-- Consolidated Business Tables Migration
-- Replaces migrations: 0002, 0008, 0012, 0021, 0022, 0028
-- Includes missing tables: shopee_settled, shopee_settled_orders, reconciled_transactions

CREATE TABLE IF NOT EXISTS dropship_purchases (
  kode_pesanan             TEXT        PRIMARY KEY,
  kode_transaksi           TEXT,
  waktu_pesanan_terbuat    TIMESTAMP,
  status_pesanan_terakhir  TEXT,
  biaya_lainnya            NUMERIC,
  biaya_mitra_jakmall      NUMERIC,
  total_transaksi          NUMERIC,
  dibuat_oleh              TEXT,
  jenis_channel            TEXT,
  nama_toko                TEXT,
  kode_invoice_channel     TEXT,
  gudang_pengiriman        TEXT,
  jenis_ekspedisi          TEXT,
  cashless                 TEXT,
  nomor_resi               TEXT,
  waktu_pengiriman         TIMESTAMP,
  provinsi                 TEXT,
  kota                     TEXT
);

CREATE TABLE IF NOT EXISTS dropship_purchase_details (
  id                          SERIAL PRIMARY KEY,
  kode_pesanan                TEXT NOT NULL REFERENCES dropship_purchases(kode_pesanan) ON DELETE CASCADE,
  sku                         TEXT,
  nama_produk                 TEXT,
  harga_produk                NUMERIC,
  qty                         INTEGER,
  total_harga_produk          NUMERIC,
  harga_produk_channel        NUMERIC,
  total_harga_produk_channel  NUMERIC,
  potensi_keuntungan          NUMERIC
);

CREATE TABLE IF NOT EXISTS expenses (
  expense_id       SERIAL PRIMARY KEY,
  expense_date     DATE NOT NULL,
  description      TEXT NOT NULL,
  total_amount     NUMERIC(14,2) NOT NULL,
  asset_account_id INT NOT NULL REFERENCES asset_accounts(asset_account_id),
  created_at       TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS expense_lines (
  line_id      SERIAL PRIMARY KEY,
  expense_id   INT NOT NULL REFERENCES expenses(expense_id) ON DELETE CASCADE,
  account_id   INT NOT NULL REFERENCES accounts(account_id),
  amount       NUMERIC(14,2) NOT NULL,
  description  TEXT
);

CREATE TABLE IF NOT EXISTS shopee_affiliate_sales (
  id                   SERIAL PRIMARY KEY,
  item_name            VARCHAR(256) NOT NULL,
  order_id             VARCHAR(64) NOT NULL,
  qty                  INT NOT NULL,
  selling_price        NUMERIC(12,2) NOT NULL,
  affiliate_commission NUMERIC(12,2) NOT NULL,
  order_time           TIMESTAMP NOT NULL,
  username             VARCHAR(64) NOT NULL,
  created_at           TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at           TIMESTAMP NOT NULL DEFAULT NOW(),
  nama_toko            TEXT
);

-- Shopee Settlement Tables (missing from original migrations but used in code)
CREATE TABLE IF NOT EXISTS shopee_settled_orders (
  id               SERIAL PRIMARY KEY,
  order_id         VARCHAR(64) NOT NULL UNIQUE,
  net_income       NUMERIC(15,2) NOT NULL,
  service_fee      NUMERIC(15,2) NOT NULL,
  campaign_fee     NUMERIC(15,2) NOT NULL,
  credit_card_fee  NUMERIC(15,2) NOT NULL,
  shipping_subsidy NUMERIC(15,2) NOT NULL,
  tax_and_import_fee NUMERIC(15,2) NOT NULL,
  settled_date     DATE NOT NULL,
  seller_username  VARCHAR(64) NOT NULL,
  created_at       TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS shopee_settled (
  nama_toko                                         VARCHAR(100) NOT NULL,
  no_pesanan                                        VARCHAR(100) NOT NULL PRIMARY KEY,
  no_pengajuan                                      VARCHAR(100) NOT NULL,
  username_pembeli                                  VARCHAR(100) NOT NULL,
  waktu_pesanan_dibuat                              TIMESTAMP NOT NULL,
  metode_pembayaran_pembeli                         VARCHAR(100) NOT NULL,
  tanggal_dana_dilepaskan                           DATE NOT NULL,
  harga_asli_produk                                 NUMERIC(15,2) NOT NULL,
  total_diskon_produk                               NUMERIC(15,2) NOT NULL,
  diskon_voucher_ditanggung_penjual                 NUMERIC(15,2) NOT NULL,
  biaya_administrasi                                NUMERIC(15,2) NOT NULL,
  biaya_layanan_termasuk_ppn_11                     NUMERIC(15,2) NOT NULL,
  total_penghasilan                                 NUMERIC(15,2) NOT NULL,
  kompensasi_ongkir_dibayar_oleh_penjual            NUMERIC(15,2) NOT NULL,
  ongkir_dibayar_pembeli                            NUMERIC(15,2) NOT NULL,
  diskon_ongkir_ditanggung_jasa_kirim               NUMERIC(15,2) NOT NULL,
  nama_produk                                       TEXT NOT NULL,
  nomor_referensi_sku                               VARCHAR(100) NOT NULL,
  nama_variasi                                      TEXT,
  jumlah_produk_dibeli                              INT NOT NULL,
  berat_produk_gram                                 NUMERIC(10,2),
  nilai_produk_sebelum_diskon                       NUMERIC(15,2) NOT NULL,
  biaya_transaksi                                   NUMERIC(15,2) DEFAULT 0
);

-- Reconciliation Table (missing from original migrations but used in code)
CREATE TABLE IF NOT EXISTS reconciled_transactions (
  id            SERIAL PRIMARY KEY,
  shop_username VARCHAR(64) NOT NULL,
  dropship_id   VARCHAR(64),
  shopee_id     VARCHAR(100),
  status        VARCHAR(32) NOT NULL,
  matched_at    TIMESTAMP NOT NULL DEFAULT NOW()
);
