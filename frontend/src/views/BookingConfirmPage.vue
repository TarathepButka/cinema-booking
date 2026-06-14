<script setup lang="ts">
import { onMounted, ref, computed, watch } from 'vue'
import { useRouter, RouterLink, onBeforeRouteLeave } from 'vue-router'
import { useBookingStore } from '@/stores/booking'
import { useShowtimeStore } from '@/stores/showtime'
import { useAuthStore } from '@/stores/auth'
import { useCountdown } from '@/composables/useCountdown'
import CountdownTimer from '@/components/CountdownTimer.vue'
import BookingSteps from '@/components/BookingSteps.vue'
import ModalDialog from '@/components/ModalDialog.vue'
import AppButton from '@/components/AppButton.vue'
import { formatBookingCode, formatCurrency, formatSeatLabels } from '@/utils/format'

const router = useRouter()
const bookingStore = useBookingStore()
const showtimeStore = useShowtimeStore()
const authStore = useAuthStore()

const pending = computed(() => bookingStore.pendingBooking)
const loading = computed(() => bookingStore.loading)
const error = computed(() => bookingStore.error)

const successBooking = ref<any>(null)
const confirmError = ref<string | null>(null)
const validatingReservation = ref(true)
const showInvalidReservationModal = ref(false)
const invalidReservationMessage = ref('')
let recoveringInvalidReservation = false

// Countdown timer setup
const countdown = useCountdown(300)

onMounted(async () => {
  if (!pending.value) {
    validatingReservation.value = false
    router.replace('/')
    return
  }

  const currentPending = pending.value
  const showtime = await showtimeStore.fetchById(currentPending.showtimeId)
  if (!showtime) {
    validatingReservation.value = false
    invalidReservationMessage.value = 'This showtime is no longer available. Please select another showtime.'
    showInvalidReservationModal.value = true
    return
  }

  const now = Date.now()
  const userId = authStore.user?.userId
  const lockedSeats = currentPending.seatLabels.map((label) =>
    showtime.seats?.find((seat) => seat.seatLabel === label),
  )
  const locksAreValid = lockedSeats.every((seat) =>
    seat?.status === 'LOCKED' &&
    seat.lockedBy === userId &&
    Boolean(seat.lockedUntil) &&
    new Date(seat.lockedUntil as string).getTime() > now,
  )

  if (!locksAreValid) {
    validatingReservation.value = false
    invalidReservationMessage.value = 'Your seat reservation has expired or is no longer valid. Please select your seats again.'
    showInvalidReservationModal.value = true
    return
  }

  const lockExpiryTimes = lockedSeats.map((seat) => new Date(seat!.lockedUntil as string).getTime())
  const pendingExpiry = currentPending.expiresAt
    ? new Date(currentPending.expiresAt).getTime()
    : Number.POSITIVE_INFINITY
  const expiry = new Date(Math.min(pendingExpiry, ...lockExpiryTimes))
  countdown.start(expiry)
  validatingReservation.value = false
})

async function recoverFromInvalidReservation() {
  const currentPending = pending.value
  recoveringInvalidReservation = true
  showInvalidReservationModal.value = false
  bookingStore.setPendingBooking(null)

  if (currentPending?.movieId) {
    await router.replace(`/movies/${currentPending.movieId}/showtime`)
  } else {
    await router.replace('/')
  }
}

// Auto-handle expiry
const watchExpired = computed(() => {
  if (countdown.isExpired.value && !successBooking.value) {
    confirmError.value = 'Your reservation has expired. Please select showtime and seats again.'
  }
  return countdown.isExpired.value
})

const showTimeoutModal = ref(false)

watch(
  () => countdown.isExpired.value,
  (expired) => {
    if (expired && !successBooking.value) {
      showTimeoutModal.value = true
    }
  }
)

async function handleTimeoutRedirect() {
  showTimeoutModal.value = false
  const currentPending = pending.value
  if (currentPending) {
    const showtimeId = currentPending.showtimeId
    const movieId = currentPending.movieId
    
    // Explicitly release locked seats in backend on session timeout
    const releasePromises = currentPending.seatLabels.map((label: string) => {
      return bookingStore.releaseSeatOnly(showtimeId, label)
    })
    await Promise.all(releasePromises)
    
    bookingStore.setPendingBooking(null)
    router.replace(`/movies/${movieId}/showtime/${showtimeId}/select-seat`)
  } else {
    bookingStore.setPendingBooking(null)
    router.replace('/')
  }
}

const formattedTotalPrice = computed(() => {
  if (!pending.value) return ''
  return formatCurrency(pending.value.totalPrice)
})

const seatLabelsJoined = computed(() => {
  if (!pending.value) return ''
  return formatSeatLabels(pending.value.seatLabels)
})

async function handleConfirm() {
  if (!pending.value || countdown.isExpired.value) return

  confirmError.value = null
  try {
    const booking = await bookingStore.confirmBooking(
      pending.value.showtimeId,
      pending.value.seatLabels,
    )
    successBooking.value = booking
    countdown.stop()
  } catch (err: any) {
    confirmError.value = err.response?.data?.error || 'Confirmation failed. Please try again.'
  }
}

// Release seats if the user navigates away (unless they successfully booked or went back to seat selection)
onBeforeRouteLeave((to, from, next) => {
  if (recoveringInvalidReservation) {
    next()
    return
  }

  const currentPending = pending.value
  if (successBooking.value) {
    next()
  } else if (currentPending && to.path === `/movies/${currentPending.movieId}/showtime/${currentPending.showtimeId}/select-seat`) {
    // Release locked seats in backend when going back to the seat selection page
    const releasePromises = currentPending.seatLabels.map((label: string) => {
      return bookingStore.releaseSeatOnly(currentPending.showtimeId, label)
    })
    Promise.all(releasePromises).then(() => {
      bookingStore.setPendingBooking({
        ...currentPending,
        locksReleased: true
      })
      next()
    })
  } else {
    if (currentPending) {
      const releasePromises = currentPending.seatLabels.map((label: string) => {
        return bookingStore.releaseSeat(currentPending.showtimeId, label)
      })
      Promise.all(releasePromises).then(() => {
        bookingStore.setPendingBooking(null)
        next()
      })
    } else {
      next()
    }
  }
})
</script>

<template>
  <div class="booking-flow-page">
    <BookingSteps :current-step="3" :back-url="`/movies/${pending?.movieId}/showtime/${pending?.showtimeId}/select-seat`" v-if="pending" />
    <div class="container page-content">
      <!-- Watcher dummy evaluation -->
      <span v-show="false">{{ watchExpired }}</span>

      <ModalDialog :isOpen="showInvalidReservationModal" maxWidth="450px" @close="recoverFromInvalidReservation">
        <template #header>
          <div style="display: flex; align-items: center; gap: var(--space-md);">
            <span class="modal-icon">⚠️</span>
            <h3 class="modal-title" style="margin: 0;">Reservation Unavailable</h3>
          </div>
        </template>
        <p class="modal-desc">{{ invalidReservationMessage }}</p>
        <template #footer>
          <AppButton @click="recoverFromInvalidReservation">Select Another Showtime</AppButton>
        </template>
      </ModalDialog>

      <!-- Timeout Expiry Modal -->
      <ModalDialog :isOpen="showTimeoutModal" maxWidth="450px" @close="handleTimeoutRedirect">
        <template #header>
          <div style="display: flex; align-items: center; gap: var(--space-md);">
            <span class="modal-icon">⚠️</span>
            <h3 class="modal-title" style="margin: 0;">Session Expired</h3>
          </div>
        </template>
        <p class="modal-desc">Sorry, your seat reservation session has expired.</p>
        <p class="modal-help">Please click OK to go back and select showtime and seats again.</p>
        <template #footer>
          <AppButton @click="handleTimeoutRedirect">OK</AppButton>
        </template>
      </ModalDialog>

      <LoadingSpinner
        v-if="validatingReservation"
        message="Validating your reservation..."
      />

      <div v-else-if="successBooking" class="success-container animate-fade-up">
        <div class="success-card glass-card">
          <div class="success-icon-wrap">
            <div class="success-icon">✓</div>
          </div>
          <h1 class="success-title">Booking Successful!</h1>
          <p class="success-desc">We have sent your booking code and ticket details to your email.</p>

          <!-- Ticket Receipt representation -->
          <div class="receipt-ticket">
            <div class="ticket-top">
              <span class="ticket-brand">CINEMA TICKET</span>
              <div class="ticket-status badge badge-success">PAID</div>
            </div>
            <div class="ticket-middle">
              <h3 class="ticket-movie">{{ successBooking.movieTitle }}</h3>
              <div class="ticket-grid">
                <div class="ticket-info-item">
                  <span class="ticket-label">Theater</span>
                  <span class="ticket-val text-accent">{{ successBooking.theaterName }}</span>
                </div>
                <div class="ticket-info-item">
                  <span class="ticket-label">Seats</span>
                  <span class="ticket-val font-mono">{{ formatSeatLabels(successBooking.seats) }}</span>
                </div>
                <div class="ticket-info-item">
                  <span class="ticket-label">Date & Time</span>
                  <span class="ticket-val">{{ new Date(successBooking.startTime).toLocaleString('en-US', { dateStyle: 'medium', timeStyle: 'short' }) }}</span>
                </div>
                <div class="ticket-info-item">
                  <span class="ticket-label">Booking ID</span>
                  <span class="ticket-val booking-code">{{ formatBookingCode(successBooking.id) }}</span>
                </div>
              </div>
            </div>
            <div class="ticket-border-dashed">
              <div class="notch left-notch"></div>
              <div class="notch right-notch"></div>
            </div>
            <div class="ticket-bottom">
              <div class="ticket-price-row">
                <span>Total Paid</span>
                <span class="ticket-price-val">฿{{ successBooking.totalPrice }}</span>
              </div>
            </div>
          </div>

          <div class="success-actions">
            <RouterLink to="/my-bookings" class="btn btn-primary btn-lg">View Booking History</RouterLink>
            <RouterLink to="/" class="btn btn-secondary">Back to Home</RouterLink>
          </div>
        </div>
      </div>

      <div v-else-if="pending" class="confirm-layout animate-fade-up">
        <div class="confirm-card glass-card">
          <h1 class="page-title text-center">Payment & Confirmation</h1>
          <p class="subtitle text-center">Please verify your booking details and confirm payment to complete.</p>

          <!-- Display timer -->
          <div class="timer-wrapper">
            <CountdownTimer
              :expires-at="pending.expiresAt"
              :remaining-seconds="countdown.remaining.value"
              :urgency="countdown.urgency.value"
            />
            <p class="timer-alert" v-if="countdown.urgency.value === 'critical'">⚠️ Please confirm quickly! Your reserved seats are about to release.</p>
          </div>

          <!-- Confirm Error -->
          <div class="error-banner" v-if="confirmError">
            <span class="error-icon">⚠️</span>
            <p class="error-msg">{{ confirmError }}</p>
          </div>

          <!-- Decorative Movie Ticket Layout -->
          <div class="ticket-display">
            <div class="movie-banner">
              <div class="ticket-label-tag">Ticket Summary</div>
              <h2 class="movie-name">{{ pending.movieTitle }}</h2>
              <div class="theater-name-tag">{{ pending.theaterName }}</div>
            </div>

            <div class="ticket-info-details">
              <div class="info-row">
                <span class="info-label">All Seats</span>
                <span class="info-val text-accent font-mono">{{ seatLabelsJoined }}</span>
              </div>
              <div class="info-row">
                <span class="info-label">Seat Count</span>
                <span class="info-val">{{ pending.seatLabels.length }} seats</span>
              </div>
              <div class="info-row">
                <span class="info-label">Total Price</span>
                <span class="info-val font-mono">{{ formattedTotalPrice }}</span>
              </div>
            </div>

            <div class="ticket-border-dashed">
              <div class="notch left-notch"></div>
              <div class="notch right-notch"></div>
            </div>

            <div class="ticket-payment-section">
              <div class="payment-method">
                <div class="payment-title">Payment Method</div>
                <div class="payment-option">
                  <span class="payment-icon">💳</span>
                  <div class="payment-desc">
                    <strong>Credit Card / Debit Card (Mock)</strong>
                    <span>Simulated payment system. Payment will succeed instantly.</span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Form Actions -->
          <div class="confirm-actions">
            <AppButton
              variant="primary"
              size="lg"
              class="w-full pay-btn"
              :loading="loading"
              :disabled="validatingReservation || countdown.isExpired.value"
              @click="handleConfirm"
            >
              Confirm & Pay {{ formattedTotalPrice }}
            </AppButton>
          </div>

        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.success-container, .confirm-layout {
  max-width: 580px;
  margin: 0 auto;
}

.success-card {
  text-align: center;
  padding: var(--space-2xl) var(--space-xl);
  background: rgba(255, 255, 255, 0.03);
}

.success-icon-wrap {
  width: 70px;
  height: 70px;
  background: rgba(34, 197, 94, 0.15);
  border: 2px solid var(--seat-available);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto var(--space-lg);
}

.success-icon {
  font-size: 2.2rem;
  color: var(--seat-available);
  font-weight: bold;
}

.success-title {
  font-size: 2rem;
  font-weight: 800;
  margin-bottom: var(--space-sm);
  background: linear-gradient(135deg, #fff 40%, var(--seat-available) 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.success-desc {
  color: var(--text-secondary);
  font-size: 0.95rem;
  margin-bottom: var(--space-xl);
}

.receipt-ticket {
  background: rgba(10, 10, 20, 0.6);
  border: 1px solid var(--border-card);
  border-radius: var(--radius-md);
  margin-bottom: var(--space-xl);
  text-align: left;
}

.ticket-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  background: rgba(255,255,255,0.02);
  border-bottom: 1px solid var(--border-subtle);
}

.ticket-brand {
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.15em;
  color: var(--text-muted);
}

.ticket-middle {
  padding: var(--space-md) var(--space-lg);
}

.ticket-movie {
  font-size: 1.35rem;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: var(--space-md);
}

.ticket-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.ticket-info-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.ticket-label {
  font-size: 0.7rem;
  color: var(--text-muted);
  text-transform: uppercase;
}

.ticket-val {
  font-size: 0.9rem;
  color: var(--text-primary);
  font-weight: 600;
}

.booking-code {
  color: var(--color-accent);
  letter-spacing: 0.05em;
}

.ticket-border-dashed {
  position: relative;
  height: 2px;
  border-top: 2px dashed var(--border-card);
  margin: var(--space-sm) 0;
}

.notch {
  position: absolute;
  top: -10px;
  width: 18px;
  height: 18px;
  background: var(--color-bg);
  border-radius: 50%;
  border: 1px solid var(--border-card);
  z-index: 10;
}

.left-notch {
  left: -10px;
}

.right-notch {
  right: -10px;
}

.ticket-bottom {
  padding: 16px var(--space-lg);
  background: rgba(255,255,255,0.01);
}

.ticket-price-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 700;
}

.ticket-price-row span:first-child {
  color: var(--text-secondary);
}

.ticket-price-val {
  font-size: 1.4rem;
  color: var(--color-accent);
}

.success-actions {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.confirm-card {
  padding: var(--space-xl);
}

.subtitle {
  color: var(--text-secondary);
  font-size: 0.95rem;
  margin-bottom: var(--space-lg);
}

.timer-wrapper {
  margin-bottom: var(--space-lg);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.timer-alert {
  font-size: 0.75rem;
  color: var(--seat-locked);
  font-weight: 600;
  animation: pulse-red 1.5s infinite;
}

@keyframes pulse-red {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.6; }
}

.error-banner {
  background: rgba(239, 68, 68, 0.15);
  border: 1px solid rgba(239, 68, 68, 0.3);
  border-radius: var(--radius-md);
  padding: 12px 16px;
  margin-bottom: var(--space-lg);
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.error-icon {
  font-size: 1.2rem;
  color: #ef4444;
}

.error-msg {
  font-size: 0.88rem;
  color: #f87171;
  line-height: 1.4;
  margin: 0;
}

.ticket-display {
  background: rgba(255,255,255,0.015);
  border: 1px solid var(--border-card);
  border-radius: var(--radius-md);
  margin-bottom: var(--space-xl);
  overflow: hidden;
}

.movie-banner {
  background: linear-gradient(135deg, rgba(229, 9, 20, 0.15) 0%, rgba(10, 10, 20, 0.6) 100%);
  padding: 20px;
  border-bottom: 1px solid var(--border-subtle);
  position: relative;
}

.ticket-label-tag {
  font-size: 0.65rem;
  font-weight: 700;
  text-transform: uppercase;
  color: var(--text-muted);
  letter-spacing: 0.15em;
  margin-bottom: 4px;
}

.movie-name {
  font-size: 1.5rem;
  font-weight: 800;
  color: var(--text-primary);
  margin-bottom: 4px;
}

.theater-name-tag {
  font-size: 0.85rem;
  color: var(--color-accent);
  font-weight: 600;
}

.ticket-info-details {
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.info-row {
  display: flex;
  justify-content: space-between;
  font-size: 0.95rem;
}

.info-label {
  color: var(--text-secondary);
}

.info-val {
  color: var(--text-primary);
  font-weight: 600;
}

.ticket-payment-section {
  padding: 20px;
  background: rgba(255,255,255,0.01);
}

.payment-title {
  font-size: 0.8rem;
  font-weight: 700;
  color: var(--text-muted);
  text-transform: uppercase;
  margin-bottom: 10px;
}

.payment-option {
  display: flex;
  gap: 12px;
  background: rgba(255,255,255,0.04);
  border: 1px solid var(--border-subtle);
  padding: 12px;
  border-radius: var(--radius-md);
  align-items: center;
}

.payment-icon {
  font-size: 1.8rem;
}

.payment-desc {
  display: flex;
  flex-direction: column;
}

.payment-desc strong {
  font-size: 0.9rem;
  color: var(--text-primary);
}

.payment-desc span {
  font-size: 0.72rem;
  color: var(--text-secondary);
}

.confirm-actions {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.pay-btn {
  font-size: 1.05rem;
}

.text-center {
  text-align: center;
}

.font-mono {
  font-family: monospace;
}

.btn-spinner {
  display: inline-block;
  width: 20px;
  height: 20px;
  border: 2px solid rgba(255,255,255,0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

.disabled {
  opacity: 0.5;
  pointer-events: none;
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

@media (max-width: 640px) {
  .confirm-card {
    padding: var(--space-md);
  }
  .success-card {
    padding: var(--space-xl) var(--space-md);
  }
}

@media (max-width: 480px) {
  .ticket-grid {
    grid-template-columns: 1fr;
    gap: 12px;
  }
  .ticket-middle {
    padding: var(--space-md);
  }
  .ticket-bottom {
    padding: 12px var(--space-md);
  }
  .success-title {
    font-size: 1.6rem;
  }
}
</style>
