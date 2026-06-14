import api from '@/composables/useApi'

export const movieApi = {
  getAll: () => api.get('/movies'),
  getById: (id: string) => api.get(`/movies/${id}`),
  getSuggestions: (q: string, page = 1, limit = 10) =>
    api.get('/movies/suggestions', { params: { q, page, limit } }),
  create: (payload: any) => api.post('/movies', payload),
}
