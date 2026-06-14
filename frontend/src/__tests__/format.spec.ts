import { describe, expect, it } from 'vitest'

import {
  formatBookingCode,
  formatCurrency,
  formatDateTime,
  formatDurationMinutes,
  formatSeatLabels,
  formatTime,
} from '@/utils/format'

describe('format utilities', () => {
  it.each([
    [0, '0m'],
    [45, '45m'],
    [60, '1h 0m'],
    [135, '2h 15m'],
  ])('formats %i minutes', (minutes, expected) => {
    expect(formatDurationMinutes(minutes)).toBe(expected)
  })

  it('formats mixed seat values and booking codes', () => {
    expect(formatSeatLabels(['A1', { seatLabel: 'A2' }])).toBe('A1, A2')
    expect(formatSeatLabels([])).toBe('')
    expect(formatBookingCode('booking-abcdef12')).toBe('ABCDEF12')
  })

  it('formats currency correctly', () => {
    const formatted = formatCurrency(150)
    expect(typeof formatted).toBe('string')
    expect(formatted).toContain('150')
  })

  it('formats date and time', () => {
    const date = new Date('2026-06-14T15:30:00Z')
    // Use UTC to avoid local timezone variance in test runners
    const dateStr = formatDateTime(date, { year: 'numeric', month: 'numeric', day: 'numeric', timeZone: 'UTC' })
    expect(dateStr).toContain('2026')

    const timeStr = formatTime(date)
    expect(timeStr).toMatch(/^\d{2}:\d{2}$/)
  })
})

