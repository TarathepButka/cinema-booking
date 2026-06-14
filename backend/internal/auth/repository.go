package auth

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

// UserRepo handles all user persistence operations.
type UserRepo struct {
	col *mongo.Collection
}

// NewUserRepo creates a new UserRepo.
func NewUserRepo(database *db.MongoDB) *UserRepo {
	return &UserRepo{col: database.DB.Collection("users")}
}

// FindByID returns a user by their ObjectID.
func (r *UserRepo) FindByID(ctx context.Context, id bson.ObjectID) (*User, error) {
	var user User
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	return &user, err
}

// FindByEmail returns a user by email address.
func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := r.col.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	return &user, err
}

// FindEmailSuggestions returns paginated user emails matching the query.
func (r *UserRepo) FindEmailSuggestions(ctx context.Context, query string, page, limit int64) ([]string, int64, error) {
	filter := bson.M{}
	if query != "" {
		filter["email"] = bson.M{"$regex": regexp.QuoteMeta(query), "$options": "i"}
	}

	total, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count user suggestions: %w", err)
	}

	opts := options.Find().
		SetProjection(bson.M{"email": 1}).
		SetSort(bson.D{{Key: "email", Value: 1}}).
		SetSkip((page - 1) * limit).
		SetLimit(limit)
	cursor, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("find user suggestions: %w", err)
	}
	defer cursor.Close(ctx)

	var results []struct {
		Email string `bson:"email"`
	}
	if err := cursor.All(ctx, &results); err != nil {
		return nil, 0, fmt.Errorf("decode user suggestions: %w", err)
	}

	suggestions := make([]string, 0, len(results))
	for _, result := range results {
		if result.Email != "" {
			suggestions = append(suggestions, result.Email)
		}
	}
	return suggestions, total, nil
}

// FindByGoogleID returns a user by Google OAuth ID.
func (r *UserRepo) FindByGoogleID(ctx context.Context, googleID string) (*User, error) {
	var user User
	err := r.col.FindOne(ctx, bson.M{"googleId": googleID}).Decode(&user)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	return &user, err
}

// Create inserts a new user and returns it with the generated ID.
func (r *UserRepo) Create(ctx context.Context, user *User) (*User, error) {
	user.ID = bson.NewObjectID()
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	if _, err := r.col.InsertOne(ctx, user); err != nil {
		return nil, fmt.Errorf("insert user: %w", err)
	}
	return user, nil
}

// UpsertByGoogleID creates or updates a user based on their Google ID.
func (r *UserRepo) UpsertByGoogleID(ctx context.Context, user *User) (*User, error) {
	existing, err := r.FindByGoogleID(ctx, *user.GoogleID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		// Update name and picture if changed
		_, err = r.col.UpdateOne(ctx,
			bson.M{"_id": existing.ID},
			bson.M{"$set": bson.M{"name": user.Name, "picture": user.Picture, "updatedAt": time.Now()}},
		)
		if err != nil {
			return nil, fmt.Errorf("update user: %w", err)
		}
		existing.Name = user.Name
		existing.Picture = user.Picture
		return existing, nil
	}
	return r.Create(ctx, user)
}
