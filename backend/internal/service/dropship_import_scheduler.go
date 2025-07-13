package service

import (
	"context"
	"log"
	"time"
)

// DropshipImportScheduler periodically processes pending dropship import batches.
type DropshipImportScheduler struct {
	batch    *BatchService
	svc      *DropshipService
	interval time.Duration
}

// NewDropshipImportScheduler creates a scheduler with the given interval.
// If interval is zero a default of one minute is used.
func NewDropshipImportScheduler(batch *BatchService, svc *DropshipService, interval time.Duration) *DropshipImportScheduler {
	if interval <= 0 {
		interval = time.Minute
	}
	return &DropshipImportScheduler{batch: batch, svc: svc, interval: interval}
}

// Start launches the scheduler loop.
func (s *DropshipImportScheduler) Start(ctx context.Context) {
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

func (s *DropshipImportScheduler) run(ctx context.Context) {
	list, err := s.batch.ListPendingByType(ctx, "dropship_import")
	if err != nil {
		log.Printf("scheduler list pending: %v", err)
		return
	}
	for _, b := range list {
		s.svc.ProcessImportFile(ctx, b.ID, b.FilePath, "")
	}
}
