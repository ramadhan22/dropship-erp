package handlers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/ramadhan22/dropship-erp/backend/internal/service"
)

// BulkImportHandler handles bulk import operations with enhanced performance
type BulkImportHandler struct {
	dropshipService    *service.DropshipService
	batchService       *service.BatchService
	streamingProcessor *service.StreamingImportProcessor
	enhancedScheduler  *service.EnhancedImportScheduler
}

// NewBulkImportHandler creates a new bulk import handler
func NewBulkImportHandler(
	dropshipService *service.DropshipService,
	batchService *service.BatchService,
	streamingProcessor *service.StreamingImportProcessor,
	enhancedScheduler *service.EnhancedImportScheduler,
) *BulkImportHandler {
	return &BulkImportHandler{
		dropshipService:    dropshipService,
		batchService:       batchService,
		streamingProcessor: streamingProcessor,
		enhancedScheduler:  enhancedScheduler,
	}
}

// HandleBulkImport handles bulk import of multiple files with streaming processing
func (h *BulkImportHandler) HandleBulkImport(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "files are required"})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no files provided"})
		return
	}

	// Get optional parameters
	channel := c.PostForm("channel")
	useStreaming := c.PostForm("use_streaming") == "true"
	processConcurrently := c.PostForm("process_concurrently") == "true"

	// Validate file count
	if len(files) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "maximum 100 files allowed per batch"})
		return
	}

	// Save files and create batch records
	var filePaths []string
	var batchIDs []int64

	for _, fileHeader := range files {
		// Validate file size (100MB limit)
		if fileHeader.Size > 100*1024*1024 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("file %s exceeds 100MB limit", fileHeader.Filename),
			})
			return
		}

		// Save file
		dir := filepath.Join("backend", "uploads", "bulk_dropship")
		if err := os.MkdirAll(dir, 0755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create upload directory"})
			return
		}

		filename := fmt.Sprintf("%s_%s", time.Now().Format("20060102150405"), fileHeader.Filename)
		filePath := filepath.Join(dir, filename)

		if err := c.SaveUploadedFile(fileHeader, filePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("failed to save file %s", fileHeader.Filename),
			})
			return
		}

		filePaths = append(filePaths, filePath)

		// Create batch record
		processType := "dropship_import"
		if useStreaming {
			processType = "streaming_dropship_import"
		}

		batch := &models.BatchHistory{
			ProcessType: processType,
			TotalData:   0,
			DoneData:    0,
			Status:      "pending",
			FileName:    fileHeader.Filename,
			FilePath:    filePath,
		}

		batchID, err := h.batchService.Create(context.Background(), batch)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("failed to create batch record for %s", fileHeader.Filename),
			})
			return
		}

		batchIDs = append(batchIDs, batchID)
	}

	// Process files
	if processConcurrently && h.streamingProcessor != nil {
		// Process files concurrently using streaming processor
		go func() {
			ctx := context.Background()
			if err := h.streamingProcessor.ProcessMultipleFiles(ctx, filePaths, channel); err != nil {
				// Log error but don't fail the request since it's async
				fmt.Printf("Error processing files concurrently: %v\n", err)
			}
		}()
	}

	response := gin.H{
		"queued_files":         len(files),
		"batch_ids":            batchIDs,
		"use_streaming":        useStreaming,
		"process_concurrently": processConcurrently,
		"estimated_time":       h.estimateProcessingTime(len(files)),
	}

	c.JSON(http.StatusOK, response)
}

// HandleImportStatus returns the status of import operations
func (h *BulkImportHandler) HandleImportStatus(c *gin.Context) {
	batchIDStr := c.Param("batch_id")
	batchID, err := strconv.ParseInt(batchIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid batch_id"})
		return
	}

	// Get batch details
	batches, err := h.batchService.List(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get batch status"})
		return
	}

	var targetBatch *models.BatchHistory
	for _, batch := range batches {
		if batch.ID == batchID {
			targetBatch = &batch
			break
		}
	}

	if targetBatch == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "batch not found"})
		return
	}

	// Calculate progress
	progress := 0.0
	if targetBatch.TotalData > 0 {
		progress = float64(targetBatch.DoneData) / float64(targetBatch.TotalData) * 100
	}

	// Get estimated time remaining
	var etaStr string
	if h.streamingProcessor != nil && progress > 0 && progress < 100 {
		if eta := h.streamingProcessor.EstimateRemainingTime(); eta > 0 {
			etaStr = eta.String()
		}
	}

	response := gin.H{
		"batch_id":       targetBatch.ID,
		"file_name":      targetBatch.FileName,
		"status":         targetBatch.Status,
		"progress":       progress,
		"rows_processed": targetBatch.DoneData,
		"total_rows":     targetBatch.TotalData,
		"created_at":     targetBatch.CreatedAt,
		"updated_at":     targetBatch.UpdatedAt,
		"estimated_eta":  etaStr,
	}

	if targetBatch.ErrorMessage != "" {
		response["error"] = targetBatch.ErrorMessage
	}

	c.JSON(http.StatusOK, response)
}

// HandleBulkImportStatus returns the status of all active import operations
func (h *BulkImportHandler) HandleBulkImportStatus(c *gin.Context) {
	var response gin.H

	// Get scheduler status if available
	if h.enhancedScheduler != nil {
		activeJobs := h.enhancedScheduler.GetActiveJobs()
		queueStatus := h.enhancedScheduler.GetQueueStatus()

		response = gin.H{
			"active_jobs":  activeJobs,
			"queue_status": queueStatus,
			"total_active": len(activeJobs),
		}
	} else {
		// Fallback to batch service
		batches, err := h.batchService.ListFiltered(context.Background(),
			[]string{"dropship_import", "streaming_dropship_import"},
			[]string{"pending", "processing"})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get batch status"})
			return
		}

		response = gin.H{
			"active_batches": batches,
			"total_active":   len(batches),
		}
	}

	// Add streaming processor stats if available
	if h.streamingProcessor != nil {
		stats := h.streamingProcessor.GetStats()
		response["streaming_stats"] = gin.H{
			"total_files":     stats.TotalFiles,
			"processed_files": stats.ProcessedFiles,
			"failed_files":    stats.FailedFiles,
			"total_rows":      stats.TotalRows,
			"processed_rows":  stats.ProcessedRows,
			"failed_rows":     stats.FailedRows,
			"start_time":      stats.StartTime,
			"last_update":     stats.LastUpdateTime,
		}
	}

	c.JSON(http.StatusOK, response)
}

// HandleForceProcessBatch forces processing of a specific batch
func (h *BulkImportHandler) HandleForceProcessBatch(c *gin.Context) {
	batchIDStr := c.Param("batch_id")
	batchID, err := strconv.ParseInt(batchIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid batch_id"})
		return
	}

	if h.enhancedScheduler != nil {
		if err := h.enhancedScheduler.ForceProcessBatch(batchID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "enhanced scheduler not available"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "batch queued for processing"})
}

// HandleImportRecommendations provides recommendations for import optimization
func (h *BulkImportHandler) HandleImportRecommendations(c *gin.Context) {
	fileCountStr := c.Query("file_count")
	avgFileSizeStr := c.Query("avg_file_size")

	fileCount, _ := strconv.Atoi(fileCountStr)
	avgFileSize, _ := strconv.ParseInt(avgFileSizeStr, 10, 64)

	recommendations := h.generateRecommendations(fileCount, avgFileSize)

	c.JSON(http.StatusOK, gin.H{
		"recommendations": recommendations,
		"estimated_time":  h.estimateProcessingTime(fileCount),
	})
}

// estimateProcessingTime estimates the total processing time for a given number of files
func (h *BulkImportHandler) estimateProcessingTime(fileCount int) string {
	// Base estimates (these could be made more sophisticated)
	baseTimePerFile := 30 * time.Second // 30 seconds per file

	// Adjust for concurrency if streaming processor is available
	if h.streamingProcessor != nil {
		// With streaming and concurrency, reduce time significantly
		baseTimePerFile = 10 * time.Second

		// Account for concurrent processing
		concurrentFiles := 5 // Default concurrent files
		estimatedTime := time.Duration(fileCount/concurrentFiles) * baseTimePerFile
		if fileCount%concurrentFiles > 0 {
			estimatedTime += baseTimePerFile
		}

		return estimatedTime.String()
	}

	// Sequential processing
	estimatedTime := time.Duration(fileCount) * baseTimePerFile
	return estimatedTime.String()
}

// generateRecommendations generates optimization recommendations based on file characteristics
func (h *BulkImportHandler) generateRecommendations(fileCount int, avgFileSize int64) []string {
	var recommendations []string

	if fileCount > 50 {
		recommendations = append(recommendations, "Consider enabling concurrent processing for better performance with many files")
	}

	if avgFileSize > 50*1024*1024 { // 50MB
		recommendations = append(recommendations, "Large files detected - streaming processing is recommended")
	}

	if fileCount > 10 && avgFileSize > 10*1024*1024 { // 10MB
		recommendations = append(recommendations, "Enable both streaming and concurrent processing for optimal performance")
	}

	if fileCount > 100 {
		recommendations = append(recommendations, "Consider splitting import into smaller batches for better reliability")
	}

	if h.streamingProcessor != nil {
		recommendations = append(recommendations, "Streaming processor is available - use 'use_streaming=true' for better memory efficiency")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Current settings should work well for your file characteristics")
	}

	return recommendations
}
