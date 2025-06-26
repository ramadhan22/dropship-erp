package repository

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func TestGetAccountBalancesBetween_DifferentDates(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New failed: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewJournalRepo(sqlxDB)

	ctx := context.Background()
	from1 := time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC)
	to1 := from1.AddDate(0, 1, 0).Add(-time.Nanosecond)

	query := regexp.QuoteMeta(`
        SELECT
          a.account_id,
          a.account_code,
          a.account_name,
          a.account_type,
          a.parent_id,
          COALESCE(SUM(
            CASE WHEN jl.is_debit THEN jl.amount ELSE -jl.amount END
          ), 0) AS balance
        FROM accounts a
        LEFT JOIN (
          SELECT jl.account_id, jl.is_debit, jl.amount
            FROM journal_lines jl
            JOIN journal_entries je ON jl.journal_id = je.journal_id
           WHERE je.entry_date BETWEEN $1 AND $2
             AND ($3 = '' OR je.shop_username = $3)
        ) jl ON a.account_id = jl.account_id
        GROUP BY
          a.account_id, a.account_code, a.account_name,
          a.account_type, a.parent_id
        ORDER BY a.account_code;`)

	rows1 := sqlmock.NewRows([]string{"account_id", "account_code", "account_name", "account_type", "parent_id", "balance"}).
		AddRow(int64(1), "100", "Cash", "Asset", nil, 100.0)
	mock.ExpectQuery(query).WithArgs(from1, to1, "ShopA").WillReturnRows(rows1)

	res1, err := repo.GetAccountBalancesBetween(ctx, "ShopA", from1, to1)
	if err != nil {
		t.Fatalf("GetAccountBalancesBetween failed: %v", err)
	}
	if len(res1) != 1 || res1[0].Balance != 100 {
		t.Fatalf("unexpected result for May: %+v", res1)
	}

	from2 := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	to2 := from2.AddDate(0, 1, 0).Add(-time.Nanosecond)
	rows2 := sqlmock.NewRows([]string{"account_id", "account_code", "account_name", "account_type", "parent_id", "balance"}).
		AddRow(int64(1), "100", "Cash", "Asset", nil, 0.0)
	mock.ExpectQuery(query).WithArgs(from2, to2, "ShopA").WillReturnRows(rows2)

	res2, err := repo.GetAccountBalancesBetween(ctx, "ShopA", from2, to2)
	if err != nil {
		t.Fatalf("GetAccountBalancesBetween failed: %v", err)
	}
	if len(res2) != 1 || res2[0].Balance != 0 {
		t.Fatalf("unexpected result for June: %+v", res2)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
