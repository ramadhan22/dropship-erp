package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// BatchDetailRepo handles operations for batch_history_details table.
type BatchDetailRepo struct{ db DBTX }

func NewBatchDetailRepo(db DBTX) *BatchDetailRepo { return &BatchDetailRepo{db: db} }

func (r *BatchDetailRepo) Insert(ctx context.Context, d *models.BatchHistoryDetail) error {
	query := `INSERT INTO batch_history_details (batch_id, reference, store, status, error_message)
              VALUES (:batch_id, :reference, :store, :status, :error_message)`
	_, err := sqlx.NamedExecContext(ctx, r.db, query, d)
	return err
}

func (r *BatchDetailRepo) ListByBatchID(ctx context.Context, id int64) ([]models.BatchHistoryDetail, error) {
	var list []models.BatchHistoryDetail
	err := r.db.SelectContext(ctx, &list,
		`SELECT * FROM batch_history_details WHERE batch_id=$1 ORDER BY id`, id)
	if list == nil {
		list = []models.BatchHistoryDetail{}
	}
	return list, err
}

func (r *BatchDetailRepo) UpdateStatus(ctx context.Context, id int64, status, msg string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE batch_history_details SET status=$2, error_message=$3 WHERE id=$1`,
		id, status, msg)
	return err
}
