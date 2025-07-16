package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/logutil"
	"github.com/ramadhan22/dropship-erp/backend/internal/service"
)

// AdsPerformanceHandler handles HTTP requests for ads performance data
type AdsPerformanceHandler struct {
	adsService *service.AdsPerformanceService
}

// NewAdsPerformanceHandler creates a new ads performance handler
func NewAdsPerformanceHandler(adsService *service.AdsPerformanceService) *AdsPerformanceHandler {
	return &AdsPerformanceHandler{
		adsService: adsService,
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
// Body: {"store_id": 1, "access_token": "token"}
func (h *AdsPerformanceHandler) FetchAdsCampaigns(c *gin.Context) {
	var request struct {
		StoreID     int    `json:"store_id" binding:"required"`
		AccessToken string `json:"access_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	ctx := context.Background()
	err := h.adsService.FetchAdsCampaigns(ctx, request.StoreID, request.AccessToken)
	if err != nil {
		logutil.Errorf("Failed to fetch campaigns from Shopee: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch campaigns"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Campaigns fetched successfully"})
}

// FetchAdsPerformance triggers fetching of performance data from Shopee API
// POST /api/ads/performance/fetch
// Body: {"store_id": 1, "campaign_id": 123, "start_date": "2024-01-01", "end_date": "2024-01-31", "access_token": "token"}
func (h *AdsPerformanceHandler) FetchAdsPerformance(c *gin.Context) {
	var request struct {
		StoreID     int    `json:"store_id" binding:"required"`
		CampaignID  int64  `json:"campaign_id" binding:"required"`
		StartDate   string `json:"start_date" binding:"required"`
		EndDate     string `json:"end_date" binding:"required"`
		AccessToken string `json:"access_token" binding:"required"`
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
	err = h.adsService.FetchAdsPerformance(ctx, request.StoreID, request.CampaignID, startDate, endDate, request.AccessToken)
	if err != nil {
		logutil.Errorf("Failed to fetch performance data from Shopee: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch performance data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Performance data fetched successfully"})
}

// RegisterRoutes registers all ads performance routes
func (h *AdsPerformanceHandler) RegisterRoutes(apiGroup *gin.RouterGroup) {
	adsGroup := apiGroup.Group("/ads")
	{
		adsGroup.GET("/campaigns", h.GetAdsCampaigns)
		adsGroup.GET("/summary", h.GetPerformanceSummary)
		adsGroup.POST("/campaigns/fetch", h.FetchAdsCampaigns)
		adsGroup.POST("/performance/fetch", h.FetchAdsPerformance)
	}
}