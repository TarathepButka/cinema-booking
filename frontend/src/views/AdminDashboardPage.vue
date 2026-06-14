<script setup lang="ts">
import { onMounted, ref, watch, nextTick } from 'vue'
import { bookingApi, movieApi, userApi } from '@/api'
import { PAGINATION } from '@/constants'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import EmptyState from '@/components/EmptyState.vue'
import AppButton from '@/components/AppButton.vue'
import AutocompleteInput from '@/components/AutocompleteInput.vue'
import { useAutocompleteSearch } from '@/composables/useAutocompleteSearch'
import { formatBookingCode, formatCurrency, formatDateTime, formatSeatLabels } from '@/utils/format'
import { getBookingStatusPresentation } from '@/utils/status'

// State
const showFilters = ref(false)
const bookings = ref<any[]>([])
const totalBookings = ref(0)
const stats = ref<any[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const isResettingFilters = ref(false)

// Filters
const searchTitle = ref('')
const searchEmail = ref('')
const selectedStatus = ref('')
const startDate = ref('')
const endDate = ref('')
const page = ref(PAGINATION.DEFAULT_PAGE)
const limit = ref(PAGINATION.DEFAULT_LIMIT)
const PAGE_SIZE_OPTIONS = PAGINATION.PAGE_SIZE_OPTIONS
const endDateInput = ref<HTMLInputElement | null>(null)

// Aggregate stats values
const totalRevenue = ref(0)
const totalConfirmedBookingsCount = ref(0)

function buildFilterParams() {
  const params: Record<string, any> = {}
  if (searchTitle.value) params.movie_title = searchTitle.value
  if (searchEmail.value) params.user_email = searchEmail.value
  if (selectedStatus.value) params.status = selectedStatus.value
  if (startDate.value) params.start_date = startDate.value
  if (endDate.value) params.end_date = endDate.value
  return params
}

async function fetchStats() {
  try {
    const params = buildFilterParams()
    const res = await bookingApi.getStats(params)
    stats.value = res.data.data || []
    
    // Sum up totals
    let revenueSum = 0
    let countSum = 0
    stats.value.forEach((item: any) => {
      revenueSum += item.totalRevenue || 0
      countSum += item.count || 0
    })
    totalRevenue.value = revenueSum
    totalConfirmedBookingsCount.value = countSum
  } catch (e) {
    console.error('Failed to load stats:', e)
  }
}

async function fetchBookings(scrollTop = false) {
  loading.value = true
  error.value = null
  try {
    const params = buildFilterParams()
    params.page = page.value
    params.limit = limit.value

    const res = await bookingApi.getAll(params)
    bookings.value = res.data.data || []
    totalBookings.value = res.data.total || 0
    if (scrollTop) {
      await nextTick()
      window.scrollTo({ top: 0, behavior: 'instant' })
    }
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to fetch bookings'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchStats()
  fetchBookings()
})

function handleStartDateChange() {
  // Clamp end date if it's before the new start date
  if (startDate.value && endDate.value && endDate.value < startDate.value) {
    endDate.value = startDate.value
  }
}

watch(startDate, (newStart) => {
  if (newStart && endDate.value && endDate.value < newStart) {
    endDate.value = newStart
  }
})

// Refetch on page/filter/limit change
watch([page, selectedStatus, startDate, endDate, limit], () => {
  if (isResettingFilters.value) return
  // scroll to top only on page/limit change (not filter change)
  fetchBookings(true)
  fetchStats()
})

function handleLimitChange() {
  page.value = 1
}

function triggerSearch() {
  if (page.value !== 1) {
    page.value = 1
  } else {
    fetchBookings()
    fetchStats()
  }
}

const {
  suggestions: movieSuggestions,
  showSuggestions: showMovieSuggestions,
  selectSuggestion: selectMovieSuggestion,
  triggerSearchInstant: triggerMovieSearchInstant,
  clearSuggestions: clearMovieSuggestions,
  clearTimers: clearMovieAutocompleteTimers,
} = useAutocompleteSearch({
  query: searchTitle,
  fetchSuggestions: async (query) => {
    const res = await movieApi.getSuggestions(query, 1, 10)
    return res.data.data || []
  },
  onSearch: () => {
    triggerSearch()
  },
  searchDelay: 3000,
  shouldSkip: () => isResettingFilters.value,
  onSuggestionError: (error) => {
    console.error('Failed to fetch movie suggestions:', error)
  },
})

const {
  suggestions: emailSuggestions,
  showSuggestions: showEmailSuggestions,
  selectSuggestion: selectEmailSuggestion,
  triggerSearchInstant: triggerEmailSearchInstant,
  clearSuggestions: clearEmailSuggestions,
  clearTimers: clearEmailAutocompleteTimers,
} = useAutocompleteSearch({
  query: searchEmail,
  fetchSuggestions: async (query) => {
    const res = await userApi.getSuggestions(query, 1, 10)
    return res.data.data || []
  },
  onSearch: () => {
    triggerSearch()
  },
  searchDelay: 5000,
  shouldSkip: () => isResettingFilters.value,
  onSuggestionError: (error) => {
    console.error('Failed to fetch email suggestions:', error)
  },
})

function clearFilters() {
  isResettingFilters.value = true

  clearMovieAutocompleteTimers()
  clearEmailAutocompleteTimers()
  clearMovieSuggestions()
  clearEmailSuggestions()

  searchTitle.value = ''
  searchEmail.value = ''
  startDate.value = ''
  endDate.value = ''
  selectedStatus.value = ''
  
  nextTick(() => {
    isResettingFilters.value = false
    triggerSearch()
  })
}
</script>

<template>
  <main class="admin-main animate-fade-up">
        <div class="admin-header">
          <h1 class="page-title">All Booking Transactions</h1>
        </div>

        <!-- Mobile filter toggle -->
        <div class="filter-toggle-row">
          <button class="btn btn-secondary btn-sm filter-toggle-btn" @click="showFilters = !showFilters" style="width: 100%">
            <span v-if="showFilters">Hide Filters ✕</span>
            <span v-else>Show Filters 🔍</span>
          </button>
        </div>

        <!-- Filter bar -->
        <div class="filter-bar glass-card" :class="{ 'mobile-hidden': !showFilters }">
          <!-- Movie Name Search with Auto-Suggestion -->
          <div class="form-group search-input-group autocomplete-wrap">
            <label class="form-label">Search Movie</label>
            <AutocompleteInput
              v-model="searchTitle"
              placeholder="Movie name..."
              :suggestions="movieSuggestions"
              v-model:show-suggestions="showMovieSuggestions"
              @select="selectMovieSuggestion"
              @enter="triggerMovieSearchInstant"
            />
          </div>

          <!-- User Email Search with Auto-Suggestion -->
          <div class="form-group search-input-group autocomplete-wrap">
            <label class="form-label">Search User Email</label>
            <AutocompleteInput
              v-model="searchEmail"
              placeholder="Email..."
              :suggestions="emailSuggestions"
              v-model:show-suggestions="showEmailSuggestions"
              @select="selectEmailSuggestion"
              @enter="triggerEmailSearchInstant"
            />
          </div>

          <!-- Start Date -->
          <div class="form-group filter-date-group">
            <label class="form-label">Start Date</label>
            <input
              type="date"
              class="form-input date-input"
              :class="{ 'date-empty': !startDate }"
              v-model="startDate"
              @change="handleStartDateChange"
            />
          </div>

          <!-- End Date -->
          <div class="form-group filter-date-group">
            <label class="form-label">End Date</label>
            <input
              type="date"
              class="form-input date-input"
              :class="{ 'date-empty': !endDate }"
              v-model="endDate"
              :min="startDate"
              ref="endDateInput"
            />
          </div>

          <!-- Status -->
          <div class="form-group filter-select-group">
            <label class="form-label">Booking Status</label>
            <select class="form-select" v-model="selectedStatus">
              <option value="">All</option>
              <option value="CONFIRMED">SUCCESS</option>
              <option value="PENDING">PENDING</option>
              <option value="CANCELLED">CANCELLED</option>
              <option value="EXPIRED">EXPIRED</option>
            </select>
          </div>

          <!-- Clear Filters Link -->
          <div class="form-group clear-filters-wrap">
            <label class="form-label">&nbsp;</label>
            <div class="clear-filters-content">
              <button class="clear-filters-btn" @click="clearFilters">Clear Filters</button>
            </div>
          </div>
        </div>

        <!-- Stats Grid widgets -->
        <div class="stats-widgets-grid">
          <div class="glass-card stat-widget">
            <span class="widget-label">Total Revenue (Confirmed)</span>
            <h2 class="widget-val text-accent">{{ formatCurrency(totalRevenue) }}</h2>
            <div class="widget-sub">From completed transactions</div>
          </div>
          <div class="glass-card stat-widget">
            <span class="widget-label">Successful Bookings</span>
            <h2 class="widget-val text-primary">{{ totalConfirmedBookingsCount }} bookings</h2>
            <div class="widget-sub">Excluding expired / cancelled</div>
          </div>
          <div class="glass-card stat-widget">
            <span class="widget-label">Total System Bookings</span>
            <h2 class="widget-val">{{ totalBookings }} bookings</h2>
            <div class="widget-sub">Total bookings in active table view</div>
          </div>
        </div>

        <!-- Movie Revenue breakdown -->
        <div class="glass-card table-section stats-table-card">
          <h3 class="section-subtitle">Revenue Breakdown by Movie</h3>
          <div v-if="stats.length === 0" class="empty-table-msg">No movie statistics available</div>
          <div v-else class="table-responsive">
            <table class="data-table">
              <thead>
                <tr>
                  <th>Movie Title</th>
                  <th>Booking Count</th>
                  <th>Total Revenue</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in stats" :key="item._id">
                  <td class="font-bold">{{ item._id }}</td>
                  <td>{{ item.count }} bookings</td>
                  <td class="text-accent">{{ formatCurrency(item.totalRevenue) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- Bookings table -->
        <div class="glass-card table-section">
          <LoadingSpinner v-if="loading && bookings.length === 0" message="Loading bookings..." />
          <EmptyState
            v-else-if="error"
            title="An error occurred"
          />
          <EmptyState
            v-else-if="bookings.length === 0 && !loading"
            title="No bookings found"
          />
          <div v-else :style="{ opacity: loading ? 0.5 : 1, transition: 'opacity 0.2s', pointerEvents: loading ? 'none' : 'auto' }">
            <div class="table-responsive">
              <table class="data-table">
                <thead>
                  <tr>
                    <th>Booking ID</th>
                    <th>User (Email)</th>
                    <th>Movie</th>
                    <th>Theater</th>
                    <th>Seats</th>
                    <th>Showtime</th>
                    <th>Total Price</th>
                    <th>Status</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="b in bookings" :key="b.id">
                    <td class="font-mono text-sm">{{ formatBookingCode(b.id) }}</td>
                    <td>{{ b.userEmail }}</td>
                    <td class="font-bold">{{ b.movieTitle }}</td>
                    <td>{{ b.theaterName }}</td>
                    <td>
                      <span class="seats-list-badge">{{ formatSeatLabels(b.seats) }}</span>
                    </td>
                    <td>{{ formatDateTime(b.startTime, { day: 'numeric', month: 'short', year: '2-digit', hour: '2-digit', minute: '2-digit' }) }}</td>
                    <td class="text-accent">{{ formatCurrency(b.totalPrice) }}</td>
                    <td>
                      <span class="badge" :class="getBookingStatusPresentation(b.status, { confirmedLabel: 'SUCCESS' }).badgeClass">{{ getBookingStatusPresentation(b.status, { confirmedLabel: 'SUCCESS' }).label }}</span>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            <!-- Pagination -->
            <div class="pagination-section">
              <span class="record-count">
                {{ (page - 1) * limit + 1 }}–{{ Math.min(page * limit, totalBookings) }} of {{ totalBookings }}
              </span>
              <div class="pagination-right">
                <AppButton
                  variant="secondary"
                  size="sm"
                  :disabled="page === 1 || loading"
                  @click="page--"
                >
                  &lt;
                </AppButton>
                <span class="page-indicator">{{ page }} / {{ Math.ceil(totalBookings / limit) || 1 }}</span>
                <AppButton
                  variant="secondary"
                  size="sm"
                  :disabled="page * limit >= totalBookings || loading"
                  @click="page++"
                >
                  &gt;
                </AppButton>
                <label class="page-size-label">Rows:</label>
                <select class="page-size-select" v-model.number="limit" @change="handleLimitChange">
                  <option v-for="size in PAGE_SIZE_OPTIONS" :key="size" :value="size">{{ size }}</option>
                </select>
              </div>
            </div>

          </div>
        </div>
  </main>
</template>

<style scoped>
.admin-main {
  display: flex;
  flex-direction: column;
  gap: var(--space-lg);
}

.admin-header {
  border-bottom: 1px solid var(--border-subtle);
  padding-bottom: var(--space-sm);
}

/* Stats grid */
.stats-widgets-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: var(--space-sm);
}

.stat-widget {
  background: rgba(255,255,255,0.02);
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: var(--space-md) !important;
}
.stat-widget:hover {
  transform: none;
  box-shadow: var(--shadow-card);
}

.widget-label {
  font-size: 0.7rem;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.widget-val {
  font-size: 1.4rem;
  font-weight: 800;
  font-family: 'Outfit', sans-serif;
}

.widget-sub {
  font-size: 0.68rem;
  color: var(--text-muted);
}

.section-subtitle {
  font-size: 0.95rem;
  margin-bottom: var(--space-sm);
}

.stats-table-card {
  background: rgba(255,255,255,0.015);
  padding: var(--space-md) !important;
}
.stats-table-card:hover {
  transform: none;
  box-shadow: var(--shadow-card);
}

.empty-table-msg {
  text-align: center;
  color: var(--text-muted);
  padding: var(--space-sm);
}

.filter-toggle-row {
  display: none;
}

/* Filter bar */
.filter-bar {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-sm);
  background: rgba(255,255,255,0.02);
  align-items: flex-end;
  padding: var(--space-md) !important;
}

@media (min-width: 993px) {
  .admin-main {
    margin-left: 260px;
    max-width: calc(100vw - 260px);
  }
}

/* ── Tablet (≤992px) ── */
@media (max-width: 992px) {
  .admin-main {
    margin-left: 0;
    max-width: 100%;
    padding: var(--space-md);
  }
  .filter-toggle-row {
    display: flex;
    margin-bottom: var(--space-xs);
  }
  .filter-bar.mobile-hidden {
    display: none !important;
  }
  .filter-bar {
    gap: var(--space-xs);
  }
  .filter-bar .filter-date-group,
  .filter-bar .filter-select-group {
    flex: 1 1 140px;
  }
  .filter-bar .search-input-group {
    flex: 1 1 160px;
  }
  .stats-widgets-grid {
    grid-template-columns: repeat(3, 1fr);
  }
  .widget-val {
    font-size: 1.2rem;
  }
}

/* ── Mobile (≤640px) ── */
@media (max-width: 640px) {
  .admin-main {
    padding: var(--space-sm);
    gap: var(--space-md);
  }
  .stats-widgets-grid {
    grid-template-columns: 1fr;
  }
  .stat-widget {
    flex-direction: row;
    align-items: center;
    justify-content: space-between;
    padding: var(--space-sm) var(--space-md) !important;
  }
  .widget-val {
    font-size: 1.1rem;
  }
  .widget-sub {
    display: none;
  }
  .filter-bar .filter-date-group,
  .filter-bar .filter-select-group,
  .filter-bar .search-input-group {
    flex: 1 1 100%;
  }
  .clear-filters-wrap {
    flex: 1 1 100%;
  }
  .clear-filters-content {
    justify-content: flex-start;
    height: auto;
  }
  .table-section {
    padding: var(--space-sm) !important;
  }
  .pagination-section {
    justify-content: center;
    margin-top: var(--space-md);
  }
}
.filter-bar:hover {
  transform: none;
  box-shadow: var(--shadow-card);
}

.filter-bar .form-group {
  margin-bottom: 0;
}

.filter-bar .filter-date-group,
.filter-bar .filter-select-group {
  flex: 0 1 160px;
}

.filter-bar .search-input-group {
  flex: 1 1 180px;
}

.filter-bar .clear-filters-wrap {
  flex: 0 1 auto;
}

.clear-filters-wrap {
  display: flex;
  flex-direction: column;
}

.clear-filters-content {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 42px;
}

.clear-filters-btn {
  background: none;
  border: none;
  padding: 0;
  color: var(--text-muted);
  font-size: 0.85rem;
  font-family: 'Inter', sans-serif;
  text-decoration: underline;
  text-underline-offset: 3px;
  cursor: pointer;
  transition: color var(--transition-fast);
}

.clear-filters-btn:hover {
  color: var(--text-primary);
}

/* Date input — dim placeholder when empty */
.date-input.date-empty {
  color: var(--text-muted);
}
.date-input.date-empty::-webkit-calendar-picker-indicator {
  opacity: 0.35;
}
.date-input:not(.date-empty) {
  color: var(--text-primary);
}

/* Auto-complete suggestions */
.autocomplete-wrap {
  position: relative;
}

.suggestions-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  right: 0;
  z-index: 100;
  max-height: 200px;
  overflow-y: auto;
  padding: 6px 0;
  background: var(--color-bg-elevated);
  border: 1px solid var(--border-card);
  box-shadow: var(--shadow-lift);
}

.suggestion-item {
  padding: 10px 16px;
  font-size: 0.88rem;
  color: var(--text-secondary);
  cursor: pointer;
  text-align: left;
  transition: all var(--transition-fast);
}

.suggestion-item:hover {
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-primary);
}

.table-section {
  background: rgba(255,255,255,0.02);
  padding: var(--space-md) !important;
}
.table-section:hover {
  transform: none;
  box-shadow: var(--shadow-card);
}

.table-responsive {
  width: 100%;
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
}

.seats-list-badge {
  background: rgba(255, 255, 255, 0.08);
  padding: 2px 6px;
  border-radius: var(--radius-sm);
  font-family: monospace;
  font-size: 0.8rem;
  color: var(--text-primary);
}

.font-bold {
  font-weight: 600;
}

.text-danger {
  color: #ef4444;
}

/* Pagination */
.pagination-section {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--space-md);
  margin-top: var(--space-lg);
  padding-top: var(--space-md);
  border-top: 1px solid var(--border-subtle);
  flex-wrap: wrap;
}

.pagination-left {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
}

.pagination-right {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
}

.page-size-label {
  font-size: 0.78rem;
  color: var(--text-muted);
  white-space: nowrap;
}

.page-size-select {
  padding: 4px 8px;
  background: rgba(255,255,255,0.04);
  border: 1px solid var(--border-card);
  border-radius: var(--radius-sm);
  color: var(--text-primary);
  font-size: 0.82rem;
  font-family: 'Inter', sans-serif;
  cursor: pointer;
  outline: none;
  transition: border-color var(--transition-fast);
}
.page-size-select:focus {
  border-color: var(--color-primary);
}

.record-count {
  font-size: 0.78rem;
  color: var(--text-muted);
  white-space: nowrap;
}

.page-indicator {
  font-size: 0.85rem;
  color: var(--text-secondary);
}

@media (max-width: 640px) {
  .pagination-section {
    flex-direction: column;
    align-items: stretch;
    gap: var(--space-sm);
  }
  .pagination-left, .pagination-right {
    justify-content: center;
  }
}

</style>
