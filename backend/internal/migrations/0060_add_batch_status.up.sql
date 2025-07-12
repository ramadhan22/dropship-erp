ALTER TABLE batch_history
    ADD COLUMN status TEXT NOT NULL DEFAULT 'processing',
    ADD COLUMN error_message TEXT;
