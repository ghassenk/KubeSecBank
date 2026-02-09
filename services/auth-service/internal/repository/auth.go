package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/ghassenk/KubeSecBank/services/auth-service/internal/models"
)

// AuthRepository defines the interface for auth-related data operations.
type AuthRepository interface {
	// Session operations
	CreateSession(ctx context.Context, session *models.Session) error
	GetSessionByToken(ctx context.Context, token string) (*models.Session, error)
	DeleteSession(ctx context.Context, token string) error
	DeleteSessionsByUserID(ctx context.Context, userID string) error

	// Login attempt operations
	RecordLoginAttempt(ctx context.Context, attempt *models.LoginAttempt) error
	GetRecentFailedAttempts(ctx context.Context, email string, since time.Time) (int, error)

	// Token blacklist (Redis)
	BlacklistToken(ctx context.Context, token string, expiry time.Duration) error
	IsTokenBlacklisted(ctx context.Context, token string) (bool, error)

	// Session cache (Redis)
	CacheSession(ctx context.Context, token string, session *models.Session, expiry time.Duration) error
	GetCachedSession(ctx context.Context, token string) (*models.Session, error)
	InvalidateCachedSession(ctx context.Context, token string) error
}

// PostgresAuthRepository implements AuthRepository using PostgreSQL and Redis.
type PostgresAuthRepository struct {
	db    *sql.DB
	redis *redis.Client
}

// NewPostgresAuthRepository creates a new repository backed by PostgreSQL and Redis.
func NewPostgresAuthRepository(db *sql.DB, redisClient *redis.Client) *PostgresAuthRepository {
	return &PostgresAuthRepository{
		db:    db,
		redis: redisClient,
	}
}

// ---------------------------------------------------------------------------
// Session operations (PostgreSQL)
// ---------------------------------------------------------------------------

func (r *PostgresAuthRepository) CreateSession(ctx context.Context, session *models.Session) error {
	query := `
		INSERT INTO sessions (id, user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.ExecContext(ctx, query,
		session.ID, session.UserID, session.Token, session.ExpiresAt, session.CreatedAt,
	)
	return err
}

func (r *PostgresAuthRepository) GetSessionByToken(ctx context.Context, token string) (*models.Session, error) {
	query := `
		SELECT id, user_id, token, expires_at, created_at
		FROM sessions
		WHERE token = $1 AND expires_at > NOW()`
	s := &models.Session{}
	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&s.ID, &s.UserID, &s.Token, &s.ExpiresAt, &s.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found or expired")
	}
	return s, err
}

func (r *PostgresAuthRepository) DeleteSession(ctx context.Context, token string) error {
	query := `DELETE FROM sessions WHERE token = $1`
	_, err := r.db.ExecContext(ctx, query, token)
	return err
}

func (r *PostgresAuthRepository) DeleteSessionsByUserID(ctx context.Context, userID string) error {
	query := `DELETE FROM sessions WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

// ---------------------------------------------------------------------------
// Login attempt operations (PostgreSQL)
// ---------------------------------------------------------------------------

func (r *PostgresAuthRepository) RecordLoginAttempt(ctx context.Context, attempt *models.LoginAttempt) error {
	query := `
		INSERT INTO login_attempts (id, email, success, ip_address, created_at)
		VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.ExecContext(ctx, query,
		attempt.ID, attempt.Email, attempt.Success, attempt.IPAddress, attempt.CreatedAt,
	)
	return err
}

func (r *PostgresAuthRepository) GetRecentFailedAttempts(ctx context.Context, email string, since time.Time) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM login_attempts
		WHERE email = $1 AND success = false AND created_at > $2`
	var count int
	err := r.db.QueryRowContext(ctx, query, email, since).Scan(&count)
	return count, err
}

// ---------------------------------------------------------------------------
// Token blacklist (Redis)
// ---------------------------------------------------------------------------

const blacklistPrefix = "blacklist:"

func (r *PostgresAuthRepository) BlacklistToken(ctx context.Context, token string, expiry time.Duration) error {
	return r.redis.Set(ctx, blacklistPrefix+token, "1", expiry).Err()
}

func (r *PostgresAuthRepository) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	val, err := r.redis.Exists(ctx, blacklistPrefix+token).Result()
	if err != nil {
		return false, err
	}
	return val > 0, nil
}

// ---------------------------------------------------------------------------
// Session cache (Redis)
// ---------------------------------------------------------------------------

const sessionCachePrefix = "session:"

func (r *PostgresAuthRepository) CacheSession(ctx context.Context, token string, session *models.Session, expiry time.Duration) error {
	// Store a simple mapping of token -> userID for fast lookups.
	return r.redis.Set(ctx, sessionCachePrefix+token, session.UserID, expiry).Err()
}

func (r *PostgresAuthRepository) GetCachedSession(ctx context.Context, token string) (*models.Session, error) {
	userID, err := r.redis.Get(ctx, sessionCachePrefix+token).Result()
	if err == redis.Nil {
		return nil, nil // cache miss
	}
	if err != nil {
		return nil, err
	}
	return &models.Session{
		Token:  token,
		UserID: userID,
	}, nil
}

func (r *PostgresAuthRepository) InvalidateCachedSession(ctx context.Context, token string) error {
	return r.redis.Del(ctx, sessionCachePrefix+token).Err()
}
