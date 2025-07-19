package service

import (
	"context"
	"sync"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/logutil"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// ReconcileBatchCreationScheduler processes pending reconcile batch creation requests in the background.
type ReconcileBatchCreationScheduler struct {
	batch    *BatchService
	svc      *ReconcileService
	interval time.Duration
	logger   *logutil.Logger
}

// NewReconcileBatchCreationScheduler creates a scheduler with the given interval.
func NewReconcileBatchCreationScheduler(batch *BatchService, svc *ReconcileService, interval time.Duration) *ReconcileBatchCreationScheduler {
	if interval <= 0 {
		interval = time.Minute
	}
	return &ReconcileBatchCreationScheduler{
		batch:    batch,
		svc:      svc,
		interval: interval,
		logger:   logutil.NewLogger("reconcile-batch-creation-scheduler", logutil.INFO),
	}
}

// Start launches the scheduler loop.
func (s *ReconcileBatchCreationScheduler) Start(ctx context.Context) {
	if s == nil {
		return
	}

	ctx = logutil.WithNewCorrelationID(ctx)
	s.logger.Info(ctx, "Start", "Starting reconcile batch creation scheduler", map[string]interface{}{
		"interval": s.interval,
	})

	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				s.logger.Info(ctx, "Start", "Reconcile batch creation scheduler stopped due to context cancellation")
				return
			case <-ticker.C:
				s.run(ctx)
			}
		}
	}()
}

func (s *ReconcileBatchCreationScheduler) run(ctx context.Context) {
	runCtx := logutil.WithNewCorrelationID(ctx)
	timer := s.logger.WithTimer(runCtx, "ProcessBatchCreationRequests")

	list, err := s.batch.ListPendingByType(runCtx, "reconcile_batch_creation")
	if err != nil {
		timer.FinishWithError("Failed to list pending batch creation requests", err)
		s.logger.Error(runCtx, "ProcessBatchCreationRequests", "Failed to list pending batch creation requests", err)
		return
	}

	if len(list) == 0 {
		s.logger.Debug(runCtx, "ProcessBatchCreationRequests", "No pending batch creation requests found")
		return
	}

	s.logger.Info(runCtx, "ProcessBatchCreationRequests", "Found pending batch creation requests", map[string]interface{}{
		"batch_count": len(list),
	})

	limit := s.svc.maxThreads
	if limit <= 0 {
		limit = 5
	}
	sem := make(chan struct{}, limit)
	var wg sync.WaitGroup

	for _, b := range list {
		batch := b
		wg.Add(1)
		sem <- struct{}{}
		go func(bb models.BatchHistory) {
			defer func() { <-sem; wg.Done() }()

			batchCtx := logutil.WithNewCorrelationID(runCtx)

			s.logger.Info(batchCtx, "ProcessSingleBatchCreation", "Processing batch creation request", map[string]interface{}{
				"batch_id":   bb.ID,
				"batch_type": bb.ProcessType,
			})

			s.svc.ProcessReconcileBatchCreation(batchCtx, bb.ID)
			s.logger.Info(batchCtx, "ProcessSingleBatchCreation", "Batch creation request processed successfully", map[string]interface{}{
				"batch_id": bb.ID,
			})
		}(batch)
	}
	wg.Wait()

	timer.Finish("All batch creation requests processed")
	s.logger.Info(runCtx, "ProcessBatchCreationRequests", "Completed processing all batch creation requests", map[string]interface{}{
		"processed_count": len(list),
	})
}
