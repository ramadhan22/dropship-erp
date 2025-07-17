package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/logutil"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// AdsPerformanceServiceInterface defines the methods needed by the handler
type AdsPerformanceServiceInterface interface {
	GetAdsCampaigns(ctx context.Context, storeID *int, status string, limit, offset int) ([]models.AdsCampaignWithMetrics, error)
	GetPerformanceSummary(ctx context.Context, storeID *int, startDate, endDate time.Time) (*models.AdsPerformanceSummary, error)
	FetchAdsCampaigns(ctx context.Context, storeID int) error
	FetchAdsCampaignSettings(ctx context.Context, storeID int, campaignIDs []int64) error
	FetchAdsPerformance(ctx context.Context, storeID int, campaignID int64, startDate, endDate time.Time) error
}

// AdsPerformanceBatchSchedulerInterface defines methods needed by the handler
type AdsPerformanceBatchSchedulerInterface interface {
	CreateSyncBatch(ctx context.Context, storeID int) (int64, error)
}

// AdsPerformanceHandler handles HTTP requests for ads performance data
type AdsPerformanceHandler struct {
	adsService     AdsPerformanceServiceInterface
	batchScheduler AdsPerformanceBatchSchedulerInterface
}

// NewAdsPerformanceHandler creates a new ads performance handler
func NewAdsPerformanceHandler(adsService AdsPerformanceServiceInterface, batchScheduler AdsPerformanceBatchSchedulerInterface) *AdsPerformanceHandler {
	return &AdsPerformanceHandler{
		adsService:     adsService,
		batchScheduler: batchScheduler,
	}
}

// GetAdsCampaigns returns ads campaigns with optional filters
// GET /api/ads/campaigns?store_id=1&status=ongoing&limit=50&offset=0
func (h *AdsPerformanceHandler) GetAdsCampaigns(c *gin.Context) {
	var storeID *int
	if storeIDStr := c.Query("store_id"); storeIDStr != "" {
		if id, err := strconv.Atoi(storeIDStr); err == nil {
			storeID = &id
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid store_id parameter"})
			return
		}
	}

	status := c.Query("status")

	limit := 50 // default
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	ctx := context.Background()
	campaigns, err := h.adsService.GetAdsCampaigns(ctx, storeID, status, limit, offset)
	if err != nil {
		logutil.Errorf("Failed to get ads campaigns: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve campaigns"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"campaigns": campaigns,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
			"count":  len(campaigns),
		},
	})
}

// GetPerformanceSummary returns aggregated performance metrics
// GET /api/ads/summary?store_id=1&start_date=2024-01-01&end_date=2024-01-31
func (h *AdsPerformanceHandler) GetPerformanceSummary(c *gin.Context) {
	var storeID *int
	if storeIDStr := c.Query("store_id"); storeIDStr != "" {
		if id, err := strconv.Atoi(storeIDStr); err == nil {
			storeID = &id
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid store_id parameter"})
			return
		}
	}

	// Default to last 30 days if no dates provided
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if date, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = date
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format (use YYYY-MM-DD)"})
			return
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if date, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = date
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format (use YYYY-MM-DD)"})
			return
		}
	}

	ctx := context.Background()
	summary, err := h.adsService.GetPerformanceSummary(ctx, storeID, startDate, endDate)
	if err != nil {
		logutil.Errorf("Failed to get performance summary: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve summary"})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// FetchAdsCampaigns triggers fetching of campaigns from Shopee API
// POST /api/ads/campaigns/fetch
// Body: {"store_id": 1}
func (h *AdsPerformanceHandler) FetchAdsCampaigns(c *gin.Context) {
	var request struct {
		StoreID int `json:"store_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	ctx := context.Background()
	err := h.adsService.FetchAdsCampaigns(ctx, request.StoreID)
	if err != nil {
		logutil.Errorf("Failed to fetch campaigns from Shopee: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch campaigns"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Campaigns fetched successfully"})
}

// FetchAdsCampaignSettings triggers fetching of campaign settings from Shopee API
// POST /api/ads/campaigns/settings/fetch
// Body: {"store_id": 1, "campaign_ids": [123, 456]}
func (h *AdsPerformanceHandler) FetchAdsCampaignSettings(c *gin.Context) {
	var request struct {
		StoreID     int     `json:"store_id" binding:"required"`
		CampaignIDs []int64 `json:"campaign_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if len(request.CampaignIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Campaign IDs list cannot be empty"})
		return
	}

	ctx := context.Background()
	err := h.adsService.FetchAdsCampaignSettings(ctx, request.StoreID, request.CampaignIDs)
	if err != nil {
		logutil.Errorf("Failed to fetch campaign settings from Shopee: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch campaign settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "Campaign settings fetched successfully",
		"campaigns_count": len(request.CampaignIDs),
	})
}
// POST /api/ads/performance/fetch
// Body: {"store_id": 1, "campaign_id": 123, "start_date": "2024-01-01", "end_date": "2024-01-31"}
func (h *AdsPerformanceHandler) FetchAdsPerformance(c *gin.Context) {
	var request struct {
		StoreID    int    `json:"store_id" binding:"required"`
		CampaignID int64  `json:"campaign_id" binding:"required"`
		StartDate  string `json:"start_date" binding:"required"`
		EndDate    string `json:"end_date" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", request.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format (use YYYY-MM-DD)"})
		return
	}

	endDate, err := time.Parse("2006-01-02", request.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format (use YYYY-MM-DD)"})
		return
	}

	ctx := context.Background()
	err = h.adsService.FetchAdsPerformance(ctx, request.StoreID, request.CampaignID, startDate, endDate)
	if err != nil {
		logutil.Errorf("Failed to fetch performance data from Shopee: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch performance data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Performance data fetched successfully"})
}

// SyncHistoricalAdsPerformance triggers background sync of all historical ads performance data
// POST /api/ads/sync/historical
// Body: {"store_id": 1}
func (h *AdsPerformanceHandler) SyncHistoricalAdsPerformance(c *gin.Context) {
	var request struct {
		StoreID int `json:"store_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if h.batchScheduler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Batch scheduler not available"})
		return
	}

	ctx := context.Background()
	batchID, err := h.batchScheduler.CreateSyncBatch(ctx, request.StoreID)
	if err != nil {
		logutil.Errorf("Failed to create historical sync batch: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create sync batch"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Historical sync batch created successfully",
		"batch_id": batchID,
	})
}

// GetStoresWithShopeeCredentials returns stores that have Shopee credentials configured
// GET /api/ads/stores-with-credentials
func (h *AdsPerformanceHandler) GetStoresWithShopeeCredentials(c *gin.Context) {
	// This is a simple implementation. In a real app, you'd add this to the service layer
	// For now, we'll use the existing /api/stores/all endpoint and filter in frontend
	c.JSON(http.StatusOK, gin.H{
		"message": "Use /api/stores/all endpoint - frontend will filter stores with credentials",
	})
}

// RegisterRoutes registers all ads performance routes
func (h *AdsPerformanceHandler) RegisterRoutes(apiGroup *gin.RouterGroup) {
	adsGroup := apiGroup.Group("/ads")
	{
		adsGroup.GET("/campaigns", h.GetAdsCampaigns)
		adsGroup.GET("/summary", h.GetPerformanceSummary)
		adsGroup.GET("/stores-with-credentials", h.GetStoresWithShopeeCredentials)
		adsGroup.POST("/campaigns/fetch", h.FetchAdsCampaigns)
		adsGroup.POST("/campaigns/settings/fetch", h.FetchAdsCampaignSettings)
		adsGroup.POST("/performance/fetch", h.FetchAdsPerformance)
		adsGroup.POST("/sync/historical", h.SyncHistoricalAdsPerformance)
	}
}
