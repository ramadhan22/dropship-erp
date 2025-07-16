CREATE TABLE failed_reconciliations (
    id BIGSERIAL PRIMARY KEY,
    purchase_id VARCHAR(255) NOT NULL,
    order_id VARCHAR(255),
    shop VARCHAR(255) NOT NULL,
    error_type VARCHAR(100) NOT NULL,
    error_message TEXT NOT NULL,
    context TEXT,
    failed_at TIMESTAMP NOT NULL DEFAULT NOW(),
    retried BOOLEAN NOT NULL DEFAULT FALSE,
    retried_at TIMESTAMP,
    batch_id BIGINT REFERENCES batch_history(id)
);

-- Add indexes for better query performance
CREATE INDEX idx_failed_reconciliations_shop ON failed_reconciliations(shop);
CREATE INDEX idx_failed_reconciliations_error_type ON failed_reconciliations(error_type);
CREATE INDEX idx_failed_reconciliations_failed_at ON failed_reconciliations(failed_at);
CREATE INDEX idx_failed_reconciliations_retried ON failed_reconciliations(retried);
CREATE INDEX idx_failed_reconciliations_batch_id ON failed_reconciliations(batch_id);