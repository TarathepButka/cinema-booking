package booking

import (
	"context"
	"fmt"
	"time"

	"backend/internal/shared/db"
	"backend/internal/shared/mq"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// ─── Booking Repository ───────────────────────────────────────────────────────

// Repo handles booking persistence.
type Repo struct {
	client *mongo.Client
	col    *mongo.Collection
}

// NewRepo creates a new booking Repo.
func NewRepo(database *db.MongoDB) *Repo {
	return &Repo{
		client: database.Client,
		col:    database.DB.Collection("bookings"),
	}
}

// WithTransaction runs a booking operation in a MongoDB transaction.
func (r *Repo) WithTransaction(ctx context.Context, fn func(context.Context) (*Booking, error)) (*Booking, error) {
	session, err := r.client.StartSession()
	if err != nil {
		return nil, fmt.Errorf("start booking transaction: %w", err)
	}
	defer session.EndSession(ctx)

	result, err := session.WithTransaction(ctx, func(txCtx context.Context) (any, error) {
		return fn(txCtx)
	})
	if err != nil {
		return nil, err
	}

	booking, ok := result.(*Booking)
	if !ok {
		return nil, fmt.Errorf("booking transaction returned unexpected result")
	}
	return booking, nil
}

// Create inserts a new booking record.
func (r *Repo) Create(ctx context.Context, booking *Booking) (*Booking, error) {
	booking.ID = bson.NewObjectID()
	now := time.Now()
	booking.CreatedAt = now
	booking.UpdatedAt = now

	if _, err := r.col.InsertOne(ctx, booking); err != nil {
		return nil, fmt.Errorf("insert booking: %w", err)
	}
	return booking, nil
}

// FindByUserID returns all bookings for a user, newest first.
func (r *Repo) FindByUserID(ctx context.Context, userID bson.ObjectID) ([]Booking, error) {
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	cursor, err := r.col.Find(ctx, bson.M{"userId": userID}, opts)
	if err != nil {
		return nil, fmt.Errorf("find user bookings: %w", err)
	}
	defer cursor.Close(ctx)

	var bookings []Booking
	if err := cursor.All(ctx, &bookings); err != nil {
		return nil, fmt.Errorf("decode bookings: %w", err)
	}
	return bookings, nil
}

// FindAll returns bookings with optional filters (admin use).
func (r *Repo) FindAll(ctx context.Context, filter bson.M, page, limit int64) ([]Booking, int64, error) {
	total, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count bookings: %w", err)
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetSkip((page - 1) * limit).
		SetLimit(limit)

	cursor, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("find bookings: %w", err)
	}
	defer cursor.Close(ctx)

	var bookings []Booking
	if err := cursor.All(ctx, &bookings); err != nil {
		return nil, 0, fmt.Errorf("decode bookings: %w", err)
	}
	return bookings, total, nil
}

// UpdateStatus changes the booking status (e.g., PENDING → CONFIRMED).
func (r *Repo) UpdateStatus(ctx context.Context, id bson.ObjectID, status BookingStatus) error {
	_, err := r.col.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"status": status, "updatedAt": time.Now()}},
	)
	return err
}

// AggregateByMovie returns booking counts grouped by movie title (admin stats), optionally filtered.
func (r *Repo) AggregateByMovie(ctx context.Context, filter bson.M) ([]bson.M, error) {
	matchFilter := bson.M{"status": string(StatusConfirmed)}
	for k, v := range filter {
		if k != "status" {
			matchFilter[k] = v
		}
	}

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: matchFilter}},
		bson.D{{Key: "$group", Value: bson.M{
			"_id":          "$movieTitle",
			"count":        bson.M{"$sum": 1},
			"totalRevenue": bson.M{"$sum": "$totalPrice"},
		}}},
		bson.D{{Key: "$sort", Value: bson.M{"count": -1}}},
	}
	cursor, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

// GetMovieSuggestions returns unique movie titles matching query.
func (r *Repo) GetMovieSuggestions(ctx context.Context, q string, limit int64) ([]string, error) {
	match := bson.M{}
	if q != "" {
		match["movieTitle"] = bson.M{"$regex": q, "$options": "i"}
	}
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: match}},
		bson.D{{Key: "$group", Value: bson.M{"_id": "$movieTitle"}}},
		bson.D{{Key: "$limit", Value: limit}},
	}
	cursor, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		ID string `bson:"_id"`
	}
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	suggestions := make([]string, len(results))
	for i, res := range results {
		suggestions[i] = res.ID
	}
	return suggestions, nil
}

// GetUserSuggestions returns unique user emails matching query.
func (r *Repo) GetUserSuggestions(ctx context.Context, q string, limit int64) ([]string, error) {
	match := bson.M{}
	if q != "" {
		match["userEmail"] = bson.M{"$regex": q, "$options": "i"}
	}
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: match}},
		bson.D{{Key: "$group", Value: bson.M{"_id": "$userEmail"}}},
		bson.D{{Key: "$limit", Value: limit}},
	}
	cursor, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		ID string `bson:"_id"`
	}
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	suggestions := make([]string, len(results))
	for i, res := range results {
		suggestions[i] = res.ID
	}
	return suggestions, nil
}

// ─── Audit Log Repository ─────────────────────────────────────────────────────

// AuditLogRepo handles audit log persistence.
type AuditLogRepo struct {
	col *mongo.Collection
}

// NewAuditLogRepo creates a new AuditLogRepo.
func NewAuditLogRepo(database *db.MongoDB) *AuditLogRepo {
	return &AuditLogRepo{col: database.DB.Collection("audit_logs")}
}

// Insert creates a new audit log entry.
func (r *AuditLogRepo) Insert(ctx context.Context, entry *AuditLog) error {
	entry.ID = bson.NewObjectID()
	entry.CreatedAt = time.Now()

	if _, err := r.col.InsertOne(ctx, entry); err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}
	return nil
}

// InsertEntry implements mq.AuditLogInserter — converts mq.AuditLogEntry to AuditLog.
func (r *AuditLogRepo) InsertEntry(ctx context.Context, entry mq.AuditLogEntry) error {
	log := &AuditLog{
		Event:      entry.Event,
		UserID:     entry.UserID,
		UserEmail:  entry.UserEmail,
		ShowtimeID: entry.ShowtimeID,
		BookingID:  entry.BookingID,
		SeatLabels: entry.SeatLabels,
		Details:    entry.Details,
	}
	return r.Insert(ctx, log)
}

// FindAll returns audit logs with optional event filter and pagination.
func (r *AuditLogRepo) FindAll(ctx context.Context, event string, page, limit int64) ([]AuditLog, int64, error) {
	filter := bson.M{}
	if event != "" {
		filter["event"] = event
	}

	total, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count audit logs: %w", err)
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetSkip((page - 1) * limit).
		SetLimit(limit)

	cursor, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("find audit logs: %w", err)
	}
	defer cursor.Close(ctx)

	var logs []AuditLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, 0, fmt.Errorf("decode audit logs: %w", err)
	}
	return logs, total, nil
}
