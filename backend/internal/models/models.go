// File: backend/internal/models/models.go

package models

import (
	"time"

	"github.com/lib/pq"
)

// Account represents the D6 table: accounts
type Account struct {
	AccountID   int64     `db:"account_id" json:"account_id"`
	AccountCode string    `db:"account_code" json:"account_code"`
	AccountName string    `db:"account_name" json:"account_name"`
	AccountType string    `db:"account_type" json:"account_type"`
	ParentID    *int64    `db:"parent_id" json:"parent_id"` // NULLABLE
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// DropshipPurchase represents the header table: dropship_purchases
type DropshipPurchase struct {
	KodePesanan           string    `db:"kode_pesanan" json:"kode_pesanan"`
	KodeTransaksi         string    `db:"kode_transaksi" json:"kode_transaksi"`
	WaktuPesananTerbuat   time.Time `db:"waktu_pesanan_terbuat" json:"waktu_pesanan_terbuat"`
	StatusPesananTerakhir string    `db:"status_pesanan_terakhir" json:"status_pesanan_terakhir"`
	BiayaLainnya          float64   `db:"biaya_lainnya" json:"biaya_lainnya"`
	BiayaMitraJakmall     float64   `db:"biaya_mitra_jakmall" json:"biaya_mitra_jakmall"`
	TotalTransaksi        float64   `db:"total_transaksi" json:"total_transaksi"`
	DibuatOleh            string    `db:"dibuat_oleh" json:"dibuat_oleh"`
	JenisChannel          string    `db:"jenis_channel" json:"jenis_channel"`
	NamaToko              string    `db:"nama_toko" json:"nama_toko"`
	KodeInvoiceChannel    string    `db:"kode_invoice_channel" json:"kode_invoice_channel"`
	GudangPengiriman      string    `db:"gudang_pengiriman" json:"gudang_pengiriman"`
	JenisEkspedisi        string    `db:"jenis_ekspedisi" json:"jenis_ekspedisi"`
	Cashless              string    `db:"cashless" json:"cashless"`
	NomorResi             string    `db:"nomor_resi" json:"nomor_resi"`
	WaktuPengiriman       time.Time `db:"waktu_pengiriman" json:"waktu_pengiriman"`
	Provinsi              string    `db:"provinsi" json:"provinsi"`
	Kota                  string    `db:"kota" json:"kota"`
}

// DropshipPurchaseDetail represents the detail table: dropship_purchase_details
type DropshipPurchaseDetail struct {
	ID                      int64   `db:"id" json:"id"`
	KodePesanan             string  `db:"kode_pesanan" json:"kode_pesanan"`
	SKU                     string  `db:"sku" json:"sku"`
	NamaProduk              string  `db:"nama_produk" json:"nama_produk"`
	HargaProduk             float64 `db:"harga_produk" json:"harga_produk"`
	Qty                     int     `db:"qty" json:"qty"`
	TotalHargaProduk        float64 `db:"total_harga_produk" json:"total_harga_produk"`
	HargaProdukChannel      float64 `db:"harga_produk_channel" json:"harga_produk_channel"`
	TotalHargaProdukChannel float64 `db:"total_harga_produk_channel" json:"total_harga_produk_channel"`
	PotensiKeuntungan       float64 `db:"potensi_keuntungan" json:"potensi_keuntungan"`
}

// ShopeeSettledOrder represents the D1 table: shopee_settled_orders
type ShopeeSettledOrder struct {
	ID              int64     `db:"id" json:"id"`
	OrderID         string    `db:"order_id" json:"order_id"`
	NetIncome       float64   `db:"net_income" json:"net_income"`
	ServiceFee      float64   `db:"service_fee" json:"service_fee"`
	CampaignFee     float64   `db:"campaign_fee" json:"campaign_fee"`
	CreditCardFee   float64   `db:"credit_card_fee" json:"credit_card_fee"`
	ShippingSubsidy float64   `db:"shipping_subsidy" json:"shipping_subsidy"`
	TaxImportFee    float64   `db:"tax_and_import_fee" json:"tax_and_import_fee"`
	SettledDate     time.Time `db:"settled_date" json:"settled_date"`
	SellerUsername  string    `db:"seller_username" json:"seller_username"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time `db:"updated_at" json:"updated_at"`
}

// ShopeeSettled represents rows of the shopee_settled table.
type ShopeeSettled struct {
	NamaToko                                         string    `db:"nama_toko" json:"nama_toko"`
	NoPesanan                                        string    `db:"no_pesanan" json:"no_pesanan"`
	NoPengajuan                                      string    `db:"no_pengajuan" json:"no_pengajuan"`
	UsernamePembeli                                  string    `db:"username_pembeli" json:"username_pembeli"`
	WaktuPesananDibuat                               time.Time `db:"waktu_pesanan_dibuat" json:"waktu_pesanan_dibuat"`
	MetodePembayaranPembeli                          string    `db:"metode_pembayaran_pembeli" json:"metode_pembayaran_pembeli"`
	TanggalDanaDilepaskan                            time.Time `db:"tanggal_dana_dilepaskan" json:"tanggal_dana_dilepaskan"`
	HargaAsliProduk                                  float64   `db:"harga_asli_produk" json:"harga_asli_produk"`
	TotalDiskonProduk                                float64   `db:"total_diskon_produk" json:"total_diskon_produk"`
	JumlahPengembalianDanaKePembeli                  float64   `db:"jumlah_pengembalian_dana_ke_pembeli" json:"jumlah_pengembalian_dana_ke_pembeli"`
	KomisiShopee                                     float64   `db:"diskon_produk_dari_shopee" json:"diskon_produk_dari_shopee"`
	BiayaAdminShopee                                 float64   `db:"diskon_voucher_ditanggung_penjual" json:"diskon_voucher_ditanggung_penjual"`
	BiayaLayanan                                     float64   `db:"cashback_koin_ditanggung_penjual" json:"cashback_koin_ditanggung_penjual"`
	BiayaLayananEkstra                               float64   `db:"ongkir_dibayar_pembeli" json:"ongkir_dibayar_pembeli"`
	BiayaPenyediaPembayaran                          float64   `db:"diskon_ongkir_ditanggung_jasa_kirim" json:"diskon_ongkir_ditanggung_jasa_kirim"`
	Asuransi                                         float64   `db:"gratis_ongkir_dari_shopee" json:"gratis_ongkir_dari_shopee"`
	TotalBiayaTransaksi                              float64   `db:"ongkir_yang_diteruskan_oleh_shopee_ke_jasa_kirim" json:"ongkir_yang_diteruskan_oleh_shopee_ke_jasa_kirim"`
	BiayaPengiriman                                  float64   `db:"ongkos_kirim_pengembalian_barang" json:"ongkos_kirim_pengembalian_barang"`
	TotalDiskonPengiriman                            float64   `db:"pengembalian_biaya_kirim" json:"pengembalian_biaya_kirim"`
	PromoGratisOngkirShopee                          float64   `db:"biaya_komisi_ams" json:"biaya_komisi_ams"`
	PromoGratisOngkirPenjual                         float64   `db:"biaya_administrasi" json:"biaya_administrasi"`
	PromoDiskonShopee                                float64   `db:"biaya_layanan_termasuk_ppn_11" json:"biaya_layanan_termasuk_ppn_11"`
	PromoDiskonPenjual                               float64   `db:"premi" json:"premi"`
	CashbackShopee                                   float64   `db:"biaya_program" json:"biaya_program"`
	CashbackPenjual                                  float64   `db:"biaya_kartu_kredit" json:"biaya_kartu_kredit"`
	BiayaTransaksi                                   float64   `db:"biaya_transaksi" json:"biaya_transaksi"`
	KoinShopee                                       float64   `db:"biaya_kampanye" json:"biaya_kampanye"`
	PotonganLainnya                                  float64   `db:"bea_masuk_ppn_pph" json:"bea_masuk_ppn_pph"`
	TotalPenerimaan                                  float64   `db:"total_penghasilan" json:"total_penghasilan"`
	Kompensasi                                       float64   `db:"kompensasi" json:"kompensasi"`
	PromoGratisOngkirDariPenjual                     float64   `db:"promo_gratis_ongkir_dari_penjual" json:"promo_gratis_ongkir_dari_penjual"`
	JasaKirim                                        string    `db:"jasa_kirim" json:"jasa_kirim"`
	NamaKurir                                        string    `db:"nama_kurir" json:"nama_kurir"`
	PengembalianDanaKePembeli                        float64   `db:"pengembalian_dana_ke_pembeli" json:"pengembalian_dana_ke_pembeli"`
	ProRataKoinYangDitukarkanUntukPengembalianBarang float64   `db:"pro_rata_koin_yang_ditukarkan_untuk_pengembalian_barang" json:"pro_rata_koin_yang_ditukarkan_untuk_pengembalian_barang"`
	ProRataVoucherShopeeUntukPengembalianBarang      float64   `db:"pro_rata_voucher_shopee_untuk_pengembalian_barang" json:"pro_rata_voucher_shopee_untuk_pengembalian_barang"`
	ProRatedBankPaymentChannelPromotionForReturns    float64   `db:"pro_rated_bank_payment_channel_promotion_for_returns" json:"pro_rated_bank_payment_channel_promotion_for_returns"`
	ProRatedShopeePaymentChannelPromotionForReturns  float64   `db:"pro_rated_shopee_payment_channel_promotion_for_returns" json:"pro_rated_shopee_payment_channel_promotion_for_returns"`
	IsDataMismatch                                   bool      `db:"is_data_mismatch" json:"is_data_mismatch"`
	IsSettledConfirmed                               bool      `db:"is_settled_confirmed" json:"is_settled_confirmed"`
}

// JournalEntry represents the D7 header table: journal_entries
type JournalEntry struct {
	JournalID    int64     `db:"journal_id" json:"journal_id"`
	EntryDate    time.Time `db:"entry_date" json:"entry_date"`
	Description  *string   `db:"description" json:"description"` // NULLABLE
	SourceType   string    `db:"source_type" json:"source_type"`
	SourceID     string    `db:"source_id" json:"source_id"`
	ShopUsername string    `db:"shop_username" json:"shop_username"`
	Store        string    `db:"store" json:"store"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

// JournalLine represents the D7 detail table: journal_lines
type JournalLine struct {
	LineID    int64   `db:"line_id" json:"line_id"`
	JournalID int64   `db:"journal_id" json:"journal_id"` // FK → journal_entries(journal_id)
	AccountID int64   `db:"account_id" json:"account_id"` // FK → accounts(account_id)
	IsDebit   bool    `db:"is_debit" json:"is_debit"`
	Amount    float64 `db:"amount" json:"amount"`
	Memo      *string `db:"memo" json:"memo"` // NULLABLE
}

// ReconciledTransaction represents the D3 table: reconciled_transactions
type ReconciledTransaction struct {
	ID           int64     `db:"id" json:"id"`
	ShopUsername string    `db:"shop_username" json:"shop_username"`
	DropshipID   *string   `db:"dropship_id" json:"dropship_id"` // NULLABLE
	ShopeeID     *string   `db:"shopee_id" json:"shopee_id"`     // NULLABLE
	Status       string    `db:"status" json:"status"`           // e.g., "matched", "unmatched"
	MatchedAt    time.Time `db:"matched_at" json:"matched_at"`
}

// ReconcileCandidate is used by the dashboard to display purchases that
// require attention. If NoPesanan is empty, the purchase has no matching
// record in shopee_settled.
type ReconcileCandidate struct {
	KodePesanan           string  `db:"kode_pesanan" json:"kode_pesanan"`
	KodeInvoiceChannel    string  `db:"kode_invoice_channel" json:"kode_invoice_channel"`
	NamaToko              string  `db:"nama_toko" json:"nama_toko"`
	StatusPesananTerakhir string  `db:"status_pesanan_terakhir" json:"status_pesanan_terakhir"`
	NoPesanan             *string `db:"no_pesanan" json:"no_pesanan"`
	ShopeeOrderStatus     string  `json:"shopee_order_status"`
}

// CachedMetric represents the D5 table: cached_metrics
type CachedMetric struct {
	ID                int64     `db:"id" json:"id"`
	ShopUsername      string    `db:"shop_username" json:"shop_username"`
	Period            string    `db:"period" json:"period"` // e.g., "2025-05"
	SumRevenue        float64   `db:"sum_revenue" json:"sum_revenue"`
	SumCOGS           float64   `db:"sum_cogs" json:"sum_cogs"`
	SumFees           float64   `db:"sum_fees" json:"sum_fees"`
	NetProfit         float64   `db:"net_profit" json:"net_profit"`
	EndingCashBalance float64   `db:"ending_cash_balance" json:"ending_cash_balance"`
	UpdatedAt         time.Time `db:"updated_at" json:"updated_at"`
}

// JenisChannel represents e-commerce channel types such as Shopee or Tokopedia.
type JenisChannel struct {
	JenisChannelID int64  `db:"jenis_channel_id" json:"jenis_channel_id"`
	JenisChannel   string `db:"jenis_channel" json:"jenis_channel"`
}

// Store represents a store/shop under a jenis channel.
type Store struct {
	StoreID        int64      `db:"store_id" json:"store_id"`
	JenisChannelID int64      `db:"jenis_channel_id" json:"jenis_channel_id"`
	NamaToko       string     `db:"nama_toko" json:"nama_toko"`
	CodeID         *string    `db:"code_id" json:"code_id"`
	ShopID         *string    `db:"shop_id" json:"shop_id"`
	AccessToken    *string    `db:"access_token" json:"-"`
	RefreshToken   *string    `db:"refresh_token" json:"-"`
	ExpireIn       *int       `db:"expire_in" json:"-"`
	RequestID      *string    `db:"request_id" json:"-"`
	LastUpdated    *time.Time `db:"last_updated" json:"-"`
}

// StoreWithChannel joins a store with its channel name.
type StoreWithChannel struct {
	Store
	JenisChannel string `db:"jenis_channel" json:"jenis_channel"`
}

// ProductSales represents aggregated sales for a product.
type ProductSales struct {
	NamaProduk string  `db:"nama_produk" json:"nama_produk"`
	TotalQty   int     `db:"total_qty" json:"total_qty"`
	TotalValue float64 `db:"total_value" json:"total_value"`
}

// ShopeeSummary aggregates numeric columns from shopee_settled for summary views.
type ShopeeSummary struct {
	HargaAsliProduk                float64 `db:"harga_asli_produk" json:"harga_asli_produk"`
	TotalDiskonProduk              float64 `db:"total_diskon_produk" json:"total_diskon_produk"`
	GMV                            float64 `db:"-" json:"gmv"`
	DiskonVoucherDitanggungPenjual float64 `db:"diskon_voucher_ditanggung_penjual" json:"diskon_voucher_ditanggung_penjual"`
	BiayaAdministrasi              float64 `db:"biaya_administrasi" json:"biaya_administrasi"`
	BiayaLayanan                   float64 `db:"biaya_layanan_termasuk_ppn_11" json:"biaya_layanan_termasuk_ppn_11"`
	TotalPenghasilan               float64 `db:"total_penghasilan" json:"total_penghasilan"`
}

// ShopeeAffiliateSale represents rows from shopee_affiliate_sales table.
type ShopeeAffiliateSale struct {
	NamaToko                           string    `db:"nama_toko" json:"nama_toko"`
	KodePesanan                        string    `db:"kode_pesanan" json:"kode_pesanan"`
	StatusPesanan                      string    `db:"status_pesanan" json:"status_pesanan"`
	StatusTerverifikasi                string    `db:"status_terverifikasi" json:"status_terverifikasi"`
	WaktuPesanan                       time.Time `db:"waktu_pesanan" json:"waktu_pesanan"`
	WaktuPesananSelesai                time.Time `db:"waktu_pesanan_selesai" json:"waktu_pesanan_selesai"`
	WaktuPesananTerverifikasi          time.Time `db:"waktu_pesanan_terverifikasi" json:"waktu_pesanan_terverifikasi"`
	KodeProduk                         string    `db:"kode_produk" json:"kode_produk"`
	NamaProduk                         string    `db:"nama_produk" json:"nama_produk"`
	IDModel                            string    `db:"id_model" json:"id_model"`
	L1KategoriGlobal                   string    `db:"l1_kategori_global" json:"l1_kategori_global"`
	L2KategoriGlobal                   string    `db:"l2_kategori_global" json:"l2_kategori_global"`
	L3KategoriGlobal                   string    `db:"l3_kategori_global" json:"l3_kategori_global"`
	KodePromo                          string    `db:"kode_promo" json:"kode_promo"`
	Harga                              float64   `db:"harga" json:"harga"`
	Jumlah                             int       `db:"jumlah" json:"jumlah"`
	NamaAffiliate                      string    `db:"nama_affiliate" json:"nama_affiliate"`
	UsernameAffiliate                  string    `db:"username_affiliate" json:"username_affiliate"`
	MCNTerhubung                       string    `db:"mcn_terhubung" json:"mcn_terhubung"`
	IDKomisiPesanan                    string    `db:"id_komisi_pesanan" json:"id_komisi_pesanan"`
	PartnerPromo                       string    `db:"partner_promo" json:"partner_promo"`
	JenisPromo                         string    `db:"jenis_promo" json:"jenis_promo"`
	NilaiPembelian                     float64   `db:"nilai_pembelian" json:"nilai_pembelian"`
	JumlahPengembalian                 float64   `db:"jumlah_pengembalian" json:"jumlah_pengembalian"`
	TipePesanan                        string    `db:"tipe_pesanan" json:"tipe_pesanan"`
	EstimasiKomisiPerProduk            float64   `db:"estimasi_komisi_per_produk" json:"estimasi_komisi_per_produk"`
	EstimasiKomisiAffiliatePerProduk   float64   `db:"estimasi_komisi_affiliate_per_produk" json:"estimasi_komisi_affiliate_per_produk"`
	PersentaseKomisiAffiliatePerProduk float64   `db:"persentase_komisi_affiliate_per_produk" json:"persentase_komisi_affiliate_per_produk"`
	EstimasiKomisiMCNPerProduk         float64   `db:"estimasi_komisi_mcn_per_produk" json:"estimasi_komisi_mcn_per_produk"`
	PersentaseKomisiMCNPerProduk       float64   `db:"persentase_komisi_mcn_per_produk" json:"persentase_komisi_mcn_per_produk"`
	EstimasiKomisiPerPesanan           float64   `db:"estimasi_komisi_per_pesanan" json:"estimasi_komisi_per_pesanan"`
	EstimasiKomisiAffiliatePerPesanan  float64   `db:"estimasi_komisi_affiliate_per_pesanan" json:"estimasi_komisi_affiliate_per_pesanan"`
	EstimasiKomisiMCNPerPesanan        float64   `db:"estimasi_komisi_mcn_per_pesanan" json:"estimasi_komisi_mcn_per_pesanan"`
	CatatanProduk                      string    `db:"catatan_produk" json:"catatan_produk"`
	Platform                           string    `db:"platform" json:"platform"`
	TingkatKomisi                      float64   `db:"tingkat_komisi" json:"tingkat_komisi"`
	Pengeluaran                        float64   `db:"pengeluaran" json:"pengeluaran"`
	StatusPemotongan                   string    `db:"status_pemotongan" json:"status_pemotongan"`
	MetodePemotongan                   string    `db:"metode_pemotongan" json:"metode_pemotongan"`
	WaktuPemotongan                    time.Time `db:"waktu_pemotongan" json:"waktu_pemotongan"`
}

// ShopeeAffiliateSummary aggregates purchase and commission values.
type ShopeeAffiliateSummary struct {
	TotalNilaiPembelian  float64 `db:"total_nilai_pembelian" json:"total_nilai_pembelian"`
	TotalKomisiAffiliate float64 `db:"total_komisi_affiliate" json:"total_komisi_affiliate"`
}

// SalesProfit represents sales along with cost and fee breakdowns.
type SalesProfit struct {
	KodePesanan       string    `db:"kode_pesanan" json:"kode_pesanan"`
	TanggalPesanan    time.Time `db:"tanggal_pesanan" json:"tanggal_pesanan"`
	ModalPurchase     float64   `db:"modal_purchase" json:"modal_purchase"`
	AmountSales       float64   `db:"amount_sales" json:"amount_sales"`
	BiayaMitraJakmall float64   `db:"biaya_mitra_jakmall" json:"biaya_mitra_jakmall"`
	BiayaAdministrasi float64   `db:"biaya_administrasi" json:"biaya_administrasi"`
	BiayaLayanan      float64   `db:"biaya_layanan" json:"biaya_layanan"`
	BiayaVoucher      float64   `db:"biaya_voucher" json:"biaya_voucher"`
	BiayaTransaksi    float64   `db:"biaya_transaksi" json:"biaya_transaksi"`
	DiskonOngkir      float64   `db:"diskon_ongkir" json:"diskon_ongkir"`
	BiayaAffiliate    float64   `db:"biaya_affiliate" json:"biaya_affiliate"`
	BiayaRefund       float64   `db:"biaya_refund" json:"biaya_refund"`
	SelisihOngkir     float64   `db:"selisih_ongkir" json:"selisih_ongkir"`
	AdjustmentIncome  float64   `db:"adjustment_income" json:"adjustment_income"`
	Discount          float64   `db:"discount" json:"discount"`
	Profit            float64   `db:"profit" json:"profit"`
	ProfitPercent     float64   `db:"profit_percent" json:"profit_percent"`
}

// Withdrawal represents cash out from Shopee balance to bank.
type Withdrawal struct {
	ID        int64     `db:"id" json:"id"`
	Store     string    `db:"store" json:"store"`
	Date      time.Time `db:"date" json:"date"`
	Amount    float64   `db:"amount" json:"amount"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// ShopeeAdjustment records adjustment entries from Shopee income reports.
type ShopeeAdjustment struct {
	ID                 int64     `db:"id" json:"id"`
	NamaToko           string    `db:"nama_toko" json:"nama_toko"`
	TanggalPenyesuaian time.Time `db:"tanggal_penyesuaian" json:"tanggal_penyesuaian"`
	TipePenyesuaian    string    `db:"tipe_penyesuaian" json:"tipe_penyesuaian"`
	AlasanPenyesuaian  string    `db:"alasan_penyesuaian" json:"alasan_penyesuaian"`
	BiayaPenyesuaian   float64   `db:"biaya_penyesuaian" json:"biaya_penyesuaian"`
	NoPesanan          string    `db:"no_pesanan" json:"no_pesanan"`
	CreatedAt          time.Time `db:"created_at" json:"created_at"`
}

// ShopeeOrderDetailRow stores key fields from Shopee order detail.
type ShopeeOrderDetailRow struct {
	OrderSN                    string     `db:"order_sn" json:"order_sn"`
	NamaToko                   string     `db:"nama_toko" json:"nama_toko"`
	Status                     *string    `db:"status" json:"status,omitempty"`
	OrderStatus                *string    `db:"order_status" json:"order_status,omitempty"`
	CheckoutTime               *time.Time `db:"checkout_time" json:"checkout_time,omitempty"`
	UpdateTime                 *time.Time `db:"update_time" json:"update_time,omitempty"`
	PayTime                    *time.Time `db:"pay_time" json:"pay_time,omitempty"`
	TotalAmount                *float64   `db:"total_amount" json:"total_amount,omitempty"`
	Currency                   *string    `db:"currency" json:"currency,omitempty"`
	ActualShippingFeeConfirmed *bool      `db:"actual_shipping_fee_confirmed" json:"actual_shipping_fee_confirmed,omitempty"`
	BuyerCancelReason          *string    `db:"buyer_cancel_reason" json:"buyer_cancel_reason,omitempty"`
	BuyerCPFID                 *string    `db:"buyer_cpf_id" json:"buyer_cpf_id,omitempty"`
	BuyerUserID                *int64     `db:"buyer_user_id" json:"buyer_user_id,omitempty"`
	BuyerUsername              *string    `db:"buyer_username" json:"buyer_username,omitempty"`
	CancelBy                   *string    `db:"cancel_by" json:"cancel_by,omitempty"`
	CancelReason               *string    `db:"cancel_reason" json:"cancel_reason,omitempty"`
	COD                        *bool      `db:"cod" json:"cod,omitempty"`
	CreateTime                 *time.Time `db:"create_time" json:"create_time,omitempty"`
	DaysToShip                 *int       `db:"days_to_ship" json:"days_to_ship,omitempty"`
	Dropshipper                *string    `db:"dropshipper" json:"dropshipper,omitempty"`
	DropshipperPhone           *string    `db:"dropshipper_phone" json:"dropshipper_phone,omitempty"`
	EstimatedShippingFee       *float64   `db:"estimated_shipping_fee" json:"estimated_shipping_fee,omitempty"`
	FulfillmentFlag            *string    `db:"fulfillment_flag" json:"fulfillment_flag,omitempty"`
	GoodsToDeclare             *bool      `db:"goods_to_declare" json:"goods_to_declare,omitempty"`
	MessageToSeller            *string    `db:"message_to_seller" json:"message_to_seller,omitempty"`
	Note                       *string    `db:"note" json:"note,omitempty"`
	NoteUpdateTime             *time.Time `db:"note_update_time" json:"note_update_time,omitempty"`
	PickupDoneTime             *time.Time `db:"pickup_done_time" json:"pickup_done_time,omitempty"`
	Region                     *string    `db:"region" json:"region,omitempty"`
	ReverseShippingFee         *float64   `db:"reverse_shipping_fee" json:"reverse_shipping_fee,omitempty"`
	ShipByDate                 *time.Time `db:"ship_by_date" json:"ship_by_date,omitempty"`
	ShippingCarrier            *string    `db:"shipping_carrier" json:"shipping_carrier,omitempty"`
	SplitUp                    *bool      `db:"split_up" json:"split_up,omitempty"`
	PaymentMethod              *string    `db:"payment_method" json:"payment_method,omitempty"`
	RecipientName              *string    `db:"recipient_name" json:"recipient_name,omitempty"`
	RecipientPhone             *string    `db:"recipient_phone" json:"recipient_phone,omitempty"`
	RecipientFullAddress       *string    `db:"recipient_full_address" json:"recipient_full_address,omitempty"`
	RecipientCity              *string    `db:"recipient_city" json:"recipient_city,omitempty"`
	RecipientDistrict          *string    `db:"recipient_district" json:"recipient_district,omitempty"`
	RecipientState             *string    `db:"recipient_state" json:"recipient_state,omitempty"`
	RecipientTown              *string    `db:"recipient_town" json:"recipient_town,omitempty"`
	RecipientZipcode           *string    `db:"recipient_zipcode" json:"recipient_zipcode,omitempty"`
	CreatedAt                  time.Time  `db:"created_at" json:"created_at"`
}

// ShopeeOrderItemRow stores individual item fields from an order detail.
type ShopeeOrderItemRow struct {
	ID                     int64          `db:"id" json:"id"`
	OrderSN                string         `db:"order_sn" json:"order_sn"`
	OrderItemID            *int64         `db:"order_item_id" json:"order_item_id,omitempty"`
	ItemName               *string        `db:"item_name" json:"item_name,omitempty"`
	ModelOriginalPrice     *float64       `db:"model_original_price" json:"model_original_price,omitempty"`
	ModelQuantityPurchased *int           `db:"model_quantity_purchased" json:"model_quantity_purchased,omitempty"`
	ItemID                 *int64         `db:"item_id" json:"item_id,omitempty"`
	ItemSKU                *string        `db:"item_sku" json:"item_sku,omitempty"`
	ModelID                *int64         `db:"model_id" json:"model_id,omitempty"`
	ModelName              *string        `db:"model_name" json:"model_name,omitempty"`
	ModelSKU               *string        `db:"model_sku" json:"model_sku,omitempty"`
	ModelDiscountedPrice   *float64       `db:"model_discounted_price" json:"model_discounted_price,omitempty"`
	Weight                 *float64       `db:"weight" json:"weight,omitempty"`
	PromotionID            *int64         `db:"promotion_id" json:"promotion_id,omitempty"`
	PromotionType          *string        `db:"promotion_type" json:"promotion_type,omitempty"`
	PromotionGroupID       *int64         `db:"promotion_group_id" json:"promotion_group_id,omitempty"`
	AddOnDeal              *bool          `db:"add_on_deal" json:"add_on_deal,omitempty"`
	AddOnDealID            *int64         `db:"add_on_deal_id" json:"add_on_deal_id,omitempty"`
	MainItem               *bool          `db:"main_item" json:"main_item,omitempty"`
	IsB2COwnedItem         *bool          `db:"is_b2c_owned_item" json:"is_b2c_owned_item,omitempty"`
	IsPrescriptionItem     *bool          `db:"is_prescription_item" json:"is_prescription_item,omitempty"`
	Wholesale              *bool          `db:"wholesale" json:"wholesale,omitempty"`
	ProductLocationID      pq.StringArray `db:"product_location_id" json:"product_location_id,omitempty"`
	ImageURL               *string        `db:"image_url" json:"image_url,omitempty"`
}

// ShopeeOrderPackageRow stores shipping package details associated with an order.
type ShopeeOrderPackageRow struct {
	ID                         int64   `db:"id" json:"id"`
	OrderSN                    string  `db:"order_sn" json:"order_sn"`
	PackageNumber              *string `db:"package_number" json:"package_number,omitempty"`
	LogisticsStatus            *string `db:"logistics_status" json:"logistics_status,omitempty"`
	ShippingCarrier            *string `db:"shipping_carrier" json:"shipping_carrier,omitempty"`
	LogisticsChannelID         *int64  `db:"logistics_channel_id" json:"logistics_channel_id,omitempty"`
	ParcelChargeableWeightGram *int    `db:"parcel_chargeable_weight_gram" json:"parcel_chargeable_weight_gram,omitempty"`
	AllowSelfDesignAWB         *bool   `db:"allow_self_design_awb" json:"allow_self_design_awb,omitempty"`
	SortingGroup               *string `db:"sorting_group" json:"sorting_group,omitempty"`
	GroupShipmentID            *string `db:"group_shipment_id" json:"group_shipment_id,omitempty"`
}

// ReconcileBatchInfo contains information about created reconcile batches
type ReconcileBatchInfo struct {
	BatchCount        int `json:"batch_count"`
	TotalTransactions int `json:"total_transactions"`
}
