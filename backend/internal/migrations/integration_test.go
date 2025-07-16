package migrations

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

const testDatabaseURL = "postgres://erp_user:erp_pass@localhost:5432/dropship_erp_test?sslmode=disable"

func TestMigrationWithRealDatabase(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Only run this test if we have a test database available
	if os.Getenv("TEST_DATABASE_URL") == "" && !databaseExists(testDatabaseURL) {
		t.Skip("No test database available. Set TEST_DATABASE_URL or have test DB at localhost")
	}

	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = testDatabaseURL
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Skipf("Cannot connect to test database: %v", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		t.Skipf("Cannot ping test database: %v", err)
	}

	// Clean up any existing migrations for a clean test
	_, err = db.Exec("DROP TABLE IF EXISTS schema_migrations CASCADE")
	if err != nil {
		t.Logf("Warning: Could not drop schema_migrations table: %v", err)
	}

	// Run migrations
	t.Log("Running migrations against test database...")
	err = Run(db)
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Verify that key tables exist and have seed data
	tables := []string{
		"accounts",
		"asset_accounts",
		"jenis_channels",
		"stores",
	}

	for _, table := range tables {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count)
		if err != nil {
			t.Errorf("Failed to query table %s: %v", table, err)
			continue
		}

		if count == 0 {
			t.Errorf("Table %s exists but has no data (expected seed data)", table)
		} else {
			t.Logf("Table %s has %d rows", table, count)
		}
	}

	// Verify specific seed data
	t.Run("VerifyAccountSeedData", func(t *testing.T) {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM accounts WHERE account_code IN ('1', '2', '3', '4', '5')").Scan(&count)
		if err != nil {
			t.Errorf("Failed to verify root accounts: %v", err)
			return
		}
		if count != 5 {
			t.Errorf("Expected 5 root accounts, got %d", count)
		}
	})

	t.Run("VerifyAssetAccountSeedData", func(t *testing.T) {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM asset_accounts").Scan(&count)
		if err != nil {
			t.Errorf("Failed to verify asset accounts: %v", err)
			return
		}
		if count == 0 {
			t.Error("Expected asset accounts to be seeded")
		}
	})

	t.Run("VerifyChannelSeedData", func(t *testing.T) {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM jenis_channels WHERE jenis_channel = 'Shopee'").Scan(&count)
		if err != nil {
			t.Errorf("Failed to verify channel seed data: %v", err)
			return
		}
		if count == 0 {
			t.Error("Expected Shopee channel to be seeded")
		}
	})

	t.Log("All migration and seed data tests passed!")
}

func databaseExists(dbURL string) bool {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return false
	}
	defer db.Close()

	err = db.Ping()
	return err == nil
}
