-- 0049_adjust_varchar_lengths.up.sql
-- Ensure all VARCHAR columns are at least length 100

ALTER TABLE accounts
    ALTER COLUMN account_code TYPE VARCHAR(100),
    ALTER COLUMN account_type TYPE VARCHAR(100);

ALTER TABLE shopee_settled_orders
    ALTER COLUMN order_id TYPE VARCHAR(100),
    ALTER COLUMN seller_username TYPE VARCHAR(100);

ALTER TABLE journal_entries
    ALTER COLUMN source_type TYPE VARCHAR(100),
    ALTER COLUMN source_id TYPE VARCHAR(100),
    ALTER COLUMN shop_username TYPE VARCHAR(100);

ALTER TABLE reconciled_transactions
    ALTER COLUMN shop_username TYPE VARCHAR(100),
    ALTER COLUMN dropship_id TYPE VARCHAR(100),
    ALTER COLUMN shopee_id TYPE VARCHAR(100),
    ALTER COLUMN status TYPE VARCHAR(100);

ALTER TABLE cached_metrics
    ALTER COLUMN shop_username TYPE VARCHAR(100),
    ALTER COLUMN period TYPE VARCHAR(100);

ALTER TABLE shopee_settled
    ALTER COLUMN no_pesanan TYPE VARCHAR(100),
    ALTER COLUMN no_pengajuan TYPE VARCHAR(100),
    ALTER COLUMN username_pembeli TYPE VARCHAR(100),
    ALTER COLUMN metode_pembayaran_pembeli TYPE VARCHAR(100),
    ALTER COLUMN jasa_kirim TYPE VARCHAR(100),
    ALTER COLUMN nama_kurir TYPE VARCHAR(100),
    ALTER COLUMN nama_toko TYPE VARCHAR(100);

ALTER TABLE shopee_affiliate_sales
    ALTER COLUMN kode_pesanan TYPE VARCHAR(100),
    ALTER COLUMN status_pesanan TYPE VARCHAR(100),
    ALTER COLUMN status_terverifikasi TYPE VARCHAR(100),
    ALTER COLUMN kode_produk TYPE VARCHAR(100),
    ALTER COLUMN id_model TYPE VARCHAR(100),
    ALTER COLUMN l1_kategori_global TYPE VARCHAR(100),
    ALTER COLUMN l2_kategori_global TYPE VARCHAR(100),
    ALTER COLUMN l3_kategori_global TYPE VARCHAR(100),
    ALTER COLUMN kode_promo TYPE VARCHAR(100),
    ALTER COLUMN id_komisi_pesanan TYPE VARCHAR(100),
    ALTER COLUMN partner_promo TYPE VARCHAR(100),
    ALTER COLUMN jenis_promo TYPE VARCHAR(100),
    ALTER COLUMN tipe_pesanan TYPE VARCHAR(100),
    ALTER COLUMN platform TYPE VARCHAR(100),
    ALTER COLUMN status_pemotongan TYPE VARCHAR(100),
    ALTER COLUMN metode_pemotongan TYPE VARCHAR(100);

ALTER TABLE ad_invoices
    ALTER COLUMN invoice_no TYPE VARCHAR(100),
    ALTER COLUMN username TYPE VARCHAR(100);

ALTER TABLE shopee_adjustments
    ALTER COLUMN no_pesanan TYPE VARCHAR(100);
