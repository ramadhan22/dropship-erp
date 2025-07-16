package service

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// AdsPerformanceBatchScheduler processes pending ads performance sync batches in the background.
type AdsPerformanceBatchScheduler struct {
	batch    *BatchService
	svc      *AdsPerformanceService
	interval time.Duration
}

// NewAdsPerformanceBatchScheduler creates a scheduler with the given interval.
func NewAdsPerformanceBatchScheduler(batch *BatchService, svc *AdsPerformanceService, interval time.Duration) *AdsPerformanceBatchScheduler {
	if interval <= 0 {
		interval = time.Minute
	}
	return &AdsPerformanceBatchScheduler{batch: batch, svc: svc, interval: interval}
}

// Start launches the scheduler loop.
func (s *AdsPerformanceBatchScheduler) Start(ctx context.Context) {
	if s == nil {
		return
	}
	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.run(ctx)
			}
		}
	}()
}

func (s *AdsPerformanceBatchScheduler) run(ctx context.Context) {
	list, err := s.batch.ListPendingByType(ctx, "ads_performance_sync")
	if err != nil {
		log.Printf("ads performance scheduler list pending: %v", err)
		return
	}
	
	for _, b := range list {
		s.processBatch(ctx, b)
	}
}

func (s *AdsPerformanceBatchScheduler) processBatch(ctx context.Context, batch models.BatchHistory) {
	log.Printf("Processing ads performance sync batch %d", batch.ID)
	
	// Update batch status to processing
	err := s.batch.UpdateStatus(ctx, batch.ID, "processing", "Starting ads performance sync")
	if err != nil {
		log.Printf("Failed to update batch status: %v", err)
		return
	}

	// Parse batch request from file path or other source
	var syncRequest struct {
		StoreID     int    `json:"store_id"`
		AccessToken string `json:"access_token"`
	}

	if batch.FilePath != "" {
		// If we stored the request as JSON in FilePath, parse it
		err := json.Unmarshal([]byte(batch.FilePath), &syncRequest)
		if err != nil {
			log.Printf("Failed to parse sync request: %v", err)
			s.batch.UpdateStatus(ctx, batch.ID, "failed", "Failed to parse sync request")
			return
		}
	} else {
		log.Printf("No sync request data found for batch %d", batch.ID)
		s.batch.UpdateStatus(ctx, batch.ID, "failed", "No sync request data found")
		return
	}

	// Perform the historical sync
	err = s.svc.SyncHistoricalAdsPerformance(ctx, syncRequest.StoreID, syncRequest.AccessToken)
	if err != nil {
		log.Printf("Failed to sync historical ads performance: %v", err)
		s.batch.UpdateStatus(ctx, batch.ID, "failed", err.Error())
		return
	}

	// Update batch status to completed
	err = s.batch.UpdateStatus(ctx, batch.ID, "completed", "Ads performance sync completed successfully")
	if err != nil {
		log.Printf("Failed to update batch completion status: %v", err)
	}

	log.Printf("Completed ads performance sync batch %d", batch.ID)
}

// CreateSyncBatch creates a new batch for historical ads performance sync
func (s *AdsPerformanceBatchScheduler) CreateSyncBatch(ctx context.Context, storeID int, accessToken string) (int64, error) {
	syncRequest := struct {
		StoreID     int    `json:"store_id"`
		AccessToken string `json:"access_token"`
	}{
		StoreID:     storeID,
		AccessToken: accessToken,
	}

	requestJSON, err := json.Marshal(syncRequest)
	if err != nil {
		return 0, err
	}

	batch := &models.BatchHistory{
		ProcessType: "ads_performance_sync",
		TotalData:   1, // We don't know total until we start processing
		DoneData:    0,
		Status:      "pending",
		FilePath:    string(requestJSON), // Store request as JSON in FilePath
	}

	return s.batch.Create(ctx, batch)
}