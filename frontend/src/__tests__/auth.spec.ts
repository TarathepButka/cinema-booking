import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

const { getMe, logout, switchRole, login } = vi.hoisted(() => ({
  getMe: vi.fn(),
  logout: vi.fn(),
  switchRole: vi.fn(),
  login: vi.fn(),
}))

vi.mock('@/api', () => ({
  authApi: {
    getMe,
    logout,
    switchRole,
    login,
  },
}))

import { useAuthStore } from '@/stores/auth'

describe('auth store cookie session', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    getMe.mockReset()
    logout.mockReset()
    switchRole.mockReset()
    login.mockReset()
  })

  it('initializes an authenticated session from /auth/me', async () => {
    getMe.mockResolvedValue({
      data: {
        userId: '123',
        email: 'user@example.com',
        role: 'USER',
      },
    })

    const authStore = useAuthStore()
    await authStore.initialize()

    expect(authStore.initialized).toBe(true)
    expect(authStore.isAuthenticated).toBe(true)
    expect(authStore.user?.email).toBe('user@example.com')
  })

  it('initializes as anonymous when the session cookie is absent', async () => {
    getMe.mockRejectedValue({ response: { status: 401 } })

    const authStore = useAuthStore()
    await authStore.initialize()

    expect(authStore.initialized).toBe(true)
    expect(authStore.isAuthenticated).toBe(false)
  })

  it('clears local state after server logout', async () => {
    logout.mockResolvedValue({ status: 204 })
    const authStore = useAuthStore()
    authStore.setUser({
      userId: '123',
      email: 'user@example.com',
      role: 'USER',
    })

    await authStore.logout()

    expect(logout).toHaveBeenCalledOnce()
    expect(authStore.user).toBeNull()
  })

  it('computes isAuthenticated, isAdmin and isUser correctly based on user role', () => {
    const authStore = useAuthStore()

    // anonymous
    expect(authStore.isAuthenticated).toBe(false)
    expect(authStore.isAdmin).toBe(false)
    expect(authStore.isUser).toBe(false)

    // admin
    authStore.setUser({ userId: '1', email: 'admin@example.com', role: 'ADMIN' })
    expect(authStore.isAuthenticated).toBe(true)
    expect(authStore.isAdmin).toBe(true)
    expect(authStore.isUser).toBe(false)

    // normal user
    authStore.setUser({ userId: '2', email: 'user@example.com', role: 'USER' })
    expect(authStore.isAuthenticated).toBe(true)
    expect(authStore.isAdmin).toBe(false)
    expect(authStore.isUser).toBe(true)
  })

  it('performs adminLogin, sets user, and sets initialized to true', async () => {
    login.mockResolvedValue({
      data: {
        user: { userId: 'admin1', email: 'admin@example.com', role: 'ADMIN' },
      },
    })

    const authStore = useAuthStore()
    const data = await authStore.adminLogin('admin@example.com', 'password')

    expect(login).toHaveBeenCalledWith('admin@example.com', 'password')
    expect(authStore.user?.email).toBe('admin@example.com')
    expect(authStore.initialized).toBe(true)
    expect(data.user.role).toBe('ADMIN')
  })

  it('deduplicates concurrent initialize calls to fetchMe/getMe once', async () => {
    getMe.mockResolvedValue({
      data: { userId: '123', email: 'user@example.com', role: 'USER' },
    })

    const authStore = useAuthStore()

    // Call twice concurrently
    await Promise.all([authStore.initialize(), authStore.initialize()])

    expect(getMe).toHaveBeenCalledOnce()
    expect(authStore.initialized).toBe(true)
  })

  it('clears session locally without calling server logout', () => {
    const authStore = useAuthStore()
    authStore.setUser({ userId: '123', email: 'user@example.com', role: 'USER' })

    authStore.clearSession()

    expect(authStore.user).toBeNull()
    expect(authStore.isAuthenticated).toBe(false)
    // server logout shouldn't be called
    expect(logout).not.toHaveBeenCalled()
  })

  it('switches role and updates user', async () => {
    switchRole.mockResolvedValue({
      data: {
        user: { userId: '123', email: 'user@example.com', role: 'ADMIN' },
      },
    })

    const authStore = useAuthStore()
    authStore.setUser({ userId: '123', email: 'user@example.com', role: 'USER' })

    const data = await authStore.switchRole('ADMIN')

    expect(switchRole).toHaveBeenCalledWith('ADMIN')
    expect(authStore.user?.role).toBe('ADMIN')
    expect(data.user.role).toBe('ADMIN')
  })

  it('redirects to Google login via window.location', () => {
    const originalLocation = window.location
    const locationMock = { href: '' }

    Object.defineProperty(window, 'location', {
      value: locationMock,
      configurable: true,
      writable: true,
    })

    const authStore = useAuthStore()
    authStore.loginWithGoogle()

    expect(window.location.href).toBe('/api/auth/google/login')

    Object.defineProperty(window, 'location', {
      value: originalLocation,
      configurable: true,
      writable: true,
    })
  })
})

