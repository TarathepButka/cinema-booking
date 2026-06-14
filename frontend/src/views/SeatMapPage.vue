<script setup lang="ts">
import { onMounted, onUnmounted, ref, computed } from 'vue'
import { useRoute, useRouter, onBeforeRouteLeave } from 'vue-router'
import { useShowtimeStore, type Seat } from '@/stores/showtime'
import { useBookingStore } from '@/stores/booking'
import { useAuthStore } from '@/stores/auth'
import { useWebSocket } from '@/composables/useWebSocket'
import SeatGrid from '@/components/SeatGrid.vue'
import BookingSteps from '@/components/BookingSteps.vue'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import EmptyState from '@/components/EmptyState.vue'
import ModalDialog from '@/components/ModalDialog.vue'
import AppButton from '@/components/AppButton.vue'
import { formatCurrency, formatTime } from '@/utils/format'

const route = useRoute()
const router = useRouter()
const showtimeStore = useShowtimeStore()
const bookingStore = useBookingStore()
const authStore = useAuthStore()

const showtimeId = route.params.id as string

const loading = computed(() => showtimeStore.loading)
const error = computed(() => showtimeStore.error)
const showtime = computed(() => showtimeStore.currentShowtime)
const seats = computed(() => showtime.value?.seats || [])
const myUserId = computed(() => authStore.user?.userId)

// Keep track of seats locked in backend by me
const myLockedSeats = ref<Array<{ seatLabel: string; expiresAt: string }>>([])
// Local selection state on frontend (does not trigger backend locks immediately)
const selectedSeats = ref<string[]>([])
const isReleasingAll = ref(false)
const bookingLoading = ref(false)
const showConflictModal = ref(false)
const conflictMessage = ref('')

// Toast message state
const toastMsg = ref<string | null>(null)
const toastType = ref<'success' | 'error' | 'info'>('info')
let toastTimeout: ReturnType<typeof setTimeout> | null = null

function showToast(msg: string, type: 'success' | 'error' | 'info' = 'info') {
  toastMsg.value = msg
  toastType.value = type
  if (toastTimeout) clearTimeout(toastTimeout)
  toastTimeout = setTimeout(() => {
    toastMsg.value = null
  }, 4000)
}

// Fetch showtime details
onMounted(async () => {
  await showtimeStore.fetchById(showtimeId)

  // Restore from pendingBooking in store if it matches this showtime
  if (bookingStore.pendingBooking && bookingStore.pendingBooking.showtimeId === showtimeId) {
    selectedSeats.value = [...bookingStore.pendingBooking.seatLabels]
    
    // Also sync myLockedSeats from pendingBooking info if locks were NOT released
    if (!bookingStore.pendingBooking.locksReleased) {
      bookingStore.pendingBooking.seats.forEach(s => {
        myLockedSeats.value.push({
          seatLabel: s.seatLabel,
          expiresAt: bookingStore.pendingBooking?.expiresAt || new Date(Date.now() + 300000).toISOString()
        })
      })
    }
  }
  
  // Find seats already locked by me from previous visits (if page refreshed)
  if (seats.value.length > 0 && myUserId.value) {
    seats.value.forEach((seat) => {
      if (seat.status === 'LOCKED' && seat.lockedBy === myUserId.value && seat.lockedUntil) {
        if (new Date(seat.lockedUntil).getTime() > Date.now()) {
          const alreadyAdded = myLockedSeats.value.some(s => s.seatLabel === seat.seatLabel)
          if (!alreadyAdded) {
            myLockedSeats.value.push({
              seatLabel: seat.seatLabel,
              expiresAt: seat.lockedUntil
            })
          }
          if (!selectedSeats.value.includes(seat.seatLabel)) {
            selectedSeats.value.push(seat.seatLabel)
          }
        }
      }
    })
  }
})

// Setup Real-time updates via WebSocket
useWebSocket(showtimeId, (msg) => {
  if (msg.type === 'SEAT_UPDATE' && msg.seatLabel && msg.status) {
    showtimeStore.updateSeatStatus(msg.seatLabel, msg.status as any)

    // If locked by someone else, and it was in my locked list, remove from locked list
    if (msg.status === 'LOCKED' && msg.updatedBy !== myUserId.value) {
      myLockedSeats.value = myLockedSeats.value.filter(s => s.seatLabel !== msg.seatLabel)
    }
    // If released by background cleanup or someone else, and was in my list, remove
    if (msg.status === 'AVAILABLE') {
      const isMine = myLockedSeats.value.some(s => s.seatLabel === msg.seatLabel)
      if (isMine) {
        myLockedSeats.value = myLockedSeats.value.filter(s => s.seatLabel !== msg.seatLabel)
        selectedSeats.value = selectedSeats.value.filter(s => s !== msg.seatLabel)
        showToast(`Seat ${msg.seatLabel} released due to timeout`, 'error')
      }
    }
  }
})

// Handle seat click
async function handleSeatClick(seat: Seat) {
  const isSelected = selectedSeats.value.includes(seat.seatLabel)
  const isLockedByMe = myLockedSeats.value.some(s => s.seatLabel === seat.seatLabel)

  if (isSelected) {
    if (isLockedByMe) {
      // Release seat in backend
      try {
        await bookingStore.releaseSeat(showtimeId, seat.seatLabel)
        myLockedSeats.value = myLockedSeats.value.filter(s => s.seatLabel !== seat.seatLabel)
        selectedSeats.value = selectedSeats.value.filter(s => s !== seat.seatLabel)
      } catch (e: any) {
        showToast(e.response?.data?.error || 'Could not release seat', 'error')
      }
    } else {
      // Just frontend selection, remove it
      selectedSeats.value = selectedSeats.value.filter(s => s !== seat.seatLabel)
    }
  } else {
    // Lock seat
    if (seat.status !== 'AVAILABLE') return
    selectedSeats.value.push(seat.seatLabel)
  }
}

// Compute total price and detail list
const selectedSeatDetails = computed(() => {
  return selectedSeats.value.map(label => {
    const seatInfo = seats.value.find(s => s.seatLabel === label)
    return {
      seatLabel: label,
      zone: seatInfo?.zone || 'FRONT',
      price: seatInfo?.price || 0
    }
  })
})

const totalPrice = computed(() => {
  return selectedSeatDetails.value.reduce((sum, s) => sum + s.price, 0)
})

const formattedTotalPrice = computed(() => {
  return formatCurrency(totalPrice.value)
})

// Proceed to confirm
async function proceedToConfirm() {
  if (selectedSeats.value.length === 0) return
  bookingLoading.value = true

  const newlyLockedSeats: Array<{ seatLabel: string; expiresAt: string }> = []

  try {
    // Lock all selected seats one by one
    for (const seatLabel of selectedSeats.value) {
      const existing = myLockedSeats.value.find(s => s.seatLabel === seatLabel)
      if (existing) {
        newlyLockedSeats.push(existing)
        continue
      }
      
      const lockRes = await bookingStore.lockSeat(showtimeId, seatLabel)
      newlyLockedSeats.push({
        seatLabel: seatLabel,
        expiresAt: lockRes.expiresAt
      })
    }

    // Calculate earliest expiry time
    const times = newlyLockedSeats.map(s => new Date(s.expiresAt).getTime())
    const earliestExpiryTime = new Date(Math.min(...times))

    bookingStore.setPendingBooking({
      showtimeId: showtimeId,
      movieId: showtime.value?.movieId || '',
      seatLabels: selectedSeats.value,
      movieTitle: showtime.value?.movie?.title || 'Movie',
      theaterName: showtime.value?.theaterName || 'Theater',
      seats: selectedSeatDetails.value,
      totalPrice: totalPrice.value,
      expiresAt: earliestExpiryTime.toISOString()
    })

    router.push('/payment')
  } catch (e: any) {
    // Release newly locked seats to avoid dangling locks
    for (const locked of newlyLockedSeats) {
      const isPreExisting = myLockedSeats.value.some(s => s.seatLabel === locked.seatLabel)
      if (!isPreExisting) {
        await bookingStore.releaseSeat(showtimeId, locked.seatLabel)
      }
    }
    
    const errorMsg = e.response?.data?.error || 'Could not reserve seats. Some seats may have already been reserved by another user.'
    conflictMessage.value = errorMsg
    showConflictModal.value = true
    
    // Refresh showtime
    await showtimeStore.fetchById(showtimeId)
    // Filter out unavailable seats
    selectedSeats.value = selectedSeats.value.filter(label => {
      const s = seats.value.find(seat => seat.seatLabel === label)
      return s && (s.status === 'AVAILABLE' || (s.status === 'LOCKED' && s.lockedBy === myUserId.value))
    })
  } finally {
    bookingLoading.value = false
  }
}

// Release all seats locked by me on exit
async function releaseAllMyLockedSeats() {
  if (selectedSeats.value.length === 0) return
  isReleasingAll.value = true
  
  if (myLockedSeats.value.length > 0) {
    const releasePromises = myLockedSeats.value.map((s) => {
      return bookingStore.releaseSeat(showtimeId, s.seatLabel)
    })
    await Promise.all(releasePromises)
  }
  
  myLockedSeats.value = []
  selectedSeats.value = []
  isReleasingAll.value = false
}

// Navigation guard to release seats when navigating away
onBeforeRouteLeave((to, from, next) => {
  if (to.path === '/payment') {
    next()
  } else {
    releaseAllMyLockedSeats().then(() => next())
  }
})

// Safety fallback cleanup in case component is destroyed
onUnmounted(() => {
  if (toastTimeout) clearTimeout(toastTimeout)
})
</script>

<template>
  <div class="booking-flow-page">
    <BookingSteps :current-step="2" :back-url="`/movies/${showtime?.movieId}/showtime`" v-if="showtime" />
    <div class="container page-content">

      <!-- Toast -->
      <div class="toast-container" v-if="toastMsg">
        <div class="toast" :class="`toast-${toastType}`">
          <span>{{ toastMsg }}</span>
        </div>
      </div>

      <!-- Seat Conflict / Unavailable Modal -->
      <ModalDialog :isOpen="showConflictModal" maxWidth="450px" @close="showConflictModal = false">
        <template #header>
          <div style="display: flex; align-items: center; gap: var(--space-md);">
            <span class="modal-icon">⚠️</span>
            <h3 class="modal-title" style="margin: 0;">Seat Unavailable</h3>
          </div>
        </template>
        <p class="modal-desc">{{ conflictMessage }}</p>
        <p class="modal-help">The seat map has been refreshed. Please select another available seat.</p>
        <template #footer>
          <AppButton @click="showConflictModal = false">OK</AppButton>
        </template>
      </ModalDialog>

      <LoadingSpinner v-slot:default v-if="loading && seats.length === 0" message="Loading seat map..." />

      <EmptyState
        v-else-if="error"
        title="Showtime Not Found"
      >
        <RouterLink to="/" class="btn btn-primary mt-4">Back to Home</RouterLink>
      </EmptyState>

      <div v-else-if="showtime" class="booking-grid animate-fade-up">
        <!-- Left: Seat Map -->
        <div class="seatmap-card">
          <!-- Header -->
          <div class="seatmap-header">
            <div class="showtime-info">
              <h2 class="showtime-movie-title">{{ showtime.movie?.title }}</h2>
              <p class="showtime-meta">{{ showtime.theaterName }} &nbsp;·&nbsp; {{ formatTime(showtime.startTime) }} &nbsp;·&nbsp; {{ showtime.slotLabel }}</p>
            </div>
          </div>

          <!-- Curved Screen Visual -->
          <div class="screen-arc-wrap">
            <div class="screen-arc">
              <span class="screen-label">SCREEN</span>
            </div>
          </div>

          <SeatGrid
            :seats="seats"
            :my-user-id="myUserId"
            :selected-seats="selectedSeats"
            @seat-click="handleSeatClick"
          />
        </div>

        <!-- Right: Summary Panel -->
        <div class="panel-sticky">
          <div class="summary-panel" :class="{ 'panel-has-seats': selectedSeats.length > 0 }">
            <div class="panel-header-row">
              <span class="panel-heading">Order Summary</span>
              <span v-if="selectedSeats.length > 0" class="seat-count-badge">{{ selectedSeats.length }} seats</span>
            </div>

            <!-- Empty state -->
            <div v-if="selectedSeats.length === 0" class="panel-empty">
              <div class="panel-empty-icon">🎟️</div>
              <p>Select seats from the map to start booking</p>
            </div>

            <!-- Has seats -->
            <div v-else class="panel-details">
              <div class="seats-list">
                <div v-for="s in selectedSeatDetails" :key="s.seatLabel" class="seat-row">
                  <div class="seat-info">
                    <span class="zone-dot" :style="{ background: s.zone === 'FRONT' ? '#ef4444' : s.zone === 'MIDDLE' ? '#f59e0b' : '#22c55e' }"></span>
                    <span class="seat-lbl">{{ s.seatLabel }}</span>
                    <span class="seat-zone-tag">{{ s.zone }}</span>
                  </div>
                  <span class="seat-p">฿{{ s.price }}</span>
                </div>
              </div>

              <div class="divider"></div>

              <div class="total-bar">
                <span class="total-lbl">Total</span>
                <span class="total-amount">{{ formattedTotalPrice }}</span>
              </div>

              <AppButton
                variant="primary"
                size="lg"
                class="w-full"
                style="margin-top: 8px"
                :loading="bookingLoading"
                :disabled="selectedSeats.length === 0"
                @click="proceedToConfirm"
              >
                Proceed
              </AppButton>
              <AppButton
                variant="ghost"
                size="sm"
                class="w-full"
                style="margin-top: 6px; color: var(--text-muted)"
                :disabled="bookingLoading || isReleasingAll"
                @click="releaseAllMyLockedSeats"
              >
                Deselect All
              </AppButton>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* ── Screen arc ── */
.screen-arc-wrap {
  display: flex;
  justify-content: center;
  padding: 0 var(--space-xl);
  margin-bottom: var(--space-lg);
  overflow: hidden;
}

.screen-arc {
  width: 80%;
  height: 50px;
  border-top: 3px solid rgba(229, 9, 20, 0.5);
  border-radius: 50% 50% 0 0 / 30px 30px 0 0;
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding-top: 6px;
  box-shadow: 0 -10px 30px rgba(229, 9, 20, 0.12);
  position: relative;
}

.screen-arc::before {
  content: '';
  position: absolute;
  top: -3px;
  left: 0;
  right: 0;
  height: 3px;
  background: linear-gradient(90deg, transparent, var(--color-primary), transparent);
  border-radius: var(--radius-full);
}

.screen-label {
  font-family: 'Barlow Condensed', sans-serif;
  font-size: 0.65rem;
  font-weight: 800;
  letter-spacing: 0.3em;
  color: rgba(229, 9, 20, 0.6);
  text-transform: uppercase;
}

.booking-grid {
  display: grid;
  grid-template-columns: 1fr 360px;
  gap: var(--space-xl);
  align-items: start;
}

.seatmap-card {
  background: var(--color-bg-card);
  border: 1px solid var(--border-card);
  border-radius: var(--radius-lg);
  padding: var(--space-xl) var(--space-lg);
  padding-top: var(--space-lg);
}

.seatmap-header {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-bottom: var(--space-lg);
  border-bottom: 1px solid var(--border-subtle);
  padding-bottom: var(--space-md);
}

.btn-back-sm {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: var(--text-muted);
  font-size: 0.82rem;
  font-weight: 600;
  text-decoration: none;
  padding: 5px 12px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-full);
  transition: all var(--transition-fast);
  width: fit-content;
}
.btn-back-sm:hover {
  color: var(--text-primary);
  border-color: var(--border-card);
  background: rgba(255,255,255,0.04);
}

.showtime-movie-title {
  font-family: 'Barlow Condensed', sans-serif;
  font-size: 1.6rem;
  font-weight: 800;
  letter-spacing: 0.03em;
  color: var(--text-primary);
}

.showtime-meta {
  font-size: 0.85rem;
  color: var(--text-muted);
  margin-top: 2px;
}

/* Panel */
.panel-sticky {
  position: sticky;
  top: calc(64px + var(--space-xl));
}

.summary-panel {
  background: var(--color-bg-card);
  border: 1px solid var(--border-card);
  border-radius: var(--radius-lg);
  padding: var(--space-lg);
  transition: all var(--transition-base);
}

.summary-panel.panel-has-seats {
  border-color: rgba(229, 9, 20, 0.25);
  box-shadow: 0 4px 30px rgba(229, 9, 20, 0.08);
}

.panel-header-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-lg);
}

.panel-heading {
  font-family: 'Barlow Condensed', sans-serif;
  font-size: 1.2rem;
  font-weight: 700;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--text-primary);
}

.seat-count-badge {
  background: rgba(229, 9, 20, 0.12);
  color: var(--color-primary);
  border: 1px solid rgba(229, 9, 20, 0.2);
  padding: 3px 10px;
  border-radius: var(--radius-full);
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.04em;
}

.panel-empty {
  text-align: center;
  padding: var(--space-2xl) 0;
  color: var(--text-muted);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  font-size: 0.88rem;
}

.panel-empty-icon {
  font-size: 2.5rem;
  opacity: 0.4;
}

.panel-details {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.timer-wrap {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.timer-note {
  font-size: 0.7rem;
  color: var(--text-muted);
  line-height: 1.4;
  text-align: center;
}

.seats-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 180px;
  overflow-y: auto;
  padding-right: 4px;
}

.seat-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.88rem;
}

.seat-info {
  display: flex;
  align-items: center;
  gap: 7px;
}

.zone-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  flex-shrink: 0;
}

.seat-lbl {
  font-family: 'Outfit', sans-serif;
  font-weight: 700;
  color: var(--text-primary);
  font-size: 0.9rem;
}

.seat-zone-tag {
  font-size: 0.68rem;
  color: var(--text-muted);
  background: rgba(255,255,255,0.05);
  padding: 1px 6px;
  border-radius: var(--radius-full);
}

.seat-p {
  font-family: 'Barlow Condensed', sans-serif;
  font-size: 1rem;
  font-weight: 700;
  color: var(--text-secondary);
}

.total-bar {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
}

.total-lbl {
  font-size: 0.78rem;
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.06em;
}

.total-amount {
  font-family: 'Barlow Condensed', sans-serif;
  font-size: 2rem;
  font-weight: 800;
  color: var(--text-primary);
  letter-spacing: 0.02em;
}

.w-full { width: 100%; }
.mt-4 { margin-top: var(--space-md); }

@media (max-width: 900px) {
  .booking-grid { grid-template-columns: 1fr; }
  .panel-sticky { position: static; margin-top: var(--space-md); }
}

/* ── Modal overlay ── */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.75);
  backdrop-filter: blur(8px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  animation: fade-in 0.25s ease-out;
}

.modal-content {
  width: 90%;
  max-width: 450px;
  background: var(--color-bg-card);
  border: 1px solid rgba(239, 68, 68, 0.3);
  border-radius: var(--radius-lg);
  padding: var(--space-xl);
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.5), 0 0 20px rgba(239, 68, 68, 0.1);
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
}

.modal-header {
  display: flex;
  align-items: center;
  gap: var(--space-md);
  border-bottom: 1px solid var(--border-subtle);
  padding-bottom: var(--space-sm);
}

.modal-icon {
  font-size: 2rem;
  color: #ef4444;
}

.modal-title {
  font-family: 'Outfit', sans-serif;
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--text-primary);
}

.modal-body {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.modal-desc {
  font-size: 0.95rem;
  color: #f87171;
  font-weight: 600;
  line-height: 1.5;
}

.modal-help {
  font-size: 0.85rem;
  color: var(--text-secondary);
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  margin-top: var(--space-sm);
}

/* Animations */
@keyframes fade-in {
  from { opacity: 0; }
  to { opacity: 1; }
}

@keyframes scale-up {
  from { transform: scale(0.95); opacity: 0; }
  to { transform: scale(1); opacity: 1; }
}

.animate-scale-up {
  animation: scale-up 0.25s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}
</style>
