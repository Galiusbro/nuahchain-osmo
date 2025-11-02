package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/osmosis-labs/osmosis/v30/server/api"
	"github.com/osmosis-labs/osmosis/v30/server/auth"
	"github.com/osmosis-labs/osmosis/v30/server/config"
	"github.com/osmosis-labs/osmosis/v30/server/database"
	"github.com/osmosis-labs/osmosis/v30/server/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		// Use standard log if config loading fails (logger not yet initialized)
		os.Stderr.WriteString("Failed to load configuration: " + err.Error() + "\n")
		os.Exit(1)
	}

	// Initialize logger
	appLogger, err := logger.New(logger.Config{
		Enabled:     cfg.Logger.Enabled,
		Level:       cfg.Logger.Level,
		Format:      cfg.Logger.Format,
		Environment: cfg.Logger.Environment,
	})
	if err != nil {
		os.Stderr.WriteString("Failed to initialize logger: " + err.Error() + "\n")
		os.Exit(1)
	}

	appLogger.Info("Server initializing...")

	// Initialize database
	db, err := database.New(cfg.Database.DSN(), database.DatabasePoolConfig{
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
		ConnMaxIdleTime: cfg.Database.ConnMaxIdleTime,
		ConnectTimeout:  cfg.Database.ConnectTimeout,
	})
	if err != nil {
		appLogger.WithError(err).Fatal("Failed to initialize database")
	}
	defer db.Close()

	appLogger.Info("Database connected successfully")

	// Run database migrations
	if err := database.RunMigrations(db.DB); err != nil {
		appLogger.WithError(err).Error("Failed to run migrations")
		// Don't fail - migrations might have already been run
	} else {
		appLogger.Info("Database migrations completed")
	}

	// Initialize authentication service
	authRepo := auth.NewRepository(db.DB)
	authService := auth.NewService(
		authRepo,
		cfg.Auth.JWTSecret,
		cfg.Auth.TokenExpiry,
		cfg.Auth.RefreshExpiry,
	)
	api.SetAuthService(authService)

	// Create HTTP router and set database health checker
	router := api.NewRouter(appLogger)
	api.SetDBHealthChecker(db)

	// Create HTTP server
	srv := &http.Server{
		Addr:         cfg.Server.Address,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		appLogger.WithField("address", cfg.Server.Address).Info("Server starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.WithError(err).Fatal("Server failed to start")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Server shutting down...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		appLogger.WithError(err).Fatal("Server forced to shutdown")
	}

	appLogger.Info("Server exited")
}
