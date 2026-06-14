<script setup lang="ts">
import { RouterLink, useRoute } from 'vue-router'

const route = useRoute()

const isActive = (path: string) => {
  return route.path === path
}
</script>

<template>
  <aside class="admin-sidebar glass-card">
    <nav class="sidebar-nav">
      <RouterLink
        to="/manage/bookings"
        class="nav-item"
        :class="{ active: isActive('/manage/bookings') }"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="7" height="9"></rect><rect x="14" y="3" width="7" height="5"></rect><rect x="14" y="12" width="7" height="9"></rect><rect x="3" y="16" width="7" height="5"></rect></svg>
        <span>Bookings</span>
      </RouterLink>

      <RouterLink
        to="/manage/movies-showtimes"
        class="nav-item"
        :class="{ active: isActive('/manage/movies-showtimes') }"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="2" width="20" height="20" rx="2.18" ry="2.18"></rect><line x1="7" y1="2" x2="7" y2="22"></line><line x1="17" y1="2" x2="17" y2="22"></line><line x1="2" y1="12" x2="22" y2="12"></line></svg>
        <span>Movies<span class="desktop-only"> &amp; Showtimes</span></span>
      </RouterLink>

      <RouterLink
        to="/manage/audit-logs"
        class="nav-item"
        :class="{ active: isActive('/manage/audit-logs') }"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path><polyline points="14 2 14 8 20 8"></polyline><line x1="16" y1="13" x2="8" y2="13"></line><line x1="16" y1="17" x2="8" y2="17"></line><polyline points="10 9 9 9 8 9"></polyline></svg>
        <span>Audit Logs</span>
      </RouterLink>
    </nav>


  </aside>
</template>

<style scoped>
.admin-sidebar {
  display: flex;
  flex-direction: column;
  gap: var(--space-xl);
  padding: var(--space-lg);
  border-radius: var(--radius-lg);
  height: max-content;
  position: sticky;
  top: calc(70px + var(--space-xl));
  border: 1px solid var(--color-glass-border);
}

.logo {
  display: flex;
  align-items: center;
  gap: 10px;
  font-family: 'Outfit', sans-serif;
  font-weight: 800;
  font-size: 1.1rem;
  color: var(--text-primary);
}

.logo-icon {
  color: var(--color-primary);
}

.sidebar-nav {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  color: var(--text-secondary);
  border-radius: var(--radius-md);
  font-weight: 500;
  transition: all var(--transition-base);
}

.nav-item svg {
  transition: transform var(--transition-base);
}

@media (hover: hover) {
  .nav-item:hover {
    background: var(--color-bg-card-hover);
    color: var(--text-primary);
  }

  .nav-item:hover svg {
    transform: translateX(2px);
  }
}

.nav-item.active {
  background: var(--color-primary-glow);
  color: var(--text-primary);
  border-left: 3px solid var(--color-primary);
}

.w-full {
  width: 100%;
}

.desktop-only {
  display: inline;
}

/* Mobile responsive sidebar rules */
@media (max-width: 992px) {
  .admin-sidebar {
    position: sticky !important;
    top: 64px !important;
    left: 0;
    right: 0;
    width: 100% !important;
    height: auto !important;
    border-radius: 0 !important;
    border: none !important;
    border-bottom: 1px solid rgba(255, 255, 255, 0.08) !important;
    background: rgba(8, 8, 15, 0.95) !important;
    backdrop-filter: blur(20px) !important;
    -webkit-backdrop-filter: blur(20px) !important;
    z-index: 90 !important;
    padding: var(--space-xs) var(--space-md) !important;
    box-shadow: none !important;
    margin-bottom: var(--space-md);
  }

  .sidebar-nav {
    flex-direction: row !important;
    justify-content: space-around !important;
    gap: var(--space-xs) !important;
    width: 100%;
  }

  .nav-item {
    padding: 10px 12px !important;
    font-size: 0.8rem !important;
    flex: 1 !important;
    justify-content: center !important;
    gap: 8px !important;
  }

  .nav-item.active {
    border-left: none !important;
    border-bottom: 3px solid var(--color-primary) !important;
    border-radius: var(--radius-sm) var(--radius-sm) 0 0 !important;
    background: rgba(229, 9, 20, 0.08) !important;
  }
}

@media (max-width: 800px) {
  .desktop-only {
    display: none;
  }
  .nav-item {
    gap: 6px !important;
    padding: 10px 8px !important;
  }
}

/* Desktop Stuck-to-Left styling overrides */
@media (min-width: 993px) {
  .admin-sidebar {
    position: fixed !important;
    top: 64px !important;
    left: 0 !important;
    bottom: 0 !important;
    width: 260px !important;
    height: calc(100vh - 64px) !important;
    border-radius: 0 !important;
    border: none !important;
    border-right: 1px solid rgba(255, 255, 255, 0.08) !important;
    background: rgba(8, 8, 15, 0.95) !important;
    backdrop-filter: blur(20px) !important;
    -webkit-backdrop-filter: blur(20px) !important;
    z-index: 90 !important;
    padding: var(--space-xl) var(--space-lg) !important;
  }

  .admin-sidebar:hover {
    transform: none !important;
    box-shadow: none !important;
    background: rgba(8, 8, 15, 0.95) !important;
    border-right: 1px solid rgba(255, 255, 255, 0.08) !important;
  }
}
</style>
