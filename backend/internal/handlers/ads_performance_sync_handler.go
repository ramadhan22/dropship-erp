package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/logutil"
	"github.com/ramadhan22/dropship-erp/backend/internal/service"
)

// AdsPerformanceSyncHandler handles ads performance sync related HTTP requests.
type AdsPerformanceSyncHandler struct {
	syncService *service.AdsPerformanceSyncService
}

// NewAdsPerformanceSyncHandler creates a new sync handler instance.
func NewAdsPerformanceSyncHandler(syncService *service.AdsPerformanceSyncService) *AdsPerformanceSyncHandler {
	return &AdsPerformanceSyncHandler{syncService: syncService}
}

// TriggerSync handles POST /api/ads-performance/sync
func (h *AdsPerformanceSyncHandler) TriggerSync(c *gin.Context) {
	var request struct {
		StoreID   int    `json:"store_id" binding:"required"`
		StartDate string `json:"start_date,omitempty"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	var startDate time.Time
	var err error
	
	if request.StartDate != "" {
		// Parse custom start date
		startDate, err = time.Parse("2006-01-02", request.StartDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format"})
			return
		}
	} else {
		// Default: start from yesterday
		startDate = time.Now().AddDate(0, 0, -1).Truncate(24 * time.Hour)
	}
	
	// Create sync job
	job, err := h.syncService.CreateSyncJob(c.Request.Context(), request.StoreID, startDate)
	if err != nil {
		logutil.Errorf("Failed to create sync job: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create sync job"})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{
		"message": "Sync job created successfully",
		"job_id":  job.ID,
		"store_id": job.StoreID,
		"start_date": job.StartDate.Format("2006-01-02"),
		"status": job.Status,
	})
}

// TriggerFullHistorySync handles POST /api/ads-performance/sync-all
func (h *AdsPerformanceSyncHandler) TriggerFullHistorySync(c *gin.Context) {
	var request struct {
		StoreID int `json:"store_id" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Create full history sync job
	job, err := h.syncService.CreateFullHistorySyncJob(c.Request.Context(), request.StoreID)
	if err != nil {
		logutil.Errorf("Failed to create full history sync job: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create sync job"})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{
		"message": "Full history sync job created successfully",
		"job_id":  job.ID,
		"store_id": job.StoreID,
		"start_date": job.StartDate.Format("2006-01-02"),
		"status": job.Status,
	})
}

// GetSyncJob handles GET /api/ads-performance/sync/:id
func (h *AdsPerformanceSyncHandler) GetSyncJob(c *gin.Context) {
	jobIDStr := c.Param("id")
	jobID, err := strconv.ParseInt(jobIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}
	
	job, err := h.syncService.GetSyncJob(jobID)
	if err != nil {
		logutil.Errorf("Failed to get sync job %d: %v", jobID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Sync job not found"})
		return
	}
	
	c.JSON(http.StatusOK, job)
}

// ListSyncJobs handles GET /api/ads-performance/sync
func (h *AdsPerformanceSyncHandler) ListSyncJobs(c *gin.Context) {
	var storeID *int
	if storeIDStr := c.Query("store_id"); storeIDStr != "" {
		if id, err := strconv.Atoi(storeIDStr); err == nil {
			storeID = &id
		}
	}
	
	// Parse pagination
	limit := 20
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
	
	jobs, err := h.syncService.ListSyncJobs(storeID, limit, offset)
	if err != nil {
		logutil.Errorf("Failed to list sync jobs: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list sync jobs"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"jobs":   jobs,
		"limit":  limit,
		"offset": offset,
		"count":  len(jobs),
	})
}

// RegisterRoutes registers all sync-related routes.
func (h *AdsPerformanceSyncHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/ads-performance/sync", h.TriggerSync)
	router.POST("/ads-performance/sync-all", h.TriggerFullHistorySync)
	router.GET("/ads-performance/sync/:id", h.GetSyncJob)
	router.GET("/ads-performance/sync", h.ListSyncJobs)
}