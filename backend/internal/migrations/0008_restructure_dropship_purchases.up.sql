-- 0008_restructure_dropship_purchases.up.sql
DROP TABLE IF EXISTS dropship_purchase_details;
DROP TABLE IF EXISTS dropship_purchases;

CREATE TABLE dropship_purchases (
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

CREATE TABLE dropship_purchase_details (
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
