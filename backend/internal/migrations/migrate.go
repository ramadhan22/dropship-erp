package migrations

import (
	"database/sql"
	"embed"
	"log"
	"net/http"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/httpfs"
)

//go:embed consolidated/*.sql
var migrationFS embed.FS

// Run applies all pending migrations.
func Run(db *sql.DB) error {
	log.Printf("Starting database migrations...")
	
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Printf("Failed to create postgres driver: %v", err)
		return err
	}
	
	source, err := httpfs.New(http.FS(migrationFS), "consolidated")
	if err != nil {
		log.Printf("Failed to create migration source: %v", err)
		return err
	}
	
	m, err := migrate.NewWithInstance("httpfs", source, "postgres", driver)
	if err != nil {
		log.Printf("Failed to create migrate instance: %v", err)
		return err
	}
	
	// Check current version
	currentVersion, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		log.Printf("Failed to get current migration version: %v", err)
		return err
	}
	
	if err == migrate.ErrNilVersion {
		log.Printf("No previous migrations found, starting from version 0")
	} else {
		log.Printf("Current migration version: %d, dirty: %t", currentVersion, dirty)
	}
	
	// Apply migrations
	log.Printf("Applying database migrations...")
	err = m.Up()
	if err != nil {
		log.Printf("Migration result: %v", err)
	} else {
		log.Printf("All migrations applied successfully")
	}
	
	// Check final version
	finalVersion, dirty, versionErr := m.Version()
	if versionErr == nil {
		log.Printf("Final migration version: %d, dirty: %t", finalVersion, dirty)
	}
	
	return err
}
