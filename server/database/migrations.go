package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
)

// RunMigrations runs all migrations from the migrations directory
func RunMigrations(db *sql.DB) error {
	migrationsDir := "server/migrations"

	// Read migration files
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return fmt.Errorf("failed to read migration files: %w", err)
	}

	// Sort files by name (assuming they start with numbers like 001_, 002_, etc.)
	// For now, just run them in order
	for _, file := range files {
		if err := runMigration(db, file); err != nil {
			return fmt.Errorf("failed to run migration %s: %w", file, err)
		}
	}

	return nil
}

// runMigration runs a single migration file
func runMigration(db *sql.DB, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	_, err = db.Exec(string(data))
	if err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	return nil
}
