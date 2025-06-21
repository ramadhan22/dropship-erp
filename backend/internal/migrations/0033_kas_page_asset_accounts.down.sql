-- Restore asset_accounts to all Kas accounts
DELETE FROM asset_accounts;

INSERT INTO asset_accounts (account_id)
SELECT account_id FROM accounts
WHERE account_type = 'Kas'
  AND account_id NOT IN (SELECT account_id FROM asset_accounts);
