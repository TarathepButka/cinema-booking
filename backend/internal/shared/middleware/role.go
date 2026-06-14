package middleware

import (
	"net/http"

	sharedauth "backend/internal/shared/auth"

	"github.com/gin-gonic/gin"
)

// RequireRole returns a middleware that restricts access to users with one of the specified roles.
// Designed for extensibility — pass multiple roles to allow any of them.
// Usage:
//
//	router.Use(RequireRole("ADMIN"))
//	router.Use(RequireRole("ADMIN", "MODERATOR"))
func RequireRole(roles ...sharedauth.Role) gin.HandlerFunc {
	roleSet := make(map[sharedauth.Role]struct{}, len(roles))
	for _, r := range roles {
		roleSet[r] = struct{}{}
	}

	return func(c *gin.Context) {
		claims := GetCurrentUser(c)
		if claims == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}

		if _, allowed := roleSet[claims.Role]; !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":         "access denied",
				"requiredRoles": roles,
				"yourRole":      claims.Role,
			})
			return
		}

		c.Next()
	}
}
