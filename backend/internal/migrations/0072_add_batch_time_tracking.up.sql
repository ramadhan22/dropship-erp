ALTER TABLE batch_history
    ADD COLUMN ended_at TIMESTAMPTZ,
    ADD COLUMN time_spent INTERVAL;