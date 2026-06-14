package auth

import sharedauth "backend/internal/shared/auth"

// LoginRequest is the request body for POST /api/auth/login.
type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// MeResponse is the response body for GET /api/auth/me.
type MeResponse struct {
	UserID  string          `json:"userId"`
	Email   string          `json:"email"`
	Name    string          `json:"name"`
	Role    sharedauth.Role `json:"role"`
	DbRole  sharedauth.Role `json:"dbRole"`
	Picture string          `json:"picture"`
}

// SwitchRoleRequest is the request body for POST /api/auth/switch-role.
type SwitchRoleRequest struct {
	Role sharedauth.Role `json:"role" binding:"required"`
}
