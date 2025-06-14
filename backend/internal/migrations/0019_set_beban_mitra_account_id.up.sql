-- 0019_set_beban_mitra_account_id.up.sql
-- Assign fixed account_id for 'Beban Mitra'
DO $$
DECLARE
    mitra_old int;
BEGIN
    SELECT account_id INTO mitra_old FROM accounts WHERE account_code='5.2.7';

    UPDATE journal_lines SET account_id=52007 WHERE account_id=mitra_old;
    UPDATE expenses SET account_id=52007 WHERE account_id=mitra_old;
    UPDATE accounts SET account_id=52007 WHERE account_code='5.2.7';

    PERFORM setval('accounts_account_id_seq', (SELECT MAX(account_id) FROM accounts));
END $$;
