DO $$
DECLARE
    old_id INT;
BEGIN
    SELECT account_id INTO old_id FROM accounts WHERE account_code='5.2.3.4';
    IF old_id IS NULL THEN
        RETURN;
    END IF;
    IF old_id <> 52002 THEN
        UPDATE journal_lines SET account_id=52002 WHERE account_id=old_id;
        UPDATE expenses SET account_id=52002 WHERE account_id=old_id;
        UPDATE expense_lines SET account_id=52002 WHERE account_id=old_id;
        UPDATE accounts SET account_id=52002 WHERE account_code='5.2.3.4';
        PERFORM setval('accounts_account_id_seq', (SELECT MAX(account_id) FROM accounts));
    END IF;
END $$;
