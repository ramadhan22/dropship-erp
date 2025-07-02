DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM accounts WHERE account_code='5.5.5') THEN
        INSERT INTO accounts (account_id, account_code, account_name, account_type, parent_id)
        VALUES (52009, '5.5.5', 'Refund', 'Expense',
            (SELECT account_id FROM accounts WHERE account_code='5.5'));
    END IF;
END $$;
