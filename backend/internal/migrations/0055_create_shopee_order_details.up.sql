CREATE TABLE IF NOT EXISTS shopee_order_details (
    order_sn VARCHAR(64) PRIMARY KEY,
    nama_toko TEXT,
    detail JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS shopee_order_items (
    id SERIAL PRIMARY KEY,
    order_sn VARCHAR(64) REFERENCES shopee_order_details(order_sn) ON DELETE CASCADE,
    item JSONB
);
