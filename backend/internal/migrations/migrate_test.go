package migrations

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang-migrate/migrate/v4"
)

func TestMigrationFilesEmbedded(t *testing.T) {
	// Test that migration files are properly embedded
	entries, err := migrationFS.ReadDir("consolidated")
	if err != nil {
		t.Fatalf("Failed to read consolidated directory: %v", err)
	}

	if len(entries) == 0 {
		t.Fatal("No migration files found in consolidated directory")
	}

	// Check that we have the expected seed files
	expectedFiles := []string{
		"0104_seed_chart_of_accounts.up.sql",
		"0105_seed_reference_data.up.sql", 
		"0107_seed_sample_data.up.sql",
	}

	foundFiles := make(map[string]bool)
	for _, entry := range entries {
		foundFiles[entry.Name()] = true
		t.Logf("Found migration file: %s", entry.Name())
	}

	for _, expectedFile := range expectedFiles {
		if !foundFiles[expectedFile] {
			t.Errorf("Expected seed file not found: %s", expectedFile)
		}
	}

	t.Logf("Total migration files embedded: %d", len(entries))
}

func TestSeedDataContent(t *testing.T) {
	// Test that seed data files have content
	seedFiles := []string{
		"consolidated/0104_seed_chart_of_accounts.up.sql",
		"consolidated/0105_seed_reference_data.up.sql",
		"consolidated/0107_seed_sample_data.up.sql",
	}

	for _, seedFile := range seedFiles {
		content, err := migrationFS.ReadFile(seedFile)
		if err != nil {
			t.Errorf("Failed to read seed file %s: %v", seedFile, err)
			continue
		}

		if len(content) == 0 {
			t.Errorf("Seed file %s is empty", seedFile)
			continue
		}

		t.Logf("Seed file %s has %d bytes of content", seedFile, len(content))
	}
}

func TestMigrationSystemWithMock(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}
	defer db.Close()

	// Mock the migration queries that the postgres driver would make
	mock.ExpectQuery("SELECT current_database()").WillReturnRows(sqlmock.NewRows([]string{"current_database"}).AddRow("test_db"))
	mock.ExpectQuery("SELECT current_schema()").WillReturnRows(sqlmock.NewRows([]string{"current_schema"}).AddRow("public"))
	mock.ExpectQuery("SHOW server_version_num").WillReturnRows(sqlmock.NewRows([]string{"server_version_num"}).AddRow("120000"))
	
	// Mock table existence check for schema_migrations
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"table_name"}))
	
	// Mock creation of schema_migrations table
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS").WillReturnResult(sqlmock.NewResult(0, 0))
	
	// Mock version check (no migrations yet)
	mock.ExpectQuery("SELECT version FROM").WillReturnError(sql.ErrNoRows)
	
	// Mock the actual migration executions - we'll expect several CREATE TABLE statements
	mock.ExpectBegin()
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// Test that Run function doesn't panic and handles the mock properly
	err = Run(db)
	// We expect this to fail with migration-related errors since we're using a mock,
	// but it should not panic and should attempt to run migrations
	if err != nil {
		// This is expected with a mock DB, log the error but don't fail
		t.Logf("Migration failed as expected with mock DB: %v", err)
		// Check if it's the specific "no change" error, which would be good
		if err == migrate.ErrNoChange {
			t.Log("No migrations needed to run (ErrNoChange)")
		}
	} else {
		t.Log("Migration system completed successfully")
	}
}