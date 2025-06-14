package repository

import (
	"context"
	"fmt"
	"strings"

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

// InsertShopeeSettled inserts a row into the shopee_settled table.
func (r *ShopeeRepo) InsertShopeeSettled(ctx context.Context, s *models.ShopeeSettled) error {
	query := `
        INSERT INTO shopee_settled (
            nama_toko, no_pesanan, no_pengajuan, username_pembeli, waktu_pesanan_dibuat,
            metode_pembayaran_pembeli, tanggal_dana_dilepaskan, harga_asli_produk,
            total_diskon_produk, jumlah_pengembalian_dana_ke_pembeli, diskon_produk_dari_shopee,
            diskon_voucher_ditanggung_penjual, cashback_koin_ditanggung_penjual, ongkir_dibayar_pembeli,
            diskon_ongkir_ditanggung_jasa_kirim, gratis_ongkir_dari_shopee, ongkir_yang_diteruskan_oleh_shopee_ke_jasa_kirim,
            ongkos_kirim_pengembalian_barang, pengembalian_biaya_kirim, biaya_komisi_ams,
            biaya_administrasi, biaya_layanan_termasuk_ppn_11, premi,
            biaya_program, biaya_kartu_kredit, biaya_kampanye, bea_masuk_ppn_pph,
            total_penghasilan, kompensasi, promo_gratis_ongkir_dari_penjual,
            jasa_kirim, nama_kurir, pengembalian_dana_ke_pembeli,
            pro_rata_koin_yang_ditukarkan_untuk_pengembalian_barang,
            pro_rata_voucher_shopee_untuk_pengembalian_barang,
            pro_rated_bank_payment_channel_promotion_for_returns,
            pro_rated_shopee_payment_channel_promotion_for_returns
        ) VALUES (
            :nama_toko, :no_pesanan, :no_pengajuan, :username_pembeli, :waktu_pesanan_dibuat,
            :metode_pembayaran_pembeli, :tanggal_dana_dilepaskan, :harga_asli_produk,
            :total_diskon_produk, :jumlah_pengembalian_dana_ke_pembeli, :diskon_produk_dari_shopee,
            :diskon_voucher_ditanggung_penjual, :cashback_koin_ditanggung_penjual, :ongkir_dibayar_pembeli,
            :diskon_ongkir_ditanggung_jasa_kirim, :gratis_ongkir_dari_shopee, :ongkir_yang_diteruskan_oleh_shopee_ke_jasa_kirim,
            :ongkos_kirim_pengembalian_barang, :pengembalian_biaya_kirim, :biaya_komisi_ams,
            :biaya_administrasi, :biaya_layanan_termasuk_ppn_11, :premi,
            :biaya_program, :biaya_kartu_kredit, :biaya_kampanye, :bea_masuk_ppn_pph,
            :total_penghasilan, :kompensasi, :promo_gratis_ongkir_dari_penjual,
            :jasa_kirim, :nama_kurir, :pengembalian_dana_ke_pembeli,
            :pro_rata_koin_yang_ditukarkan_untuk_pengembalian_barang,
            :pro_rata_voucher_shopee_untuk_pengembalian_barang,
            :pro_rated_bank_payment_channel_promotion_for_returns,
            :pro_rated_shopee_payment_channel_promotion_for_returns
        )`
	_, err := r.db.NamedExecContext(ctx, query, s)
	return err
}

// ExistsShopeeSettled checks if a row with the given order number already exists.
func (r *ShopeeRepo) ExistsShopeeSettled(ctx context.Context, noPesanan string) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists,
		"SELECT EXISTS(SELECT 1 FROM shopee_settled WHERE no_pesanan=$1)", noPesanan)
	return exists, err
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

// ListShopeeSettled returns shopee_settled rows filtered by optional parameters.
// Pagination is controlled via limit and offset and the total count is returned
// alongside the slice of rows.
func (r *ShopeeRepo) ListShopeeSettled(
	ctx context.Context,
	channel, store, date, month, year string,
	limit, offset int,
) ([]models.ShopeeSettled, int, error) {
	base := `SELECT s.* FROM shopee_settled s
        LEFT JOIN stores st ON s.nama_toko = st.nama_toko
        LEFT JOIN jenis_channels jc ON st.jenis_channel_id = jc.jenis_channel_id`
	args := []interface{}{}
	conds := []string{}
	arg := 1
	if channel != "" {
		conds = append(conds, fmt.Sprintf("jc.jenis_channel = $%d", arg))
		args = append(args, channel)
		arg++
	}
	if store != "" {
		conds = append(conds, fmt.Sprintf("s.nama_toko = $%d", arg))
		args = append(args, store)
		arg++
	}
	if date != "" {
		conds = append(conds, fmt.Sprintf("s.waktu_pesanan_dibuat = $%d::date", arg))
		args = append(args, date)
		arg++
	}
	if month != "" {
		conds = append(conds, fmt.Sprintf("EXTRACT(MONTH FROM s.waktu_pesanan_dibuat) = $%d::int", arg))
		args = append(args, month)
		arg++
	}
	if year != "" {
		conds = append(conds, fmt.Sprintf("EXTRACT(YEAR FROM s.waktu_pesanan_dibuat) = $%d::int", arg))
		args = append(args, year)
		arg++
	}
	query := base
	if len(conds) > 0 {
		query += " WHERE " + strings.Join(conds, " AND ")
	}
	countQuery := "SELECT COUNT(*) FROM (" + query + ") AS sub"
	var count int
	if err := r.db.GetContext(ctx, &count, countQuery, args...); err != nil {
		return nil, 0, err
	}
	query += fmt.Sprintf(" ORDER BY s.waktu_pesanan_dibuat DESC LIMIT %d OFFSET %d", limit, offset)
	var list []models.ShopeeSettled
	if err := r.db.SelectContext(ctx, &list, query, args...); err != nil {
		return nil, 0, err
	}
	if list == nil {
		list = []models.ShopeeSettled{}
	}
	return list, count, nil
}

// SumShopeeSettled returns the sum of total_penghasilan for rows matching the filters.
func (r *ShopeeRepo) SumShopeeSettled(
	ctx context.Context,
	channel, store, date, month, year string,
) (*models.ShopeeSummary, error) {
	base := `SELECT
                COALESCE(SUM(s.harga_asli_produk),0) AS harga_asli_produk,
                COALESCE(SUM(s.total_diskon_produk),0) AS total_diskon_produk,
                COALESCE(SUM(s.diskon_voucher_ditanggung_penjual),0) AS diskon_voucher_ditanggung_penjual,
                COALESCE(SUM(s.biaya_administrasi),0) AS biaya_administrasi,
                COALESCE(SUM(s.biaya_layanan_termasuk_ppn_11),0) AS biaya_layanan_termasuk_ppn_11,
                COALESCE(SUM(s.total_penghasilan),0) AS total_penghasilan
        FROM shopee_settled s
        LEFT JOIN stores st ON s.nama_toko = st.nama_toko
        LEFT JOIN jenis_channels jc ON st.jenis_channel_id = jc.jenis_channel_id`
	args := []interface{}{}
	conds := []string{}
	arg := 1
	if channel != "" {
		conds = append(conds, fmt.Sprintf("jc.jenis_channel = $%d", arg))
		args = append(args, channel)
		arg++
	}
	if store != "" {
		conds = append(conds, fmt.Sprintf("s.nama_toko = $%d", arg))
		args = append(args, store)
		arg++
	}
	if date != "" {
		conds = append(conds, fmt.Sprintf("s.waktu_pesanan_dibuat = $%d::date", arg))
		args = append(args, date)
		arg++
	}
	if month != "" {
		conds = append(conds, fmt.Sprintf("EXTRACT(MONTH FROM s.waktu_pesanan_dibuat) = $%d::int", arg))
		args = append(args, month)
		arg++
	}
	if year != "" {
		conds = append(conds, fmt.Sprintf("EXTRACT(YEAR FROM s.waktu_pesanan_dibuat) = $%d::int", arg))
		args = append(args, year)
		arg++
	}
	query := base
	if len(conds) > 0 {
		query += " WHERE " + strings.Join(conds, " AND ")
	}
	var sum models.ShopeeSummary
	if err := r.db.GetContext(ctx, &sum, query, args...); err != nil {
		return nil, err
	}
	sum.GMV = sum.HargaAsliProduk - sum.TotalDiskonProduk
	return &sum, nil
}
