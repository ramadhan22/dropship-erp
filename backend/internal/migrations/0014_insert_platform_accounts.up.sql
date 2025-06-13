-- 0014_insert_platform_accounts.up.sql
-- Menambahkan akun khusus platform dan saldo pending/settled untuk Shopee & Jakmall

WITH parent AS (
  SELECT account_id AS pid FROM accounts WHERE account_code = '1.1'
)
INSERT INTO accounts (account_code, account_name, account_type, parent_id) VALUES
  ('1.1.9',  'Saldo Jakmall',            'Asset', (SELECT pid FROM parent)),
  ('1.1.10', 'MR eStore Shopee Pending', 'Asset', (SELECT pid FROM parent)),
  ('1.1.11', 'MR eStore Shopee Balance', 'Asset', (SELECT pid FROM parent)),
  ('1.1.12', 'MR Barista Gear Pending',  'Asset', (SELECT pid FROM parent)),
  ('1.1.13', 'MR Barista Gear Balance',  'Asset', (SELECT pid FROM parent));
