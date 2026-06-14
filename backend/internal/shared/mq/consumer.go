package mq

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// AuditLogEntry is shared by synchronous audit writers without importing a
// feature model into infrastructure packages.
type AuditLogEntry struct {
	Event      string
	UserID     *string
	UserEmail  *string
	ShowtimeID *string
	BookingID  *string
	SeatLabels []string
	Details    map[string]any
	CreatedAt  time.Time
}

type AuditLogInserter interface {
	InsertEntry(ctx context.Context, entry AuditLogEntry) error
}

type NotificationSender interface {
	SendBookingConfirmation(ctx context.Context, event BookingEvent) error
}

// Consumer subscribes to booking events for asynchronous notifications.
type Consumer struct {
	rdb          *redis.Client
	notification NotificationSender
}

func NewConsumer(rdb *redis.Client, notification NotificationSender) *Consumer {
	return &Consumer{rdb: rdb, notification: notification}
}

func (c *Consumer) Start(ctx context.Context) {
	go func() {
		pubsub := c.rdb.Subscribe(ctx, BookingEventChannel)
		defer pubsub.Close()

		log.Printf("MQ consumer listening on channel: %s", BookingEventChannel)
		ch := pubsub.Channel()
		for {
			select {
			case <-ctx.Done():
				log.Println("MQ consumer shutting down")
				return
			case msg, ok := <-ch:
				if !ok {
					log.Println("MQ channel closed")
					return
				}
				c.handleMessage(msg.Payload)
			}
		}
	}()
}

func (c *Consumer) handleMessage(payload string) {
	var event BookingEvent
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		log.Printf("MQ: failed to unmarshal event: %v", err)
		return
	}
	if event.Type != EventBookingSuccess || event.UserEmail == "" {
		return
	}

	go func() {
		if err := c.notification.SendBookingConfirmation(context.Background(), event); err != nil {
			log.Printf("notification failed for booking %s: %v", event.BookingID, err)
		}
	}()
}
