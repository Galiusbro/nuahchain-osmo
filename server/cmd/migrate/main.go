package main

import (
	"log"

	"github.com/osmosis-labs/osmosis/v30/server/config"
	"github.com/osmosis-labs/osmosis/v30/server/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	db, err := database.New(cfg.Database.DSN(), database.DatabasePoolConfig{
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
		ConnMaxIdleTime: cfg.Database.ConnMaxIdleTime,
		ConnectTimeout:  cfg.Database.ConnectTimeout,
	})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := database.RunMigrations(db.DB); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	log.Println("database migrations applied successfully")
}
