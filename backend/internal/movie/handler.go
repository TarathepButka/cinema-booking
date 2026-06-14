package movie

import (
	"context"
	"net/http"
	"time"

	"backend/internal/shared/query"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	suggestionLimit    = 10
	suggestionMaxLimit = 50
)

// ShowtimeCreator is the showtime behavior needed when creating a movie.
type ShowtimeCreator interface {
	CreateShowtimesForMovie(ctx context.Context, movieID bson.ObjectID, duration int, halls []string, slots []string) error
}

// Handler handles movie-related HTTP endpoints.
type Handler struct {
	svc             *Service
	showtimeCreator ShowtimeCreator
}

// NewHandler creates a new movie Handler.
func NewHandler(svc *Service, showtimeCreator ShowtimeCreator) *Handler {
	return &Handler{svc: svc, showtimeCreator: showtimeCreator}
}

// GetAll returns all active movies.
// GET /api/movies
func (h *Handler) GetAll(c *gin.Context) {
	movies, err := h.svc.GetAllMovies(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": movies, "total": len(movies)})
}

// GetByID returns a single movie.
// GET /api/movies/:id
func (h *Handler) GetByID(c *gin.Context) {
	movie, err := h.svc.GetMovieByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": movie})
}

// GetSuggestions returns movie title recommendations based on search input.
// GET /api/movies/suggestions?q=&page=&limit=
func (h *Handler) GetSuggestions(c *gin.Context) {
	q := c.Query("q")
	page, limit := query.Pagination(c.Request.URL.Query(), suggestionLimit, suggestionMaxLimit)
	suggestions, total, err := h.svc.GetTitleSuggestions(c.Request.Context(), q, int64(page), int64(limit))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  suggestions,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// Create creates a new movie (admin only).
// POST /api/movies
func (h *Handler) Create(c *gin.Context) {
	var req CreateMovieRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	movie := &Movie{
		Title:       req.Title,
		Description: req.Description,
		Genre:       req.Genre,
		Duration:    req.Duration,
		Rating:      req.Rating,
		Language:    req.Language,
		Director:    req.Director,
		Cast:        req.Cast,
		PosterURL:   req.PosterURL,
		ReleaseDate: req.ReleaseDate,
		IsActive:    req.IsActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	created, err := h.svc.CreateMovie(c.Request.Context(), movie)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(req.Halls) > 0 && len(req.Slots) > 0 {
		if err := h.showtimeCreator.CreateShowtimesForMovie(
			c.Request.Context(),
			created.ID,
			created.Duration,
			req.Halls,
			req.Slots,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Movie created, but showtime generation failed: " + err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{"data": created})
}
