package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// ShopeeRepo handles interactions with the shopee_settled_orders table.
type ShopeeRepo struct {
	db *sqlx.DB
}

// ListShopeeOrdersByShopAndDate implements service.MetricServiceShopeeRepo.
func (r *ShopeeRepo) ListShopeeOrdersByShopAndDate(
	ctx context.Context,
	shop string,
	from string,
	to string,
) ([]models.ShopeeSettledOrder, error) {
	var list []models.ShopeeSettledOrder
	err := r.db.SelectContext(ctx, &list,
		`SELECT * FROM shopee_settled_orders
         WHERE seller_username = $1
           AND settled_date BETWEEN $2 AND $3
         ORDER BY settled_date`,
		shop, from, to)
	if list == nil {
		list = []models.ShopeeSettledOrder{}
	}
	return list, err
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
            pro_rated_shopee_payment_channel_promotion_for_returns,
            is_data_mismatch,
            is_settled_confirmed
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
            :pro_rated_shopee_payment_channel_promotion_for_returns,
            :is_data_mismatch,
            :is_settled_confirmed
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

// MarkMismatch updates the is_data_mismatch flag for a given order number.
func (r *ShopeeRepo) MarkMismatch(ctx context.Context, orderSN string, mismatch bool) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE shopee_settled SET is_data_mismatch=$2 WHERE no_pesanan=$1`,
		orderSN, mismatch)
	return err
}

// ConfirmSettle sets the is_settled_confirmed flag to true for the given order number.
func (r *ShopeeRepo) ConfirmSettle(ctx context.Context, orderSN string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE shopee_settled SET is_settled_confirmed=TRUE WHERE no_pesanan=$1`,
		orderSN)
	return err
}

// GetBySN retrieves a single shopee_settled row by order number.
func (r *ShopeeRepo) GetBySN(ctx context.Context, orderSN string) (*models.ShopeeSettled, error) {
	var s models.ShopeeSettled
	err := r.db.GetContext(ctx, &s, `SELECT * FROM shopee_settled WHERE no_pesanan=$1`, orderSN)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// InsertShopeeAffiliateSale inserts a row into the shopee_affiliate_sales table.
func (r *ShopeeRepo) InsertShopeeAffiliateSale(ctx context.Context, s *models.ShopeeAffiliateSale) error {
	query := `
        INSERT INTO shopee_affiliate_sales (
                nama_toko,
                kode_pesanan, status_pesanan, status_terverifikasi,
                waktu_pesanan, waktu_pesanan_selesai, waktu_pesanan_terverifikasi,
                kode_produk, nama_produk, id_model, l1_kategori_global, l2_kategori_global,
                l3_kategori_global, kode_promo, harga, jumlah, nama_affiliate,
                username_affiliate, mcn_terhubung, id_komisi_pesanan, partner_promo,
                jenis_promo, nilai_pembelian, jumlah_pengembalian, tipe_pesanan,
                estimasi_komisi_per_produk, estimasi_komisi_affiliate_per_produk,
                persentase_komisi_affiliate_per_produk, estimasi_komisi_mcn_per_produk,
                persentase_komisi_mcn_per_produk, estimasi_komisi_per_pesanan,
                estimasi_komisi_affiliate_per_pesanan, estimasi_komisi_mcn_per_pesanan,
                catatan_produk, platform, tingkat_komisi, pengeluaran,
                status_pemotongan, metode_pemotongan, waktu_pemotongan
        ) VALUES (
                :nama_toko,
                :kode_pesanan, :status_pesanan, :status_terverifikasi,
                :waktu_pesanan, :waktu_pesanan_selesai, :waktu_pesanan_terverifikasi,
                :kode_produk, :nama_produk, :id_model, :l1_kategori_global, :l2_kategori_global,
                :l3_kategori_global, :kode_promo, :harga, :jumlah, :nama_affiliate,
                :username_affiliate, :mcn_terhubung, :id_komisi_pesanan, :partner_promo,
                :jenis_promo, :nilai_pembelian, :jumlah_pengembalian, :tipe_pesanan,
                :estimasi_komisi_per_produk, :estimasi_komisi_affiliate_per_produk,
                :persentase_komisi_affiliate_per_produk, :estimasi_komisi_mcn_per_produk,
                :persentase_komisi_mcn_per_produk, :estimasi_komisi_per_pesanan,
                :estimasi_komisi_affiliate_per_pesanan, :estimasi_komisi_mcn_per_pesanan,
                :catatan_produk, :platform, :tingkat_komisi, :pengeluaran,
                :status_pemotongan, :metode_pemotongan, :waktu_pemotongan
        )`
	_, err := r.db.NamedExecContext(ctx, query, s)
	return err
}

// ExistsShopeeAffiliateSale checks if a row exists for the given order and product.
func (r *ShopeeRepo) ExistsShopeeAffiliateSale(ctx context.Context, orderID, productCode string) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists,
		`SELECT EXISTS(SELECT 1 FROM shopee_affiliate_sales WHERE kode_pesanan=$1 AND kode_produk=$2)`, orderID, productCode)
	return exists, err
}

// ListShopeeAffiliateSales returns affiliate sales filtered by optional date/month/year with pagination.
func (r *ShopeeRepo) ListShopeeAffiliateSales(
	ctx context.Context,
	noPesanan, from, to string,
	limit, offset int,
) ([]models.ShopeeAffiliateSale, int, error) {
	base := `SELECT * FROM shopee_affiliate_sales`
	args := []interface{}{}
	conds := []string{}
	arg := 1
	if noPesanan != "" {
		conds = append(conds, fmt.Sprintf("kode_pesanan = $%d", arg))
		args = append(args, noPesanan)
		arg++
	}
	if from != "" {
		conds = append(conds, fmt.Sprintf("DATE(waktu_pesanan) >= $%d::date", arg))
		args = append(args, from)
		arg++
	}
	if to != "" {
		conds = append(conds, fmt.Sprintf("DATE(waktu_pesanan) <= $%d::date", arg))
		args = append(args, to)
		arg++
	}
	query := base
	if len(conds) > 0 {
		query += " WHERE " + strings.Join(conds, " AND ")
	}
	countQuery := "SELECT COUNT(*) FROM (" + query + ") AS sub"
	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}
	query += fmt.Sprintf(" ORDER BY waktu_pesanan DESC LIMIT %d OFFSET %d", limit, offset)
	var list []models.ShopeeAffiliateSale
	if err := r.db.SelectContext(ctx, &list, query, args...); err != nil {
		return nil, 0, err
	}
	if list == nil {
		list = []models.ShopeeAffiliateSale{}
	}
	return list, total, nil
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

// ListShopeeSettled returns shopee_settled rows filtered by optional channel,
// store and date range parameters. Pagination is controlled via limit and offset
// and the total count is returned alongside the slice of rows.
func (r *ShopeeRepo) ListShopeeSettled(
	ctx context.Context,
	channel, store, from, to, orderNo, sortBy, dir string,
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
	if from != "" {
		conds = append(conds, fmt.Sprintf("DATE(s.waktu_pesanan_dibuat) >= $%d::date", arg))
		args = append(args, from)
		arg++
	}
	if to != "" {
		conds = append(conds, fmt.Sprintf("DATE(s.waktu_pesanan_dibuat) <= $%d::date", arg))
		args = append(args, to)
		arg++
	}
	if orderNo != "" {
		conds = append(conds, fmt.Sprintf("s.no_pesanan ILIKE $%d", arg))
		args = append(args, "%"+orderNo+"%")
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
	sortCol := map[string]string{
		"no_pesanan":           "no_pesanan",
		"waktu_pesanan_dibuat": "waktu_pesanan_dibuat",
		"total_penghasilan":    "total_penghasilan",
	}[sortBy]
	if sortCol == "" {
		sortCol = "waktu_pesanan_dibuat"
	}
	direction := "ASC"
	if strings.ToLower(dir) == "desc" {
		direction = "DESC"
	}
	args = append(args, limit, offset)
	query += fmt.Sprintf(" ORDER BY s.%s %s LIMIT $%d OFFSET $%d", sortCol, direction, arg, arg+1)
	var list []models.ShopeeSettled
	if err := r.db.SelectContext(ctx, &list, query, args...); err != nil {
		return nil, 0, err
	}
	if list == nil {
		list = []models.ShopeeSettled{}
	}
	return list, count, nil
}

// SumShopeeSettled returns the aggregated totals for rows matching the provided
// channel, store and optional date range filters.
func (r *ShopeeRepo) SumShopeeSettled(
	ctx context.Context,
	channel, store, from, to string,
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
	if from != "" {
		conds = append(conds, fmt.Sprintf("DATE(s.waktu_pesanan_dibuat) >= $%d::date", arg))
		args = append(args, from)
		arg++
	}
	if to != "" {
		conds = append(conds, fmt.Sprintf("DATE(s.waktu_pesanan_dibuat) <= $%d::date", arg))
		args = append(args, to)
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

// SumShopeeAffiliateSales aggregates nilai_pembelian and komisi for filtered rows.
func (r *ShopeeRepo) SumShopeeAffiliateSales(
	ctx context.Context,
	noPesanan, from, to string,
) (*models.ShopeeAffiliateSummary, error) {
	base := `SELECT
               COALESCE(SUM(nilai_pembelian),0) AS total_nilai_pembelian,
               COALESCE(SUM(estimasi_komisi_affiliate_per_pesanan),0) AS total_komisi_affiliate
               FROM shopee_affiliate_sales`
	args := []interface{}{}
	conds := []string{}
	arg := 1
	if noPesanan != "" {
		conds = append(conds, fmt.Sprintf("kode_pesanan = $%d", arg))
		args = append(args, noPesanan)
		arg++
	}
	if from != "" {
		conds = append(conds, fmt.Sprintf("DATE(waktu_pesanan) >= $%d::date", arg))
		args = append(args, from)
		arg++
	}
	if to != "" {
		conds = append(conds, fmt.Sprintf("DATE(waktu_pesanan) <= $%d::date", arg))
		args = append(args, to)
		arg++
	}
	query := base
	if len(conds) > 0 {
		query += " WHERE " + strings.Join(conds, " AND ")
	}
	var sum models.ShopeeAffiliateSummary
	if err := r.db.GetContext(ctx, &sum, query, args...); err != nil {
		return nil, err
	}
	return &sum, nil
}

// GetAffiliateExpenseByOrder sums the pengeluaran for a given kode_pesanan.
func (r *ShopeeRepo) GetAffiliateExpenseByOrder(ctx context.Context, kodePesanan string) (float64, error) {
	var amt float64
	err := r.db.GetContext(ctx, &amt,
		`SELECT COALESCE(SUM(pengeluaran),0) FROM shopee_affiliate_sales WHERE kode_pesanan=$1`,
		kodePesanan)
	return amt, err
}

// ListSalesProfit returns joined dropship and shopee sales with cost breakdown.
func (r *ShopeeRepo) ListSalesProfit(
	ctx context.Context,
	channel, store, from, to, orderNo, sortBy, dir string,
	limit, offset int,
) ([]models.SalesProfit, int, error) {
	base := `SELECT
                je.source_id AS kode_pesanan,
                dp.waktu_pesanan_terbuat AS tanggal_pesanan,
                SUM(CASE WHEN jl.account_id = 5001 THEN jl.amount ELSE 0 END) AS modal_purchase,
                SUM(CASE WHEN jl.account_id IN (11010,11012) AND jl.is_debit = false THEN jl.amount ELSE 0 END) AS amount_sales,
                SUM(CASE WHEN jl.account_id = 52007 THEN jl.amount ELSE 0 END) AS biaya_mitra_jakmall,
                SUM(CASE WHEN jl.account_id = 52006 THEN jl.amount ELSE 0 END) AS biaya_administrasi,
                SUM(CASE WHEN jl.account_id = 52004 THEN jl.amount ELSE 0 END) AS biaya_layanan,
                SUM(CASE WHEN jl.account_id = 52003 THEN jl.amount ELSE 0 END) AS biaya_voucher,
                SUM(CASE WHEN jl.account_id = 52005 THEN jl.amount ELSE 0 END) AS biaya_affiliate,
                SUM(CASE WHEN jl.account_id IN (11010,11012) AND jl.is_debit = false THEN jl.amount ELSE 0 END)
                  - (SUM(CASE WHEN jl.account_id = 5001 THEN jl.amount ELSE 0 END)
                     + SUM(CASE WHEN jl.account_id = 52007 THEN jl.amount ELSE 0 END)
                     + SUM(CASE WHEN jl.account_id = 52006 THEN jl.amount ELSE 0 END)
                     + SUM(CASE WHEN jl.account_id = 52004 THEN jl.amount ELSE 0 END)
                     + SUM(CASE WHEN jl.account_id = 52003 THEN jl.amount ELSE 0 END)
                     + SUM(CASE WHEN jl.account_id = 52005 THEN jl.amount ELSE 0 END)) AS profit,
                CASE WHEN SUM(CASE WHEN jl.account_id IN (11010,11012) AND jl.is_debit = false THEN jl.amount ELSE 0 END) = 0 THEN 0
                     ELSE (SUM(CASE WHEN jl.account_id IN (11010,11012) AND jl.is_debit = false THEN jl.amount ELSE 0 END)
                          - (SUM(CASE WHEN jl.account_id = 5001 THEN jl.amount ELSE 0 END)
                             + SUM(CASE WHEN jl.account_id = 52007 THEN jl.amount ELSE 0 END)
                             + SUM(CASE WHEN jl.account_id = 52006 THEN jl.amount ELSE 0 END)
                             + SUM(CASE WHEN jl.account_id = 52004 THEN jl.amount ELSE 0 END)
                             + SUM(CASE WHEN jl.account_id = 52003 THEN jl.amount ELSE 0 END)
                             + SUM(CASE WHEN jl.account_id = 52005 THEN jl.amount ELSE 0 END))
                     ) / SUM(CASE WHEN jl.account_id IN (11010,11012) AND jl.is_debit = false THEN jl.amount ELSE 0 END) * 100 END AS profit_percent
                FROM journal_entries je
                JOIN dropship_purchases dp ON dp.kode_invoice_channel = je.source_id
                JOIN journal_lines jl ON jl.journal_id = je.journal_id
                JOIN stores st ON dp.nama_toko = st.nama_toko
                JOIN jenis_channels jc ON st.jenis_channel_id = jc.jenis_channel_id
                WHERE je.source_type IN ('pending_sales','shopee_settled')`
	args := []interface{}{}
	conds := []string{}
	arg := 1
	if channel != "" {
		conds = append(conds, fmt.Sprintf("jc.jenis_channel = $%d", arg))
		args = append(args, channel)
		arg++
	}
	if store != "" {
		conds = append(conds, fmt.Sprintf("dp.nama_toko = $%d", arg))
		args = append(args, store)
		arg++
	}
	if from != "" {
		conds = append(conds, fmt.Sprintf("DATE(dp.waktu_pesanan_terbuat) >= $%d::date", arg))
		args = append(args, from)
		arg++
	}
	if to != "" {
		conds = append(conds, fmt.Sprintf("DATE(dp.waktu_pesanan_terbuat) <= $%d::date", arg))
		args = append(args, to)
		arg++
	}
	if orderNo != "" {
		conds = append(conds, fmt.Sprintf("je.source_id ILIKE $%d", arg))
		args = append(args, "%"+orderNo+"%")
		arg++
	}
	query := base
	if len(conds) > 0 {
		query += " AND " + strings.Join(conds, " AND ")
	}
	query += " GROUP BY je.source_id, dp.waktu_pesanan_terbuat"
	countQuery := "SELECT COUNT(*) FROM (" + query + ") AS sub"
	var count int
	if err := r.db.GetContext(ctx, &count, countQuery, args...); err != nil {
		return nil, 0, err
	}
	sortCol := map[string]string{
		"kode_pesanan":        "kode_pesanan",
		"tanggal_pesanan":     "tanggal_pesanan",
		"modal_purchase":      "modal_purchase",
		"amount_sales":        "amount_sales",
		"biaya_mitra_jakmall": "biaya_mitra_jakmall",
		"biaya_administrasi":  "biaya_administrasi",
		"biaya_layanan":       "biaya_layanan",
		"biaya_voucher":       "biaya_voucher",
		"biaya_affiliate":     "biaya_affiliate",
		"profit":              "profit",
		"profit_percent":      "profit_percent",
	}[sortBy]
	if sortCol == "" {
		sortCol = "tanggal_pesanan"
	}
	direction := "ASC"
	if strings.ToLower(dir) == "desc" {
		direction = "DESC"
	}
	args = append(args, limit, offset)
	finalQuery := query + fmt.Sprintf(" ORDER BY %s %s LIMIT $%d OFFSET $%d", sortCol, direction, arg, arg+1)
	var rows []struct {
		KodePesanan       string    `db:"kode_pesanan"`
		TanggalPesanan    time.Time `db:"tanggal_pesanan"`
		ModalPurchase     float64   `db:"modal_purchase"`
		AmountSales       float64   `db:"amount_sales"`
		BiayaMitraJakmall float64   `db:"biaya_mitra_jakmall"`
		BiayaAdministrasi float64   `db:"biaya_administrasi"`
		BiayaLayanan      float64   `db:"biaya_layanan"`
		BiayaVoucher      float64   `db:"biaya_voucher"`
		BiayaAffiliate    float64   `db:"biaya_affiliate"`
		Profit            float64   `db:"profit"`
		ProfitPercent     float64   `db:"profit_percent"`
	}
	if err := r.db.SelectContext(ctx, &rows, finalQuery, args...); err != nil {
		return nil, 0, err
	}
	list := make([]models.SalesProfit, len(rows))
	for i, r := range rows {
		list[i] = models.SalesProfit{
			KodePesanan:       r.KodePesanan,
			TanggalPesanan:    r.TanggalPesanan,
			ModalPurchase:     r.ModalPurchase,
			AmountSales:       r.AmountSales,
			BiayaMitraJakmall: r.BiayaMitraJakmall,
			BiayaAdministrasi: r.BiayaAdministrasi,
			BiayaLayanan:      r.BiayaLayanan,
			BiayaVoucher:      r.BiayaVoucher,
			BiayaAffiliate:    r.BiayaAffiliate,
			Profit:            r.Profit,
			ProfitPercent:     r.ProfitPercent,
		}
	}
	return list, count, nil
}
