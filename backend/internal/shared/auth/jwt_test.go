package auth

import (
	"testing"
	"time"
)

func TestJWTGenerateAndValidate(t *testing.T) {
	token, err := GenerateJWT("user-1", "user@example.com", RoleAdmin, "test-secret", 1)
	if err != nil {
		t.Fatalf("GenerateJWT() error = %v", err)
	}

	claims, err := ValidateJWT(token, "test-secret")
	if err != nil {
		t.Fatalf("ValidateJWT() error = %v", err)
	}
	if claims.UserID != "user-1" || claims.Email != "user@example.com" || claims.Role != RoleAdmin {
		t.Fatalf("unexpected claims: %#v", claims)
	}
	if claims.ExpiresAt == nil || time.Until(claims.ExpiresAt.Time) <= 0 {
		t.Fatalf("token expiry was not set in the future")
	}
}

func TestJWTRejectsWrongSecret(t *testing.T) {
	token, err := GenerateJWT("user-1", "user@example.com", RoleUser, "correct-secret", 1)
	if err != nil {
		t.Fatalf("GenerateJWT() error = %v", err)
	}

	if _, err := ValidateJWT(token, "wrong-secret"); err == nil {
		t.Fatal("ValidateJWT() expected an error for the wrong secret")
	}
}
