DO $$
DECLARE
    beban_usaha_id INT;
    marketing_id INT;
    aset_tetap_id INT;
BEGIN
    SELECT account_id INTO beban_usaha_id FROM accounts WHERE account_code='5.2';

    -- rename existing Voucher account to allow new subgroup code
    UPDATE accounts SET account_code='5.2.3.1' WHERE account_code='5.2.3';

    INSERT INTO accounts (account_code, account_name, account_type, parent_id)
    VALUES ('5.2.3', 'Beban Pemasaran', 'Expense', beban_usaha_id)
    RETURNING account_id INTO marketing_id;

    UPDATE accounts SET parent_id=marketing_id WHERE account_code='5.2.3.1';
    UPDATE accounts SET account_code='5.2.3.2', parent_id=marketing_id WHERE account_code='5.2.5';
    UPDATE accounts SET account_code='5.2.3.3', parent_id=marketing_id WHERE account_code='5.2.9';

    IF NOT EXISTS (SELECT 1 FROM accounts WHERE account_code='5.2.3.4') THEN
        INSERT INTO accounts (account_code, account_name, account_type, parent_id)
        VALUES ('5.2.3.4', 'Diskon', 'Expense', marketing_id);
    END IF;

    SELECT account_id INTO aset_tetap_id FROM accounts WHERE account_code='1.2';
    UPDATE accounts
        SET account_code='1.2.8', account_type='Asset', parent_id=aset_tetap_id
        WHERE account_code='5.2.8';
END $$;
