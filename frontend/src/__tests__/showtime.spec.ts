import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

const { getById, getByMovieId } = vi.hoisted(() => ({
  getById: vi.fn(),
  getByMovieId: vi.fn(),
}))

vi.mock('@/api', () => ({
  showtimeApi: {
    getById,
    getByMovieId,
  },
}))

import { SEAT_STATUS } from '@/constants'
import { useShowtimeStore } from '@/stores/showtime'

describe('showtime store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    getById.mockReset()
    getByMovieId.mockReset()
  })

  it('loads a showtime and updates seat state from real-time events', async () => {
    getById.mockResolvedValue({
      data: {
        data: {
          id: 'showtime-1',
          seats: [{
            row: 'A',
            number: 1,
            seatLabel: 'A1',
            zone: 'FRONT',
            price: 180,
            status: SEAT_STATUS.LOCKED,
            lockedBy: 'user-1',
            lockedUntil: '2026-06-14T12:00:00Z',
          }],
        },
      },
    })

    const store = useShowtimeStore()
    await store.fetchById('showtime-1')
    store.updateSeatStatus('A1', SEAT_STATUS.AVAILABLE)

    expect(store.currentShowtime?.seats?.[0]).toMatchObject({
      status: SEAT_STATUS.AVAILABLE,
      lockedBy: undefined,
      lockedUntil: undefined,
    })
  })

  it('exposes an API error and stops loading', async () => {
    getById.mockRejectedValue({ response: { data: { error: 'Showtime unavailable' } } })

    const store = useShowtimeStore()
    await expect(store.fetchById('missing')).resolves.toBeNull()

    expect(store.error).toBe('Showtime unavailable')
    expect(store.loading).toBe(false)
  })

  it('sets loading flag while fetching showtime', async () => {
    let resolvePromise: any
    const promise = new Promise((resolve) => {
      resolvePromise = resolve
    })
    getById.mockReturnValue(promise)

    const store = useShowtimeStore()
    const fetchPromise = store.fetchById('showtime-1')

    expect(store.loading).toBe(true)

    resolvePromise({ data: { data: { id: 'showtime-1', seats: [] } } })
    await fetchPromise

    expect(store.loading).toBe(false)
  })

  it('fetches showtimes by movieId and handles success & error states', async () => {
    const mockShowtimes = [
      { id: 'st-1', movieId: 'movie-123', theaterId: 'th-1', theaterName: 'Hall A' }
    ]
    getByMovieId.mockResolvedValue({
      data: {
        data: mockShowtimes,
      },
    })

    const store = useShowtimeStore()
    const fetchPromise = store.fetchByMovieId('movie-123')
    expect(store.loading).toBe(true)

    await fetchPromise

    expect(store.loading).toBe(false)
    expect(store.showtimes).toEqual(mockShowtimes)
    expect(store.error).toBeNull()

    // Test error case
    getByMovieId.mockRejectedValue({ response: { data: { error: 'Movie not found' } } })
    await store.fetchByMovieId('invalid-movie')
    expect(store.error).toBe('Movie not found')
    expect(store.loading).toBe(false)
  })

  it('no-ops updateSeatStatus when currentShowtime or seats are not found', () => {
    const store = useShowtimeStore()

    // No current showtime
    expect(() => store.updateSeatStatus('A1', SEAT_STATUS.AVAILABLE)).not.toThrow()

    // Showtime exists but has no seats property
    store.currentShowtime = { id: 'st-1', seats: undefined } as any
    expect(() => store.updateSeatStatus('A1', SEAT_STATUS.AVAILABLE)).not.toThrow()

    // Showtime exists with seats, but seat is not found
    store.currentShowtime = {
      id: 'st-1',
      seats: [{ seatLabel: 'A2', status: SEAT_STATUS.LOCKED }]
    } as any
    store.updateSeatStatus('A1', SEAT_STATUS.AVAILABLE)
    expect(store.currentShowtime?.seats?.[0]?.status).toBe(SEAT_STATUS.LOCKED)
  })
})

