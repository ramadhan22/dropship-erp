DROP TABLE IF EXISTS expense_lines;
ALTER TABLE expenses RENAME COLUMN asset_account_id TO account_id;
