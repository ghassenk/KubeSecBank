package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"

	"github.com/kubesec-bank/auth-service/internal/config"
	"github.com/kubesec-bank/auth-service/internal/handlers"
	"github.com/kubesec-bank/auth-service/internal/middleware"
	"github.com/kubesec-bank/auth-service/internal/repository"
)

func main() {
	// Load configuration.
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Connect to PostgreSQL.
	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	log.Println("connected to PostgreSQL")

	// Connect to Redis.
	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("failed to ping Redis: %v", err)
	}
	defer redisClient.Close()
	log.Println("connected to Redis")

	// Initialise repository and handlers.
	repo := repository.NewPostgresAuthRepository(db, redisClient)
	authHandler := handlers.NewAuthHandler(repo, cfg)

	// Set up middleware.
	jwtMiddleware := middleware.JWTAuth(cfg.JWTSecret)
	rateLimiter := middleware.NewRateLimiter(60, 1*time.Minute)

	// Set up routes.
	mux := http.NewServeMux()

	// Health check.
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Public routes (rate limited but no auth required).
	mux.Handle("/api/v1/auth/login", rateLimiter.Middleware(http.HandlerFunc(authHandler.Login)))
	mux.Handle("/api/v1/auth/refresh", rateLimiter.Middleware(http.HandlerFunc(authHandler.RefreshToken)))

	// Service-to-service validation (rate limited, no user auth).
	mux.Handle("/api/v1/auth/validate", rateLimiter.Middleware(http.HandlerFunc(authHandler.ValidateToken)))

	// Protected routes (require valid JWT).
	mux.Handle("/api/v1/auth/logout", rateLimiter.Middleware(jwtMiddleware(http.HandlerFunc(authHandler.Logout))))

	// Create HTTP server.
	addr := fmt.Sprintf(":%d", cfg.ServerPort)
	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine.
	go func() {
		log.Printf("auth-service listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Graceful shutdown.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down auth-service...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	log.Println("auth-service stopped")
}
