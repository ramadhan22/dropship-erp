-- 0014_insert_platform_accounts.down.sql
-- Remove platform-specific accounts

DELETE FROM accounts WHERE account_code IN (
  '1.1.13','1.1.12','1.1.11','1.1.10','1.1.9'
);
