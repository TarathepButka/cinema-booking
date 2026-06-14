package auth

import (
	"context"
	"fmt"

	"backend/config"
	sharedauth "backend/internal/shared/auth"

	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	googleoauth "google.golang.org/api/oauth2/v2"
)

// Service handles authentication: Google OAuth + JWT + admin login.
type Service struct {
	cfg         *config.Config
	userRepo    *UserRepo
	oauthConfig *oauth2.Config
}

// NewService creates a new auth Service.
func NewService(cfg *config.Config, userRepo *UserRepo) *Service {
	oauthConfig := &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.GoogleRedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	return &Service{
		cfg:         cfg,
		userRepo:    userRepo,
		oauthConfig: oauthConfig,
	}
}

// GetGoogleAuthURL returns the Google OAuth consent URL.
func (s *Service) GetGoogleAuthURL(state string) string {
	return s.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// GetFrontendURL returns the frontend base URL.
func (s *Service) GetFrontendURL() string {
	return s.cfg.FrontendURL
}

// HandleGoogleCallback exchanges the code for a user, creates/updates the user, and returns a JWT.
func (s *Service) HandleGoogleCallback(ctx context.Context, code string) (*User, string, error) {
	token, err := s.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, "", fmt.Errorf("oauth exchange: %w", err)
	}

	httpClient := s.oauthConfig.Client(ctx, token)
	svc, err := googleoauth.New(httpClient)
	if err != nil {
		return nil, "", fmt.Errorf("google oauth client: %w", err)
	}

	info, err := svc.Userinfo.Get().Do()
	if err != nil {
		return nil, "", fmt.Errorf("get google userinfo: %w", err)
	}

	user := &User{
		Email:    info.Email,
		Name:     info.Name,
		Role:     RoleUser,
		GoogleID: &info.Id,
		Picture:  info.Picture,
	}

	saved, err := s.userRepo.UpsertByGoogleID(ctx, user)
	if err != nil {
		return nil, "", fmt.Errorf("upsert user: %w", err)
	}

	jwtToken, err := s.GenerateJWT(saved)
	if err != nil {
		return nil, "", err
	}

	return saved, jwtToken, nil
}

// AdminLogin validates admin credentials and returns a JWT.
func (s *Service) AdminLogin(ctx context.Context, email, password string) (*User, string, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, "", fmt.Errorf("find admin: %w", err)
	}
	if user == nil || user.Role != RoleAdmin {
		return nil, "", fmt.Errorf("invalid credentials")
	}
	if user.PasswordHash == nil {
		return nil, "", fmt.Errorf("admin has no password set")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password)); err != nil {
		return nil, "", fmt.Errorf("invalid credentials")
	}

	jwtToken, err := s.GenerateJWT(user)
	if err != nil {
		return nil, "", err
	}
	return user, jwtToken, nil
}

// GetEmailSuggestions returns paginated user emails matching the query.
func (s *Service) GetEmailSuggestions(ctx context.Context, query string, page, limit int64) ([]string, int64, error) {
	return s.userRepo.FindEmailSuggestions(ctx, query, page, limit)
}

// GenerateJWT creates a signed JWT token for the given user.
func (s *Service) GenerateJWT(user *User) (string, error) {
	return sharedauth.GenerateJWT(user.ID.Hex(), user.Email, user.Role, s.cfg.JWTSecret, s.cfg.JWTExpireHours)
}

// ValidateJWT parses and validates a JWT token string.
// Signature matches middleware.TokenValidator — passed directly to AuthMiddleware in main.go.
func (s *Service) ValidateJWT(tokenString string) (*sharedauth.Claims, error) {
	return sharedauth.ValidateJWT(tokenString, s.cfg.JWTSecret)
}

// GetUserByID fetches a user by their ID.
func (s *Service) GetUserByID(ctx context.Context, idStr string) (*User, error) {
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}
	return s.userRepo.FindByID(ctx, id)
}

// GenerateJWTWithRole creates a signed JWT token for the user with an overridden role.
func (s *Service) GenerateJWTWithRole(user *User, role sharedauth.Role) (string, error) {
	return sharedauth.GenerateJWT(user.ID.Hex(), user.Email, role, s.cfg.JWTSecret, s.cfg.JWTExpireHours)
}
