package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// MongoDB wraps the MongoDB client and exposes the primary database.
type MongoDB struct {
	Client *mongo.Client
	DB     *mongo.Database
}

// NewMongoDB creates and verifies a MongoDB connection.
func NewMongoDB(uri, dbName string) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("mongo connect: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("mongo ping: %w", err)
	}

	if err := verifyTransactionSupport(ctx, client); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, err
	}

	log.Println("✅ MongoDB connected")
	return &MongoDB{Client: client, DB: client.Database(dbName)}, nil
}

func verifyTransactionSupport(ctx context.Context, client *mongo.Client) error {
	var hello struct {
		SetName string `bson:"setName"`
		Message string `bson:"msg"`
	}
	if err := client.Database("admin").RunCommand(ctx, bson.D{{Key: "hello", Value: 1}}).Decode(&hello); err != nil {
		return fmt.Errorf("mongo deployment check: %w", err)
	}
	if hello.SetName == "" && hello.Message != "isdbgrid" {
		return fmt.Errorf(
			"MongoDB transactions require a replica set or mongos; start local MongoDB with --replSet rs0, run rs.initiate(), and set MONGO_REPLICA_SET=rs0",
		)
	}
	return nil
}

// EnsureIndexes creates necessary indexes if they don't already exist.
func (m *MongoDB) EnsureIndexes(ctx context.Context) error {
	indexes := []struct {
		collection string
		model      mongo.IndexModel
	}{
		{
			collection: "users",
			model: mongo.IndexModel{
				Keys:    bson.D{{Key: "email", Value: 1}},
				Options: options.Index().SetUnique(true),
			},
		},
		{
			collection: "users",
			model: mongo.IndexModel{
				Keys:    bson.D{{Key: "googleId", Value: 1}},
				Options: options.Index().SetSparse(true),
			},
		},
		{
			collection: "movies",
			model: mongo.IndexModel{
				Keys: bson.D{{Key: "isActive", Value: 1}},
			},
		},
		{
			collection: "showtimes",
			model: mongo.IndexModel{
				Keys: bson.D{
					{Key: "movieId", Value: 1},
					{Key: "startTime", Value: 1},
				},
			},
		},
		{
			collection: "showtimes",
			model: mongo.IndexModel{
				Keys: bson.D{{Key: "seats.lockedUntil", Value: 1}},
			},
		},
		{
			collection: "bookings",
			model: mongo.IndexModel{
				Keys: bson.D{{Key: "userId", Value: 1}},
			},
		},
		{
			collection: "bookings",
			model: mongo.IndexModel{
				Keys: bson.D{{Key: "showtimeId", Value: 1}},
			},
		},
		{
			collection: "audit_logs",
			model: mongo.IndexModel{
				Keys: bson.D{{Key: "createdAt", Value: -1}},
			},
		},
		{
			collection: "audit_logs",
			model: mongo.IndexModel{
				Keys: bson.D{{Key: "event", Value: 1}},
			},
		},
	}

	for _, idx := range indexes {
		col := m.DB.Collection(idx.collection)
		if _, err := col.Indexes().CreateOne(ctx, idx.model); err != nil {
			log.Printf("⚠️  Index on %s: %v", idx.collection, err)
		}
	}

	log.Println("✅ MongoDB indexes ensured")
	return nil
}

// Disconnect closes the MongoDB connection gracefully.
func (m *MongoDB) Disconnect(ctx context.Context) {
	if err := m.Client.Disconnect(ctx); err != nil {
		log.Printf("MongoDB disconnect error: %v", err)
	}
}
