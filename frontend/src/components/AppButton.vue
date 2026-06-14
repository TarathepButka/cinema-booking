<script setup lang="ts">
withDefaults(
  defineProps<{
    variant?: 'primary' | 'secondary' | 'accent' | 'danger' | 'ghost'
    size?: 'sm' | 'md' | 'lg'
    loading?: boolean
    disabled?: boolean
    type?: 'button' | 'submit' | 'reset'
  }>(),
  {
    variant: 'primary',
    size: 'md',
    loading: false,
    disabled: false,
    type: 'button',
  }
)
</script>

<template>
  <button
    :type="type"
    :class="['btn', `btn-${variant}`, size !== 'md' ? `btn-${size}` : '']"
    :disabled="disabled || loading"
  >
    <span v-if="loading" class="btn-spinner"></span>
    <slot v-else />
  </button>
</template>

<style scoped>
.btn-spinner {
  display: inline-block;
  width: 1.25em;
  height: 1.25em;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: currentColor;
  border-radius: 50%;
  animation: spin-loading 0.8s linear infinite;
}

@keyframes spin-loading {
  to {
    transform: rotate(360deg);
  }
}
</style>
