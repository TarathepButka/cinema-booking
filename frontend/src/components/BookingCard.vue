<script setup lang="ts">
import { computed } from 'vue'
import type { Booking } from '@/stores/booking'
import { formatBookingCode, formatCurrency, formatDateTime, formatSeatLabels, formatTime } from '@/utils/format'
import { getBookingStatusPresentation } from '@/utils/status'

const props = defineProps<{
  booking: Booking
  showUser?: boolean
}>()

const formattedDate = computed(() => {
  if (!props.booking.startTime) return ''
  return formatDateTime(props.booking.startTime, {
    weekday: 'short',
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
})

const formattedTime = computed(() => {
  if (!props.booking.startTime) return ''
  return formatTime(props.booking.startTime)
})

const statusLabel = computed(() => getBookingStatusPresentation(props.booking.status))

const formattedPrice = computed(() => formatCurrency(props.booking.totalPrice))

const seatLabelsJoined = computed(() => formatSeatLabels(props.booking.seats))
const bookingCode = computed(() => formatBookingCode(props.booking.id))
const movieTitle = computed(() => props.booking.movieTitle || 'Movie unavailable')
const theaterName = computed(() => props.booking.theaterName || 'Theater unavailable')
</script>

<template>
  <div class="booking-ticket">
    <!-- Top section — movie info + status -->
    <div class="ticket-top">
      <div class="ticket-brand-line">
        <span class="ticket-brand">CINEPLEX</span>
        <span class="ticket-logo">🎬</span>
      </div>
      <span class="status-badge" :class="statusLabel.ticketClass">{{ statusLabel.label }}</span>
    </div>

    <!-- Movie title section -->
    <div class="ticket-movie-section">
      <h3 class="ticket-movie-title">{{ movieTitle }}</h3>
      <div class="ticket-theater">
        <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"/><circle cx="12" cy="10" r="3"/></svg>
        {{ theaterName }}
      </div>
    </div>

    <!-- Perforated separator -->
    <div class="ticket-perforated"></div>

    <!-- Details grid -->
    <div class="ticket-details">
      <div class="detail-cell">
        <span class="detail-label">DATE</span>
        <span class="detail-value">{{ formattedDate }}</span>
      </div>
      <div class="detail-cell">
        <span class="detail-label">TIME</span>
        <span class="detail-value text-primary">{{ formattedTime }}</span>
      </div>
      <div class="detail-cell">
        <span class="detail-label">SEATS</span>
        <span class="detail-value seat-val">{{ seatLabelsJoined }}</span>
      </div>
      <div class="detail-cell" v-if="showUser">
        <span class="detail-label">GUEST</span>
        <span class="detail-value email-val">{{ booking.userEmail }}</span>
      </div>
    </div>

    <!-- Perforated separator -->
    <div class="ticket-perforated"></div>

    <!-- Footer — price + booking code -->
    <div class="ticket-footer">
      <div class="ticket-price-section">
        <span class="footer-label">TOTAL</span>
        <span class="ticket-price">{{ formattedPrice }}</span>
      </div>
      <div class="ticket-code-section">
        <span class="footer-label">BOOKING ID</span>
        <span class="ticket-code">{{ bookingCode }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.booking-ticket {
  background: var(--color-bg-card);
  border: 1px solid var(--border-card);
  border-radius: var(--radius-lg);
  overflow: hidden;
  transition: all var(--transition-base);
  position: relative;
}

.booking-ticket::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: linear-gradient(90deg, var(--color-primary) 0%, rgba(229,9,20,0.3) 100%);
}

.booking-ticket:hover {
  border-color: rgba(229, 9, 20, 0.2);
  box-shadow: 0 8px 40px rgba(0,0,0,0.5), 0 0 0 1px rgba(229,9,20,0.1);
  transform: translateY(-2px);
}

/* Top */
.ticket-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 20px 10px;
  background: rgba(229, 9, 20, 0.04);
}

.ticket-brand-line {
  display: flex;
  align-items: center;
  gap: 6px;
}

.ticket-brand {
  font-family: 'Barlow Condensed', sans-serif;
  font-size: 0.85rem;
  font-weight: 800;
  letter-spacing: 0.15em;
  color: var(--text-muted);
}

.ticket-logo { font-size: 0.9rem; }

.status-badge {
  font-size: 0.65rem;
  font-weight: 800;
  letter-spacing: 0.1em;
  padding: 3px 10px;
  border-radius: var(--radius-full);
}

.status-confirmed {
  background: rgba(34,197,94,0.12);
  color: #4ade80;
  border: 1px solid rgba(34,197,94,0.25);
}
.status-pending {
  background: rgba(245,158,11,0.12);
  color: #fbbf24;
  border: 1px solid rgba(245,158,11,0.25);
}
.status-cancelled {
  background: rgba(239,68,68,0.12);
  color: #f87171;
  border: 1px solid rgba(239,68,68,0.25);
}
.status-expired {
  background: rgba(78,78,100,0.15);
  color: var(--text-muted);
  border: 1px solid var(--border-subtle);
}

/* Movie section */
.ticket-movie-section {
  padding: 14px 20px;
}

.ticket-movie-title {
  font-family: 'Outfit', sans-serif;
  font-size: 1.25rem;
  font-weight: 800;
  color: var(--text-primary);
  margin-bottom: 6px;
  line-height: 1.2;
}

.ticket-theater {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 0.8rem;
  color: var(--color-accent);
  font-weight: 600;
}

/* Perforated line */
.ticket-perforated {
  position: relative;
  height: 1px;
  border-top: 2px dashed rgba(255,255,255,0.08);
  margin: 0 0;
}

.ticket-perforated::before,
.ticket-perforated::after {
  content: '';
  position: absolute;
  top: -10px;
  width: 18px;
  height: 18px;
  background: var(--color-bg);
  border-radius: 50%;
  border: 1px solid var(--border-card);
}

.ticket-perforated::before { left: -10px; }
.ticket-perforated::after  { right: -10px; }

/* Details */
.ticket-details {
  padding: 16px 20px;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px;
}

.detail-cell {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.detail-label {
  font-size: 0.62rem;
  font-weight: 700;
  letter-spacing: 0.1em;
  color: var(--text-muted);
  text-transform: uppercase;
}

.detail-value {
  font-size: 0.88rem;
  font-weight: 600;
  color: var(--text-secondary);
}

.seat-val {
  font-family: 'Outfit', monospace;
  color: var(--color-accent);
  font-size: 0.85rem;
  letter-spacing: 0.03em;
}

.email-val {
  font-size: 0.78rem;
  color: var(--text-muted);
  word-break: break-all;
}

/* Footer */
.ticket-footer {
  padding: 14px 20px;
  background: rgba(255,255,255,0.02);
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
}

.ticket-price-section,
.ticket-code-section {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.ticket-code-section { text-align: right; }

.footer-label {
  font-size: 0.6rem;
  font-weight: 700;
  letter-spacing: 0.1em;
  color: var(--text-muted);
  text-transform: uppercase;
}

.ticket-price {
  font-family: 'Barlow Condensed', sans-serif;
  font-size: 1.6rem;
  font-weight: 800;
  letter-spacing: 0.02em;
  color: var(--text-primary);
}

.ticket-code {
  font-family: 'Courier New', monospace;
  font-size: 1rem;
  font-weight: 700;
  color: var(--color-accent);
  letter-spacing: 0.08em;
}
</style>
