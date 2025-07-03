package repository

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/ramadhan22/dropship-erp/backend/internal/logutil"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// AccountRepo handles CRUD operations for accounts table.
type AccountRepo struct {
	db *sqlx.DB
}

// NewAccountRepo constructs an AccountRepo.
func NewAccountRepo(db *sqlx.DB) *AccountRepo {
	return &AccountRepo{db: db}
}

// CreateAccount inserts a new account row and returns its ID.
func (r *AccountRepo) CreateAccount(ctx context.Context, a *models.Account) (int64, error) {
	log.Printf("AccountRepo.CreateAccount %s", a.AccountCode)
	var id int64
	err := r.db.QueryRowxContext(ctx,
		`INSERT INTO accounts (account_code, account_name, account_type, parent_id)
         VALUES ($1, $2, $3, $4) RETURNING account_id`,
		a.AccountCode, a.AccountName, a.AccountType, a.ParentID,
	).Scan(&id)
	if err != nil {
		logutil.Errorf("AccountRepo.CreateAccount error: %v", err)
		return 0, err
	}
	log.Printf("AccountRepo.CreateAccount id=%d", id)
	return id, nil
}

// GetAccountByID fetches a single account by its ID.
func (r *AccountRepo) GetAccountByID(ctx context.Context, id int64) (*models.Account, error) {
	var a models.Account
	if err := r.db.GetContext(ctx, &a, `SELECT * FROM accounts WHERE account_id=$1`, id); err != nil {
		return nil, err
	}
	return &a, nil
}

// ListAccounts returns all accounts ordered by code.
func (r *AccountRepo) ListAccounts(ctx context.Context) ([]models.Account, error) {
	var list []models.Account
	err := r.db.SelectContext(ctx, &list, `SELECT * FROM accounts ORDER BY account_code`)
	if list == nil {
		list = []models.Account{}
	}
	return list, err
}

// UpdateAccount updates an existing account by ID.
func (r *AccountRepo) UpdateAccount(ctx context.Context, a *models.Account) error {
	log.Printf("AccountRepo.UpdateAccount %d", a.AccountID)
	_, err := r.db.ExecContext(ctx,
		`UPDATE accounts
         SET account_code=$1, account_name=$2, account_type=$3, parent_id=$4, updated_at=NOW()
         WHERE account_id=$5`,
		a.AccountCode, a.AccountName, a.AccountType, a.ParentID, a.AccountID,
	)
	if err != nil {
		logutil.Errorf("AccountRepo.UpdateAccount error: %v", err)
	}
	return err
}

// DeleteAccount removes an account row.
func (r *AccountRepo) DeleteAccount(ctx context.Context, id int64) error {
	log.Printf("AccountRepo.DeleteAccount %d", id)
	_, err := r.db.ExecContext(ctx, `DELETE FROM accounts WHERE account_id=$1`, id)
	if err != nil {
		logutil.Errorf("AccountRepo.DeleteAccount error: %v", err)
	}
	return err
}
