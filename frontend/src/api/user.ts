import api from '@/composables/useApi'

export const userApi = {
  getSuggestions: (q: string, page = 1, limit = 10) =>
    api.get('/users/suggestions', { params: { q, page, limit } }),
}
