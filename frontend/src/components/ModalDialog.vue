<script setup lang="ts">
defineProps<{
  isOpen: boolean
  maxWidth?: string
}>()

defineEmits<{
  close: []
}>()
</script>

<template>
  <Teleport to="body">
    <Transition name="fade">
      <div class="modal-overlay" v-if="isOpen" @click.self="$emit('close')">
        <div 
          class="modal-content glass-card animate-scale-up" 
          :style="{ maxWidth: maxWidth || '450px' }"
        >
          <div class="modal-header">
            <slot name="header" />
            <button class="btn-close" @click="$emit('close')">✕</button>
          </div>
          
          <div class="divider" v-if="$slots.header"></div>
          
          <div class="modal-body">
            <slot />
          </div>
          
          <div class="divider" v-if="$slots.footer"></div>
          
          <div class="modal-footer" v-if="$slots.footer">
            <slot name="footer" />
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.75);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  width: 95%;
  background: var(--color-bg-card);
  border: 1px solid var(--border-card);
  border-radius: var(--radius-lg);
  padding: var(--space-xl);
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.5);
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
  max-height: 90vh;
  overflow: hidden;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-md);
  flex-shrink: 0;
}

.btn-close {
  background: transparent;
  border: none;
  color: var(--text-secondary);
  font-size: 1.2rem;
  cursor: pointer;
  padding: 4px;
  line-height: 1;
  transition: color var(--transition-fast);
}

.btn-close:hover {
  color: var(--color-primary);
}

.modal-body {
  display: flex;
  flex-direction: column;
  gap: 8px;
  flex: 1;
  overflow-y: auto;
  padding-right: 4px;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-sm);
  flex-shrink: 0;
}

.divider {
  height: 1px;
  background: var(--border-subtle);
  margin: var(--space-xs) 0;
  flex-shrink: 0;
}

/* Modal Transitions */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.25s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

@keyframes scale-up {
  from { transform: scale(0.95); opacity: 0; }
  to { transform: scale(1); opacity: 1; }
}

.animate-scale-up {
  animation: scale-up 0.25s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}
</style>
