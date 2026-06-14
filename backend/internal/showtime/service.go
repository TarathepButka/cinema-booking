package showtime

import (
	"context"
	"fmt"
	"time"

	"backend/internal/movie"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// MovieFinder is the interface used by ShowtimeService to fetch movie data.
// Defined here to avoid tight coupling to movie.Repo — main.go wires the concrete type.
type MovieFinder interface {
	FindByID(ctx context.Context, id bson.ObjectID) (*movie.Movie, error)
}

// Service handles showtime business logic.
type Service struct {
	repo        *Repo
	theaterRepo *TheaterRepo
	movieRepo   MovieFinder
}

// NewService creates a new showtime Service.
func NewService(repo *Repo, theaterRepo *TheaterRepo, movieRepo MovieFinder) *Service {
	return &Service{repo: repo, theaterRepo: theaterRepo, movieRepo: movieRepo}
}

// GetByMovieID returns upcoming showtimes for a given movie.
func (s *Service) GetByMovieID(ctx context.Context, movieID string) ([]Showtime, error) {
	oid, err := bson.ObjectIDFromHex(movieID)
	if err != nil {
		return nil, fmt.Errorf("invalid movie id: %w", err)
	}
	return s.repo.FindByMovieID(ctx, oid)
}

// GetByID returns a showtime with full seat details, including populated movie data.
func (s *Service) GetByID(ctx context.Context, id string) (*ShowtimeWithMovie, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid showtime id: %w", err)
	}

	showtime, err := s.repo.FindByID(ctx, oid)
	if err != nil {
		return nil, err
	}
	if showtime == nil {
		return nil, fmt.Errorf("showtime not found")
	}

	// Populate movie data
	mv, err := s.movieRepo.FindByID(ctx, showtime.MovieID)
	if err != nil {
		return nil, err
	}

	return &ShowtimeWithMovie{Showtime: *showtime, Movie: mv}, nil
}

// GenerateSeats creates the seat layouts dynamically from the theater's zones.
func GenerateSeats(zones []Zone) []Seat {
	var seats []Seat
	for _, z := range zones {
		for _, row := range z.Rows {
			for col := 1; col <= z.SeatsPerRow; col++ {
				seats = append(seats, Seat{
					Row:       row,
					Number:    col,
					SeatLabel: fmt.Sprintf("%s%d", row, col),
					Zone:      z.Name,
					Price:     z.Price,
					Status:    SeatAvailable,
				})
			}
		}
	}
	return seats
}

// CreateShowtimesForMovie generates showtimes for a movie across selected halls and slots for the next 5 days.
func (s *Service) CreateShowtimesForMovie(ctx context.Context, movieID bson.ObjectID, duration int, halls []string, slots []string) error {
	theaters, err := s.theaterRepo.FindAll(ctx)
	if err != nil {
		return fmt.Errorf("find theaters: %w", err)
	}

	// Filter theaters matching selected halls
	var targetTheaters []Theater
	for _, t := range theaters {
		for _, name := range halls {
			if t.Name == name {
				targetTheaters = append(targetTheaters, t)
				break
			}
		}
	}

	if len(targetTheaters) == 0 {
		return fmt.Errorf("no matching theaters found for: %v", halls)
	}

	// Slot definitions
	slotDefs := map[string]struct{ hour, min int }{
		"Morning":   {10, 0},
		"Afternoon": {13, 30},
		"Evening":   {17, 0},
		"Night":     {20, 30},
	}

	// Generate showtimes for the next 5 days in Thai timezone (ICT, UTC+7)
	ictZone := time.FixedZone("ICT", 7*3600)
	nowICT := time.Now().In(ictZone)
	today := time.Date(nowICT.Year(), nowICT.Month(), nowICT.Day(), 0, 0, 0, 0, ictZone)

	var showtimes []any
	for dayOffset := 0; dayOffset < 5; dayOffset++ {
		date := today.AddDate(0, 0, dayOffset)

		for _, t := range targetTheaters {
			priceByZone := make(map[string]float64)
			for _, z := range t.Zones {
				priceByZone[z.Name] = z.Price
			}
			seats := GenerateSeats(t.Zones)

			for _, slotLabel := range slots {
				def, exists := slotDefs[slotLabel]
				if !exists {
					continue
				}

				startTime := time.Date(date.Year(), date.Month(), date.Day(), def.hour, def.min, 0, 0, ictZone)
				endTime := startTime.Add(time.Duration(duration) * time.Minute)

				st := Showtime{
					ID:          bson.NewObjectID(),
					MovieID:     movieID,
					TheaterID:   t.ID,
					TheaterName: t.Name,
					StartTime:   startTime,
					EndTime:     endTime,
					SlotLabel:   slotLabel,
					PriceByZone: priceByZone,
					Seats:       seats,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				showtimes = append(showtimes, st)
			}
		}
	}

	if len(showtimes) > 0 {
		if err := s.repo.InsertMany(ctx, showtimes); err != nil {
			return fmt.Errorf("insert showtimes: %w", err)
		}
	}

	return nil
}
