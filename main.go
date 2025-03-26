package main

import (
	"context"
	"errors"
	"github.com/gorilla/mux"
	"heart-rate-server/internal/config"
	"heart-rate-server/internal/handlers"
	"heart-rate-server/internal/middleware"
	"heart-rate-server/internal/storage"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize storage
	db, err := storage.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	redisClient, err := storage.InitRedis(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}

	// Initialize secure cookie
	secureCookie := middleware.NewSecureCookie(cfg.CookieHashKey, cfg.CookieBlockKey)

	// Create app with dependencies
	app := &handlers.App{
		DB:           db,
		Redis:        redisClient,
		Config:       cfg,
		SecureCookie: secureCookie,
	}

	// Create router
	r := mux.NewRouter()

	// Global middleware
	r.Use(middleware.LoggingMiddleware)
	r.Use(middleware.RecoveryMiddleware)
	r.Use(middleware.JSONContentTypeMiddleware)

	// Public routes
	publicRouter := r.PathPrefix("").Subrouter()
	publicRouter.HandleFunc("/health", handlers.HealthHandler).Methods("GET")
	publicRouter.HandleFunc("/register", app.RegisterHandler).Methods("POST")
	publicRouter.HandleFunc("/login", app.LoginHandler).Methods("POST")

	// Authenticated routes
	authRouter := r.PathPrefix("").Subrouter()
	authRouter.Use(middleware.AuthMiddleware(secureCookie, app.Config))
	authRouter.HandleFunc("/receive_data", app.ReceiveDataHandler).Methods("POST")
	authRouter.HandleFunc("/latest-heart-rate", app.LatestHeartRateHandler).Methods("GET")
	authRouter.HandleFunc("/logout", app.LogoutHandler).Methods("POST")

	// Create server
	server := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
