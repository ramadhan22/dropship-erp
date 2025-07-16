-- Create ads performance metrics table for daily/hourly ad metrics
CREATE TABLE IF NOT EXISTS ads_performance_metrics (
    id SERIAL PRIMARY KEY,
    campaign_id BIGINT NOT NULL REFERENCES ads_campaigns(campaign_id) ON DELETE CASCADE,
    store_id INT NOT NULL REFERENCES stores(store_id) ON DELETE CASCADE,
    date_recorded DATE NOT NULL,
    hour_recorded INT, -- NULL for daily aggregates, 0-23 for hourly data
    
    -- Viewability metrics
    ads_viewed BIGINT DEFAULT 0,
    ads_impressions BIGINT DEFAULT 0,
    
    -- Engagement metrics  
    total_clicks BIGINT DEFAULT 0,
    click_percentage NUMERIC(5,4) DEFAULT 0, -- CTR as decimal (e.g., 0.0250 for 2.5%)
    
    -- Conversion metrics
    orders_count BIGINT DEFAULT 0,
    products_sold BIGINT DEFAULT 0,
    
    -- Financial metrics (in cents/smallest currency unit to avoid floating point issues)
    sales_from_ads_cents BIGINT DEFAULT 0, -- Revenue from ads
    ad_costs_cents BIGINT DEFAULT 0, -- Total ad spend
    roas NUMERIC(8,4) DEFAULT 0, -- Return on Ad Spend
    
    -- Additional metrics
    avg_cpc_cents BIGINT DEFAULT 0, -- Average cost per click
    avg_cpm_cents BIGINT DEFAULT 0, -- Average cost per mille (1000 impressions)
    conversion_rate NUMERIC(5,4) DEFAULT 0, -- Orders/Clicks ratio
    
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Ensure no duplicate records for same campaign/date/hour
    UNIQUE(campaign_id, date_recorded, hour_recorded)
);

-- Create indexes for performance queries
CREATE INDEX idx_ads_performance_campaign_date ON ads_performance_metrics(campaign_id, date_recorded);
CREATE INDEX idx_ads_performance_store_date ON ads_performance_metrics(store_id, date_recorded);
CREATE INDEX idx_ads_performance_date_recorded ON ads_performance_metrics(date_recorded);