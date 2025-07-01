-- 0040_insert_shopee_channel.up.sql
-- Seed Shopee channel and stores

INSERT INTO jenis_channels (jenis_channel)
SELECT 'Shopee'
WHERE NOT EXISTS (
    SELECT 1 FROM jenis_channels WHERE jenis_channel = 'Shopee'
);

INSERT INTO stores (jenis_channel_id, nama_toko)
SELECT jc.jenis_channel_id, 'MR eStore Shopee'
FROM jenis_channels jc
WHERE jc.jenis_channel = 'Shopee'
  AND NOT EXISTS (
      SELECT 1 FROM stores st WHERE st.nama_toko = 'MR eStore Shopee'
  );

INSERT INTO stores (jenis_channel_id, nama_toko)
SELECT jc.jenis_channel_id, 'MR Barista Gear'
FROM jenis_channels jc
WHERE jc.jenis_channel = 'Shopee'
  AND NOT EXISTS (
      SELECT 1 FROM stores st WHERE st.nama_toko = 'MR Barista Gear'
  );
