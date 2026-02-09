package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/ghassenk/KubeSecBank/services/auth-service/internal/config"
	"github.com/ghassenk/KubeSecBank/services/auth-service/internal/middleware"
	"github.com/ghassenk/KubeSecBank/services/auth-service/internal/models"
	"github.com/ghassenk/KubeSecBank/services/auth-service/internal/repository"
)

// AuthHandler holds dependencies for authentication endpoints.
type AuthHandler struct {
	repo   repository.AuthRepository
	config *config.Config
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(repo repository.AuthRepository, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		repo:   repo,
		config: cfg,
	}
}

// Login validates credentials, records the attempt, and issues a token pair.
// POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var creds models.Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if creds.Email == "" || creds.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	// Check for too many recent failed attempts (simple brute-force protection).
	ctx := r.Context()
	failedCount, err := h.repo.GetRecentFailedAttempts(ctx, creds.Email, time.Now().Add(-15*time.Minute))
	if err != nil {
		log.Printf("error checking failed attempts: %v", err)
	}
	if failedCount >= 5 {
		writeError(w, http.StatusTooManyRequests, "too many failed login attempts, try again later")
		return
	}

	// -----------------------------------------------------------------------
	// TODO: Replace this stub with a real call to the account-service or a
	// local user lookup. For now we accept any non-empty credentials.
	// -----------------------------------------------------------------------
	userID := "user-" + creds.Email // placeholder
	authenticated := true           // placeholder

	// Record the login attempt.
	attempt := &models.LoginAttempt{
		ID:        generateID(),
		Email:     creds.Email,
		Success:   authenticated,
		IPAddress: r.RemoteAddr,
		CreatedAt: time.Now().UTC(),
	}
	if err := h.repo.RecordLoginAttempt(ctx, attempt); err != nil {
		log.Printf("error recording login attempt: %v", err)
	}

	if !authenticated {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	// Issue tokens.
	tokenPair, err := h.issueTokens(userID, creds.Email)
	if err != nil {
		log.Printf("error issuing tokens: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to issue tokens")
		return
	}

	// Persist session in DB and cache.
	session := &models.Session{
		ID:        generateID(),
		UserID:    userID,
		Token:     tokenPair.AccessToken,
		ExpiresAt: time.Now().UTC().Add(h.config.JWTExpiry),
		CreatedAt: time.Now().UTC(),
	}
	if err := h.repo.CreateSession(ctx, session); err != nil {
		log.Printf("error creating session: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to create session")
		return
	}
	_ = h.repo.CacheSession(ctx, session.Token, session, h.config.JWTExpiry)

	writeJSON(w, http.StatusOK, tokenPair)
}

// Logout invalidates the current session.
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Extract the raw token from the header for blacklisting.
	tokenStr := r.Header.Get("Authorization")
	if len(tokenStr) > 7 {
		tokenStr = tokenStr[7:] // strip "Bearer "
	}

	ctx := r.Context()

	// Blacklist the token so it cannot be reused.
	if err := h.repo.BlacklistToken(ctx, tokenStr, h.config.JWTExpiry); err != nil {
		log.Printf("error blacklisting token: %v", err)
	}

	// Remove session from DB and cache.
	if err := h.repo.DeleteSession(ctx, tokenStr); err != nil {
		log.Printf("error deleting session: %v", err)
	}
	_ = h.repo.InvalidateCachedSession(ctx, tokenStr)

	log.Printf("user %s logged out", claims.UserID)
	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out successfully"})
}

// RefreshToken issues a new token pair given a valid refresh token.
// POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		writeError(w, http.StatusBadRequest, "refresh_token is required")
		return
	}

	ctx := r.Context()

	// Check if the refresh token has been blacklisted.
	blacklisted, err := h.repo.IsTokenBlacklisted(ctx, req.RefreshToken)
	if err != nil {
		log.Printf("error checking blacklist: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if blacklisted {
		writeError(w, http.StatusUnauthorized, "token has been revoked")
		return
	}

	// Validate the refresh token.
	token, err := jwt.Parse(req.RefreshToken, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(h.config.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		writeError(w, http.StatusUnauthorized, "invalid or expired refresh token")
		return
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		writeError(w, http.StatusUnauthorized, "invalid token claims")
		return
	}

	userID, _ := mapClaims["user_id"].(string)
	email, _ := mapClaims["email"].(string)
	tokenType, _ := mapClaims["type"].(string)

	if tokenType != "refresh" {
		writeError(w, http.StatusUnauthorized, "not a refresh token")
		return
	}

	// Blacklist the old refresh token.
	_ = h.repo.BlacklistToken(ctx, req.RefreshToken, 7*24*time.Hour)

	// Issue new pair.
	newPair, err := h.issueTokens(userID, email)
	if err != nil {
		log.Printf("error issuing tokens: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to issue tokens")
		return
	}

	writeJSON(w, http.StatusOK, newPair)
}

// ValidateToken checks whether a token is valid and returns the claims.
// Used for service-to-service authentication.
// POST /api/v1/auth/validate
func (h *AuthHandler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Token == "" {
		writeError(w, http.StatusBadRequest, "token is required")
		return
	}

	ctx := r.Context()

	// Check blacklist first.
	blacklisted, err := h.repo.IsTokenBlacklisted(ctx, req.Token)
	if err != nil {
		log.Printf("error checking blacklist: %v", err)
	}
	if blacklisted {
		writeJSON(w, http.StatusOK, models.TokenValidationResponse{Valid: false})
		return
	}

	token, err := jwt.Parse(req.Token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(h.config.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		writeJSON(w, http.StatusOK, models.TokenValidationResponse{Valid: false})
		return
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		writeJSON(w, http.StatusOK, models.TokenValidationResponse{Valid: false})
		return
	}

	resp := models.TokenValidationResponse{
		Valid:  true,
		UserID: getStringClaim(mapClaims, "user_id"),
		Email:  getStringClaim(mapClaims, "email"),
	}
	writeJSON(w, http.StatusOK, resp)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func (h *AuthHandler) issueTokens(userID, email string) (*models.TokenPair, error) {
	now := time.Now().UTC()

	// Access token.
	accessClaims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"type":    "access",
		"iat":     now.Unix(),
		"exp":     now.Add(h.config.JWTExpiry).Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessStr, err := accessToken.SignedString([]byte(h.config.JWTSecret))
	if err != nil {
		return nil, err
	}

	// Refresh token (longer-lived).
	refreshClaims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"type":    "refresh",
		"iat":     now.Unix(),
		"exp":     now.Add(7 * 24 * time.Hour).Unix(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshStr, err := refreshToken.SignedString([]byte(h.config.JWTSecret))
	if err != nil {
		return nil, err
	}

	return &models.TokenPair{
		AccessToken:  accessStr,
		RefreshToken: refreshStr,
	}, nil
}

func generateID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func getStringClaim(m jwt.MapClaims, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, models.ErrorResponse{Error: msg})
}
