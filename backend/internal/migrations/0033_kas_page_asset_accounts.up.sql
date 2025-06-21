-- Restrict asset_accounts to accounts with codes under 1.1.1*
DELETE FROM asset_accounts
WHERE account_id NOT IN (
  SELECT account_id FROM accounts WHERE account_code LIKE '1.1.1%'
);

INSERT INTO asset_accounts (account_id)
SELECT account_id FROM accounts
WHERE account_code LIKE '1.1.1%'
  AND account_id NOT IN (SELECT account_id FROM asset_accounts);
