export interface Metric {
  shop_username: string;
  period: string;
  sum_revenue: number;
  sum_cogs: number;
  sum_fees: number;
  net_profit: number;
  ending_cash_balance: number;
}

export interface Account {
  account_id: number;
  account_code: string;
  account_name: string;
  account_type: string;
  parent_id: number | null;
  balance: number;
}

export interface BalanceCategory {
  category: string; // e.g. "Assets"
  accounts: Account[]; // list of accounts in this category
  total: number; // aggregated total
}

export interface JenisChannel {
  jenis_channel_id: number;
  jenis_channel: string;
}

export interface Store {
  store_id: number;
  jenis_channel_id: number;
  nama_toko: string;
}

export interface ShopeeSettled {
  nama_toko: string;
  no_pesanan: string;
  no_pengajuan: string;
  username_pembeli: string;
  waktu_pesanan_dibuat: string;
  metode_pembayaran_pembeli: string;
  tanggal_dana_dilepaskan: string;
  harga_asli_produk: number;
  total_diskon_produk: number;
  jumlah_pengembalian_dana_ke_pembeli: number;
  diskon_produk_dari_shopee: number;
  diskon_voucher_ditanggung_penjual: number;
  cashback_koin_ditanggung_penjual: number;
  ongkir_dibayar_pembeli: number;
  diskon_ongkir_ditanggung_jasa_kirim: number;
  gratis_ongkir_dari_shopee: number;
  ongkir_yang_diteruskan_oleh_shopee_ke_jasa_kirim: number;
  ongkos_kirim_pengembalian_barang: number;
  pengembalian_biaya_kirim: number;
  biaya_komisi_ams: number;
  biaya_administrasi: number;
  biaya_layanan_termasuk_ppn_11: number;
  premi: number;
  biaya_program: number;
  biaya_kartu_kredit: number;
  biaya_kampanye: number;
  bea_masuk_ppn_pph: number;
  total_penghasilan: number;
  kompensasi: number;
  promo_gratis_ongkir_dari_penjual: number;
  jasa_kirim: string;
  nama_kurir: string;
  pengembalian_dana_ke_pembeli: number;
  pro_rata_koin_yang_ditukarkan_untuk_pengembalian_barang: number;
  pro_rata_voucher_shopee_untuk_pengembalian_barang: number;
  pro_rated_bank_payment_channel_promotion_for_returns: number;
  pro_rated_shopee_payment_channel_promotion_for_returns: number;
}

export interface ShopeeSettledSummary {
  harga_asli_produk: number;
  total_diskon_produk: number;
  gmv: number;
  diskon_voucher_ditanggung_penjual: number;
  biaya_administrasi: number;
  biaya_layanan_termasuk_ppn_11: number;
  total_penghasilan: number;
}
export interface DropshipPurchase {
  kode_pesanan: string;
  kode_transaksi: string;
  waktu_pesanan_terbuat: string;
  status_pesanan_terakhir: string;
  biaya_lainnya: number;
  biaya_mitra_jakmall: number;
  total_transaksi: number;
  dibuat_oleh: string;
  jenis_channel: string;
  nama_toko: string;
  kode_invoice_channel: string;
  gudang_pengiriman: string;
  jenis_ekspedisi: string;
  cashless: string;
  nomor_resi: string;
  waktu_pengiriman: string;
  provinsi: string;
  kota: string;
}

export interface DropshipPurchaseDetail {
  id: number;
  kode_pesanan: string;
  sku: string;
  nama_produk: string;
  harga_produk: number;
  qty: number;
  total_harga_produk: number;
  harga_produk_channel: number;
  total_harga_produk_channel: number;
  potensi_keuntungan: number;
}

export interface Expense {
  id: string;
  date: string;
  description: string;
  amount: number;
  account_id: number;
  created_at: string;
}

export interface JournalEntry {
  journal_id: number;
  entry_date: string;
  description: string | null;
  source_type: string;
  source_id: string;
  shop_username: string;
  created_at: string;
}

export interface JournalLine {
  line_id: number;
  journal_id: number;
  account_id: number;
  is_debit: boolean;
  amount: number;
  memo: string | null;
}

export interface JournalLineDetail extends JournalLine {
  account_name: string;
}

export interface ReconciledTransaction {
  id: number;
  shop_username: string;
  dropship_id: string | null;
  shopee_id: string | null;
  status: string;
  matched_at: string;
}

export interface ReconcileCandidate {
  kode_pesanan: string;
  kode_invoice_channel: string;
  nama_toko: string;
  status_pesanan_terakhir: string;
  no_pesanan: string | null;
}

export interface ProductSales {
  nama_produk: string;
  total_qty: number;
  total_value: number;
}

export interface ShopeeAffiliateSale {
  kode_pesanan: string;
  status_pesanan: string;
  nama_affiliate: string;
  username_affiliate: string;
  waktu_pesanan: string;
  nilai_pembelian: number;
  estimasi_komisi_affiliate_per_pesanan: number;
}
