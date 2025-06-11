// File: backend/internal/repository/repository_test.go

package repository

import (
	"fmt"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/ramadhan22/dropship-erp/backend/internal/config"
	"github.com/ramadhan22/dropship-erp/backend/internal/db"
)

var testDB *sqlx.DB

// ptrString is a helper to create *string from a literal.
func ptrString(s string) *string {
	return &s
}

// TestMain sets up the shared *sqlx.DB for all repository tests.
func TestMain(m *testing.M) {
	// Change working directory to project root (backend/) so config file can be found
	if err := os.Chdir("../../"); err != nil {
		panic(err)
	}

	// Load configuration
	cfg := config.MustLoadConfig()

	// Connect to Postgres (skip tests if not available)
	var err error
	testDB, err = db.ConnectPostgres(cfg.Database.URL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "skipping repository tests: %v\n", err)
		os.Exit(0)
	}

	// Run migrations (ensure tables exist)
	migrationsDir := "file://internal/migrations"
	_ = db.RunMigrations(cfg.Database.URL, migrationsDir)

	// Run tests
	code := m.Run()

	// Close DB and exit
	testDB.Close()
	os.Exit(code)
}
