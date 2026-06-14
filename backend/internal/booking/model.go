package booking

import (
	"time"

	"backend/internal/showtime"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// BookingStatus represents the lifecycle state of a booking.
type BookingStatus string

const (
	StatusPending   BookingStatus = "PENDING"
	StatusConfirmed BookingStatus = "CONFIRMED"
	StatusCancelled BookingStatus = "CANCELLED"
	StatusExpired   BookingStatus = "EXPIRED"
)

// BookedSeat is a snapshot of a seat at the time of booking.
// Mirrors showtime.Seat fields needed for receipt/display.
type BookedSeat struct {
	SeatLabel string  `bson:"seatLabel" json:"seatLabel"`
	Row       string  `bson:"row"       json:"row"`
	Number    int     `bson:"number"    json:"number"`
	Zone      string  `bson:"zone"      json:"zone"`
	Price     float64 `bson:"price"     json:"price"`
}

// Booking represents a completed or in-progress seat reservation.
type Booking struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      bson.ObjectID `bson:"userId"        json:"userId"`
	ShowtimeID  bson.ObjectID `bson:"showtimeId"    json:"showtimeId"`
	Seats       []BookedSeat  `bson:"seats"         json:"seats"`
	Status      BookingStatus `bson:"status"        json:"status"`
	TotalPrice  float64       `bson:"totalPrice"    json:"totalPrice"`
	UserEmail   string        `bson:"userEmail"     json:"userEmail"`
	UserName    string        `bson:"userName"      json:"userName"`
	MovieTitle  string        `bson:"movieTitle"    json:"movieTitle"`
	TheaterName string        `bson:"theaterName"   json:"theaterName"`
	StartTime   time.Time     `bson:"startTime"     json:"startTime"`
	CreatedAt   time.Time     `bson:"createdAt"     json:"createdAt"`
	UpdatedAt   time.Time     `bson:"updatedAt"     json:"updatedAt"`
}

// BookingWithDetails includes populated showtime and user for admin views.
type BookingWithDetails struct {
	Booking  `bson:",inline"`
	Showtime *showtime.Showtime `bson:"showtime,omitempty" json:"showtime,omitempty"`
}

// AuditLog records significant system events for traceability.
type AuditLog struct {
	ID         bson.ObjectID  `bson:"_id,omitempty" json:"id"`
	Event      string         `bson:"event"         json:"event"`
	UserID     *string        `bson:"userId"        json:"userId,omitempty"`
	UserEmail  *string        `bson:"userEmail"     json:"userEmail,omitempty"`
	ShowtimeID *string        `bson:"showtimeId"    json:"showtimeId,omitempty"`
	BookingID  *string        `bson:"bookingId"     json:"bookingId,omitempty"`
	SeatLabels []string       `bson:"seatLabels"    json:"seatLabels,omitempty"`
	Details    map[string]any `bson:"details"       json:"details,omitempty"`
	CreatedAt  time.Time      `bson:"createdAt"     json:"createdAt"`
}
