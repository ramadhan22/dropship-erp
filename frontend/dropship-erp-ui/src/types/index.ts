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
  category: string;       // e.g. "Assets"
  accounts: Account[];    // list of accounts in this category
  total: number;          // aggregated total
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