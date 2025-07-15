package repository

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/ramadhan22/dropship-erp/backend/internal/logutil"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// DropshipRepo handles all database operations related to the dropship_purchases table.
type DropshipRepo struct {
	db DBTX
}

// DailyPurchaseTotal represents aggregated purchase totals per day.
type DailyPurchaseTotal struct {
	Date  string  `db:"date" json:"date"`
	Total float64 `db:"total" json:"total"`
	Count int     `db:"count" json:"count"`
}

// MonthlyPurchaseTotal represents aggregated purchase totals per month.
type MonthlyPurchaseTotal struct {
	Month string  `db:"month" json:"month"`
	Total float64 `db:"total" json:"total"`
	Count int     `db:"count" json:"count"`
}

// CancelledSummary represents aggregated stats for cancelled orders.
type CancelledSummary struct {
	Count int     `db:"count" json:"count"`
	Biaya float64 `db:"biaya" json:"biaya_mitra"`
}

// NewDropshipRepo constructs a DropshipRepo given an *sqlx.DB connection.
func NewDropshipRepo(db DBTX) *DropshipRepo {
	return &DropshipRepo{db: db}
}

// InsertDropshipPurchase receives a *models.DropshipPurchase and executes an INSERT into dropship_purchases.
// It uses NamedExecContext so the struct fields map to column names automatically (via db tags).
func (r *DropshipRepo) InsertDropshipPurchase(ctx context.Context, p *models.DropshipPurchase) error {
	log.Printf("DropshipRepo.InsertDropshipPurchase %s", p.KodePesanan)
	query := `
        INSERT INTO dropship_purchases (
            kode_pesanan, kode_transaksi, waktu_pesanan_terbuat, status_pesanan_terakhir,
            biaya_lainnya, biaya_mitra_jakmall, total_transaksi, dibuat_oleh,
            jenis_channel, nama_toko, kode_invoice_channel, gudang_pengiriman,
            jenis_ekspedisi, cashless, nomor_resi, waktu_pengiriman,
            provinsi, kota
        ) VALUES (
            :kode_pesanan, :kode_transaksi, :waktu_pesanan_terbuat, :status_pesanan_terakhir,
            :biaya_lainnya, :biaya_mitra_jakmall, :total_transaksi, :dibuat_oleh,
            :jenis_channel, :nama_toko, :kode_invoice_channel, :gudang_pengiriman,
            :jenis_ekspedisi, :cashless, :nomor_resi, :waktu_pengiriman,
            :provinsi, :kota
        )
        ON CONFLICT (kode_pesanan) DO NOTHING`
	_, err := r.db.NamedExecContext(ctx, query, p)
	if err != nil {
		logutil.Errorf("InsertDropshipPurchase error: %v", err)
	}
	return err
}

// ExistsDropshipPurchase checks if a dropship purchase with the given kode_pesanan already exists.
func (r *DropshipRepo) ExistsDropshipPurchase(ctx context.Context, kodePesanan string) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM dropship_purchases WHERE kode_pesanan = $1)`, kodePesanan)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// ListExistingPurchases returns a map of kode_pesanan that already exist.
func (r *DropshipRepo) ListExistingPurchases(ctx context.Context, ids []string) (map[string]bool, error) {
	if len(ids) == 0 {
		return map[string]bool{}, nil
	}
	query, args, err := sqlx.In(`SELECT kode_pesanan FROM dropship_purchases WHERE kode_pesanan IN (?)`, ids)
	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)
	var list []string
	if err := r.db.SelectContext(ctx, &list, query, args...); err != nil {
		return nil, err
	}
	m := make(map[string]bool, len(list))
	for _, id := range list {
		m[id] = true
	}
	return m, nil
}

// InsertDropshipPurchaseDetail inserts a record into dropship_purchase_details.
func (r *DropshipRepo) InsertDropshipPurchaseDetail(ctx context.Context, d *models.DropshipPurchaseDetail) error {
	log.Printf("DropshipRepo.InsertDropshipPurchaseDetail %s %s", d.KodePesanan, d.SKU)
	query := `
        INSERT INTO dropship_purchase_details (
            kode_pesanan, sku, nama_produk, harga_produk, qty,
            total_harga_produk, harga_produk_channel, total_harga_produk_channel,
            potensi_keuntungan
        ) VALUES (
            :kode_pesanan, :sku, :nama_produk, :harga_produk, :qty,
            :total_harga_produk, :harga_produk_channel, :total_harga_produk_channel,
            :potensi_keuntungan
        )`
	_, err := r.db.NamedExecContext(ctx, query, d)
	if err != nil {
		logutil.Errorf("InsertDropshipPurchaseDetail error: %v", err)
	}
	return err
}

// GetDropshipPurchaseByID looks up a single row by purchase_id (the unique identifier in that table).
// It fills a models.DropshipPurchase struct with all columns from that row.
func (r *DropshipRepo) GetDropshipPurchaseByID(ctx context.Context, kodePesanan string) (*models.DropshipPurchase, error) {
	var p models.DropshipPurchase
	err := r.db.GetContext(ctx, &p,
		`SELECT * FROM dropship_purchases WHERE kode_pesanan = $1`, kodePesanan)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// GetDropshipPurchaseByInvoice retrieves a purchase by kode_invoice_channel.
func (r *DropshipRepo) GetDropshipPurchaseByInvoice(ctx context.Context, kodeInvoice string) (*models.DropshipPurchase, error) {
	log.Printf("DropshipRepo.GetDropshipPurchaseByInvoice %s", kodeInvoice)
	var p models.DropshipPurchase
	err := r.db.GetContext(ctx, &p,
		`SELECT * FROM dropship_purchases WHERE kode_invoice_channel = $1`, kodeInvoice)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// GetDropshipPurchaseByTransaction retrieves a purchase by kode_transaksi.
func (r *DropshipRepo) GetDropshipPurchaseByTransaction(ctx context.Context, kodeTransaksi string) (*models.DropshipPurchase, error) {
	var p models.DropshipPurchase
	err := r.db.GetContext(ctx, &p,
		`SELECT * FROM dropship_purchases WHERE kode_transaksi = $1`, kodeTransaksi)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// GetDropshipPurchasesByInvoices retrieves multiple purchases by invoice codes in a single query.
// This is an optimization to reduce N+1 query problems.
func (r *DropshipRepo) GetDropshipPurchasesByInvoices(ctx context.Context, invoices []string) ([]*models.DropshipPurchase, error) {
	if len(invoices) == 0 {
		return []*models.DropshipPurchase{}, nil
	}
	log.Printf("DropshipRepo.GetDropshipPurchasesByInvoices fetching %d invoices", len(invoices))
	
	// Build IN clause with placeholders
	placeholders := make([]string, len(invoices))
	args := make([]interface{}, len(invoices))
	for i, inv := range invoices {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = inv
	}
	
	query := fmt.Sprintf(`SELECT * FROM dropship_purchases WHERE kode_invoice_channel IN (%s)`, 
		strings.Join(placeholders, ","))
	
	var purchases []models.DropshipPurchase
	err := r.db.SelectContext(ctx, &purchases, query, args...)
	if err != nil {
		return nil, err
	}
	
	// Convert to slice of pointers
	result := make([]*models.DropshipPurchase, len(purchases))
	for i := range purchases {
		result[i] = &purchases[i]
	}
	
	log.Printf("DropshipRepo.GetDropshipPurchasesByInvoices found %d purchases", len(result))
	return result, nil
}

// SumDetailByInvoice sums total_harga_produk_channel for a given invoice.
func (r *DropshipRepo) SumDetailByInvoice(ctx context.Context, kodeInvoice string) (float64, error) {
	var sum float64
	err := r.db.GetContext(ctx, &sum,
		`SELECT COALESCE(SUM(d.total_harga_produk_channel),0)
                FROM dropship_purchase_details d
                JOIN dropship_purchases p ON d.kode_pesanan = p.kode_pesanan
                WHERE p.kode_invoice_channel=$1`, kodeInvoice)
	return sum, err
}

// SumProductCostByInvoice sums total_harga_produk for a given invoice.
func (r *DropshipRepo) SumProductCostByInvoice(ctx context.Context, kodeInvoice string) (float64, error) {
	var sum float64
	err := r.db.GetContext(ctx, &sum,
		`SELECT COALESCE(SUM(d.total_harga_produk),0)
                FROM dropship_purchase_details d
                JOIN dropship_purchases p ON d.kode_pesanan = p.kode_pesanan
                WHERE p.kode_invoice_channel=$1`, kodeInvoice)
	return sum, err
}

// UpdatePurchaseStatus sets status_pesanan_terakhir for the given kode_pesanan.
func (r *DropshipRepo) UpdatePurchaseStatus(ctx context.Context, kodePesanan, status string) error {
	log.Printf("DropshipRepo.UpdatePurchaseStatus %s %s", kodePesanan, status)
	_, err := r.db.ExecContext(ctx,
		`UPDATE dropship_purchases SET status_pesanan_terakhir=$2 WHERE kode_pesanan=$1`,
		kodePesanan, status)
	if err != nil {
		logutil.Errorf("UpdatePurchaseStatus error: %v", err)
	}
	return err
}

// UpdateDropshipStatus is a compatibility alias for UpdatePurchaseStatus.
// ShopeeService expects this method to exist on its dropship repository
// dependency, so we simply delegate to UpdatePurchaseStatus.
func (r *DropshipRepo) UpdateDropshipStatus(ctx context.Context, kodePesanan, status string) error {
	return r.UpdatePurchaseStatus(ctx, kodePesanan, status)
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
         WHERE nama_toko = $1
           AND waktu_pesanan_terbuat BETWEEN $2 AND $3
         ORDER BY waktu_pesanan_terbuat`,
		shop, from, to)
	if list == nil {
		list = []models.DropshipPurchase{}
	}
	return list, err
}

// ListDropshipPurchases returns dropship purchases filtered by optional channel,
// store and date range with pagination. Empty filter values are ignored.
func (r *DropshipRepo) ListDropshipPurchases(
	ctx context.Context,
	channel, store, from, to, orderNo, sortBy, dir string,
	limit, offset int,
) ([]models.DropshipPurchase, int, error) {
	countQuery := `SELECT COUNT(*) FROM dropship_purchases
                WHERE ($1 = '' OR jenis_channel = $1)
                  AND ($2 = '' OR nama_toko = $2)
                  AND ($3 = '' OR DATE(waktu_pesanan_terbuat) >= $3::date)
                  AND ($4 = '' OR DATE(waktu_pesanan_terbuat) <= $4::date)
                  AND ($5 = '' OR kode_invoice_channel ILIKE '%' || $5 || '%')`
	var total int
	if err := r.db.GetContext(ctx, &total, countQuery,
		channel, store, from, to, orderNo); err != nil {
		return nil, 0, err
	}

	sortCol := map[string]string{
		"kode_pesanan":          "kode_pesanan",
		"waktu_pesanan_terbuat": "waktu_pesanan_terbuat",
		"total_transaksi":       "total_transaksi",
	}[sortBy]
	if sortCol == "" {
		sortCol = "waktu_pesanan_terbuat"
	}
	direction := "ASC"
	if strings.ToLower(dir) == "desc" {
		direction = "DESC"
	}
	query := fmt.Sprintf(`SELECT * FROM dropship_purchases
                WHERE ($1 = '' OR jenis_channel = $1)
                  AND ($2 = '' OR nama_toko = $2)
                  AND ($3 = '' OR DATE(waktu_pesanan_terbuat) >= $3::date)
                  AND ($4 = '' OR DATE(waktu_pesanan_terbuat) <= $4::date)
                  AND ($5 = '' OR kode_invoice_channel ILIKE '%%' || $5 || '%%')
                ORDER BY %s %s
                LIMIT $6 OFFSET $7`, sortCol, direction)
	var list []models.DropshipPurchase
	err := r.db.SelectContext(ctx, &list, query,
		channel, store, from, to, orderNo, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	if list == nil {
		list = []models.DropshipPurchase{}
	}
	return list, total, nil
}

// SumDropshipPurchases returns the total sum of total_transaksi for all rows
// matching the provided filters.
func (r *DropshipRepo) SumDropshipPurchases(
	ctx context.Context,
	channel, store, from, to string,
) (float64, error) {
	query := `SELECT COALESCE(SUM(total_transaksi),0) FROM dropship_purchases
                WHERE ($1 = '' OR jenis_channel = $1)
                  AND ($2 = '' OR nama_toko = $2)
                  AND ($3 = '' OR DATE(waktu_pesanan_terbuat) >= $3::date)
                  AND ($4 = '' OR DATE(waktu_pesanan_terbuat) <= $4::date)`
	var sum float64
	if err := r.db.GetContext(ctx, &sum, query,
		channel, store, from, to); err != nil {
		return 0, err
	}
	return sum, nil
}

// ListDropshipPurchaseDetails returns detail rows for a given kode_pesanan.
func (r *DropshipRepo) ListDropshipPurchaseDetails(ctx context.Context, kodePesanan string) ([]models.DropshipPurchaseDetail, error) {
	var list []models.DropshipPurchaseDetail
	query := `SELECT * FROM dropship_purchase_details WHERE kode_pesanan=$1 ORDER BY id`
	if err := r.db.SelectContext(ctx, &list, query, kodePesanan); err != nil {
		return nil, err
	}
	if list == nil {
		list = []models.DropshipPurchaseDetail{}
	}
	return list, nil
}

// TopProducts aggregates sales by product name filtered by optional channel,
// store and date range parameters.
func (r *DropshipRepo) TopProducts(
	ctx context.Context,
	channel, store, from, to string,
	limit int,
) ([]models.ProductSales, error) {
	query := `SELECT d.nama_produk,
                COALESCE(SUM(d.qty),0) AS total_qty,
                COALESCE(SUM(d.total_harga_produk_channel),0) AS total_value
                FROM dropship_purchase_details d
                JOIN dropship_purchases p ON d.kode_pesanan = p.kode_pesanan
                WHERE ($1 = '' OR p.jenis_channel = $1)
                  AND ($2 = '' OR p.nama_toko = $2)
                  AND ($3 = '' OR DATE(p.waktu_pesanan_terbuat) >= $3::date)
                  AND ($4 = '' OR DATE(p.waktu_pesanan_terbuat) <= $4::date)
                GROUP BY d.nama_produk
                ORDER BY total_value DESC
                LIMIT $5`
	var list []models.ProductSales
	if err := r.db.SelectContext(ctx, &list, query, channel, store, from, to, limit); err != nil {
		return nil, err
	}
	if list == nil {
		list = []models.ProductSales{}
	}
	return list, nil
}

// DailyTotals sums total_harga_produk_channel grouped by date with optional
// channel, store and date range filters. It also counts distinct purchases per
// day.
func (r *DropshipRepo) DailyTotals(
	ctx context.Context,
	channel, store, from, to string,
) ([]DailyPurchaseTotal, error) {
	query := `SELECT
                DATE(p.waktu_pesanan_terbuat) AS date,
                COUNT(DISTINCT p.kode_pesanan) AS count,
                COALESCE(SUM(d.total_harga_produk_channel),0) AS total
                FROM dropship_purchase_details d
                JOIN dropship_purchases p ON d.kode_pesanan = p.kode_pesanan
                WHERE ($1 = '' OR p.jenis_channel = $1)
                  AND ($2 = '' OR p.nama_toko = $2)
                  AND ($3 = '' OR DATE(p.waktu_pesanan_terbuat) >= $3::date)
                  AND ($4 = '' OR DATE(p.waktu_pesanan_terbuat) <= $4::date)
                GROUP BY DATE(p.waktu_pesanan_terbuat)
                ORDER BY DATE(p.waktu_pesanan_terbuat)`
	var list []DailyPurchaseTotal
	if err := r.db.SelectContext(ctx, &list, query, channel, store, from, to); err != nil {
		return nil, err
	}
	if list == nil {
		list = []DailyPurchaseTotal{}
	}
	return list, nil
}

// MonthlyTotals sums totals grouped by month with optional filters.
func (r *DropshipRepo) MonthlyTotals(
	ctx context.Context,
	channel, store, from, to string,
) ([]MonthlyPurchaseTotal, error) {
	query := `SELECT
                TO_CHAR(DATE_TRUNC('month', p.waktu_pesanan_terbuat), 'YYYY-MM') AS month,
                COUNT(DISTINCT p.kode_pesanan) AS count,
                COALESCE(SUM(d.total_harga_produk_channel),0) AS total
                FROM dropship_purchase_details d
                JOIN dropship_purchases p ON d.kode_pesanan = p.kode_pesanan
                WHERE ($1 = '' OR p.jenis_channel = $1)
                  AND ($2 = '' OR p.nama_toko = $2)
                  AND ($3 = '' OR DATE(p.waktu_pesanan_terbuat) >= $3::date)
                  AND ($4 = '' OR DATE(p.waktu_pesanan_terbuat) <= $4::date)
                GROUP BY DATE_TRUNC('month', p.waktu_pesanan_terbuat)
                ORDER BY DATE_TRUNC('month', p.waktu_pesanan_terbuat)`
	var list []MonthlyPurchaseTotal
	if err := r.db.SelectContext(ctx, &list, query, channel, store, from, to); err != nil {
		return nil, err
	}
	if list == nil {
		list = []MonthlyPurchaseTotal{}
	}
	return list, nil
}

// CancelledSummary returns count of cancelled orders and total Biaya Mitra
// filtered by optional channel, store and date range.
func (r *DropshipRepo) CancelledSummary(
	ctx context.Context,
	channel, store, from, to string,
) (CancelledSummary, error) {
	query := `SELECT COUNT(*) AS count,
                COALESCE(SUM(biaya_mitra_jakmall),0) AS biaya
                FROM dropship_purchases
                WHERE status_pesanan_terakhir = 'Cancelled Shopee'
                  AND ($1 = '' OR jenis_channel = $1)
                  AND ($2 = '' OR nama_toko = $2)
                  AND ($3 = '' OR DATE(waktu_pesanan_terbuat) >= $3::date)
                  AND ($4 = '' OR DATE(waktu_pesanan_terbuat) <= $4::date)`
	var res CancelledSummary
	if err := r.db.GetContext(ctx, &res, query, channel, store, from, to); err != nil {
		return CancelledSummary{}, err
	}
	return res, nil
}

// CountOrders returns the number of purchases matching the filters.
func (r *DropshipRepo) CountOrders(ctx context.Context, channel, store, from, to string) (int, error) {
	query := `SELECT COUNT(*) FROM dropship_purchases
                WHERE ($1 = '' OR jenis_channel = $1)
                  AND ($2 = '' OR nama_toko = $2)
                  AND ($3 = '' OR DATE(waktu_pesanan_terbuat) >= $3::date)
                  AND ($4 = '' OR DATE(waktu_pesanan_terbuat) <= $4::date)`
	var n int
	if err := r.db.GetContext(ctx, &n, query, channel, store, from, to); err != nil {
		return 0, err
	}
	return n, nil
}

// AvgOrderValue returns the average total_transaksi for purchases matching the filters.
func (r *DropshipRepo) AvgOrderValue(ctx context.Context, channel, store, from, to string) (float64, error) {
	query := `SELECT COALESCE(AVG(total_transaksi),0) FROM dropship_purchases
                WHERE ($1 = '' OR jenis_channel = $1)
                  AND ($2 = '' OR nama_toko = $2)
                  AND ($3 = '' OR DATE(waktu_pesanan_terbuat) >= $3::date)
                  AND ($4 = '' OR DATE(waktu_pesanan_terbuat) <= $4::date)`
	var v float64
	if err := r.db.GetContext(ctx, &v, query, channel, store, from, to); err != nil {
		return 0, err
	}
	return v, nil
}

// DistinctCustomers counts unique dibuat_oleh values matching the filters.
func (r *DropshipRepo) DistinctCustomers(ctx context.Context, channel, store, from, to string) (int, error) {
	query := `SELECT COUNT(DISTINCT dibuat_oleh) FROM dropship_purchases
                WHERE ($1 = '' OR jenis_channel = $1)
                  AND ($2 = '' OR nama_toko = $2)
                  AND ($3 = '' OR DATE(waktu_pesanan_terbuat) >= $3::date)
                  AND ($4 = '' OR DATE(waktu_pesanan_terbuat) <= $4::date)`
	var n int
	if err := r.db.GetContext(ctx, &n, query, channel, store, from, to); err != nil {
		return 0, err
	}
	return n, nil
}
