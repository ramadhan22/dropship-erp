ALTER TABLE batch_history
    DROP COLUMN IF EXISTS ended_at,
    DROP COLUMN IF EXISTS time_spent;