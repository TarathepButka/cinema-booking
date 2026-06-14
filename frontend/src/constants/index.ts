// ─── Auth / Roles ────────────────────────────────────────────────────────────
export const ROLES = {
  ADMIN: 'ADMIN',
  USER: 'USER',
} as const

export type Role = (typeof ROLES)[keyof typeof ROLES]

// ─── LocalStorage Keys ───────────────────────────────────────────────────────
// ─── Booking Status ──────────────────────────────────────────────────────────
export const BOOKING_STATUS = {
  PENDING: 'PENDING',
  CONFIRMED: 'CONFIRMED',
  CANCELLED: 'CANCELLED',
  EXPIRED: 'EXPIRED',
} as const

export type BookingStatus = (typeof BOOKING_STATUS)[keyof typeof BOOKING_STATUS]

// ─── Seat Status ─────────────────────────────────────────────────────────────
export const SEAT_STATUS = {
  AVAILABLE: 'AVAILABLE',
  LOCKED: 'LOCKED',
  BOOKED: 'BOOKED',
} as const

export type SeatStatus = (typeof SEAT_STATUS)[keyof typeof SEAT_STATUS]

// ─── Audit Event Types ───────────────────────────────────────────────────────
export const AUDIT_EVENT = {
  BOOKING_SUCCESS: 'BOOKING_SUCCESS',
  BOOKING_TIMEOUT: 'BOOKING_TIMEOUT',
  SEAT_LOCKED: 'SEAT_LOCKED',
  SEAT_RELEASED: 'SEAT_RELEASED',
  SYSTEM_ERROR: 'SYSTEM_ERROR',
} as const

export type AuditEvent = (typeof AUDIT_EVENT)[keyof typeof AUDIT_EVENT]

// ─── Badge CSS Classes ───────────────────────────────────────────────────────
export const BADGE_CLASS = {
  SUCCESS: 'badge-success',
  WARNING: 'badge-warning',
  DANGER: 'badge-danger',
  MUTED: 'badge-muted',
  ACCENT: 'badge-accent',
  INFO: 'badge-info',
} as const

// ─── Pagination ───────────────────────────────────────────────────────────────
export const PAGINATION = {
  DEFAULT_PAGE: 1,
  DEFAULT_LIMIT: 20,
  PAGE_SIZE_OPTIONS: [10, 20, 50, 100] as const,
} as const

// ─── Cinema Config ────────────────────────────────────────────────────────────
export const CINEMA_HALLS = [
  { value: 'Hall A', label: 'Hall A (Premium)' },
  { value: 'Hall B', label: 'Hall B (Standard)' },
] as const

export const SHOWTIME_SLOTS = [
  { value: 'Morning',   label: 'Morning (10:00)' },
  { value: 'Afternoon', label: 'Afternoon (13:30)' },
  { value: 'Evening',   label: 'Evening (17:00)' },
  { value: 'Night',     label: 'Night (20:30)' },
] as const

// ─── API ──────────────────────────────────────────────────────────────────────
export const API_TIMEOUT_MS = 15_000
