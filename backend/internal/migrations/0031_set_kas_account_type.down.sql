-- Revert kas accounts back to Asset type
UPDATE accounts
SET account_type = 'Asset'
WHERE account_type = 'Kas';

-- Repopulate asset_accounts with asset (kas) accounts
INSERT INTO asset_accounts (account_id)
SELECT account_id FROM accounts
WHERE account_type = 'Kas'
  AND account_id NOT IN (SELECT account_id FROM asset_accounts);
