-- Modify ads_performance table for hourly data tracking
-- Drop the existing table and recreate with hourly structure

DROP TABLE IF EXISTS ads_performance;

CREATE TABLE IF NOT EXISTS ads_performance (
    id SERIAL PRIMARY KEY,
    store_id INT REFERENCES stores(store_id),
    campaign_id VARCHAR(64) NOT NULL,
    campaign_name VARCHAR(256),
    campaign_type VARCHAR(64),
    campaign_status VARCHAR(32),
    performance_hour TIMESTAMP NOT NULL, -- Changed from date_from/date_to to single hourly timestamp
    
    -- Core metrics
    ads_viewed BIGINT DEFAULT 0,
    total_clicks BIGINT DEFAULT 0,
    orders_count BIGINT DEFAULT 0,
    products_sold BIGINT DEFAULT 0,
    sales_from_ads NUMERIC(15,2) DEFAULT 0,
    ad_costs NUMERIC(15,2) DEFAULT 0,
    
    -- Calculated metrics
    click_rate NUMERIC(5,4) DEFAULT 0, -- Percentage as decimal (0.1234 for 12.34%)
    roas NUMERIC(8,4) DEFAULT 0, -- Return on Ad Spend
    
    -- Budget and targeting
    daily_budget NUMERIC(15,2) DEFAULT 0,
    target_roas NUMERIC(8,4) DEFAULT 0,
    
    -- Performance indicators
    performance_change_percentage NUMERIC(6,2) DEFAULT 0,
    
    -- Metadata
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Unique constraint to prevent duplicates for same hour
    UNIQUE(store_id, campaign_id, performance_hour)
);

-- Indexes for efficient queries
CREATE INDEX idx_ads_performance_store_hour ON ads_performance(store_id, performance_hour);
CREATE INDEX idx_ads_performance_campaign ON ads_performance(campaign_id);
CREATE INDEX idx_ads_performance_status ON ads_performance(campaign_status);
CREATE INDEX idx_ads_performance_hour ON ads_performance(performance_hour);

-- Create ads sync batch history for tracking background sync jobs
CREATE TABLE IF NOT EXISTS ads_sync_jobs (
    id SERIAL PRIMARY KEY,
    store_id INT REFERENCES stores(store_id),
    start_date DATE NOT NULL,
    end_date DATE,
    total_campaigns INT DEFAULT 0,
    processed_campaigns INT DEFAULT 0,
    total_hours INT DEFAULT 0,
    processed_hours INT DEFAULT 0,
    status VARCHAR(32) DEFAULT 'pending', -- pending, running, completed, failed
    error_message TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    started_at TIMESTAMP,
    completed_at TIMESTAMP
);

-- Index for sync jobs
CREATE INDEX idx_ads_sync_jobs_status ON ads_sync_jobs(status);
CREATE INDEX idx_ads_sync_jobs_store ON ads_sync_jobs(store_id);