CREATE TABLE shopee_adjustments (
  id SERIAL PRIMARY KEY,
  nama_toko VARCHAR(128) NOT NULL,
  tanggal_penyesuaian DATE NOT NULL,
  tipe_penyesuaian VARCHAR(128) NOT NULL,
  alasan_penyesuaian TEXT,
  biaya_penyesuaian NUMERIC NOT NULL,
  no_pesanan VARCHAR(64) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  UNIQUE(no_pesanan, tanggal_penyesuaian, tipe_penyesuaian)
);
