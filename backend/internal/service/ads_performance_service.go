package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/ramadhan22/dropship-erp/backend/internal/logutil"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
)

// AdsPerformanceService handles business logic for ads performance data.
type AdsPerformanceService struct {
	repo         *repository.AdsPerformanceRepository
	storeRepo    *repository.ChannelRepo
	shopeeClient *ShopeeClient
}

// NewAdsPerformanceService creates a new service instance.
func NewAdsPerformanceService(db *sqlx.DB, shopeeClient *ShopeeClient) *AdsPerformanceService {
	return &AdsPerformanceService{
		repo:         repository.NewAdsPerformanceRepository(db),
		storeRepo:    repository.NewChannelRepo(db),
		shopeeClient: shopeeClient,
	}
}

// ShopeeAdsResponse represents the response from Shopee Marketing API
type ShopeeAdsResponse struct {
	Error   string                     `json:"error"`
	Message string                     `json:"message"`
	Ads     []ShopeeAdsPerformanceData `json:"ads"`
}

// ShopeeAdsPerformanceData represents ads performance data from Shopee API
type ShopeeAdsPerformanceData struct {
	CampaignID   string  `json:"campaign_id"`
	CampaignName string  `json:"campaign_name"`
	CampaignType string  `json:"campaign_type"`
	Status       string  `json:"status"`
	Budget       float64 `json:"budget"`
	TargetROAS   float64 `json:"target_roas"`

	// Performance metrics
	Impression int64   `json:"impression"`
	Click      int64   `json:"click"`
	Order      int64   `json:"order"`
	GMV        float64 `json:"gmv"`
	Cost       float64 `json:"cost"`
	CVR        float64 `json:"cvr"`
	CTR        float64 `json:"ctr"`
	CPC        float64 `json:"cpc"`
	ROAS       float64 `json:"roas"`

	// Time range
	DateFrom string `json:"date_from"`
	DateTo   string `json:"date_to"`
}

// FetchAdsPerformanceFromShopee retrieves ads performance data from Shopee API
func (s *AdsPerformanceService) FetchAdsPerformanceFromShopee(ctx context.Context, storeID int, dateFrom, dateTo time.Time) error {
	logutil.Errorf("Starting ads performance fetch for store %d from %s to %s", storeID, dateFrom.Format("2006-01-02"), dateTo.Format("2006-01-02"))

	// Get store information
	store, err := s.storeRepo.GetStoreByID(ctx, int64(storeID))
	if err != nil {
		return fmt.Errorf("failed to get store: %w", err)
	}

	if store.ShopID == nil || *store.ShopID == "" {
		return fmt.Errorf("store %d does not have shop_id configured", storeID)
	}

	// Configure Shopee client for this store
	s.shopeeClient.ShopID = *store.ShopID
	if store.AccessToken != nil && *store.AccessToken != "" {
		s.shopeeClient.AccessToken = *store.AccessToken
	}

	// API path for marketing performance
	path := "/api/v2/ads/get_ads_performance"
	ts := time.Now().Unix()

	// Prepare request parameters
	params := map[string]interface{}{
		"partner_id": s.shopeeClient.PartnerID,
		"shop_id":    *store.ShopID,
		"timestamp":  ts,
		"date_from":  dateFrom.Format("2006-01-02"),
		"date_to":    dateTo.Format("2006-01-02"),
		"page_size":  100,
		"page_no":    1,
	}

	// Generate signature
	sign := s.shopeeClient.signWithTokenShop(path, ts, s.shopeeClient.AccessToken, *store.ShopID)
	params["sign"] = sign
	params["access_token"] = s.shopeeClient.AccessToken

	log.Printf("Fetching ads performance data from Shopee API for store %d", storeID)

	// Make API request
	response, err := s.shopeeClient.makeGetRequest(ctx, path, params)
	if err != nil {
		logutil.Errorf("Failed to fetch ads performance from Shopee API: %v", err)
		return fmt.Errorf("failed to fetch ads performance: %w", err)
	}

	var adsResponse ShopeeAdsResponse
	if err := json.Unmarshal(response, &adsResponse); err != nil {
		logutil.Errorf("Failed to parse ads performance response: %v", err)
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if adsResponse.Error != "" {
		logutil.Errorf("Shopee API returned error: %s - %s", adsResponse.Error, adsResponse.Message)
		return fmt.Errorf("shopee API error: %s", adsResponse.Message)
	}

	// Convert and save ads performance data
	for _, adsData := range adsResponse.Ads {
		// For backwards compatibility, create hourly records for each hour in the date range
		currentHour := dateFrom
		for currentHour.Before(dateTo.Add(time.Hour)) || currentHour.Equal(dateTo) {
			adsPerformance, err := s.convertShopeeDataToModel(storeID, adsData, currentHour)
			if err != nil {
				logutil.Errorf("Failed to convert ads data: %v", err)
				currentHour = currentHour.Add(time.Hour)
				continue
			}

			if err := s.repo.Create(adsPerformance); err != nil {
				logutil.Errorf("Failed to save ads performance: %v", err)
			}

			currentHour = currentHour.Add(time.Hour)
		}
	}

	log.Printf("Successfully fetched and saved %d ads performance records for store %d", len(adsResponse.Ads), storeID)
	return nil
}

// convertShopeeDataToModel converts Shopee API data to our model
func (s *AdsPerformanceService) convertShopeeDataToModel(storeID int, data ShopeeAdsPerformanceData, performanceHour time.Time) (*models.AdsPerformance, error) {
	ap := &models.AdsPerformance{
		StoreID:         storeID,
		CampaignID:      data.CampaignID,
		CampaignName:    data.CampaignName,
		CampaignType:    data.CampaignType,
		CampaignStatus:  data.Status,
		PerformanceHour: performanceHour,

		// Map Shopee metrics to our model
		AdsViewed:    data.Impression,
		TotalClicks:  data.Click,
		OrdersCount:  data.Order,
		ProductsSold: data.Order, // Assuming 1:1 for now, can be adjusted
		SalesFromAds: data.GMV,
		AdCosts:      data.Cost,
		ClickRate:    data.CTR,
		ROAS:         data.ROAS,
		DailyBudget:  data.Budget,
		TargetROAS:   data.TargetROAS,
	}

	return ap, nil
}

// GetAdsPerformance retrieves ads performance data with filtering
func (s *AdsPerformanceService) GetAdsPerformance(filter *models.AdsPerformanceFilter, limit, offset int) ([]models.AdsPerformance, error) {
	return s.repo.List(filter, limit, offset)
}

// GetAdsPerformanceSummary calculates aggregated metrics
func (s *AdsPerformanceService) GetAdsPerformanceSummary(filter *models.AdsPerformanceFilter) (*models.AdsPerformanceSummary, error) {
	return s.repo.GetSummary(filter)
}

// RefreshAdsData fetches fresh data from Shopee API for all stores
func (s *AdsPerformanceService) RefreshAdsData(ctx context.Context, dateFrom, dateTo time.Time) error {
	// Get all stores
	stores, err := s.storeRepo.ListAllStores(ctx)
	if err != nil {
		return fmt.Errorf("failed to get stores: %w", err)
	}

	var errors []error
	for _, store := range stores {
		if store.ShopID == nil || *store.ShopID == "" {
			log.Printf("Skipping store %d: no shop_id configured", store.StoreID)
			continue
		}

		if err := s.FetchAdsPerformanceFromShopee(ctx, int(store.StoreID), dateFrom, dateTo); err != nil {
			logutil.Errorf("Failed to fetch ads data for store %d: %v", store.StoreID, err)
			errors = append(errors, fmt.Errorf("store %d: %w", store.StoreID, err))
			continue
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to fetch data for some stores: %v", errors)
	}

	return nil
}
