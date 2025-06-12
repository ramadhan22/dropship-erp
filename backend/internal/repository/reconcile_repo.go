package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// ReconcileRepo handles database operations for reconciled_transactions.
type ReconcileRepo struct {
	db *sqlx.DB
}

// NewReconcileRepo constructs a ReconcileRepo.
func NewReconcileRepo(db *sqlx.DB) *ReconcileRepo {
	return &ReconcileRepo{db: db}
}

// InsertReconciledTransaction saves a new matched (or unmatched) transaction.
// Fields dropship_id and shopee_id may be NULL if no match is found.
func (r *ReconcileRepo) InsertReconciledTransaction(
	ctx context.Context,
	rec *models.ReconciledTransaction,
) error {
	query := `
        INSERT INTO reconciled_transactions
          (shop_username, dropship_id, shopee_id, status, matched_at)
        VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.ExecContext(ctx, query,
		rec.ShopUsername, rec.DropshipID, rec.ShopeeID, rec.Status, rec.MatchedAt)
	return err
}

// GetReconciledTransactionsByShopAndPeriod fetches all reconciled rows (matched/unmatched)
// for a given shop in a year-month (YYYY-MM). Uses to_char on matched_at to filter.
func (r *ReconcileRepo) GetReconciledTransactionsByShopAndPeriod(
	ctx context.Context,
	shop, period string,
) ([]models.ReconciledTransaction, error) {
	var list []models.ReconciledTransaction
	err := r.db.SelectContext(ctx, &list,
		`SELECT * FROM reconciled_transactions
         WHERE shop_username = $1
           AND to_char(matched_at, 'YYYY-MM') = $2
         ORDER BY matched_at`,
		shop, period)
	return list, err
}

// ListUnmatched returns rows with status='unmatched'.
func (r *ReconcileRepo) ListUnmatched(ctx context.Context, shop string) ([]models.ReconciledTransaction, error) {
	var list []models.ReconciledTransaction
	err := r.db.SelectContext(ctx, &list,
		`SELECT * FROM reconciled_transactions WHERE shop_username=$1 AND status='unmatched'`, shop)
	return list, err
}
