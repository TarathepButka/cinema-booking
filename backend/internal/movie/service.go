package movie

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Service handles movie business logic.
type Service struct {
	repo *Repo
}

// NewService creates a new movie Service.
func NewService(repo *Repo) *Service {
	return &Service{repo: repo}
}

// GetAllMovies returns all active movies.
func (s *Service) GetAllMovies(ctx context.Context) ([]Movie, error) {
	return s.repo.FindAll(ctx)
}

// GetMovieByID returns a movie or error if not found.
func (s *Service) GetMovieByID(ctx context.Context, id string) (*Movie, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid movie id: %w", err)
	}
	movie, err := s.repo.FindByID(ctx, oid)
	if err != nil {
		return nil, err
	}
	if movie == nil {
		return nil, fmt.Errorf("movie not found")
	}
	return movie, nil
}

// GetTitleSuggestions returns paginated movie titles matching the query.
func (s *Service) GetTitleSuggestions(ctx context.Context, query string, page, limit int64) ([]string, int64, error) {
	return s.repo.FindTitleSuggestions(ctx, query, page, limit)
}

// CreateMovie creates a new movie (admin only).
func (s *Service) CreateMovie(ctx context.Context, movie *Movie) (*Movie, error) {
	return s.repo.Create(ctx, movie)
}

// UpdateMovie modifies an existing movie (admin only).
func (s *Service) UpdateMovie(ctx context.Context, id string, update bson.M) error {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid movie id: %w", err)
	}
	return s.repo.Update(ctx, oid, update)
}

// DeleteMovie soft-deletes a movie (admin only).
func (s *Service) DeleteMovie(ctx context.Context, id string) error {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid movie id: %w", err)
	}
	return s.repo.Delete(ctx, oid)
}
