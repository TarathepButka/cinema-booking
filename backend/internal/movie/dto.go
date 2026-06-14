package movie

import "time"

// CreateMovieRequest is the request body for POST /api/movies.
type CreateMovieRequest struct {
	Title       string    `json:"title" binding:"required"`
	Description string    `json:"description"`
	Genre       []string  `json:"genre"`
	Duration    int       `json:"duration"`
	Rating      float64   `json:"rating"`
	Language    string    `json:"language"`
	Director    string    `json:"director"`
	Cast        []string  `json:"cast"`
	PosterURL   string    `json:"posterUrl"`
	ReleaseDate time.Time `json:"releaseDate"`
	IsActive    bool      `json:"isActive"`
	Halls       []string  `json:"halls"`
	Slots       []string  `json:"slots"`
}
