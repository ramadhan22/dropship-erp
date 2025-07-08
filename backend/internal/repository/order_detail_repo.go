package repository

import (
	"context"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// OrderDetailRepo manages shopee_order_details and shopee_order_items tables.
type OrderDetailRepo struct{ db DBTX }

func NewOrderDetailRepo(db DBTX) *OrderDetailRepo { return &OrderDetailRepo{db: db} }

// SaveOrderDetail replaces any existing rows for the order_sn then inserts detail and items.
func (r *OrderDetailRepo) SaveOrderDetail(ctx context.Context, detail *models.ShopeeOrderDetailRow, items []models.ShopeeOrderItemRow) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM shopee_order_items WHERE order_sn=$1`, detail.OrderSN); err != nil {
		return err
	}
	if _, err := r.db.ExecContext(ctx, `DELETE FROM shopee_order_details WHERE order_sn=$1`, detail.OrderSN); err != nil {
		return err
	}
	if detail.CreatedAt.IsZero() {
		detail.CreatedAt = time.Now()
	}
	if _, err := r.db.NamedExecContext(ctx, `INSERT INTO shopee_order_details (order_sn, nama_toko, detail, created_at)
        VALUES (:order_sn,:nama_toko,:detail,:created_at)`, detail); err != nil {
		return err
	}
	for i := range items {
		if _, err := r.db.NamedExecContext(ctx, `INSERT INTO shopee_order_items (order_sn, item) VALUES (:order_sn,:item)`, &items[i]); err != nil {
			return err
		}
	}
	return nil
}
