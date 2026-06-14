import { ref, computed, onUnmounted } from 'vue'

export function useCountdown(durationSeconds: number) {
  const remaining = ref(durationSeconds)
  const isExpired = ref(false)
  let interval: ReturnType<typeof setInterval> | null = null

  const minutes = computed(() => Math.floor(remaining.value / 60))
  const seconds = computed(() => remaining.value % 60)

  const formatted = computed(() => {
    const m = String(minutes.value).padStart(2, '0')
    const s = String(seconds.value).padStart(2, '0')
    return `${m}:${s}`
  })

  const progress = computed(() => (remaining.value / durationSeconds) * 100)

  const urgency = computed(() => {
    if (remaining.value <= 30) return 'critical'
    if (remaining.value <= 60) return 'warning'
    return 'normal'
  })

  function start(targetExpiry?: Date) {
    if (targetExpiry) {
      // Sync to server expiry time
      const diff = Math.max(0, Math.floor((targetExpiry.getTime() - Date.now()) / 1000))
      remaining.value = diff
    }

    interval = setInterval(() => {
      if (remaining.value <= 0) {
        remaining.value = 0
        isExpired.value = true
        stop()
        return
      }
      remaining.value--
    }, 1000)
  }

  function stop() {
    if (interval) {
      clearInterval(interval)
      interval = null
    }
  }

  function reset(newDuration?: number) {
    stop()
    remaining.value = newDuration ?? durationSeconds
    isExpired.value = false
  }

  onUnmounted(stop)

  return { remaining, minutes, seconds, formatted, progress, urgency, isExpired, start, stop, reset }
}
