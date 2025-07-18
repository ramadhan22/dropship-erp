package service

import (
	"context"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// EnhancedImportScheduler handles concurrent processing of multiple import files
type EnhancedImportScheduler struct {
	batch              *BatchService
	dropshipService    *DropshipService
	streamingProcessor *StreamingImportProcessor
	interval           time.Duration
	maxConcurrentFiles int
	mu                 sync.RWMutex
	activeJobs         map[int64]*ImportJob
	jobQueue           chan *ImportJob
	workers            int
	ctx                context.Context
	cancel             context.CancelFunc
}

// ImportJob represents a single import job
type ImportJob struct {
	BatchID       int64
	FilePath      string
	Channel       string
	Priority      int
	SubmittedAt   time.Time
	StartedAt     time.Time
	CompletedAt   time.Time
	Status        string
	Error         error
	RowsProcessed int
}

// ImportJobStatus represents the status of an import job
type ImportJobStatus struct {
	BatchID         int64     `json:"batch_id"`
	FileName        string    `json:"file_name"`
	Status          string    `json:"status"`
	Progress        float64   `json:"progress"`
	RowsProcessed   int       `json:"rows_processed"`
	TotalRows       int       `json:"total_rows"`
	StartedAt       time.Time `json:"started_at"`
	EstimatedETA    string    `json:"estimated_eta"`
	Error           string    `json:"error,omitempty"`
}

// NewEnhancedImportScheduler creates a new enhanced scheduler
func NewEnhancedImportScheduler(
	batch *BatchService,
	dropshipService *DropshipService,
	streamingProcessor *StreamingImportProcessor,
	interval time.Duration,
	maxConcurrentFiles int,
) *EnhancedImportScheduler {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	if maxConcurrentFiles <= 0 {
		maxConcurrentFiles = 3
	}

	ctx, cancel := context.WithCancel(context.Background())
	
	return &EnhancedImportScheduler{
		batch:              batch,
		dropshipService:    dropshipService,
		streamingProcessor: streamingProcessor,
		interval:           interval,
		maxConcurrentFiles: maxConcurrentFiles,
		activeJobs:         make(map[int64]*ImportJob),
		jobQueue:           make(chan *ImportJob, 100), // Buffer for 100 jobs
		workers:            maxConcurrentFiles,
		ctx:                ctx,
		cancel:             cancel,
	}
}

// Start launches the enhanced scheduler
func (s *EnhancedImportScheduler) Start() {
	if s == nil {
		return
	}

	log.Printf("Starting enhanced import scheduler with %d workers", s.workers)
	
	// Start worker goroutines
	for i := 0; i < s.workers; i++ {
		go s.worker(i)
	}

	// Start job discovery goroutine
	go s.jobDiscovery()

	// Start cleanup goroutine
	go s.cleanup()
}

// Stop stops the scheduler
func (s *EnhancedImportScheduler) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
}

// worker processes import jobs from the queue
func (s *EnhancedImportScheduler) worker(workerID int) {
	log.Printf("Worker %d started", workerID)
	
	for {
		select {
		case <-s.ctx.Done():
			log.Printf("Worker %d stopped", workerID)
			return
		case job := <-s.jobQueue:
			s.processJob(workerID, job)
		}
	}
}

// jobDiscovery discovers and queues pending import jobs
func (s *EnhancedImportScheduler) jobDiscovery() {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.discoverPendingJobs()
		}
	}
}

// discoverPendingJobs finds pending import jobs and adds them to the queue
func (s *EnhancedImportScheduler) discoverPendingJobs() {
	// Get pending dropship imports
	pendingDropship, err := s.batch.ListPendingByType(s.ctx, "dropship_import")
	if err != nil {
		log.Printf("Error getting pending dropship imports: %v", err)
		return
	}

	// Get pending streaming imports
	pendingStreaming, err := s.batch.ListPendingByType(s.ctx, "streaming_dropship_import")
	if err != nil {
		log.Printf("Error getting pending streaming imports: %v", err)
		return
	}

	// Combine and prioritize jobs
	allPending := append(pendingDropship, pendingStreaming...)
	jobs := s.prioritizeJobs(allPending)

	// Queue jobs that aren't already active
	for _, job := range jobs {
		s.mu.RLock()
		_, isActive := s.activeJobs[job.BatchID]
		s.mu.RUnlock()

		if !isActive {
			select {
			case s.jobQueue <- job:
				log.Printf("Queued job %d: %s", job.BatchID, job.FilePath)
			default:
				log.Printf("Job queue full, skipping job %d", job.BatchID)
			}
		}
	}
}

// prioritizeJobs sorts jobs by priority (smaller files first, then by submission time)
func (s *EnhancedImportScheduler) prioritizeJobs(batches []models.BatchHistory) []*ImportJob {
	jobs := make([]*ImportJob, 0, len(batches))

	for _, batch := range batches {
		job := &ImportJob{
			BatchID:     batch.ID,
			FilePath:    batch.FilePath,
			Channel:     "",
			Priority:    s.calculatePriority(batch),
			SubmittedAt: batch.CreatedAt,
			Status:      "pending",
		}
		jobs = append(jobs, job)
	}

	// Sort by priority (ascending - lower number = higher priority)
	sort.Slice(jobs, func(i, j int) bool {
		if jobs[i].Priority == jobs[j].Priority {
			return jobs[i].SubmittedAt.Before(jobs[j].SubmittedAt)
		}
		return jobs[i].Priority < jobs[j].Priority
	})

	return jobs
}

// calculatePriority calculates job priority based on file size and age
func (s *EnhancedImportScheduler) calculatePriority(batch models.BatchHistory) int {
	// Base priority on file age (older files get higher priority)
	age := time.Since(batch.CreatedAt)
	agePriority := int(age.Minutes())

	// Adjust priority based on file size (smaller files get higher priority)
	// This is a simplified approach - in reality, you'd want to check actual file size
	sizePriority := 0
	if batch.TotalData > 0 {
		// Larger files get lower priority (higher number)
		sizePriority = batch.TotalData / 1000
	}

	return agePriority + sizePriority
}

// processJob processes a single import job
func (s *EnhancedImportScheduler) processJob(workerID int, job *ImportJob) {
	log.Printf("Worker %d processing job %d: %s", workerID, job.BatchID, job.FilePath)

	// Mark job as active
	s.mu.Lock()
	job.Status = "processing"
	job.StartedAt = time.Now()
	s.activeJobs[job.BatchID] = job
	s.mu.Unlock()

	// Update batch status
	if err := s.batch.UpdateStatus(s.ctx, job.BatchID, "processing", ""); err != nil {
		log.Printf("Error updating batch status: %v", err)
	}

	// Process the file
	var err error
	if s.streamingProcessor != nil {
		// Use streaming processor for better performance
		err = s.streamingProcessor.processFileWithStreaming(s.ctx, job.FilePath, job.Channel)
	} else {
		// Fallback to original processor
		s.dropshipService.ProcessImportFile(s.ctx, job.BatchID, job.FilePath, job.Channel)
	}

	// Update job status
	s.mu.Lock()
	job.CompletedAt = time.Now()
	if err != nil {
		job.Status = "failed"
		job.Error = err
		s.batch.UpdateStatus(s.ctx, job.BatchID, "failed", err.Error())
	} else {
		job.Status = "completed"
		s.batch.UpdateStatus(s.ctx, job.BatchID, "completed", "")
	}
	s.mu.Unlock()

	duration := job.CompletedAt.Sub(job.StartedAt)
	log.Printf("Worker %d completed job %d in %v (status: %s)", workerID, job.BatchID, duration, job.Status)
}

// cleanup removes completed jobs from active jobs map
func (s *EnhancedImportScheduler) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.cleanupCompletedJobs()
		}
	}
}

// cleanupCompletedJobs removes old completed jobs
func (s *EnhancedImportScheduler) cleanupCompletedJobs() {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().Add(-30 * time.Minute)
	
	for batchID, job := range s.activeJobs {
		if (job.Status == "completed" || job.Status == "failed") && job.CompletedAt.Before(cutoff) {
			delete(s.activeJobs, batchID)
		}
	}
}

// GetActiveJobs returns currently active import jobs
func (s *EnhancedImportScheduler) GetActiveJobs() []ImportJobStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	jobs := make([]ImportJobStatus, 0, len(s.activeJobs))
	
	for _, job := range s.activeJobs {
		status := ImportJobStatus{
			BatchID:       job.BatchID,
			FileName:      job.FilePath,
			Status:        job.Status,
			RowsProcessed: job.RowsProcessed,
			StartedAt:     job.StartedAt,
		}

		if job.Error != nil {
			status.Error = job.Error.Error()
		}

		// Calculate progress and ETA if available
		if job.Status == "processing" && s.streamingProcessor != nil {
			stats := s.streamingProcessor.GetStats()
			if stats.TotalRows > 0 {
				status.Progress = float64(stats.ProcessedRows) / float64(stats.TotalRows) * 100
				status.TotalRows = stats.TotalRows
				status.RowsProcessed = stats.ProcessedRows
				
				if eta := s.streamingProcessor.EstimateRemainingTime(); eta > 0 {
					status.EstimatedETA = eta.String()
				}
			}
		}

		jobs = append(jobs, status)
	}

	return jobs
}

// GetQueueStatus returns information about the job queue
func (s *EnhancedImportScheduler) GetQueueStatus() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"queue_length":     len(s.jobQueue),
		"active_jobs":      len(s.activeJobs),
		"max_workers":      s.workers,
		"queue_capacity":   cap(s.jobQueue),
	}
}

// ForceProcessBatch forces processing of a specific batch
func (s *EnhancedImportScheduler) ForceProcessBatch(batchID int64) error {
	// Get batch details
	batches, err := s.batch.List(s.ctx)
	if err != nil {
		return fmt.Errorf("list batches: %w", err)
	}

	var targetBatch *models.BatchHistory
	for _, batch := range batches {
		if batch.ID == batchID {
			targetBatch = &batch
			break
		}
	}

	if targetBatch == nil {
		return fmt.Errorf("batch %d not found", batchID)
	}

	// Create and queue job
	job := &ImportJob{
		BatchID:     targetBatch.ID,
		FilePath:    targetBatch.FilePath,
		Channel:     "",
		Priority:    0, // Highest priority
		SubmittedAt: time.Now(),
		Status:      "pending",
	}

	select {
	case s.jobQueue <- job:
		log.Printf("Force-queued job %d: %s", job.BatchID, job.FilePath)
		return nil
	default:
		return fmt.Errorf("job queue is full")
	}
}