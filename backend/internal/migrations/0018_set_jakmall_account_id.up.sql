-- 0018_set_jakmall_account_id.up.sql
-- Assign fixed account_id for 'Saldo Jakmall'
DO $$
DECLARE
    jakmall_old int;
BEGIN
    SELECT account_id INTO jakmall_old FROM accounts WHERE account_code='1.1.9';

    UPDATE journal_lines SET account_id=11009 WHERE account_id=jakmall_old;
    UPDATE expenses SET account_id=11009 WHERE account_id=jakmall_old;
    UPDATE accounts SET account_id=11009 WHERE account_code='1.1.9';

    PERFORM setval('accounts_account_id_seq', (SELECT MAX(account_id) FROM accounts));
END $$;
