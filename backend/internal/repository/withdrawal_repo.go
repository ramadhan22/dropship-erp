package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// WithdrawalRepo handles CRUD for withdrawals table.

type WithdrawalRepo struct{ db DBTX }

func NewWithdrawalRepo(db DBTX) *WithdrawalRepo { return &WithdrawalRepo{db: db} }

// GetByStoreDate fetches a withdrawal by store and date if it exists.
func (r *WithdrawalRepo) GetByStoreDate(ctx context.Context, store string, date time.Time) (*models.Withdrawal, error) {
	var w models.Withdrawal
	if err := r.db.GetContext(ctx, &w, `SELECT * FROM withdrawals WHERE store=$1 AND date=$2 LIMIT 1`, store, date); err != nil {
		return nil, err
	}
	return &w, nil
}

// Delete removes a withdrawal row by id.
func (r *WithdrawalRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM withdrawals WHERE id=$1`, id)
	return err
}

func (r *WithdrawalRepo) Insert(ctx context.Context, w *models.Withdrawal) error {
	q := `INSERT INTO withdrawals (store, date, amount, created_at)
          VALUES (:store,:date,:amount,:created_at) RETURNING id`
	stmt, args, err := r.db.BindNamed(q, w)
	if err != nil {
		return err
	}
	return r.db.QueryRowxContext(ctx, stmt, args...).Scan(&w.ID)
}

func (r *WithdrawalRepo) List(ctx context.Context, sortBy, dir string) ([]models.Withdrawal, error) {
	if sortBy == "" {
		sortBy = "date"
	}
	if dir != "asc" && dir != "desc" {
		dir = "desc"
	}
	query := fmt.Sprintf(`SELECT * FROM withdrawals ORDER BY %s %s`, sortBy, dir)
	var res []models.Withdrawal
	err := r.db.SelectContext(ctx, &res, query)
	if res == nil {
		res = []models.Withdrawal{}
	}
	return res, err
}
