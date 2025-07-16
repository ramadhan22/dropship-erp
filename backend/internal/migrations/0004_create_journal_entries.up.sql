CREATE TABLE IF NOT EXISTS journal_entries (
  journal_id    SERIAL PRIMARY KEY,
  entry_date    DATE NOT NULL,
  description   TEXT,
  source_type   VARCHAR(32) NOT NULL,
  source_id     VARCHAR(64) NOT NULL,
  shop_username VARCHAR(64) NOT NULL,
  created_at    TIMESTAMP NOT NULL DEFAULT NOW()
);