-- 0006_create_reconciled_transactions.up.sql

CREATE TABLE IF NOT EXISTS reconciled_transactions (
  id            SERIAL PRIMARY KEY,
  shop_username VARCHAR(64) NOT NULL,
  dropship_id   VARCHAR(64),
  shopee_id     VARCHAR(64),
  status        VARCHAR(16) NOT NULL,
  matched_at    TIMESTAMP NOT NULL DEFAULT NOW()
);
