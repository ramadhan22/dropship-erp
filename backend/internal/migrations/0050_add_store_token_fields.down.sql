-- 0050_add_store_token_fields.down.sql
ALTER TABLE stores DROP COLUMN IF EXISTS access_token;
ALTER TABLE stores DROP COLUMN IF EXISTS refresh_token;
ALTER TABLE stores DROP COLUMN IF EXISTS expire_in;
ALTER TABLE stores DROP COLUMN IF EXISTS request_id;
