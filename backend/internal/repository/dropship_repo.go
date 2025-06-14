package repository

import (
	"context"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// DropshipRepo handles all database operations related to the dropship_purchases table.
type DropshipRepo struct {
	db DBTX
}

// NewDropshipRepo constructs a DropshipRepo given an *sqlx.DB connection.
func NewDropshipRepo(db DBTX) *DropshipRepo {
	return &DropshipRepo{db: db}
}

// InsertDropshipPurchase receives a *models.DropshipPurchase and executes an INSERT into dropship_purchases.
// It uses NamedExecContext so the struct fields map to column names automatically (via db tags).
func (r *DropshipRepo) InsertDropshipPurchase(ctx context.Context, p *models.DropshipPurchase) error {
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

// InsertDropshipPurchaseDetail inserts a record into dropship_purchase_details.
func (r *DropshipRepo) InsertDropshipPurchaseDetail(ctx context.Context, d *models.DropshipPurchaseDetail) error {
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
	var p models.DropshipPurchase
	err := r.db.GetContext(ctx, &p,
		`SELECT * FROM dropship_purchases WHERE kode_invoice_channel = $1`, kodeInvoice)
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
// store, date, month and year with pagination.
// Empty filter values are ignored. Pagination uses limit & offset.
func (r *DropshipRepo) ListDropshipPurchases(
	ctx context.Context,
	channel, store, date, month, year string,
	limit, offset int,
) ([]models.DropshipPurchase, int, error) {
	countQuery := `SELECT COUNT(*) FROM dropship_purchases
                WHERE ($1 = '' OR jenis_channel = $1)
                  AND ($2 = '' OR nama_toko = $2)
                  AND ($3 = '' OR DATE(waktu_pesanan_terbuat) = $3::date)
                  AND ($4 = '' OR EXTRACT(MONTH FROM waktu_pesanan_terbuat) = $4::int)
                  AND ($5 = '' OR EXTRACT(YEAR FROM waktu_pesanan_terbuat) = $5::int)`
	var total int
	if err := r.db.GetContext(ctx, &total, countQuery,
		channel, store, date, month, year); err != nil {
		return nil, 0, err
	}

	query := `SELECT * FROM dropship_purchases
                WHERE ($1 = '' OR jenis_channel = $1)
                  AND ($2 = '' OR nama_toko = $2)
                  AND ($3 = '' OR DATE(waktu_pesanan_terbuat) = $3::date)
                  AND ($4 = '' OR EXTRACT(MONTH FROM waktu_pesanan_terbuat) = $4::int)
                  AND ($5 = '' OR EXTRACT(YEAR FROM waktu_pesanan_terbuat) = $5::int)
                ORDER BY waktu_pesanan_terbuat DESC
                LIMIT $6 OFFSET $7`
	var list []models.DropshipPurchase
	err := r.db.SelectContext(ctx, &list, query,
		channel, store, date, month, year, limit, offset)
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
	channel, store, date, month, year string,
) (float64, error) {
	query := `SELECT COALESCE(SUM(total_transaksi),0) FROM dropship_purchases
                WHERE ($1 = '' OR jenis_channel = $1)
                  AND ($2 = '' OR nama_toko = $2)
                  AND ($3 = '' OR DATE(waktu_pesanan_terbuat) = $3::date)
                  AND ($4 = '' OR EXTRACT(MONTH FROM waktu_pesanan_terbuat) = $4::int)
                  AND ($5 = '' OR EXTRACT(YEAR FROM waktu_pesanan_terbuat) = $5::int)`
	var sum float64
	if err := r.db.GetContext(ctx, &sum, query,
		channel, store, date, month, year); err != nil {
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

// TopProducts aggregates sales by product name filtered by optional parameters.
func (r *DropshipRepo) TopProducts(
	ctx context.Context,
	channel, store, month, year string,
	limit int,
) ([]models.ProductSales, error) {
	query := `SELECT d.nama_produk,
                COALESCE(SUM(d.qty),0) AS total_qty,
                COALESCE(SUM(d.total_harga_produk_channel),0) AS total_value
                FROM dropship_purchase_details d
                JOIN dropship_purchases p ON d.kode_pesanan = p.kode_pesanan
                WHERE ($1 = '' OR p.jenis_channel = $1)
                  AND ($2 = '' OR p.nama_toko = $2)
                  AND ($3 = '' OR EXTRACT(MONTH FROM p.waktu_pesanan_terbuat) = $3::int)
                  AND ($4 = '' OR EXTRACT(YEAR FROM p.waktu_pesanan_terbuat) = $4::int)
                GROUP BY d.nama_produk
                ORDER BY total_value DESC
                LIMIT $5`
	var list []models.ProductSales
	if err := r.db.SelectContext(ctx, &list, query, channel, store, month, year, limit); err != nil {
		return nil, err
	}
	if list == nil {
		list = []models.ProductSales{}
	}
	return list, nil
}
