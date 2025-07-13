package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

const insertJournalSQL = `
INSERT INTO journal_entries (
  entry_date, description, source_type, source_id, shop_username, store, created_at
) VALUES (
  :entry_date, :description, :source_type, :source_id, :shop_username, :store, :created_at
) RETURNING journal_id`

// JournalRepo manages the journal_entries and journal_lines tables,
// which are at the heart of double-entry bookkeeping.
type JournalRepo struct {
	db DBTX
}

// NewJournalRepo constructs a new JournalRepo.
func NewJournalRepo(db DBTX) *JournalRepo {
	return &JournalRepo{db: db}
}

// CreateJournalEntry inserts a row into journal_entries and returns the new journal_id.
// We need this so we can capture the returned primary key for inserting lines.
func (r *JournalRepo) CreateJournalEntry(ctx context.Context, e *models.JournalEntry) (int64, error) {
	if e.SourceType != "" && e.SourceID != "" {
		if old, err := r.GetJournalEntryBySource(ctx, e.SourceType, e.SourceID); err == nil && old != nil {
			return 0, fmt.Errorf("journal entry exists")
		}
	}
	rows, err := sqlx.NamedQueryContext(ctx, r.db, insertJournalSQL, e)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	if rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return 0, err
		}
		return id, nil
	}
	return 0, fmt.Errorf("no id returned")
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

// JournalLineDetail represents a journal line joined with its account name.
type JournalLineDetail struct {
	LineID      int64   `db:"line_id" json:"line_id"`
	JournalID   int64   `db:"journal_id" json:"journal_id"`
	AccountID   int64   `db:"account_id" json:"account_id"`
	AccountName string  `db:"account_name" json:"account_name"`
	IsDebit     bool    `db:"is_debit" json:"is_debit"`
	Amount      float64 `db:"amount" json:"amount"`
	Memo        *string `db:"memo" json:"memo"`
}

// EntryWithLines bundles a journal entry with its lines.
type EntryWithLines struct {
	Entry models.JournalEntry `json:"entry"`
	Lines []JournalLineDetail `json:"lines"`
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
        FROM accounts a
        LEFT JOIN (
          SELECT jl.account_id, jl.is_debit, jl.amount
            FROM journal_lines jl
            JOIN journal_entries je ON jl.journal_id = je.journal_id
           WHERE je.entry_date <= $1
             AND ($2 = '' OR je.shop_username = $2)
        ) jl ON a.account_id = jl.account_id
        GROUP BY
          a.account_id, a.account_code, a.account_name,
          a.account_type, a.parent_id
        ORDER BY a.account_code;`

	args := []interface{}{asOfDate, shop}

	var result []AccountBalance
	if err := r.db.SelectContext(ctx, &result, query, args...); err != nil {
		return nil, fmt.Errorf("GetAccountBalancesAsOf: %w", err)
	}
	return result, nil
}

// GetAccountBalancesBetween returns each account's net balance within the given
// date range. It sums debit amounts as positive and credit amounts as negative
// for journal entries between the from and to dates (inclusive).
func (r *JournalRepo) GetAccountBalancesBetween(
	ctx context.Context,
	shop string,
	from, to time.Time,
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
        ORDER BY a.account_code;`

	args := []interface{}{from, to, shop}

	var result []AccountBalance
	if err := r.db.SelectContext(ctx, &result, query, args...); err != nil {
		return nil, fmt.Errorf("GetAccountBalancesBetween: %w", err)
	}
	return result, nil
}

// GetLinesByJournalID returns all journal lines for a given journal entry
// joined with the account name.
func (r *JournalRepo) GetLinesByJournalID(ctx context.Context, id int64) ([]JournalLineDetail, error) {
	var list []JournalLineDetail
	err := r.db.SelectContext(ctx, &list, `
                SELECT jl.line_id, jl.journal_id, jl.account_id,
                       a.account_name, jl.is_debit, jl.amount, jl.memo
                  FROM journal_lines jl
                  JOIN accounts a ON jl.account_id = a.account_id
                 WHERE jl.journal_id = $1
                 ORDER BY jl.line_id
        `, id)
	if list == nil {
		list = []JournalLineDetail{}
	}
	return list, err
}

// GetJournalEntry fetches a journal entry by id.
func (r *JournalRepo) GetJournalEntry(ctx context.Context, id int64) (*models.JournalEntry, error) {
	var je models.JournalEntry
	if err := r.db.GetContext(ctx, &je, `SELECT * FROM journal_entries WHERE journal_id=$1`, id); err != nil {
		return nil, err
	}
	return &je, nil
}

// ListJournalEntries returns all entries ordered by date desc.
// ListJournalEntries returns journal entries filtered by optional date range and
// description substring. Empty strings are ignored.
func (r *JournalRepo) ListJournalEntries(
	ctx context.Context,
	from, to, desc string,
) ([]models.JournalEntry, error) {
	var list []models.JournalEntry
	query := `SELECT * FROM journal_entries
                WHERE ($1 = '' OR DATE(entry_date) >= $1::date)
                  AND ($2 = '' OR DATE(entry_date) <= $2::date)
                  AND ($3 = '' OR COALESCE(description,'') ILIKE '%' || $3 || '%')
                ORDER BY entry_date DESC`
	err := r.db.SelectContext(ctx, &list, query, from, to, desc)
	if list == nil {
		list = []models.JournalEntry{}
	}
	return list, err
}

// DeleteJournalEntry removes the entry (lines cascade).
func (r *JournalRepo) DeleteJournalEntry(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM journal_entries WHERE journal_id=$1`, id)
	return err
}

// GetJournalEntryBySource fetches a journal entry by its source type and source ID.
func (r *JournalRepo) GetJournalEntryBySource(ctx context.Context, sourceType, sourceID string) (*models.JournalEntry, error) {
	var je models.JournalEntry
	err := r.db.GetContext(ctx, &je,
		`SELECT * FROM journal_entries WHERE source_type=$1 AND source_id=$2 LIMIT 1`,
		sourceType, sourceID)
	if err != nil {
		return nil, err
	}
	return &je, nil
}

// ListEntriesBySourceID returns all journal entries that share the given source_id.
func (r *JournalRepo) ListEntriesBySourceID(ctx context.Context, sourceID string) ([]models.JournalEntry, error) {
	var list []models.JournalEntry
	err := r.db.SelectContext(ctx, &list,
		`SELECT * FROM journal_entries WHERE source_id=$1 OR source_id LIKE $1 || '-%' ORDER BY journal_id`,
		sourceID)
	if list == nil {
		list = []models.JournalEntry{}
	}
	return list, err
}

// UpdateJournalLineAmount updates the amount of a journal line identified by line_id.
func (r *JournalRepo) UpdateJournalLineAmount(ctx context.Context, lineID int64, amount float64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE journal_lines SET amount=$1 WHERE line_id=$2`, amount, lineID)
	return err
}
