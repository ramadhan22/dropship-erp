package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// BatchRepo handles database operations for batch_history table.
type BatchRepo struct{ db DBTX }

func NewBatchRepo(db DBTX) *BatchRepo { return &BatchRepo{db: db} }

func (r *BatchRepo) Insert(ctx context.Context, b *models.BatchHistory) (int64, error) {
	query := `INSERT INTO batch_history (process_type, started_at, total_data, done_data, status, error_message)
              VALUES (:process_type, NOW(), :total_data, :done_data, :status, :error_message)
              RETURNING id`
	rows, err := sqlx.NamedQueryContext(ctx, r.db, query, b)
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
	return 0, nil
}

func (r *BatchRepo) UpdateDone(ctx context.Context, id int64, done int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE batch_history SET done_data=$2 WHERE id=$1`, id, done)
	return err
}

// UpdateTotal sets the total_data count for a batch record.
func (r *BatchRepo) UpdateTotal(ctx context.Context, id int64, total int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE batch_history SET total_data=$2 WHERE id=$1`, id, total)
	return err
}

func (r *BatchRepo) UpdateStatus(ctx context.Context, id int64, status, msg string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE batch_history SET status=$2, error_message=$3 WHERE id=$1`,
		id, status, msg)
	return err
}

func (r *BatchRepo) List(ctx context.Context) ([]models.BatchHistory, error) {
	var list []models.BatchHistory
	err := r.db.SelectContext(ctx, &list, `SELECT * FROM batch_history ORDER BY started_at DESC`)
	if list == nil {
		list = []models.BatchHistory{}
	}
	return list, err
}
