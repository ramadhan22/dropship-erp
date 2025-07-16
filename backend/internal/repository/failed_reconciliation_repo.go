package repository

import (
	"context"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// FailedReconciliationRepo handles database operations for failed_reconciliations.
type FailedReconciliationRepo struct {
	db DBTX
}

// NewFailedReconciliationRepo constructs a FailedReconciliationRepo.
func NewFailedReconciliationRepo(db DBTX) *FailedReconciliationRepo {
	return &FailedReconciliationRepo{db: db}
}

// InsertFailedReconciliation saves a failed reconciliation record.
func (r *FailedReconciliationRepo) InsertFailedReconciliation(
	ctx context.Context,
	failed *models.FailedReconciliation,
) error {
	query := `
        INSERT INTO failed_reconciliations
          (purchase_id, order_id, shop, error_type, error_message, context, failed_at, batch_id)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.ExecContext(ctx, query,
		failed.PurchaseID, failed.OrderID, failed.Shop, failed.ErrorType,
		failed.ErrorMsg, failed.Context, failed.FailedAt, failed.BatchID)
	return err
}

// GetFailedReconciliationsByShop retrieves failed reconciliations for a specific shop.
func (r *FailedReconciliationRepo) GetFailedReconciliationsByShop(
	ctx context.Context,
	shop string,
	limit, offset int,
) ([]models.FailedReconciliation, error) {
	var list []models.FailedReconciliation
	query := `SELECT * FROM failed_reconciliations 
              WHERE shop = $1 
              ORDER BY failed_at DESC 
              LIMIT $2 OFFSET $3`
	err := r.db.SelectContext(ctx, &list, query, shop, limit, offset)
	if list == nil {
		list = []models.FailedReconciliation{}
	}
	return list, err
}

// GetFailedReconciliationsByBatch retrieves failed reconciliations for a specific batch.
func (r *FailedReconciliationRepo) GetFailedReconciliationsByBatch(
	ctx context.Context,
	batchID int64,
) ([]models.FailedReconciliation, error) {
	var list []models.FailedReconciliation
	query := `SELECT * FROM failed_reconciliations 
              WHERE batch_id = $1 
              ORDER BY failed_at DESC`
	err := r.db.SelectContext(ctx, &list, query, batchID)
	if list == nil {
		list = []models.FailedReconciliation{}
	}
	return list, err
}

// CountFailedReconciliationsByErrorType counts failures grouped by error type.
func (r *FailedReconciliationRepo) CountFailedReconciliationsByErrorType(
	ctx context.Context,
	shop string,
	since time.Time,
) (map[string]int, error) {
	query := `SELECT error_type, COUNT(*) as count 
              FROM failed_reconciliations 
              WHERE shop = $1 AND failed_at >= $2 
              GROUP BY error_type`
	
	rows, err := r.db.QueryContext(ctx, query, shop, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var errorType string
		var count int
		if err := rows.Scan(&errorType, &count); err != nil {
			return nil, err
		}
		result[errorType] = count
	}
	return result, rows.Err()
}

// MarkAsRetried updates a failed reconciliation as retried.
func (r *FailedReconciliationRepo) MarkAsRetried(
	ctx context.Context,
	id int64,
) error {
	query := `UPDATE failed_reconciliations 
              SET retried = TRUE, retried_at = NOW() 
              WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// GetUnretriedFailedReconciliations gets failed reconciliations that haven't been retried.
func (r *FailedReconciliationRepo) GetUnretriedFailedReconciliations(
	ctx context.Context,
	shop string,
	limit int,
) ([]models.FailedReconciliation, error) {
	var list []models.FailedReconciliation
	query := `SELECT * FROM failed_reconciliations 
              WHERE shop = $1 AND retried = FALSE 
              ORDER BY failed_at ASC 
              LIMIT $1`
	err := r.db.SelectContext(ctx, &list, query, shop, limit)
	if list == nil {
		list = []models.FailedReconciliation{}
	}
	return list, err
}

// DeleteFailedReconciliation removes a failed reconciliation record (useful for cleanup).
func (r *FailedReconciliationRepo) DeleteFailedReconciliation(
	ctx context.Context,
	id int64,
) error {
	query := `DELETE FROM failed_reconciliations WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}