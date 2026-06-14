<script setup lang="ts">
import { onMounted, computed } from 'vue'
import { useBookingStore } from '@/stores/booking'
import BookingCard from '@/components/BookingCard.vue'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import EmptyState from '@/components/EmptyState.vue'

const bookingStore = useBookingStore()

onMounted(() => {
  bookingStore.fetchMyBookings()
})

const bookings = computed(() => {
  const list = bookingStore.myBookings || []
  // Sort bookings so newest are first
  return [...list].sort((a, b) => {
    return new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
  })
})

const loading = computed(() => bookingStore.loading)
const error = computed(() => bookingStore.error)
</script>

<template>
  <div class="container page-content">
    <!-- Page header -->
    <div class="page-header animate-fade-up">
      <h1 class="page-title">My Tickets</h1>
    </div>

    <LoadingSpinner v-if="loading" message="Loading booking history..." />

    <EmptyState
      v-else-if="error"
      title="An error occurred"
    />

    <EmptyState
      v-else-if="bookings.length === 0"
      title="No Bookings Yet"
    >
      <RouterLink to="/" class="btn btn-primary" style="margin-top: 16px">Book Tickets Now →</RouterLink>
    </EmptyState>

    <div v-else class="tickets-grid animate-fade-up">
      <BookingCard
        v-for="booking in bookings"
        :key="booking.id"
        :booking="booking"
      />
    </div>
  </div>
</template>

<style scoped>
.page-header {
  margin-bottom: var(--space-2xl);
}

.tickets-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
  gap: var(--space-lg);
  align-items: start;
}

@media (max-width: 768px) {
  .tickets-grid { grid-template-columns: 1fr; }
}
</style>
