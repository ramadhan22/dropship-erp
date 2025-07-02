DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM accounts WHERE account_code='5.4.1') THEN
        INSERT INTO accounts (account_id, account_code, account_name, account_type, parent_id)
        VALUES (54001, '5.4.1', 'PPh Final UMKM', 'Expense',
            (SELECT account_id FROM accounts WHERE account_code='5.4'));
    END IF;
END $$;
