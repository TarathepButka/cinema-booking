<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  expiresAt: string | null
  remainingSeconds: number
  urgency: 'normal' | 'warning' | 'critical'
}>()

const progressPct = computed(() => {
  const total = 5 * 60 // 5 minutes
  return Math.max(0, (props.remainingSeconds / total) * 100)
})

const formatted = computed(() => {
  const m = String(Math.floor(props.remainingSeconds / 60)).padStart(2, '0')
  const s = String(props.remainingSeconds % 60).padStart(2, '0')
  return `${m}:${s}`
})

const colorClass = computed(() => {
  if (props.urgency === 'critical') return 'critical'
  if (props.urgency === 'warning') return 'warning'
  return 'normal'
})
</script>

<template>
  <div class="countdown-wrap" :class="colorClass">
    <div class="countdown-header">
      <span class="clock-icon">⏱️</span>
      <span class="countdown-label">Time Remaining</span>
    </div>

    <div class="countdown-time" :class="{ pulse: urgency === 'critical' }">
      {{ formatted }}
    </div>

    <div class="progress-bar">
      <div
        class="progress-fill"
        :style="{ width: progressPct + '%' }"
      ></div>
    </div>

    <p v-if="urgency === 'critical'" class="urgency-msg">
      ⚠️ Hurry! Your seat reservation is about to expire!
    </p>
    <p v-else-if="urgency === 'warning'" class="urgency-msg">
      Complete your booking soon
    </p>
  </div>
</template>

<style scoped>
.countdown-wrap {
  background: var(--color-glass);
  border: 1px solid var(--color-glass-border);
  border-radius: var(--radius-lg);
  padding: var(--space-lg);
  text-align: center;
  transition: all var(--transition-base);
}

.countdown-wrap.warning {
  border-color: rgba(245, 158, 11, 0.3);
  background: rgba(245, 158, 11, 0.05);
}

.countdown-wrap.critical {
  border-color: rgba(239, 68, 68, 0.4);
  background: rgba(239, 68, 68, 0.08);
  animation: pulse-border 1s ease-in-out infinite;
}

@keyframes pulse-border {
  0%, 100% { border-color: rgba(239, 68, 68, 0.4); }
  50%       { border-color: rgba(239, 68, 68, 0.8); }
}

.countdown-header {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-sm);
  margin-bottom: var(--space-sm);
}

.clock-icon { font-size: 1.2rem; }

.countdown-label {
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--text-muted);
}

.countdown-time {
  font-family: 'Outfit', sans-serif;
  font-size: 3rem;
  font-weight: 800;
  color: var(--text-primary);
  line-height: 1;
  margin-bottom: var(--space-md);
  letter-spacing: 0.05em;
}

.normal .countdown-time  { color: var(--seat-available); }
.warning .countdown-time { color: var(--seat-locked-mine); }
.critical .countdown-time { color: var(--seat-locked); }

.countdown-time.pulse {
  animation: pulse-text 0.5s ease-in-out infinite alternate;
}

@keyframes pulse-text {
  from { opacity: 1; }
  to   { opacity: 0.6; }
}

.progress-bar {
  height: 6px;
  background: rgba(255,255,255,0.08);
  border-radius: var(--radius-full);
  overflow: hidden;
  margin-bottom: var(--space-sm);
}

.progress-fill {
  height: 100%;
  border-radius: var(--radius-full);
  transition: width 1s linear, background 0.5s ease;
}

.normal .progress-fill   { background: var(--seat-available); }
.warning .progress-fill  { background: var(--seat-locked-mine); }
.critical .progress-fill { background: var(--seat-locked); }

.urgency-msg {
  font-size: 0.8rem;
  color: var(--text-secondary);
  margin-top: var(--space-sm);
}
</style>
