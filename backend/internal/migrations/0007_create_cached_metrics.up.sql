-- 0007_create_cached_metrics.up.sql

CREATE TABLE IF NOT EXISTS cached_metrics (
  id                  SERIAL PRIMARY KEY,
  shop_username       VARCHAR(64) NOT NULL,
  period              VARCHAR(7) NOT NULL,  -- e.g. '2025-05'
  sum_revenue         NUMERIC(14,2) NOT NULL,
  sum_cogs            NUMERIC(14,2) NOT NULL,
  sum_fees            NUMERIC(14,2) NOT NULL,
  net_profit          NUMERIC(14,2) NOT NULL,
  ending_cash_balance NUMERIC(14,2) NOT NULL,
  updated_at          TIMESTAMP NOT NULL DEFAULT NOW(),
  UNIQUE (shop_username, period)
);
