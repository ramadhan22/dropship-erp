-- 0049_adjust_varchar_lengths.down.sql
-- Revert VARCHAR length adjustments

ALTER TABLE accounts
    ALTER COLUMN account_code TYPE VARCHAR(16),
    ALTER COLUMN account_type TYPE VARCHAR(16);

ALTER TABLE shopee_settled_orders
    ALTER COLUMN order_id TYPE VARCHAR(64),
    ALTER COLUMN seller_username TYPE VARCHAR(64);

ALTER TABLE journal_entries
    ALTER COLUMN source_type TYPE VARCHAR(32),
    ALTER COLUMN source_id TYPE VARCHAR(64),
    ALTER COLUMN shop_username TYPE VARCHAR(64);

ALTER TABLE reconciled_transactions
    ALTER COLUMN shop_username TYPE VARCHAR(64),
    ALTER COLUMN dropship_id TYPE VARCHAR(64),
    ALTER COLUMN shopee_id TYPE VARCHAR(64),
    ALTER COLUMN status TYPE VARCHAR(16);

ALTER TABLE cached_metrics
    ALTER COLUMN shop_username TYPE VARCHAR(64),
    ALTER COLUMN period TYPE VARCHAR(7);

ALTER TABLE shopee_settled
    ALTER COLUMN no_pesanan TYPE VARCHAR(32),
    ALTER COLUMN no_pengajuan TYPE VARCHAR(32),
    ALTER COLUMN username_pembeli TYPE VARCHAR(64),
    ALTER COLUMN metode_pembayaran_pembeli TYPE VARCHAR(64),
    ALTER COLUMN jasa_kirim TYPE VARCHAR(64),
    ALTER COLUMN nama_kurir TYPE VARCHAR(64),
    ALTER COLUMN nama_toko TYPE VARCHAR(64);

ALTER TABLE shopee_affiliate_sales
    ALTER COLUMN kode_pesanan TYPE VARCHAR(64),
    ALTER COLUMN status_pesanan TYPE VARCHAR(64),
    ALTER COLUMN status_terverifikasi TYPE VARCHAR(64),
    ALTER COLUMN kode_produk TYPE VARCHAR(64),
    ALTER COLUMN id_model TYPE VARCHAR(64),
    ALTER COLUMN l1_kategori_global TYPE VARCHAR(64),
    ALTER COLUMN l2_kategori_global TYPE VARCHAR(64),
    ALTER COLUMN l3_kategori_global TYPE VARCHAR(64),
    ALTER COLUMN kode_promo TYPE VARCHAR(64),
    ALTER COLUMN id_komisi_pesanan TYPE VARCHAR(64),
    ALTER COLUMN partner_promo TYPE VARCHAR(64),
    ALTER COLUMN jenis_promo TYPE VARCHAR(64),
    ALTER COLUMN tipe_pesanan TYPE VARCHAR(64),
    ALTER COLUMN platform TYPE VARCHAR(64),
    ALTER COLUMN status_pemotongan TYPE VARCHAR(64),
    ALTER COLUMN metode_pemotongan TYPE VARCHAR(64);

ALTER TABLE ad_invoices
    ALTER COLUMN invoice_no TYPE VARCHAR(64),
    ALTER COLUMN username TYPE VARCHAR(64);

ALTER TABLE shopee_adjustments
    ALTER COLUMN no_pesanan TYPE VARCHAR(64);
