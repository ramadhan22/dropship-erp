-- Update account_type to 'Kas' for cash/bank accounts
UPDATE accounts
SET account_type = 'Kas'
WHERE account_name ILIKE '%kas%' OR account_name ILIKE 'bank%';

-- Ensure asset_accounts only include Kas accounts
DELETE FROM asset_accounts
WHERE account_id NOT IN (
  SELECT account_id FROM accounts WHERE account_type='Kas'
);
