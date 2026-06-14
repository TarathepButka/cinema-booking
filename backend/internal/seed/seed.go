package seed

import (
	"context"
	"fmt"
	"log"
	"time"

	"backend/config"
	"backend/internal/auth"
	"backend/internal/movie"
	"backend/internal/shared/db"
	"backend/internal/showtime"

	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"
)

type showtimeSlot struct {
	label     string
	startHour int
	startMin  int
}

var slots = []showtimeSlot{
	{"Morning", 10, 0},
	{"Afternoon", 13, 30},
	{"Evening", 17, 0},
	{"Night", 20, 30},
}

func generateHallASeats() []showtime.Seat {
	zones := []struct {
		name  string
		rows  []string
		cols  int
		price float64
	}{
		{"FRONT", []string{"A", "B"}, 5, 150},
		{"MIDDLE", []string{"C", "D"}, 5, 200},
		{"BACK", []string{"E", "F"}, 5, 280},
	}
	var seats []showtime.Seat
	for _, z := range zones {
		for _, row := range z.rows {
			for col := 1; col <= z.cols; col++ {
				seats = append(seats, showtime.Seat{
					Row:       row,
					Number:    col,
					SeatLabel: fmt.Sprintf("%s%d", row, col),
					Zone:      z.name,
					Price:     z.price,
					Status:    showtime.SeatAvailable,
				})
			}
		}
	}
	return seats
}

func generateHallBSeats() []showtime.Seat {
	zones := []struct {
		name  string
		rows  []string
		cols  int
		price float64
	}{
		{"FRONT", []string{"A", "B"}, 8, 120},
		{"MIDDLE", []string{"C", "D"}, 6, 180},
		{"BACK", []string{"E", "F"}, 6, 250},
	}
	var seats []showtime.Seat
	for _, z := range zones {
		for _, row := range z.rows {
			for col := 1; col <= z.cols; col++ {
				seats = append(seats, showtime.Seat{
					Row:       row,
					Number:    col,
					SeatLabel: fmt.Sprintf("%s%d", row, col),
					Zone:      z.name,
					Price:     z.price,
					Status:    showtime.SeatAvailable,
				})
			}
		}
	}
	return seats
}

func buildShowtime(movieID bson.ObjectID, theaterID bson.ObjectID, theaterName string, movieDuration int, date time.Time, slot showtimeSlot, priceByZone map[string]float64) showtime.Showtime {
	startTime := time.Date(date.Year(), date.Month(), date.Day(), slot.startHour, slot.startMin, 0, 0, date.Location())
	endTime := startTime.Add(time.Duration(movieDuration) * time.Minute)

	var seats []showtime.Seat
	if theaterName == "Hall A" {
		seats = generateHallASeats()
	} else {
		seats = generateHallBSeats()
	}

	return showtime.Showtime{
		ID:          bson.NewObjectID(),
		MovieID:     movieID,
		TheaterID:   theaterID,
		TheaterName: theaterName,
		StartTime:   startTime,
		EndTime:     endTime,
		SlotLabel:   slot.label,
		PriceByZone: priceByZone,
		Seats:       seats,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// SeedDatabase checks if the database is empty. If so, it seeds default theaters, movies, showtimes, and an admin user.
func SeedDatabase(ctx context.Context, mongoDB *db.MongoDB, cfg *config.Config) error {
	log.Println("🎬 Checking database seeding status...")

	count, err := mongoDB.DB.Collection("movies").CountDocuments(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("check existing movies count: %w", err)
	}
	if count > 0 {
		log.Println("⏭️  Database already seeded. Skipping.")
		return nil
	}

	log.Println("🎬 Database is empty. Starting database seeding...")

	// 1. Seed Theaters
	log.Println("🏬 Seeding theaters...")
	hallAId := bson.NewObjectID()
	hallBId := bson.NewObjectID()

	theaters := []any{
		showtime.Theater{
			ID:          hallAId,
			Name:        "Hall A",
			Description: "Premium theater with superior sound & screen",
			TotalSeats:  30,
			Zones: []showtime.Zone{
				{Name: "FRONT", Rows: []string{"A", "B"}, SeatsPerRow: 5, Price: 150},
				{Name: "MIDDLE", Rows: []string{"C", "D"}, SeatsPerRow: 5, Price: 200},
				{Name: "BACK", Rows: []string{"E", "F"}, SeatsPerRow: 5, Price: 280},
			},
			CreatedAt: time.Now(),
		},
		showtime.Theater{
			ID:          hallBId,
			Name:        "Hall B",
			Description: "Standard theater with great value for money",
			TotalSeats:  40,
			Zones: []showtime.Zone{
				{Name: "FRONT", Rows: []string{"A", "B"}, SeatsPerRow: 8, Price: 120},
				{Name: "MIDDLE", Rows: []string{"C", "D"}, SeatsPerRow: 6, Price: 180},
				{Name: "BACK", Rows: []string{"E", "F"}, SeatsPerRow: 6, Price: 250},
			},
			CreatedAt: time.Now(),
		},
	}

	_, err = mongoDB.DB.Collection("theaters").InsertMany(ctx, theaters)
	if err != nil {
		return fmt.Errorf("seed theaters: %w", err)
	}
	log.Printf("   ✅ Inserted %d theaters\n", len(theaters))

	// 2. Seed Movies
	log.Println("🎬 Seeding movies...")

	duneRelease, _ := time.Parse("2006-01-02", "2025-11-15")
	batmanRelease, _ := time.Parse("2006-01-02", "2025-10-03")
	spiritedRelease, _ := time.Parse("2006-01-02", "2025-09-20")
	oppenheimerRelease, _ := time.Parse("2006-01-02", "2025-08-06")
	fastRelease, _ := time.Parse("2006-01-02", "2025-07-18")

	movies := []movie.Movie{
		{
			ID:          bson.NewObjectID(),
			Title:       "Dune: Part Three",
			Description: "The epic conclusion of Paul Atreides' journey as he leads the Fremen in an interstellar holy war. The fate of Arrakis — and the universe — hangs in the balance.",
			Genre:       []string{"Sci-Fi", "Adventure", "Drama"},
			Duration:    165,
			Rating:      8.5,
			Language:    "English",
			Director:    "Denis Villeneuve",
			Cast:        []string{"Timothée Chalamet", "Zendaya", "Rebecca Ferguson"},
			PosterURL:   "https://images.unsplash.com/photo-1446776811953-b23d57bd21aa?w=400&h=600&fit=crop",
			ReleaseDate: duneRelease,
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          bson.NewObjectID(),
			Title:       "The Batman: Arkham",
			Description: "Bruce Wayne faces his darkest night yet when the Arkham Asylum falls into chaos. A new breed of villain threatens to tear Gotham City apart from the inside.",
			Genre:       []string{"Action", "Thriller", "Crime"},
			Duration:    152,
			Rating:      8.2,
			Language:    "English",
			Director:    "Matt Reeves",
			Cast:        []string{"Robert Pattinson", "Zoë Kravitz", "Paul Dano"},
			PosterURL:   "https://images.unsplash.com/photo-1531259683007-016a7b628fc3?w=400&h=600&fit=crop",
			ReleaseDate: batmanRelease,
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          bson.NewObjectID(),
			Title:       "Spirited Away 2",
			Description: "Chihiro returns to the spirit world, now as a young adult, to rescue a new friend trapped by the mysterious ruler of the bathhouse. A breathtaking sequel from Studio Ghibli.",
			Genre:       []string{"Animation", "Fantasy", "Adventure"},
			Duration:    130,
			Rating:      9.0,
			Language:    "Japanese",
			Director:    "Hayao Miyazaki",
			Cast:        []string{"Daveigh Chase", "Suzanne Pleshette"},
			PosterURL:   "https://images.unsplash.com/photo-1607604276583-eef5d076aa5f?w=400&h=600&fit=crop",
			ReleaseDate: spiritedRelease,
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          bson.NewObjectID(),
			Title:       "Oppenheimer: Legacy",
			Description: "A dramatic retelling of J. Robert Oppenheimer's final years, exploring the ethical consequences of the Manhattan Project through the eyes of those who lived through the atomic age.",
			Genre:       []string{"Drama", "History", "Biography"},
			Duration:    148,
			Rating:      8.7,
			Language:    "English",
			Director:    "Christopher Nolan",
			Cast:        []string{"Cillian Murphy", "Emily Blunt", "Matt Damon"},
			PosterURL:   "https://images.unsplash.com/photo-1518818608552-195ed130cdf4?w=400&h=600&fit=crop",
			ReleaseDate: oppenheimerRelease,
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          bson.NewObjectID(),
			Title:       "Fast & Furious 12",
			Description: "Dom Toretto and his family face their most impossible mission yet — infiltrate a global criminal syndicate that has taken one of their own. Buckle up for the final ride.",
			Genre:       []string{"Action", "Thriller", "Racing"},
			Duration:    135,
			Rating:      7.1,
			Language:    "English",
			Director:    "Louis Leterrier",
			Cast:        []string{"Vin Diesel", "Michelle Rodriguez", "Tyrese Gibson"},
			PosterURL:   "https://images.unsplash.com/photo-1568605117036-5fe5e7bab0b7?w=400&h=600&fit=crop",
			ReleaseDate: fastRelease,
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	movieAnys := make([]any, len(movies))
	for i, m := range movies {
		movieAnys[i] = m
	}

	_, err = mongoDB.DB.Collection("movies").InsertMany(ctx, movieAnys)
	if err != nil {
		return fmt.Errorf("seed movies: %w", err)
	}
	log.Printf("   ✅ Inserted %d movies\n", len(movies))

	// 3. Seed Showtimes — 5 days × movies × slots
	log.Println("🕐 Seeding showtimes (5 days)...")

	theaterAssignments := []struct {
		theaterID   bson.ObjectID
		theaterName string
		priceByZone map[string]float64
	}{
		{hallAId, "Hall A", map[string]float64{"FRONT": 150, "MIDDLE": 200, "BACK": 280}},
		{hallBId, "Hall B", map[string]float64{"FRONT": 120, "MIDDLE": 180, "BACK": 250}},
		{hallAId, "Hall A", map[string]float64{"FRONT": 150, "MIDDLE": 200, "BACK": 280}},
		{hallBId, "Hall B", map[string]float64{"FRONT": 120, "MIDDLE": 180, "BACK": 250}},
		{hallAId, "Hall A", map[string]float64{"FRONT": 150, "MIDDLE": 200, "BACK": 280}},
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var showtimeAnys []any
	for dayOffset := 0; dayOffset < 5; dayOffset++ {
		date := today.AddDate(0, 0, dayOffset)

		for idx, m := range movies {
			theater := theaterAssignments[idx]
			for _, slot := range slots {
				st := buildShowtime(m.ID, theater.theaterID, theater.theaterName, m.Duration, date, slot, theater.priceByZone)
				showtimeAnys = append(showtimeAnys, st)
			}
		}
	}

	_, err = mongoDB.DB.Collection("showtimes").InsertMany(ctx, showtimeAnys)
	if err != nil {
		return fmt.Errorf("seed showtimes: %w", err)
	}
	log.Printf("   ✅ Inserted %d showtimes (5 movies × 4 slots × 5 days)\n", len(showtimeAnys))

	// 4. Seed Admin User
	log.Println("👤 Seeding admin user...")
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), 12)
	if err != nil {
		return fmt.Errorf("hash admin password: %w", err)
	}
	passwordHashStr := string(hashedPassword)

	adminUser := auth.User{
		ID:           bson.NewObjectID(),
		Email:        cfg.AdminEmail,
		Name:         "Admin",
		Role:         auth.RoleAdmin,
		PasswordHash: &passwordHashStr,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	_, err = mongoDB.DB.Collection("users").InsertOne(ctx, adminUser)
	if err != nil {
		return fmt.Errorf("seed admin user: %w", err)
	}
	log.Printf("   ✅ Admin user created: %s\n", cfg.AdminEmail)

	log.Println("🎉 Database seeding completed successfully!")
	return nil
}
