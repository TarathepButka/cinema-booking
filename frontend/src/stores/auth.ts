import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { authApi } from '@/api'
import { ROLES, type Role } from '@/constants'

interface User {
  userId: string
  email: string
  role: string
  name?: string
  picture?: string
  dbRole?: string
}

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const initialized = ref(false)
  let initialization: Promise<void> | null = null

  const isAuthenticated = computed(() => user.value !== null)
  const isAdmin = computed(() => user.value?.role === ROLES.ADMIN)
  const isUser = computed(() => user.value?.role === ROLES.USER)

  function setUser(newUser: User) {
    user.value = newUser
  }

  function clearSession() {
    user.value = null
  }

  async function fetchMe() {
    const res = await authApi.getMe()
    setUser(res.data)
    return res.data
  }

  async function initialize() {
    if (initialized.value) {
      return
    }
    if (!initialization) {
      initialization = (async () => {
        try {
          await fetchMe()
        } catch {
          clearSession()
        } finally {
          initialized.value = true
        }
      })()
    }
    await initialization
  }

  async function switchRole(targetRole: Role) {
    const res = await authApi.switchRole(targetRole)
    setUser(res.data.user)
    return res.data
  }

  function loginWithGoogle() {
    window.location.href = '/api/auth/google/login'
  }

  async function adminLogin(email: string, password: string) {
    const res = await authApi.login(email, password)
    setUser(res.data.user)
    initialized.value = true
    return res.data
  }

  async function logout() {
    try {
      await authApi.logout()
    } finally {
      clearSession()
      initialized.value = true
    }
  }

  return {
    user,
    initialized,
    isAuthenticated,
    isAdmin,
    isUser,
    setUser,
    clearSession,
    fetchMe,
    initialize,
    switchRole,
    loginWithGoogle,
    adminLogin,
    logout,
  }
})
