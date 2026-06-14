package movie

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"backend/internal/shared/db"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Repo handles all movie persistence operations.
type Repo struct {
	col *mongo.Collection
}

// NewRepo creates a new movie Repo.
func NewRepo(database *db.MongoDB) *Repo {
	return &Repo{col: database.DB.Collection("movies")}
}

// FindAll returns all active movies, sorted by release date descending.
func (r *Repo) FindAll(ctx context.Context) ([]Movie, error) {
	opts := options.Find().SetSort(bson.D{{Key: "releaseDate", Value: -1}})
	cursor, err := r.col.Find(ctx, bson.M{"isActive": true}, opts)
	if err != nil {
		return nil, fmt.Errorf("find movies: %w", err)
	}
	defer cursor.Close(ctx)

	var movies []Movie
	if err := cursor.All(ctx, &movies); err != nil {
		return nil, fmt.Errorf("decode movies: %w", err)
	}
	return movies, nil
}

// FindTitleSuggestions returns paginated movie titles matching the query.
func (r *Repo) FindTitleSuggestions(ctx context.Context, query string, page, limit int64) ([]string, int64, error) {
	filter := bson.M{}
	if query != "" {
		filter["title"] = bson.M{"$regex": regexp.QuoteMeta(query), "$options": "i"}
	}

	total, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count movie suggestions: %w", err)
	}

	opts := options.Find().
		SetProjection(bson.M{"title": 1}).
		SetSort(bson.D{{Key: "title", Value: 1}}).
		SetSkip((page - 1) * limit).
		SetLimit(limit)
	cursor, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("find movie suggestions: %w", err)
	}
	defer cursor.Close(ctx)

	var results []struct {
		Title string `bson:"title"`
	}
	if err := cursor.All(ctx, &results); err != nil {
		return nil, 0, fmt.Errorf("decode movie suggestions: %w", err)
	}

	suggestions := make([]string, 0, len(results))
	for _, result := range results {
		if result.Title != "" {
			suggestions = append(suggestions, result.Title)
		}
	}
	return suggestions, total, nil
}

// FindByID returns a movie by its ObjectID.
func (r *Repo) FindByID(ctx context.Context, id bson.ObjectID) (*Movie, error) {
	var movie Movie
	err := r.col.FindOne(ctx, bson.M{"_id": id, "isActive": true}).Decode(&movie)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	return &movie, err
}

// Create inserts a new movie.
func (r *Repo) Create(ctx context.Context, movie *Movie) (*Movie, error) {
	movie.ID = bson.NewObjectID()
	now := time.Now()
	movie.CreatedAt = now
	movie.UpdatedAt = now
	movie.IsActive = true

	if _, err := r.col.InsertOne(ctx, movie); err != nil {
		return nil, fmt.Errorf("insert movie: %w", err)
	}
	return movie, nil
}

// Update modifies an existing movie by ID.
func (r *Repo) Update(ctx context.Context, id bson.ObjectID, update bson.M) error {
	update["updatedAt"] = time.Now()
	_, err := r.col.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": update})
	return err
}

// Delete soft-deletes a movie by setting isActive = false.
func (r *Repo) Delete(ctx context.Context, id bson.ObjectID) error {
	_, err := r.col.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"isActive": false, "updatedAt": time.Now()}},
	)
	return err
}
