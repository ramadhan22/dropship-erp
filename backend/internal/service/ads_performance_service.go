package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/ramadhan22/dropship-erp/backend/internal/config"
	"github.com/ramadhan22/dropship-erp/backend/internal/logutil"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
)

// AdsPerformanceService handles ads campaign and performance metrics
type AdsPerformanceService struct {
	db           *sqlx.DB
	shopeeClient *ShopeeClient
	repo         *repository.Repository
}

// NewAdsPerformanceService creates a new ads performance service
func NewAdsPerformanceService(db *sqlx.DB, cfg config.ShopeeAPIConfig, repo *repository.Repository) *AdsPerformanceService {
	return &AdsPerformanceService{
		db:           db,
		shopeeClient: NewShopeeClient(cfg),
		repo:         repo,
	}
}

// Shopee Marketing API response structures
type ShopeeAdsCampaignsResponse struct {
	Response struct {
		ShopID       int64 `json:"shop_id"`
		Region       string `json:"region"`
		HasNextPage  bool  `json:"has_next_page"`
		CampaignList []struct {
			CampaignID int64  `json:"campaign_id"`
			AdType     string `json:"ad_type"`
		} `json:"campaign_list"`
	} `json:"response"`
	Error     string `json:"error"`
	Message   string `json:"message"`
	Warning   string `json:"warning"`
	RequestID string `json:"request_id"`
}

type ShopeeAdsPerformanceResponse struct {
	Response []struct {
		ShopID       int64  `json:"shop_id"`
		Region       string `json:"region"`
		CampaignList []struct {
			CampaignID          int64  `json:"campaign_id"`
			AdType              string `json:"ad_type"`
			CampaignPlacement   string `json:"campaign_placement"`
			AdName              string `json:"ad_name"`
			MetricsList         []struct {
				Hour               int     `json:"hour"`
				Date               string  `json:"date"`
				Impression         int64   `json:"impression"`
				Clicks             int64   `json:"clicks"`
				Ctr                float64 `json:"ctr"`
				Expense            float64 `json:"expense"`
				BroadGmv           float64 `json:"broad_gmv"`
				BroadOrder         int64   `json:"broad_order"`
				BroadOrderAmount   float64 `json:"broad_order_amount"`
				BroadRoi           float64 `json:"broad_roi"`
				BroadCir           float64 `json:"broad_cir"`
				Cr                 float64 `json:"cr"`
				Cpc                float64 `json:"cpc"`
				DirectOrder        int64   `json:"direct_order"`
				DirectOrderAmount  float64 `json:"direct_order_amount"`
				DirectGmv          float64 `json:"direct_gmv"`
				DirectRoi          float64 `json:"direct_roi"`
				DirectCir          float64 `json:"direct_cir"`
				DirectCr           float64 `json:"direct_cr"`
				Cpdc               float64 `json:"cpdc"`
			} `json:"metrics_list"`
		} `json:"campaign_list"`
	} `json:"response"`
	Error     string `json:"error"`
	Message   string `json:"message"`
	Warning   string `json:"warning"`
	RequestID string `json:"request_id"`
}

// FetchAdsCampaigns retrieves ads campaigns from Shopee Marketing API for a specific store
func (s *AdsPerformanceService) FetchAdsCampaigns(ctx context.Context, storeID int) error {
	log.Printf("Starting to fetch ads campaigns for store %d", storeID)

	// Get store details and validate credentials
	store, err := s.repo.ChannelRepo.GetStoreByID(ctx, int64(storeID))
	if err != nil {
		logutil.Errorf("Failed to get store details for store %d: %v", storeID, err)
		return fmt.Errorf("failed to get store details: %w", err)
	}

	if store.ShopID == nil || store.AccessToken == nil {
		logutil.Errorf("Store %d does not have shop_id or access_token configured - shop_id: %v, access_token configured: %v", storeID, store.ShopID != nil, store.AccessToken != nil)
		return fmt.Errorf("store %d does not have shop_id or access_token configured", storeID)
	}

	log.Printf("Store %d credentials validated - shop_id: %s", storeID, *store.ShopID)

	// Update client with store-specific credentials
	s.shopeeClient.ShopID = *store.ShopID
	s.shopeeClient.AccessToken = *store.AccessToken

	totalCampaigns := 0
	totalSuccessCount := 0
	pageNo := 1
	pageSize := 100 // Shopee API default page size

	// Fetch campaigns with pagination
	for {
		log.Printf("Fetching campaigns page %d for store %d", pageNo, storeID)

		// Build API request
		path := "/api/v2/ads/get_product_level_campaign_id_list"
		ts := time.Now().Unix()
		sign := s.shopeeClient.signWithToken(path, ts, *store.AccessToken)

		params := url.Values{}
		params.Set("partner_id", s.shopeeClient.PartnerID)
		params.Set("shop_id", s.shopeeClient.ShopID)
		params.Set("timestamp", strconv.FormatInt(ts, 10))
		params.Set("access_token", *store.AccessToken)
		params.Set("sign", sign)
		params.Set("ad_type", "all")
		params.Set("page_no", strconv.Itoa(pageNo))
		params.Set("page_size", strconv.Itoa(pageSize))

		apiURL := s.shopeeClient.BaseURL + path + "?" + params.Encode()
		log.Printf("Making API request to Shopee for store %d: %s", storeID, path)

		// Make API request
		resp, err := s.shopeeClient.makeRequestWithRetry(ctx, "GET", apiURL, nil, nil)
		if err != nil {
			logutil.Errorf("Failed to fetch ads campaigns from Shopee API for store %d: %v", storeID, err)
			return fmt.Errorf("failed to fetch ads campaigns: %w", err)
		}
		defer resp.Body.Close()

		var campaignsResp ShopeeAdsCampaignsResponse
		if err := json.NewDecoder(resp.Body).Decode(&campaignsResp); err != nil {
			logutil.Errorf("Failed to decode campaigns response for store %d: %v", storeID, err)
			return fmt.Errorf("failed to decode campaigns response: %w", err)
		}

		if campaignsResp.Error != "" {
			logutil.Errorf("Shopee API error for store %d: %s - %s", storeID, campaignsResp.Error, campaignsResp.Message)
			return fmt.Errorf("Shopee API error: %s - %s", campaignsResp.Error, campaignsResp.Message)
		}

		log.Printf("Successfully received %d campaign objects from Shopee API for store %d on page %d", len(campaignsResp.Response.CampaignList), storeID, pageNo)

		// Store campaign IDs in database (with minimal campaign data)
		successCount := 0
		for _, campaign := range campaignsResp.Response.CampaignList {
			// Create minimal campaign object with just ID and store info
			campaignData := struct {
				CampaignID   int64  `json:"campaign_id"`
				CampaignName string `json:"campaign_name"`
				CampaignType string `json:"campaign_type"`
				StoreID      int    `json:"store_id"`
			}{
				CampaignID:   campaign.CampaignID,
				CampaignName: fmt.Sprintf("Campaign %d", campaign.CampaignID), // Placeholder name
				CampaignType: campaign.AdType,
				StoreID:      storeID,
			}
			
			err := s.upsertCampaign(ctx, storeID, &campaignData)
			if err != nil {
				logutil.Errorf("Failed to upsert campaign %d for store %d: %v", campaign.CampaignID, storeID, err)
				continue
			}
			successCount++
		}

		totalCampaigns += len(campaignsResp.Response.CampaignList)
		totalSuccessCount += successCount

		log.Printf("Successfully processed %d/%d campaigns from page %d for store %d", successCount, len(campaignsResp.Response.CampaignList), pageNo, storeID)

		// Check if there are more pages
		if !campaignsResp.Response.HasNextPage {
			log.Printf("No more pages available for store %d", storeID)
			break
		}

		pageNo++
		
		// Add small delay between pages to respect rate limits
		time.Sleep(100 * time.Millisecond)
	}

	log.Printf("Successfully fetched and stored %d/%d campaigns across all pages for store %d", totalSuccessCount, totalCampaigns, storeID)
	return nil
}

// FetchAdsPerformance retrieves ads performance data from Shopee Marketing API
func (s *AdsPerformanceService) FetchAdsPerformance(ctx context.Context, storeID int, campaignID int64, startDate, endDate time.Time) error {
	log.Printf("Starting to fetch ads performance for campaign %d, store %d, from %s to %s",
		campaignID, storeID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	// Get store details and validate credentials
	store, err := s.repo.ChannelRepo.GetStoreByID(ctx, int64(storeID))
	if err != nil {
		logutil.Errorf("Failed to get store details for store %d when fetching performance for campaign %d: %v", storeID, campaignID, err)
		return fmt.Errorf("failed to get store details: %w", err)
	}

	if store.ShopID == nil || store.AccessToken == nil {
		logutil.Errorf("Store %d does not have shop_id or access_token configured for campaign %d performance fetch", storeID, campaignID)
		return fmt.Errorf("store %d does not have shop_id or access_token configured", storeID)
	}

	// Update client with store-specific credentials
	s.shopeeClient.ShopID = *store.ShopID
	s.shopeeClient.AccessToken = *store.AccessToken

	// Since the new API only accepts one date at a time, iterate through each date
	totalDataPoints := 0
	successCount := 0
	currentDate := startDate
	
	for currentDate.Before(endDate) || currentDate.Equal(endDate) {
		log.Printf("Fetching performance data for campaign %d on date %s", campaignID, currentDate.Format("2006-01-02"))
		
		dataPoints, err := s.fetchAdsPerformanceForDate(ctx, storeID, campaignID, currentDate, store)
		if err != nil {
			logutil.Errorf("Failed to fetch performance data for campaign %d on date %s: %v", 
				campaignID, currentDate.Format("2006-01-02"), err)
			// Continue with next date instead of failing entire operation
			currentDate = currentDate.AddDate(0, 0, 1)
			continue
		}
		
		totalDataPoints += dataPoints
		successCount += dataPoints
		
		// Move to next day
		currentDate = currentDate.AddDate(0, 0, 1)
		
		// Small delay to respect API rate limits
		time.Sleep(100 * time.Millisecond)
	}

	log.Printf("Successfully fetched and stored %d/%d performance data points for campaign %d, store %d", 
		successCount, totalDataPoints, campaignID, storeID)
	return nil
}

// fetchAdsPerformanceForDate fetches performance data for a specific date
func (s *AdsPerformanceService) fetchAdsPerformanceForDate(ctx context.Context, storeID int, campaignID int64, date time.Time, store interface{}) (int, error) {
	// Build API request
	path := "/api/v2/ads/get_product_campaign_hourly_performance"
	ts := time.Now().Unix()
	sign := s.shopeeClient.signWithToken(path, ts, s.shopeeClient.AccessToken)

	params := url.Values{}
	params.Set("partner_id", s.shopeeClient.PartnerID)
	params.Set("shop_id", s.shopeeClient.ShopID)
	params.Set("timestamp", strconv.FormatInt(ts, 10))
	params.Set("access_token", s.shopeeClient.AccessToken)
	params.Set("sign", sign)
	params.Set("campaign_id_list", strconv.FormatInt(campaignID, 10))
	params.Set("performance_date", date.Format("02-01-2006")) // DD-MM-YYYY format as per Shopee API

	apiURL := s.shopeeClient.BaseURL + path + "?" + params.Encode()
	log.Printf("Making API request to Shopee for performance data - campaign %d, store %d, date %s: %s", 
		campaignID, storeID, date.Format("2006-01-02"), path)

	// Make API request
	resp, err := s.shopeeClient.makeRequestWithRetry(ctx, "GET", apiURL, nil, nil)
	if err != nil {
		logutil.Errorf("Failed to fetch ads performance from Shopee API for campaign %d, store %d, date %s: %v", 
			campaignID, storeID, date.Format("2006-01-02"), err)
		return 0, fmt.Errorf("failed to fetch ads performance: %w", err)
	}
	defer resp.Body.Close()

	var performanceResp ShopeeAdsPerformanceResponse
	if err := json.NewDecoder(resp.Body).Decode(&performanceResp); err != nil {
		logutil.Errorf("Failed to decode performance response for campaign %d, store %d, date %s: %v", 
			campaignID, storeID, date.Format("2006-01-02"), err)
		return 0, fmt.Errorf("failed to decode performance response: %w", err)
	}

	if performanceResp.Error != "" {
		logutil.Errorf("Shopee API error for campaign %d, store %d, date %s: %s - %s", 
			campaignID, storeID, date.Format("2006-01-02"), performanceResp.Error, performanceResp.Message)
		return 0, fmt.Errorf("Shopee API error: %s - %s", performanceResp.Error, performanceResp.Message)
	}

	// Store performance data in database
	totalDataPoints := 0
	successCount := 0
	for _, responseItem := range performanceResp.Response {
		for _, campaign := range responseItem.CampaignList {
			totalDataPoints += len(campaign.MetricsList)
			for _, metrics := range campaign.MetricsList {
				// Convert date from DD-MM-YYYY to YYYY-MM-DD format
				convertedDate, err := convertDateFormat(metrics.Date)
				if err != nil {
					logutil.Errorf("Failed to convert date format for campaign %d, date %s: %v", campaign.CampaignID, metrics.Date, err)
					continue
				}

				// Convert the new metrics format to the old format expected by upsertPerformanceMetrics
				data := struct {
					Date            string  `json:"date"`
					Hour            *int    `json:"hour,omitempty"`
					Impression      int64   `json:"impression"`
					Click           int64   `json:"click"`
					Ctr             float64 `json:"ctr"`
					Cpc             int64   `json:"cpc"`   // in cents
					Spend           int64   `json:"spend"` // in cents
					GmvOrder        int64   `json:"gmv_order"`
					GmvSales        int64   `json:"gmv_sales"` // in cents
					ConversionRate  float64 `json:"conversion_rate"`
					OrderConversion int64   `json:"order_conversion"`
					Roas            float64 `json:"roas"`
				}{
					Date:            convertedDate,
					Hour:            &metrics.Hour,
					Impression:      metrics.Impression,
					Click:           metrics.Clicks,
					Ctr:             metrics.Ctr,
					Cpc:             int64(metrics.Cpc * 100),      // Convert to cents
					Spend:           int64(metrics.Expense * 100),  // Convert to cents
					GmvOrder:        metrics.DirectOrder,
					GmvSales:        int64(metrics.DirectGmv * 100), // Convert to cents
					ConversionRate:  metrics.DirectCr,
					OrderConversion: metrics.DirectOrder,
					Roas:            metrics.DirectRoi,
				}
				
				if err = s.upsertPerformanceMetrics(ctx, storeID, campaign.CampaignID, &data); err != nil {
					logutil.Errorf("Failed to upsert performance data for campaign %d, date %s, store %d: %v",
						campaign.CampaignID, metrics.Date, storeID, err)
					continue
				}
				successCount++
			}
		}
	}

	log.Printf("Processed %d/%d performance data points for campaign %d, store %d, date %s", 
		successCount, totalDataPoints, campaignID, storeID, date.Format("2006-01-02"))
	return successCount, nil
}

// upsertCampaign stores or updates campaign data in the database
func (s *AdsPerformanceService) upsertCampaign(ctx context.Context, storeID int, campaign interface{}) error {
	// Type assertion for the campaign data structure
	type campaignData struct {
		CampaignID     int64  `json:"campaign_id"`
		CampaignName   string `json:"campaign_name"`
		CampaignType   string `json:"campaign_type"`
		CampaignStatus string `json:"campaign_status"`
		PlacementType  string `json:"placement_type"`
		DailyBudget    int64  `json:"daily_budget"`
		TotalBudget    int64  `json:"total_budget"`
		TargetRoas     int    `json:"target_roas"`
		StartTime      int64  `json:"start_time"`
		EndTime        int64  `json:"end_time"`
		StoreID        int    `json:"store_id"`
	}

	// Convert interface{} to our expected structure
	jsonData, err := json.Marshal(campaign)
	if err != nil {
		return fmt.Errorf("failed to marshal campaign data: %w", err)
	}

	var c campaignData
	if err := json.Unmarshal(jsonData, &c); err != nil {
		return fmt.Errorf("failed to unmarshal campaign data: %w", err)
	}

	// Use provided store ID if campaign doesn't have one
	if c.StoreID == 0 {
		c.StoreID = storeID
	}

	query := `
		INSERT INTO ads_campaigns (
			campaign_id, store_id, campaign_name, campaign_type, campaign_status,
			placement_type, daily_budget, total_budget, target_roas, start_date, end_date, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW()
		)
		ON CONFLICT (campaign_id) DO UPDATE SET
			campaign_name = EXCLUDED.campaign_name,
			campaign_type = EXCLUDED.campaign_type,
			campaign_status = EXCLUDED.campaign_status,
			placement_type = EXCLUDED.placement_type,
			daily_budget = EXCLUDED.daily_budget,
			total_budget = EXCLUDED.total_budget,
			target_roas = EXCLUDED.target_roas,
			start_date = EXCLUDED.start_date,
			end_date = EXCLUDED.end_date,
			updated_at = NOW()
	`

	var startDate, endDate *time.Time
	if c.StartTime > 0 {
		t := time.Unix(c.StartTime, 0)
		startDate = &t
	}
	if c.EndTime > 0 {
		t := time.Unix(c.EndTime, 0)
		endDate = &t
	}

	// Provide default values for missing fields
	campaignType := c.CampaignType
	if campaignType == "" {
		campaignType = "product"
	}
	
	campaignStatus := c.CampaignStatus
	if campaignStatus == "" {
		campaignStatus = "unknown"
	}

	_, err = s.db.ExecContext(ctx, query,
		c.CampaignID,
		c.StoreID,
		c.CampaignName,
		campaignType,
		campaignStatus,
		c.PlacementType,
		float64(c.DailyBudget)/100.0, // Convert from cents
		float64(c.TotalBudget)/100.0, // Convert from cents
		float64(c.TargetRoas)/100.0,  // Convert from percentage
		startDate,
		endDate,
	)

	return err
}

// upsertPerformanceMetrics stores or updates performance metrics in the database
func (s *AdsPerformanceService) upsertPerformanceMetrics(ctx context.Context, storeID int, campaignID int64, data interface{}) error {
	// Type assertion for the performance data structure
	type performanceData struct {
		Date            string  `json:"date"`
		Hour            *int    `json:"hour,omitempty"`
		Impression      int64   `json:"impression"`
		Click           int64   `json:"click"`
		Ctr             float64 `json:"ctr"`
		Cpc             int64   `json:"cpc"`
		Spend           int64   `json:"spend"`
		GmvOrder        int64   `json:"gmv_order"`
		GmvSales        int64   `json:"gmv_sales"`
		ConversionRate  float64 `json:"conversion_rate"`
		OrderConversion int64   `json:"order_conversion"`
		Roas            float64 `json:"roas"`
	}

	// Convert interface{} to our expected structure
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal performance data: %w", err)
	}

	var p performanceData
	if err := json.Unmarshal(jsonData, &p); err != nil {
		return fmt.Errorf("failed to unmarshal performance data: %w", err)
	}

	// Parse date
	dateRecorded, err := time.Parse("2006-01-02", p.Date)
	if err != nil {
		return fmt.Errorf("failed to parse date %s: %w", p.Date, err)
	}

	query := `
		INSERT INTO ads_performance_metrics (
			campaign_id, store_id, date_recorded, hour_recorded, ads_viewed, ads_impressions,
			total_clicks, click_percentage, orders_count, products_sold,
			sales_from_ads_cents, ad_costs_cents, roas, avg_cpc_cents,
			conversion_rate
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)
		ON CONFLICT (campaign_id, date_recorded, hour_recorded) DO UPDATE SET
			ads_viewed = EXCLUDED.ads_viewed,
			ads_impressions = EXCLUDED.ads_impressions,
			total_clicks = EXCLUDED.total_clicks,
			click_percentage = EXCLUDED.click_percentage,
			orders_count = EXCLUDED.orders_count,
			products_sold = EXCLUDED.products_sold,
			sales_from_ads_cents = EXCLUDED.sales_from_ads_cents,
			ad_costs_cents = EXCLUDED.ad_costs_cents,
			roas = EXCLUDED.roas,
			avg_cpc_cents = EXCLUDED.avg_cpc_cents,
			conversion_rate = EXCLUDED.conversion_rate
	`

	_, err = s.db.ExecContext(ctx, query,
		campaignID,
		storeID,
		dateRecorded,
		p.Hour,            // hour_recorded (can be nil for daily aggregates)
		p.Impression,      // ads_viewed
		p.Impression,      // ads_impressions (same as viewed)
		p.Click,           // total_clicks
		p.Ctr,             // click_percentage (already in decimal format)
		p.OrderConversion, // orders_count
		p.GmvOrder,        // products_sold (approximation)
		p.GmvSales,        // sales_from_ads_cents (already in cents)
		p.Spend,           // ad_costs_cents (already in cents)
		p.Roas,            // roas
		p.Cpc,             // avg_cpc_cents (already in cents)
		p.ConversionRate,  // conversion_rate
	)

	return err
}

// GetAdsCampaigns retrieves campaigns with optional filters
func (s *AdsPerformanceService) GetAdsCampaigns(ctx context.Context, storeID *int, status string, limit, offset int) ([]models.AdsCampaignWithMetrics, error) {
	query := `
		SELECT 
			c.campaign_id, c.store_id, c.campaign_name, c.campaign_type, c.campaign_status,
			c.placement_type, c.daily_budget, c.total_budget, c.target_roas,
			c.start_date, c.end_date, c.created_at, c.updated_at,
			st.nama_toko as store_name
		FROM ads_campaigns c
		JOIN stores st ON c.store_id = st.store_id
		WHERE 1=1
	`

	args := []interface{}{}
	argIndex := 1

	if storeID != nil {
		query += fmt.Sprintf(" AND c.store_id = $%d", argIndex)
		args = append(args, *storeID)
		argIndex++
	}

	if status != "" {
		query += fmt.Sprintf(" AND c.campaign_status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	query += " ORDER BY c.updated_at DESC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, limit)
		argIndex++
	}

	if offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, offset)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query campaigns: %w", err)
	}
	defer rows.Close()

	var campaigns []models.AdsCampaignWithMetrics
	for rows.Next() {
		var campaign models.AdsCampaignWithMetrics
		var storeName string
		err := rows.Scan(
			&campaign.CampaignID, &campaign.StoreID, &campaign.CampaignName,
			&campaign.CampaignType, &campaign.CampaignStatus, &campaign.PlacementType,
			&campaign.DailyBudget, &campaign.TotalBudget, &campaign.TargetRoas,
			&campaign.StartDate, &campaign.EndDate, &campaign.CreatedAt, &campaign.UpdatedAt,
			&storeName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan campaign: %w", err)
		}

		// TODO: Fetch latest metrics for this campaign
		campaigns = append(campaigns, campaign)
	}

	return campaigns, nil
}

// GetPerformanceSummary retrieves aggregated performance metrics
func (s *AdsPerformanceService) GetPerformanceSummary(ctx context.Context, storeID *int, startDate, endDate time.Time) (*models.AdsPerformanceSummary, error) {
	query := `
		SELECT 
			COUNT(DISTINCT c.campaign_id) as total_campaigns,
			COUNT(DISTINCT CASE WHEN c.campaign_status = 'ongoing' THEN c.campaign_id END) as active_campaigns,
			COALESCE(SUM(m.ads_viewed), 0) as total_ads_viewed,
			COALESCE(SUM(m.total_clicks), 0) as total_clicks,
			COALESCE(SUM(m.orders_count), 0) as total_orders,
			COALESCE(SUM(m.products_sold), 0) as total_products_sold,
			COALESCE(SUM(m.sales_from_ads_cents), 0) as total_sales_cents,
			COALESCE(SUM(m.ad_costs_cents), 0) as total_costs_cents,
			CASE 
				WHEN SUM(m.ad_costs_cents) > 0 THEN SUM(m.sales_from_ads_cents)::FLOAT / SUM(m.ad_costs_cents)::FLOAT
				ELSE 0 
			END as overall_roas,
			CASE 
				WHEN SUM(m.total_clicks) > 0 THEN SUM(m.orders_count)::FLOAT / SUM(m.total_clicks)::FLOAT
				ELSE 0 
			END as overall_conversion_rate
		FROM ads_campaigns c
		LEFT JOIN ads_performance_metrics m ON c.campaign_id = m.campaign_id 
			AND m.date_recorded BETWEEN $1 AND $2
		WHERE 1=1
	`

	args := []interface{}{startDate, endDate}
	argIndex := 3

	if storeID != nil {
		query += fmt.Sprintf(" AND c.store_id = $%d", argIndex)
		args = append(args, *storeID)
	}

	var summary models.AdsPerformanceSummary
	var totalSalesCents, totalCostsCents int64

	err := s.db.QueryRowContext(ctx, query, args...).Scan(
		&summary.TotalCampaigns, &summary.ActiveCampaigns,
		&summary.TotalAdsViewed, &summary.TotalClicks,
		&summary.TotalOrders, &summary.TotalProductsSold,
		&totalSalesCents, &totalCostsCents,
		&summary.OverallRoas, &summary.OverallConversionRate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get performance summary: %w", err)
	}

	// Convert from cents to currency
	summary.TotalSalesFromAds = float64(totalSalesCents) / 100.0
	summary.TotalAdCosts = float64(totalCostsCents) / 100.0

	// Calculate click percentage
	if summary.TotalAdsViewed > 0 {
		summary.OverallClickPercent = float64(summary.TotalClicks) / float64(summary.TotalAdsViewed)
	}

	summary.DateRange = fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	summary.StoreFilter = storeID

	return &summary, nil
}

// SyncHistoricalAdsPerformance syncs all historical ads performance data in background
func (s *AdsPerformanceService) SyncHistoricalAdsPerformance(ctx context.Context, storeID int) error {
	log.Printf("Starting historical ads performance sync for store %d", storeID)

	// First, fetch campaigns from Shopee API to ensure we have the latest campaign data
	log.Printf("Fetching campaigns from Shopee API for store %d", storeID)
	err := s.FetchAdsCampaigns(ctx, storeID)
	if err != nil {
		logutil.Errorf("Failed to fetch campaigns from Shopee API for store %d: %v", storeID, err)
		return fmt.Errorf("failed to fetch campaigns from Shopee API for store %d: %w", storeID, err)
	}

	// Get all campaigns for the store from database
	campaigns, err := s.GetAdsCampaigns(ctx, &storeID, "", 0, 0)
	if err != nil {
		logutil.Errorf("Failed to get campaigns from database for store %d: %v", storeID, err)
		return fmt.Errorf("failed to get campaigns for store %d: %w", storeID, err)
	}

	if len(campaigns) == 0 {
		logutil.Errorf("No campaigns found for store %d after fetching from Shopee API", storeID)
		return fmt.Errorf("no campaigns found for store %d after fetching from Shopee API", storeID)
	}

	log.Printf("Found %d campaigns for store %d", len(campaigns), storeID)

	// Split campaigns into batches of 100
	const batchSize = 100
	batches := make([][]models.AdsCampaignWithMetrics, 0)

	for i := 0; i < len(campaigns); i += batchSize {
		end := i + batchSize
		if end > len(campaigns) {
			end = len(campaigns)
		}
		batches = append(batches, campaigns[i:end])
	}

	log.Printf("Split %d campaigns into %d batches of max %d campaigns each for store %d", len(campaigns), len(batches), batchSize, storeID)

	// Process each batch going back in time
	currentDate := time.Now().Truncate(24 * time.Hour)
	consecutiveEmptyDays := 0
	maxConsecutiveEmptyDays := 2

	log.Printf("Starting historical sync for store %d from date %s, will stop after %d consecutive empty days", storeID, currentDate.Format("2006-01-02"), maxConsecutiveEmptyDays)

	for consecutiveEmptyDays < maxConsecutiveEmptyDays {
		dayHasData := false

		// Process each batch for the current date
		for batchIndex, batch := range batches {
			log.Printf("Processing batch %d/%d for store %d on date %s", batchIndex+1, len(batches), storeID, currentDate.Format("2006-01-02"))

			batchHasData, err := s.syncBatchForDate(ctx, storeID, batch, currentDate)
			if err != nil {
				logutil.Errorf("Failed to sync batch %d for store %d on date %s: %v", batchIndex+1, storeID, currentDate.Format("2006-01-02"), err)
				// Continue with next batch instead of failing entire operation
				continue
			}

			if batchHasData {
				dayHasData = true
			}
		}

		if dayHasData {
			consecutiveEmptyDays = 0
			log.Printf("Successfully processed data for store %d on date %s", storeID, currentDate.Format("2006-01-02"))
		} else {
			consecutiveEmptyDays++
			log.Printf("No data found for store %d on date %s, consecutive empty days: %d", storeID, currentDate.Format("2006-01-02"), consecutiveEmptyDays)
		}

		// Move to previous day
		currentDate = currentDate.AddDate(0, 0, -1)
	}

	log.Printf("Historical sync completed for store %d. Stopped after %d consecutive empty days", storeID, consecutiveEmptyDays)
	return nil
}

// syncBatchForDate syncs a batch of campaigns for a specific date
func (s *AdsPerformanceService) syncBatchForDate(ctx context.Context, storeID int, campaigns []models.AdsCampaignWithMetrics, date time.Time) (bool, error) {
	anyDataFound := false
	campaignCount := len(campaigns)
	log.Printf("Syncing batch of %d campaigns for store %d on date %s", campaignCount, storeID, date.Format("2006-01-02"))

	for i, campaign := range campaigns {
		log.Printf("Fetching performance for campaign %d (%s) [%d/%d] for store %d on date %s", campaign.CampaignID, campaign.CampaignName, i+1, campaignCount, storeID, date.Format("2006-01-02"))

		err := s.FetchAdsPerformance(ctx, storeID, campaign.CampaignID, date, date)
		if err != nil {
			logutil.Errorf("Failed to fetch performance for campaign %d (%s) on date %s for store %d: %v", campaign.CampaignID, campaign.CampaignName, date.Format("2006-01-02"), storeID, err)
			// Continue with next campaign instead of failing entire batch
			continue
		}

		// Check if we actually got data for this campaign and date
		hasData, err := s.hasPerformanceDataForDate(ctx, campaign.CampaignID, date)
		if err != nil {
			logutil.Errorf("Failed to check if performance data exists for campaign %d on date %s: %v", campaign.CampaignID, date.Format("2006-01-02"), err)
			continue
		}

		if hasData {
			anyDataFound = true
			log.Printf("Successfully fetched performance data for campaign %d (%s) on date %s", campaign.CampaignID, campaign.CampaignName, date.Format("2006-01-02"))
		} else {
			log.Printf("No performance data available for campaign %d (%s) on date %s", campaign.CampaignID, campaign.CampaignName, date.Format("2006-01-02"))
		}

		// Add small delay to respect API rate limits
		time.Sleep(100 * time.Millisecond)
	}

	log.Printf("Completed batch sync for store %d on date %s - data found: %v", storeID, date.Format("2006-01-02"), anyDataFound)
	return anyDataFound, nil
}

// hasPerformanceDataForDate checks if performance data exists for a campaign on a specific date
func (s *AdsPerformanceService) hasPerformanceDataForDate(ctx context.Context, campaignID int64, date time.Time) (bool, error) {
	query := `
		SELECT COUNT(*) 
		FROM ads_performance_metrics 
		WHERE campaign_id = $1 AND date_recorded = $2
	`

	var count int
	err := s.db.QueryRowContext(ctx, query, campaignID, date).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// convertDateFormat converts date from DD-MM-YYYY format to YYYY-MM-DD format
func convertDateFormat(dateStr string) (string, error) {
	// Parse date in DD-MM-YYYY format
	parsedDate, err := time.Parse("02-01-2006", dateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse date %s in DD-MM-YYYY format: %w", dateStr, err)
	}
	
	// Return in YYYY-MM-DD format
	return parsedDate.Format("2006-01-02"), nil
}
