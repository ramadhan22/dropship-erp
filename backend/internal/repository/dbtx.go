package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// DBTX is implemented by both *sqlx.DB and *sqlx.Tx.
type DBTX interface {
	sqlx.ExtContext
	sqlx.PreparerContext
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}
