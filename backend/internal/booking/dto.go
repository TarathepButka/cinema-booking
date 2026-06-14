package booking

// LockSeatRequest is the HTTP input for locking a seat.
type LockSeatRequest struct {
	ShowtimeID string `json:"showtimeId" binding:"required"`
	SeatLabel  string `json:"seatLabel"  binding:"required"`
}

// LockSeatResponse is returned after a successful lock.
type LockSeatResponse struct {
	SeatLabel  string `json:"seatLabel"`
	ShowtimeID string `json:"showtimeId"`
	ExpiresAt  string `json:"expiresAt"`
}

// ConfirmBookingRequest is the HTTP input for confirming a booking.
type ConfirmBookingRequest struct {
	ShowtimeID string   `json:"showtimeId" binding:"required"`
	SeatLabels []string `json:"seatLabels" binding:"required,min=1"`
}

// ReleaseSeatRequest is the HTTP input for releasing a locked seat.
type ReleaseSeatRequest struct {
	ShowtimeID string `json:"showtimeId" binding:"required"`
	SeatLabel  string `json:"seatLabel"  binding:"required"`
}
