import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest'
import { defineComponent } from 'vue'
import { mount } from '@vue/test-utils'
import { useCountdown } from '@/composables/useCountdown'

function runInComponent(fn: () => any) {
  let result: any
  const TestComponent = defineComponent({
    setup() {
      result = fn()
      return () => null
    }
  })
  const wrapper = mount(TestComponent)
  return { result, wrapper }
}

describe('useCountdown composable', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('initializes with correct defaults and computed properties', () => {
    const { result } = runInComponent(() => useCountdown(120))
    const { remaining, minutes, seconds, formatted, progress, urgency, isExpired } = result

    expect(remaining.value).toBe(120)
    expect(isExpired.value).toBe(false)
    expect(minutes.value).toBe(2)
    expect(seconds.value).toBe(0)
    expect(formatted.value).toBe('02:00')
    expect(progress.value).toBe(100)
    expect(urgency.value).toBe('normal')
  })

  it('calculates urgency levels correctly', () => {
    const { result } = runInComponent(() => useCountdown(120))

    // normal
    expect(result.urgency.value).toBe('normal')

    // warning (<= 60s)
    result.remaining.value = 60
    expect(result.urgency.value).toBe('warning')
    result.remaining.value = 31
    expect(result.urgency.value).toBe('warning')

    // critical (<= 30s)
    result.remaining.value = 30
    expect(result.urgency.value).toBe('critical')
    result.remaining.value = 0
    expect(result.urgency.value).toBe('critical')
  })

  it('decrements remaining seconds when started', () => {
    const { result } = runInComponent(() => useCountdown(60))
    result.start()

    expect(result.remaining.value).toBe(60)

    // Advance 5 seconds
    vi.advanceTimersByTime(5000)
    expect(result.remaining.value).toBe(55)
    expect(result.formatted.value).toBe('00:55')
    expect(result.progress.value).toBe((55 / 60) * 100)
  })

  it('expires and stops when remaining hits 0', () => {
    const { result } = runInComponent(() => useCountdown(2))
    result.start()

    expect(result.isExpired.value).toBe(false)

    vi.advanceTimersByTime(2000)
    expect(result.remaining.value).toBe(0)
    expect(result.isExpired.value).toBe(false) // It hits 0 on the second tick, but checks <= 0 on the third tick

    vi.advanceTimersByTime(1000)
    expect(result.isExpired.value).toBe(true)

    // Advance more to check that it has stopped decrementing
    vi.advanceTimersByTime(2000)
    expect(result.remaining.value).toBe(0)
  })

  it('stops decrementing when stop() is called', () => {
    const { result } = runInComponent(() => useCountdown(60))
    result.start()

    vi.advanceTimersByTime(5000)
    expect(result.remaining.value).toBe(55)

    result.stop()

    vi.advanceTimersByTime(5000)
    // Should stay at 55
    expect(result.remaining.value).toBe(55)
  })

  it('resets correctly with current or new duration', () => {
    const { result } = runInComponent(() => useCountdown(60))
    result.start()

    vi.advanceTimersByTime(5000)
    expect(result.remaining.value).toBe(55)

    result.reset()
    expect(result.remaining.value).toBe(60)
    expect(result.isExpired.value).toBe(false)

    // Reset with new duration
    result.reset(10)
    expect(result.remaining.value).toBe(10)
  })

  it('synchronizes target expiry on start(targetExpiry)', () => {
    const { result } = runInComponent(() => useCountdown(60))

    // Set a date 15 seconds in the future
    const futureDate = new Date(Date.now() + 15000)
    result.start(futureDate)

    expect(result.remaining.value).toBe(15)
  })

  it('stops the interval when the component is unmounted', () => {
    const { result, wrapper } = runInComponent(() => useCountdown(60))
    result.start()

    vi.advanceTimersByTime(5000)
    expect(result.remaining.value).toBe(55)

    wrapper.unmount()

    vi.advanceTimersByTime(5000)
    // Should not decrement after unmounting
    expect(result.remaining.value).toBe(55)
  })
})
