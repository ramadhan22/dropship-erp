-- Revert to original ads_performance table structure
DROP TABLE IF EXISTS ads_sync_jobs;
DROP TABLE IF EXISTS ads_performance;

CREATE TABLE IF NOT EXISTS ads_performance (
    id SERIAL PRIMARY KEY,
    store_id INT REFERENCES stores(id),
    campaign_id VARCHAR(64) NOT NULL,
    campaign_name VARCHAR(256),
    campaign_type VARCHAR(64),
    campaign_status VARCHAR(32),
    date_from DATE NOT NULL,
    date_to DATE NOT NULL,
    
    -- Core metrics
    ads_viewed BIGINT DEFAULT 0,
    total_clicks BIGINT DEFAULT 0,
    orders_count BIGINT DEFAULT 0,
    products_sold BIGINT DEFAULT 0,
    sales_from_ads NUMERIC(15,2) DEFAULT 0,
    ad_costs NUMERIC(15,2) DEFAULT 0,
    
    -- Calculated metrics
    click_rate NUMERIC(5,4) DEFAULT 0,
    roas NUMERIC(8,4) DEFAULT 0,
    
    -- Budget and targeting
    daily_budget NUMERIC(15,2) DEFAULT 0,
    target_roas NUMERIC(8,4) DEFAULT 0,
    
    -- Performance indicators
    performance_change_percentage NUMERIC(6,2) DEFAULT 0,
    
    -- Metadata
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Unique constraint to prevent duplicates
    UNIQUE(store_id, campaign_id, date_from, date_to)
);

-- Restore original indexes
CREATE INDEX idx_ads_performance_store_date ON ads_performance(store_id, date_from, date_to);
CREATE INDEX idx_ads_performance_campaign ON ads_performance(campaign_id);
CREATE INDEX idx_ads_performance_status ON ads_performance(campaign_status);