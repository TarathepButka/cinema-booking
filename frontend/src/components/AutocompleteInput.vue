<script setup lang="ts">
defineProps<{
  modelValue: string
  placeholder?: string
  suggestions: string[]
  showSuggestions: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [val: string]
  'update:showSuggestions': [val: boolean]
  'select': [val: string]
  'enter': [val: string]
}>()

function handleBlur() {
  setTimeout(() => {
    emit('update:showSuggestions', false)
  }, 200)
}
</script>

<template>
  <div class="search-wrap autocomplete-wrap">
    <svg class="search-icon-svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
    <input
      type="text"
      class="search-input"
      :placeholder="placeholder"
      :value="modelValue"
      @input="emit('update:modelValue', ($event.target as HTMLInputElement).value)"
      @focus="emit('update:showSuggestions', true)"
      @blur="handleBlur"
      @keydown.enter="emit('enter', modelValue)"
    />
    <div v-if="showSuggestions && suggestions.length > 0" class="suggestions-dropdown glass-card">
      <div
        v-for="item in suggestions"
        :key="item"
        class="suggestion-item"
        @mousedown="emit('select', item)"
      >
        {{ item }}
      </div>
    </div>
  </div>
</template>

<style scoped>
.search-wrap {
  position: relative;
  width: 100%;
}

.search-icon-svg {
  position: absolute;
  left: 14px;
  top: 50%;
  transform: translateY(-50%);
  color: var(--text-muted);
  pointer-events: none;
  z-index: 10;
}

.search-input {
  width: 100%;
  padding: 11px 14px 11px 44px;
  background: var(--color-bg-card);
  border: 1px solid var(--border-card);
  border-radius: var(--radius-md);
  color: var(--text-primary);
  font-family: 'Inter', sans-serif;
  font-size: 0.9rem;
  transition: all var(--transition-fast);
  outline: none;
}

.search-input:focus {
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px var(--color-primary-subtle);
  background: var(--color-bg-elevated);
}

.search-input::placeholder { color: var(--text-muted); }

.autocomplete-wrap {
  position: relative;
}

.suggestions-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  right: 0;
  z-index: 100;
  max-height: 200px;
  overflow-y: auto;
  padding: 6px 0;
  background: var(--color-bg-elevated);
  border: 1px solid var(--border-card);
  box-shadow: var(--shadow-lift);
}

.suggestion-item {
  padding: 10px 16px;
  font-size: 0.88rem;
  color: var(--text-secondary);
  cursor: pointer;
  text-align: left;
  transition: all var(--transition-fast);
}

.suggestion-item:hover {
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-primary);
}
</style>
