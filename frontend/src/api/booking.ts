import api from '@/composables/useApi'

export const bookingApi = {
  getStats: (params: any) => api.get('/bookings/stats', { params }),
  getAll: (params: any) => api.get('/bookings', { params }),
  lockSeat: (showtimeId: string, seatLabel: string) => 
    api.post('/bookings/locks', { showtimeId, seatLabel }),
  unlockSeat: (showtimeId: string, seatLabel: string) => 
    api.delete('/bookings/locks', { data: { showtimeId, seatLabel } }),
  create: (payload: any) => api.post('/bookings', payload),
  getMyBookings: () => api.get('/users/me/bookings'),
}
