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

export interface KasAccountBalance {
  asset_id: number;
  account_id: number;
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
  code_id: string | null;
  shop_id: string | null;
}

export interface StoreWithChannel extends Store {
  jenis_channel: string;
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
  is_data_mismatch: boolean;
  is_settled_confirmed: boolean;
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

export interface ExpenseLine {
  line_id: number;
  expense_id: string;
  account_id: number;
  amount: number;
}

export interface Expense {
  id: string;
  date: string;
  description: string;
  amount: number;
  asset_account_id: number;
  created_at: string;
  lines: ExpenseLine[];
}

export interface AdInvoice {
  invoice_no: string;
  username: string;
  store: string;
  invoice_date: string;
  total: number;
  created_at: string;
}

export interface Withdrawal {
  id: number;
  store: string;
  date: string;
  amount: number;
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

export interface JournalEntryWithLines {
  entry: JournalEntry;
  lines: JournalLineDetail[];
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
  shopee_order_status: string;
}

export interface ShopeeOrderDetail {
  order_sn?: string;
  status?: string;
  checkout_time?: number;
  update_time?: number;
  order_status?: string;
  buyer_username?: string;
  payment_method?: string;
  total_amount?: number;
  currency?: string;
  item_list?: {
    order_item_id: number;
    item_name: string;
    model_quantity_purchased: number;
  }[];
  recipient_address?: {
    name: string;
    full_address: string;
  };
  [key: string]: any;
}

export interface ShopeeEscrowDetail {
  buyer_payment_info?: {
    buyer_paid_extended_warranty?: number;
    buyer_paid_installation_fee?: number;
    buyer_payment_method?: string;
    buyer_service_fee?: number;
    buyer_tax_amount?: number;
    buyer_total_amount?: number;
    credit_card_promotion?: number;
    footwear_tax?: number;
    icms_tax_amount?: number;
    import_processing_charge?: number;
    import_tax_amount?: number;
    initial_buyer_txn_fee?: number;
    insurance_premium?: number;
    iof_tax_amount?: number;
    is_paid_by_credit_card?: boolean;
    merchant_subtotal?: number;
    seller_voucher?: number;
    shipping_fee?: number;
    shipping_fee_sst_amount?: number;
    shopee_coins_redeemed?: number;
    shopee_voucher?: number;
    trade_in_bonus?: number;
    trade_in_discount?: number;
  };
  buyer_user_name?: string;
  order_income?: {
    actual_installation_fee?: number;
    actual_shipping_fee?: number;
    buyer_paid_extended_warranty?: number;
    buyer_paid_shipping_fee?: number;
    buyer_payment_method?: string;
    buyer_total_amount?: number;
    buyer_transaction_fee?: number;
    campaign_fee?: number;
    coins?: number;
    commission_fee?: number;
    cost_of_goods_sold?: number;
    credit_card_promotion?: number;
    credit_card_transaction_fee?: number;
    cross_border_tax?: number;
    delivery_seller_protection_fee_premium_amount?: number;
    drc_adjustable_refund?: number;
    escrow_amount?: number;
    escrow_amount_after_adjustment?: number;
    escrow_import_tax?: number;
    escrow_tax?: number;
    estimated_shipping_fee?: number;
    final_escrow_product_gst?: number;
    final_escrow_shipping_gst?: number;
    final_product_protection?: number;
    final_product_vat_tax?: number;
    final_return_to_seller_shipping_fee?: number;
    final_shipping_fee?: number;
    final_shipping_vat_tax?: number;
    fsf_seller_protection_fee_claim_amount?: number;
    installation_fee_paid_by_buyer?: number;
    instalment_plan?: string;
    items?: {
      activity_id?: number;
      activity_type?: string;
      ams_commission_fee?: number;
      buyer_paid_extended_warranty?: number;
      discount_from_coin?: number;
      discount_from_voucher_seller?: number;
      discount_from_voucher_shopee?: number;
      discounted_price?: number;
      installation_fee_paid_by_buyer?: number;
      is_b2c_shop_item?: boolean;
      is_main_item?: boolean;
      item_id?: number;
      item_name?: string;
      item_sku?: string;
      model_id?: number;
      model_name?: string;
      model_sku?: string;
      original_price?: number;
      quantity_purchased?: number;
      seller_discount?: number;
      selling_price?: number;
      shopee_discount?: number;
    }[];
    order_ams_commission_fee?: number;
    order_chargeable_weight?: number;
    order_discounted_price?: number;
    order_original_price?: number;
    order_seller_discount?: number;
    order_selling_price?: number;
    original_cost_of_goods_sold?: number;
    original_price?: number;
    original_shopee_discount?: number;
    overseas_return_service_fee?: number;
    payment_promotion?: number;
    prorated_coins_value_offset_return_items?: number;
    prorated_payment_channel_promo_bank_offset_return_items?: number;
    prorated_payment_channel_promo_shopee_offset_return_items?: number;
    prorated_seller_voucher_offset_return_items?: number;
    prorated_shopee_voucher_offset_return_items?: number;
    reverse_shipping_fee?: number;
    reverse_shipping_fee_sst?: number;
    rsf_seller_protection_fee_claim_amount?: number;
    sales_tax_on_lvg?: number;
    seller_coin_cash_back?: number;
    seller_discount?: number;
    seller_lost_compensation?: number;
    seller_return_refund?: number;
    seller_shipping_discount?: number;
    seller_transaction_fee?: number;
    seller_voucher_code?: string[];
    service_fee?: number;
    shipping_fee_discount_from_3pl?: number;
    shipping_fee_sst?: number;
    shipping_seller_protection_fee_amount?: number;
    shopee_discount?: number;
    shopee_shipping_rebate?: number;
    tenure_info_list?: { instalment_plan: string }[];
    total_adjustment_amount?: number;
    trade_in_bonus_by_seller?: number;
    vat_on_imported_goods?: number;
    voucher_from_seller?: number;
    voucher_from_shopee?: number;
    withholding_pit_tax?: number;
    withholding_tax?: number;
    withholding_vat_tax?: number;
  };
  order_sn?: string;
  return_order_sn_list?: string[];
  [key: string]: any;
}

export interface ProductSales {
  nama_produk: string;
  total_qty: number;
  total_value: number;
}

export interface DailyPurchaseTotal {
  date: string;
  total: number;
  count: number;
}

export interface MonthlyPurchaseTotal {
  month: string;
  total: number;
  count: number;
}

export interface ShopeeAffiliateSale {
  nama_toko: string;
  kode_pesanan: string;
  status_pesanan: string;
  status_terverifikasi: string;
  waktu_pesanan: string;
  waktu_pesanan_selesai: string;
  waktu_pesanan_terverifikasi: string;
  kode_produk: string;
  nama_produk: string;
  id_model: string;
  l1_kategori_global: string;
  l2_kategori_global: string;
  l3_kategori_global: string;
  kode_promo: string;
  harga: number;
  jumlah: number;
  nama_affiliate: string;
  username_affiliate: string;
  mcn_terhubung: string;
  id_komisi_pesanan: string;
  partner_promo: string;
  jenis_promo: string;
  nilai_pembelian: number;
  jumlah_pengembalian: number;
  tipe_pesanan: string;
  estimasi_komisi_per_produk: number;
  estimasi_komisi_affiliate_per_produk: number;
  persentase_komisi_affiliate_per_produk: number;
  estimasi_komisi_mcn_per_produk: number;
  persentase_komisi_mcn_per_produk: number;
  estimasi_komisi_per_pesanan: number;
  estimasi_komisi_affiliate_per_pesanan: number;
  estimasi_komisi_mcn_per_pesanan: number;
  catatan_produk: string;
  platform: string;
  tingkat_komisi: number;
  pengeluaran: number;
  status_pemotongan: string;
  metode_pemotongan: string;
  waktu_pemotongan: string;
}

export interface ShopeeAffiliateSummary {
  total_nilai_pembelian: number;
  total_komisi_affiliate: number;
}

export interface SalesProfit {
  kode_pesanan: string;
  tanggal_pesanan: string;
  modal_purchase: number;
  amount_sales: number;
  biaya_mitra_jakmall: number;
  biaya_administrasi: number;
  biaya_layanan: number;
  biaya_voucher: number;
  biaya_transaksi: number;
  diskon_ongkir: number;
  biaya_affiliate: number;
  biaya_refund: number;
  selisih_ongkir: number;
  adjustment_income: number;
  discount: number;
  profit: number;
  profit_percent: number;
}

export interface CancelledSummary {
  count: number;
  biaya_mitra: number;
}

export interface ShopeeAdjustment {
  id: number;
  nama_toko: string;
  tanggal_penyesuaian: string;
  tipe_penyesuaian: string;
  alasan_penyesuaian: string;
  biaya_penyesuaian: number;
  no_pesanan: string;
  created_at: string;
}

export interface TaxPayment {
  id: string;
  store: string;
  period_type: string;
  period_value: string;
  revenue: number;
  tax_rate: number;
  tax_amount: number;
  is_paid: boolean;
  paid_at: string;
}

export interface ShopeeOrderDetailRow {
  order_sn: string;
  nama_toko: string;
  status?: string;
  order_status?: string;
  checkout_time?: string;
  update_time?: string;
  pay_time?: string;
  total_amount?: number;
  currency?: string;
  actual_shipping_fee_confirmed?: boolean;
  buyer_cancel_reason?: string;
  buyer_cpf_id?: string;
  buyer_user_id?: number;
  buyer_username?: string;
  cancel_by?: string;
  cancel_reason?: string;
  cod?: boolean;
  create_time?: string;
  days_to_ship?: number;
  dropshipper?: string;
  dropshipper_phone?: string;
  estimated_shipping_fee?: number;
  fulfillment_flag?: string;
  goods_to_declare?: boolean;
  message_to_seller?: string;
  note?: string;
  note_update_time?: string;
  pickup_done_time?: string;
  region?: string;
  reverse_shipping_fee?: number;
  ship_by_date?: string;
  shipping_carrier?: string;
  split_up?: boolean;
  payment_method?: string;
  recipient_name?: string;
  recipient_phone?: string;
  recipient_full_address?: string;
  recipient_city?: string;
  recipient_district?: string;
  recipient_state?: string;
  recipient_town?: string;
  recipient_zipcode?: string;
  created_at: string;
}

export interface ShopeeOrderItemRow {
  id: number;
  order_sn: string;
  order_item_id?: number;
  item_name?: string;
  model_original_price?: number;
  model_quantity_purchased?: number;
  item_id?: number;
  item_sku?: string;
  model_id?: number;
  model_name?: string;
  model_sku?: string;
  model_discounted_price?: number;
  weight?: number;
  promotion_id?: number;
  promotion_type?: string;
  promotion_group_id?: number;
  add_on_deal?: boolean;
  add_on_deal_id?: number;
  main_item?: boolean;
  is_b2c_owned_item?: boolean;
  is_prescription_item?: boolean;
  wholesale?: boolean;
  product_location_id?: string[];
  image_url?: string;
}

export interface ShopeeOrderPackageRow {
  id: number;
  order_sn: string;
  package_number?: string;
  logistics_status?: string;
  shipping_carrier?: string;
  logistics_channel_id?: number;
  parcel_chargeable_weight_gram?: number;
  allow_self_design_awb?: boolean;
  sorting_group?: string;
  group_shipment_id?: string;
}

export interface WalletTransaction {
  transaction_id: number;
  status: string;
  transaction_type: string;
  amount: number;
  current_balance: number;
  create_time: number;
  order_sn?: string;
  refund_sn?: string;
  withdrawal_type?: string;
  transaction_fee?: number;
  description?: string;
  buyer_name?: string;
  pay_order_list?: { order_sn: string; shop_name: string }[];
  withdrawal_id?: number;
  reason?: string;
  root_withdrawal_id?: number;
  transaction_tab_type?: string;
  money_flow?: string;
  journaled?: boolean;
}

export interface BatchHistory {
  id: number;
  process_type: string;
  started_at: string;
  total_data: number;
  done_data: number;
  status: string;
  error_message: string;
  file_name: string;
  file_path: string;
}

export interface BatchHistoryDetail {
  id: number;
  batch_id: number;
  reference: string;
  store: string;
  status: string;
  error_message: string;
}

export interface DashboardData {
  summary: DashboardMetrics;
  charts: Record<string, { date: string; value: number }[]>;
}

export interface DashboardMetric {
  value: number;
  change: number;
}

export interface DashboardMetrics {
  total_orders?: DashboardMetric;
  avg_order_value?: DashboardMetric;
  total_cancelled?: DashboardMetric;
  total_customers?: DashboardMetric;
  total_price?: DashboardMetric;
  total_discounts?: DashboardMetric;
  total_net_profit?: DashboardMetric;
  outstanding_amount?: DashboardMetric;
  [key: string]: DashboardMetric | undefined;
}

export interface ShopeeOrderReturn {
  request_id: string;
  return_sn: string;
  order_sn: string;
  status: string;
  negotiation_status: string;
  seller_proof_status: string;
  seller_compensation_status: string;
  refund_amount: number;
  currency: string;
  create_time: number;
  update_time: number;
  user: ShopeeReturnUser;
  item: ShopeeReturnItem[];
  reason: string;
  due_date: number;
  image_info: ShopeeReturnImageInfo[];
  video_info: ShopeeReturnVideoInfo[];
  seller_proof: ShopeeReturnSellerProof[];
}

export interface ShopeeReturnUser {
  userid: number;
  username: string;
}

export interface ShopeeReturnItem {
  itemid: number;
  item_name: string;
  item_sku: string;
  modelid: number;
  model_name: string;
  model_sku: string;
  amount: number;
  item_price: number;
  is_main_item: boolean;
}

export interface ShopeeReturnImageInfo {
  image_id: string;
  image_url: string;
}

export interface ShopeeReturnVideoInfo {
  video_id: string;
  video_url: string;
}

export interface ShopeeReturnSellerProof {
  image_id: string;
  image_url: string;
}

export interface ShopeeReturnResponse {
  data: ShopeeOrderReturn[];
  has_more: boolean;
  total: number;
}
