package service

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/logutil"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"golang.org/x/sync/errgroup"
)

// ReconcileStreamConfig defines configuration for streaming reconciliation
type ReconcileStreamConfig struct {
	// ChunkSize is the number of records to process in each chunk
	ChunkSize int
	
	// MaxConcurrency is the maximum number of concurrent chunk processors
	MaxConcurrency int
	
	// MemoryThreshold is the memory limit in bytes before forcing garbage collection
	MemoryThreshold int64
	
	// ProgressReportInterval is how often to report progress
	ProgressReportInterval time.Duration
	
	// RetryAttempts is the number of retry attempts for failed chunks
	RetryAttempts int
	
	// RetryDelay is the delay between retry attempts
	RetryDelay time.Duration
	
	// TimeoutPerChunk is the timeout for processing each chunk
	TimeoutPerChunk time.Duration
}

// DefaultReconcileStreamConfig returns a default configuration for streaming reconciliation
func DefaultReconcileStreamConfig() *ReconcileStreamConfig {
	return &ReconcileStreamConfig{
		ChunkSize:              1000,
		MaxConcurrency:         5,
		MemoryThreshold:        500 * 1024 * 1024, // 500MB
		ProgressReportInterval: 30 * time.Second,
		RetryAttempts:          3,
		RetryDelay:             5 * time.Second,
		TimeoutPerChunk:        10 * time.Minute,
	}
}

// ReconcileProgress tracks progress of reconciliation operations
type ReconcileProgress struct {
	TotalRecords      int64
	ProcessedRecords  int64
	SuccessfulRecords int64
	FailedRecords     int64
	CurrentChunk      int
	TotalChunks       int
	StartTime         time.Time
	LastUpdate        time.Time
	EstimatedTimeLeft time.Duration
	CurrentRate       float64 // records per second
	ErrorRate         float64 // percentage of failed records
}

// ReconcileChunkResult represents the result of processing a chunk
type ReconcileChunkResult struct {
	ChunkIndex    int
	ProcessedAt   time.Time
	ProcessedRows int
	SuccessRows   int
	FailedRows    int
	Duration      time.Duration
	Error         error
	FailedRecords []models.FailedReconciliation
}

// ReconcileStreamResult represents the final result of streaming reconciliation
type ReconcileStreamResult struct {
	TotalProcessed    int64
	TotalSuccessful   int64
	TotalFailed       int64
	StartTime         time.Time
	EndTime           time.Time
	Duration          time.Duration
	ChunkResults      []ReconcileChunkResult
	ProgressSnapshots []ReconcileProgress
	FinalReport       *models.ReconciliationReport
}

// ReconcileStreamProcessor handles streaming reconciliation of large datasets
type ReconcileStreamProcessor struct {
	service        *ReconcileService
	config         *ReconcileStreamConfig
	logger         *logutil.Logger
	progressMu     sync.RWMutex
	progress       *ReconcileProgress
	progressChan   chan ReconcileProgress
	stopChan       chan struct{}
	resultChan     chan ReconcileChunkResult
}

// NewReconcileStreamProcessor creates a new streaming reconciliation processor
func NewReconcileStreamProcessor(service *ReconcileService, config *ReconcileStreamConfig) *ReconcileStreamProcessor {
	if config == nil {
		config = DefaultReconcileStreamConfig()
	}
	
	return &ReconcileStreamProcessor{
		service:      service,
		config:       config,
		logger:       logutil.NewLogger("reconcile-stream", logutil.INFO),
		progressChan: make(chan ReconcileProgress, 100),
		stopChan:     make(chan struct{}),
		resultChan:   make(chan ReconcileChunkResult, config.MaxConcurrency*2),
	}
}

// StreamReconcileAll processes reconciliation for millions of records using streaming
func (p *ReconcileStreamProcessor) StreamReconcileAll(ctx context.Context, shop string, filters map[string]interface{}) (*ReconcileStreamResult, error) {
	timer := p.logger.WithTimer(ctx, "StreamReconcileAll")
	defer timer.Finish("Stream reconciliation completed")

	ctx = logutil.WithShop(ctx, shop)
	
	// Initialize progress tracking
	p.initializeProgress(ctx)
	defer p.cleanup()

	// Start progress monitoring
	go p.monitorProgress(ctx)

	// Get total count for progress tracking
	totalCount, err := p.getTotalRecordsCount(ctx, shop, filters)
	if err != nil {
		timer.FinishWithError("Failed to get total records count", err)
		return nil, fmt.Errorf("failed to get total records count: %w", err)
	}

	p.logger.Info(ctx, "StreamReconcileAll", "Starting stream reconciliation", map[string]interface{}{
		"shop":         shop,
		"total_count":  totalCount,
		"chunk_size":   p.config.ChunkSize,
		"concurrency":  p.config.MaxConcurrency,
	})

	// Update progress with total count
	p.updateProgress(func(progress *ReconcileProgress) {
		progress.TotalRecords = totalCount
		progress.TotalChunks = int(math.Ceil(float64(totalCount) / float64(p.config.ChunkSize)))
	})

	// Create chunk processor
	result := &ReconcileStreamResult{
		StartTime:         time.Now(),
		ChunkResults:      make([]ReconcileChunkResult, 0),
		ProgressSnapshots: make([]ReconcileProgress, 0),
	}

	// Process chunks concurrently
	if err := p.processChunks(ctx, shop, filters, result); err != nil {
		timer.FinishWithError("Failed to process chunks", err)
		return nil, fmt.Errorf("failed to process chunks: %w", err)
	}

	// Finalize result
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.FinalReport = p.buildFinalReport(result)

	p.logger.Info(ctx, "StreamReconcileAll", "Stream reconciliation completed", map[string]interface{}{
		"total_processed":  result.TotalProcessed,
		"total_successful": result.TotalSuccessful,
		"total_failed":     result.TotalFailed,
		"duration":         result.Duration,
		"error_rate":       float64(result.TotalFailed) / float64(result.TotalProcessed) * 100,
	})

	return result, nil
}

// processChunks processes reconciliation in chunks using concurrent workers
func (p *ReconcileStreamProcessor) processChunks(ctx context.Context, shop string, filters map[string]interface{}, result *ReconcileStreamResult) error {
	// Create error group for concurrent processing
	eg, egCtx := errgroup.WithContext(ctx)
	eg.SetLimit(p.config.MaxConcurrency)

	// Channel to collect results
	resultCollector := make(chan ReconcileChunkResult, p.config.MaxConcurrency*2)
	
	// Start result collector goroutine
	go func() {
		defer close(resultCollector)
		for chunkResult := range p.resultChan {
			select {
			case resultCollector <- chunkResult:
			case <-egCtx.Done():
				return
			}
		}
	}()

	// Process chunks
	chunkIndex := 0
	offset := 0
	
	for {
		select {
		case <-egCtx.Done():
			return egCtx.Err()
		default:
		}

		// Check if we've processed all records
		if offset >= int(p.getProgress().TotalRecords) {
			break
		}

		// Create chunk context with timeout
		chunkCtx, cancel := context.WithTimeout(egCtx, p.config.TimeoutPerChunk)
		currentChunkIndex := chunkIndex
		currentOffset := offset

		// Submit chunk processing job
		eg.Go(func() error {
			defer cancel()
			return p.processChunk(chunkCtx, shop, filters, currentChunkIndex, currentOffset, p.config.ChunkSize)
		})

		chunkIndex++
		offset += p.config.ChunkSize
	}

	// Wait for all chunks to complete
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("chunk processing failed: %w", err)
	}

	// Close result channel and collect all results
	close(p.resultChan)
	for chunkResult := range resultCollector {
		result.ChunkResults = append(result.ChunkResults, chunkResult)
		result.TotalProcessed += int64(chunkResult.ProcessedRows)
		result.TotalSuccessful += int64(chunkResult.SuccessRows)
		result.TotalFailed += int64(chunkResult.FailedRows)
	}

	return nil
}

// processChunk processes a single chunk of reconciliation records
func (p *ReconcileStreamProcessor) processChunk(ctx context.Context, shop string, filters map[string]interface{}, chunkIndex, offset, limit int) error {
	chunkTimer := p.logger.WithTimer(ctx, "ProcessChunk")
	defer chunkTimer.Finish("Chunk processing completed")

	chunkCtx := logutil.WithOperation(ctx, fmt.Sprintf("chunk-%d", chunkIndex))
	
	p.logger.Debug(chunkCtx, "ProcessChunk", "Starting chunk processing", map[string]interface{}{
		"chunk_index": chunkIndex,
		"offset":      offset,
		"limit":       limit,
	})

	startTime := time.Now()
	result := ReconcileChunkResult{
		ChunkIndex:  chunkIndex,
		ProcessedAt: startTime,
	}

	// Get chunk data
	candidates, err := p.getChunkData(chunkCtx, shop, filters, offset, limit)
	if err != nil {
		result.Error = err
		result.Duration = time.Since(startTime)
		p.resultChan <- result
		return fmt.Errorf("failed to get chunk data: %w", err)
	}

	result.ProcessedRows = len(candidates)
	
	if len(candidates) == 0 {
		result.Duration = time.Since(startTime)
		p.resultChan <- result
		return nil
	}

	// Process reconciliation pairs with error handling
	pairs := make([][2]string, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.NoPesanan != nil && *candidate.NoPesanan != "" {
			pairs = append(pairs, [2]string{candidate.KodeInvoiceChannel, *candidate.NoPesanan})
		}
	}

	if len(pairs) > 0 {
		// Create a batch for this chunk
		batchID, err := p.createChunkBatch(chunkCtx, shop, chunkIndex, len(pairs))
		if err != nil {
			p.logger.Error(chunkCtx, "ProcessChunk", "Failed to create batch", err)
		}

		// Process reconciliation with error handling
		report, err := p.service.BulkReconcileWithErrorHandling(chunkCtx, pairs, shop, &batchID)
		if err != nil {
			result.Error = err
			p.logger.Error(chunkCtx, "ProcessChunk", "Bulk reconciliation failed", err)
		} else {
			result.SuccessRows = report.SuccessfulTransactions
			result.FailedRows = report.FailedTransactions
			result.FailedRecords = report.FailedTransactionList
		}
	}

	result.Duration = time.Since(startTime)
	
	// Update progress
	p.updateProgress(func(progress *ReconcileProgress) {
		progress.ProcessedRecords += int64(result.ProcessedRows)
		progress.SuccessfulRecords += int64(result.SuccessRows)
		progress.FailedRecords += int64(result.FailedRows)
		progress.CurrentChunk = chunkIndex + 1
		progress.LastUpdate = time.Now()
		
		// Calculate rate and ETA
		elapsed := progress.LastUpdate.Sub(progress.StartTime)
		if elapsed > 0 {
			progress.CurrentRate = float64(progress.ProcessedRecords) / elapsed.Seconds()
			if progress.CurrentRate > 0 {
				remaining := progress.TotalRecords - progress.ProcessedRecords
				progress.EstimatedTimeLeft = time.Duration(float64(remaining)/progress.CurrentRate) * time.Second
			}
		}
		
		if progress.ProcessedRecords > 0 {
			progress.ErrorRate = float64(progress.FailedRecords) / float64(progress.ProcessedRecords) * 100
		}
	})

	p.logger.Info(chunkCtx, "ProcessChunk", "Chunk processing completed", map[string]interface{}{
		"chunk_index":    chunkIndex,
		"processed_rows": result.ProcessedRows,
		"success_rows":   result.SuccessRows,
		"failed_rows":    result.FailedRows,
		"duration":       result.Duration,
		"records_per_sec": float64(result.ProcessedRows) / result.Duration.Seconds(),
	})

	p.resultChan <- result
	return nil
}

// Helper methods for the streaming processor

func (p *ReconcileStreamProcessor) initializeProgress(ctx context.Context) {
	p.progressMu.Lock()
	defer p.progressMu.Unlock()
	
	p.progress = &ReconcileProgress{
		StartTime:  time.Now(),
		LastUpdate: time.Now(),
	}
}

func (p *ReconcileStreamProcessor) updateProgress(updateFunc func(*ReconcileProgress)) {
	p.progressMu.Lock()
	defer p.progressMu.Unlock()
	
	updateFunc(p.progress)
	
	// Send progress update to channel (non-blocking)
	select {
	case p.progressChan <- *p.progress:
	default:
	}
}

func (p *ReconcileStreamProcessor) getProgress() ReconcileProgress {
	p.progressMu.RLock()
	defer p.progressMu.RUnlock()
	
	return *p.progress
}

func (p *ReconcileStreamProcessor) monitorProgress(ctx context.Context) {
	ticker := time.NewTicker(p.config.ProgressReportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-p.stopChan:
			return
		case <-ticker.C:
			progress := p.getProgress()
			p.logger.Info(ctx, "ProgressMonitor", "Progress update", map[string]interface{}{
				"processed":     progress.ProcessedRecords,
				"total":         progress.TotalRecords,
				"successful":    progress.SuccessfulRecords,
				"failed":        progress.FailedRecords,
				"current_chunk": progress.CurrentChunk,
				"total_chunks":  progress.TotalChunks,
				"rate":          progress.CurrentRate,
				"error_rate":    progress.ErrorRate,
				"eta":           progress.EstimatedTimeLeft,
			})
		}
	}
}

func (p *ReconcileStreamProcessor) cleanup() {
	close(p.stopChan)
}

func (p *ReconcileStreamProcessor) getTotalRecordsCount(ctx context.Context, shop string, filters map[string]interface{}) (int64, error) {
	// This would need to be implemented based on your repository interface
	// For now, return a placeholder
	return 0, fmt.Errorf("not implemented - need to add count method to repository")
}

func (p *ReconcileStreamProcessor) getChunkData(ctx context.Context, shop string, filters map[string]interface{}, offset, limit int) ([]models.ReconcileCandidate, error) {
	// This would need to be implemented based on your repository interface
	// Use the existing ListCandidates method but with chunking
	if repo, ok := p.service.recRepo.(interface {
		ListCandidates(context.Context, string, string, string, string, string, int, int) ([]models.ReconcileCandidate, int, error)
	}); ok {
		candidates, _, err := repo.ListCandidates(ctx, shop, "", "", "", "", limit, offset)
		return candidates, err
	}
	return nil, fmt.Errorf("repository does not support chunked listing")
}

func (p *ReconcileStreamProcessor) createChunkBatch(ctx context.Context, shop string, chunkIndex, recordCount int) (int64, error) {
	batch := &models.BatchHistory{
		ProcessType: "reconcile_chunk",
		Status:      "pending",
		TotalData:   recordCount,
		DoneData:    0,
		ErrorMsg:    fmt.Sprintf("Chunk %d reconciliation", chunkIndex),
		StartedAt:   time.Now().Format(time.RFC3339),
	}
	
	return p.service.batchSvc.Create(ctx, batch)
}

func (p *ReconcileStreamProcessor) buildFinalReport(result *ReconcileStreamResult) *models.ReconciliationReport {
	report := &models.ReconciliationReport{
		TotalTransactions:      int(result.TotalProcessed),
		SuccessfulTransactions: int(result.TotalSuccessful),
		FailedTransactions:     int(result.TotalFailed),
		ProcessingStartTime:    result.StartTime,
		ProcessingEndTime:      result.EndTime,
		FailureCategories:      make(map[string]int),
		FailedTransactionList:  make([]models.FailedReconciliation, 0),
	}

	// Aggregate failed transactions from all chunks
	for _, chunkResult := range result.ChunkResults {
		for _, failed := range chunkResult.FailedRecords {
			report.FailedTransactionList = append(report.FailedTransactionList, failed)
			
			// Count failure categories
			if failed.ErrorType != "" {
				report.FailureCategories[failed.ErrorType]++
			}
		}
	}

	return report
}