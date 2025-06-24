DO $$
DECLARE
    seq_id INT;
    cur_id INT;
BEGIN
    SELECT account_id INTO cur_id FROM accounts WHERE account_code='5.2.3.4';
    IF cur_id IS NULL THEN
        RETURN;
    END IF;
    IF cur_id = 52002 THEN
        seq_id := nextval('accounts_account_id_seq');
        UPDATE journal_lines SET account_id=seq_id WHERE account_id=cur_id;
        UPDATE expenses SET account_id=seq_id WHERE account_id=cur_id;
        UPDATE expense_lines SET account_id=seq_id WHERE account_id=cur_id;
        UPDATE accounts SET account_id=seq_id WHERE account_code='5.2.3.4';
        PERFORM setval('accounts_account_id_seq', (SELECT MAX(account_id) FROM accounts));
    END IF;
END $$;
