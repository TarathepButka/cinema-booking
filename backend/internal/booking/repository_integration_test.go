package booking

import (
	"context"
	"os"
	"testing"
	"time"

	"backend/internal/shared/db"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestMongoRepositoryIntegration(t *testing.T) {
	uri := os.Getenv("TEST_MONGO_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017/?replicaSet=rs0"
	}

	mongoDB, err := db.NewMongoDB(uri, "cinema_test")
	if err != nil {
		t.Skipf("Skipping MongoDB repository integration test: %v", err)
	}
	defer mongoDB.Disconnect(context.Background())

	_ = mongoDB.EnsureIndexes(context.Background())

	repo := NewRepo(mongoDB)
	
	// Create a dummy booking
	userID := bson.NewObjectID()
	showtimeID := bson.NewObjectID()
	b := &Booking{
		UserID:     userID,
		ShowtimeID: showtimeID,
		Seats: []BookedSeat{
			{
				SeatLabel: "C3",
				Row:       "C",
				Number:    3,
				Zone:      "REGULAR",
				Price:     180,
			},
			{
				SeatLabel: "C4",
				Row:       "C",
				Number:    4,
				Zone:      "REGULAR",
				Price:     180,
			},
		},
		TotalPrice:  360,
		Status:      StatusPending,
		UserEmail:   "integration-test@example.com",
		MovieTitle:  "Integration Test Movie",
		TheaterName: "Premium Hall",
		StartTime:   time.Now().Add(24 * time.Hour),
	}

	t.Run("Create and Find Booking", func(t *testing.T) {
		created, err := repo.Create(context.Background(), b)
		if err != nil {
			t.Fatalf("Failed to create booking: %v", err)
		}
		if created.ID.IsZero() {
			t.Fatal("Expected non-zero ObjectID for booking ID")
		}

		// Clean up this booking at the end of the test
		t.Cleanup(func() {
			_, _ = mongoDB.DB.Collection("bookings").DeleteOne(context.Background(), bson.M{"_id": created.ID})
		})

		// Find by UserID
		bookings, err := repo.FindByUserID(context.Background(), userID)
		if err != nil {
			t.Fatalf("Failed to find bookings by userID: %v", err)
		}
		if len(bookings) != 1 {
			t.Fatalf("Expected 1 booking, got %d", len(bookings))
		}
		if bookings[0].MovieTitle != "Integration Test Movie" {
			t.Errorf("Expected movie title 'Integration Test Movie', got %s", bookings[0].MovieTitle)
		}
	})

	t.Run("Update Status and Aggregate", func(t *testing.T) {
		created, err := repo.Create(context.Background(), b)
		if err != nil {
			t.Fatalf("Failed to create booking: %v", err)
		}
		t.Cleanup(func() {
			_, _ = mongoDB.DB.Collection("bookings").DeleteOne(context.Background(), bson.M{"_id": created.ID})
		})

		// Update to CONFIRMED
		err = repo.UpdateStatus(context.Background(), created.ID, StatusConfirmed)
		if err != nil {
			t.Fatalf("Failed to update status: %v", err)
		}

		// Aggregate revenue by movie
		agg, err := repo.AggregateByMovie(context.Background(), bson.M{"movieTitle": "Integration Test Movie"})
		if err != nil {
			t.Fatalf("Failed to aggregate by movie: %v", err)
		}
		if len(agg) == 0 {
			t.Fatal("Expected at least one aggregation result")
		}

		result := agg[0]
		if result["_id"] != "Integration Test Movie" {
			t.Errorf("Expected aggregated movie title to be 'Integration Test Movie', got %v", result["_id"])
		}
	})
}
