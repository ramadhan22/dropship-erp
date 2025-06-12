package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// JournalRepo manages the journal_entries and journal_lines tables,
// which are at the heart of double-entry bookkeeping.
type JournalRepo struct {
	db *sqlx.DB
}

// NewJournalRepo constructs a new JournalRepo.
func NewJournalRepo(db *sqlx.DB) *JournalRepo {
	return &JournalRepo{db: db}
}

// CreateJournalEntry inserts a row into journal_entries and returns the new journal_id.
// We need this so we can capture the returned primary key for inserting lines.
func (r *JournalRepo) CreateJournalEntry(ctx context.Context, e *models.JournalEntry) (int64, error) {
	var newID int64
	err := r.db.QueryRowxContext(ctx,
		`INSERT INTO journal_entries
          (entry_date, description, source_type, source_id, shop_username)
         VALUES ($1, $2, $3, $4, $5)
         RETURNING journal_id`,
		e.EntryDate, e.Description, e.SourceType, e.SourceID, e.ShopUsername,
	).Scan(&newID)
	if err != nil {
		return 0, err
	}
	return newID, nil
}

// InsertJournalLine inserts a single debit or credit row into journal_lines.
func (r *JournalRepo) InsertJournalLine(ctx context.Context, l *models.JournalLine) error {
	query := `
        INSERT INTO journal_lines (
          journal_id, account_id, is_debit, amount, memo
        ) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.ExecContext(ctx, query,
		l.JournalID, l.AccountID, l.IsDebit, l.Amount, l.Memo)
	return err
}

// GetJournalLinesByShopAndDate fetches all journal lines for a given shop within a date range.
// Joins through journal_entries to filter by shop_username and entry_date.
func (r *JournalRepo) GetJournalLinesByShopAndDate(
	ctx context.Context,
	shop string,
	from, to time.Time,
) ([]models.JournalLine, error) {
	var lines []models.JournalLine
	err := r.db.SelectContext(ctx, &lines,
		`SELECT jl.* 
         FROM journal_lines jl
         JOIN journal_entries je ON jl.journal_id = je.journal_id
         WHERE je.shop_username = $1
           AND je.entry_date BETWEEN $2 AND $3
         ORDER BY je.entry_date, jl.line_id`,
		shop, from, to)
	return lines, err
}

// AccountBalance is a helper type for producing Balance Sheet data.
// It represents the net (debit – credit) balance of one account as of a given date.
type AccountBalance struct {
	AccountID   int64   `db:"account_id" json:"account_id"`
	AccountCode string  `db:"account_code" json:"account_code"`
	AccountName string  `db:"account_name" json:"account_name"`
	AccountType string  `db:"account_type" json:"account_type"` // e.g. “Asset”/“Liability”/“Equity”
	ParentID    *int64  `db:"parent_id" json:"parent_id"`
	Balance     float64 `db:"balance" json:"balance"`
}

// GetAccountBalancesAsOf returns each account’s cumulative balance up to and including asOfDate.
// It sums debit amounts as positive and credit amounts as negative.
func (r *JournalRepo) GetAccountBalancesAsOf(
	ctx context.Context,
	shop string,
	asOfDate time.Time,
) ([]AccountBalance, error) {
	query := `
        SELECT
          a.account_id,
          a.account_code,
          a.account_name,
          a.account_type,
          a.parent_id,
          COALESCE(SUM(
            CASE WHEN jl.is_debit THEN jl.amount ELSE -jl.amount END
          ), 0) AS balance
        FROM journal_lines jl
        JOIN journal_entries je ON jl.journal_id = je.journal_id
        JOIN accounts a ON jl.account_id = a.account_id
        WHERE je.shop_username = $1
          AND je.entry_date <= $2
        GROUP BY
          a.account_id, a.account_code, a.account_name,
          a.account_type, a.parent_id
        ORDER BY a.account_code;
    `
	var result []AccountBalance
	if err := r.db.SelectContext(ctx, &result, query, shop, asOfDate); err != nil {
		return nil, fmt.Errorf("GetAccountBalancesAsOf: %w", err)
	}
	return result, nil
}
