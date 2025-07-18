package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
)

const (
	DefaultChunkSize         = 1000  // Process 1000 rows at a time
	DefaultMaxConcurrentFiles = 5    // Process 5 files concurrently
	DefaultMaxFileSize       = 100 * 1024 * 1024 // 100MB max file size
	DefaultProgressUpdateInterval = 100 // Update progress every 100 rows
)

// StreamingImportConfig contains configuration for streaming import processing
type StreamingImportConfig struct {
	ChunkSize                int
	MaxConcurrentFiles      int
	MaxFileSize             int64
	ProgressUpdateInterval  int
	EnableMemoryOptimization bool
}

// DefaultStreamingImportConfig returns default configuration for streaming imports
func DefaultStreamingImportConfig() *StreamingImportConfig {
	return &StreamingImportConfig{
		ChunkSize:                DefaultChunkSize,
		MaxConcurrentFiles:      DefaultMaxConcurrentFiles,
		MaxFileSize:             DefaultMaxFileSize,
		ProgressUpdateInterval:  DefaultProgressUpdateInterval,
		EnableMemoryOptimization: true,
	}
}

// StreamingImportProcessor handles large-scale dropship imports with streaming and chunking
type StreamingImportProcessor struct {
	service *DropshipService
	config  *StreamingImportConfig
	mu      sync.RWMutex
	stats   *ImportStats
}

// ImportStats tracks import statistics
type ImportStats struct {
	TotalFiles     int
	ProcessedFiles int
	FailedFiles    int
	TotalRows      int
	ProcessedRows  int
	FailedRows     int
	StartTime      time.Time
	LastUpdateTime time.Time
}

// NewStreamingImportProcessor creates a new streaming import processor
func NewStreamingImportProcessor(service *DropshipService, config *StreamingImportConfig) *StreamingImportProcessor {
	if config == nil {
		config = DefaultStreamingImportConfig()
	}
	return &StreamingImportProcessor{
		service: service,
		config:  config,
		stats:   &ImportStats{StartTime: time.Now()},
	}
}

// ProcessMultipleFiles processes multiple import files concurrently with streaming
func (p *StreamingImportProcessor) ProcessMultipleFiles(ctx context.Context, filePaths []string, channel string) error {
	p.mu.Lock()
	p.stats.TotalFiles = len(filePaths)
	p.stats.StartTime = time.Now()
	p.mu.Unlock()

	// Create a semaphore to limit concurrent files
	sem := make(chan struct{}, p.config.MaxConcurrentFiles)
	var wg sync.WaitGroup
	var globalErr error
	var errMu sync.Mutex

	log.Printf("Starting streaming import of %d files with %d concurrent workers", len(filePaths), p.config.MaxConcurrentFiles)

	for _, filePath := range filePaths {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			sem <- struct{}{} // Acquire semaphore
			defer func() { <-sem }() // Release semaphore

			if err := p.processFileWithStreaming(ctx, path, channel); err != nil {
				errMu.Lock()
				if globalErr == nil {
					globalErr = fmt.Errorf("file %s: %w", path, err)
				}
				p.stats.FailedFiles++
				errMu.Unlock()
				log.Printf("Failed to process file %s: %v", path, err)
			} else {
				p.mu.Lock()
				p.stats.ProcessedFiles++
				p.mu.Unlock()
				log.Printf("Successfully processed file %s", path)
			}
		}(filePath)
	}

	wg.Wait()
	
	p.mu.Lock()
	duration := time.Since(p.stats.StartTime)
	p.mu.Unlock()

	log.Printf("Streaming import completed. Files: %d processed, %d failed. Duration: %v", 
		p.stats.ProcessedFiles, p.stats.FailedFiles, duration)
	
	return globalErr
}

// processFileWithStreaming processes a single file using streaming and chunking
func (p *StreamingImportProcessor) processFileWithStreaming(ctx context.Context, filePath, channel string) error {
	// Validate file size
	if err := p.validateFileSize(filePath); err != nil {
		return err
	}

	// Create batch record
	batchID := int64(0)
	if p.service.batchSvc != nil {
		filename := filepath.Base(filePath)
		batch := &models.BatchHistory{
			ProcessType: "streaming_dropship_import",
			TotalData:   0,
			DoneData:    0,
			Status:      "processing",
			FileName:    filename,
			FilePath:    filePath,
		}
		
		var err error
		batchID, err = p.service.batchSvc.Create(ctx, batch)
		if err != nil {
			return fmt.Errorf("create batch record: %w", err)
		}
	}

	// Process file in chunks
	if err := p.processFileInChunks(ctx, filePath, channel, batchID); err != nil {
		if p.service.batchSvc != nil && batchID != 0 {
			p.service.batchSvc.UpdateStatus(ctx, batchID, "failed", err.Error())
		}
		return err
	}

	if p.service.batchSvc != nil && batchID != 0 {
		p.service.batchSvc.UpdateStatus(ctx, batchID, "completed", "")
	}

	return nil
}

// validateFileSize checks if file size is within acceptable limits
func (p *StreamingImportProcessor) validateFileSize(filePath string) error {
	stat, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("stat file: %w", err)
	}

	if stat.Size() > p.config.MaxFileSize {
		return fmt.Errorf("file size %d bytes exceeds maximum %d bytes", stat.Size(), p.config.MaxFileSize)
	}

	return nil
}

// processFileInChunks processes a CSV file in chunks to optimize memory usage
func (p *StreamingImportProcessor) processFileInChunks(ctx context.Context, filePath, channel string, batchID int64) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.ReuseRecord = true

	// Read and validate header
	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("read header: %w", err)
	}

	if err := p.validateHeader(header); err != nil {
		return fmt.Errorf("validate header: %w", err)
	}

	// Count total rows for progress tracking
	totalRows, err := p.countTotalRows(filePath)
	if err != nil {
		log.Printf("Warning: could not count total rows: %v", err)
	} else if p.service.batchSvc != nil && batchID != 0 {
		p.service.batchSvc.UpdateTotal(ctx, batchID, totalRows)
	}

	// Process file in chunks
	chunkNum := 0
	processedRows := 0
	
	for {
		chunk, err := p.readChunk(reader, p.config.ChunkSize)
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("read chunk %d: %w", chunkNum, err)
		}

		if len(chunk) == 0 {
			break
		}

		chunkNum++
		log.Printf("Processing chunk %d with %d rows", chunkNum, len(chunk))

		// Process chunk with transaction
		rowsProcessed, err := p.processChunk(ctx, chunk, channel, batchID)
		if err != nil {
			log.Printf("Error processing chunk %d: %v", chunkNum, err)
			// Continue processing other chunks rather than failing entirely
		}

		processedRows += rowsProcessed
		
		// Update progress
		p.mu.Lock()
		p.stats.ProcessedRows += rowsProcessed
		p.stats.LastUpdateTime = time.Now()
		p.mu.Unlock()

		if p.service.batchSvc != nil && batchID != 0 {
			p.service.batchSvc.UpdateDone(ctx, batchID, processedRows)
		}

		// Check for cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}

	log.Printf("Completed processing file %s: %d rows in %d chunks", filePath, processedRows, chunkNum)
	return nil
}

// readChunk reads up to chunkSize rows from the CSV reader
func (p *StreamingImportProcessor) readChunk(reader *csv.Reader, chunkSize int) ([][]string, error) {
	var chunk [][]string
	
	for i := 0; i < chunkSize; i++ {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return chunk, err
		}

		// Create a copy of the record to avoid issues with ReuseRecord
		recordCopy := make([]string, len(record))
		copy(recordCopy, record)
		chunk = append(chunk, recordCopy)
	}

	return chunk, nil
}

// processChunk processes a chunk of CSV records
func (p *StreamingImportProcessor) processChunk(ctx context.Context, chunk [][]string, channel string, batchID int64) (int, error) {
	if len(chunk) == 0 {
		return 0, nil
	}

	// Start transaction for this chunk
	tx, err := p.service.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	repoTx := repository.NewDropshipRepo(tx)
	jrTx := repository.NewJournalRepo(tx)

	// Process records in this chunk
	processedCount := 0
	for _, record := range chunk {
		if err := p.processRecord(ctx, repoTx, jrTx, record, channel, batchID); err != nil {
			log.Printf("Error processing record: %v", err)
			// Continue processing other records in the chunk
			continue
		}
		processedCount++
	}

	// Commit transaction for this chunk
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit transaction: %w", err)
	}

	return processedCount, nil
}

// processRecord processes a single CSV record
func (p *StreamingImportProcessor) processRecord(ctx context.Context, repoTx DropshipRepoInterface, jrTx DropshipJournalRepo, record []string, channel string, batchID int64) error {
	// Validate record length
	if len(record) < 20 {
		return fmt.Errorf("invalid record length: %d", len(record))
	}

	// Parse record data
	qty, err := strconv.Atoi(record[8])
	if err != nil {
		return fmt.Errorf("parse qty: %w", err)
	}

	hargaProduk, _ := strconv.ParseFloat(record[7], 64)
	totalHargaProduk, _ := strconv.ParseFloat(record[9], 64)
	biayaLain, _ := strconv.ParseFloat(record[10], 64)
	biayaMitra, _ := strconv.ParseFloat(record[11], 64)
	totalTransaksi, _ := strconv.ParseFloat(record[12], 64)
	hargaChannel, _ := strconv.ParseFloat(record[13], 64)
	totalHargaChannel, _ := strconv.ParseFloat(record[14], 64)
	potensi, _ := strconv.ParseFloat(record[15], 64)

	waktuPesanan, err := time.Parse("02 January 2006, 15:04:05", record[1])
	if err != nil {
		return fmt.Errorf("parse waktu_pesanan: %w", err)
	}

	// Create purchase record
	purchase := &models.DropshipPurchase{
		KodePesanan:           record[3],
		WaktuPesananTerbuat:   waktuPesanan,
		JenisChannel:          record[17],
		NamaToko:              record[18],
		KodeInvoiceChannel:    strings.TrimPrefix(record[19], "'"),
		BiayaLainnya:          biayaLain,
		BiayaMitraJakmall:     biayaMitra,
		TotalTransaksi:        totalTransaksi,
		StatusPesananTerakhir: record[16],
	}

	// Filter by channel if specified
	if channel != "" && purchase.JenisChannel != channel {
		return nil // Skip this record
	}

	// Check if purchase already exists
	exists, err := repoTx.ExistsDropshipPurchase(ctx, purchase.KodePesanan)
	if err != nil {
		return fmt.Errorf("check purchase exists: %w", err)
	}
	if exists {
		return nil // Skip existing purchase
	}

	// Insert purchase
	if err := repoTx.InsertDropshipPurchase(ctx, purchase); err != nil {
		return fmt.Errorf("insert purchase: %w", err)
	}

	// Insert purchase detail
	detail := &models.DropshipPurchaseDetail{
		KodePesanan:             purchase.KodePesanan,
		SKU:                     record[5],
		NamaProduk:              record[6],
		HargaProduk:             hargaProduk,
		Qty:                     qty,
		TotalHargaProduk:        totalHargaProduk,
		HargaProdukChannel:      hargaChannel,
		TotalHargaProdukChannel: totalHargaChannel,
		PotensiKeuntungan:       potensi,
	}

	if err := repoTx.InsertDropshipPurchaseDetail(ctx, detail); err != nil {
		return fmt.Errorf("insert purchase detail: %w", err)
	}

	// Create journal entry if needed
	if !strings.EqualFold(purchase.NamaToko, "MR eStore Free Sample") {
		pending := totalHargaChannel
		if err := p.service.createPendingSalesJournal(ctx, jrTx, purchase, totalHargaProduk, pending); err != nil {
			return fmt.Errorf("create journal entry: %w", err)
		}
	}

	return nil
}

// countTotalRows counts the total number of data rows in a CSV file
func (p *StreamingImportProcessor) countTotalRows(filePath string) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	
	// Skip header
	if _, err := reader.Read(); err != nil {
		return 0, err
	}

	count := 0
	for {
		_, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return 0, err
		}
		count++
	}

	return count, nil
}

// validateHeader validates the CSV header format
func (p *StreamingImportProcessor) validateHeader(header []string) error {
	if len(header) < 20 {
		return fmt.Errorf("header too short: expected at least 20 columns, got %d", len(header))
	}
	
	// Could add more specific header validation here
	return nil
}

// GetStats returns current import statistics
func (p *StreamingImportProcessor) GetStats() ImportStats {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return *p.stats
}

// EstimateRemainingTime estimates remaining time based on current progress
func (p *StreamingImportProcessor) EstimateRemainingTime() time.Duration {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.stats.ProcessedRows == 0 {
		return 0
	}

	elapsed := time.Since(p.stats.StartTime)
	rowsPerSecond := float64(p.stats.ProcessedRows) / elapsed.Seconds()
	
	if rowsPerSecond == 0 {
		return 0
	}

	remainingRows := p.stats.TotalRows - p.stats.ProcessedRows
	return time.Duration(float64(remainingRows)/rowsPerSecond) * time.Second
}