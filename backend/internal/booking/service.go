package booking

import (
	"context"
	"fmt"
	"strings"
	"time"

	"backend/internal/movie"
	"backend/internal/shared/mq"
	"backend/internal/showtime"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// ShowtimeRepository is the interface booking/Service needs from the showtime package.
// Using an interface here makes the dependency explicit and testable.
// main.go wires *showtime.Repo to this interface.
type ShowtimeRepository interface {
	FindByID(ctx context.Context, id bson.ObjectID) (*showtime.Showtime, error)
	LockSeat(ctx context.Context, showtimeID bson.ObjectID, seatLabel, userID string, lockUntil interface{ IsZero() bool }) (bool, error)
	ConfirmSeat(ctx context.Context, showtimeID bson.ObjectID, seatLabel, userID string, confirmedAt time.Time) (bool, error)
	ReleaseSeat(ctx context.Context, showtimeID bson.ObjectID, seatLabel, userID string) (bool, error)
}

type MovieRepository interface {
	FindByID(ctx context.Context, id bson.ObjectID) (*movie.Movie, error)
}

type LockManager interface {
	AcquireLock(ctx context.Context, showtimeID, seatLabel, userID string) (bool, error)
	GetLockOwner(ctx context.Context, showtimeID, seatLabel string) (string, error)
	GetLockTTL(ctx context.Context, showtimeID, seatLabel string) (time.Duration, error)
	ReleaseLock(ctx context.Context, showtimeID, seatLabel, userID string) error
	ForceReleaseLock(ctx context.Context, showtimeID, seatLabel string) error
}

type AuditLogger interface {
	InsertEntry(ctx context.Context, entry mq.AuditLogEntry) error
}

type confirmSeatInput struct {
	ShowtimeID string
	SeatLabels []string
	UserID     string
	UserEmail  string
	UserName   string
}

// Service orchestrates the full booking flow:
// lock → confirm → release, with event publishing for async processing.
type Service struct {
	showtimeRepo *showtime.Repo
	movieRepo    MovieRepository
	bookingRepo  *Repo
	lockSvc      LockManager
	publisher    *mq.Publisher
	auditLog     AuditLogger
}

// NewService creates a new booking Service.
func NewService(
	showtimeRepo *showtime.Repo,
	movieRepo MovieRepository,
	bookingRepo *Repo,
	lockSvc LockManager,
	publisher *mq.Publisher,
	auditLog AuditLogger,
) *Service {
	return &Service{
		showtimeRepo: showtimeRepo,
		movieRepo:    movieRepo,
		bookingRepo:  bookingRepo,
		lockSvc:      lockSvc,
		publisher:    publisher,
		auditLog:     auditLog,
	}
}

// lockSeatParams groups inputs for the LockSeat service method.
type lockSeatParams struct {
	ShowtimeID string
	SeatLabel  string
	UserID     string
}

// LockSeat implements Step 2 of the booking flow.
// 1. Try Redis lock (SET NX EX) — fast path, prevents race conditions
// 2. Try MongoDB findOneAndUpdate — safety net for double booking
func (s *Service) LockSeat(ctx context.Context, req lockSeatParams) (*LockSeatResponse, error) {
	// Step 1: Redis distributed lock
	acquired, err := s.lockSvc.AcquireLock(ctx, req.ShowtimeID, req.SeatLabel, req.UserID)
	if err != nil {
		s.writeAudit(ctx, mq.EventSystemError, req.UserID, "", req.ShowtimeID, "", []string{req.SeatLabel},
			map[string]any{"error": err.Error(), "step": "redis_lock"})
		return nil, fmt.Errorf("lock service error: %w", err)
	}
	if !acquired {
		return nil, fmt.Errorf("seat %s is not available", req.SeatLabel)
	}

	lockUntil := LockExpiry()

	// Step 2: MongoDB atomic update (double-lock safety net)
	showtimeOID, err := bson.ObjectIDFromHex(req.ShowtimeID)
	if err != nil {
		_ = s.lockSvc.ForceReleaseLock(ctx, req.ShowtimeID, req.SeatLabel)
		return nil, fmt.Errorf("invalid showtime id: %w", err)
	}

	locked, err := s.showtimeRepo.LockSeat(ctx, showtimeOID, req.SeatLabel, req.UserID, lockUntil)
	if err != nil {
		_ = s.lockSvc.ForceReleaseLock(ctx, req.ShowtimeID, req.SeatLabel)
		return nil, fmt.Errorf("database lock failed: %w", err)
	}
	if !locked {
		_ = s.lockSvc.ForceReleaseLock(ctx, req.ShowtimeID, req.SeatLabel)
		return nil, fmt.Errorf("seat %s is not available", req.SeatLabel)
	}

	s.writeAudit(ctx, mq.EventSeatLocked, req.UserID, "", req.ShowtimeID, "", []string{req.SeatLabel}, nil)

	return &LockSeatResponse{
		SeatLabel:  req.SeatLabel,
		ShowtimeID: req.ShowtimeID,
		ExpiresAt:  lockUntil.UTC().Format("2006-01-02T15:04:05Z"),
	}, nil
}

// LockSeatForUser is the public-facing method called from the handler.
func (s *Service) LockSeatForUser(ctx context.Context, showtimeID, seatLabel, userID string) (*LockSeatResponse, error) {
	return s.LockSeat(ctx, lockSeatParams{
		ShowtimeID: showtimeID,
		SeatLabel:  seatLabel,
		UserID:     userID,
	})
}

// ConfirmBooking confirms all locked seats and creates a booking record.
func (s *Service) ConfirmBooking(ctx context.Context, req confirmSeatInput) (*Booking, error) {
	showtimeOID, err := bson.ObjectIDFromHex(req.ShowtimeID)
	if err != nil {
		return nil, fmt.Errorf("invalid showtime id: %w", err)
	}
	userOID, err := bson.ObjectIDFromHex(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}
	seatLabels, err := normalizeSeatLabels(req.SeatLabels)
	if err != nil {
		return nil, err
	}

	for _, label := range seatLabels {
		owner, ownerErr := s.lockSvc.GetLockOwner(ctx, req.ShowtimeID, label)
		if ownerErr != nil {
			return nil, fmt.Errorf("verify lock owner for seat %s: %w", label, ownerErr)
		}
		ttl, ttlErr := s.lockSvc.GetLockTTL(ctx, req.ShowtimeID, label)
		if ttlErr != nil {
			return nil, fmt.Errorf("verify lock expiry for seat %s: %w", label, ttlErr)
		}
		if owner != req.UserID || ttl <= 0 {
			return nil, fmt.Errorf("seat %s lock has expired or is not owned by you", label)
		}
	}

	var movieTitle string
	var theaterName string
	var startTime time.Time
	var totalPrice float64
	created, err := s.bookingRepo.WithTransaction(ctx, func(txCtx context.Context) (*Booking, error) {
		confirmedAt := time.Now()
		st, findErr := s.showtimeRepo.FindByID(txCtx, showtimeOID)
		if findErr != nil {
			return nil, fmt.Errorf("find showtime: %w", findErr)
		}
		if st == nil {
			return nil, fmt.Errorf("showtime not found")
		}
		mv, movieErr := s.movieRepo.FindByID(txCtx, st.MovieID)
		if movieErr != nil {
			return nil, fmt.Errorf("find movie: %w", movieErr)
		}
		if mv == nil {
			return nil, fmt.Errorf("movie not found")
		}

		bookedSeats, price, validateErr := validateLockedSeats(st, seatLabels, req.UserID, confirmedAt)
		if validateErr != nil {
			return nil, validateErr
		}
		for _, label := range seatLabels {
			confirmed, confirmErr := s.showtimeRepo.ConfirmSeat(txCtx, showtimeOID, label, req.UserID, confirmedAt)
			if confirmErr != nil {
				return nil, fmt.Errorf("confirm seat %s failed: %w", label, confirmErr)
			}
			if !confirmed {
				return nil, fmt.Errorf("seat %s confirmation failed (lock may have expired)", label)
			}
		}

		movieTitle = mv.Title
		theaterName = st.TheaterName
		startTime = st.StartTime
		totalPrice = price
		createdBooking, createErr := s.bookingRepo.Create(txCtx, &Booking{
			UserID:      userOID,
			ShowtimeID:  showtimeOID,
			Seats:       bookedSeats,
			Status:      StatusConfirmed,
			TotalPrice:  totalPrice,
			UserEmail:   req.UserEmail,
			UserName:    req.UserName,
			MovieTitle:  movieTitle,
			TheaterName: theaterName,
			StartTime:   startTime,
		})
		if createErr != nil {
			return nil, createErr
		}
		if auditErr := s.auditLog.InsertEntry(txCtx, newAuditEntry(
			mq.EventBookingSuccess,
			req.UserID,
			req.UserEmail,
			req.ShowtimeID,
			createdBooking.ID.Hex(),
			seatLabels,
			nil,
		)); auditErr != nil {
			return nil, fmt.Errorf("write booking audit log: %w", auditErr)
		}
		return createdBooking, nil
	})
	if err != nil {
		return nil, err
	}

	for _, label := range seatLabels {
		_ = s.lockSvc.ReleaseLock(ctx, req.ShowtimeID, label, req.UserID)
	}

	_ = s.publisher.Publish(ctx, mq.BookingEvent{
		Type:         mq.EventBookingSuccess,
		UserID:       req.UserID,
		UserEmail:    req.UserEmail,
		UserName:     req.UserName,
		ShowtimeID:   req.ShowtimeID,
		BookingID:    created.ID.Hex(),
		MovieTitle:   movieTitle,
		TheaterName:  theaterName,
		ShowtimeTime: startTime,
		SeatLabels:   seatLabels,
		TotalPrice:   totalPrice,
	})

	return created, nil
}

// ConfirmBookingForUser is the public-facing method called from the handler.
func (s *Service) ConfirmBookingForUser(ctx context.Context, showtimeID string, seatLabels []string, userID, userEmail, userName string) (*Booking, error) {
	return s.ConfirmBooking(ctx, confirmSeatInput{
		ShowtimeID: showtimeID,
		SeatLabels: seatLabels,
		UserID:     userID,
		UserEmail:  userEmail,
		UserName:   userName,
	})
}

// ReleaseSeat releases a locked seat back to AVAILABLE.
func (s *Service) ReleaseSeat(ctx context.Context, showtimeID, seatLabel, userID string) (bool, error) {
	showtimeOID, err := bson.ObjectIDFromHex(showtimeID)
	if err != nil {
		return false, fmt.Errorf("invalid showtime id: %w", err)
	}

	released, err := s.showtimeRepo.ReleaseSeat(ctx, showtimeOID, seatLabel, userID)
	if err != nil {
		return false, fmt.Errorf("release seat in db: %w", err)
	}

	if released {
		_ = s.lockSvc.ReleaseLock(ctx, showtimeID, seatLabel, userID)
		s.writeAudit(ctx, mq.EventSeatReleased, userID, "", showtimeID, "", []string{seatLabel}, nil)
	}

	return released, nil
}

// GetUserBookings returns all bookings for the given user.
func (s *Service) GetUserBookings(ctx context.Context, userID string) ([]Booking, error) {
	oid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}
	bookings, err := s.bookingRepo.FindByUserID(ctx, oid)
	if err != nil {
		return nil, err
	}

	for i := range bookings {
		if bookings[i].MovieTitle != "" && bookings[i].TheaterName != "" && !bookings[i].StartTime.IsZero() {
			continue
		}

		st, findErr := s.showtimeRepo.FindByID(ctx, bookings[i].ShowtimeID)
		if findErr != nil || st == nil {
			continue
		}
		if bookings[i].TheaterName == "" {
			bookings[i].TheaterName = st.TheaterName
		}
		if bookings[i].StartTime.IsZero() {
			bookings[i].StartTime = st.StartTime
		}
		if bookings[i].MovieTitle == "" {
			mv, movieErr := s.movieRepo.FindByID(ctx, st.MovieID)
			if movieErr == nil && mv != nil {
				bookings[i].MovieTitle = mv.Title
			}
		}
	}

	return bookings, nil
}

// GetAllBookings returns paginated bookings matching the supplied filters.
func (s *Service) GetAllBookings(ctx context.Context, filter bson.M, page, limit int64) ([]Booking, int64, error) {
	return s.bookingRepo.FindAll(ctx, filter, page, limit)
}

// GetBookingStats aggregates confirmed booking counts and revenue by movie.
func (s *Service) GetBookingStats(ctx context.Context, filter bson.M) ([]bson.M, error) {
	return s.bookingRepo.AggregateByMovie(ctx, filter)
}

func normalizeSeatLabels(labels []string) ([]string, error) {
	seen := make(map[string]struct{}, len(labels))
	normalized := make([]string, 0, len(labels))
	for _, raw := range labels {
		label := strings.TrimSpace(raw)
		if label == "" {
			return nil, fmt.Errorf("seat label cannot be empty")
		}
		if _, exists := seen[label]; exists {
			return nil, fmt.Errorf("duplicate seat label %s", label)
		}
		seen[label] = struct{}{}
		normalized = append(normalized, label)
	}
	return normalized, nil
}

func validateLockedSeats(st *showtime.Showtime, labels []string, userID string, confirmedAt time.Time) ([]BookedSeat, float64, error) {
	seatMap := make(map[string]showtime.Seat, len(st.Seats))
	for _, seat := range st.Seats {
		seatMap[seat.SeatLabel] = seat
	}

	bookedSeats := make([]BookedSeat, 0, len(labels))
	var totalPrice float64
	for _, label := range labels {
		seat, ok := seatMap[label]
		if !ok {
			return nil, 0, fmt.Errorf("seat %s not found in showtime", label)
		}
		if seat.Status != showtime.SeatLocked || seat.LockedBy == nil || *seat.LockedBy != userID ||
			seat.LockedUntil == nil || !seat.LockedUntil.After(confirmedAt) {
			return nil, 0, fmt.Errorf("seat %s lock has expired or is not owned by you", label)
		}
		bookedSeats = append(bookedSeats, BookedSeat{
			SeatLabel: seat.SeatLabel,
			Row:       seat.Row,
			Number:    seat.Number,
			Zone:      seat.Zone,
			Price:     seat.Price,
		})
		totalPrice += seat.Price
	}
	return bookedSeats, totalPrice, nil
}

func (s *Service) writeAudit(ctx context.Context, event mq.EventType, userID, userEmail, showtimeID, bookingID string, seats []string, details map[string]any) {
	if err := s.auditLog.InsertEntry(ctx, newAuditEntry(event, userID, userEmail, showtimeID, bookingID, seats, details)); err != nil {
		fmt.Printf("audit log write failed for %s: %v\n", event, err)
	}
}

func newAuditEntry(event mq.EventType, userID, userEmail, showtimeID, bookingID string, seats []string, details map[string]any) mq.AuditLogEntry {
	entry := mq.AuditLogEntry{
		Event:      string(event),
		SeatLabels: seats,
		Details:    details,
		CreatedAt:  time.Now(),
	}
	if userID != "" {
		entry.UserID = &userID
	}
	if userEmail != "" {
		entry.UserEmail = &userEmail
	}
	if showtimeID != "" {
		entry.ShowtimeID = &showtimeID
	}
	if bookingID != "" {
		entry.BookingID = &bookingID
	}
	return entry
}
