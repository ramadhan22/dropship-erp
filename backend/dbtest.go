// backend/cmd/dbtest/main.go
package main

import (
	"fmt"

	"github.com/ramadhan22/dropship-erp/backend/internal/config"
	"github.com/ramadhan22/dropship-erp/backend/internal/db"
)

func main() {
	// Load config to get the database URL
	cfg := config.MustLoadConfig()

	// Connect to Postgres
	sqlxDB, err := db.ConnectPostgres(cfg.Database.URL)
	if err != nil {
		panic(fmt.Errorf("ConnectPostgres failed: %w", err))
	}
	fmt.Println("Connected to Postgres successfully")

	// Run migrations (point to the migrations folder)
	migrationsDir := "file:///Users/rama/Documents/dropship-erp/backend/internal/migrations"
	fmt.Println("Migration URL is:", migrationsDir)
	if err := db.RunMigrations(cfg.Database.URL, migrationsDir); err != nil {
		panic(fmt.Errorf("RunMigrations failed: %w", err))
	}
	fmt.Println("Migrations ran successfully: all tables should now exist")

	// Optionally, verify by querying pg tables:
	var version string
	err = sqlxDB.Get(&version, "SELECT version();")
	if err != nil {
		panic(fmt.Errorf("SELECT version() failed: %w", err))
	}
	fmt.Println("Postgres version:", version)
}
