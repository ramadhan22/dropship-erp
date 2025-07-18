-- Remove additional fields from ads_campaigns table
DROP INDEX IF EXISTS idx_ads_campaigns_bidding_method;

ALTER TABLE ads_campaigns 
DROP COLUMN IF EXISTS bidding_method,
DROP COLUMN IF EXISTS campaign_budget,
DROP COLUMN IF EXISTS item_id_list,
DROP COLUMN IF EXISTS enhanced_cpc;