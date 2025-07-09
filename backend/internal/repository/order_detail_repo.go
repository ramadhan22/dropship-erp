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
func (r *OrderDetailRepo) SaveOrderDetail(ctx context.Context, detail *models.ShopeeOrderDetailRow, items []models.ShopeeOrderItemRow, packages []models.ShopeeOrderPackageRow) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM shopee_order_packages WHERE order_sn=$1`, detail.OrderSN); err != nil {
		return err
	}
	if _, err := r.db.ExecContext(ctx, `DELETE FROM shopee_order_items WHERE order_sn=$1`, detail.OrderSN); err != nil {
		return err
	}
	if _, err := r.db.ExecContext(ctx, `DELETE FROM shopee_order_details WHERE order_sn=$1`, detail.OrderSN); err != nil {
		return err
	}
	if detail.CreatedAt.IsZero() {
		detail.CreatedAt = time.Now()
	}
	if _, err := r.db.NamedExecContext(ctx, `INSERT INTO shopee_order_details (
                order_sn, nama_toko, status, checkout_time, update_time, pay_time,
                total_amount, currency, actual_shipping_fee_confirmed, buyer_cancel_reason,
                buyer_cpf_id, buyer_user_id, buyer_username, cancel_by, cancel_reason, cod,
                create_time, days_to_ship, dropshipper, dropshipper_phone, estimated_shipping_fee,
                fulfillment_flag, goods_to_declare, message_to_seller, note, note_update_time,
                order_status, pickup_done_time, region, reverse_shipping_fee, ship_by_date,
                shipping_carrier, split_up, payment_method, recipient_name, recipient_phone,
                recipient_full_address, recipient_city, recipient_district, recipient_state,
                recipient_town, recipient_zipcode, created_at
        ) VALUES (
                :order_sn,:nama_toko,:status,:checkout_time,:update_time,:pay_time,
                :total_amount,:currency,:actual_shipping_fee_confirmed,:buyer_cancel_reason,
                :buyer_cpf_id,:buyer_user_id,:buyer_username,:cancel_by,:cancel_reason,:cod,
                :create_time,:days_to_ship,:dropshipper,:dropshipper_phone,:estimated_shipping_fee,
                :fulfillment_flag,:goods_to_declare,:message_to_seller,:note,:note_update_time,
                :order_status,:pickup_done_time,:region,:reverse_shipping_fee,:ship_by_date,
                :shipping_carrier,:split_up,:payment_method,:recipient_name,:recipient_phone,
                :recipient_full_address,:recipient_city,:recipient_district,:recipient_state,
                :recipient_town,:recipient_zipcode,:created_at
        )`, detail); err != nil {
		return err
	}
	for i := range items {
		if _, err := r.db.NamedExecContext(ctx, `INSERT INTO shopee_order_items (
                        order_sn, order_item_id, item_name, model_original_price, model_quantity_purchased,
                        item_id, item_sku, model_id, model_name, model_sku, model_discounted_price, weight,
                        promotion_id, promotion_type, promotion_group_id, add_on_deal, add_on_deal_id,
                        main_item, is_b2c_owned_item, is_prescription_item, wholesale, product_location_id,
                        image_url
                ) VALUES (
                        :order_sn,:order_item_id,:item_name,:model_original_price,:model_quantity_purchased,
                        :item_id,:item_sku,:model_id,:model_name,:model_sku,:model_discounted_price,:weight,
                        :promotion_id,:promotion_type,:promotion_group_id,:add_on_deal,:add_on_deal_id,
                        :main_item,:is_b2c_owned_item,:is_prescription_item,:wholesale,:product_location_id,
                        :image_url
                )`, &items[i]); err != nil {
			return err
		}
	}
	for i := range packages {
		if _, err := r.db.NamedExecContext(ctx, `INSERT INTO shopee_order_packages (
                        order_sn, package_number, logistics_status, shipping_carrier,
                        logistics_channel_id, parcel_chargeable_weight_gram, allow_self_design_awb,
                        sorting_group, group_shipment_id
                ) VALUES (
                        :order_sn,:package_number,:logistics_status,:shipping_carrier,
                        :logistics_channel_id,:parcel_chargeable_weight_gram,:allow_self_design_awb,
                        :sorting_group,:group_shipment_id
                )`, &packages[i]); err != nil {
			return err
		}
	}
	return nil
}

// ListOrderDetails returns rows filtered by store name and partial order_sn with pagination.
func (r *OrderDetailRepo) ListOrderDetails(ctx context.Context, store, order string, limit, offset int) ([]models.ShopeeOrderDetailRow, int, error) {
	countQuery := `SELECT COUNT(*) FROM shopee_order_details
                WHERE ($1 = '' OR nama_toko = $1)
                  AND ($2 = '' OR order_sn ILIKE '%' || $2 || '%')`
	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, store, order); err != nil {
		return nil, 0, err
	}
	query := `SELECT * FROM shopee_order_details
                WHERE ($1 = '' OR nama_toko = $1)
                  AND ($2 = '' OR order_sn ILIKE '%' || $2 || '%')
                ORDER BY checkout_time DESC
                LIMIT $3 OFFSET $4`
	var list []models.ShopeeOrderDetailRow
	if err := r.db.SelectContext(ctx, &list, query, store, order, limit, offset); err != nil {
		return nil, 0, err
	}
	if list == nil {
		list = []models.ShopeeOrderDetailRow{}
	}
	return list, total, nil
}

// GetOrderDetail fetches a detail row and associated items and packages by order_sn.
func (r *OrderDetailRepo) GetOrderDetail(ctx context.Context, sn string) (*models.ShopeeOrderDetailRow, []models.ShopeeOrderItemRow, []models.ShopeeOrderPackageRow, error) {
	var det models.ShopeeOrderDetailRow
	if err := r.db.GetContext(ctx, &det, `SELECT * FROM shopee_order_details WHERE order_sn=$1`, sn); err != nil {
		return nil, nil, nil, err
	}
	var items []models.ShopeeOrderItemRow
	if err := r.db.SelectContext(ctx, &items, `SELECT * FROM shopee_order_items WHERE order_sn=$1 ORDER BY id`, sn); err != nil {
		return nil, nil, nil, err
	}
	if items == nil {
		items = []models.ShopeeOrderItemRow{}
	}
	var packs []models.ShopeeOrderPackageRow
	if err := r.db.SelectContext(ctx, &packs, `SELECT * FROM shopee_order_packages WHERE order_sn=$1 ORDER BY id`, sn); err != nil {
		return nil, nil, nil, err
	}
	if packs == nil {
		packs = []models.ShopeeOrderPackageRow{}
	}
	return &det, items, packs, nil
}

// UpdateOrderDetailStatus updates status fields and update_time for the given order_sn.
func (r *OrderDetailRepo) UpdateOrderDetailStatus(ctx context.Context, sn, status, orderStatus string, updateTime time.Time) error {
	var statusVal, orderStatusVal interface{}
	if status != "" {
		statusVal = status
	}
	if orderStatus != "" {
		orderStatusVal = orderStatus
	}
	_, err := r.db.ExecContext(ctx,
		`UPDATE shopee_order_details
                 SET status=$2, order_status=$3, update_time=$4
                 WHERE order_sn=$1`,
		sn, statusVal, orderStatusVal, updateTime)
	return err
}
