-- 0010_add_nama_toko_to_shopee_settled.down.sql
ALTER TABLE shopee_settled
    DROP COLUMN IF EXISTS nama_toko;
