package service

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// ReconcileBatchScheduler processes pending reconcile batches in the background.
type ReconcileBatchScheduler struct {
	batch    *BatchService
	svc      *ReconcileService
	interval time.Duration
}

// NewReconcileBatchScheduler creates a scheduler with the given interval.
func NewReconcileBatchScheduler(batch *BatchService, svc *ReconcileService, interval time.Duration) *ReconcileBatchScheduler {
	if interval <= 0 {
		interval = time.Minute
	}
	return &ReconcileBatchScheduler{batch: batch, svc: svc, interval: interval}
}

// Start launches the scheduler loop.
func (s *ReconcileBatchScheduler) Start(ctx context.Context) {
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

func (s *ReconcileBatchScheduler) run(ctx context.Context) {
	list, err := s.batch.ListPendingByType(ctx, "reconcile_batch")
	if err != nil {
		log.Printf("scheduler list pending: %v", err)
		return
	}
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
			s.svc.ProcessReconcileBatch(ctx, bb.ID)
		}(batch)
	}
	wg.Wait()
}
