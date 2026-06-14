package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	sharedauth "backend/internal/shared/auth"

	"github.com/gin-gonic/gin"
)

func TestAuthMiddlewareAcceptsSessionCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AuthMiddleware(func(token string) (*sharedauth.Claims, error) {
		if token != "valid-token" {
			t.Fatalf("unexpected token: %q", token)
		}
		return &sharedauth.Claims{UserID: "user-1"}, nil
	}, "cinema_session"))
	router.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "cinema_session", Value: "valid-token"})
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected %d, got %d", http.StatusNoContent, recorder.Code)
	}
}

func TestAuthMiddlewarePrefersBearerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AuthMiddleware(func(token string) (*sharedauth.Claims, error) {
		if token != "bearer-token" {
			t.Fatalf("unexpected token: %q", token)
		}
		return &sharedauth.Claims{UserID: "user-1"}, nil
	}, "cinema_session"))
	router.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer bearer-token")
	req.AddCookie(&http.Cookie{Name: "cinema_session", Value: "cookie-token"})
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected %d, got %d", http.StatusNoContent, recorder.Code)
	}
}
