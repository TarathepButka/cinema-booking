package email

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"time"

	"backend/internal/shared/mq"

	"github.com/resend/resend-go/v2"
)

// NotificationService sends email notifications for booking events.
// Uses Resend if RESEND_API_KEY is set, otherwise mocks the send (logs + no error).
type NotificationService struct {
	fromEmail string
	client    *resend.Client
}

// NewNotificationService creates a new NotificationService.
func NewNotificationService(resendAPIKey, fromEmail string) *NotificationService {
	var client *resend.Client
	if resendAPIKey != "" {
		client = resend.NewClient(resendAPIKey)
	}
	return &NotificationService{fromEmail: fromEmail, client: client}
}

//go:embed booking.html
var bookingEmailHTML string

// bookingEmailTemplate is the HTML email template for booking confirmation.
var bookingEmailTemplate = template.Must(template.New("booking").Parse(bookingEmailHTML))

// SendBookingConfirmation sends a booking confirmation email.
// Implements mq.NotificationSender interface.
func (s *NotificationService) SendBookingConfirmation(ctx context.Context, event mq.BookingEvent) error {
	// Convert ShowtimeTime to ICT (Asia/Bangkok timezone) to show correct local time
	ict := time.FixedZone("ICT", 7*60*60)
	event.ShowtimeTime = event.ShowtimeTime.In(ict)

	// Render HTML
	var buf bytes.Buffer
	if err := bookingEmailTemplate.Execute(&buf, event); err != nil {
		return fmt.Errorf("render email template: %w", err)
	}
	htmlBody := buf.String()

	// If no Resend API key, mock the send
	if s.client == nil {
		log.Printf("📧 [MOCK EMAIL] To: %s | Subject: Booking Confirmed — %s | Seats: %v | Total: ฿%.2f",
			event.UserEmail, event.MovieTitle, event.SeatLabels, event.TotalPrice)
		return nil
	}

	// Send via Resend
	params := &resend.SendEmailRequest{
		From:    s.fromEmail,
		To:      []string{event.UserEmail},
		Subject: fmt.Sprintf("🎬 Booking Confirmed — %s", event.MovieTitle),
		Html:    htmlBody,
	}

	resp, err := s.client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("resend send: %w", err)
	}

	log.Printf("📧 Email sent via Resend: %s → %s", resp.Id, event.UserEmail)
	return nil
}
