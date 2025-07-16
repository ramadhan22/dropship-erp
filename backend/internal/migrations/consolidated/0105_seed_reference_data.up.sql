-- Consolidated Reference Data Seed
-- Replaces reference data from migrations: 0029, 0032, 0033, 0034, 0040

-- Asset Accounts (from migration 0029, 0032, 0033)
INSERT INTO asset_accounts (account_id, account_name, balance)
VALUES
  ((SELECT account_id FROM accounts WHERE account_code = '1.1.1'), 'Kas', 0),
  ((SELECT account_id FROM accounts WHERE account_code = '1.1.2'), 'Bank A', 0),
  ((SELECT account_id FROM accounts WHERE account_code = '1.1.3'), 'Bank B', 0),
  ((SELECT account_id FROM accounts WHERE account_code = '1.1.4'), 'Bank C', 0),
  ((SELECT account_id FROM accounts WHERE account_code = '1.1.9'), 'Bank Jakmall', 0),
  ((SELECT account_id FROM accounts WHERE account_code = '1.1.10'), 'Bank SeaBank', 0);

-- Channel Types (from migration 0040)
INSERT INTO jenis_channels (jenis_channel)
VALUES ('Shopee');

-- Stores (from migration 0040)
INSERT INTO stores (jenis_channel_id, nama_toko)
VALUES 
  ((SELECT jenis_channel_id FROM jenis_channels WHERE jenis_channel = 'Shopee'), 'dropship.id'),
  ((SELECT jenis_channel_id FROM jenis_channels WHERE jenis_channel = 'Shopee'), 'kedaitoko22');