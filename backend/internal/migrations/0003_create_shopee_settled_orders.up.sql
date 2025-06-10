-- 0003_create_shopee_settled_orders.up.sql

CREATE TABLE IF NOT EXISTS shopee_settled_orders (
  id                SERIAL PRIMARY KEY,
  order_id          VARCHAR(64) NOT NULL UNIQUE,
  net_income        NUMERIC(12,2) NOT NULL,
  service_fee       NUMERIC(12,2) NOT NULL,
  campaign_fee      NUMERIC(12,2) NOT NULL,
  credit_card_fee   NUMERIC(12,2) NOT NULL,
  shipping_subsidy  NUMERIC(12,2) NOT NULL,
  tax_and_import_fee NUMERIC(12,2) NOT NULL,
  settled_date      DATE NOT NULL,
  seller_username   VARCHAR(64) NOT NULL,
  created_at        TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMP NOT NULL DEFAULT NOW()
);
