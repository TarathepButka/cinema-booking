import api from '@/composables/useApi'

export const showtimeApi = {
  getByMovieId: (movieId: string) => api.get('/showtimes', { params: { movie_id: movieId } }),
  getById: (id: string) => api.get(`/showtimes/${id}`),
}
