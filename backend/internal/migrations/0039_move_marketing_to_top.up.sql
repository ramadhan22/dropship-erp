DO $$
DECLARE
    marketing_id INT;
    beban_id INT;
BEGIN
    SELECT account_id INTO marketing_id FROM accounts WHERE account_code='5.2.3';
    SELECT account_id INTO beban_id FROM accounts WHERE account_code='5';
    IF marketing_id IS NULL OR beban_id IS NULL THEN
        RETURN;
    END IF;
    UPDATE accounts SET account_code='5.5', parent_id=beban_id WHERE account_id=marketing_id;
    UPDATE accounts SET account_code='5.5.1' WHERE account_code='5.2.3.1';
    UPDATE accounts SET account_code='5.5.2' WHERE account_code='5.2.3.2';
    UPDATE accounts SET account_code='5.5.3' WHERE account_code='5.2.3.3';
    UPDATE accounts SET account_code='5.5.4' WHERE account_code='5.2.3.4';
END $$;
