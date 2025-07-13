ALTER TABLE batch_history
    DROP COLUMN IF EXISTS file_name,
    DROP COLUMN IF EXISTS file_path;
