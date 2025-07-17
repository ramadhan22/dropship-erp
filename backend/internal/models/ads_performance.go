package models

import (
	"time"
)

// AdsCampaign represents an ads campaign from Shopee Marketing API
type AdsCampaign struct {
	CampaignID     int64      `db:"campaign_id" json:"campaign_id"`
	StoreID        int        `db:"store_id" json:"store_id"`
	CampaignName   string     `db:"campaign_name" json:"campaign_name"`
	CampaignType   *string    `db:"campaign_type" json:"campaign_type"`
	CampaignStatus string     `db:"campaign_status" json:"campaign_status"`
	PlacementType  *string    `db:"placement_type" json:"placement_type"`
	DailyBudget    *float64   `db:"daily_budget" json:"daily_budget"`
	TotalBudget    *float64   `db:"total_budget" json:"total_budget"`
	TargetRoas     *float64   `db:"target_roas" json:"target_roas"`
	StartDate      *time.Time `db:"start_date" json:"start_date"`
	EndDate        *time.Time `db:"end_date" json:"end_date"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at" json:"updated_at"`
}

// AdsPerformanceMetrics represents daily/hourly ads performance data
type AdsPerformanceMetrics struct {
	ID                int64     `db:"id" json:"id"`
	CampaignID        int64     `db:"campaign_id" json:"campaign_id"`
	StoreID           int       `db:"store_id" json:"store_id"`
	DateRecorded      time.Time `db:"date_recorded" json:"date_recorded"`
	HourRecorded      *int      `db:"hour_recorded" json:"hour_recorded"`
	AdsViewed         int64     `db:"ads_viewed" json:"ads_viewed"`
	AdsImpressions    int64     `db:"ads_impressions" json:"ads_impressions"`
	TotalClicks       int64     `db:"total_clicks" json:"total_clicks"`
	ClickPercentage   float64   `db:"click_percentage" json:"click_percentage"`
	OrdersCount       int64     `db:"orders_count" json:"orders_count"`
	ProductsSold      int64     `db:"products_sold" json:"products_sold"`
	SalesFromAdsCents int64     `db:"sales_from_ads_cents" json:"sales_from_ads_cents"`
	AdCostsCents      int64     `db:"ad_costs_cents" json:"ad_costs_cents"`
	Roas              float64   `db:"roas" json:"roas"`
	AvgCpcCents       int64     `db:"avg_cpc_cents" json:"avg_cpc_cents"`
	AvgCpmCents       int64     `db:"avg_cpm_cents" json:"avg_cpm_cents"`
	ConversionRate    float64   `db:"conversion_rate" json:"conversion_rate"`
	CreatedAt         time.Time `db:"created_at" json:"created_at"`
}

// AdsPerformanceMetricsWithCurrency provides convenient access to monetary values
type AdsPerformanceMetricsWithCurrency struct {
	AdsPerformanceMetrics
	SalesFromAds float64 `json:"sales_from_ads"` // Converted from cents
	AdCosts      float64 `json:"ad_costs"`       // Converted from cents
	AvgCpc       float64 `json:"avg_cpc"`        // Converted from cents
	AvgCpm       float64 `json:"avg_cpm"`        // Converted from cents
}

// ConvertToCurrency converts the cents-based fields to currency values
func (m *AdsPerformanceMetrics) ConvertToCurrency() *AdsPerformanceMetricsWithCurrency {
	return &AdsPerformanceMetricsWithCurrency{
		AdsPerformanceMetrics: *m,
		SalesFromAds:          float64(m.SalesFromAdsCents) / 100.0,
		AdCosts:               float64(m.AdCostsCents) / 100.0,
		AvgCpc:                float64(m.AvgCpcCents) / 100.0,
		AvgCpm:                float64(m.AvgCpmCents) / 100.0,
	}
}

// AdsCampaignWithMetrics combines campaign info with latest performance data
type AdsCampaignWithMetrics struct {
	AdsCampaign
	// Latest performance metrics (can be null if no data available)
	LatestMetrics *AdsPerformanceMetricsWithCurrency `json:"latest_metrics"`

	// Aggregated performance over selected period
	TotalViews      int64   `json:"total_views"`
	TotalClicks     int64   `json:"total_clicks"`
	TotalOrders     int64   `json:"total_orders"`
	TotalSales      float64 `json:"total_sales"`
	TotalAdCosts    float64 `json:"total_ad_costs"`
	AverageRoas     float64 `json:"average_roas"`
	AverageCtr      float64 `json:"average_ctr"`
	PeriodStartDate string  `json:"period_start_date"`
	PeriodEndDate   string  `json:"period_end_date"`
}

// AdsPerformanceSummary provides dashboard summary metrics
type AdsPerformanceSummary struct {
	TotalCampaigns        int     `json:"total_campaigns"`
	ActiveCampaigns       int     `json:"active_campaigns"`
	TotalAdsViewed        int64   `json:"total_ads_viewed"`
	TotalClicks           int64   `json:"total_clicks"`
	OverallClickPercent   float64 `json:"overall_click_percent"`
	TotalOrders           int64   `json:"total_orders"`
	TotalProductsSold     int64   `json:"total_products_sold"`
	TotalSalesFromAds     float64 `json:"total_sales_from_ads"`
	TotalAdCosts          float64 `json:"total_ad_costs"`
	OverallRoas           float64 `json:"overall_roas"`
	OverallConversionRate float64 `json:"overall_conversion_rate"`
	DateRange             string  `json:"date_range"`
	StoreFilter           *int    `json:"store_filter,omitempty"` // null means all stores
}
