package booking

import (
	"strings"
	"testing"
	"time"

	"backend/internal/shared/mq"
	"backend/internal/showtime"
)

func TestNormalizeSeatLabelsRejectsDuplicates(t *testing.T) {
	_, err := normalizeSeatLabels([]string{"A1", " A1 "})
	if err == nil || !strings.Contains(err.Error(), "duplicate seat label") {
		t.Fatalf("normalizeSeatLabels() error = %v, want duplicate error", err)
	}
}

func TestValidateLockedSeatsRejectsExpiredLock(t *testing.T) {
	userID := "user-1"
	expiredAt := time.Now().Add(-time.Second)
	st := &showtime.Showtime{Seats: []showtime.Seat{{
		SeatLabel:   "A1",
		Status:      showtime.SeatLocked,
		LockedBy:    &userID,
		LockedUntil: &expiredAt,
	}}}

	_, _, err := validateLockedSeats(st, []string{"A1"}, userID, time.Now())
	if err == nil || !strings.Contains(err.Error(), "expired") {
		t.Fatalf("validateLockedSeats() error = %v, want expired lock error", err)
	}
}

func TestValidateLockedSeatsRejectsWrongOwner(t *testing.T) {
	owner := "user-1"
	expiresAt := time.Now().Add(time.Minute)
	st := &showtime.Showtime{Seats: []showtime.Seat{{
		SeatLabel:   "A1",
		Status:      showtime.SeatLocked,
		LockedBy:    &owner,
		LockedUntil: &expiresAt,
	}}}

	_, _, err := validateLockedSeats(st, []string{"A1"}, "user-2", time.Now())
	if err == nil || !strings.Contains(err.Error(), "not owned") {
		t.Fatalf("validateLockedSeats() error = %v, want ownership error", err)
	}
}

func TestValidateLockedSeatsUsesDatabasePrice(t *testing.T) {
	userID := "user-1"
	expiresAt := time.Now().Add(time.Minute)
	st := &showtime.Showtime{Seats: []showtime.Seat{{
		SeatLabel:   "A1",
		Row:         "A",
		Number:      1,
		Zone:        "FRONT",
		Price:       250,
		Status:      showtime.SeatLocked,
		LockedBy:    &userID,
		LockedUntil: &expiresAt,
	}}}

	seats, total, err := validateLockedSeats(st, []string{"A1"}, userID, time.Now())
	if err != nil {
		t.Fatalf("validateLockedSeats() error = %v", err)
	}
	if len(seats) != 1 || seats[0].Price != 250 || total != 250 {
		t.Fatalf("database seat metadata not preserved: seats=%#v total=%v", seats, total)
	}
}

func TestValidateLockedSeatsRejectsMultiSeatRequestBeforeMutation(t *testing.T) {
	userID := "user-1"
	future := time.Now().Add(time.Minute)
	past := time.Now().Add(-time.Minute)
	st := &showtime.Showtime{Seats: []showtime.Seat{
		{
			SeatLabel:   "A1",
			Status:      showtime.SeatLocked,
			LockedBy:    &userID,
			LockedUntil: &future,
		},
		{
			SeatLabel:   "A2",
			Status:      showtime.SeatLocked,
			LockedBy:    &userID,
			LockedUntil: &past,
		},
	}}

	_, _, err := validateLockedSeats(st, []string{"A1", "A2"}, userID, time.Now())
	if err == nil {
		t.Fatal("validateLockedSeats() succeeded with a partially invalid seat set")
	}
	if st.Seats[0].Status != showtime.SeatLocked || st.Seats[1].Status != showtime.SeatLocked {
		t.Fatal("validation mutated seat state before the transaction")
	}
}

func TestValidateLockedSeatsRejectsRepeatedConfirmation(t *testing.T) {
	st := &showtime.Showtime{Seats: []showtime.Seat{{
		SeatLabel: "A1",
		Status:    showtime.SeatBooked,
	}}}

	_, _, err := validateLockedSeats(st, []string{"A1"}, "user-1", time.Now())
	if err == nil {
		t.Fatal("validateLockedSeats() accepted an already booked seat")
	}
}

func TestNewAuditEntryIncludesRequiredFields(t *testing.T) {
	entry := newAuditEntry(
		mq.EventBookingSuccess,
		"user-1",
		"user@example.com",
		"showtime-1",
		"booking-1",
		[]string{"A1"},
		nil,
	)
	if entry.UserID == nil || *entry.UserID != "user-1" ||
		entry.BookingID == nil || *entry.BookingID != "booking-1" {
		t.Fatalf("newAuditEntry() omitted required identifiers: %#v", entry)
	}
}
