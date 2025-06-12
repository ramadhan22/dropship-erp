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
  tanggal_dana_dilepaskan: string;
  total_penerimaan: number;
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
