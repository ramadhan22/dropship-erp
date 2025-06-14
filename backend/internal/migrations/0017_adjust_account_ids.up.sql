-- 0017_adjust_account_ids.up.sql
-- Update account_id values to match journal entry expectations
DO $$
DECLARE
    cash_old int;
    rev_old int;
    cogs_old int;
    shopee_p_old int;
    shopee_b_old int;
    barista_p_old int;
    barista_b_old int;
BEGIN
    SELECT account_id INTO cash_old FROM accounts WHERE account_code='1.1.1';
    SELECT account_id INTO rev_old FROM accounts WHERE account_code='4.1';
    SELECT account_id INTO cogs_old FROM accounts WHERE account_code='5.1';
    SELECT account_id INTO shopee_p_old FROM accounts WHERE account_code='1.1.10';
    SELECT account_id INTO shopee_b_old FROM accounts WHERE account_code='1.1.11';
    SELECT account_id INTO barista_p_old FROM accounts WHERE account_code='1.1.12';
    SELECT account_id INTO barista_b_old FROM accounts WHERE account_code='1.1.13';

    UPDATE journal_lines SET account_id=1001 WHERE account_id=cash_old;
    UPDATE journal_lines SET account_id=4001 WHERE account_id=rev_old;
    UPDATE journal_lines SET account_id=5001 WHERE account_id=cogs_old;
    UPDATE journal_lines SET account_id=11010 WHERE account_id=shopee_p_old;
    UPDATE journal_lines SET account_id=11011 WHERE account_id=shopee_b_old;
    UPDATE journal_lines SET account_id=11012 WHERE account_id=barista_p_old;
    UPDATE journal_lines SET account_id=11013 WHERE account_id=barista_b_old;

    UPDATE expenses SET account_id=1001 WHERE account_id=cash_old;
    UPDATE expenses SET account_id=4001 WHERE account_id=rev_old;
    UPDATE expenses SET account_id=5001 WHERE account_id=cogs_old;
    UPDATE expenses SET account_id=11010 WHERE account_id=shopee_p_old;
    UPDATE expenses SET account_id=11011 WHERE account_id=shopee_b_old;
    UPDATE expenses SET account_id=11012 WHERE account_id=barista_p_old;
    UPDATE expenses SET account_id=11013 WHERE account_id=barista_b_old;

    UPDATE accounts SET account_id=1001 WHERE account_code='1.1.1';
    UPDATE accounts SET account_id=4001 WHERE account_code='4.1';
    UPDATE accounts SET account_id=5001 WHERE account_code='5.1';
    UPDATE accounts SET account_id=11010 WHERE account_code='1.1.10';
    UPDATE accounts SET account_id=11011 WHERE account_code='1.1.11';
    UPDATE accounts SET account_id=11012 WHERE account_code='1.1.12';
    UPDATE accounts SET account_id=11013 WHERE account_code='1.1.13';

    PERFORM setval('accounts_account_id_seq', (SELECT MAX(account_id) FROM accounts));
END $$;
