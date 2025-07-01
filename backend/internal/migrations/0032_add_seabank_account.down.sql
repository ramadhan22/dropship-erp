-- File: backend/internal/migrations/0002_add_seabank_account.down.sql

-- Remove from asset_accounts first
DELETE FROM asset_accounts
 WHERE account_id IN (
   SELECT account_id FROM accounts WHERE account_code = '1.1.14'
 );

-- Then remove the Seabank account
DELETE FROM accounts
 WHERE account_code = '1.1.14';