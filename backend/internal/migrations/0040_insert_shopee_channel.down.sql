-- 0040_insert_shopee_channel.down.sql
-- Remove seeded Shopee channel and stores

DELETE FROM stores WHERE nama_toko IN ('MR eStore Shopee', 'MR Barista Gear');
DELETE FROM jenis_channels WHERE jenis_channel = 'Shopee';
