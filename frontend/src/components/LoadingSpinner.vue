<script setup lang="ts">
defineProps<{
  message?: string
  fullscreen?: boolean
}>()
</script>

<template>
  <div :class="['loading-state', { 'loading-fullscreen': fullscreen }]">
    <div class="spinner"></div>
    <span>{{ message || 'Loading...' }}</span>
  </div>
</template>

<style scoped>
.loading-fullscreen {
  position: fixed;
  inset: 0;
  background: rgba(8, 8, 15, 0.9);
  backdrop-filter: blur(8px);
  z-index: 9999;
  padding: 0;
}
</style>
