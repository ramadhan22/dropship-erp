-- 0010_create_jenis_channels_and_stores.up.sql
CREATE TABLE IF NOT EXISTS jenis_channels (
    jenis_channel_id SERIAL PRIMARY KEY,
    jenis_channel TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS stores (
    store_id SERIAL PRIMARY KEY,
    jenis_channel_id INT NOT NULL REFERENCES jenis_channels(jenis_channel_id) ON DELETE CASCADE,
    nama_toko TEXT NOT NULL
);
