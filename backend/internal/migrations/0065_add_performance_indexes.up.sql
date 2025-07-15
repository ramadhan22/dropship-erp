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

-- Index for journal lines by account ID (used in account balances)
CREATE INDEX IF NOT EXISTS idx_journal_lines_account_id 
ON journal_lines (account_id);

-- Composite index for journal lines filtering
CREATE INDEX IF NOT EXISTS idx_journal_lines_composite 
ON journal_lines (journal_id, account_id, amount, is_debit);

-- Index for shopee settled orders by order ID (used in reconciliation)
CREATE INDEX IF NOT EXISTS idx_shopee_settled_orders_order_id 
ON shopee_settled_orders (order_id);

-- Index for dropship purchase details by kode pesanan (used in product analysis)
CREATE INDEX IF NOT EXISTS idx_dropship_purchase_details_kode 
ON dropship_purchase_details (kode_pesanan);

-- Index for expenses by date (used in expense reporting)
CREATE INDEX IF NOT EXISTS idx_expenses_date 
ON expenses (date DESC);

-- Index for batch history by process type and started time (used in monitoring)
CREATE INDEX IF NOT EXISTS idx_batch_history_process_started 
ON batch_history (process_type, started_at DESC);