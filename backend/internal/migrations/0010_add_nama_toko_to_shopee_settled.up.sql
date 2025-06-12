-- 0010_add_nama_toko_to_shopee_settled.up.sql
ALTER TABLE shopee_settled
    ADD COLUMN IF NOT EXISTS nama_toko VARCHAR(64);
