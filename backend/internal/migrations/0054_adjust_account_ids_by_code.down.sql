DO $$
DECLARE
    id_seq INT := 1;
    rec RECORD;
BEGIN
    FOR rec IN SELECT account_id FROM accounts ORDER BY account_id LOOP
        UPDATE journal_lines SET account_id=id_seq WHERE account_id=rec.account_id;
        UPDATE expenses SET account_id=id_seq WHERE account_id=rec.account_id;
        UPDATE expense_lines SET account_id=id_seq WHERE account_id=rec.account_id;
        UPDATE asset_accounts SET account_id=id_seq WHERE account_id=rec.account_id;
        UPDATE accounts SET account_id=id_seq WHERE account_id=rec.account_id;
        id_seq := id_seq + 1;
    END LOOP;
    PERFORM setval('accounts_account_id_seq', (SELECT MAX(account_id) FROM accounts));
END $$;
