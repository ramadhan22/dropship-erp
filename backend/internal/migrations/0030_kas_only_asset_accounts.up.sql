DELETE FROM asset_accounts WHERE account_id NOT IN (
  SELECT account_id FROM accounts
  WHERE account_name ILIKE '%kas%' OR account_name ILIKE 'bank%'
);
