-- File: 0065_add_performance_indexes.down.sql
-- Remove performance optimization indexes

DROP INDEX IF EXISTS idx_dropship_purchases_composite;
DROP INDEX IF EXISTS idx_dropship_purchases_invoice_channel;
DROP INDEX IF EXISTS idx_purchases_pending_reconcile;
DROP INDEX IF EXISTS idx_journal_entries_created_at;
DROP INDEX IF EXISTS idx_journal_lines_account_id;
DROP INDEX IF EXISTS idx_journal_lines_composite;
DROP INDEX IF EXISTS idx_shopee_settled_orders_order_id;
DROP INDEX IF EXISTS idx_dropship_purchase_details_kode;
DROP INDEX IF EXISTS idx_expenses_date;
DROP INDEX IF EXISTS idx_batch_history_process_started;