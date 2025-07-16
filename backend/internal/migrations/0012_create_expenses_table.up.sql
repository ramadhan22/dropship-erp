CREATE TABLE expenses (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  date TIMESTAMP NOT NULL,
  description TEXT NOT NULL,
  amount NUMERIC NOT NULL,
  account_id INT NOT NULL REFERENCES accounts(account_id),
  created_at TIMESTAMP NOT NULL DEFAULT now()
);