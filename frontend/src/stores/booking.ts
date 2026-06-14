import { defineStore } from 'pinia'
import { ref } from 'vue'
import { bookingApi } from '@/api'
import { type BookingStatus } from '@/constants'

export interface BookedSeat {
  seatLabel: string
  row: string
  number: number
  zone: string
  price: number
}

export interface Booking {
  id: string
  userId: string
  showtimeId: string
  seats: BookedSeat[]
  status: BookingStatus
  totalPrice: number
  userEmail: string
  userName: string
  movieTitle: string
  theaterName: string
  startTime: string
  createdAt: string
}

export const useBookingStore = defineStore('booking', () => {
  const myBookings = ref<Booking[]>([])
  const currentLock = ref<{
    showtimeId: string
    seatLabel: string
    expiresAt: string
  } | null>(null)
  const pendingBooking = ref<{
    showtimeId: string
    movieId: string
    seatLabels: string[]
    movieTitle: string
    theaterName: string
    seats: Array<{ seatLabel: string; zone: string; price: number }>
    totalPrice: number
    expiresAt: string
    locksReleased?: boolean
  } | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function lockSeat(showtimeId: string, seatLabel: string) {
    loading.value = true
    error.value = null
    try {
      const res = await bookingApi.lockSeat(showtimeId, seatLabel)
      currentLock.value = res.data.data
      return res.data.data
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to lock seat'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function confirmBooking(
    showtimeId: string,
    seatLabels: string[],
  ) {
    loading.value = true
    error.value = null
    try {
      const res = await bookingApi.create({
        showtimeId,
        seatLabels,
      })
      pendingBooking.value = null
      currentLock.value = null
      return res.data.data as Booking
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Booking confirmation failed'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function releaseSeat(showtimeId: string, seatLabel: string) {
    try {
      await bookingApi.unlockSeat(showtimeId, seatLabel)
      currentLock.value = null
      pendingBooking.value = null
    } catch (e: any) {
      console.error('Release seat error:', e)
    }
  }

  async function releaseSeatOnly(showtimeId: string, seatLabel: string) {
    try {
      await bookingApi.unlockSeat(showtimeId, seatLabel)
    } catch (e: any) {
      console.error('Release seat only error:', e)
    }
  }

  async function fetchMyBookings() {
    loading.value = true
    try {
      const res = await bookingApi.getMyBookings()
      myBookings.value = res.data.data || []
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to load bookings'
    } finally {
      loading.value = false
    }
  }

  function setPendingBooking(data: typeof pendingBooking.value) {
    pendingBooking.value = data
  }

  return {
    myBookings,
    currentLock,
    pendingBooking,
    loading,
    error,
    lockSeat,
    confirmBooking,
    releaseSeat,
    releaseSeatOnly,
    fetchMyBookings,
    setPendingBooking,
  }
})
