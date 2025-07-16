-- Consolidated Analytics and Tracking Tables Migration
-- Replaces migrations: 0025, 0047, 0059, 0060, 0062, 0063, 0066, 0067

CREATE TABLE IF NOT EXISTS ad_invoices (
  id               SERIAL PRIMARY KEY,
  invoice_number   VARCHAR(100) NOT NULL UNIQUE,
  invoice_date     DATE NOT NULL,
  due_date         DATE,
  total_amount     NUMERIC(14,2) NOT NULL,
  tax_amount       NUMERIC(14,2) DEFAULT 0,
  vendor_name      VARCHAR(255) NOT NULL,
  description      TEXT,
  status           VARCHAR(50) DEFAULT 'pending',
  file_path        TEXT,
  created_at       TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS tax_payments (
  id               SERIAL PRIMARY KEY,
  payment_date     DATE NOT NULL,
  tax_period       VARCHAR(20) NOT NULL,
  tax_type         VARCHAR(50) NOT NULL,
  tax_amount       NUMERIC(14,2) NOT NULL,
  penalty_amount   NUMERIC(14,2) DEFAULT 0,
  total_payment    NUMERIC(14,2) NOT NULL,
  reference_number VARCHAR(100),
  description      TEXT,
  created_at       TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS batch_history (
  id               SERIAL PRIMARY KEY,
  batch_type       VARCHAR(50) NOT NULL,
  filename         VARCHAR(255) NOT NULL,
  total_records    INT DEFAULT 0,
  processed_records INT DEFAULT 0,
  error_records    INT DEFAULT 0,
  status           VARCHAR(20) DEFAULT 'pending',
  file_name        VARCHAR(255),
  file_size        BIGINT,
  file_content_type VARCHAR(100),
  error_message    TEXT,
  created_at       TIMESTAMP NOT NULL DEFAULT NOW(),
  started_at       TIMESTAMP,
  completed_at     TIMESTAMP
);

CREATE TABLE IF NOT EXISTS batch_history_details (
  id              SERIAL PRIMARY KEY,
  batch_id        INT NOT NULL REFERENCES batch_history(id) ON DELETE CASCADE,
  record_id       VARCHAR(100),
  record_data     JSONB,
  status          VARCHAR(20) DEFAULT 'pending',
  error_message   TEXT,
  created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
  processed_at    TIMESTAMP
);

CREATE TABLE IF NOT EXISTS ads_performance (
  id                              SERIAL PRIMARY KEY,
  store_id                        INT REFERENCES stores(store_id),
  campaign_id                     VARCHAR(64) NOT NULL,
  campaign_name                   VARCHAR(256),
  campaign_type                   VARCHAR(64),
  campaign_status                 VARCHAR(32),
  performance_hour                TIMESTAMP NOT NULL,
  ads_viewed                      BIGINT DEFAULT 0,
  total_clicks                    BIGINT DEFAULT 0,
  orders_count                    BIGINT DEFAULT 0,
  products_sold                   BIGINT DEFAULT 0,
  sales_from_ads                  NUMERIC(15,2) DEFAULT 0,
  ad_costs                        NUMERIC(15,2) DEFAULT 0,
  click_rate                      NUMERIC(5,4) DEFAULT 0,
  roas                            NUMERIC(8,4) DEFAULT 0,
  daily_budget                    NUMERIC(15,2) DEFAULT 0,
  target_roas                     NUMERIC(8,4) DEFAULT 0,
  performance_change_percentage   NUMERIC(6,2) DEFAULT 0,
  created_at                      TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at                      TIMESTAMP NOT NULL DEFAULT NOW(),
  UNIQUE(store_id, campaign_id, performance_hour)
);

CREATE TABLE IF NOT EXISTS ads_sync_jobs (
  id                   SERIAL PRIMARY KEY,
  store_id             INT REFERENCES stores(store_id),
  start_date           DATE NOT NULL,
  end_date             DATE,
  total_campaigns      INT DEFAULT 0,
  processed_campaigns  INT DEFAULT 0,
  total_hours          INT DEFAULT 0,
  processed_hours      INT DEFAULT 0,
  status               VARCHAR(32) DEFAULT 'pending',
  error_message        TEXT,
  created_at           TIMESTAMP NOT NULL DEFAULT NOW(),
  started_at           TIMESTAMP,
  completed_at         TIMESTAMP
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_ads_performance_store_hour ON ads_performance(store_id, performance_hour);
CREATE INDEX IF NOT EXISTS idx_ads_performance_campaign ON ads_performance(campaign_id);
CREATE INDEX IF NOT EXISTS idx_ads_performance_status ON ads_performance(campaign_status);
CREATE INDEX IF NOT EXISTS idx_ads_performance_hour ON ads_performance(performance_hour);
CREATE INDEX IF NOT EXISTS idx_ads_sync_jobs_status ON ads_sync_jobs(status);
CREATE INDEX IF NOT EXISTS idx_ads_sync_jobs_store ON ads_sync_jobs(store_id);
