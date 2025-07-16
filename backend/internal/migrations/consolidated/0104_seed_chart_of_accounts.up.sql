-- Consolidated Chart of Accounts Seed Data
-- Replaces account-related data from migrations: 0013, 0014, 0017, 0018, 0019, 0023, 0024, 0026, 0032, 0035, 0037, 0038, 0039, 0043, 0045, 0046, 0048, 0052, 0054, 0058, 0064

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
  ('1.2.6', 'Peralatan', 'Asset', (SELECT pid FROM parent)),
  ('1.2.7', 'Akumulasi Penyusutan Peralatan', 'ContraAsset', (SELECT pid FROM parent));

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
  ('2.1.2', 'Utang Pajak', 'Liability', (SELECT pid FROM parent)),
  ('2.1.3', 'Utang Gaji', 'Liability', (SELECT pid FROM parent)),
  ('2.1.4', 'Utang Bunga', 'Liability', (SELECT pid FROM parent));

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
  ('2.2.1', 'Utang Bank', 'Liability', (SELECT pid FROM parent)),
  ('2.2.2', 'Utang Obligasi', 'Liability', (SELECT pid FROM parent));

-- Ekuitas
WITH parent AS (
  SELECT account_id AS pid FROM accounts WHERE account_code = '3'
)
INSERT INTO accounts (account_code, account_name, account_type, parent_id) VALUES
  ('3.1', 'Modal Saham', 'Equity', (SELECT pid FROM parent)),
  ('3.2', 'Laba Ditahan', 'Equity', (SELECT pid FROM parent)),
  ('3.3', 'Laba Tahun Berjalan', 'Equity', (SELECT pid FROM parent));

-- Pendapatan
WITH parent AS (
  SELECT account_id AS pid FROM accounts WHERE account_code = '4'
)
INSERT INTO accounts (account_code, account_name, account_type, parent_id) VALUES
  ('4.1', 'Pendapatan Operasional', 'Revenue', (SELECT pid FROM parent)),
  ('4.2', 'Pendapatan Non-Operasional', 'Revenue', (SELECT pid FROM parent));

-- Sub akun Pendapatan
WITH parent AS (
  SELECT account_id AS pid FROM accounts WHERE account_code = '4.1'
)
INSERT INTO accounts (account_code, account_name, account_type, parent_id) VALUES
  ('4.1.1', 'Penjualan', 'Revenue', (SELECT pid FROM parent)),
  ('4.1.2', 'Retur Penjualan', 'ContraRevenue', (SELECT pid FROM parent)),
  ('4.1.3', 'Potongan Penjualan', 'ContraRevenue', (SELECT pid FROM parent));

-- Beban
WITH parent AS (
  SELECT account_id AS pid FROM accounts WHERE account_code = '5'
)
INSERT INTO accounts (account_code, account_name, account_type, parent_id) VALUES
  ('5.1', 'Beban Operasional', 'Expense', (SELECT pid FROM parent)),
  ('5.2', 'Beban Non-Operasional', 'Expense', (SELECT pid FROM parent));

-- Sub akun Beban Operasional
WITH parent AS (
  SELECT account_id AS pid FROM accounts WHERE account_code = '5.1'
)
INSERT INTO accounts (account_code, account_name, account_type, parent_id) VALUES
  ('5.1.1', 'Harga Pokok Penjualan', 'Expense', (SELECT pid FROM parent)),
  ('5.1.2', 'Beban Gaji', 'Expense', (SELECT pid FROM parent)),
  ('5.1.3', 'Beban Sewa', 'Expense', (SELECT pid FROM parent)),
  ('5.1.4', 'Beban Listrik', 'Expense', (SELECT pid FROM parent)),
  ('5.1.5', 'Beban Telepon', 'Expense', (SELECT pid FROM parent)),
  ('5.1.6', 'Beban Pemasaran', 'Expense', (SELECT pid FROM parent)),
  ('5.1.7', 'Beban Administrasi', 'Expense', (SELECT pid FROM parent)),
  ('5.1.8', 'Beban Penyusutan', 'Expense', (SELECT pid FROM parent));

-- Platform-specific accounts (from migration 0014)
INSERT INTO accounts (account_code, account_name, account_type, parent_id)
VALUES
  ('1.1.9', 'Bank Jakmall', 'Asset', (SELECT account_id FROM accounts WHERE account_code = '1.1')),
  ('5.1.9', 'Beban Mitra Jakmall', 'Expense', (SELECT account_id FROM accounts WHERE account_code = '5.1'));

-- Additional expense accounts (from various migrations)
INSERT INTO accounts (account_code, account_name, account_type, parent_id)
VALUES
  ('5.1.10', 'Biaya Iklan', 'Expense', (SELECT account_id FROM accounts WHERE account_code = '5.1')),
  ('1.1.10', 'Bank SeaBank', 'Asset', (SELECT account_id FROM accounts WHERE account_code = '1.1')),
  ('4.1.4', 'Diskon Produk', 'ContraRevenue', (SELECT account_id FROM accounts WHERE account_code = '4.1')),
  ('4.1.5', 'Refund', 'ContraRevenue', (SELECT account_id FROM accounts WHERE account_code = '4.1')),
  ('5.1.11', 'Selisih Ongkir', 'Expense', (SELECT account_id FROM accounts WHERE account_code = '5.1')),
  ('2.1.5', 'Pajak UMKM', 'Liability', (SELECT account_id FROM accounts WHERE account_code = '2.1')),
  ('5.1.12', 'Biaya Transaksi', 'Expense', (SELECT account_id FROM accounts WHERE account_code = '5.1')),
  ('4.1.6', 'Diskon Ongkir', 'ContraRevenue', (SELECT account_id FROM accounts WHERE account_code = '4.1')),
  ('5.1.13', 'Free Sample', 'Expense', (SELECT account_id FROM accounts WHERE account_code = '5.1'));