-- 0005_create_journal_lines.up.sql

CREATE TABLE IF NOT EXISTS journal_lines (
  line_id    SERIAL PRIMARY KEY,
  journal_id INT NOT NULL REFERENCES journal_entries(journal_id) ON DELETE CASCADE,
  account_id INT NOT NULL REFERENCES accounts(account_id),
  is_debit   BOOLEAN NOT NULL,
  amount     NUMERIC(14,2) NOT NULL,
  memo       TEXT
);
