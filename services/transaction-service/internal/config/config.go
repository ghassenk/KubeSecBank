package config

import (
	"fmt"
	"os"
)

type Config struct {
	DBHost            string
	DBPort            string
	DBUser            string
	DBPassword        string
	DBName            string
	NatsURL           string
	AuthServiceURL    string
	AccountServiceURL string
	ServerPort        string
}

func Load() *Config {
	return &Config{
		DBHost:            getEnv("DB_HOST", "localhost"),
		DBPort:            getEnv("DB_PORT", "5432"),
		DBUser:            getEnv("DB_USER", "postgres"),
		DBPassword:        getEnv("DB_PASSWORD", "postgres"),
		DBName:            getEnv("DB_NAME", "transactions"),
		NatsURL:           getEnv("NATS_URL", "nats://localhost:4222"),
		AuthServiceURL:    getEnv("AUTH_SERVICE_URL", "http://localhost:8081"),
		AccountServiceURL: getEnv("ACCOUNT_SERVICE_URL", "http://localhost:8082"),
		ServerPort:        getEnv("SERVER_PORT", "8083"),
	}
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName,
	)
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
