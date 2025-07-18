package repository

import (
	"context"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// ShippingDiscrepancyRepo handles database operations for shipping_discrepancies.
type ShippingDiscrepancyRepo struct {
	db DBTX
}

// NewShippingDiscrepancyRepo constructs a ShippingDiscrepancyRepo.
func NewShippingDiscrepancyRepo(db DBTX) *ShippingDiscrepancyRepo {
	return &ShippingDiscrepancyRepo{db: db}
}

// InsertShippingDiscrepancy saves a shipping discrepancy record.
// It uses INSERT ON CONFLICT DO NOTHING to prevent duplicates.
func (r *ShippingDiscrepancyRepo) InsertShippingDiscrepancy(
	ctx context.Context,
	discrepancy *models.ShippingDiscrepancy,
) error {
	query := `
        INSERT INTO shipping_discrepancies
          (invoice_number, return_id, discrepancy_type, discrepancy_amount, 
           actual_shipping_fee, buyer_paid_shipping_fee, shopee_shipping_rebate, 
           seller_shipping_discount, reverse_shipping_fee, order_date, store_name)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        ON CONFLICT (invoice_number, discrepancy_type) DO NOTHING`
	_, err := r.db.ExecContext(ctx, query,
		discrepancy.InvoiceNumber, discrepancy.ReturnID, discrepancy.DiscrepancyType,
		discrepancy.DiscrepancyAmount, discrepancy.ActualShippingFee,
		discrepancy.BuyerPaidShippingFee, discrepancy.ShopeeShippingRebate,
		discrepancy.SellerShippingDiscount, discrepancy.ReverseShippingFee,
		discrepancy.OrderDate, discrepancy.StoreName)
	return err
}

// GetShippingDiscrepanciesByStore retrieves shipping discrepancies for a specific store.
func (r *ShippingDiscrepancyRepo) GetShippingDiscrepanciesByStore(
	ctx context.Context,
	storeName string,
	limit, offset int,
) ([]models.ShippingDiscrepancy, error) {
	var list []models.ShippingDiscrepancy
	query := `SELECT * FROM shipping_discrepancies 
              WHERE store_name = $1 
              ORDER BY created_at DESC 
              LIMIT $2 OFFSET $3`
	err := r.db.SelectContext(ctx, &list, query, storeName, limit, offset)
	if list == nil {
		list = []models.ShippingDiscrepancy{}
	}
	return list, err
}

// GetAllShippingDiscrepancies retrieves all shipping discrepancies with pagination.
func (r *ShippingDiscrepancyRepo) GetAllShippingDiscrepancies(
	ctx context.Context,
	limit, offset int,
) ([]models.ShippingDiscrepancy, error) {
	var list []models.ShippingDiscrepancy
	query := `SELECT * FROM shipping_discrepancies 
              ORDER BY created_at DESC 
              LIMIT $1 OFFSET $2`
	err := r.db.SelectContext(ctx, &list, query, limit, offset)
	if list == nil {
		list = []models.ShippingDiscrepancy{}
	}
	return list, err
}

// GetShippingDiscrepanciesByType retrieves shipping discrepancies by type.
func (r *ShippingDiscrepancyRepo) GetShippingDiscrepanciesByType(
	ctx context.Context,
	discrepancyType string,
	limit, offset int,
) ([]models.ShippingDiscrepancy, error) {
	var list []models.ShippingDiscrepancy
	query := `SELECT * FROM shipping_discrepancies 
              WHERE discrepancy_type = $1 
              ORDER BY created_at DESC 
              LIMIT $2 OFFSET $3`
	err := r.db.SelectContext(ctx, &list, query, discrepancyType, limit, offset)
	if list == nil {
		list = []models.ShippingDiscrepancy{}
	}
	return list, err
}

// CountShippingDiscrepanciesByDateRange counts shipping discrepancies in a date range.
func (r *ShippingDiscrepancyRepo) CountShippingDiscrepanciesByDateRange(
	ctx context.Context,
	startDate, endDate time.Time,
) (map[string]int, error) {
	result := make(map[string]int)
	query := `SELECT discrepancy_type, COUNT(*) as count 
              FROM shipping_discrepancies 
              WHERE created_at >= $1 AND created_at <= $2 
              GROUP BY discrepancy_type`

	rows, err := r.db.QueryContext(ctx, query, startDate, endDate)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		var discrepancyType string
		var count int
		if err := rows.Scan(&discrepancyType, &count); err != nil {
			return result, err
		}
		result[discrepancyType] = count
	}

	return result, rows.Err()
}

// GetShippingDiscrepancySumsByDateRange returns total amounts for each discrepancy type within a date range.
func (r *ShippingDiscrepancyRepo) GetShippingDiscrepancySumsByDateRange(
	ctx context.Context,
	startDate, endDate time.Time,
) (map[string]float64, error) {
	result := make(map[string]float64)
	query := `SELECT discrepancy_type, COALESCE(SUM(discrepancy_amount), 0) as total_amount
              FROM shipping_discrepancies 
              WHERE created_at >= $1 AND created_at <= $2 
              GROUP BY discrepancy_type`

	rows, err := r.db.QueryContext(ctx, query, startDate, endDate)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		var discrepancyType string
		var totalAmount float64
		if err := rows.Scan(&discrepancyType, &totalAmount); err != nil {
			return result, err
		}
		result[discrepancyType] = totalAmount
	}

	return result, rows.Err()
}

// GetShippingDiscrepancyByInvoice retrieves a shipping discrepancy by invoice number.
func (r *ShippingDiscrepancyRepo) GetShippingDiscrepancyByInvoice(
	ctx context.Context,
	invoiceNumber string,
) (*models.ShippingDiscrepancy, error) {
	var discrepancy models.ShippingDiscrepancy
	query := `SELECT * FROM shipping_discrepancies WHERE invoice_number = $1 LIMIT 1`
	err := r.db.GetContext(ctx, &discrepancy, query, invoiceNumber)
	if err != nil {
		return nil, err
	}
	return &discrepancy, nil
}
