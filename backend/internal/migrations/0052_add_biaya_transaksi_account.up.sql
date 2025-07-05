DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM accounts WHERE account_code='5.2.11') THEN
        INSERT INTO accounts (account_id, account_code, account_name, account_type, parent_id)
        VALUES (52012, '5.2.11', 'Biaya Transaksi', 'Expense',
            (SELECT account_id FROM accounts WHERE account_code='5.2'));
    END IF;
END $$;
