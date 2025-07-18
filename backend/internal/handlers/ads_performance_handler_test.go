package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// Mock service for testing
type mockAdsPerformanceService struct{}

func (m *mockAdsPerformanceService) GetAdsCampaigns(ctx context.Context, storeID *int, status string, startDate, endDate *time.Time, limit, offset int) ([]models.AdsCampaignWithMetrics, error) {
	return []models.AdsCampaignWithMetrics{
		{
			AdsCampaign: models.AdsCampaign{
				CampaignID:     123,
				StoreID:        1,
				CampaignName:   "Test Campaign",
				CampaignStatus: "ongoing",
				CampaignType:   &[]string{"keyword"}[0],
				DailyBudget:    &[]float64{50000}[0],
				TargetRoas:     &[]float64{2.5}[0],
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			},
		},
	}, nil
}

func (m *mockAdsPerformanceService) GetPerformanceSummary(ctx context.Context, storeID *int, startDate, endDate time.Time) (*models.AdsPerformanceSummary, error) {
	return &models.AdsPerformanceSummary{
		TotalCampaigns:        5,
		ActiveCampaigns:       3,
		TotalAdsViewed:        10000,
		TotalClicks:           500,
		OverallClickPercent:   0.05,
		TotalOrders:           25,
		TotalProductsSold:     30,
		TotalSalesFromAds:     1250000,
		TotalAdCosts:          500000,
		OverallRoas:           2.5,
		OverallConversionRate: 0.05,
		DateRange:             "2024-01-01 to 2024-01-31",
	}, nil
}

func (m *mockAdsPerformanceService) FetchAdsCampaigns(ctx context.Context, storeID int) error {
	return nil
}

func (m *mockAdsPerformanceService) FetchAdsCampaignSettings(ctx context.Context, storeID int, campaignIDs []int64) error {
	return nil
}

func (m *mockAdsPerformanceService) FetchAdsPerformance(ctx context.Context, storeID int, campaignID int64, startDate, endDate time.Time) error {
	return nil
}

func (m *mockAdsPerformanceService) SyncAdsPerformanceBatch(ctx context.Context, storeID int, campaigns []models.AdsCampaignWithMetrics) error {
	return nil
}

// Mock batch scheduler for testing
type mockBatchScheduler struct{}

func (m *mockBatchScheduler) CreateSyncBatch(ctx context.Context, storeID int) (int64, error) {
	return 123, nil
}

func TestAdsPerformanceHandler_GetAdsCampaigns(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &AdsPerformanceHandler{
		adsService:     &mockAdsPerformanceService{},
		batchScheduler: &mockBatchScheduler{},
	}

	router := gin.New()
	router.GET("/campaigns", handler.GetAdsCampaigns)

	req := httptest.NewRequest("GET", "/campaigns?limit=10", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	campaigns, ok := response["campaigns"].([]interface{})
	if !ok {
		t.Error("Expected campaigns array in response")
	}

	if len(campaigns) != 1 {
		t.Errorf("Expected 1 campaign, got %d", len(campaigns))
	}
}

func TestAdsPerformanceHandler_GetPerformanceSummary(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &AdsPerformanceHandler{
		adsService:     &mockAdsPerformanceService{},
		batchScheduler: &mockBatchScheduler{},
	}

	router := gin.New()
	router.GET("/summary", handler.GetPerformanceSummary)

	req := httptest.NewRequest("GET", "/summary?start_date=2024-01-01&end_date=2024-01-31", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var summary models.AdsPerformanceSummary
	err := json.Unmarshal(w.Body.Bytes(), &summary)
	if err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if summary.TotalCampaigns != 5 {
		t.Errorf("Expected 5 total campaigns, got %d", summary.TotalCampaigns)
	}

	if summary.OverallRoas != 2.5 {
		t.Errorf("Expected ROAS 2.5, got %f", summary.OverallRoas)
	}
}

func TestAdsPerformanceHandler_FetchAdsCampaigns(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &AdsPerformanceHandler{
		adsService:     &mockAdsPerformanceService{},
		batchScheduler: &mockBatchScheduler{},
	}

	router := gin.New()
	router.POST("/campaigns/fetch", handler.FetchAdsCampaigns)

	requestBody := map[string]interface{}{
		"store_id": 1,
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/campaigns/fetch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	message, ok := response["message"].(string)
	if !ok || message != "Campaigns fetched successfully" {
		t.Errorf("Expected success message, got %v", response)
	}
}
