-- 0016_drop_foreign_key_constraints.up.sql
ALTER TABLE accounts
  DROP CONSTRAINT IF EXISTS accounts_parent_id_fkey;

ALTER TABLE journal_lines
  DROP CONSTRAINT IF EXISTS journal_lines_journal_id_fkey,
  DROP CONSTRAINT IF EXISTS journal_lines_account_id_fkey;

ALTER TABLE dropship_purchase_details
  DROP CONSTRAINT IF EXISTS dropship_purchase_details_kode_pesanan_fkey;

ALTER TABLE stores
  DROP CONSTRAINT IF EXISTS stores_jenis_channel_id_fkey;

ALTER TABLE expenses
  DROP CONSTRAINT IF EXISTS expenses_account_id_fkey;
