<script setup lang="ts">
import { onMounted, nextTick, ref, watch } from 'vue'
import { auditApi } from '@/api'
import { PAGINATION } from '@/constants'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import EmptyState from '@/components/EmptyState.vue'
import AppButton from '@/components/AppButton.vue'
import { formatBookingCode, formatDateTime } from '@/utils/format'
import { getAuditEventBadgeClass } from '@/utils/status'

// State
const showFilters = ref(false)
const logs = ref<any[]>([])
const totalLogs = ref(0)
const loading = ref(false)
const error = ref<string | null>(null)

// Filters
const selectedEvent = ref('')
const page = ref(PAGINATION.DEFAULT_PAGE)
const limit = ref(PAGINATION.DEFAULT_LIMIT)
const PAGE_SIZE_OPTIONS = PAGINATION.PAGE_SIZE_OPTIONS

async function fetchAuditLogs(scrollTop = false) {
  loading.value = true
  error.value = null
  try {
    const params: Record<string, any> = {
      page: page.value,
      limit: limit.value,
    }
    if (selectedEvent.value) params.event = selectedEvent.value

    const res = await auditApi.getLogs(params)
    logs.value = res.data.data || []
    totalLogs.value = res.data.total || 0
    if (scrollTop) {
      await nextTick()
      window.scrollTo({ top: 0, behavior: 'instant' })
    }
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to fetch audit logs'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchAuditLogs()
})

watch([page, selectedEvent, limit], () => {
  fetchAuditLogs(true)
})

function handleLimitChange() {
  page.value = 1
}

const isoDateRegex = /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}/

const formatValue = (val: any) => {
  if (typeof val === 'string' && isoDateRegex.test(val)) {
    return formatDateTime(val, {
      day: 'numeric',
      month: 'short',
      year: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    })
  }
  return typeof val === 'string' ? val : JSON.stringify(val)
}

const formatDetails = (details: Record<string, any>) => {
  if (!details) return '-'
  return Object.entries(details)
    .map(([key, val]) => `${key}: ${formatValue(val)}`)
    .join(', ')
}
</script>

<template>
  <main class="admin-main animate-fade-up">
        <div class="admin-header">
          <h1 class="page-title">System Audit Logs</h1>
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
          <div class="form-group filter-select-group">
            <label class="form-label">Filter by Event Type</label>
            <select class="form-select" v-model="selectedEvent">
              <option value="">All Events</option>
              <option value="BOOKING_SUCCESS">BOOKING_SUCCESS</option>
              <option value="BOOKING_TIMEOUT">BOOKING_TIMEOUT</option>
              <option value="SEAT_LOCKED">SEAT_LOCKED</option>
              <option value="SEAT_RELEASED">SEAT_RELEASED</option>
              <option value="SYSTEM_ERROR">SYSTEM_ERROR</option>
            </select>
          </div>
        </div>

        <!-- Logs Table -->
        <div class="glass-card table-section">
          <LoadingSpinner v-slot:default v-if="loading && logs.length === 0" message="Loading logs..." />
          <EmptyState
            v-else-if="error"
            title="An error occurred"
          />
          <EmptyState
            v-else-if="logs.length === 0 && !loading"
            title="No events found"
          />
          <div v-else :style="{ opacity: loading ? 0.5 : 1, transition: 'opacity 0.2s', pointerEvents: loading ? 'none' : 'auto' }">
            <div class="table-responsive">
              <table class="data-table">
                <thead>
                  <tr>
                    <th>Date & Time</th>
                    <th>Event</th>
                    <th>User (Email)</th>
                    <th>Showtime ID</th>
                    <th>Seats</th>
                    <th>Booking ID</th>
                    <th>Additional Details</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="log in logs" :key="log.id">
                    <td class="timestamp-col">{{ formatDateTime(log.createdAt, { day: 'numeric', month: 'short', year: '2-digit', hour: '2-digit', minute: '2-digit', second: '2-digit' }) }}</td>
                    <td>
                      <span class="badge" :class="getAuditEventBadgeClass(log.event)">{{ log.event }}</span>
                    </td>
                    <td>
                      <span v-if="log.userEmail" class="email-text">{{ log.userEmail }}</span>
                      <span v-else class="system-tag">SYSTEM</span>
                    </td>
                    <td class="font-mono text-sm">{{ log.showtimeId ? formatBookingCode(log.showtimeId) : '-' }}</td>
                    <td>
                      <span v-if="log.seatLabels && log.seatLabels.length" class="seats-list-badge">
                        {{ log.seatLabels.join(', ') }}
                      </span>
                      <span v-else>-</span>
                    </td>
                    <td class="font-mono text-sm text-accent">
                      {{ log.bookingId ? formatBookingCode(log.bookingId) : '-' }}
                    </td>
                    <td class="details-col" :title="JSON.stringify(log.details)">
                      {{ formatDetails(log.details) }}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            <!-- Pagination -->
            <div class="pagination-section">
              <span class="record-count">
                {{ (page - 1) * limit + 1 }}–{{ Math.min(page * limit, totalLogs) }} of {{ totalLogs }}
              </span>
              <div class="pagination-right">
                <AppButton
                  variant="secondary"
                  size="sm"
                  :disabled="page === 1 || loading"
                  @click="page--"
                >
                  Previous
                </AppButton>
                <span class="page-indicator">{{ page }} / {{ Math.ceil(totalLogs / limit) || 1 }}</span>
                <AppButton
                  variant="secondary"
                  size="sm"
                  :disabled="page * limit >= totalLogs || loading"
                  @click="page++"
                >
                  Next
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
  gap: var(--space-xl);
}

.admin-header {
  border-bottom: 1px solid var(--border-subtle);
  padding-bottom: var(--space-md);
}

.filter-toggle-row {
  display: none;
}

/* Filter bar */
.filter-bar {
  display: flex;
  background: rgba(255,255,255,0.02);
}

@media (max-width: 992px) {
  .filter-toggle-row {
    display: flex;
    justify-content: flex-end;
    margin-bottom: var(--space-sm);
  }
  .filter-bar.mobile-hidden {
    display: none !important;
  }
  .filter-select-group {
    width: 100% !important;
  }
}
.filter-bar:hover {
  transform: none;
  box-shadow: var(--shadow-card);
}

.filter-select-group {
  margin-bottom: 0;
  width: 320px;
}

.table-section {
  background: rgba(255,255,255,0.02);
}
.table-section:hover {
  transform: none;
  box-shadow: var(--shadow-card);
}

.table-responsive {
  width: 100%;
  overflow-x: auto;
}

.timestamp-col {
  white-space: nowrap;
  font-size: 0.8rem;
}

.email-text {
  color: var(--text-primary);
}

.system-tag {
  color: var(--text-muted);
  font-size: 0.72rem;
  letter-spacing: 0.05em;
  font-weight: 700;
  background: rgba(255, 255, 255, 0.04);
  padding: 2px 6px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-subtle);
}

.seats-list-badge {
  background: rgba(255, 255, 255, 0.08);
  padding: 2px 6px;
  border-radius: var(--radius-sm);
  font-family: monospace;
  font-size: 0.8rem;
  color: var(--text-primary);
}

.details-col {
  max-width: 250px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  font-size: 0.8rem;
  color: var(--text-muted);
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
