-- Consolidated Shopee Integration Tables Migration
-- Replaces migrations: 0042, 0044, 0049, 0055, 0056, 0057

CREATE TABLE IF NOT EXISTS shopee_order_details (
  order_sn       VARCHAR(64) PRIMARY KEY,
  nama_toko      TEXT,
  status         TEXT,
  checkout_time  BIGINT,
  update_time    BIGINT,
  pay_time       BIGINT,
  total_amount   NUMERIC,
  currency       TEXT,
  created_at     TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS shopee_order_items (
  id                         SERIAL PRIMARY KEY,
  order_sn                   VARCHAR(64) REFERENCES shopee_order_details(order_sn) ON DELETE CASCADE,
  order_item_id              BIGINT,
  item_name                  TEXT,
  model_original_price       NUMERIC,
  model_quantity_purchased   INT
);

CREATE TABLE IF NOT EXISTS shopee_order_packages (
  id                    SERIAL PRIMARY KEY,
  order_sn              VARCHAR(64) REFERENCES shopee_order_details(order_sn) ON DELETE CASCADE,
  package_id            VARCHAR(64),
  tracking_number       VARCHAR(100),
  shipping_carrier      VARCHAR(100),
  item_list             JSONB,
  created_at            TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS shopee_adjustments (
  id                    SERIAL PRIMARY KEY,
  nama_toko             VARCHAR(100) NOT NULL,
  no_pesanan            VARCHAR(100) NOT NULL,
  username_pembeli      VARCHAR(100) NOT NULL,
  waktu_pesanan_dibuat  TIMESTAMP NOT NULL,
  tanggal_penyesuaian   DATE NOT NULL,
  jenis_penyesuaian     VARCHAR(100) NOT NULL,
  jumlah_penyesuaian    NUMERIC(15,2) NOT NULL,
  alasan_penyesuaian    TEXT,
  created_at            TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS withdrawals (
  id             SERIAL PRIMARY KEY,
  nama_toko      VARCHAR(100) NOT NULL,
  tanggal_tarik  DATE NOT NULL,
  jumlah_tarik   NUMERIC(15,2) NOT NULL,
  status         VARCHAR(50) NOT NULL,
  nomor_referensi VARCHAR(100),
  created_at     TIMESTAMP DEFAULT NOW()
);
