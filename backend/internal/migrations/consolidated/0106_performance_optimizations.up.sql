-- Consolidated Performance Optimizations
-- Replaces performance migrations: 0016, 0061, 0065

-- Drop foreign key constraints for performance (from 0016)
ALTER TABLE accounts DROP CONSTRAINT IF EXISTS accounts_parent_id_fkey;
ALTER TABLE journal_lines DROP CONSTRAINT IF EXISTS journal_lines_account_id_fkey;
ALTER TABLE asset_accounts DROP CONSTRAINT IF EXISTS asset_accounts_account_id_fkey;
ALTER TABLE expense_lines DROP CONSTRAINT IF EXISTS expense_lines_account_id_fkey;
ALTER TABLE dropship_purchase_details DROP CONSTRAINT IF EXISTS dropship_purchase_details_kode_pesanan_fkey;
ALTER TABLE stores DROP CONSTRAINT IF EXISTS stores_jenis_channel_id_fkey;
ALTER TABLE expenses DROP CONSTRAINT IF EXISTS expenses_account_id_fkey;

-- Performance indexes (from 0061, 0065)
CREATE INDEX IF NOT EXISTS idx_journal_entries_source ON journal_entries(source_type, source_id);
CREATE INDEX IF NOT EXISTS idx_journal_lines_account ON journal_lines(account_id);
CREATE INDEX IF NOT EXISTS idx_journal_lines_journal ON journal_lines(journal_id);
CREATE INDEX IF NOT EXISTS idx_dropship_purchases_status ON dropship_purchases(status_pesanan_terakhir);
CREATE INDEX IF NOT EXISTS idx_dropship_purchases_toko ON dropship_purchases(nama_toko);
CREATE INDEX IF NOT EXISTS idx_dropship_purchases_channel ON dropship_purchases(jenis_channel);
CREATE INDEX IF NOT EXISTS idx_dropship_purchase_details_kode ON dropship_purchase_details(kode_pesanan);
CREATE INDEX IF NOT EXISTS idx_expenses_date ON expenses(expense_date);
CREATE INDEX IF NOT EXISTS idx_expenses_asset_account ON expenses(asset_account_id);
CREATE INDEX IF NOT EXISTS idx_shopee_settled_orders_order_id ON shopee_settled_orders(order_id);
CREATE INDEX IF NOT EXISTS idx_shopee_settled_orders_date ON shopee_settled_orders(settled_date);
CREATE INDEX IF NOT EXISTS idx_shopee_settled_no_pesanan ON shopee_settled(no_pesanan);
CREATE INDEX IF NOT EXISTS idx_shopee_settled_date ON shopee_settled(tanggal_dana_dilepaskan);
CREATE INDEX IF NOT EXISTS idx_shopee_order_details_toko ON shopee_order_details(nama_toko);
CREATE INDEX IF NOT EXISTS idx_shopee_order_details_status ON shopee_order_details(status);
CREATE INDEX IF NOT EXISTS idx_accounts_code ON accounts(account_code);
CREATE INDEX IF NOT EXISTS idx_accounts_type ON accounts(account_type);
CREATE INDEX IF NOT EXISTS idx_accounts_parent ON accounts(parent_id);