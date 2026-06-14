<script setup lang="ts">
import { computed } from 'vue'
import type { Seat } from '@/stores/showtime'
import SeatItem from './SeatItem.vue'

const props = defineProps<{
  seats: Seat[]
  myUserId?: string
  selectedSeats?: string[]
}>()

const emit = defineEmits<{
  seatClick: [seat: Seat]
}>()

// Group seats by zone, then by row
const zones = computed(() => {
  const zoneOrder = ['FRONT', 'MIDDLE', 'BACK']
  const grouped: Record<string, Record<string, Seat[]>> = {}

  for (const seat of props.seats) {
    if (!grouped[seat.zone]) {
      grouped[seat.zone] = {}
    }
    const zoneGroup = grouped[seat.zone]
    if (zoneGroup) {
      if (!zoneGroup[seat.row]) {
        zoneGroup[seat.row] = []
      }
      const rowGroup = zoneGroup[seat.row]
      if (rowGroup) {
        rowGroup.push(seat)
      }
    }
  }

  return zoneOrder
    .filter((z) => grouped[z])
    .map((zoneName) => {
      const rows = Object.entries(grouped[zoneName] || {})
        .sort(([a], [b]) => a.localeCompare(b))
        .map(([row, seats]) => ({
          row,
          seats: seats.sort((a, b) => a.number - b.number),
        }))
      return { name: zoneName, rows }
    })
})

const zoneColors: Record<string, string> = {
  FRONT: '#ef4444',
  MIDDLE: '#f59e0b',
  BACK: '#22c55e',
}

const zonePrices = computed(() => {
  const prices: Record<string, number> = {}
  for (const seat of props.seats) {
    if (!prices[seat.zone]) prices[seat.zone] = seat.price
  }
  return prices
})

</script>

<template>
  <div class="seat-grid-wrap">
    <!-- Zones -->
    <div class="zones">
      <div v-for="zone in zones" :key="zone.name" class="zone">
        <div class="zone-header">
          <div class="zone-indicator" :style="{ background: zoneColors[zone.name] }"></div>
          <span class="zone-name">{{ zone.name }} ZONE</span>
          <span class="zone-price">฿{{ zonePrices[zone.name] }}</span>
        </div>

        <div class="rows">
          <div v-for="rowData in zone.rows" :key="rowData.row" class="row">
            <span class="row-label">{{ rowData.row }}</span>
            <div class="seats-row">
              <SeatItem
                v-for="seat in rowData.seats"
                :key="seat.seatLabel"
                :seat="seat"
                :my-user-id="myUserId"
                :is-selected="selectedSeats?.includes(seat.seatLabel)"
                @click="emit('seatClick', seat)"
              />
            </div>
            <span class="row-label">{{ rowData.row }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Legend -->
    <div class="legend">
      <div class="legend-item">
        <span class="legend-seat available">A1</span>
        <span>Available</span>
      </div>
      <div class="legend-item">
        <span class="legend-seat locked">A1</span>
        <span>Reserved</span>
      </div>
      <div class="legend-item">
        <span class="legend-seat booked">A1</span>
        <span>Booked</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.seat-grid-wrap {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-xl);
  padding: var(--space-lg);
}

/* Zones */
.zones {
  display: flex;
  flex-direction: column;
  gap: var(--space-xl);
  width: 100%;
  max-width: 700px;
  overflow-x: auto;
  padding-bottom: var(--space-md);
  /* Custom scrollbar for custom aesthetics */
  scrollbar-width: thin;
  scrollbar-color: rgba(229, 9, 20, 0.3) transparent;
}

.zones::-webkit-scrollbar {
  height: 4px;
}
.zones::-webkit-scrollbar-thumb {
  background: rgba(229, 9, 20, 0.3);
  border-radius: var(--radius-full);
}

.zone {
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
  min-width: max-content; /* Prevent zone header collapse */
}

.zone-header {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
}

.zone-indicator {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  opacity: 0.8;
}

.zone-name {
  font-size: 0.75rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--text-muted);
}

.zone-price {
  margin-left: auto;
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--color-accent);
}

.rows {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.row {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  justify-content: center;
  min-width: max-content;
}

.row-label {
  width: 20px;
  text-align: center;
  font-size: 0.7rem;
  font-weight: 700;
  color: var(--text-muted);
  flex-shrink: 0;
}

.seats-row {
  display: flex;
  gap: 6px;
  flex-wrap: nowrap;
  justify-content: center;
}

/* Legend */
.legend {
  display: flex;
  align-items: center;
  gap: var(--space-xl);
  flex-wrap: wrap;
  justify-content: center;
  padding-top: var(--space-md);
  border-top: 1px solid var(--border-subtle);
  width: 100%;
  max-width: 600px;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 0.8rem;
  color: var(--text-secondary);
}

.legend-seat {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 28px;
  border-radius: 4px;
  font-size: 0.55rem;
  font-weight: 700;
  border: 2px solid;
}

.legend-seat.available {
  background: var(--seat-available-bg);
  border-color: var(--seat-available);
  color: var(--seat-available);
}

.legend-seat.selected {
  background: var(--seat-available);
  border-color: var(--seat-available);
  color: #000;
}

.legend-seat.locked {
  background: var(--seat-locked-bg);
  border-color: var(--seat-locked);
  color: var(--seat-locked);
}

.legend-seat.booked {
  background: var(--seat-booked-bg);
  border-color: var(--seat-booked);
  color: var(--text-muted);
}
</style>
