CREATE TABLE asset_accounts (
  id SERIAL PRIMARY KEY,
  account_id INT NOT NULL UNIQUE REFERENCES accounts(account_id),
  created_at TIMESTAMP NOT NULL DEFAULT now()
);
