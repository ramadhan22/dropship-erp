-- Temporarily disable triggers to freely update account relations
ALTER TABLE journal_lines DISABLE TRIGGER ALL;
ALTER TABLE expenses DISABLE TRIGGER ALL;
ALTER TABLE expense_lines DISABLE TRIGGER ALL;
ALTER TABLE asset_accounts DISABLE TRIGGER ALL;
ALTER TABLE accounts DISABLE TRIGGER ALL;

DO $$
DECLARE
    rec RECORD;
    parts TEXT[];
    expected INT;
    offset_id INT := 1000000;
BEGIN
    -- Temporarily shift all account IDs to avoid primary key conflicts
    FOR rec IN SELECT account_id FROM accounts ORDER BY account_id DESC LOOP
        UPDATE journal_lines SET account_id = rec.account_id + offset_id WHERE account_id = rec.account_id;
        UPDATE expenses SET asset_account_id = rec.account_id + offset_id WHERE asset_account_id = rec.account_id;
        UPDATE expense_lines SET account_id = rec.account_id + offset_id WHERE account_id = rec.account_id;
        UPDATE asset_accounts SET account_id = rec.account_id + offset_id WHERE account_id = rec.account_id;
        UPDATE accounts SET account_id = rec.account_id + offset_id WHERE account_id = rec.account_id;
        UPDATE accounts SET parent_id = rec.account_id + offset_id WHERE parent_id = rec.account_id;
    END LOOP;

    -- Reassign IDs based on account codes from the shifted values
    FOR rec IN
        SELECT account_id, account_code
        FROM accounts
        ORDER BY array_length(string_to_array(account_code, '.'), 1) DESC
    LOOP
        parts := string_to_array(rec.account_code, '.');
        IF array_length(parts, 1) = 1 THEN
            expected := parts[1]::INT;
        ELSIF array_length(parts, 1) = 2 THEN
            expected := parts[1]::INT * 1000 + parts[2]::INT;
        ELSIF array_length(parts, 1) >= 3 THEN
            expected := parts[1]::INT * 10000 + parts[2]::INT * 1000 + parts[3]::INT;
        ELSE
            CONTINUE;
        END IF;
        IF rec.account_id <> expected THEN
            UPDATE journal_lines SET account_id = expected WHERE account_id = rec.account_id;
            UPDATE expenses SET asset_account_id = expected WHERE asset_account_id = rec.account_id;
            UPDATE expense_lines SET account_id = expected WHERE account_id = rec.account_id;
            UPDATE asset_accounts SET account_id = expected WHERE account_id = rec.account_id;
            UPDATE accounts SET account_id = expected WHERE account_id = rec.account_id;
        END IF;
        UPDATE accounts SET parent_id = expected WHERE parent_id = rec.account_id;
    END LOOP;
    PERFORM setval('accounts_account_id_seq', (SELECT MAX(account_id) FROM accounts));
END $$;

-- Re-enable triggers after updates
ALTER TABLE accounts ENABLE TRIGGER ALL;
ALTER TABLE journal_lines ENABLE TRIGGER ALL;
ALTER TABLE expenses ENABLE TRIGGER ALL;
ALTER TABLE expense_lines ENABLE TRIGGER ALL;
ALTER TABLE asset_accounts ENABLE TRIGGER ALL;
