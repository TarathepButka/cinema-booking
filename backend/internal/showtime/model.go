package showtime

import (
	"time"

	"backend/internal/movie"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// SeatStatus represents the current booking state of a seat.
type SeatStatus string

const (
	SeatAvailable SeatStatus = "AVAILABLE"
	SeatLocked    SeatStatus = "LOCKED"
	SeatBooked    SeatStatus = "BOOKED"
)

// Seat is an embedded document inside a Showtime.
type Seat struct {
	Row         string     `bson:"row"         json:"row"`
	Number      int        `bson:"number"      json:"number"`
	SeatLabel   string     `bson:"seatLabel"   json:"seatLabel"` // e.g., "A1", "B3"
	Zone        string     `bson:"zone"        json:"zone"`      // FRONT, MIDDLE, BACK
	Price       float64    `bson:"price"       json:"price"`
	Status      SeatStatus `bson:"status"      json:"status"`
	LockedBy    *string    `bson:"lockedBy"    json:"lockedBy,omitempty"`    // user ID
	LockedUntil *time.Time `bson:"lockedUntil" json:"lockedUntil,omitempty"` // lock expiry
}

// Zone defines a seating zone within a theater.
type Zone struct {
	Name        string   `bson:"name"        json:"name"`        // FRONT, MIDDLE, BACK
	Rows        []string `bson:"rows"        json:"rows"`
	SeatsPerRow int      `bson:"seatsPerRow" json:"seatsPerRow"`
	Price       float64  `bson:"price"       json:"price"`
}

// Theater represents a cinema hall with its seating layout.
type Theater struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string        `bson:"name"          json:"name"`
	Description string        `bson:"description"   json:"description"`
	TotalSeats  int           `bson:"totalSeats"    json:"totalSeats"`
	Zones       []Zone        `bson:"zones"         json:"zones"`
	CreatedAt   time.Time     `bson:"createdAt"     json:"createdAt"`
}

// Showtime represents a scheduled movie screening in a specific theater.
type Showtime struct {
	ID          bson.ObjectID      `bson:"_id,omitempty" json:"id"`
	MovieID     bson.ObjectID      `bson:"movieId"       json:"movieId"`
	TheaterID   bson.ObjectID      `bson:"theaterId"     json:"theaterId"`
	TheaterName string             `bson:"theaterName"   json:"theaterName"`
	StartTime   time.Time          `bson:"startTime"     json:"startTime"`
	EndTime     time.Time          `bson:"endTime"       json:"endTime"`
	SlotLabel   string             `bson:"slotLabel"     json:"slotLabel"` // Morning, Afternoon, Evening, Night
	PriceByZone map[string]float64 `bson:"priceByZone"   json:"priceByZone"`
	Seats       []Seat             `bson:"seats"         json:"seats"`
	CreatedAt   time.Time          `bson:"createdAt"     json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt"     json:"updatedAt"`
}

// ShowtimeWithMovie is used when returning showtime data with populated movie info.
type ShowtimeWithMovie struct {
	Showtime `bson:",inline"`
	Movie    *movie.Movie `bson:"movie,omitempty" json:"movie,omitempty"`
}
