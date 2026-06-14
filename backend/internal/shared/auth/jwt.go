// Package auth provides shared JWT utilities used across the application.
// Keeping Claims and JWT logic here (instead of in the auth feature package)
// prevents import cycles between internal/auth and internal/shared/middleware.
package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Role is an authorization role stored in user records and JWT claims.
type Role string

const (
	RoleUser  Role = "USER"
	RoleAdmin Role = "ADMIN"
)

// Claims is the JWT payload structure embedded in every signed token.
type Claims struct {
	UserID string `json:"userId"`
	Email  string `json:"email"`
	Role   Role   `json:"role"`
	jwt.RegisteredClaims
}

// GenerateJWT creates a signed HS256 JWT token for the given user identity.
func GenerateJWT(userID, email string, role Role, secret string, expireHours int) (string, error) {
	expiry := time.Now().Add(time.Duration(expireHours) * time.Hour)
	claims := &Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiry),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateJWT parses and validates a JWT token string.
// Returns the embedded Claims on success.
func ValidateJWT(tokenString, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}
