<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useMovieStore } from '@/stores/movie'
import { movieApi } from '@/api'
import { CINEMA_HALLS, SHOWTIME_SLOTS } from '@/constants'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import EmptyState from '@/components/EmptyState.vue'
import ModalDialog from '@/components/ModalDialog.vue'
import AppButton from '@/components/AppButton.vue'
import AutocompleteInput from '@/components/AutocompleteInput.vue'
import { useAutocompleteSearch } from '@/composables/useAutocompleteSearch'
import { createDefaultMovieForm, parseCommaSeparatedList } from '@/utils/movieForm'

const movieStore = useMovieStore()

// State
const isModalOpen = ref(false)
const submitting = ref(false)
const formError = ref<string | null>(null)
const successMsg = ref<string | null>(null)
const searchTitle = ref('')
const activeSearchQuery = ref('')

// Form Data
const newMovie = ref(createDefaultMovieForm())

onMounted(() => {
  movieStore.fetchAll()
})

const movies = computed(() => movieStore.movies)
const filteredMovies = computed(() => {
  const query = activeSearchQuery.value.trim().toLowerCase()
  if (!query) return movies.value

  return movies.value.filter((movie) =>
    movie.title.toLowerCase().includes(query) ||
    movie.director.toLowerCase().includes(query),
  )
})
const loading = computed(() => movieStore.loading)
const error = computed(() => movieStore.error)

const {
  suggestions: movieSuggestions,
  showSuggestions: showMovieSuggestions,
  selectSuggestion: selectMovieSuggestion,
  triggerSearchInstant,
} = useAutocompleteSearch({
  query: searchTitle,
  fetchSuggestions: async (query) => {
    const res = await movieApi.getSuggestions(query, 1, 10)
    return res.data.data || []
  },
  onSearch: (query) => {
    activeSearchQuery.value = query
  },
  searchDelay: 3000,
  onSuggestionError: (error) => {
    console.error('Failed to fetch movie suggestions:', error)
  },
})

function openModal() {
  isModalOpen.value = true
  formError.value = null
  successMsg.value = null
  newMovie.value = createDefaultMovieForm()
}

function closeModal() {
  isModalOpen.value = false
}

async function handleSubmit() {
  submitting.value = true
  formError.value = null
  successMsg.value = null

  // Format array fields
  const payload = {
    ...newMovie.value,
    genre: parseCommaSeparatedList(newMovie.value.genre),
    cast: parseCommaSeparatedList(newMovie.value.cast)
  }

  try {
    await movieApi.create(payload)
    successMsg.value = 'Movie added successfully!'
    // Refetch list
    await movieStore.fetchAll()
    setTimeout(() => {
      closeModal()
    }, 1500)
  } catch (err: any) {
    formError.value = err.response?.data?.error || 'Could not add movie. Please verify information.'
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <main class="admin-main animate-fade-up">
        <div class="admin-header">
          <div class="header-action-row">
            <div>
              <h1 class="page-title">Manage Movies</h1>
            </div>
            <AppButton variant="primary" @click="openModal">
              <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="margin-right: 4px; vertical-align: middle;"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
              Add New Movie
            </AppButton>
          </div>
        </div>

        <div class="filter-bar glass-card">
          <div class="form-group search-input-group">
            <label class="form-label">Search Movie</label>
            <AutocompleteInput
              v-model="searchTitle"
              placeholder="Movie name or director..."
              :suggestions="movieSuggestions"
              v-model:show-suggestions="showMovieSuggestions"
              @select="selectMovieSuggestion"
              @enter="triggerSearchInstant"
            />
          </div>
        </div>

         <LoadingSpinner v-slot:default v-if="loading" message="Loading movies..." />

        <EmptyState
          v-else-if="error"
          title="An error occurred"
        />

        <EmptyState v-else-if="movies.length === 0" title="No Movies Found" />

        <EmptyState
          v-else-if="filteredMovies.length === 0"
          title="No movies match your search"
        />

        <!-- Movies Grid -->
        <div v-else class="movies-admin-grid">
          <div
            v-for="movie in filteredMovies"
            :key="movie.id"
            class="movie-admin-card glass-card"
          >
            <div class="card-inner">
              <div class="poster-preview">
                <img :src="movie.posterUrl || 'https://images.unsplash.com/photo-1536440136628-849c177e76a1?auto=format&fit=crop&w=400&q=80'" :alt="movie.title" />
              </div>
              <div class="card-details">
                <div class="status-row">
                  <span class="badge" :class="movie.isActive ? 'badge-success' : 'badge-muted'">
                    {{ movie.isActive ? 'Active' : 'Inactive' }}
                  </span>
                  <span class="rating-val" v-if="movie.rating">⭐ {{ movie.rating.toFixed(1) }}</span>
                </div>
                <h3 class="movie-title-admin">{{ movie.title }}</h3>
                <div class="meta-row">
                  <span>{{ movie.duration }} min</span> | 
                  <span>{{ movie.language }}</span>
                </div>
                <div class="genres-row">
                  <span class="genre-tag" v-for="g in movie.genre" :key="g">{{ g }}</span>
                </div>
                <p class="desc-preview">{{ movie.description }}</p>
              </div>
            </div>
          </div>
        </div>
    <!-- Modal Form Dialog -->
    <!-- Modal Form Dialog -->
    <ModalDialog :isOpen="isModalOpen" maxWidth="680px" @close="closeModal">
      <template #header>
        <h3 style="margin: 0;">Add New Movie</h3>
      </template>

      <div class="alert alert-success" v-if="successMsg">🎉 {{ successMsg }}</div>
      <div class="alert alert-danger" v-if="formError">⚠️ {{ formError }}</div>

      <form id="add-movie-form" @submit.prevent="handleSubmit" v-if="!successMsg">
        <div class="form-grid">
          <div class="form-group span-2">
            <label class="form-label">Movie Title *</label>
            <input type="text" class="form-input" v-model="newMovie.title" required placeholder="e.g. Dune: Part Two" />
          </div>

          <div class="form-group span-2">
            <label class="form-label">Synopsis *</label>
            <textarea class="form-input text-area" v-model="newMovie.description" required placeholder="Brief description..."></textarea>
          </div>

          <div class="form-group">
            <label class="form-label">Genres (separated by commas) *</label>
            <input type="text" class="form-input" v-model="newMovie.genre" required placeholder="Action, Sci-Fi, Adventure" />
          </div>

          <div class="form-group">
            <label class="form-label">Poster Image Link (URL)</label>
            <input type="url" class="form-input" v-model="newMovie.posterUrl" placeholder="https://..." />
          </div>

          <div class="form-group">
            <label class="form-label">Duration (minutes) *</label>
            <input type="number" class="form-input" v-model.number="newMovie.duration" required min="1" />
          </div>

          <div class="form-group">
            <label class="form-label">Movie Rating (Rating) *</label>
            <input type="number" class="form-input" v-model.number="newMovie.rating" required step="0.1" min="0" max="10" />
          </div>

          <div class="form-group">
            <label class="form-label">Language *</label>
            <input type="text" class="form-input" v-model="newMovie.language" required placeholder="EN/TH" />
          </div>

          <div class="form-group">
            <label class="form-label">Director</label>
            <input type="text" class="form-input" v-model="newMovie.director" placeholder="Director's name" />
          </div>

          <div class="form-group span-2">
            <label class="form-label">Cast members (separated by commas)</label>
            <input type="text" class="form-input" v-model="newMovie.cast" placeholder="Actor 1, Actor 2" />
          </div>

          <div class="form-group">
            <label class="form-label">Release Date</label>
            <input type="date" class="form-input" v-model="newMovie.releaseDate" />
          </div>

          <div class="form-group checkbox-group">
            <label class="checkbox-label">
              <input type="checkbox" v-model="newMovie.isActive" />
              Show in cinema (Active)
            </label>
          </div>

          <!-- Showtime Configurations -->
          <div class="form-group span-2 showtimes-config">
            <h4 class="config-title">Assign Showtimes & Halls</h4>
            <p class="config-help">Select which halls and time slots will be scheduled for this movie (generates schedules for the next 5 days).</p>
            
            <div class="config-grid">
              <div class="config-column">
                <label class="config-section-label">Theater Halls</label>
                <div class="checkbox-list">
                  <label v-for="hall in CINEMA_HALLS" :key="hall.value" class="checkbox-label">
                    <input type="checkbox" :value="hall.value" v-model="newMovie.halls" />
                    {{ hall.label }}
                  </label>
                </div>
              </div>

              <div class="config-column">
                <label class="config-section-label">Time Slots</label>
                <div class="checkbox-list">
                  <label v-for="slot in SHOWTIME_SLOTS" :key="slot.value" class="checkbox-label">
                    <input type="checkbox" :value="slot.value" v-model="newMovie.slots" />
                    {{ slot.label }}
                  </label>
                </div>
              </div>
            </div>
          </div>
        </div>
      </form>

      <template #footer>
        <template v-if="!successMsg">
          <AppButton variant="secondary" @click="closeModal" :disabled="submitting">Cancel</AppButton>
          <AppButton type="submit" variant="primary" form="add-movie-form" :loading="submitting">Save Movie</AppButton>
        </template>
        <template v-else>
          <AppButton variant="primary" @click="closeModal">OK</AppButton>
        </template>
      </template>
    </ModalDialog>
  </main>
</template>

<style scoped>
.admin-main {
  display: flex;
  flex-direction: column;
  gap: var(--space-xl);
}

.admin-header {
  border-bottom: 1px solid var(--border-subtle);
  padding-bottom: var(--space-md);
}

.header-action-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  gap: var(--space-md);
  flex-wrap: wrap;
}

.text-danger {
  color: #ef4444;
}

.filter-bar {
  display: flex;
  align-items: flex-end;
  padding: var(--space-md) !important;
  background: rgba(255,255,255,0.02);
}

.filter-bar:hover {
  transform: none;
  box-shadow: var(--shadow-card);
}

.search-input-group {
  width: min(100%, 420px);
  margin-bottom: 0;
}

/* Movies Grid */
.movies-admin-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: var(--space-md);
}

.movie-admin-card {
  padding: 16px;
  background: rgba(255,255,255,0.02);
}
.movie-admin-card:hover {
  transform: translateY(-1px);
}

.card-inner {
  display: flex;
  gap: var(--space-md);
}

.poster-preview {
  width: 100px;
  height: 140px;
  border-radius: var(--radius-sm);
  overflow: hidden;
  border: 1px solid var(--border-subtle);
  flex-shrink: 0;
}

.poster-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.card-details {
  display: flex;
  flex-direction: column;
  gap: 6px;
  flex-grow: 1;
}

.status-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.rating-val {
  font-size: 0.82rem;
  color: var(--color-accent);
  font-weight: 700;
}

.movie-title-admin {
  font-size: 1.25rem;
  font-weight: 700;
  margin: 0;
  color: var(--text-primary);
}

.meta-row {
  font-size: 0.8rem;
  color: var(--text-muted);
}

.genres-row {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.genre-tag {
  font-size: 0.7rem;
  background: rgba(255,255,255,0.05);
  border: 1px solid var(--border-subtle);
  padding: 2px 8px;
  border-radius: var(--radius-full);
  color: var(--text-secondary);
}

.desc-preview {
  font-size: 0.85rem;
  color: var(--text-secondary);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  line-height: 1.5;
  margin-top: 4px;
}

/* Modal form styling */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.8);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  padding: var(--space-md);
}

.modal-card {
  width: 100%;
  max-width: 680px;
  background: var(--color-bg-secondary);
  border: 1px solid var(--border-card);
  padding: var(--space-xl);
  max-height: 90vh;
  overflow-y: auto;
}
.modal-card:hover {
  transform: none;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.modal-header h3 {
  font-size: 1.35rem;
  font-weight: 700;
}

.btn-close {
  background: transparent;
  border: none;
  color: var(--text-secondary);
  font-size: 1.2rem;
  cursor: pointer;
}
.btn-close:hover {
  color: var(--color-primary);
}

.modal-body {
  margin-top: var(--space-md);
}

.alert {
  padding: 12px;
  border-radius: var(--radius-md);
  margin-bottom: var(--space-md);
  font-size: 0.9rem;
}
.alert-success { background: rgba(34,197,94,0.15); border: 1px solid rgba(34,197,94,0.3); color: #4ade80; }
.alert-danger { background: rgba(239,68,68,0.15); border: 1px solid rgba(239,68,68,0.3); color: #f87171; }

.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--space-md);
}

.span-2 {
  grid-column: span 2;
}

.text-area {
  height: 80px;
  resize: vertical;
}

.checkbox-group {
  display: flex;
  align-items: center;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 0.9rem;
  color: var(--text-primary);
  cursor: pointer;
}

.checkbox-label input {
  width: 18px;
  height: 18px;
  accent-color: var(--color-primary);
}

.form-actions-row {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: var(--space-md);
}

.spinner-sm {
  display: inline-block;
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255,255,255,0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}



.showtimes-config {
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  padding: var(--space-md);
  margin-top: var(--space-xs);
}

.config-title {
  font-size: 0.95rem;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 4px;
}

.config-help {
  font-size: 0.8rem;
  color: var(--text-secondary);
  margin-bottom: var(--space-md);
}

.config-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--space-md);
}

.config-section-label {
  display: block;
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--color-accent);
  margin-bottom: var(--space-sm);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.checkbox-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

@media (max-width: 600px) {
  .form-grid {
    grid-template-columns: 1fr;
  }
  .span-2 {
    grid-column: span 1;
  }
  .config-grid {
    grid-template-columns: 1fr;
    gap: var(--space-md);
  }
}
</style>
