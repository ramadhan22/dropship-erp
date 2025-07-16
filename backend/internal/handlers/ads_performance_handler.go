package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/logutil"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/ramadhan22/dropship-erp/backend/internal/service"
)

// AdsPerformanceHandler handles ads performance related HTTP requests.
type AdsPerformanceHandler struct {
	service *service.AdsPerformanceService
}

// NewAdsPerformanceHandler creates a new handler instance.
func NewAdsPerformanceHandler(service *service.AdsPerformanceService) *AdsPerformanceHandler {
	return &AdsPerformanceHandler{service: service}
}

// GetAdsPerformance handles GET /api/ads-performance
func (h *AdsPerformanceHandler) GetAdsPerformance(c *gin.Context) {
	filter := &models.AdsPerformanceFilter{}

	// Parse query parameters
	if storeIDStr := c.Query("store_id"); storeIDStr != "" {
		if storeID, err := strconv.Atoi(storeIDStr); err == nil {
			filter.StoreID = &storeID
		}
	}

	if status := c.Query("campaign_status"); status != "" {
		filter.CampaignStatus = &status
	}

	if campaignType := c.Query("campaign_type"); campaignType != "" {
		filter.CampaignType = &campaignType
	}

	if dateFromStr := c.Query("date_from"); dateFromStr != "" {
		if dateFrom, err := time.Parse("2006-01-02", dateFromStr); err == nil {
			filter.DateFrom = &dateFrom
		}
	}

	if dateToStr := c.Query("date_to"); dateToStr != "" {
		if dateTo, err := time.Parse("2006-01-02", dateToStr); err == nil {
			filter.DateTo = &dateTo
		}
	}

	// Parse pagination
	limit := 50
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Get ads performance data
	ads, err := h.service.GetAdsPerformance(filter, limit, offset)
	if err != nil {
		logutil.Errorf("Failed to get ads performance: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get ads performance"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ads":    ads,
		"limit":  limit,
		"offset": offset,
		"count":  len(ads),
	})
}

// GetAdsPerformanceSummary handles GET /api/ads-performance/summary
func (h *AdsPerformanceHandler) GetAdsPerformanceSummary(c *gin.Context) {
	filter := &models.AdsPerformanceFilter{}

	// Parse query parameters (same as above)
	if storeIDStr := c.Query("store_id"); storeIDStr != "" {
		if storeID, err := strconv.Atoi(storeIDStr); err == nil {
			filter.StoreID = &storeID
		}
	}

	if status := c.Query("campaign_status"); status != "" {
		filter.CampaignStatus = &status
	}

	if dateFromStr := c.Query("date_from"); dateFromStr != "" {
		if dateFrom, err := time.Parse("2006-01-02", dateFromStr); err == nil {
			filter.DateFrom = &dateFrom
		}
	}

	if dateToStr := c.Query("date_to"); dateToStr != "" {
		if dateTo, err := time.Parse("2006-01-02", dateToStr); err == nil {
			filter.DateTo = &dateTo
		}
	}

	// Get summary
	summary, err := h.service.GetAdsPerformanceSummary(filter)
	if err != nil {
		logutil.Errorf("Failed to get ads performance summary: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get summary"})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// RefreshAdsData handles POST /api/ads-performance/refresh
func (h *AdsPerformanceHandler) RefreshAdsData(c *gin.Context) {
	var request struct {
		DateFrom string `json:"date_from" binding:"required"`
		DateTo   string `json:"date_to" binding:"required"`
		StoreID  *int   `json:"store_id,omitempty"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse dates
	dateFrom, err := time.Parse("2006-01-02", request.DateFrom)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date_from format"})
		return
	}

	dateTo, err := time.Parse("2006-01-02", request.DateTo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date_to format"})
		return
	}

	// Validate date range
	if dateTo.Before(dateFrom) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date_to must be after date_from"})
		return
	}

	// Limit to 30 days max
	if dateTo.Sub(dateFrom) > 30*24*time.Hour {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Date range cannot exceed 30 days"})
		return
	}

	// Refresh data
	if request.StoreID != nil {
		// Refresh for specific store
		err = h.service.FetchAdsPerformanceFromShopee(c.Request.Context(), *request.StoreID, dateFrom, dateTo)
	} else {
		// Refresh for all stores
		err = h.service.RefreshAdsData(c.Request.Context(), dateFrom, dateTo)
	}

	if err != nil {
		logutil.Errorf("Failed to refresh ads data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refresh data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Ads data refreshed successfully",
		"date_from": request.DateFrom,
		"date_to":   request.DateTo,
	})
}
