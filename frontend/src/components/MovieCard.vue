<script setup lang="ts">
import { computed } from 'vue'
import type { Movie } from '@/stores/movie'
import { formatDurationMinutes } from '@/utils/format'

const props = defineProps<{ movie: Movie }>()

const genres = computed(() => props.movie.genre?.slice(0, 2) || [])
const durationFormatted = computed(() => formatDurationMinutes(props.movie.duration))

const posterUrl = computed(() =>
  props.movie.posterUrl || 'https://images.unsplash.com/photo-1489599849927-2ee91cede3ba?w=400&h=600&fit=crop',
)
</script>

<template>
  <article class="movie-card">
    <!-- Poster -->
    <div class="poster-wrap">
      <img :src="posterUrl" :alt="movie.title" class="poster" loading="lazy" />

      <!-- Rating chip always visible -->
      <div class="rating-chip" v-if="movie.rating">
        <svg width="11" height="11" viewBox="0 0 24 24" fill="currentColor">
          <path d="M12 17.27L18.18 21l-1.64-7.03L22 9.24l-7.19-.61L12 2 9.19 8.63 2 9.24l5.46 4.73L5.82 21z"/>
        </svg>
        {{ movie.rating.toFixed(1) }}
      </div>

      <!-- Hover Overlay -->
      <div class="poster-overlay">
        <div class="overlay-inner">
          <div class="overlay-meta">
            <span class="duration-chip">🕒 {{ durationFormatted }}</span>
          </div>
          <span class="book-cta">Book Now →</span>
        </div>
      </div>
    </div>

    <!-- Card Body -->
    <div class="card-body">
      <h3 class="movie-title">{{ movie.title }}</h3>
      <p class="movie-director">{{ movie.director }}</p>
      <div class="genres">
        <span v-for="genre in genres" :key="genre" class="genre-tag">{{ genre }}</span>
        <span class="lang-tag">{{ movie.language }}</span>
      </div>
    </div>
  </article>
</template>

<style scoped>
.movie-card {
  background: var(--color-bg-card);
  border: 1px solid var(--color-glass-border);
  border-radius: var(--radius-lg);
  overflow: hidden;
  cursor: pointer;
  transition: all var(--transition-base);
  display: flex;
  flex-direction: column;
}

.movie-card:hover {
  border-color: rgba(229, 9, 20, 0.3);
  transform: translateY(-8px);
  box-shadow: 0 20px 60px rgba(0,0,0,0.7), 0 0 0 1px rgba(229,9,20,0.15);
}

/* Poster */
.poster-wrap {
  position: relative;
  aspect-ratio: 2/3;
  overflow: hidden;
}

.poster {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform var(--transition-slow);
}

.movie-card:hover .poster {
  transform: scale(1.06);
}

/* Rating chip */
.rating-chip {
  position: absolute;
  top: 10px;
  left: 10px;
  display: flex;
  align-items: center;
  gap: 3px;
  background: rgba(0,0,0,0.75);
  backdrop-filter: blur(6px);
  -webkit-backdrop-filter: blur(6px);
  border: 1px solid rgba(245, 166, 35, 0.4);
  color: var(--color-accent);
  padding: 4px 8px;
  border-radius: var(--radius-full);
  font-size: 0.72rem;
  font-weight: 700;
  z-index: 2;
}

/* Hover overlay */
.poster-overlay {
  position: absolute;
  inset: 0;
  background: linear-gradient(to top, rgba(8,8,15,0.95) 0%, rgba(8,8,15,0.4) 50%, transparent 100%);
  display: flex;
  align-items: flex-end;
  opacity: 0;
  transition: opacity var(--transition-base);
  z-index: 1;
}

.movie-card:hover .poster-overlay {
  opacity: 1;
}

.overlay-inner {
  width: 100%;
  padding: var(--space-md);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
}


.overlay-meta {
  display: flex;
  gap: 8px;
}

.duration-chip {
  background: rgba(0,0,0,0.6);
  backdrop-filter: blur(4px);
  padding: 3px 10px;
  border-radius: var(--radius-full);
  font-size: 0.72rem;
  font-weight: 600;
  color: rgba(255,255,255,0.85);
}

.book-cta {
  font-size: 0.82rem;
  font-weight: 700;
  color: var(--color-primary);
  letter-spacing: 0.04em;
}

/* Card Body */
.card-body {
  padding: 12px var(--space-md);
  display: flex;
  flex-direction: column;
  gap: 4px;
  flex: 1;
  border-top: 1px solid rgba(229, 9, 20, 0.08);
}

.movie-title {
  font-family: 'Outfit', sans-serif;
  font-size: 0.95rem;
  font-weight: 700;
  color: var(--text-primary);
  line-height: 1.35;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.movie-director {
  font-size: 0.72rem;
  color: var(--text-muted);
  margin-top: 2px;
}

.genres {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  margin-top: 6px;
}

.genre-tag {
  background: rgba(229, 9, 20, 0.1);
  color: rgba(229, 9, 20, 0.85);
  border: 1px solid rgba(229, 9, 20, 0.2);
  padding: 2px 8px;
  border-radius: var(--radius-full);
  font-size: 0.67rem;
  font-weight: 700;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.lang-tag {
  background: rgba(255,255,255,0.05);
  color: var(--text-muted);
  border: 1px solid var(--border-subtle);
  padding: 2px 8px;
  border-radius: var(--radius-full);
  font-size: 0.67rem;
  font-weight: 600;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}
</style>
