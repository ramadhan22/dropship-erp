package repository

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/ramadhan22/dropship-erp/backend/internal/logutil"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

type ExpenseRepo struct{ db DBTX }

func NewExpenseRepo(db DBTX) *ExpenseRepo { return &ExpenseRepo{db: db} }

func (r *ExpenseRepo) Create(ctx context.Context, e *models.Expense) error {
	log.Printf("ExpenseRepo.Create %s", e.ID)
	// Ensure an ID exists so that expense_lines can reference it
	if e.ID == "" {
		e.ID = uuid.NewString()
	}
	query := `INSERT INTO expenses (id, date, description, amount, asset_account_id)
       VALUES (:id,:date,:description,:amount,:asset_account_id) RETURNING id`
	q, args, err := sqlx.Named(query, e)
	if err != nil {
		return err
	}
	q = r.db.Rebind(q)
	if err := r.db.QueryRowxContext(ctx, q, args...).Scan(&e.ID); err != nil {
		logutil.Errorf("ExpenseRepo.Create error: %v", err)
		return err
	}
	for i := range e.Lines {
		l := e.Lines[i]
		l.ExpenseID = e.ID
		if _, err := r.db.NamedExecContext(ctx,
			`INSERT INTO expense_lines (expense_id, account_id, amount) VALUES (:expense_id,:account_id,:amount)`,
			l); err != nil {
			return err
		}
	}
	log.Printf("ExpenseRepo.Create done %s", e.ID)
	return nil
}

func (r *ExpenseRepo) GetByID(ctx context.Context, id string) (*models.Expense, error) {
	var ex models.Expense
	err := r.db.GetContext(ctx, &ex, `SELECT * FROM expenses WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	var lines []models.ExpenseLine
	if err := r.db.SelectContext(ctx, &lines, `SELECT * FROM expense_lines WHERE expense_id=$1 ORDER BY line_id`, id); err != nil {
		return nil, err
	}
	ex.Lines = lines
	return &ex, nil
}

func (r *ExpenseRepo) List(ctx context.Context, accountID int64, sortBy, dir string, limit, offset int) ([]models.Expense, int, error) {
	base := `SELECT * FROM expenses`
	args := []interface{}{}
	conds := []string{}
	arg := 1
	if accountID != 0 {
		conds = append(conds, fmt.Sprintf(`(asset_account_id = $%d OR id IN (SELECT expense_id FROM expense_lines WHERE account_id=$%d))`, arg, arg))
		args = append(args, accountID)
		arg++
	}
	query := base
	if len(conds) > 0 {
		query += " WHERE " + strings.Join(conds, " AND ")
	}
	countQuery := "SELECT COUNT(*) FROM (" + query + ") AS sub"
	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}
	sortCol := map[string]string{"date": "date", "amount": "amount"}[sortBy]
	if sortCol == "" {
		sortCol = "date"
	}
	direction := "ASC"
	if strings.ToLower(dir) == "desc" {
		direction = "DESC"
	}
	args = append(args, limit, offset)
	query += fmt.Sprintf(" ORDER BY %s %s LIMIT $%d OFFSET $%d", sortCol, direction, arg, arg+1)
	var list []models.Expense
	if err := r.db.SelectContext(ctx, &list, query, args...); err != nil {
		return nil, 0, err
	}
	if list == nil {
		list = []models.Expense{}
	}
	if len(list) == 0 {
		return list, total, nil
	}
	ids := make([]interface{}, len(list))
	for i, e := range list {
		ids[i] = e.ID
	}
	queryLines, args, err := sqlx.In(`SELECT * FROM expense_lines WHERE expense_id IN (?) ORDER BY line_id`, ids)
	if err != nil {
		return nil, 0, err
	}
	queryLines = r.db.Rebind(queryLines)
	var lines []models.ExpenseLine
	if err := r.db.SelectContext(ctx, &lines, queryLines, args...); err != nil {
		return nil, 0, err
	}
	lineMap := map[string][]models.ExpenseLine{}
	for _, l := range lines {
		lineMap[l.ExpenseID] = append(lineMap[l.ExpenseID], l)
	}
	for i := range list {
		list[i].Lines = lineMap[list[i].ID]
	}
	return list, total, nil
}

func (r *ExpenseRepo) Update(ctx context.Context, e *models.Expense) error {
	log.Printf("ExpenseRepo.Update %s", e.ID)
	_, err := r.db.NamedExecContext(ctx,
		`UPDATE expenses SET date=:date, description=:description, amount=:amount, asset_account_id=:asset_account_id WHERE id=:id`, e)
	if err != nil {
		logutil.Errorf("ExpenseRepo.Update error: %v", err)
		return err
	}
	_, err = r.db.ExecContext(ctx, `DELETE FROM expense_lines WHERE expense_id=$1`, e.ID)
	if err != nil {
		logutil.Errorf("ExpenseRepo.Update delete lines error: %v", err)
		return err
	}
	for i := range e.Lines {
		l := e.Lines[i]
		l.ExpenseID = e.ID
		if _, err := r.db.NamedExecContext(ctx,
			`INSERT INTO expense_lines (expense_id, account_id, amount) VALUES (:expense_id,:account_id,:amount)`, l); err != nil {
			logutil.Errorf("ExpenseRepo.Update insert line error: %v", err)
			return err
		}
	}
	log.Printf("ExpenseRepo.Update done %s", e.ID)
	return nil
}

func (r *ExpenseRepo) Delete(ctx context.Context, id string) error {
	log.Printf("ExpenseRepo.Delete %s", id)
	_, err := r.db.ExecContext(ctx, `DELETE FROM expenses WHERE id=$1`, id)
	if err != nil {
		logutil.Errorf("ExpenseRepo.Delete error: %v", err)
	}
	return err
}
