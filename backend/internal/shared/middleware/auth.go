package middleware

import (
	"net/http"
	"strings"

	sharedauth "backend/internal/shared/auth"

	"github.com/gin-gonic/gin"
)

const UserContextKey = "currentUser"

// TokenValidator is a function that validates a JWT string and returns Claims.
// Passing a function (instead of *auth.AuthService) breaks the import cycle:
// middleware does NOT import internal/auth — only internal/shared/auth.
type TokenValidator func(tokenString string) (*sharedauth.Claims, error)

// AuthMiddleware validates the JWT Bearer token and sets user claims in context.
func AuthMiddleware(validate TokenValidator, cookieName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		tokenString := ""
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format, expected: Bearer <token>"})
				return
			}
			tokenString = parts[1]
		} else if cookie, err := c.Cookie(cookieName); err == nil {
			tokenString = cookie
		}

		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}

		claims, err := validate(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set(UserContextKey, claims)
		c.Next()
	}
}

// GetCurrentUser retrieves the authenticated user's claims from the Gin context.
func GetCurrentUser(c *gin.Context) *sharedauth.Claims {
	val, exists := c.Get(UserContextKey)
	if !exists {
		return nil
	}
	claims, _ := val.(*sharedauth.Claims)
	return claims
}
