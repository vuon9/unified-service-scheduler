package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/vuon9/unified-service-scheduler/internal/handler"
	"github.com/vuon9/unified-service-scheduler/internal/repository"
	"github.com/vuon9/unified-service-scheduler/internal/service"
)

func main() {
	// Configure structured logging
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	// Database path from env or default
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./scheduler.db"
	}

	// Migrations directory - check env first, then common locations
	// This is for demo - so I wanted to keep it simple
	// but for production application, migration could be moved out and to be running as a separated command,
	// depends on how the app infrastructure looks like
	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		// Try common locations: relative to binary, Docker path, local dev path
		candidates := []string{"./migrations", "/migrations", "../migrations"}
		for _, dir := range candidates {
			if info, err := os.Stat(dir); err == nil && info.IsDir() {
				migrationsDir = dir
				break
			}
		}
	}
	if migrationsDir == "" {
		slog.Warn("no migrations directory found, using ./migrations as default")
		migrationsDir = "./migrations"
	}

	// Initialize repository (opens DB and runs migrations)
	repo, err := repository.New(dbPath, migrationsDir)
	if err != nil {
		slog.Error("failed to initialize repository", "error", err)
		os.Exit(1)
	}
	defer repo.Close()

	// Initialize service
	svc := service.New(repo)

	// Create router
	router := handler.NewRouter(svc, repo)

	// Determine port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := fmt.Sprintf(":%s", port)
	slog.Info("starting server", "addr", addr, "db_path", dbPath, "migrations_dir", migrationsDir)

	if err := http.ListenAndServe(addr, router); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
