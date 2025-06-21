-- Add Seabank cash/bank account
WITH parent AS (
  SELECT account_id AS pid FROM accounts WHERE account_code = '1.1'
)
INSERT INTO accounts (account_code, account_name, account_type, parent_id)
VALUES ('1.1.14', 'Seabank', 'Kas', (SELECT pid FROM parent));

INSERT INTO asset_accounts (account_id)
SELECT account_id FROM accounts WHERE account_code = '1.1.14';
