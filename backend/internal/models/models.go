// File: backend/internal/models/models.go

package models

import "time"

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
	KomisiShopee                                     float64   `db:"komisi_shopee" json:"komisi_shopee"`
	BiayaAdminShopee                                 float64   `db:"biaya_admin_shopee" json:"biaya_admin_shopee"`
	BiayaLayanan                                     float64   `db:"biaya_layanan" json:"biaya_layanan"`
	BiayaLayananEkstra                               float64   `db:"biaya_layanan_ekstra" json:"biaya_layanan_ekstra"`
	BiayaPenyediaPembayaran                          float64   `db:"biaya_penyedia_pembayaran" json:"biaya_penyedia_pembayaran"`
	Asuransi                                         float64   `db:"asuransi" json:"asuransi"`
	TotalBiayaTransaksi                              float64   `db:"total_biaya_transaksi" json:"total_biaya_transaksi"`
	BiayaPengiriman                                  float64   `db:"biaya_pengiriman" json:"biaya_pengiriman"`
	TotalDiskonPengiriman                            float64   `db:"total_diskon_pengiriman" json:"total_diskon_pengiriman"`
	PromoGratisOngkirShopee                          float64   `db:"promo_gratis_ongkir_shopee" json:"promo_gratis_ongkir_shopee"`
	PromoGratisOngkirPenjual                         float64   `db:"promo_gratis_ongkir_penjual" json:"promo_gratis_ongkir_penjual"`
	PromoDiskonShopee                                float64   `db:"promo_diskon_shopee" json:"promo_diskon_shopee"`
	PromoDiskonPenjual                               float64   `db:"promo_diskon_penjual" json:"promo_diskon_penjual"`
	CashbackShopee                                   float64   `db:"cashback_shopee" json:"cashback_shopee"`
	CashbackPenjual                                  float64   `db:"cashback_penjual" json:"cashback_penjual"`
	KoinShopee                                       float64   `db:"koin_shopee" json:"koin_shopee"`
	PotonganLainnya                                  float64   `db:"potongan_lainnya" json:"potongan_lainnya"`
	TotalPenerimaan                                  float64   `db:"total_penerimaan" json:"total_penerimaan"`
	Kompensasi                                       float64   `db:"kompensasi" json:"kompensasi"`
	PromoGratisOngkirDariPenjual                     float64   `db:"promo_gratis_ongkir_dari_penjual" json:"promo_gratis_ongkir_dari_penjual"`
	JasaKirim                                        string    `db:"jasa_kirim" json:"jasa_kirim"`
	NamaKurir                                        string    `db:"nama_kurir" json:"nama_kurir"`
	PengembalianDanaKePembeli                        float64   `db:"pengembalian_dana_ke_pembeli" json:"pengembalian_dana_ke_pembeli"`
	ProRataKoinYangDitukarkanUntukPengembalianBarang float64   `db:"pro_rata_koin_yang_ditukarkan_untuk_pengembalian_barang" json:"pro_rata_koin_yang_ditukarkan_untuk_pengembalian_barang"`
	ProRataVoucherShopeeUntukPengembalianBarang      float64   `db:"pro_rata_voucher_shopee_untuk_pengembalian_barang" json:"pro_rata_voucher_shopee_untuk_pengembalian_barang"`
	ProRatedBankPaymentChannelPromotionForReturns    float64   `db:"pro_rated_bank_payment_channel_promotion_for_returns" json:"pro_rated_bank_payment_channel_promotion_for_returns"`
	ProRatedShopeePaymentChannelPromotionForReturns  float64   `db:"pro_rated_shopee_payment_channel_promotion_for_returns" json:"pro_rated_shopee_payment_channel_promotion_for_returns"`
}

// JournalEntry represents the D7 header table: journal_entries
type JournalEntry struct {
	JournalID    int64     `db:"journal_id" json:"journal_id"`
	EntryDate    time.Time `db:"entry_date" json:"entry_date"`
	Description  *string   `db:"description" json:"description"` // NULLABLE
	SourceType   string    `db:"source_type" json:"source_type"`
	SourceID     string    `db:"source_id" json:"source_id"`
	ShopUsername string    `db:"shop_username" json:"shop_username"`
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
	StoreID        int64  `db:"store_id" json:"store_id"`
	JenisChannelID int64  `db:"jenis_channel_id" json:"jenis_channel_id"`
	NamaToko       string `db:"nama_toko" json:"nama_toko"`
}
