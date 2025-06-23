ALTER TABLE shopee_settled
  DROP COLUMN IF EXISTS is_data_mismatch,
  DROP COLUMN IF EXISTS is_settled_confirmed;
