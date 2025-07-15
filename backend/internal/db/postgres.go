// backend/internal/db/postgres.go
package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // pq registers driver "postgres"
)

// ConnectPostgres opens a connection pool to PostgreSQL using the given DSN.
// It returns a *sqlx.DB that you can use throughout your application.
func ConnectPostgres(dsn string) (*sqlx.DB, error) {
	// Use "postgres" since pq registers under that name
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	// Set default pool settings (can be overridden with ConnectPostgresWithConfig)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	return db, nil
}

// ConnectPostgresWithConfig opens a connection pool with custom configuration
func ConnectPostgresWithConfig(dsn string, maxOpenConns, maxIdleConns int, connMaxLifetime string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	
	// Parse and set connection max lifetime if provided
	if connMaxLifetime != "" {
		if duration, err := time.ParseDuration(connMaxLifetime); err == nil {
			db.SetConnMaxLifetime(duration)
		}
	}

	return db, nil
}

// RunMigrations applies all up migrations from the specified directory.
// `migrationsDir` should be a file:// URL or a relative path to your SQL files.
// For example: "file://./internal/migrations"
func RunMigrations(dsn, migrationsDir string) error {
	// Open a plain *sql.DB (not *sqlx.DB) with the same "postgres" driver
	sqldb, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open sql.DB for migrations: %w", err)
	}
	defer sqldb.Close()

	driver, err := postgres.WithInstance(sqldb, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		migrationsDir,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	// Run all up migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("error applying migrations: %w", err)
	}

	return nil
}
