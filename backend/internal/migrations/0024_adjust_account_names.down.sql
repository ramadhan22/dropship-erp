-- 0024_adjust_account_names.down.sql
-- Revert expense account name changes

UPDATE accounts
SET account_name = 'Beban Iklan'
WHERE account_code = '5.2.6';

UPDATE accounts
SET account_name = 'Biaya Layanan Ecommerce'
WHERE account_code = '5.2.4';
