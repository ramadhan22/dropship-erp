package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// ShopeeRepo handles interactions with the shopee_settled_orders table.
type ShopeeRepo struct {
	db *sqlx.DB
}

// ListShopeeOrdersByShopAndDate implements service.MetricServiceShopeeRepo.
func (r *ShopeeRepo) ListShopeeOrdersByShopAndDate(ctx context.Context, shop string, from string, to string) ([]models.ShopeeSettledOrder, error) {
	panic("unimplemented")
}

// NewShopeeRepo constructs a ShopeeRepo given an *sqlx.DB.
func NewShopeeRepo(db *sqlx.DB) *ShopeeRepo {
	return &ShopeeRepo{db: db}
}

// InsertShopeeOrder inserts a ShopeeSettledOrder into the database.
// Similar pattern as InsertDropshipPurchase: it uses NamedExecContext to map struct fields to columns.
func (r *ShopeeRepo) InsertShopeeOrder(ctx context.Context, o *models.ShopeeSettledOrder) error {
	query := `
        INSERT INTO shopee_settled_orders (
            order_id, net_income, service_fee, campaign_fee,
            credit_card_fee, shipping_subsidy, tax_and_import_fee,
            settled_date, seller_username
        ) VALUES (
            :order_id, :net_income, :service_fee, :campaign_fee,
            :credit_card_fee, :shipping_subsidy, :tax_and_import_fee,
            :settled_date, :seller_username
        )`
	_, err := r.db.NamedExecContext(ctx, query, o)
	return err
}

// GetShopeeOrderByID retrieves one settled order by its unique order_id.
// This is used when reconciling with dropship purchases or calculating revenue.
func (r *ShopeeRepo) GetShopeeOrderByID(ctx context.Context, orderID string) (*models.ShopeeSettledOrder, error) {
	var o models.ShopeeSettledOrder
	err := r.db.GetContext(ctx, &o,
		`SELECT * FROM shopee_settled_orders WHERE order_id = $1`, orderID)
	if err != nil {
		return nil, err
	}
	return &o, nil
}
