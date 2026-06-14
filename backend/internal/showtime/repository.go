package showtime

import (
	"context"
	"errors"
	"fmt"
	"time"

	"backend/internal/shared/db"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Repo handles showtime and seat-level persistence.
type Repo struct {
	col *mongo.Collection
}

// NewRepo creates a new showtime Repo.
func NewRepo(database *db.MongoDB) *Repo {
	return &Repo{col: database.DB.Collection("showtimes")}
}

// InsertMany inserts multiple showtime documents.
func (r *Repo) InsertMany(ctx context.Context, showtimes []any) error {
	_, err := r.col.InsertMany(ctx, showtimes)
	return err
}

// FindByMovieID returns all showtimes from today onwards for a given movie.
func (r *Repo) FindByMovieID(ctx context.Context, movieID bson.ObjectID) ([]Showtime, error) {
	// Calculate the beginning of today in Thai timezone (ICT, UTC+7)
	ictZone := time.FixedZone("ICT", 7*3600)
	nowICT := time.Now().In(ictZone)
	beginningOfToday := time.Date(nowICT.Year(), nowICT.Month(), nowICT.Day(), 0, 0, 0, 0, ictZone)

	filter := bson.M{
		"movieId":   movieID,
		"startTime": bson.M{"$gte": beginningOfToday},
	}
	opts := options.Find().
		SetSort(bson.D{{Key: "startTime", Value: 1}}).
		SetProjection(bson.M{"seats": 0}) // exclude seats in list view

	cursor, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("find showtimes: %w", err)
	}
	defer cursor.Close(ctx)

	var showtimes []Showtime
	if err := cursor.All(ctx, &showtimes); err != nil {
		return nil, fmt.Errorf("decode showtimes: %w", err)
	}
	return showtimes, nil
}

// FindByID returns a full showtime document (including seats).
func (r *Repo) FindByID(ctx context.Context, id bson.ObjectID) (*Showtime, error) {
	var showtime Showtime
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&showtime)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	return &showtime, err
}

// LockSeat atomically sets a seat to LOCKED status using findOneAndUpdate.
// Returns false if the seat is not AVAILABLE (already locked or booked).
func (r *Repo) LockSeat(ctx context.Context, showtimeID bson.ObjectID, seatLabel, userID string, lockUntil time.Time) (bool, error) {
	filter := bson.M{
		"_id": showtimeID,
		"seats": bson.M{
			"$elemMatch": bson.M{
				"seatLabel": seatLabel,
				"status":    string(SeatAvailable),
			},
		},
	}
	update := bson.M{
		"$set": bson.M{
			"seats.$.status":      string(SeatLocked),
			"seats.$.lockedBy":    userID,
			"seats.$.lockedUntil": lockUntil,
			"updatedAt":           time.Now(),
		},
	}

	result := r.col.FindOneAndUpdate(ctx, filter, update)
	if result.Err() != nil {
		if errors.Is(result.Err(), mongo.ErrNoDocuments) {
			return false, nil // seat not available
		}
		return false, fmt.Errorf("lock seat: %w", result.Err())
	}
	return true, nil
}

// ConfirmSeat changes a seat from LOCKED to BOOKED, verifying the owner.
func (r *Repo) ConfirmSeat(ctx context.Context, showtimeID bson.ObjectID, seatLabel, userID string, confirmedAt time.Time) (bool, error) {
	filter := bson.M{
		"_id": showtimeID,
		"seats": bson.M{
			"$elemMatch": bson.M{
				"seatLabel":   seatLabel,
				"status":      string(SeatLocked),
				"lockedBy":    userID,
				"lockedUntil": bson.M{"$gt": confirmedAt},
			},
		},
	}
	update := bson.M{
		"$set": bson.M{
			"seats.$.status":      string(SeatBooked),
			"seats.$.lockedBy":    nil,
			"seats.$.lockedUntil": nil,
			"updatedAt":           time.Now(),
		},
	}

	result := r.col.FindOneAndUpdate(ctx, filter, update)
	if result.Err() != nil {
		if errors.Is(result.Err(), mongo.ErrNoDocuments) {
			return false, nil
		}
		return false, fmt.Errorf("confirm seat: %w", result.Err())
	}
	return true, nil
}

// ReleaseSeat changes a seat from LOCKED back to AVAILABLE.
func (r *Repo) ReleaseSeat(ctx context.Context, showtimeID bson.ObjectID, seatLabel, userID string) (bool, error) {
	filter := bson.M{
		"_id": showtimeID,
		"seats": bson.M{
			"$elemMatch": bson.M{
				"seatLabel": seatLabel,
				"status":    string(SeatLocked),
				"lockedBy":  userID,
			},
		},
	}
	update := bson.M{
		"$set": bson.M{
			"seats.$.status":      string(SeatAvailable),
			"seats.$.lockedBy":    nil,
			"seats.$.lockedUntil": nil,
			"updatedAt":           time.Now(),
		},
	}

	result := r.col.FindOneAndUpdate(ctx, filter, update)
	if result.Err() != nil {
		if errors.Is(result.Err(), mongo.ErrNoDocuments) {
			return false, nil
		}
		return false, fmt.Errorf("release seat: %w", result.Err())
	}
	return true, nil
}

// FindExpiredLockedSeats returns showtimes that have seats with expired locks.
func (r *Repo) FindExpiredLockedSeats(ctx context.Context) ([]Showtime, error) {
	now := time.Now()
	filter := bson.M{
		"seats": bson.M{
			"$elemMatch": bson.M{
				"status":      string(SeatLocked),
				"lockedUntil": bson.M{"$lt": now},
			},
		},
	}

	cursor, err := r.col.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("find expired locks: %w", err)
	}
	defer cursor.Close(ctx)

	var showtimes []Showtime
	if err := cursor.All(ctx, &showtimes); err != nil {
		return nil, fmt.Errorf("decode expired showtimes: %w", err)
	}
	return showtimes, nil
}

// ReleaseExpiredSeat forces a LOCKED seat back to AVAILABLE (no user check).
func (r *Repo) ReleaseExpiredSeat(ctx context.Context, showtimeID bson.ObjectID, seatLabel string, expiredAt time.Time) (bool, error) {
	filter := bson.M{
		"_id": showtimeID,
		"seats": bson.M{
			"$elemMatch": bson.M{
				"seatLabel":   seatLabel,
				"status":      string(SeatLocked),
				"lockedUntil": bson.M{"$lte": expiredAt},
			},
		},
	}
	update := bson.M{
		"$set": bson.M{
			"seats.$.status":      string(SeatAvailable),
			"seats.$.lockedBy":    nil,
			"seats.$.lockedUntil": nil,
			"updatedAt":           time.Now(),
		},
	}
	result, err := r.col.UpdateOne(ctx, filter, update)
	if err != nil {
		return false, err
	}
	return result.ModifiedCount == 1, nil
}

// ─── Theater Repo ─────────────────────────────────────────────────────────────

// TheaterRepo handles theater persistence operations.
type TheaterRepo struct {
	col *mongo.Collection
}

// NewTheaterRepo creates a new TheaterRepo.
func NewTheaterRepo(database *db.MongoDB) *TheaterRepo {
	return &TheaterRepo{col: database.DB.Collection("theaters")}
}

// FindAll returns all theaters.
func (r *TheaterRepo) FindAll(ctx context.Context) ([]Theater, error) {
	opts := options.Find().SetSort(bson.D{{Key: "name", Value: 1}})
	cursor, err := r.col.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("find theaters: %w", err)
	}
	defer cursor.Close(ctx)

	var theaters []Theater
	if err := cursor.All(ctx, &theaters); err != nil {
		return nil, fmt.Errorf("decode theaters: %w", err)
	}
	return theaters, nil
}

// FindByID returns a theater by ObjectID.
func (r *TheaterRepo) FindByID(ctx context.Context, id bson.ObjectID) (*Theater, error) {
	var theater Theater
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&theater)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	return &theater, err
}
