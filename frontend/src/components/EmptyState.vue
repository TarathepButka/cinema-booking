<script setup lang="ts">
defineProps<{
  title: string
}>()
</script>

<template>
  <div class="empty-state animate-fade-up">
    <p class="empty-state-title">{{ title }}</p>
    <div class="empty-state-actions" v-if="$slots.default">
      <slot />
    </div>
  </div>
</template>

<style scoped>
.empty-state-actions {
  margin-top: var(--space-md);
  display: flex;
  justify-content: center;
  gap: var(--space-sm);
}
</style>
