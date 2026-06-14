package showtime

// ShowtimesResponse wraps a list of showtimes.
type ShowtimesResponse struct {
	Data  []Showtime `json:"data"`
	Total int        `json:"total"`
}
