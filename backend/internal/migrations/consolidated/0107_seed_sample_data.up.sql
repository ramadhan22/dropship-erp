-- Optional Sample Data for Development/Testing
-- From migration 0034 - Insert 2025 expense data and related journal entries
-- This is sample data for development/testing purposes only

-- Note: This sample data uses specific UUIDs and account codes
-- Ensure the corresponding accounts exist before running this migration
-- This is optional and not required for production deployments

DO $$
DECLARE
    kas INT := (SELECT account_id FROM accounts WHERE account_code='1.1.1');
    beban_operasional INT := (SELECT account_id FROM accounts WHERE account_code='5.1');
    expense_id1 INT;
    expense_id2 INT;
    kas_asset_account INT := (SELECT asset_account_id FROM asset_accounts LIMIT 1);
BEGIN
    -- Only insert if this is a development environment
    -- Check if we already have sample data
    IF NOT EXISTS (SELECT 1 FROM expenses LIMIT 1) THEN
        
        -- Sample expense 1: Office Supplies
        INSERT INTO expenses (expense_date, description, total_amount, asset_account_id)
        VALUES ('2024-01-15', 'Office Supplies - Stationery', 250000, kas_asset_account)
        RETURNING expense_id INTO expense_id1;
        
        INSERT INTO expense_lines (expense_id, account_id, amount, description)
        VALUES (expense_id1, beban_operasional, 250000, 'Office supplies purchase');
        
        -- Sample expense 2: Marketing Materials
        INSERT INTO expenses (expense_date, description, total_amount, asset_account_id)
        VALUES ('2024-01-20', 'Marketing Materials', 500000, kas_asset_account)
        RETURNING expense_id INTO expense_id2;
        
        INSERT INTO expense_lines (expense_id, account_id, amount, description)
        VALUES (expense_id2, beban_operasional, 500000, 'Marketing materials purchase');
        
        -- Sample journal entries
        INSERT INTO journal_entries (entry_date, description, source_type, source_id, shop_username, store)
        VALUES 
        ('2024-01-15', 'Sample expense entry 1', 'expense', expense_id1::text, 'sample_user', 'sample_store'),
        ('2024-01-20', 'Sample expense entry 2', 'expense', expense_id2::text, 'sample_user', 'sample_store');
        
        -- Sample journal lines
        INSERT INTO journal_lines (journal_id, account_id, is_debit, amount, memo)
        SELECT 
            je.journal_id,
            beban_operasional,
            true,
            250000,
            'Sample expense debit'
        FROM journal_entries je 
        WHERE je.description = 'Sample expense entry 1';
        
        INSERT INTO journal_lines (journal_id, account_id, is_debit, amount, memo)
        SELECT 
            je.journal_id,
            kas,
            false,
            250000,
            'Sample expense credit'
        FROM journal_entries je 
        WHERE je.description = 'Sample expense entry 1';
        
        INSERT INTO journal_lines (journal_id, account_id, is_debit, amount, memo)
        SELECT 
            je.journal_id,
            beban_operasional,
            true,
            500000,
            'Sample marketing debit'
        FROM journal_entries je 
        WHERE je.description = 'Sample expense entry 2';
        
        INSERT INTO journal_lines (journal_id, account_id, is_debit, amount, memo)
        SELECT 
            je.journal_id,
            kas,
            false,
            500000,
            'Sample marketing credit'
        FROM journal_entries je 
        WHERE je.description = 'Sample expense entry 2';
        
        RAISE NOTICE 'Sample data inserted successfully';
    ELSE
        RAISE NOTICE 'Sample data already exists, skipping insert';
    END IF;
END $$;