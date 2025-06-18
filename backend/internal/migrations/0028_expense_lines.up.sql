ALTER TABLE expenses RENAME COLUMN account_id TO asset_account_id;

CREATE TABLE expense_lines (
  line_id SERIAL PRIMARY KEY,
  expense_id UUID NOT NULL REFERENCES expenses(id) ON DELETE CASCADE,
  account_id INT NOT NULL REFERENCES accounts(account_id),
  amount NUMERIC NOT NULL
);
