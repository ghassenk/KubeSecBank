package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the auth service.
type Config struct {
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	RedisAddr  string
	JWTSecret  string
	JWTExpiry  time.Duration
	ServerPort int
}

// Load reads configuration from environment variables with sensible defaults.
func Load() (*Config, error) {
	dbPort, err := getEnvInt("DB_PORT", 5432)
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}

	jwtExpiryMinutes, err := getEnvInt("JWT_EXPIRY", 15)
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_EXPIRY: %w", err)
	}

	serverPort, err := getEnvInt("SERVER_PORT", 8082)
	if err != nil {
		return nil, fmt.Errorf("invalid SERVER_PORT: %w", err)
	}

	cfg := &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     dbPort,
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "kubesec_auth"),
		RedisAddr:  getEnv("REDIS_ADDR", "localhost:6379"),
		JWTSecret:  getEnv("JWT_SECRET", "change-me-in-production"),
		JWTExpiry:  time.Duration(jwtExpiryMinutes) * time.Minute,
		ServerPort: serverPort,
	}

	return cfg, nil
}

// DSN returns the PostgreSQL connection string.
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName,
	)
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) (int, error) {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback, nil
	}
	return strconv.Atoi(val)
}
