-- 0002_create_dropship_purchases.up.sql

CREATE TABLE IF NOT EXISTS dropship_purchases (
  id               SERIAL PRIMARY KEY,
  seller_username  VARCHAR(64) NOT NULL,
  purchase_id      VARCHAR(64) NOT NULL UNIQUE,
  order_id         VARCHAR(64),
  sku              VARCHAR(64) NOT NULL,
  qty              INT NOT NULL,
  purchase_price   NUMERIC(12,2) NOT NULL,
  purchase_fee     NUMERIC(12,2) NOT NULL,
  status           VARCHAR(32) NOT NULL,
  purchase_date    DATE NOT NULL,
  supplier_name    VARCHAR(128),
  created_at       TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMP NOT NULL DEFAULT NOW()
);
