package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"time"

	"backend/config"
	"backend/internal/shared/middleware"
	"backend/internal/shared/query"

	"github.com/gin-gonic/gin"
)

// Handler handles authentication HTTP endpoints.
type Handler struct {
	svc          *Service
	cookieSecure bool
	cookieMaxAge int
}

// NewHandler creates a new auth Handler.
func NewHandler(svc *Service, cfg *config.Config) *Handler {
	return &Handler{
		svc:          svc,
		cookieSecure: cfg.CookieSecure,
		cookieMaxAge: int((time.Duration(cfg.JWTExpireHours) * time.Hour).Seconds()),
	}
}

const (
	SessionCookieName  = "cinema_session"
	oauthStateCookie   = "cinema_oauth_state"
	oauthStateSize     = 32
	oauthStateMaxAge   = 10 * 60
	suggestionLimit    = 10
	suggestionMaxLimit = 50
)

// Login authenticates an admin user with email and password.
// POST /api/auth/login
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, token, err := h.svc.AdminLogin(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	h.setSessionCookie(c, token)
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"userId":  user.ID.Hex(),
			"id":      user.ID.Hex(),
			"email":   user.Email,
			"name":    user.Name,
			"role":    user.Role,
			"dbRole":  user.Role,
			"picture": user.Picture,
		},
	})
}

// GoogleLogin redirects the user to Google's OAuth consent page.
// GET /api/auth/google/login
func (h *Handler) GoogleLogin(c *gin.Context) {
	state, err := randomToken(oauthStateSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start OAuth login"})
		return
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     oauthStateCookie,
		Value:    state,
		Path:     "/api/auth/google",
		MaxAge:   oauthStateMaxAge,
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
	url := h.svc.GetGoogleAuthURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GoogleCallback handles the OAuth callback from Google.
// GET /api/auth/google/callback
func (h *Handler) GoogleCallback(c *gin.Context) {
	state := c.Query("state")
	stateCookie, err := c.Cookie(oauthStateCookie)
	h.clearCookie(c, oauthStateCookie, "/api/auth/google")
	if err != nil || state == "" ||
		subtle.ConstantTimeCompare([]byte(state), []byte(stateCookie)) != 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid OAuth state"})
		return
	}

	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing authorization code"})
		return
	}

	_, token, err := h.svc.HandleGoogleCallback(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "OAuth login failed: " + err.Error()})
		return
	}

	h.setSessionCookie(c, token)
	frontendURL := h.svc.GetFrontendURL()
	c.Redirect(http.StatusTemporaryRedirect, frontendURL)
}

// Me returns the authenticated user's profile.
// GET /api/auth/me
func (h *Handler) Me(c *gin.Context) {
	claims := middleware.GetCurrentUser(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	user, err := h.svc.GetUserByID(c.Request.Context(), claims.UserID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, MeResponse{
		UserID:  user.ID.Hex(),
		Email:   user.Email,
		Name:    user.Name,
		Role:    claims.Role,
		DbRole:  user.Role,
		Picture: user.Picture,
	})
}

// SwitchRole handles toggling active roles for ADMIN database users.
// POST /api/auth/switch-role
func (h *Handler) SwitchRole(c *gin.Context) {
	claims := middleware.GetCurrentUser(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	var req SwitchRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Role != RoleUser && req.Role != RoleAdmin {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid target role"})
		return
	}

	user, err := h.svc.GetUserByID(c.Request.Context(), claims.UserID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Only admin database users can switch roles
	if user.Role != RoleAdmin && req.Role == RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to switch to admin role"})
		return
	}

	token, err := h.svc.GenerateJWTWithRole(user, req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token: " + err.Error()})
		return
	}

	h.setSessionCookie(c, token)
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"userId":  user.ID.Hex(),
			"id":      user.ID.Hex(),
			"email":   user.Email,
			"name":    user.Name,
			"role":    req.Role,
			"dbRole":  user.Role,
			"picture": user.Picture,
		},
	})
}

// Logout clears the browser session cookie.
func (h *Handler) Logout(c *gin.Context) {
	h.clearCookie(c, SessionCookieName, "/")
	c.Status(http.StatusNoContent)
}

func (h *Handler) setSessionCookie(c *gin.Context, token string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   h.cookieMaxAge,
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *Handler) clearCookie(c *gin.Context, name, path string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     path,
		MaxAge:   -1,
		Expires:  time.Unix(1, 0),
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

func randomToken(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// GetUserSuggestions returns user email recommendations based on search input.
// GET /api/users/suggestions?q=&page=&limit=
func (h *Handler) GetUserSuggestions(c *gin.Context) {
	q := c.Query("q")
	page, limit := query.Pagination(c.Request.URL.Query(), suggestionLimit, suggestionMaxLimit)
	suggestions, total, err := h.svc.GetEmailSuggestions(c.Request.Context(), q, int64(page), int64(limit))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  suggestions,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}
