package service

import (
	"context"
	"log"
	"time"
)

// ShopeeDetailBackgroundScheduler processes background Shopee detail fetch jobs
type ShopeeDetailBackgroundScheduler struct {
	backgroundSvc *ShopeeDetailBackgroundService
	interval      time.Duration
	stopChan      chan struct{}
}

// NewShopeeDetailBackgroundScheduler creates a new scheduler
func NewShopeeDetailBackgroundScheduler(
	backgroundSvc *ShopeeDetailBackgroundService,
	interval time.Duration,
) *ShopeeDetailBackgroundScheduler {
	return &ShopeeDetailBackgroundScheduler{
		backgroundSvc: backgroundSvc,
		interval:      interval,
		stopChan:      make(chan struct{}),
	}
}

// Start starts the background scheduler
func (s *ShopeeDetailBackgroundScheduler) Start(ctx context.Context) {
	log.Printf("Starting Shopee detail background scheduler with interval: %v", s.interval)

	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Printf("Shopee detail background scheduler stopped due to context cancellation")
				return
			case <-s.stopChan:
				log.Printf("Shopee detail background scheduler stopped")
				return
			case <-ticker.C:
				s.processPending(ctx)
			}
		}
	}()
}

// Stop stops the background scheduler
func (s *ShopeeDetailBackgroundScheduler) Stop() {
	close(s.stopChan)
}

// processPending processes pending jobs
func (s *ShopeeDetailBackgroundScheduler) processPending(ctx context.Context) {
	if err := s.backgroundSvc.ProcessPendingOrderDetailFetches(ctx); err != nil {
		log.Printf("Error processing pending Shopee detail fetches: %v", err)
	}
}
