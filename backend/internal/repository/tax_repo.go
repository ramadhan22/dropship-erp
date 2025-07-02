package repository

import (
	"context"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// TaxRepo manages tax_payments table.
type TaxRepo struct{ db DBTX }

func NewTaxRepo(db DBTX) *TaxRepo { return &TaxRepo{db: db} }

func (r *TaxRepo) Get(ctx context.Context, store, periodType, periodValue string) (*models.TaxPayment, error) {
	var tp models.TaxPayment
	err := r.db.GetContext(ctx, &tp,
		`SELECT * FROM tax_payments WHERE store=$1 AND period_type=$2 AND period_value=$3 LIMIT 1`,
		store, periodType, periodValue)
	if err != nil {
		return nil, err
	}
	return &tp, nil
}

func (r *TaxRepo) Create(ctx context.Context, tp *models.TaxPayment) error {
	query := `INSERT INTO tax_payments (store, period_type, period_value, revenue, tax_rate, tax_amount, is_paid, paid_at)
              VALUES (:store,:period_type,:period_value,:revenue,:tax_rate,:tax_amount,:is_paid,:paid_at)
              RETURNING id`
	stmt, args, err := r.db.BindNamed(query, tp)
	if err != nil {
		return err
	}
	return r.db.QueryRowxContext(ctx, stmt, args...).Scan(&tp.ID)
}

func (r *TaxRepo) MarkPaid(ctx context.Context, id string, paidAt time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE tax_payments SET is_paid=TRUE, paid_at=$2 WHERE id=$1`, id, paidAt)
	return err
}
