package service

import (
	"context"
	"sync"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/logutil"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// ReconcileBatchScheduler processes pending reconcile batches in the background.
type ReconcileBatchScheduler struct {
	batch    *BatchService
	svc      *ReconcileService
	interval time.Duration
	logger   *logutil.Logger
}

// NewReconcileBatchScheduler creates a scheduler with the given interval.
func NewReconcileBatchScheduler(batch *BatchService, svc *ReconcileService, interval time.Duration) *ReconcileBatchScheduler {
	if interval <= 0 {
		interval = time.Minute
	}
	return &ReconcileBatchScheduler{
		batch:    batch,
		svc:      svc,
		interval: interval,
		logger:   logutil.NewLogger("reconcile-batch-scheduler", logutil.INFO),
	}
}

// Start launches the scheduler loop.
func (s *ReconcileBatchScheduler) Start(ctx context.Context) {
	if s == nil {
		return
	}

	ctx = logutil.WithNewCorrelationID(ctx)
	s.logger.Info(ctx, "Start", "Starting reconcile batch scheduler", map[string]interface{}{
		"interval": s.interval,
	})

	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				s.logger.Info(ctx, "Start", "Reconcile batch scheduler stopped due to context cancellation")
				return
			case <-ticker.C:
				s.run(ctx)
			}
		}
	}()
}

func (s *ReconcileBatchScheduler) run(ctx context.Context) {
	runCtx := logutil.WithNewCorrelationID(ctx)
	timer := s.logger.WithTimer(runCtx, "ProcessBatches")

	list, err := s.batch.ListPendingByType(runCtx, "reconcile_batch")
	if err != nil {
		timer.FinishWithError("Failed to list pending batches", err)
		s.logger.Error(runCtx, "ProcessBatches", "Failed to list pending batches", err)
		return
	}

	if len(list) == 0 {
		s.logger.Debug(runCtx, "ProcessBatches", "No pending batches found")
		return
	}

	s.logger.Info(runCtx, "ProcessBatches", "Found pending batches", map[string]interface{}{
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

			s.logger.Info(batchCtx, "ProcessSingleBatch", "Processing batch", map[string]interface{}{
				"batch_id":   bb.ID,
				"batch_type": bb.ProcessType,
				"total_data": bb.TotalData,
				"done_data":  bb.DoneData,
			})

			s.svc.ProcessReconcileBatch(batchCtx, bb.ID)
			s.logger.Info(batchCtx, "ProcessSingleBatch", "Batch processed successfully", map[string]interface{}{
				"batch_id": bb.ID,
			})
		}(batch)
	}
	wg.Wait()

	timer.Finish("All batches processed")
	s.logger.Info(runCtx, "ProcessBatches", "Completed processing all batches", map[string]interface{}{
		"processed_count": len(list),
	})
}
