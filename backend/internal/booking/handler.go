package booking

import (
	"encoding/json"
	"net/http"
	"time"

	sharedauth "backend/internal/shared/auth"
	"backend/internal/shared/middleware"
	"backend/internal/shared/query"
	"backend/internal/shared/ws"
	"backend/internal/showtime"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	bookingPageLimit = 20
	auditPageLimit   = 50
	adminMaxLimit    = 100
)

// Handler handles seat locking and booking confirmation.
type Handler struct {
	svc          *Service
	auditLogRepo *AuditLogRepo
	hub          *ws.Hub
}

// NewHandler creates a new booking Handler.
func NewHandler(svc *Service, auditLogRepo *AuditLogRepo, hub *ws.Hub) *Handler {
	return &Handler{svc: svc, auditLogRepo: auditLogRepo, hub: hub}
}

// LockSeat handles Step 1-2 of the booking flow: lock a seat for the current user.
// POST /api/bookings/locks
func (h *Handler) LockSeat(c *gin.Context) {
	var req LockSeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims := middleware.GetCurrentUser(c)
	result, err := h.svc.LockSeatForUser(c.Request.Context(), req.ShowtimeID, req.SeatLabel, claims.UserID)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	h.broadcastSeatUpdate(req.ShowtimeID, req.SeatLabel, showtime.SeatLocked, claims.UserID)
	c.JSON(http.StatusOK, gin.H{"data": result})
}

// ConfirmBooking handles Step 3 of the booking flow: confirm and book the seat.
// POST /api/bookings
func (h *Handler) ConfirmBooking(c *gin.Context) {
	var req ConfirmBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims := middleware.GetCurrentUser(c)
	booking, err := h.svc.ConfirmBookingForUser(
		c.Request.Context(),
		req.ShowtimeID, req.SeatLabels,
		claims.UserID, claims.Email, claims.Email,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Broadcast BOOKED status for all confirmed seats
	for _, label := range req.SeatLabels {
		h.broadcastSeatUpdate(req.ShowtimeID, label, showtime.SeatBooked, claims.UserID)
	}

	c.JSON(http.StatusCreated, gin.H{"data": booking})
}

// ReleaseSeat releases a locked seat back to AVAILABLE.
// DELETE /api/bookings/locks
func (h *Handler) ReleaseSeat(c *gin.Context) {
	var req ReleaseSeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims := middleware.GetCurrentUser(c)
	released, err := h.svc.ReleaseSeat(c.Request.Context(), req.ShowtimeID, req.SeatLabel, claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !released {
		c.JSON(http.StatusConflict, gin.H{"error": "seat is not locked by you"})
		return
	}

	h.broadcastSeatUpdate(req.ShowtimeID, req.SeatLabel, showtime.SeatAvailable, "")
	c.JSON(http.StatusOK, gin.H{"message": "seat released"})
}

// MyBookings returns the authenticated user's booking history.
// GET /api/users/:id/bookings
func (h *Handler) MyBookings(c *gin.Context) {
	id := c.Param("id")
	claims := middleware.GetCurrentUser(c)

	targetUserID := id
	if id == "me" {
		targetUserID = claims.UserID
	} else if id != claims.UserID && claims.Role != sharedauth.RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	bookings, err := h.svc.GetUserBookings(c.Request.Context(), targetUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if bookings == nil {
		bookings = []Booking{}
	}

	c.JSON(http.StatusOK, gin.H{"data": bookings, "total": len(bookings)})
}

// GetAll returns all bookings with optional filters.
// GET /api/bookings?movie_title=&user_email=&status=&start_date=&end_date=&page=&limit=
func (h *Handler) GetAll(c *gin.Context) {
	filter := buildBookingsFilter(c)
	page, limit := query.Pagination(c.Request.URL.Query(), bookingPageLimit, adminMaxLimit)

	bookings, total, err := h.svc.GetAllBookings(c.Request.Context(), filter, int64(page), int64(limit))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  bookings,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetStats returns booking statistics for the admin dashboard.
// GET /api/bookings/stats?movie_title=&user_email=&start_date=&end_date=
func (h *Handler) GetStats(c *gin.Context) {
	stats, err := h.svc.GetBookingStats(c.Request.Context(), buildBookingsFilter(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": stats})
}

// GetAuditLogs returns audit logs with optional event filtering.
// GET /api/audit-logs?event=&page=&limit=
func (h *Handler) GetAuditLogs(c *gin.Context) {
	page, limit := query.Pagination(c.Request.URL.Query(), auditPageLimit, adminMaxLimit)

	logs, total, err := h.auditLogRepo.FindAll(
		c.Request.Context(),
		c.Query("event"),
		int64(page),
		int64(limit),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  logs,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func buildBookingsFilter(c *gin.Context) bson.M {
	filter := bson.M{}
	if title := c.Query("movie_title"); title != "" {
		filter["movieTitle"] = bson.M{"$regex": title, "$options": "i"}
	}
	if email := c.Query("user_email"); email != "" {
		filter["userEmail"] = bson.M{"$regex": email, "$options": "i"}
	}
	if status := c.Query("status"); status != "" {
		filter["status"] = status
	}

	ictZone := time.FixedZone("ICT", 7*60*60)
	if start, err := time.Parse("2006-01-02", c.Query("start_date")); err == nil {
		filter["createdAt"] = bson.M{
			"$gte": time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, ictZone),
		}
	}
	if end, err := time.Parse("2006-01-02", c.Query("end_date")); err == nil {
		dateFilter, ok := filter["createdAt"].(bson.M)
		if !ok {
			dateFilter = bson.M{}
			filter["createdAt"] = dateFilter
		}
		dateFilter["$lte"] = time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 999999999, ictZone)
	}
	return filter
}

// broadcastSeatUpdate sends a WebSocket message to all clients watching this showtime.
func (h *Handler) broadcastSeatUpdate(showtimeID, seatLabel string, status showtime.SeatStatus, updatedBy string) {
	msg := ws.SeatUpdateMessage{
		Type:       ws.MessageTypeSeatUpdate,
		ShowtimeID: showtimeID,
		SeatLabel:  seatLabel,
		Status:     string(status),
		UpdatedBy:  updatedBy,
	}
	data, _ := json.Marshal(msg)
	h.hub.Broadcast(showtimeID, data)
}
