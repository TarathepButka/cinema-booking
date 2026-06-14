import api from '@/composables/useApi'

export const auditApi = {
  getLogs: (params: any) => api.get('/audit-logs', { params }),
}
