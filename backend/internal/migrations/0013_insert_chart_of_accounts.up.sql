-- 0013_insert_chart_of_accounts.up.sql
-- Insert Chart of Accounts in Bahasa Indonesia

-- Root accounts
INSERT INTO accounts (account_code, account_name, account_type)
VALUES
  ('1', 'Aset', 'Asset'),
  ('2', 'Kewajiban', 'Liability'),
  ('3', 'Ekuitas', 'Equity'),
  ('4', 'Pendapatan', 'Revenue'),
  ('5', 'Beban', 'Expense');

-- Aset Lancar
WITH parent AS (
  SELECT account_id AS pid FROM accounts WHERE account_code = '1'
)
INSERT INTO accounts (account_code, account_name, account_type, parent_id)
SELECT '1.1', 'Aset Lancar', 'Asset', pid FROM parent;

-- Sub akun Aset Lancar
WITH parent AS (
  SELECT account_id AS pid FROM accounts WHERE account_code = '1.1'
)
INSERT INTO accounts (account_code, account_name, account_type, parent_id) VALUES
  ('1.1.1', 'Kas', 'Asset', (SELECT pid FROM parent)),
  ('1.1.2', 'Bank A', 'Asset', (SELECT pid FROM parent)),
  ('1.1.3', 'Bank B', 'Asset', (SELECT pid FROM parent)),
  ('1.1.4', 'Bank C', 'Asset', (SELECT pid FROM parent)),
  ('1.1.5', 'Persediaan', 'Asset', (SELECT pid FROM parent)),
  ('1.1.6', 'Sewa Dibayar Dimuka', 'Asset', (SELECT pid FROM parent)),
  ('1.1.7', 'Piutang Dagang', 'Asset', (SELECT pid FROM parent)),
  ('1.1.8', 'Biaya Dibayar Dimuka', 'Asset', (SELECT pid FROM parent));

-- Aset Tetap
WITH parent AS (
  SELECT account_id AS pid FROM accounts WHERE account_code = '1'
)
INSERT INTO accounts (account_code, account_name, account_type, parent_id)
SELECT '1.2', 'Aset Tetap', 'Asset', pid FROM parent;

-- Sub akun Aset Tetap
WITH parent AS (
  SELECT account_id AS pid FROM accounts WHERE account_code = '1.2'
)
INSERT INTO accounts (account_code, account_name, account_type, parent_id) VALUES
  ('1.2.1', 'Tanah', 'Asset', (SELECT pid FROM parent)),
  ('1.2.2', 'Bangunan', 'Asset', (SELECT pid FROM parent)),
  ('1.2.3', 'Akumulasi Penyusutan Bangunan', 'ContraAsset', (SELECT pid FROM parent)),
  ('1.2.4', 'Kendaraan', 'Asset', (SELECT pid FROM parent)),
  ('1.2.5', 'Akumulasi Penyusutan Kendaraan', 'ContraAsset', (SELECT pid FROM parent)),
  ('1.2.6', 'Aset Tak Berwujud', 'Asset', (SELECT pid FROM parent)),
  ('1.2.7', 'Akumulasi Amortisasi Aset Tak Berwujud', 'ContraAsset', (SELECT pid FROM parent));

-- Kewajiban Lancar
WITH parent AS (
  SELECT account_id AS pid FROM accounts WHERE account_code = '2'
)
INSERT INTO accounts (account_code, account_name, account_type, parent_id)
SELECT '2.1', 'Kewajiban Lancar', 'Liability', pid FROM parent;

-- Sub akun Kewajiban Lancar
WITH parent AS (
  SELECT account_id AS pid FROM accounts WHERE account_code = '2.1'
)
INSERT INTO accounts (account_code, account_name, account_type, parent_id) VALUES
  ('2.1.1', 'Utang Dagang', 'Liability', (SELECT pid FROM parent)),
  ('2.1.2', 'Utang Bank (Modal Kerja)', 'Liability', (SELECT pid FROM parent)),
  ('2.1.3', 'Beban Akrual', 'Liability', (SELECT pid FROM parent));

-- Kewajiban Jangka Panjang
WITH parent AS (
  SELECT account_id AS pid FROM accounts WHERE account_code = '2'
)
INSERT INTO accounts (account_code, account_name, account_type, parent_id)
SELECT '2.2', 'Kewajiban Jangka Panjang', 'Liability', pid FROM parent;

-- Sub akun Kewajiban Jangka Panjang
WITH parent AS (
  SELECT account_id AS pid FROM accounts WHERE account_code = '2.2'
)
INSERT INTO accounts (account_code, account_name, account_type, parent_id) VALUES
  ('2.2.1', 'Utang Bank Jangka Panjang', 'Liability', (SELECT pid FROM parent));

-- Ekuitas
WITH parent AS (
  SELECT account_id AS pid FROM accounts WHERE account_code = '3'
)
INSERT INTO accounts (account_code, account_name, account_type, parent_id) VALUES
  ('3.1', 'Modal Disetor', 'Equity', (SELECT pid FROM parent)),
  ('3.2', 'Laba Ditahan', 'Equity', (SELECT pid FROM parent)),
  ('3.3', 'Laba/Rugi Tahun Berjalan', 'Equity', (SELECT pid FROM parent)),
  ('3.4', 'Prive', 'ContraEquity', (SELECT pid FROM parent));

-- Pendapatan
WITH parent AS (
  SELECT account_id AS pid FROM accounts WHERE account_code = '4'
)
INSERT INTO accounts (account_code, account_name, account_type, parent_id) VALUES
  ('4.1', 'Pendapatan Operasional', 'Revenue', (SELECT pid FROM parent)),
  ('4.2', 'Pendapatan Bunga', 'Revenue', (SELECT pid FROM parent)),
  ('4.3', 'Pendapatan Jasa', 'Revenue', (SELECT pid FROM parent));

-- Beban
WITH parent AS (
  SELECT account_id AS pid FROM accounts WHERE account_code = '5'
)
INSERT INTO accounts (account_code, account_name, account_type, parent_id) VALUES
  ('5.1', 'Harga Pokok Penjualan', 'Expense', (SELECT pid FROM parent)),
  ('5.2', 'Beban Usaha', 'Expense', (SELECT pid FROM parent)),
  ('5.3', 'Beban Administrasi', 'Expense', (SELECT pid FROM parent)),
  ('5.4', 'Beban Pajak Penghasilan', 'Expense', (SELECT pid FROM parent));

-- Sub akun Beban Usaha
WITH parent AS (
  SELECT account_id AS pid FROM accounts WHERE account_code = '5.2'
)
INSERT INTO accounts (account_code, account_name, account_type, parent_id) VALUES
  ('5.2.1', 'Loan Fee', 'Expense', (SELECT pid FROM parent)),
  ('5.2.2', 'Gaji', 'Expense', (SELECT pid FROM parent)),
  ('5.2.3', 'Voucher', 'Expense', (SELECT pid FROM parent)),
  ('5.2.4', 'Biaya Layanan Ecommerce', 'Expense', (SELECT pid FROM parent)),
  ('5.2.5', 'Biaya Affiliate', 'Expense', (SELECT pid FROM parent)),
  ('5.2.6', 'Beban Iklan', 'Expense', (SELECT pid FROM parent)),
  ('5.2.7', 'Beban Mitra', 'Expense', (SELECT pid FROM parent)),
  ('5.2.8', 'Pembelian Aset (CapEx)', 'Expense', (SELECT pid FROM parent));

-- Sub akun Beban Administrasi
WITH parent AS (
  SELECT account_id AS pid FROM accounts WHERE account_code = '5.3'
)
INSERT INTO accounts (account_code, account_name, account_type, parent_id) VALUES
  ('5.3.1', 'Gaji Administrasi', 'Expense', (SELECT pid FROM parent)),
  ('5.3.2', 'Asuransi', 'Expense', (SELECT pid FROM parent));
