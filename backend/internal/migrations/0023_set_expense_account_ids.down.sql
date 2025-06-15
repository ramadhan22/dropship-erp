-- 0023_set_expense_account_ids.down.sql
-- Revert account_id changes for expense accounts
DO $$
DECLARE
    seq_id int;
BEGIN
    -- Voucher
    seq_id := nextval('accounts_account_id_seq');
    UPDATE journal_lines SET account_id=seq_id WHERE account_id=52003;
    UPDATE expenses SET account_id=seq_id WHERE account_id=52003;
    UPDATE accounts SET account_id=seq_id WHERE account_code='5.2.3';

    -- Biaya Layanan Ecommerce
    seq_id := nextval('accounts_account_id_seq');
    UPDATE journal_lines SET account_id=seq_id WHERE account_id=52004;
    UPDATE expenses SET account_id=seq_id WHERE account_id=52004;
    UPDATE accounts SET account_id=seq_id WHERE account_code='5.2.4';

    -- Biaya Affiliate
    seq_id := nextval('accounts_account_id_seq');
    UPDATE journal_lines SET account_id=seq_id WHERE account_id=52005;
    UPDATE expenses SET account_id=seq_id WHERE account_id=52005;
    UPDATE accounts SET account_id=seq_id WHERE account_code='5.2.5';

    -- Beban Iklan/Administrasi
    seq_id := nextval('accounts_account_id_seq');
    UPDATE journal_lines SET account_id=seq_id WHERE account_id=52006;
    UPDATE expenses SET account_id=seq_id WHERE account_id=52006;
    UPDATE accounts SET account_id=seq_id WHERE account_code='5.2.6';

    PERFORM setval('accounts_account_id_seq', (SELECT MAX(account_id) FROM accounts));
END $$;
