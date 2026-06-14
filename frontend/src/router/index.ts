import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    // Public
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/LoginPage.vue'),
      meta: { public: true, hideNavbar: true },
    },

    // User routes (requires authentication)
    {
      path: '/',
      name: 'home',
      component: () => import('@/views/HomePage.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/movies/:id/showtime',
      name: 'movie-detail',
      component: () => import('@/views/MovieDetailPage.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/movies/:movieId/showtime/:id/select-seat',
      name: 'seat-map',
      component: () => import('@/views/SeatMapPage.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/payment',
      name: 'payment',
      component: () => import('@/views/BookingConfirmPage.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/my-bookings',
      name: 'my-bookings',
      component: () => import('@/views/MyBookingsPage.vue'),
      meta: { requiresAuth: true },
    },

    // Admin routes (requires ADMIN role)
    {
      path: '/manage',
      component: () => import('@/views/AdminLayout.vue'),
      meta: { requiresAuth: true, requiresAdmin: true },
      children: [
        {
          path: '',
          redirect: { name: 'admin-dashboard' },
        },
        {
          path: 'bookings',
          name: 'admin-dashboard',
          component: () => import('@/views/AdminDashboardPage.vue'),
        },
        {
          path: 'audit-logs',
          name: 'admin-audit-logs',
          component: () => import('@/views/AdminAuditLogPage.vue'),
        },
        {
          path: 'movies-showtimes',
          name: 'admin-movies',
          component: () => import('@/views/AdminMoviesPage.vue'),
        },
      ]
    },
  ],
  scrollBehavior() {
    return { top: 0 }
  },
})

// Navigation guard
router.beforeEach(async (to) => {
  const authStore = useAuthStore()
  await authStore.initialize()

  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    return { name: 'login' }
  }

  if (to.meta.requiresAdmin && !authStore.isAdmin) {
    return { name: 'home' }
  }

  if (authStore.isAuthenticated && authStore.isAdmin && !to.meta.requiresAdmin) {
    return { name: 'admin-dashboard' }
  }

  if (to.name === 'login' && authStore.isAuthenticated) {
    return authStore.isAdmin ? { name: 'admin-dashboard' } : { name: 'home' }
  }
})

export default router
