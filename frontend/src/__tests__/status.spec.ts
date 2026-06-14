import { describe, expect, it } from 'vitest'
import { getBookingStatusPresentation, getAuditEventBadgeClass } from '@/utils/status'
import { AUDIT_EVENT, BADGE_CLASS, BOOKING_STATUS } from '@/constants'

describe('status utilities', () => {
  describe('getBookingStatusPresentation', () => {
    it('returns presentation for CONFIRMED status with default options', () => {
      const presentation = getBookingStatusPresentation(BOOKING_STATUS.CONFIRMED)
      expect(presentation).toEqual({
        label: 'CONFIRMED',
        badgeClass: BADGE_CLASS.SUCCESS,
        ticketClass: 'status-confirmed',
      })
    })

    it('returns presentation for CONFIRMED status with custom confirmedLabel', () => {
      const presentation = getBookingStatusPresentation(BOOKING_STATUS.CONFIRMED, {
        confirmedLabel: 'Paid & Confirmed',
      })
      expect(presentation.label).toBe('Paid & Confirmed')
      expect(presentation.badgeClass).toBe(BADGE_CLASS.SUCCESS)
      expect(presentation.ticketClass).toBe('status-confirmed')
    })

    it('returns presentation for PENDING status', () => {
      const presentation = getBookingStatusPresentation(BOOKING_STATUS.PENDING)
      expect(presentation).toEqual({
        label: 'PENDING',
        badgeClass: BADGE_CLASS.WARNING,
        ticketClass: 'status-pending',
      })
    })

    it('returns presentation for CANCELLED status', () => {
      const presentation = getBookingStatusPresentation(BOOKING_STATUS.CANCELLED)
      expect(presentation).toEqual({
        label: 'CANCELLED',
        badgeClass: BADGE_CLASS.DANGER,
        ticketClass: 'status-cancelled',
      })
    })

    it('returns presentation for EXPIRED status', () => {
      const presentation = getBookingStatusPresentation(BOOKING_STATUS.EXPIRED)
      expect(presentation).toEqual({
        label: 'EXPIRED',
        badgeClass: BADGE_CLASS.MUTED,
        ticketClass: 'status-expired',
      })
    })

    it('returns fallback presentation for unknown status', () => {
      const presentation = getBookingStatusPresentation('UNKNOWN_STATUS')
      expect(presentation).toEqual({
        label: 'UNKNOWN_STATUS',
        badgeClass: BADGE_CLASS.MUTED,
        ticketClass: 'status-expired',
      })
    })
  })

  describe('getAuditEventBadgeClass', () => {
    it('maps AUDIT_EVENT.BOOKING_SUCCESS to BADGE_CLASS.SUCCESS', () => {
      expect(getAuditEventBadgeClass(AUDIT_EVENT.BOOKING_SUCCESS)).toBe(BADGE_CLASS.SUCCESS)
    })

    it('maps AUDIT_EVENT.BOOKING_TIMEOUT to BADGE_CLASS.WARNING', () => {
      expect(getAuditEventBadgeClass(AUDIT_EVENT.BOOKING_TIMEOUT)).toBe(BADGE_CLASS.WARNING)
    })

    it('maps AUDIT_EVENT.SEAT_RELEASED to BADGE_CLASS.INFO', () => {
      expect(getAuditEventBadgeClass(AUDIT_EVENT.SEAT_RELEASED)).toBe(BADGE_CLASS.INFO)
    })

    it('maps AUDIT_EVENT.SEAT_LOCKED to BADGE_CLASS.ACCENT', () => {
      expect(getAuditEventBadgeClass(AUDIT_EVENT.SEAT_LOCKED)).toBe(BADGE_CLASS.ACCENT)
    })

    it('maps AUDIT_EVENT.SYSTEM_ERROR to BADGE_CLASS.DANGER', () => {
      expect(getAuditEventBadgeClass(AUDIT_EVENT.SYSTEM_ERROR)).toBe(BADGE_CLASS.DANGER)
    })

    it('maps unknown audit events to BADGE_CLASS.MUTED', () => {
      expect(getAuditEventBadgeClass('UNKNOWN_EVENT')).toBe(BADGE_CLASS.MUTED)
    })
  })
})
