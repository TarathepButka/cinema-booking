package movie

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Movie represents a film available for booking.
type Movie struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Title       string        `bson:"title"         json:"title"`
	Description string        `bson:"description"   json:"description"`
	Genre       []string      `bson:"genre"         json:"genre"`
	Duration    int           `bson:"duration"      json:"duration"` // minutes
	Rating      float64       `bson:"rating"        json:"rating"`
	Language    string        `bson:"language"      json:"language"`
	Director    string        `bson:"director"      json:"director"`
	Cast        []string      `bson:"cast"          json:"cast"`
	PosterURL   string        `bson:"posterUrl"     json:"posterUrl"`
	ReleaseDate time.Time     `bson:"releaseDate"   json:"releaseDate"`
	IsActive    bool          `bson:"isActive"      json:"isActive"`
	CreatedAt   time.Time     `bson:"createdAt"     json:"createdAt"`
	UpdatedAt   time.Time     `bson:"updatedAt"     json:"updatedAt"`
}
