package repository

import (
	"context"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// AssetAccountRepo handles CRUD operations for asset_accounts table.
type AssetAccountRepo struct{ db DBTX }

func NewAssetAccountRepo(db DBTX) *AssetAccountRepo { return &AssetAccountRepo{db: db} }

func (r *AssetAccountRepo) Create(ctx context.Context, a *models.AssetAccount) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO asset_accounts (account_id) VALUES ($1)`,
		a.AccountID,
	)
	return err
}

func (r *AssetAccountRepo) GetByID(ctx context.Context, id int64) (*models.AssetAccount, error) {
	var aa models.AssetAccount
	err := r.db.GetContext(ctx, &aa, `SELECT * FROM asset_accounts WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	return &aa, nil
}

func (r *AssetAccountRepo) List(ctx context.Context) ([]models.AssetAccount, error) {
	var list []models.AssetAccount
	err := r.db.SelectContext(ctx, &list, `SELECT * FROM asset_accounts ORDER BY id`)
	if list == nil {
		list = []models.AssetAccount{}
	}
	return list, err
}

func (r *AssetAccountRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM asset_accounts WHERE id=$1`, id)
	return err
}
