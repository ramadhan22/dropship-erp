package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/logutil"
	"github.com/ramadhan22/dropship-erp/backend/internal/service"
)

// ReconcileStreamHandler handles streaming reconciliation requests
type ReconcileStreamHandler struct {
	reconcileService *service.ReconcileService
	logger           *logutil.Logger
}

// NewReconcileStreamHandler creates a new handler for streaming reconciliation
func NewReconcileStreamHandler(reconcileService *service.ReconcileService) *ReconcileStreamHandler {
	return &ReconcileStreamHandler{
		reconcileService: reconcileService,
		logger:           logutil.NewLogger("reconcile-stream-handler", logutil.INFO),
	}
}

// StreamReconcileAllRequest represents the request body for streaming reconciliation
type StreamReconcileAllRequest struct {
	Shop           string `json:"shop" binding:"required"`
	ChunkSize      int    `json:"chunk_size,omitempty"`
	MaxConcurrency int    `json:"max_concurrency,omitempty"`
	MemoryThreshold int64 `json:"memory_threshold,omitempty"`
}

// StreamReconcileAllResponse represents the response for streaming reconciliation
type StreamReconcileAllResponse struct {
	Success           bool                              `json:"success"`
	Message           string                            `json:"message"`
	TotalProcessed    int64                             `json:"total_processed"`
	TotalSuccessful   int64                             `json:"total_successful"`
	TotalFailed       int64                             `json:"total_failed"`
	Duration          string                            `json:"duration"`
	ErrorRate         float64                           `json:"error_rate"`
	ProcessingStarted time.Time                         `json:"processing_started"`
	ProcessingEnded   time.Time                         `json:"processing_ended"`
	FailureCategories map[string]int                    `json:"failure_categories,omitempty"`
	CorrelationID     string                            `json:"correlation_id"`
}

// HandleStreamReconcileAll handles streaming reconciliation of all records
func (h *ReconcileStreamHandler) HandleStreamReconcileAll(c *gin.Context) {
	ctx := c.Request.Context()
	correlationID := logutil.GetCorrelationID(ctx)
	
	timer := h.logger.WithTimer(ctx, "HandleStreamReconcileAll")
	defer timer.Finish("Stream reconcile all request completed")

	var req StreamReconcileAllRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error(ctx, "HandleStreamReconcileAll", "Invalid request body", err, map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body: " + err.Error(),
			"correlation_id": correlationID,
		})
		return
	}

	h.logger.Info(ctx, "HandleStreamReconcileAll", "Starting stream reconciliation", map[string]interface{}{
		"shop":            req.Shop,
		"chunk_size":      req.ChunkSize,
		"max_concurrency": req.MaxConcurrency,
	})

	// Create streaming configuration
	config := service.DefaultReconcileStreamConfig()
	if req.ChunkSize > 0 {
		config.ChunkSize = req.ChunkSize
	}
	if req.MaxConcurrency > 0 {
		config.MaxConcurrency = req.MaxConcurrency
	}
	if req.MemoryThreshold > 0 {
		config.MemoryThreshold = req.MemoryThreshold
	}

	// Process streaming reconciliation
	result, err := h.reconcileService.StreamReconcileAll(ctx, req.Shop, config)
	if err != nil {
		h.logger.Error(ctx, "HandleStreamReconcileAll", "Stream reconciliation failed", err, map[string]interface{}{
			"shop": req.Shop,
		})
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Stream reconciliation failed: " + err.Error(),
			"correlation_id": correlationID,
		})
		return
	}

	// Calculate error rate
	errorRate := float64(0)
	if result.TotalProcessed > 0 {
		errorRate = float64(result.TotalFailed) / float64(result.TotalProcessed) * 100
	}

	h.logger.Info(ctx, "HandleStreamReconcileAll", "Stream reconciliation completed", map[string]interface{}{
		"shop":             req.Shop,
		"total_processed":  result.TotalProcessed,
		"total_successful": result.TotalSuccessful,
		"total_failed":     result.TotalFailed,
		"duration":         result.Duration,
		"error_rate":       errorRate,
	})

	// Build response
	response := StreamReconcileAllResponse{
		Success:           true,
		Message:           "Stream reconciliation completed successfully",
		TotalProcessed:    result.TotalProcessed,
		TotalSuccessful:   result.TotalSuccessful,
		TotalFailed:       result.TotalFailed,
		Duration:          result.Duration.String(),
		ErrorRate:         errorRate,
		ProcessingStarted: result.StartTime,
		ProcessingEnded:   result.EndTime,
		CorrelationID:     correlationID,
	}

	if result.FinalReport != nil {
		response.FailureCategories = result.FinalReport.FailureCategories
	}

	c.JSON(http.StatusOK, response)
}

// HandleReconcileProgress handles progress monitoring for streaming reconciliation
func (h *ReconcileStreamHandler) HandleReconcileProgress(c *gin.Context) {
	ctx := c.Request.Context()
	correlationID := logutil.GetCorrelationID(ctx)
	
	batchIDParam := c.Param("batch_id")
	batchID, err := strconv.ParseInt(batchIDParam, 10, 64)
	if err != nil {
		h.logger.Error(ctx, "HandleReconcileProgress", "Invalid batch ID", err, map[string]interface{}{
			"batch_id_param": batchIDParam,
		})
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid batch ID",
			"correlation_id": correlationID,
		})
		return
	}

	// Note: This would need to be implemented in the service layer
	// For now, return a placeholder response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Progress monitoring not yet implemented",
		"batch_id": batchID,
		"correlation_id": correlationID,
	})
}

// HandleReconcileHealthCheck handles health check for reconciliation service
func (h *ReconcileStreamHandler) HandleReconcileHealthCheck(c *gin.Context) {
	ctx := c.Request.Context()
	correlationID := logutil.GetCorrelationID(ctx)
	
	h.logger.Info(ctx, "HandleReconcileHealthCheck", "Health check requested")
	
	// Basic health check - in production, this would check database connectivity,
	// memory usage, active processes, etc.
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Reconciliation service is healthy",
		"timestamp": time.Now(),
		"correlation_id": correlationID,
	})
}

// RegisterRoutes registers the streaming reconciliation routes
func (h *ReconcileStreamHandler) RegisterRoutes(router *gin.RouterGroup) {
	reconcileGroup := router.Group("/reconcile")
	{
		reconcileGroup.POST("/stream", h.HandleStreamReconcileAll)
		reconcileGroup.GET("/progress/:batch_id", h.HandleReconcileProgress)
		reconcileGroup.GET("/health", h.HandleReconcileHealthCheck)
	}
}