-- 0016_drop_foreign_key_constraints.down.sql
ALTER TABLE accounts
  ADD CONSTRAINT accounts_parent_id_fkey FOREIGN KEY(parent_id)
    REFERENCES accounts(account_id);

ALTER TABLE journal_lines
  ADD CONSTRAINT journal_lines_journal_id_fkey FOREIGN KEY(journal_id)
    REFERENCES journal_entries(journal_id) ON DELETE CASCADE;
ALTER TABLE journal_lines
  ADD CONSTRAINT journal_lines_account_id_fkey FOREIGN KEY(account_id)
    REFERENCES accounts(account_id);

ALTER TABLE dropship_purchase_details
  ADD CONSTRAINT dropship_purchase_details_kode_pesanan_fkey FOREIGN KEY(kode_pesanan)
    REFERENCES dropship_purchases(kode_pesanan) ON DELETE CASCADE;

ALTER TABLE stores
  ADD CONSTRAINT stores_jenis_channel_id_fkey FOREIGN KEY(jenis_channel_id)
    REFERENCES jenis_channels(jenis_channel_id) ON DELETE CASCADE;

ALTER TABLE expenses
  ADD CONSTRAINT expenses_account_id_fkey FOREIGN KEY(account_id)
    REFERENCES accounts(account_id);
