<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useMovieStore } from '@/stores/movie'
import { useAuthStore } from '@/stores/auth'
import MovieCard from '@/components/MovieCard.vue'
import { movieApi } from '@/api'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import EmptyState from '@/components/EmptyState.vue'
import AutocompleteInput from '@/components/AutocompleteInput.vue'
import { useAutocompleteSearch } from '@/composables/useAutocompleteSearch'

const router = useRouter()
const movieStore = useMovieStore()
const authStore = useAuthStore()

const searchQuery = ref('')
const activeSearchQuery = ref('')
const selectedGenre = ref('')

const {
  suggestions: movieSuggestions,
  showSuggestions: showMovieSuggestions,
  selectSuggestion: selectMovieSuggestion,
  triggerSearchInstant,
} = useAutocompleteSearch({
  query: searchQuery,
  fetchSuggestions: async (query) => {
    const res = await movieApi.getSuggestions(query)
    return res.data.data || []
  },
  onSearch: (query) => {
    activeSearchQuery.value = query
  },
  onSuggestionError: (error) => {
    console.error('Failed to fetch movie suggestions:', error)
  },
})

onMounted(() => {
  movieStore.fetchAll()
})

const allGenres = computed(() => {
  const genres = new Set<string>()
  movieStore.movies.forEach((m) => m.genre?.forEach((g) => genres.add(g)))
  return ['All', ...Array.from(genres).sort()]
})

const filteredMovies = computed(() => {
  return movieStore.movies.filter((m) => {
    const matchSearch =
      !activeSearchQuery.value ||
      m.title.toLowerCase().includes(activeSearchQuery.value.toLowerCase()) ||
      m.director.toLowerCase().includes(activeSearchQuery.value.toLowerCase())
    const matchGenre =
      !selectedGenre.value ||
      selectedGenre.value === 'All' ||
      m.genre?.includes(selectedGenre.value)
    return matchSearch && matchGenre
  })
})

// Hero movie — use first movie for backdrop
const heroMovie = computed(() => movieStore.movies[0] || null)

function goToMovie(movieId: string) {
  router.push(`/movies/${movieId}/showtime`)
}
</script>

<template>
  <main class="home-page">
    <!-- ── Cinematic Hero Banner ── -->
    <section class="hero-section" v-if="heroMovie && !movieStore.loading">
      <div
        class="hero-backdrop"
        :style="{ backgroundImage: `url(${heroMovie.posterUrl || 'https://images.unsplash.com/photo-1536440136628-849c177e76a1?w=1600&q=80'})` }"
      ></div>
      <div class="hero-gradient"></div>
      <div class="container hero-content">
        <h1 class="hero-title">YOUR CINEMATIC<br/>EXPERIENCE AWAITS</h1>
        <p class="hero-subtitle">
          Book seats instantly with real-time availability across all Cineplex theaters.
        </p>
        <div class="hero-stats">
          <div class="stat-item">
            <span class="stat-num">{{ movieStore.movies.length }}</span>
            <span class="stat-label">Films</span>
          </div>
          <div class="stat-sep"></div>
          <div class="stat-item">
            <span class="stat-num">2</span>
            <span class="stat-label">Theaters</span>
          </div>
        </div>
        <div class="hero-ctas" v-if="!authStore.isAuthenticated">
          <router-link to="/login" class="btn btn-primary btn-lg">Book Now</router-link>
        </div>
      </div>
    </section>

    <!-- ── Minimal Hero (loading / no movies) ── -->
    <section class="hero-minimal page-content" v-else-if="movieStore.loading">
      <div class="container">
        <LoadingSpinner message="Loading movies..." />
      </div>
    </section>

    <!-- ── Main Content ── -->
    <div class="container main-content">
      <!-- Search + Filter Bar -->
      <section class="filter-bar animate-fade-up">
        <AutocompleteInput
          v-model="searchQuery"
          placeholder="Search movies or directors..."
          :suggestions="movieSuggestions"
          v-model:show-suggestions="showMovieSuggestions"
          @select="selectMovieSuggestion"
          @enter="triggerSearchInstant"
        />
        <div class="genre-pills">
          <button
            v-for="genre in allGenres"
            :key="genre"
            class="pill-filter"
            :class="{ active: selectedGenre === genre || (genre === 'All' && !selectedGenre) }"
            @click="selectedGenre = genre === 'All' ? '' : genre"
          >
            {{ genre }}
          </button>
        </div>
      </section>

      <!-- Movies Grid -->
      <section class="movies-section">
        <LoadingSpinner v-if="movieStore.loading" message="Loading movies..." />

        <EmptyState
          v-else-if="movieStore.error"
          title="An error occurred"
        />

        <template v-else>
          <div class="section-title-accent animate-fade-up">
            Now Showing — {{ filteredMovies.length }} Films
          </div>

          <div v-if="filteredMovies.length" class="movies-grid animate-fade-up">
            <div
              v-for="movie in filteredMovies"
              :key="movie.id"
              @click="goToMovie(movie.id)"
            >
              <MovieCard :movie="movie" />
            </div>
          </div>

          <EmptyState
            v-else
            title="No movies found"
          />
        </template>
      </section>
    </div>
  </main>
</template>

<style scoped>
.home-page {
  min-height: 100vh;
}

/* ── Hero ── */
.hero-section {
  position: relative;
  height: 520px;
  display: flex;
  align-items: flex-end;
  overflow: hidden;
  margin-top: 64px; /* navbar height */
}

.hero-backdrop {
  position: absolute;
  inset: 0;
  background-size: cover;
  background-position: center top;
  filter: blur(3px) brightness(0.35);
  transform: scale(1.05);
}

.hero-gradient {
  position: absolute;
  inset: 0;
  background: linear-gradient(
    to bottom,
    rgba(8, 8, 15, 0.2) 0%,
    rgba(8, 8, 15, 0.5) 50%,
    rgba(8, 8, 15, 0.98) 100%
  );
}

.hero-content {
  position: relative;
  z-index: 1;
  padding-bottom: var(--space-2xl);
  width: 100%;
}

.hero-title {
  font-family: 'Barlow Condensed', sans-serif;
  font-size: clamp(2.4rem, 6vw, 4rem);
  font-weight: 900;
  letter-spacing: 0.03em;
  line-height: 1.0;
  color: #fff;
  margin-bottom: var(--space-md);
  text-shadow: 0 2px 40px rgba(0,0,0,0.8);
}

.hero-subtitle {
  font-size: 1rem;
  color: rgba(255, 255, 255, 0.6);
  max-width: 480px;
  margin-bottom: var(--space-xl);
  line-height: 1.6;
}

.hero-stats {
  display: flex;
  align-items: center;
  gap: var(--space-lg);
  margin-bottom: var(--space-lg);
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
}

.stat-num {
  font-family: 'Barlow Condensed', sans-serif;
  font-size: 1.8rem;
  font-weight: 800;
  color: #fff;
  line-height: 1;
}

.stat-label {
  font-size: 0.7rem;
  font-weight: 600;
  color: rgba(255,255,255,0.45);
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.stat-sep {
  width: 1px;
  height: 32px;
  background: rgba(255,255,255,0.15);
}

.hero-ctas {
  display: flex;
  gap: var(--space-md);
}

.hero-minimal {
  padding-top: calc(64px + var(--space-xl));
}

/* ── Filter Bar ── */
.main-content {
  padding-top: var(--space-xl);
  padding-bottom: var(--space-3xl);
}

.filter-bar {
  position: relative;
  z-index: 2;
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
  margin-bottom: var(--space-xl);
}

/* Autocomplete styling relocated to BaseAutocomplete.vue */

.genre-pills {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-sm);
}

/* ── Movies Grid ── */
.movies-section {
  position: relative;
  z-index: 1;
  padding-bottom: var(--space-2xl);
}

.movies-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(190px, 1fr));
  gap: var(--space-lg);
}

@media (max-width: 992px) {
  .hero-section { height: 380px; }
  .hero-title { font-size: 2rem; }
  .genre-pills {
    flex-wrap: nowrap;
    overflow-x: auto;
    padding-bottom: 6px;
    scrollbar-width: none;
    -ms-overflow-style: none;
    width: 100%;
  }
  .genre-pills::-webkit-scrollbar {
    display: none;
  }
  .movies-grid {
    grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
    gap: var(--space-md);
  }
}
</style>
