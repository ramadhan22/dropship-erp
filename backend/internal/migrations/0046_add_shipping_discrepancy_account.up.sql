DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM accounts WHERE account_code='5.2.10') THEN
        INSERT INTO accounts (account_id, account_code, account_name, account_type, parent_id)
        VALUES (52010, '5.2.10', 'Selisih Ongkir', 'Expense',
            (SELECT account_id FROM accounts WHERE account_code='5.2'));
    END IF;
END $$;
