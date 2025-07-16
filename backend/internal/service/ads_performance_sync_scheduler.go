package service

import (
	"context"
	"log"
	"time"
)

// AdsPerformanceSyncScheduler periodically processes pending ads performance sync jobs.
type AdsPerformanceSyncScheduler struct {
	syncSvc  *AdsPerformanceSyncService
	interval time.Duration
}

// NewAdsPerformanceSyncScheduler creates a scheduler with the given interval.
// If interval is zero, a default of 2 minutes is used.
func NewAdsPerformanceSyncScheduler(syncSvc *AdsPerformanceSyncService, interval time.Duration) *AdsPerformanceSyncScheduler {
	if interval <= 0 {
		interval = 2 * time.Minute // Longer interval for ads sync due to API rate limits
	}
	return &AdsPerformanceSyncScheduler{
		syncSvc:  syncSvc,
		interval: interval,
	}
}

// Start launches the scheduler loop.
func (s *AdsPerformanceSyncScheduler) Start(ctx context.Context) {
	if s == nil || s.syncSvc == nil {
		return
	}
	
	log.Printf("Starting ads performance sync scheduler with interval %v", s.interval)
	
	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()
		
		for {
			select {
			case <-ctx.Done():
				log.Printf("Ads performance sync scheduler stopped")
				return
			case <-ticker.C:
				s.run(ctx)
			}
		}
	}()
}

func (s *AdsPerformanceSyncScheduler) run(ctx context.Context) {
	jobs, err := s.syncSvc.ListPendingSyncJobs()
	if err != nil {
		log.Printf("ads sync scheduler list pending: %v", err)
		return
	}
	
	if len(jobs) == 0 {
		return // No pending jobs
	}
	
	log.Printf("Found %d pending ads sync jobs", len(jobs))
	
	for _, job := range jobs {
		// Process each job in a separate goroutine to avoid blocking
		go func(jobID int64) {
			s.syncSvc.ProcessSyncJob(ctx, jobID)
		}(job.ID)
		
		// Add a small delay between starting jobs to avoid overwhelming the API
		time.Sleep(10 * time.Second)
	}
}