-- 0018_set_jakmall_account_id.down.sql
-- Revert account_id change for 'Saldo Jakmall'
DO $$
DECLARE
    seq_id int;
BEGIN
    seq_id := nextval('accounts_account_id_seq');

    UPDATE journal_lines SET account_id=seq_id WHERE account_id=11009;
    UPDATE expenses SET account_id=seq_id WHERE account_id=11009;
    UPDATE accounts SET account_id=seq_id WHERE account_code='1.1.9';

    PERFORM setval('accounts_account_id_seq', (SELECT MAX(account_id) FROM accounts));
END $$;
