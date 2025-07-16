-- 0008_restructure_dropship_purchases.down.sql
DROP TABLE IF EXISTS dropship_purchase_details;
DROP TABLE IF EXISTS dropship_purchases;

-- Recreate original dropship_purchases table from migration 0002
CREATE TABLE dropship_purchases (
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