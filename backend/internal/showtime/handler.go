package showtime

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler handles showtime-related HTTP endpoints.
type Handler struct {
	svc *Service
}

// NewHandler creates a new showtime Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// GetByMovieID returns showtimes for a movie.
// GET /api/showtimes?movie_id=xxx
func (h *Handler) GetByMovieID(c *gin.Context) {
	movieID := c.Query("movie_id")
	if movieID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "movie_id query parameter required"})
		return
	}

	showtimes, err := h.svc.GetByMovieID(c.Request.Context(), movieID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": showtimes, "total": len(showtimes)})
}

// GetByID returns a showtime with full seat map.
// GET /api/showtimes/:id
func (h *Handler) GetByID(c *gin.Context) {
	showtime, err := h.svc.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": showtime})
}
