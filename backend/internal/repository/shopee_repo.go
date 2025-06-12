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

// InsertShopeeSettled inserts a row into the shopee_settled table.
func (r *ShopeeRepo) InsertShopeeSettled(ctx context.Context, s *models.ShopeeSettled) error {
	query := `
        INSERT INTO shopee_settled (
            nama_toko, no_pesanan, no_pengajuan, username_pembeli, waktu_pesanan_dibuat,
            metode_pembayaran_pembeli, tanggal_dana_dilepaskan, harga_asli_produk,
            total_diskon_produk, jumlah_pengembalian_dana_ke_pembeli, komisi_shopee,
            biaya_admin_shopee, biaya_layanan, biaya_layanan_ekstra,
            biaya_penyedia_pembayaran, asuransi, total_biaya_transaksi,
            biaya_pengiriman, total_diskon_pengiriman, promo_gratis_ongkir_shopee,
            promo_gratis_ongkir_penjual, promo_diskon_shopee, promo_diskon_penjual,
            cashback_shopee, cashback_penjual, koin_shopee, potongan_lainnya,
            total_penerimaan, kompensasi, promo_gratis_ongkir_dari_penjual,
            jasa_kirim, nama_kurir, pengembalian_dana_ke_pembeli,
            pro_rata_koin_yang_ditukarkan_untuk_pengembalian_barang,
            pro_rata_voucher_shopee_untuk_pengembalian_barang,
            pro_rated_bank_payment_channel_promotion_for_returns,
            pro_rated_shopee_payment_channel_promotion_for_returns
        ) VALUES (
            :nama_toko, :no_pesanan, :no_pengajuan, :username_pembeli, :waktu_pesanan_dibuat,
            :metode_pembayaran_pembeli, :tanggal_dana_dilepaskan, :harga_asli_produk,
            :total_diskon_produk, :jumlah_pengembalian_dana_ke_pembeli, :komisi_shopee,
            :biaya_admin_shopee, :biaya_layanan, :biaya_layanan_ekstra,
            :biaya_penyedia_pembayaran, :asuransi, :total_biaya_transaksi,
            :biaya_pengiriman, :total_diskon_pengiriman, :promo_gratis_ongkir_shopee,
            :promo_gratis_ongkir_penjual, :promo_diskon_shopee, :promo_diskon_penjual,
            :cashback_shopee, :cashback_penjual, :koin_shopee, :potongan_lainnya,
            :total_penerimaan, :kompensasi, :promo_gratis_ongkir_dari_penjual,
            :jasa_kirim, :nama_kurir, :pengembalian_dana_ke_pembeli,
            :pro_rata_koin_yang_ditukarkan_untuk_pengembalian_barang,
            :pro_rata_voucher_shopee_untuk_pengembalian_barang,
            :pro_rated_bank_payment_channel_promotion_for_returns,
            :pro_rated_shopee_payment_channel_promotion_for_returns
        )`
	_, err := r.db.NamedExecContext(ctx, query, s)
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
