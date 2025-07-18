package repository

import (
	"context"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// ReconcileRepo handles database operations for reconciled_transactions.
type ReconcileRepo struct {
	db DBTX
}

// NewReconcileRepo constructs a ReconcileRepo.
func NewReconcileRepo(db DBTX) *ReconcileRepo {
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
	if list == nil {
		list = []models.ReconciledTransaction{}
	}
	return list, err
}

// ListUnmatched returns rows with status='unmatched'.
func (r *ReconcileRepo) ListUnmatched(ctx context.Context, shop string) ([]models.ReconciledTransaction, error) {
	var list []models.ReconciledTransaction
	err := r.db.SelectContext(ctx, &list,
		`SELECT * FROM reconciled_transactions WHERE shop_username=$1 AND status='unmatched'`, shop)
	if list == nil {
		list = []models.ReconciledTransaction{}
	}
	return list, err
}

// ListCandidates returns dropship purchases that either have no matching row in
// shopee_settled or have a matching row but the purchase status is not
// "pesanan selesai". Optional shop filter matches nama_toko.
func (r *ReconcileRepo) ListCandidates(ctx context.Context, shop, order, status, from, to string, limit, offset int) ([]models.ReconcileCandidate, int, error) {
	base := `SELECT dp.kode_pesanan, dp.kode_invoice_channel, dp.nama_toko, dp.status_pesanan_terakhir,
               ss.no_pesanan
               FROM dropship_purchases dp
               JOIN journal_entries je ON je.source_type = 'pending_sales' AND je.source_id = dp.kode_invoice_channel
               LEFT JOIN shopee_settled ss ON dp.kode_invoice_channel = ss.no_pesanan
               WHERE ($1 = '' OR dp.nama_toko = $1)
                 AND ($2 = '' OR dp.kode_invoice_channel ILIKE '%' || $2 || '%')
                 AND ($3 = '' OR dp.status_pesanan_terakhir ILIKE '%' || $3 || '%')
                 AND ($4 = '' OR DATE(dp.waktu_pesanan_terbuat) >= $4::date)
                 AND ($5 = '' OR DATE(dp.waktu_pesanan_terbuat) <= $5::date)
                 AND NOT EXISTS (
                       SELECT 1 FROM journal_entries je2
                       WHERE je2.source_id = dp.kode_invoice_channel
                         AND je2.source_type IN ('shopee_escrow','reconcile_cancel')
               )`

	countQuery := "SELECT COUNT(*) FROM (" + base + ") AS sub"
	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, shop, order, status, from, to); err != nil {
		return nil, 0, err
	}

	base += " ORDER BY dp.waktu_pesanan_terbuat DESC LIMIT $6 OFFSET $7"
	var list []models.ReconcileCandidate
	err := r.db.SelectContext(ctx, &list, base, shop, order, status, from, to, limit, offset)
	if list == nil {
		list = []models.ReconcileCandidate{}
	}
	return list, total, err
}
