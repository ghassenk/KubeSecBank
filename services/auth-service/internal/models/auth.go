package models

import "time"

// Credentials represents a login request payload.
type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// TokenPair holds the access and refresh tokens returned after authentication.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Session represents an active user session stored in the database.
type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// LoginAttempt records each login attempt for auditing and rate limiting.
type LoginAttempt struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Success   bool      `json:"success"`
	IPAddress string    `json:"ip_address"`
	CreatedAt time.Time `json:"created_at"`
}

// Claims holds the JWT custom claims used across the auth service.
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

// ErrorResponse is the standard JSON error envelope.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// TokenValidationResponse is returned by the ValidateToken endpoint.
type TokenValidationResponse struct {
	Valid  bool   `json:"valid"`
	UserID string `json:"user_id,omitempty"`
	Email  string `json:"email,omitempty"`
}
