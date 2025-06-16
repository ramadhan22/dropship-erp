-- 0024_adjust_account_names.up.sql
-- Rename expense account names for Shopee fees

UPDATE accounts
SET account_name = 'Biaya Administrasi Shopee'
WHERE account_code = '5.2.6';

UPDATE accounts
SET account_name = 'Biaya Layanan Shopee'
WHERE account_code = '5.2.4';
