<script setup lang="ts">
import type { Seat } from '@/stores/showtime'

const props = defineProps<{
  seat: Seat
  myUserId?: string
  isSelected?: boolean
}>()

const emit = defineEmits<{
  click: [seat: Seat]
}>()

const isLockedByMe = () =>
  props.seat.status === 'LOCKED' && props.seat.lockedBy === props.myUserId

const seatClass = () => {
  if (props.isSelected) return 'seat seat-selected'
  if (props.seat.status === 'BOOKED') return 'seat seat-booked'
  if (props.seat.status === 'LOCKED') {
    return isLockedByMe() ? 'seat seat-selected' : 'seat seat-locked'
  }
  return 'seat seat-available'
}

const isClickable = () =>
  props.isSelected || props.seat.status === 'AVAILABLE' || isLockedByMe()

const tooltip = () => {
  if (props.seat.status === 'BOOKED') return 'Already booked'
  if (props.seat.status === 'LOCKED') {
    return isLockedByMe() ? 'Your reservation (click to deselect)' : 'Reserved by someone else'
  }
  return `${props.seat.seatLabel} — ฿${props.seat.price} (${props.seat.zone})`
}

function handleClick() {
  if (isClickable()) {
    emit('click', props.seat)
  }
}
</script>

<template>
  <button
    :class="seatClass()"
    :disabled="!isSelected && (seat.status === 'BOOKED' || (seat.status === 'LOCKED' && !isLockedByMe()))"
    :title="tooltip()"
    :aria-label="`Seat ${seat.seatLabel}, ${seat.zone} zone, ฿${seat.price}, ${seat.status}`"
    @click="handleClick"
  >
    <span class="seat-label">{{ seat.seatLabel }}</span>
  </button>
</template>

<style scoped>
.seat {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 40px;
  border-radius: 6px 6px 4px 4px;
  border: 2px solid;
  font-size: 0.65rem;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.15s ease;
  position: relative;
  gap: 2px;
}

.seat-label {
  font-size: 0.6rem;
  font-weight: 700;
  letter-spacing: 0.02em;
}

.seat-check {
  font-size: 0.7rem;
  position: absolute;
  top: 2px;
  right: 4px;
}

/* Status styles */
.seat-available {
  background: var(--seat-available-bg);
  border-color: var(--seat-available);
  color: var(--seat-available);
}
.seat-available:hover:not(:disabled) {
  background: var(--seat-available);
  color: #000;
  transform: scale(1.12);
  box-shadow: 0 4px 12px rgba(34, 197, 94, 0.4);
}

.seat-selected {
  background: var(--seat-available);
  border-color: var(--seat-available);
  color: #000;
  transform: scale(1.08);
  box-shadow: 0 4px 12px rgba(34, 197, 94, 0.4);
}

.seat-locked {
  background: var(--seat-locked-bg);
  border-color: var(--seat-locked);
  color: var(--seat-locked);
  cursor: not-allowed;
  opacity: 0.7;
}

.seat-locked-mine {
  background: var(--seat-locked-mine-bg);
  border-color: var(--seat-locked-mine);
  color: var(--seat-locked-mine);
  animation: pulse-amber 2s ease-in-out infinite;
}
.seat-locked-mine:hover:not(:disabled) {
  background: rgba(245, 158, 11, 0.3);
  transform: scale(1.05);
}

.seat-booked {
  background: var(--seat-booked-bg);
  border-color: var(--seat-booked);
  color: var(--text-muted);
  cursor: not-allowed;
  opacity: 0.5;
}

.seat:disabled {
  transform: none !important;
  box-shadow: none !important;
}

/* Seat shape like a cinema seat */
.seat::before {
  content: '';
  position: absolute;
  bottom: -6px;
  left: 4px;
  right: 4px;
  height: 4px;
  background: currentColor;
  border-radius: 0 0 3px 3px;
  opacity: 0.3;
}

@media (max-width: 480px) {
  .seat {
    width: 34px;
    height: 30px;
    font-size: 0.55rem;
    border-radius: 4px 4px 3px 3px;
    border-width: 1.5px;
  }
  .seat::before {
    bottom: -4px;
    height: 3px;
    border-radius: 0 0 2px 2px;
  }
  .seat-label {
    font-size: 0.52rem;
  }
}
</style>
