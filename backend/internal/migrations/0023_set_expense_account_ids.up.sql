-- 0023_set_expense_account_ids.up.sql
-- Assign fixed account_id values for expense accounts used in Shopee journal entries
DO $$
DECLARE
    voucher_old int;
    fee_old int;
    affiliate_old int;
    admin_old int;
BEGIN
    SELECT account_id INTO voucher_old FROM accounts WHERE account_code='5.2.3';
    SELECT account_id INTO fee_old FROM accounts WHERE account_code='5.2.4';
    SELECT account_id INTO affiliate_old FROM accounts WHERE account_code='5.2.5';
    SELECT account_id INTO admin_old FROM accounts WHERE account_code='5.2.6';

    UPDATE journal_lines SET account_id=52003 WHERE account_id=voucher_old;
    UPDATE journal_lines SET account_id=52004 WHERE account_id=fee_old;
    UPDATE journal_lines SET account_id=52005 WHERE account_id=affiliate_old;
    UPDATE journal_lines SET account_id=52006 WHERE account_id=admin_old;

    UPDATE expenses SET account_id=52003 WHERE account_id=voucher_old;
    UPDATE expenses SET account_id=52004 WHERE account_id=fee_old;
    UPDATE expenses SET account_id=52005 WHERE account_id=affiliate_old;
    UPDATE expenses SET account_id=52006 WHERE account_id=admin_old;

    UPDATE accounts SET account_id=52003 WHERE account_code='5.2.3';
    UPDATE accounts SET account_id=52004 WHERE account_code='5.2.4';
    UPDATE accounts SET account_id=52005 WHERE account_code='5.2.5';
    UPDATE accounts SET account_id=52006 WHERE account_code='5.2.6';

    PERFORM setval('accounts_account_id_seq', (SELECT MAX(account_id) FROM accounts));
END $$;
