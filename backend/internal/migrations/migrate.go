package migrations

import (
	"database/sql"
	"embed"
	"net/http"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/httpfs"
)

//go:embed *.sql
var migrationFS embed.FS

// Run applies all pending migrations.
func Run(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}
	source, err := httpfs.New(http.FS(migrationFS), ".")
	if err != nil {
		return err
	}
	m, err := migrate.NewWithInstance("httpfs", source, "postgres", driver)
	if err != nil {
		return err
	}
	return m.Up()
}
