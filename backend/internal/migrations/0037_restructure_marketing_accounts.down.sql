DO $$
DECLARE
    beban_usaha_id INT;
    aset_tetap_id INT;
BEGIN
    SELECT account_id INTO beban_usaha_id FROM accounts WHERE account_code='5.2';
    SELECT account_id INTO aset_tetap_id FROM accounts WHERE account_code='1.2';

    UPDATE accounts SET account_code='5.2.3', parent_id=beban_usaha_id WHERE account_code='5.2.3.1';
    UPDATE accounts SET account_code='5.2.5', parent_id=beban_usaha_id WHERE account_code='5.2.3.2';
    UPDATE accounts SET account_code='5.2.9', parent_id=beban_usaha_id WHERE account_code='5.2.3.3';
    DELETE FROM accounts WHERE account_code='5.2.3.4';
    DELETE FROM accounts WHERE account_code='5.2.3';

    UPDATE accounts
        SET account_code='5.2.8', account_type='Expense', parent_id=beban_usaha_id
        WHERE account_code='1.2.8';
END $$;
