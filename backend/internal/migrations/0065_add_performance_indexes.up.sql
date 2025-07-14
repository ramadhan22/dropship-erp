-- File: 0065_add_performance_indexes.up.sql
-- Performance optimization indexes for faster queries

-- Composite index for dropship purchases filtering (used in dashboard and reporting)
CREATE INDEX IF NOT EXISTS idx_dropship_purchases_composite 
ON dropship_purchases (jenis_channel, nama_toko, waktu_pesanan_terbuat DESC);

-- Index for dropship purchases by invoice channel (used in reconciliation)
CREATE INDEX IF NOT EXISTS idx_dropship_purchases_invoice_channel 
ON dropship_purchases (kode_invoice_channel);

-- Partial index for pending reconciliation (optimizes dashboard loading)
CREATE INDEX IF NOT EXISTS idx_purchases_pending_reconcile 
ON dropship_purchases (kode_invoice_channel, status_pesanan_terakhir) 
WHERE status_pesanan_terakhir != 'pesanan selesai';

-- Index for journal entries by creation date (used in GL and P&L reports)
CREATE INDEX IF NOT EXISTS idx_journal_entries_created_at 
ON journal_entries (created_at DESC);

-- Index for journal lines by account code (used in account balances)
CREATE INDEX IF NOT EXISTS idx_journal_lines_account_code 
ON journal_lines (account_code);

-- Composite index for journal lines filtering
CREATE INDEX IF NOT EXISTS idx_journal_lines_composite 
ON journal_lines (journal_entry_id, account_code, debit_amount, credit_amount);

-- Index for shopee settled orders by order number (used in reconciliation)
CREATE INDEX IF NOT EXISTS idx_shopee_settled_orders_no 
ON shopee_settled_orders (no_pesanan);

-- Index for dropship purchase details by kode pesanan (used in product analysis)
CREATE INDEX IF NOT EXISTS idx_dropship_purchase_details_kode 
ON dropship_purchase_details (kode_pesanan);

-- Index for expenses by date (used in expense reporting)
CREATE INDEX IF NOT EXISTS idx_expenses_tanggal 
ON expenses (tanggal DESC);

-- Index for batch history by status and process type (used in monitoring)
CREATE INDEX IF NOT EXISTS idx_batch_history_status_type 
ON batch_history (status, process_type, created_at DESC);