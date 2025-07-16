package models

import "time"

// AdsPerformance represents advertising performance metrics for a campaign.
type AdsPerformance struct {
	ID                        int       `db:"id" json:"id"`
	StoreID                   int       `db:"store_id" json:"store_id"`
	CampaignID                string    `db:"campaign_id" json:"campaign_id"`
	CampaignName              string    `db:"campaign_name" json:"campaign_name"`
	CampaignType              string    `db:"campaign_type" json:"campaign_type"`
	CampaignStatus            string    `db:"campaign_status" json:"campaign_status"`
	DateFrom                  time.Time `db:"date_from" json:"date_from"`
	DateTo                    time.Time `db:"date_to" json:"date_to"`
	
	// Core metrics
	AdsViewed                 int64     `db:"ads_viewed" json:"ads_viewed"`
	TotalClicks               int64     `db:"total_clicks" json:"total_clicks"`
	OrdersCount               int64     `db:"orders_count" json:"orders_count"`
	ProductsSold              int64     `db:"products_sold" json:"products_sold"`
	SalesFromAds              float64   `db:"sales_from_ads" json:"sales_from_ads"`
	AdCosts                   float64   `db:"ad_costs" json:"ad_costs"`
	
	// Calculated metrics
	ClickRate                 float64   `db:"click_rate" json:"click_rate"`
	ROAS                      float64   `db:"roas" json:"roas"`
	
	// Budget and targeting
	DailyBudget               float64   `db:"daily_budget" json:"daily_budget"`
	TargetROAS                float64   `db:"target_roas" json:"target_roas"`
	
	// Performance indicators
	PerformanceChangePercentage float64 `db:"performance_change_percentage" json:"performance_change_percentage"`
	
	// Metadata
	CreatedAt                 time.Time `db:"created_at" json:"created_at"`
	UpdatedAt                 time.Time `db:"updated_at" json:"updated_at"`
}

// AdsPerformanceSummary represents aggregated metrics across multiple campaigns.
type AdsPerformanceSummary struct {
	TotalAdsViewed            int64     `json:"total_ads_viewed"`
	TotalClicks               int64     `json:"total_clicks"`
	TotalOrders               int64     `json:"total_orders"`
	TotalProductsSold         int64     `json:"total_products_sold"`
	TotalSalesFromAds         float64   `json:"total_sales_from_ads"`
	TotalAdCosts              float64   `json:"total_ad_costs"`
	AverageClickRate          float64   `json:"average_click_rate"`
	AverageROAS               float64   `json:"average_roas"`
	DateFrom                  time.Time `json:"date_from"`
	DateTo                    time.Time `json:"date_to"`
}

// AdsPerformanceFilter represents filter options for ads performance queries.
type AdsPerformanceFilter struct {
	StoreID        *int      `json:"store_id"`
	CampaignStatus *string   `json:"campaign_status"`
	DateFrom       *time.Time `json:"date_from"`
	DateTo         *time.Time `json:"date_to"`
	CampaignType   *string   `json:"campaign_type"`
}

// CalculateMetrics computes derived metrics from raw data.
func (ap *AdsPerformance) CalculateMetrics() {
	// Calculate click rate
	if ap.AdsViewed > 0 {
		ap.ClickRate = float64(ap.TotalClicks) / float64(ap.AdsViewed)
	}
	
	// Calculate ROAS
	if ap.AdCosts > 0 {
		ap.ROAS = ap.SalesFromAds / ap.AdCosts
	}
	
	ap.UpdatedAt = time.Now()
}