<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useRoute, useRouter, RouterLink } from 'vue-router'
import { useMovieStore } from '@/stores/movie'
import { useShowtimeStore, type Showtime } from '@/stores/showtime'
import BookingSteps from '@/components/BookingSteps.vue'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import EmptyState from '@/components/EmptyState.vue'
import { formatDateTime, formatTime } from '@/utils/format'

const route = useRoute()
const router = useRouter()
const movieStore = useMovieStore()
const showtimeStore = useShowtimeStore()

const movieId = ref(route.params.id as string)

onMounted(async () => {
  await Promise.all([
    movieStore.fetchById(movieId.value),
    showtimeStore.fetchByMovieId(movieId.value)
  ])
})

const movie = computed(() => movieStore.currentMovie)
const showtimes = computed(() => showtimeStore.showtimes)
const loading = computed(() => movieStore.loading || showtimeStore.loading)
const error = computed(() => movieStore.error || showtimeStore.error)

// Group showtimes by Date and then by Theater Name
const showtimesByDateAndTheater = computed(() => {
  const groups: Record<string, Record<string, Showtime[]>> = {}

  showtimes.value.forEach((showtime) => {
    const dateStr = formatDateTime(showtime.startTime, {
      weekday: 'short',
      day: 'numeric',
      month: 'short',
    })

    if (!groups[dateStr]) groups[dateStr] = {}

    const theater = showtime.theaterName || 'Theater'
    if (!groups[dateStr][theater]) groups[dateStr][theater] = []

    groups[dateStr][theater].push(showtime)
  })

  Object.keys(groups).forEach((date) => {
    const dateGroup = groups[date]
    if (dateGroup) {
      Object.keys(dateGroup).forEach((theater) => {
        const slots = dateGroup[theater]
        if (slots) {
          slots.sort((a, b) =>
            new Date(a.startTime).getTime() - new Date(b.startTime).getTime()
          )
        }
      })
    }
  })

  return groups
})

const selectedDate = ref<string>('')
const dates = computed(() => Object.keys(showtimesByDateAndTheater.value))

const watchDates = computed(() => {
  if (dates.value.length > 0 && !selectedDate.value) {
    selectedDate.value = dates.value[0] || ''
  }
  return dates.value
})

const posterUrl = computed(() =>
  movie.value?.posterUrl || 'https://images.unsplash.com/photo-1536440136628-849c177e76a1?auto=format&fit=crop&w=800&q=80'
)
</script>

<template>
  <div class="movie-detail-page">
    <BookingSteps :current-step="1" back-url="/" />
    <span v-show="false">{{ watchDates }}</span>

    <LoadingSpinner v-if="loading" message="Loading movie details..." style="padding-top: calc(64px + 80px)" />

    <EmptyState
      v-else-if="error"
      title="An error occurred"
      style="padding-top: calc(64px + 80px)"
    >
      <RouterLink to="/" class="btn btn-primary" style="margin-top: 16px">Back to Home</RouterLink>
    </EmptyState>

    <div v-else-if="movie">
      <!-- ── Backdrop Hero ── -->
      <div class="detail-hero">
        <div class="detail-backdrop" :style="{ backgroundImage: `url(${posterUrl})` }"></div>
        <div class="detail-backdrop-gradient"></div>
        <div class="container detail-hero-content">
        </div>
      </div>

      <!-- ── Main Content ── -->
      <div class="container detail-body animate-fade-up">
        <div class="detail-grid">
          <!-- Left — Sticky Poster -->
          <div class="poster-column">
            <div class="poster-frame">
              <img :src="posterUrl" :alt="movie.title" class="detail-poster" />
              <div class="poster-rating" v-if="movie.rating">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor" style="color: var(--color-accent)">
                  <path d="M12 17.27L18.18 21l-1.64-7.03L22 9.24l-7.19-.61L12 2 9.19 8.63 2 9.24l5.46 4.73L5.82 21z"/>
                </svg>
                {{ movie.rating.toFixed(1) }}
              </div>
            </div>
          </div>

          <!-- Right — Info + Showtimes -->
          <div class="info-column">
            <!-- Title + Tags -->
            <div class="movie-header">
              <div class="meta-tags">
                <span class="meta-tag">{{ movie.duration }} min</span>
                <span class="meta-tag" v-for="g in movie.genre" :key="g">{{ g }}</span>
                <span class="meta-tag lang-tag" v-if="movie.language">{{ movie.language }}</span>
              </div>
              <h1 class="movie-headline">{{ movie.title }}</h1>
              <p class="movie-director-line">Directed by <strong>{{ movie.director }}</strong></p>
            </div>

            <!-- Synopsis Card -->
            <div class="synopsis-card">
              <div class="synopsis-label">Synopsis</div>
              <p class="synopsis-text">{{ movie.description }}</p>
              <div class="cast-row" v-if="movie.cast && movie.cast.length">
                <span class="cast-label">Cast:</span>
                <span class="cast-names">{{ movie.cast.join(', ') }}</span>
              </div>
            </div>

            <!-- Showtime Section -->
            <div class="showtime-section">
              <div class="section-title-accent">
                <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect><line x1="16" y1="2" x2="16" y2="6"></line><line x1="8" y1="2" x2="8" y2="6"></line><line x1="3" y1="10" x2="21" y2="10"></line></svg>
                Select Showtime
              </div>

              <div v-if="dates.length === 0" class="no-showtimes">
                <p>Currently, there are no showtimes for this movie.</p>
              </div>

              <div v-else>
                <!-- Date Tabs -->
                <div class="date-tabs">
                  <button
                    v-for="date in dates"
                    :key="date"
                    class="date-tab"
                    :class="{ active: selectedDate === date }"
                    @click="selectedDate = date"
                  >
                    {{ date }}
                  </button>
                </div>

                <!-- Theater Groups -->
                <div v-if="selectedDate" class="theater-list animate-fade">
                  <div
                    v-for="(slots, theater) in showtimesByDateAndTheater[selectedDate]"
                    :key="theater"
                    class="theater-group"
                  >
                    <div class="theater-name-row">
                      <span class="theater-icon">🏢</span>
                      <span class="theater-name-text">{{ theater }}</span>
                    </div>
                    <div class="slots-row">
                      <template v-for="slot in slots" :key="slot.id">
                        <!-- Past Showtime (Disabled & Faded) -->
                        <div
                          v-if="new Date(slot.startTime).getTime() < Date.now()"
                          class="slot-card slot-card-disabled"
                          title="This showtime has already passed"
                        >
                          <span class="slot-time">{{ formatTime(slot.startTime) }}</span>
                          <span class="slot-type">{{ slot.slotLabel }}</span>
                        </div>

                        <!-- Upcoming Showtime -->
                        <RouterLink
                          v-else
                          :to="`/movies/${movieId}/showtime/${slot.id}/select-seat`"
                          class="slot-card"
                        >
                          <span class="slot-time">{{ formatTime(slot.startTime) }}</span>
                          <span class="slot-type">{{ slot.slotLabel }}</span>
                        </RouterLink>
                      </template>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.movie-detail-page {
  min-height: 100vh;
  overflow-x: hidden;
}

/* ── Backdrop Hero ── */
.detail-hero {
  position: relative;
  height: 240px;
  overflow: hidden;
}

.detail-backdrop {
  position: absolute;
  inset: 0;
  background-size: cover;
  background-position: center 20%;
  filter: blur(4px) brightness(0.25);
  transform: scale(1.05);
}

.detail-backdrop-gradient {
  position: absolute;
  inset: 0;
  background: linear-gradient(
    to bottom,
    rgba(8, 8, 15, 0.3) 0%,
    rgba(8, 8, 15, 0.7) 60%,
    rgba(8, 8, 15, 1) 100%
  );
}

.detail-hero-content {
  position: relative;
  z-index: 1;
  height: 100%;
  display: flex;
  align-items: flex-start;
  padding-top: 92px;
}

.btn-back {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  color: rgba(255,255,255,0.7);
  font-size: 0.875rem;
  font-weight: 600;
  text-decoration: none;
  padding: 8px 16px;
  background: rgba(255,255,255,0.08);
  border: 1px solid rgba(255,255,255,0.12);
  border-radius: var(--radius-full);
  transition: all var(--transition-fast);
  backdrop-filter: blur(8px);
}

.btn-back:hover {
  color: #fff;
  background: rgba(255,255,255,0.14);
}

/* ── Detail Body ── */
.detail-body {
  padding-top: var(--space-xl);
  padding-bottom: var(--space-3xl);
  margin-top: -180px;
  position: relative;
  z-index: 1;
}

.detail-grid {
  display: grid;
  grid-template-columns: 280px 1fr;
  gap: var(--space-xl);
  align-items: start;
}

.info-column {
  min-width: 0;
}

/* Poster */
.poster-column {
  position: sticky;
  top: calc(64px + var(--space-lg));
}

.poster-frame {
  position: relative;
  border-radius: var(--radius-lg);
  overflow: hidden;
  box-shadow: 0 20px 60px rgba(0,0,0,0.7), 0 0 0 1px rgba(255,255,255,0.08);
  aspect-ratio: 2/3;
  background: var(--color-bg-card);
}

.detail-poster {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.poster-rating {
  position: absolute;
  top: 12px;
  right: 12px;
  display: flex;
  align-items: center;
  gap: 4px;
  background: rgba(8,8,15,0.85);
  backdrop-filter: blur(8px);
  border: 1px solid rgba(245, 166, 35, 0.4);
  color: var(--color-accent);
  padding: 5px 10px;
  border-radius: var(--radius-full);
  font-size: 0.85rem;
  font-weight: 700;
}

/* Movie header */
.movie-header {
  margin-bottom: var(--space-lg);
}

.meta-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: var(--space-sm);
}

.meta-tag {
  background: rgba(255,255,255,0.05);
  border: 1px solid var(--border-subtle);
  padding: 3px 12px;
  border-radius: var(--radius-full);
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--text-muted);
  letter-spacing: 0.04em;
}

.lang-tag {
  background: rgba(229, 9, 20, 0.1);
  border-color: rgba(229, 9, 20, 0.2);
  color: rgba(229, 9, 20, 0.8);
}

.movie-headline {
  font-family: 'Barlow Condensed', sans-serif;
  font-size: clamp(1.8rem, 3.5vw, 2.8rem);
  font-weight: 800;
  letter-spacing: 0.02em;
  line-height: 1.0;
  color: #fff;
  margin-bottom: 8px;
  overflow-wrap: anywhere;
}

.movie-director-line {
  font-size: 0.88rem;
  color: var(--text-muted);
}

.movie-director-line strong {
  color: var(--text-secondary);
}

/* Synopsis */
.synopsis-card {
  background: var(--color-bg-card);
  border: 1px solid var(--border-card);
  border-radius: var(--radius-lg);
  padding: var(--space-lg);
  margin-bottom: var(--space-xl);
}

.synopsis-label {
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--color-primary);
  margin-bottom: var(--space-sm);
}

.synopsis-text {
  font-size: 0.92rem;
  color: var(--text-secondary);
  line-height: 1.75;
  margin-bottom: var(--space-md);
}

.cast-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  font-size: 0.85rem;
  padding-top: var(--space-md);
  border-top: 1px solid var(--border-subtle);
}

.cast-label {
  color: var(--text-muted);
  flex-shrink: 0;
}

.cast-names { color: var(--text-secondary); }

/* Showtime */
.showtime-section { margin-bottom: var(--space-2xl); }

.no-showtimes {
  background: var(--color-bg-card);
  border: 1px solid var(--border-card);
  border-radius: var(--radius-lg);
  padding: var(--space-xl);
  text-align: center;
  color: var(--text-muted);
  font-size: 0.9rem;
}

/* Date Tabs */
.date-tabs {
  display: flex;
  gap: 8px;
  overflow-x: auto;
  padding-bottom: 4px;
  margin-bottom: var(--space-lg);
  scrollbar-width: none;
  scroll-snap-type: x proximity;
  overscroll-behavior-inline: contain;
}
.date-tabs::-webkit-scrollbar { display: none; }

.date-tab {
  padding: 8px 20px;
  border-radius: var(--radius-full);
  background: var(--color-bg-card);
  border: 1px solid var(--border-card);
  color: var(--text-muted);
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;
  white-space: nowrap;
  transition: all var(--transition-base);
  scroll-snap-align: start;
}

.date-tab:hover {
  border-color: var(--border-card);
  color: var(--text-primary);
  background: var(--color-bg-elevated);
}

.date-tab.active {
  background: var(--color-primary);
  border-color: var(--color-primary);
  color: #fff;
  box-shadow: 0 4px 16px var(--color-primary-glow);
}

/* Theater groups */
.theater-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
}

.theater-group {
  background: var(--color-bg-card);
  border: 1px solid var(--border-card);
  border-radius: var(--radius-lg);
  padding: var(--space-lg);
}

.theater-name-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: var(--space-md);
}

.theater-icon { font-size: 1.1rem; }

.theater-name-text {
  font-family: 'Outfit', sans-serif;
  font-size: 1rem;
  font-weight: 700;
  color: var(--text-primary);
}

.slots-row {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(90px, 1fr));
  gap: 10px;
}

.slot-card {
  background: rgba(255,255,255,0.04);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  padding: 10px 18px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  text-decoration: none;
  transition: all var(--transition-base);
  min-width: 0;
}

.slot-card:not(.slot-card-disabled):hover {
  background: rgba(229, 9, 20, 0.1);
  border-color: var(--color-primary);
  transform: scale(1.04);
  box-shadow: 0 4px 16px var(--color-primary-glow);
}

.slot-card-disabled {
  opacity: 0.35;
  cursor: not-allowed;
  background: rgba(255, 255, 255, 0.02);
  border-color: rgba(255, 255, 255, 0.05);
}

.slot-time {
  font-family: 'Barlow Condensed', sans-serif;
  font-size: 1.3rem;
  font-weight: 700;
  color: var(--text-primary);
  letter-spacing: 0.04em;
}

.slot-type {
  font-size: 0.65rem;
  color: var(--text-muted);
  font-weight: 600;
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

@media (max-width: 1024px) {
  .detail-hero {
    height: 220px;
  }

  .detail-body {
    margin-top: -160px;
  }

  .detail-grid {
    grid-template-columns: minmax(190px, 220px) minmax(0, 1fr);
    gap: var(--space-lg);
  }

  .poster-column {
    top: calc(64px + var(--space-md));
  }

  .synopsis-card,
  .theater-group {
    padding: var(--space-md);
  }
}

@media (max-width: 768px) {
  .detail-hero {
    height: 190px;
  }

  .detail-body {
    margin-top: -118px;
    padding-top: var(--space-md);
    padding-bottom: var(--space-2xl);
  }

  .detail-grid {
    grid-template-columns: minmax(0, 1fr);
    gap: var(--space-lg);
  }

  .poster-column {
    position: relative;
    top: 0;
    width: min(44vw, 190px);
    margin: 0 auto;
  }

  .movie-header {
    text-align: center;
  }

  .meta-tags {
    justify-content: center;
  }

  .movie-headline {
    font-size: clamp(2rem, 10vw, 2.6rem);
  }

  .synopsis-card {
    margin-bottom: var(--space-lg);
  }

  .date-tabs {
    margin-inline: calc(var(--space-md) * -1);
    padding-inline: var(--space-md);
    scroll-padding-inline: var(--space-md);
  }
}

@media (max-width: 480px) {
  .detail-hero {
    height: 170px;
  }

  .detail-body {
    margin-top: -96px;
  }

  .poster-column {
    width: min(48vw, 172px);
  }

  .synopsis-card,
  .theater-group {
    border-radius: var(--radius-md);
    padding: 14px;
  }

  .synopsis-text {
    font-size: 0.88rem;
    line-height: 1.65;
  }

  .section-title-accent {
    font-size: 1rem;
  }

  .slots-row {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 8px;
  }

  .slot-card {
    padding: 10px 8px;
  }

  .slot-time {
    font-size: 1.2rem;
  }
}
</style>
