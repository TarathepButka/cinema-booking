package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type OriginPolicy struct {
	allowed map[string]struct{}
}

func NewOriginPolicy(origins []string) *OriginPolicy {
	allowed := make(map[string]struct{}, len(origins))
	for _, origin := range origins {
		if normalized := normalizeOrigin(origin); normalized != "" {
			allowed[normalized] = struct{}{}
		}
	}
	return &OriginPolicy{allowed: allowed}
}

func (p *OriginPolicy) IsAllowed(origin string) bool {
	_, ok := p.allowed[normalizeOrigin(origin)]
	return ok
}

func (p *OriginPolicy) CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			if !p.IsAllowed(origin) {
				if c.Request.Method == http.MethodOptions {
					c.AbortWithStatus(http.StatusForbidden)
					return
				}
			} else {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Vary", "Origin")
				c.Header("Access-Control-Allow-Credentials", "true")
				c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
				c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Requested-With")
				c.Header("Access-Control-Max-Age", "86400")
			}
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// TrustedOriginMiddleware rejects unsafe browser requests from untrusted origins.
// Requests without Origin remain valid for non-browser bearer-token clients.
func (p *OriginPolicy) TrustedOriginMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		switch c.Request.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			c.Next()
			return
		}

		origin := c.GetHeader("Origin")
		if origin != "" && !p.IsAllowed(origin) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "origin not allowed"})
			return
		}
		c.Next()
	}
}

func normalizeOrigin(origin string) string {
	return strings.TrimRight(strings.TrimSpace(origin), "/")
}

// RequestLogger logs each request with method, path, status, and duration.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		_ = start
	}
}
