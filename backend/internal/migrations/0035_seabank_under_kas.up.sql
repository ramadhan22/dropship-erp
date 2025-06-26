-- Move Seabank account under Kas
UPDATE accounts
SET parent_id = (SELECT account_id FROM accounts WHERE account_code = '1.1'),  account_type = 'Asset'
WHERE account_code = '1.1.14';
