-- 0050_add_store_token_fields.up.sql
ALTER TABLE stores ADD COLUMN IF NOT EXISTS access_token TEXT;
ALTER TABLE stores ADD COLUMN IF NOT EXISTS refresh_token TEXT;
ALTER TABLE stores ADD COLUMN IF NOT EXISTS expire_in INT;
ALTER TABLE stores ADD COLUMN IF NOT EXISTS request_id TEXT;
