package repository

import (
	"context"
	"fmt"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// WithdrawalRepo handles CRUD for withdrawals table.

type WithdrawalRepo struct{ db DBTX }

func NewWithdrawalRepo(db DBTX) *WithdrawalRepo { return &WithdrawalRepo{db: db} }

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
