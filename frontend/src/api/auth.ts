import api from '@/composables/useApi'
import { type Role } from '@/constants'

export const authApi = {
  getMe: () => api.get('/auth/me'),
  switchRole: (role: Role) => api.post('/auth/switch-role', { role }),
  login: (email: string, password: string) => api.post('/auth/login', { email, password }),
  logout: () => api.post('/auth/logout'),
}
