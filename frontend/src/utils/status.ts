import { AUDIT_EVENT, BADGE_CLASS, BOOKING_STATUS } from '@/constants'

const BOOKING_TICKET_CLASS = {
  [BOOKING_STATUS.CONFIRMED]: 'status-confirmed',
  [BOOKING_STATUS.PENDING]: 'status-pending',
  [BOOKING_STATUS.CANCELLED]: 'status-cancelled',
  [BOOKING_STATUS.EXPIRED]: 'status-expired',
} as const

const BOOKING_BADGE_CLASS = {
  [BOOKING_STATUS.CONFIRMED]: BADGE_CLASS.SUCCESS,
  [BOOKING_STATUS.PENDING]: BADGE_CLASS.WARNING,
  [BOOKING_STATUS.CANCELLED]: BADGE_CLASS.DANGER,
  [BOOKING_STATUS.EXPIRED]: BADGE_CLASS.MUTED,
} as const

export function getBookingStatusPresentation(
  status: string,
  options?: {
    confirmedLabel?: string
  },
) {
  const confirmedLabel = options?.confirmedLabel ?? BOOKING_STATUS.CONFIRMED

  switch (status) {
    case BOOKING_STATUS.CONFIRMED:
      return {
        label: confirmedLabel,
        badgeClass: BOOKING_BADGE_CLASS[BOOKING_STATUS.CONFIRMED],
        ticketClass: BOOKING_TICKET_CLASS[BOOKING_STATUS.CONFIRMED],
      }
    case BOOKING_STATUS.PENDING:
      return {
        label: BOOKING_STATUS.PENDING,
        badgeClass: BOOKING_BADGE_CLASS[BOOKING_STATUS.PENDING],
        ticketClass: BOOKING_TICKET_CLASS[BOOKING_STATUS.PENDING],
      }
    case BOOKING_STATUS.CANCELLED:
      return {
        label: BOOKING_STATUS.CANCELLED,
        badgeClass: BOOKING_BADGE_CLASS[BOOKING_STATUS.CANCELLED],
        ticketClass: BOOKING_TICKET_CLASS[BOOKING_STATUS.CANCELLED],
      }
    case BOOKING_STATUS.EXPIRED:
      return {
        label: BOOKING_STATUS.EXPIRED,
        badgeClass: BOOKING_BADGE_CLASS[BOOKING_STATUS.EXPIRED],
        ticketClass: BOOKING_TICKET_CLASS[BOOKING_STATUS.EXPIRED],
      }
    default:
      return {
        label: status,
        badgeClass: BADGE_CLASS.MUTED,
        ticketClass: BOOKING_TICKET_CLASS[BOOKING_STATUS.EXPIRED],
      }
  }
}

export function getAuditEventBadgeClass(event: string) {
  switch (event) {
    case AUDIT_EVENT.BOOKING_SUCCESS:
      return BADGE_CLASS.SUCCESS
    case AUDIT_EVENT.BOOKING_TIMEOUT:
      return BADGE_CLASS.WARNING
    case AUDIT_EVENT.SEAT_RELEASED:
      return BADGE_CLASS.INFO
    case AUDIT_EVENT.SEAT_LOCKED:
      return BADGE_CLASS.ACCENT
    case AUDIT_EVENT.SYSTEM_ERROR:
      return BADGE_CLASS.DANGER
    default:
      return BADGE_CLASS.MUTED
  }
}
