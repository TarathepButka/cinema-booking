package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCORSMiddlewareAllowsConfiguredOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	policy := NewOriginPolicy([]string{"https://cinema.example.com"})
	router := gin.New()
	router.Use(policy.CORSMiddleware())
	router.GET("/", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://cinema.example.com")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Header().Get("Access-Control-Allow-Origin") != "https://cinema.example.com" {
		t.Fatalf("trusted origin was not returned")
	}
}

func TestTrustedOriginMiddlewareRejectsUnsafeRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	policy := NewOriginPolicy([]string{"https://cinema.example.com"})
	router := gin.New()
	router.Use(policy.TrustedOriginMiddleware())
	router.POST("/", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Origin", "https://attacker.example")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected %d, got %d", http.StatusForbidden, recorder.Code)
	}
}
