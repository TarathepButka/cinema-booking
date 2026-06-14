package scheduler

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"backend/internal/shared/mq"
	"backend/internal/shared/ws"
	"backend/internal/showtime"
)

type LockReleaser interface {
	ForceReleaseLock(ctx context.Context, showtimeID, seatLabel string) error
}

type AuditLogger interface {
	InsertEntry(ctx context.Context, entry mq.AuditLogEntry) error
}

// LockExpiryScheduler periodically releases expired MongoDB seat locks.
type LockExpiryScheduler struct {
	showtimeRepo *showtime.Repo
	lockSvc      LockReleaser
	auditLog     AuditLogger
	hub          *ws.Hub
	interval     time.Duration
}

func NewLockExpiryScheduler(
	showtimeRepo *showtime.Repo,
	lockSvc LockReleaser,
	auditLog AuditLogger,
	hub *ws.Hub,
) *LockExpiryScheduler {
	return &LockExpiryScheduler{
		showtimeRepo: showtimeRepo,
		lockSvc:      lockSvc,
		auditLog:     auditLog,
		hub:          hub,
		interval:     30 * time.Second,
	}
}

func (s *LockExpiryScheduler) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()
		log.Printf("lock expiry scheduler started (interval: %s)", s.interval)

		for {
			select {
			case <-ctx.Done():
				log.Println("lock expiry scheduler stopped")
				return
			case <-ticker.C:
				s.runOnce(ctx)
			}
		}
	}()
}

func (s *LockExpiryScheduler) runOnce(ctx context.Context) {
	showtimes, err := s.showtimeRepo.FindExpiredLockedSeats(ctx)
	if err != nil {
		log.Printf("lock expiry scan error: %v", err)
		return
	}

	now := time.Now()
	for _, st := range showtimes {
		for _, seat := range st.Seats {
			if seat.LockedBy == nil || seat.LockedUntil == nil || seat.LockedUntil.After(now) {
				continue
			}

			released, err := s.showtimeRepo.ReleaseExpiredSeat(ctx, st.ID, seat.SeatLabel, now)
			if err != nil {
				log.Printf("failed to release expired seat %s: %v", seat.SeatLabel, err)
				continue
			}
			if !released {
				continue
			}

			userID := *seat.LockedBy
			showtimeID := st.ID.Hex()
			_ = s.lockSvc.ForceReleaseLock(ctx, showtimeID, seat.SeatLabel)

			data, _ := json.Marshal(ws.SeatUpdateMessage{
				Type:       ws.MessageTypeSeatUpdate,
				ShowtimeID: showtimeID,
				SeatLabel:  seat.SeatLabel,
				Status:     string(showtime.SeatAvailable),
			})
			s.hub.Broadcast(showtimeID, data)

			entry := mq.AuditLogEntry{
				Event:      string(mq.EventBookingTimeout),
				UserID:     &userID,
				ShowtimeID: &showtimeID,
				SeatLabels: []string{seat.SeatLabel},
				Details:    map[string]any{"expiredAt": seat.LockedUntil.Format(time.RFC3339)},
				CreatedAt:  time.Now(),
			}
			if err := s.auditLog.InsertEntry(ctx, entry); err != nil {
				log.Printf("failed to write timeout audit log: %v", err)
			}
		}
	}
}
