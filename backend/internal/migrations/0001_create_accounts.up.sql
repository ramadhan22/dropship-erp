CREATE TABLE IF NOT EXISTS accounts (
  account_id   SERIAL PRIMARY KEY,
  account_code VARCHAR(16) NOT NULL UNIQUE,
  account_name VARCHAR(128) NOT NULL,
  account_type VARCHAR(16) NOT NULL,
  parent_id    INT REFERENCES accounts(account_id),
  created_at   TIMESTAMP DEFAULT NOW(),
  updated_at   TIMESTAMP DEFAULT NOW()
);