-- Add additional fields to ads_campaigns table for storing detailed campaign settings
ALTER TABLE ads_campaigns 
ADD COLUMN bidding_method VARCHAR(50),
ADD COLUMN campaign_budget NUMERIC(15,2),
ADD COLUMN item_id_list TEXT, -- Store JSON array of item IDs as text
ADD COLUMN enhanced_cpc BOOLEAN;

-- Create index for bidding_method for filtering
CREATE INDEX idx_ads_campaigns_bidding_method ON ads_campaigns(bidding_method);