-- File: backend/internal/migrations/0002_add_seabank_account.up.sql

-- 1) Resynchronize the accounts.sequence so nextval > MAX(account_id)
SELECT setval(
  pg_get_serial_sequence('accounts', 'account_id'),
  (SELECT COALESCE(MAX(account_id), 0) FROM accounts) + 1
);

-- 2) Add Seabank as a cash/bank asset under code 1.1
WITH parent AS (
  SELECT account_id AS pid
    FROM accounts
   WHERE account_code = '1.1'
)
INSERT INTO accounts (account_code, account_name, account_type, parent_id)
VALUES ('1.1.14', 'Seabank', 'Asset', (SELECT pid FROM parent));

-- 3) Link it in asset_accounts
INSERT INTO asset_accounts (account_id)
SELECT account_id FROM accounts WHERE account_code = '1.1.14';