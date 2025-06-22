-- Revert Seabank parent to Aset Lancar
UPDATE accounts
SET parent_id = (SELECT account_id FROM accounts WHERE account_code = '1.1')
WHERE account_code = '1.1.14';
