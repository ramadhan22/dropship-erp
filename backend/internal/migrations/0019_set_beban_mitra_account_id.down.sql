-- 0019_set_beban_mitra_account_id.down.sql
-- Revert account_id change for 'Beban Mitra'
DO $$
DECLARE
    seq_id int;
BEGIN
    seq_id := nextval('accounts_account_id_seq');

    UPDATE journal_lines SET account_id=seq_id WHERE account_id=52007;
    UPDATE expenses SET account_id=seq_id WHERE account_id=52007;
    UPDATE accounts SET account_id=seq_id WHERE account_code='5.2.7';

    PERFORM setval('accounts_account_id_seq', (SELECT MAX(account_id) FROM accounts));
END $$;
