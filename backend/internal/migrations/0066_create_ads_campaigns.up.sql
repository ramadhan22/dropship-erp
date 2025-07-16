-- Create ads campaigns table for storing Shopee ad campaign data
CREATE TABLE IF NOT EXISTS ads_campaigns (
    campaign_id BIGINT PRIMARY KEY,
    store_id INT NOT NULL REFERENCES stores(store_id) ON DELETE CASCADE,
    campaign_name VARCHAR(255) NOT NULL,
    campaign_type VARCHAR(50), -- keyword, product, shop, etc.
    campaign_status VARCHAR(50) NOT NULL, -- ongoing, paused, ended, scheduled
    placement_type VARCHAR(50), -- search, discovery, etc.
    daily_budget NUMERIC(15,2),
    total_budget NUMERIC(15,2),
    target_roas NUMERIC(5,2),
    start_date DATE,
    end_date DATE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes for common queries
CREATE INDEX idx_ads_campaigns_store_id ON ads_campaigns(store_id);
CREATE INDEX idx_ads_campaigns_status ON ads_campaigns(campaign_status);
CREATE INDEX idx_ads_campaigns_start_date ON ads_campaigns(start_date);