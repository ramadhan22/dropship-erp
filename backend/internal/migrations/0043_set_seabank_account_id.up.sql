DO $$
DECLARE
    old_id INT;
BEGIN
    SELECT account_id INTO old_id FROM accounts WHERE account_code='1.1.14';
    IF old_id IS NULL THEN
        RETURN;
    END IF;
    IF old_id <> 11014 THEN
        UPDATE journal_lines SET account_id=11014 WHERE account_id=old_id;
        UPDATE expenses SET account_id=11014 WHERE account_id=old_id;
        UPDATE asset_accounts SET account_id=11014 WHERE account_id=old_id;
        UPDATE accounts SET account_id=11014 WHERE account_code='1.1.14';
        PERFORM setval('accounts_account_id_seq', (SELECT MAX(account_id) FROM accounts));
    END IF;
END $$;
