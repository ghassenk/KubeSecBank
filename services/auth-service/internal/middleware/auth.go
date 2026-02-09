package middleware

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/kubesec-bank/auth-service/internal/models"
)

// contextKey is an unexported type for context keys in this package.
type contextKey string

const (
	// ClaimsContextKey is the key used to store JWT claims in the request context.
	ClaimsContextKey contextKey = "claims"
)

// JWTAuth returns middleware that validates the Authorization header
// and injects the parsed claims into the request context.
func JWTAuth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				http.Error(w, `{"error":"invalid authorization format"}`, http.StatusUnauthorized)
				return
			}

			tokenStr := parts[1]

			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(jwtSecret), nil
			})
			if err != nil || !token.Valid {
				http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			mapClaims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, `{"error":"invalid token claims"}`, http.StatusUnauthorized)
				return
			}

			claims := &models.Claims{
				UserID: getString(mapClaims, "user_id"),
				Email:  getString(mapClaims, "email"),
			}

			ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetClaims extracts claims from the request context.
func GetClaims(ctx context.Context) (*models.Claims, bool) {
	claims, ok := ctx.Value(ClaimsContextKey).(*models.Claims)
	return claims, ok
}

func getString(m jwt.MapClaims, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

// ---------------------------------------------------------------------------
// Rate Limiting Middleware (stub)
// ---------------------------------------------------------------------------

// RateLimiter provides a simple in-memory rate limiter.
// For production use, replace with a Redis-backed implementation.
type RateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a rate limiter that allows `limit` requests per `window`.
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// Middleware returns an HTTP middleware that rate-limits by client IP.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr

		rl.mu.Lock()
		now := time.Now()
		windowStart := now.Add(-rl.window)

		// Remove expired entries.
		var valid []time.Time
		for _, t := range rl.requests[ip] {
			if t.After(windowStart) {
				valid = append(valid, t)
			}
		}
		rl.requests[ip] = valid

		if len(valid) >= rl.limit {
			rl.mu.Unlock()
			http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
			return
		}

		rl.requests[ip] = append(rl.requests[ip], now)
		rl.mu.Unlock()

		next.ServeHTTP(w, r)
	})
}
