-- Drop foreign keys prior to rolling back account ID changes
ALTER TABLE IF EXISTS accounts
  DROP CONSTRAINT IF EXISTS accounts_parent_id_fkey;
ALTER TABLE IF EXISTS journal_lines
  DROP CONSTRAINT IF EXISTS journal_lines_account_id_fkey;
ALTER TABLE IF EXISTS expenses
  DROP CONSTRAINT IF EXISTS expenses_account_id_fkey;
ALTER TABLE IF EXISTS expense_lines
  DROP CONSTRAINT IF EXISTS expense_lines_account_id_fkey;
ALTER TABLE IF EXISTS asset_accounts
  DROP CONSTRAINT IF EXISTS asset_accounts_account_id_fkey;

DO $$
DECLARE
    cur_id INT;
    seq_id INT;
BEGIN
    SELECT account_id INTO cur_id FROM accounts WHERE account_code='1.1.14';
    IF cur_id IS NULL THEN
        RETURN;
    END IF;
    IF cur_id = 11014 THEN
        seq_id := nextval('accounts_account_id_seq');
        UPDATE journal_lines SET account_id=seq_id WHERE account_id=cur_id;
        UPDATE expenses SET asset_account_id=seq_id WHERE asset_account_id=cur_id;
        UPDATE asset_accounts SET account_id=seq_id WHERE account_id=cur_id;
        UPDATE accounts SET account_id=seq_id WHERE account_code='1.1.14';
        PERFORM setval('accounts_account_id_seq', (SELECT MAX(account_id) FROM accounts));
    END IF;
END $$;
