DELETE FROM asset_accounts WHERE account_id = (SELECT account_id FROM accounts WHERE account_code = '1.1.14');
DELETE FROM accounts WHERE account_code = '1.1.14';
