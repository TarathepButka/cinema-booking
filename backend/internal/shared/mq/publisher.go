package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

const BookingEventChannel = "booking_events"

// EventType defines the type of booking event.
type EventType string

const (
	EventBookingSuccess EventType = "BOOKING_SUCCESS"
	EventBookingTimeout EventType = "BOOKING_TIMEOUT"
	EventSeatReleased   EventType = "SEAT_RELEASED"
	EventSeatLocked     EventType = "SEAT_LOCKED"
	EventSystemError    EventType = "SYSTEM_ERROR"
)

// BookingEvent is the message structure published to Redis Pub/Sub.
type BookingEvent struct {
	Type         EventType      `json:"type"`
	UserID       string         `json:"userId,omitempty"`
	UserEmail    string         `json:"userEmail,omitempty"`
	UserName     string         `json:"userName,omitempty"`
	ShowtimeID   string         `json:"showtimeId,omitempty"`
	BookingID    string         `json:"bookingId,omitempty"`
	MovieTitle   string         `json:"movieTitle,omitempty"`
	TheaterName  string         `json:"theaterName,omitempty"`
	ShowtimeTime time.Time      `json:"showtimeTime,omitempty"`
	SeatLabels   []string       `json:"seatLabels,omitempty"`
	TotalPrice   float64        `json:"totalPrice,omitempty"`
	Details      map[string]any `json:"details,omitempty"`
}

// Publisher publishes booking events to Redis Pub/Sub.
type Publisher struct {
	rdb *redis.Client
}

// NewPublisher creates a new event Publisher.
func NewPublisher(rdb *redis.Client) *Publisher {
	return &Publisher{rdb: rdb}
}

// Publish serializes and publishes a booking event.
func (p *Publisher) Publish(ctx context.Context, event BookingEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	if err := p.rdb.Publish(ctx, BookingEventChannel, data).Err(); err != nil {
		log.Printf("⚠️  Failed to publish event %s: %v", event.Type, err)
		return fmt.Errorf("publish event: %w", err)
	}

	log.Printf("📤 Published event: %s (booking=%s)", event.Type, event.BookingID)
	return nil
}
