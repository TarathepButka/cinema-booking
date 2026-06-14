import { defineStore } from 'pinia'
import { ref } from 'vue'
import { showtimeApi } from '@/api'
import { SEAT_STATUS, type SeatStatus } from '@/constants'

export interface Seat {
  row: string
  number: number
  seatLabel: string
  zone: string
  price: number
  status: SeatStatus
  lockedBy?: string
  lockedUntil?: string
}

export interface Showtime {
  id: string
  movieId: string
  theaterId: string
  theaterName: string
  startTime: string
  endTime: string
  slotLabel: string
  priceByZone: Record<string, number>
  seats?: Seat[]
  movie?: {
    id: string
    title: string
    duration: number
    posterUrl: string
    genre: string[]
    rating: number
  }
}

export const useShowtimeStore = defineStore('showtime', () => {
  const showtimes = ref<Showtime[]>([])
  const currentShowtime = ref<Showtime | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function fetchByMovieId(movieId: string) {
    loading.value = true
    error.value = null
    try {
      const res = await showtimeApi.getByMovieId(movieId)
      showtimes.value = res.data.data
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to load showtimes'
    } finally {
      loading.value = false
    }
  }

  async function fetchById(id: string) {
    loading.value = true
    error.value = null
    try {
      const res = await showtimeApi.getById(id)
      currentShowtime.value = res.data.data
      return res.data.data as Showtime
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Showtime not found'
      return null
    } finally {
      loading.value = false
    }
  }

  // Update a seat status in real-time from WebSocket message
  function updateSeatStatus(seatLabel: string, status: Seat['status']) {
    if (!currentShowtime.value?.seats) return
    const seat = currentShowtime.value.seats.find((s) => s.seatLabel === seatLabel)
    if (seat) {
      seat.status = status
      if (status === SEAT_STATUS.AVAILABLE) {
        seat.lockedBy = undefined
        seat.lockedUntil = undefined
      }
    }
  }

  return {
    showtimes,
    currentShowtime,
    loading,
    error,
    fetchByMovieId,
    fetchById,
    updateSeatStatus,
  }
})
