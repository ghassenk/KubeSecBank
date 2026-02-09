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
	"github.com/nats-io/nats.go"

	"github.com/ghassenk/KubeSecBank/services/transaction-service/internal/config"
	"github.com/ghassenk/KubeSecBank/services/transaction-service/internal/handlers"
	"github.com/ghassenk/KubeSecBank/services/transaction-service/internal/middleware"
	"github.com/ghassenk/KubeSecBank/services/transaction-service/internal/repository"
)

func main() {
	cfg := config.Load()

	// Connect to PostgreSQL
	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Connected to PostgreSQL")

	// Connect to NATS
	nc, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()
	log.Println("Connected to NATS")

	// Set up dependencies
	repo := repository.NewPostgresRepository(db)
	txnHandler := handlers.NewTransactionHandler(repo, nc, cfg.AccountServiceURL)
	authMiddleware := middleware.Auth(cfg.AuthServiceURL)

	// Routes
	mux := http.NewServeMux()

	// Health check (no auth required)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"healthy"}`)
	})

	// Protected routes
	protected := http.NewServeMux()
	protected.HandleFunc("POST /transactions/transfer", txnHandler.CreateTransfer)
	protected.HandleFunc("GET /transactions/{id}", txnHandler.GetTransaction)
	protected.HandleFunc("GET /transactions", txnHandler.ListTransactions)

	mux.Handle("/transactions/", authMiddleware(protected))
	mux.Handle("/transactions", authMiddleware(protected))

	// Server
	server := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Transaction service listening on :%s", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-done
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced shutdown: %v", err)
	}

	nc.Drain()
	log.Println("Server stopped gracefully")
}
