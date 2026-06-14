type DateInput = string | Date

type SeatLike = string | { seatLabel: string }

const DEFAULT_LOCALE = 'en-US'
const DEFAULT_CURRENCY = 'THB'

export function formatCurrency(value: number, locale = DEFAULT_LOCALE, currency = DEFAULT_CURRENCY) {
  return new Intl.NumberFormat(locale, {
    style: 'currency',
    currency,
    maximumFractionDigits: 0,
  }).format(value)
}

export function formatDateTime(value: DateInput, options: Intl.DateTimeFormatOptions, locale = DEFAULT_LOCALE) {
  return new Date(value).toLocaleString(locale, options)
}

export function formatTime(value: DateInput, locale = DEFAULT_LOCALE) {
  return new Date(value).toLocaleTimeString(locale, {
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  })
}

export function formatDurationMinutes(durationMinutes: number) {
  const hours = Math.floor(durationMinutes / 60)
  const minutes = durationMinutes % 60

  return hours > 0 ? `${hours}h ${minutes}m` : `${minutes}m`
}

export function formatSeatLabels(seats: SeatLike[]) {
  return seats
    .map((seat) => (typeof seat === 'string' ? seat : seat.seatLabel))
    .join(', ')
}

export function formatBookingCode(id: string) {
  return id.substring(id.length - 8).toUpperCase()
}
