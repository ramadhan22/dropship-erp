-- Consolidated Core Tables Migration
-- Replaces migrations: 0001, 0004, 0005, 0011, 0015, 0027, 0041, 0050, 0051

CREATE TABLE IF NOT EXISTS accounts (
  account_id   SERIAL PRIMARY KEY,
  account_code VARCHAR(16) NOT NULL UNIQUE,
  account_name VARCHAR(128) NOT NULL,
  account_type VARCHAR(16) NOT NULL,
  parent_id    INT REFERENCES accounts(account_id),
  created_at   TIMESTAMP DEFAULT NOW(),
  updated_at   TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS journal_entries (
  journal_id    SERIAL PRIMARY KEY,
  entry_date    DATE NOT NULL,
  description   TEXT,
  source_type   VARCHAR(32) NOT NULL,
  source_id     VARCHAR(64) NOT NULL,
  shop_username VARCHAR(64) NOT NULL,
  created_at    TIMESTAMP NOT NULL DEFAULT NOW(),
  store         TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS journal_lines (
  line_id    SERIAL PRIMARY KEY,
  journal_id INT NOT NULL REFERENCES journal_entries(journal_id) ON DELETE CASCADE,
  account_id INT NOT NULL REFERENCES accounts(account_id),
  is_debit   BOOLEAN NOT NULL,
  amount     NUMERIC(14,2) NOT NULL,
  memo       TEXT
);

CREATE TABLE IF NOT EXISTS jenis_channels (
  jenis_channel_id SERIAL PRIMARY KEY,
  jenis_channel    TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS stores (
  store_id         SERIAL PRIMARY KEY,
  jenis_channel_id INT NOT NULL REFERENCES jenis_channels(jenis_channel_id) ON DELETE CASCADE,
  nama_toko        TEXT NOT NULL,
  code_id          TEXT,
  shop_id          TEXT,
  access_token     TEXT,
  refresh_token    TEXT,
  expire_in        INT,
  request_id       TEXT,
  last_updated     TIMESTAMP
);

CREATE TABLE IF NOT EXISTS asset_accounts (
  asset_account_id SERIAL PRIMARY KEY,
  account_id       INT NOT NULL REFERENCES accounts(account_id),
  account_name     VARCHAR(128) NOT NULL,
  balance          NUMERIC(14,2) DEFAULT 0,
  created_at       TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMP NOT NULL DEFAULT NOW()
);
