<script setup lang="ts">
import { RouterView } from 'vue-router'
import AdminSidebar from '@/components/AdminSidebar.vue'
</script>

<template>
  <div class="admin-page-container">
    <div class="admin-layout">
      <!-- Sidebar is persistent and outside of child page transition animations -->
      <AdminSidebar />
      
      <!-- Only the main detail content area undergoes transitions -->
      <RouterView v-slot="{ Component }">
        <Transition name="admin-fade" mode="out-in">
          <component :is="Component" />
        </Transition>
      </RouterView>
    </div>
  </div>
</template>

<style scoped>
/* Scoped transition for child views inside the persistent admin layout */
.admin-fade-enter-active,
.admin-fade-leave-active {
  transition: opacity 0.15s ease, transform 0.15s ease;
}
.admin-fade-enter-from {
  opacity: 0;
  transform: translateY(4px);
}
.admin-fade-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>
