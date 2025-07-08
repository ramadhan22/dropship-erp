DO $$
DECLARE
    rec RECORD;
    parts TEXT[];
    expected INT;
    offset INT := 1000000;
BEGIN
    -- Temporarily shift all account IDs to avoid primary key conflicts
    FOR rec IN SELECT account_id FROM accounts ORDER BY account_id DESC LOOP
        UPDATE journal_lines SET account_id = rec.account_id + offset WHERE account_id = rec.account_id;
        UPDATE expenses SET account_id = rec.account_id + offset WHERE account_id = rec.account_id;
        UPDATE expense_lines SET account_id = rec.account_id + offset WHERE account_id = rec.account_id;
        UPDATE asset_accounts SET account_id = rec.account_id + offset WHERE account_id = rec.account_id;
        UPDATE accounts SET account_id = rec.account_id + offset WHERE account_id = rec.account_id;
    END LOOP;

    -- Reassign IDs based on account codes from the shifted values
    FOR rec IN
        SELECT account_id, account_code
        FROM accounts
        ORDER BY array_length(string_to_array(account_code, '.'), 1) DESC
    LOOP
        parts := string_to_array(rec.account_code, '.');
        IF array_length(parts, 1) = 2 THEN
            expected := parts[1]::INT * 1000 + parts[2]::INT;
        ELSIF array_length(parts, 1) >= 3 THEN
            expected := parts[1]::INT * 10000 + parts[2]::INT * 1000 + parts[3]::INT;
        ELSE
            CONTINUE;
        END IF;
        IF rec.account_id <> expected THEN
            UPDATE journal_lines SET account_id=expected WHERE account_id=rec.account_id;
            UPDATE expenses SET account_id=expected WHERE account_id=rec.account_id;
            UPDATE expense_lines SET account_id=expected WHERE account_id=rec.account_id;
            UPDATE asset_accounts SET account_id=expected WHERE account_id=rec.account_id;
            UPDATE accounts SET account_id=expected WHERE account_id=rec.account_id;
        END IF;
    END LOOP;
    PERFORM setval('accounts_account_id_seq', (SELECT MAX(account_id) FROM accounts));
END $$;
