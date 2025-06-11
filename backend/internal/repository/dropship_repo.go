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

// NewDropshipRepo constructs a DropshipRepo given an *sqlx.DB connection.
func NewDropshipRepo(db *sqlx.DB) *DropshipRepo {
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
        )`
	_, err := r.db.NamedExecContext(ctx, query, p)
	return err
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
	return list, err
}
