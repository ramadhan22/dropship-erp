ALTER TABLE batch_history
    DROP COLUMN IF EXISTS status,
    DROP COLUMN IF EXISTS error_message;
