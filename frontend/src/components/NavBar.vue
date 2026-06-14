<script setup lang="ts">
import { computed, ref } from 'vue'
import { RouterLink, useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const router = useRouter()
const route = useRoute()
const menuOpen = ref(false)

const navLinks = computed(() => {
  if (authStore.isAdmin) {
    return []
  }
  return [
    { to: '/', label: 'Movies', icon: null },
    { to: '/my-bookings', label: 'My Tickets', icon: null },
  ]
})

async function logout() {
  await authStore.logout()
  window.location.href = '/login'
}

function isActive(to: string) {
  if (to === '/') return route.path === '/'
  return route.path.startsWith(to)
}

function closeDropdown() {
  setTimeout(() => {
    menuOpen.value = false
  }, 200)
}

async function switchRoleTo(targetRole: 'ADMIN' | 'USER') {
  if ((targetRole === 'ADMIN' && authStore.isAdmin) || (targetRole === 'USER' && !authStore.isAdmin)) {
    return
  }
  
  try {
    await authStore.switchRole(targetRole)
    if (targetRole === 'ADMIN') {
      router.push({ name: 'admin-dashboard' })
    } else {
      router.push('/')
    }
  } catch (error) {
    console.error('Failed to switch role:', error)
  }
}
</script>

<template>
  <nav class="navbar">
    <div class="navbar-inner">
      <!-- Logo -->
      <RouterLink :to="authStore.isAdmin ? '/manage/bookings' : '/'" class="navbar-logo">
        <img class="logo-icon" src="/favicon.svg" alt="" aria-hidden="true" />
        <span class="logo-text">CINEPLEX</span>
      </RouterLink>

      <!-- Nav Links (desktop) -->
      <div class="navbar-links">
        <RouterLink
          v-for="link in navLinks"
          :key="link.to"
          :to="link.to"
          class="nav-link"
          :class="{ 'nav-link-active': isActive(link.to) }"
        >
          {{ link.label }}
        </RouterLink>
      </div>

      <!-- User Area -->
      <div class="navbar-user">
        <template v-if="authStore.isAuthenticated">
          <div class="user-profile-dropdown" :class="{ 'dropdown-open': menuOpen }">
            <button class="profile-trigger" @click="menuOpen = !menuOpen" @blur="closeDropdown">
              <span class="user-display-name">
                <span class="user-display-role" :class="authStore.isAdmin ? 'role-text-admin' : 'role-text-user'">
                  {{ authStore.isAdmin ? 'ADMIN' : 'MEMBER' }}
                </span>
                <span class="user-display-divider">|</span>
                <span class="user-display-name-text">
                  {{ authStore.user?.name || authStore.user?.email.split('@')[0] }}
                </span>
              </span>
              <img 
                :src="authStore.user?.picture || 'https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?w=80&h=80&fit=crop'" 
                alt="Profile picture" 
                class="user-avatar"
                referrerpolicy="no-referrer"
              />
              <svg class="dropdown-chevron" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <polyline points="6 9 12 15 18 9"/>
              </svg>
            </button>
            
            <div class="dropdown-menu">
              <!-- Switch Role Submenu (Only if dbRole is ADMIN) -->
              <div v-if="authStore.user?.dbRole === 'ADMIN'" class="dropdown-submenu-container">
                <div class="dropdown-item submenu-trigger">
                  <svg class="item-icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <path d="m16 3 4 4-4 4M20 7H4M8 21l-4-4 4-4M4 17h16"/>
                  </svg>
                  <span style="flex: 1">Switch Role</span>
                </div>
                
                <!-- Submenu panel on the left -->
                <div class="submenu-left">
                  <button 
                    class="submenu-item" 
                    :class="{ active: authStore.isAdmin }" 
                    @click="switchRoleTo('ADMIN')"
                  >
                    Admin
                  </button>
                  <button 
                    class="submenu-item" 
                    :class="{ active: !authStore.isAdmin }" 
                    @click="switchRoleTo('USER')"
                  >
                    Member
                  </button>
                </div>
              </div>
              
              <button class="dropdown-item logout-item" @click="logout">
                <svg class="item-icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" y1="12" x2="9" y2="12"/>
                </svg>
                Logout
              </button>
            </div>
          </div>
        </template>
      </div>
    </div>

    <!-- Red bottom glow line -->
    <div class="navbar-glow-line"></div>
  </nav>
</template>

<style scoped>
.navbar {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 100;
  background: rgba(8, 8, 15, 0.92);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  height: 64px;
}

.navbar-glow-line {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 1px;
  background: linear-gradient(90deg, transparent 0%, rgba(229, 9, 20, 0.4) 50%, transparent 100%);
}

.navbar-inner {
  max-width: 1260px;
  margin: 0 auto;
  padding: 0 var(--space-lg);
  height: 100%;
  display: flex;
  align-items: center;
  gap: var(--space-xl);
  position: relative;
}

/* Logo */
.navbar-logo {
  display: flex;
  align-items: center;
  gap: 8px;
  text-decoration: none;
  flex-shrink: 0;
}

.logo-icon {
  width: 28px;
  height: 28px;
  display: block;
  filter: drop-shadow(0 0 8px rgba(229, 9, 20, 0.6));
}

.logo-text {
  font-family: 'Barlow Condensed', sans-serif;
  font-size: 1.5rem;
  font-weight: 800;
  letter-spacing: 0.12em;
  color: #fff;
  background: linear-gradient(135deg, #fff 30%, rgba(229,9,20,0.9) 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

/* Nav Links */
.navbar-links {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 2px;
  flex: 1;
}

@media (min-width: 769px) {
  .navbar-links {
    position: absolute;
    left: 50%;
    transform: translateX(-50%);
  }
}

.nav-link {
  padding: 6px 16px;
  border-radius: var(--radius-md);
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--text-muted);
  text-decoration: none;
  letter-spacing: 0.03em;
  transition: all var(--transition-fast);
  position: relative;
}

@media (hover: hover) {
  .nav-link:hover {
    color: var(--text-primary);
    background: rgba(255, 255, 255, 0.05);
  }
}

.nav-link-active {
  color: var(--text-primary) !important;
  background: rgba(229, 9, 20, 0.1) !important;
}

.nav-link-active::after {
  content: '';
  position: absolute;
  bottom: -2px;
  left: 16px;
  right: 16px;
  height: 2px;
  background: var(--color-primary);
  border-radius: var(--radius-full);
  box-shadow: 0 0 8px var(--color-primary-glow);
}

/* User Area */
.navbar-user {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-shrink: 0;
  margin-left: auto;
}

.user-role-badge {
  padding: 3px 10px;
  border-radius: var(--radius-full);
  font-size: 0.68rem;
  font-weight: 700;
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.role-admin {
  background: rgba(229, 9, 20, 0.15);
  color: #f87171;
  border: 1px solid rgba(229, 9, 20, 0.25);
}

.role-user {
  background: rgba(99, 102, 241, 0.12);
  color: #a5b4fc;
  border: 1px solid rgba(99, 102, 241, 0.2);
}

/* Dropdown Container */
.user-profile-dropdown {
  position: relative;
  display: inline-block;
}

/* Trigger Button */
.profile-trigger {
  display: flex;
  align-items: center;
  gap: 10px;
  background: transparent;
  border: none;
  padding: 4px 10px;
  border-radius: var(--radius-full);
  color: var(--text-primary);
  font-family: inherit;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.profile-trigger:hover {
  background: transparent;
}

@media (hover: hover) {
  .profile-trigger:hover .user-avatar {
    border-color: var(--color-primary);
    box-shadow: 0 0 10px var(--color-primary-glow);
    transform: scale(1.05);
  }
}

.user-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  object-fit: cover;
  border: 1px solid rgba(255, 255, 255, 0.2);
  transition: all var(--transition-fast);
}.user-display-name {
  max-width: 280px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: flex;
  align-items: center;
}

.user-display-role {
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.05em;
  text-transform: uppercase;
}

.role-text-admin {
  color: #f87171;
}

.role-text-user {
  color: #a5b4fc;
}

.user-display-divider {
  color: rgba(255, 255, 255, 0.25);
  margin: 0 6px;
}

.user-display-name-text {
  font-weight: 600;
  color: var(--text-primary);
}

/* Submenu Container */
.dropdown-submenu-container {
  position: relative;
}

.submenu-trigger {
  display: flex;
  align-items: center;
  gap: 10px;
}

/* Submenu Panel positioning on the left side */
.submenu-left {
  position: absolute;
  top: 0;
  right: 100%;
  margin-right: 8px;
  width: 140px;
  background: rgba(13, 13, 25, 0.96);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: var(--radius-lg);
  box-shadow: -5px 10px 30px rgba(0, 0, 0, 0.5);
  padding: 6px 0;
  opacity: 0;
  visibility: hidden;
  transform: translateX(8px);
  transition: all var(--transition-fast);
  z-index: 210;
}

/* Hover effect to show the submenu */
.dropdown-submenu-container:hover .submenu-left {
  opacity: 1;
  visibility: visible;
  transform: translateX(0);
}

.dropdown-submenu-container:hover .submenu-arrow {
  color: var(--text-primary);
  transform: scale(1.1);
}

/* Submenu Item Styles */
.submenu-item {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  padding: 8px 14px;
  background: transparent;
  border: none;
  color: var(--text-muted);
  font-size: 0.8rem;
  font-weight: 600;
  text-align: left;
  font-family: inherit;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.submenu-item:hover {
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-primary);
}

.submenu-item.active {
  color: var(--color-primary);
  background: rgba(229, 9, 20, 0.08);
}
.dropdown-chevron {
  color: var(--text-muted);
  transition: transform var(--transition-fast);
}

.user-profile-dropdown:hover .dropdown-chevron,
.dropdown-open .dropdown-chevron {
  transform: rotate(180deg);
}

/* Dropdown Menu */
.dropdown-menu {
  position: absolute;
  top: calc(100% + 8px);
  right: 0;
  width: 180px;
  background: rgba(13, 13, 25, 0.96);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: var(--radius-lg);
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.5);
  padding: 6px 0;
  opacity: 0;
  visibility: hidden;
  transform: translateY(8px);
  transition: all var(--transition-fast);
  z-index: 200;
}

/* Show dropdown on hover for desktop */
@media (min-width: 769px) {
  .user-profile-dropdown:hover .dropdown-menu {
    opacity: 1;
    visibility: visible;
    transform: translateY(0);
  }
}

.dropdown-open .dropdown-menu {
  opacity: 1;
  visibility: visible;
  transform: translateY(0);
}

/* Header */
.dropdown-header {
  padding: 0 var(--space-md) var(--space-sm) var(--space-md);
}

.header-name {
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 2px;
}

.header-email {
  font-size: 0.78rem;
  color: var(--text-muted);
  margin-bottom: var(--space-xs);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.header-badge-row {
  display: flex;
  margin-top: 6px;
}

.dropdown-divider {
  height: 1px;
  background: rgba(255, 255, 255, 0.06);
  margin: var(--space-xs) 0;
}

/* Dropdown Items */
.dropdown-item {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 10px var(--space-md);
  background: transparent;
  border: none;
  color: var(--text-muted);
  font-size: 0.875rem;
  text-align: left;
  font-family: inherit;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.dropdown-item:hover {
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-primary);
}

.item-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  height: 16px;
}

.logout-item:hover {
  color: #f87171;
  background: rgba(239, 68, 68, 0.08);
}

@media (max-width: 768px) {
  .navbar-inner {
    padding: 0 var(--space-md);
    gap: var(--space-sm);
  }

  .navbar-links {
    justify-content: flex-start;
    flex-wrap: nowrap;
    overflow-x: auto;
    scrollbar-width: none;
    -ms-overflow-style: none;
    gap: 4px;
    padding: 4px 0;
  }

  .navbar-links::-webkit-scrollbar {
    display: none;
  }

  .nav-link {
    padding: 6px 12px;
    font-size: 0.82rem;
    white-space: nowrap;
  }

  .user-display-name {
    display: none;
  }
}
</style>
