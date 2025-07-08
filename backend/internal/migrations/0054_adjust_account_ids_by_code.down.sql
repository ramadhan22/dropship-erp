-- Disable triggers so ID changes do not violate constraints
ALTER TABLE journal_lines DISABLE TRIGGER ALL;
ALTER TABLE expenses DISABLE TRIGGER ALL;
ALTER TABLE expense_lines DISABLE TRIGGER ALL;
ALTER TABLE asset_accounts DISABLE TRIGGER ALL;
ALTER TABLE accounts DISABLE TRIGGER ALL;

DO $$
DECLARE
    id_seq INT := 1;
    rec RECORD;
BEGIN
    FOR rec IN SELECT account_id FROM accounts ORDER BY account_id LOOP
        UPDATE journal_lines SET account_id = id_seq WHERE account_id = rec.account_id;
        UPDATE expenses SET asset_account_id = id_seq WHERE asset_account_id = rec.account_id;
        UPDATE expense_lines SET account_id = id_seq WHERE account_id = rec.account_id;
        UPDATE asset_accounts SET account_id = id_seq WHERE account_id = rec.account_id;
        UPDATE accounts SET account_id = id_seq WHERE account_id = rec.account_id;
        UPDATE accounts SET parent_id = id_seq WHERE parent_id = rec.account_id;
        id_seq := id_seq + 1;
    END LOOP;
    PERFORM setval('accounts_account_id_seq', (SELECT MAX(account_id) FROM accounts));
END $$;

-- Re-enable triggers
ALTER TABLE accounts ENABLE TRIGGER ALL;
ALTER TABLE journal_lines ENABLE TRIGGER ALL;
ALTER TABLE expenses ENABLE TRIGGER ALL;
ALTER TABLE expense_lines ENABLE TRIGGER ALL;
ALTER TABLE asset_accounts ENABLE TRIGGER ALL;
