package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// DropshipRepo handles all database operations related to the dropship_purchases table.
type DropshipRepo struct {
	db *sqlx.DB
}

// UpdateDropshipPurchase implements service.DropshipRepoInterface2.
func (r *DropshipRepo) UpdateDropshipPurchase(ctx context.Context, p *models.DropshipPurchase) error {
	panic("unimplemented")
}

// NewDropshipRepo constructs a DropshipRepo given an *sqlx.DB connection.
func NewDropshipRepo(db *sqlx.DB) *DropshipRepo {
	return &DropshipRepo{db: db}
}

// InsertDropshipPurchase receives a *models.DropshipPurchase and executes an INSERT into dropship_purchases.
// It uses NamedExecContext so the struct fields map to column names automatically (via db tags).
func (r *DropshipRepo) InsertDropshipPurchase(ctx context.Context, p *models.DropshipPurchase) error {
	query := `
        INSERT INTO dropship_purchases (
            seller_username, purchase_id, order_id, sku, qty,
            purchase_price, purchase_fee, status, purchase_date, supplier_name
        ) VALUES (
            :seller_username, :purchase_id, :order_id, :sku, :qty,
            :purchase_price, :purchase_fee, :status, :purchase_date, :supplier_name
        )`
	_, err := r.db.NamedExecContext(ctx, query, p)
	return err
}

// GetDropshipPurchaseByID looks up a single row by purchase_id (the unique identifier in that table).
// It fills a models.DropshipPurchase struct with all columns from that row.
func (r *DropshipRepo) GetDropshipPurchaseByID(ctx context.Context, purchaseID string) (*models.DropshipPurchase, error) {
	var p models.DropshipPurchase
	err := r.db.GetContext(ctx, &p,
		`SELECT * FROM dropship_purchases WHERE purchase_id = $1`, purchaseID)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// ListDropshipPurchasesByShopAndDate returns all dropship purchases for a given shop_username
// whose purchase_date falls between two string‐formatted dates (YYYY-MM-DD).
// This lets you pull a slice of purchases to, for example, generate reports or feed reconciliation logic.
func (r *DropshipRepo) ListDropshipPurchasesByShopAndDate(
	ctx context.Context,
	shop string,
	from, to string, // expects "2025-05-01" format
) ([]models.DropshipPurchase, error) {
	var list []models.DropshipPurchase
	err := r.db.SelectContext(ctx, &list,
		`SELECT * FROM dropship_purchases
         WHERE seller_username = $1
           AND purchase_date BETWEEN $2 AND $3
         ORDER BY purchase_date`,
		shop, from, to)
	return list, err
}
