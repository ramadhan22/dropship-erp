DO $$
DECLARE
    marketing_id INT;
    op_id INT;
BEGIN
    SELECT account_id INTO marketing_id FROM accounts WHERE account_code='5.5';
    SELECT account_id INTO op_id FROM accounts WHERE account_code='5.2';
    IF marketing_id IS NULL OR op_id IS NULL THEN
        RETURN;
    END IF;
    UPDATE accounts SET account_code='5.2.3', parent_id=op_id WHERE account_id=marketing_id;
    UPDATE accounts SET account_code='5.2.3.1' WHERE account_code='5.5.1';
    UPDATE accounts SET account_code='5.2.3.2' WHERE account_code='5.5.2';
    UPDATE accounts SET account_code='5.2.3.3' WHERE account_code='5.5.3';
    UPDATE accounts SET account_code='5.2.3.4' WHERE account_code='5.5.4';
END $$;
