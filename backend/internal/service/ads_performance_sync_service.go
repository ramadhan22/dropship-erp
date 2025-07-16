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

// AdsPerformanceSyncService handles background sync of ads performance data.
type AdsPerformanceSyncService struct {
	repo         *repository.AdsPerformanceRepository
	storeRepo    *repository.ChannelRepo
	shopeeClient *ShopeeClient
	batchSvc     *BatchService
}

// NewAdsPerformanceSyncService creates a new sync service instance.
func NewAdsPerformanceSyncService(db *sqlx.DB, shopeeClient *ShopeeClient, batchSvc *BatchService) *AdsPerformanceSyncService {
	return &AdsPerformanceSyncService{
		repo:         repository.NewAdsPerformanceRepository(db),
		storeRepo:    repository.NewChannelRepo(db),
		shopeeClient: shopeeClient,
		batchSvc:     batchSvc,
	}
}

// ShopeeCampaign represents a campaign from Shopee API
type ShopeeCampaign struct {
	CampaignID   string `json:"campaign_id"`
	CampaignName string `json:"campaign_name"`
	CampaignType string `json:"campaign_type"`
	Status       string `json:"status"`
}

// ShopeeCampaignsResponse represents the response from get campaigns API
type ShopeeCampaignsResponse struct {
	Error     string           `json:"error"`
	Message   string           `json:"message"`
	Campaigns []ShopeeCampaign `json:"campaigns"`
	Total     int              `json:"total"`
}

// ShopeeHourlyPerformanceData represents hourly performance data from Shopee API
type ShopeeHourlyPerformanceData struct {
	CampaignID   string    `json:"campaign_id"`
	CampaignName string    `json:"campaign_name"`
	CampaignType string    `json:"campaign_type"`
	Status       string    `json:"status"`
	Hour         time.Time `json:"hour"`

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
	Budget     float64 `json:"budget"`
	TargetROAS float64 `json:"target_roas"`
}

// ShopeeHourlyPerformanceResponse represents the response from hourly performance API
type ShopeeHourlyPerformanceResponse struct {
	Error       string                        `json:"error"`
	Message     string                        `json:"message"`
	Performance []ShopeeHourlyPerformanceData `json:"performance"`
}

// CreateSyncJob creates a new background sync job.
func (s *AdsPerformanceSyncService) CreateSyncJob(ctx context.Context, storeID int, startDate time.Time) (*models.AdsSyncJob, error) {
	job := &models.AdsSyncJob{
		StoreID:   storeID,
		StartDate: startDate,
		Status:    "pending",
	}

	if err := s.repo.CreateSyncJob(job); err != nil {
		return nil, fmt.Errorf("failed to create sync job: %w", err)
	}

	log.Printf("Created ads sync job %d for store %d starting from %s", job.ID, storeID, startDate.Format("2006-01-02"))
	return job, nil
}

// CreateFullHistorySyncJob creates a sync job that will fetch all historical data until no data is found.
func (s *AdsPerformanceSyncService) CreateFullHistorySyncJob(ctx context.Context, storeID int) (*models.AdsSyncJob, error) {
	// Start from yesterday and go backwards
	startDate := time.Now().AddDate(0, 0, -1).Truncate(24 * time.Hour)
	return s.CreateSyncJob(ctx, storeID, startDate)
}

// ProcessSyncJob processes a pending sync job in the background.
func (s *AdsPerformanceSyncService) ProcessSyncJob(ctx context.Context, jobID int64) {
	job, err := s.repo.GetSyncJob(jobID)
	if err != nil {
		logutil.Errorf("Failed to get sync job %d: %v", jobID, err)
		return
	}

	if job.Status != "pending" {
		log.Printf("Sync job %d is not pending (status: %s), skipping", jobID, job.Status)
		return
	}

	// Mark job as running
	now := time.Now()
	job.Status = "running"
	job.StartedAt = &now
	if err := s.repo.UpdateSyncJob(job); err != nil {
		logutil.Errorf("Failed to update sync job %d status: %v", jobID, err)
		return
	}

	log.Printf("Starting ads sync job %d for store %d", jobID, job.StoreID)

	// Process the sync
	if err := s.syncHistoricalData(ctx, job); err != nil {
		logutil.Errorf("Sync job %d failed: %v", jobID, err)
		job.Status = "failed"
		job.ErrorMessage = err.Error()
	} else {
		log.Printf("Sync job %d completed successfully", jobID)
		job.Status = "completed"
	}

	// Mark job as completed
	completed := time.Now()
	job.CompletedAt = &completed
	if err := s.repo.UpdateSyncJob(job); err != nil {
		logutil.Errorf("Failed to update sync job %d completion: %v", jobID, err)
	}
}

// syncHistoricalData performs the actual historical data sync.
func (s *AdsPerformanceSyncService) syncHistoricalData(ctx context.Context, job *models.AdsSyncJob) error {
	// Get store information
	store, err := s.storeRepo.GetStoreByID(ctx, int64(job.StoreID))
	if err != nil {
		return fmt.Errorf("failed to get store: %w", err)
	}

	if store.ShopID == nil || *store.ShopID == "" {
		return fmt.Errorf("store %d does not have shop_id configured", job.StoreID)
	}

	// Configure Shopee client for this store
	s.shopeeClient.ShopID = *store.ShopID
	if store.AccessToken != nil && *store.AccessToken != "" {
		s.shopeeClient.AccessToken = *store.AccessToken
	}

	// Step 1: Get all campaigns
	campaigns, err := s.getAllCampaigns(ctx, job.StoreID)
	if err != nil {
		return fmt.Errorf("failed to get campaigns: %w", err)
	}

	job.TotalCampaigns = len(campaigns)
	if err := s.repo.UpdateSyncJob(job); err != nil {
		logutil.Errorf("Failed to update job campaigns count: %v", err)
	}

	if len(campaigns) == 0 {
		log.Printf("No campaigns found for store %d", job.StoreID)
		return nil
	}

	log.Printf("Found %d campaigns for store %d", len(campaigns), job.StoreID)

	// Step 2: Process historical data going backwards
	currentDate := job.StartDate
	consecutiveDaysWithoutData := 0
	totalHoursProcessed := 0

	for consecutiveDaysWithoutData < 2 {
		// Check if we already have data for this date
		hasData, err := s.repo.HasPerformanceData(job.StoreID, currentDate)
		if err != nil {
			logutil.Errorf("Failed to check existing data for %s: %v", currentDate.Format("2006-01-02"), err)
		}

		if hasData {
			log.Printf("Data already exists for store %d on %s, skipping", job.StoreID, currentDate.Format("2006-01-02"))
			currentDate = currentDate.AddDate(0, 0, -1)
			continue
		}

		// Fetch hourly data for this date
		hoursProcessed, err := s.fetchHourlyDataForDate(ctx, job.StoreID, currentDate, campaigns)
		if err != nil {
			logutil.Errorf("Failed to fetch data for %s: %v", currentDate.Format("2006-01-02"), err)
			// Continue to next date instead of failing completely
		}

		if hoursProcessed == 0 {
			consecutiveDaysWithoutData++
			log.Printf("No data found for store %d on %s (%d consecutive days without data)",
				job.StoreID, currentDate.Format("2006-01-02"), consecutiveDaysWithoutData)
		} else {
			consecutiveDaysWithoutData = 0 // Reset counter
			totalHoursProcessed += hoursProcessed
			log.Printf("Processed %d hours of data for store %d on %s",
				hoursProcessed, job.StoreID, currentDate.Format("2006-01-02"))
		}

		// Update job progress
		job.ProcessedHours = totalHoursProcessed
		if err := s.repo.UpdateSyncJob(job); err != nil {
			logutil.Errorf("Failed to update job progress: %v", err)
		}

		// Move to previous day
		currentDate = currentDate.AddDate(0, 0, -1)

		// Safety check - don't go back more than 2 years
		if time.Since(currentDate) > 2*365*24*time.Hour {
			log.Printf("Reached 2-year limit, stopping sync for store %d", job.StoreID)
			break
		}
	}

	job.EndDate = &currentDate
	job.TotalHours = totalHoursProcessed
	log.Printf("Sync completed for store %d. Processed %d hours from %s to %s",
		job.StoreID, totalHoursProcessed, job.StartDate.Format("2006-01-02"), currentDate.Format("2006-01-02"))

	return nil
}

// getAllCampaigns fetches all campaigns for a store from Shopee API.
func (s *AdsPerformanceSyncService) getAllCampaigns(ctx context.Context, storeID int) ([]ShopeeCampaign, error) {
	// Note: This is a mock implementation. In real Shopee API, you would call:
	// GET /api/v2/ads/get_campaigns
	// For now, we'll return a mock response since the exact API endpoint structure
	// depends on Shopee's actual ads API documentation

	path := "/api/v2/ads/get_campaigns"
	ts := time.Now().Unix()

	store, err := s.storeRepo.GetStoreByID(ctx, int64(storeID))
	if err != nil {
		return nil, err
	}

	params := map[string]interface{}{
		"partner_id": s.shopeeClient.PartnerID,
		"shop_id":    *store.ShopID,
		"timestamp":  ts,
		"page_size":  1000, // Get all campaigns
		"page_no":    1,
	}

	sign := s.shopeeClient.signWithTokenShop(path, ts, s.shopeeClient.AccessToken, *store.ShopID)
	params["sign"] = sign
	params["access_token"] = s.shopeeClient.AccessToken

	log.Printf("Fetching campaigns from Shopee API for store %d", storeID)

	response, err := s.shopeeClient.makeGetRequest(ctx, path, params)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch campaigns: %w", err)
	}

	var campaignsResponse ShopeeCampaignsResponse
	if err := json.Unmarshal(response, &campaignsResponse); err != nil {
		return nil, fmt.Errorf("failed to parse campaigns response: %w", err)
	}

	if campaignsResponse.Error != "" {
		return nil, fmt.Errorf("shopee API error: %s", campaignsResponse.Message)
	}

	return campaignsResponse.Campaigns, nil
}

// fetchHourlyDataForDate fetches hourly performance data for a specific date.
func (s *AdsPerformanceSyncService) fetchHourlyDataForDate(ctx context.Context, storeID int, date time.Time, campaigns []ShopeeCampaign) (int, error) {
	totalHoursProcessed := 0

	// Split campaigns into batches of 100
	batchSize := 100
	for i := 0; i < len(campaigns); i += batchSize {
		end := i + batchSize
		if end > len(campaigns) {
			end = len(campaigns)
		}

		batch := campaigns[i:end]
		hours, err := s.fetchHourlyDataForCampaignBatch(ctx, storeID, date, batch)
		if err != nil {
			logutil.Errorf("Failed to fetch data for campaign batch: %v", err)
			continue
		}

		totalHoursProcessed += hours
	}

	return totalHoursProcessed, nil
}

// fetchHourlyDataForCampaignBatch fetches hourly data for a batch of campaigns.
func (s *AdsPerformanceSyncService) fetchHourlyDataForCampaignBatch(ctx context.Context, storeID int, date time.Time, campaigns []ShopeeCampaign) (int, error) {
	// Build campaign IDs list
	var campaignIDs []string
	for _, campaign := range campaigns {
		campaignIDs = append(campaignIDs, campaign.CampaignID)
	}

	path := "/api/v2/ads/get_hourly_performance"
	ts := time.Now().Unix()

	store, err := s.storeRepo.GetStoreByID(ctx, int64(storeID))
	if err != nil {
		return 0, err
	}

	params := map[string]interface{}{
		"partner_id":   s.shopeeClient.PartnerID,
		"shop_id":      *store.ShopID,
		"timestamp":    ts,
		"date":         date.Format("2006-01-02"),
		"campaign_ids": campaignIDs,
	}

	sign := s.shopeeClient.signWithTokenShop(path, ts, s.shopeeClient.AccessToken, *store.ShopID)
	params["sign"] = sign
	params["access_token"] = s.shopeeClient.AccessToken

	response, err := s.shopeeClient.makeGetRequest(ctx, path, params)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch hourly performance: %w", err)
	}

	var perfResponse ShopeeHourlyPerformanceResponse
	if err := json.Unmarshal(response, &perfResponse); err != nil {
		return 0, fmt.Errorf("failed to parse hourly performance response: %w", err)
	}

	if perfResponse.Error != "" {
		return 0, fmt.Errorf("shopee API error: %s", perfResponse.Message)
	}

	// Convert and save hourly data
	hoursProcessed := 0
	for _, perfData := range perfResponse.Performance {
		adsPerformance := s.convertHourlyDataToModel(storeID, perfData)
		if err := s.repo.Create(adsPerformance); err != nil {
			logutil.Errorf("Failed to save hourly performance data: %v", err)
			continue
		}
		hoursProcessed++
	}

	return hoursProcessed, nil
}

// convertHourlyDataToModel converts Shopee hourly API data to our model.
func (s *AdsPerformanceSyncService) convertHourlyDataToModel(storeID int, data ShopeeHourlyPerformanceData) *models.AdsPerformance {
	return &models.AdsPerformance{
		StoreID:         storeID,
		CampaignID:      data.CampaignID,
		CampaignName:    data.CampaignName,
		CampaignType:    data.CampaignType,
		CampaignStatus:  data.Status,
		PerformanceHour: data.Hour,

		// Map Shopee metrics to our model
		AdsViewed:    data.Impression,
		TotalClicks:  data.Click,
		OrdersCount:  data.Order,
		ProductsSold: data.Order, // Assuming 1:1 for now
		SalesFromAds: data.GMV,
		AdCosts:      data.Cost,
		ClickRate:    data.CTR,
		ROAS:         data.ROAS,
		DailyBudget:  data.Budget,
		TargetROAS:   data.TargetROAS,
	}
}

// GetSyncJob retrieves a sync job by ID.
func (s *AdsPerformanceSyncService) GetSyncJob(jobID int64) (*models.AdsSyncJob, error) {
	return s.repo.GetSyncJob(jobID)
}

// ListSyncJobs lists sync jobs with optional filtering.
func (s *AdsPerformanceSyncService) ListSyncJobs(storeID *int, limit, offset int) ([]models.AdsSyncJob, error) {
	return s.repo.ListSyncJobs(storeID, limit, offset)
}

// ListPendingSyncJobs returns all pending sync jobs for processing.
func (s *AdsPerformanceSyncService) ListPendingSyncJobs() ([]models.AdsSyncJob, error) {
	return s.repo.ListPendingSyncJobs()
}
