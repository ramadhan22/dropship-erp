package repository

import (
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

type ExpenseRepo struct{ db *sqlx.DB }

func NewExpenseRepo(db *sqlx.DB) *ExpenseRepo { return &ExpenseRepo{db: db} }

func (r *ExpenseRepo) Create(ctx context.Context, e *models.Expense) error {
	_, err := r.db.NamedExecContext(ctx,
		`INSERT INTO expenses (id, date, description, amount, account_id) VALUES (:id,:date,:description,:amount,:account_id)`, e)
	return err
}

func (r *ExpenseRepo) GetByID(ctx context.Context, id string) (*models.Expense, error) {
	var ex models.Expense
	err := r.db.GetContext(ctx, &ex, `SELECT * FROM expenses WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	return &ex, nil
}

func (r *ExpenseRepo) List(ctx context.Context) ([]models.Expense, error) {
	var list []models.Expense
	err := r.db.SelectContext(ctx, &list, `SELECT * FROM expenses ORDER BY date DESC`)
	if list == nil {
		list = []models.Expense{}
	}
	return list, err
}

func (r *ExpenseRepo) Update(ctx context.Context, e *models.Expense) error {
	_, err := r.db.NamedExecContext(ctx,
		`UPDATE expenses SET date=:date, description=:description, amount=:amount, account_id=:account_id WHERE id=:id`, e)
	return err
}

func (r *ExpenseRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM expenses WHERE id=$1`, id)
	return err
}
