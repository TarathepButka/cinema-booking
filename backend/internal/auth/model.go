package auth

import (
	"time"

	sharedauth "backend/internal/shared/auth"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	RoleUser  = sharedauth.RoleUser
	RoleAdmin = sharedauth.RoleAdmin
)

// User represents a registered user in the system.
type User struct {
	ID           bson.ObjectID   `bson:"_id,omitempty"    json:"id"`
	Email        string          `bson:"email"            json:"email"`
	Name         string          `bson:"name"             json:"name"`
	Role         sharedauth.Role `bson:"role"             json:"role"`
	Picture      string          `bson:"picture"          json:"picture"`
	GoogleID     *string         `bson:"googleId"         json:"googleId,omitempty"`
	PasswordHash *string         `bson:"passwordHash"     json:"-"`
	CreatedAt    time.Time       `bson:"createdAt"        json:"createdAt"`
	UpdatedAt    time.Time       `bson:"updatedAt"        json:"updatedAt"`
}
