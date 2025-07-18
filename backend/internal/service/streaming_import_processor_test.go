package service

import (
	"context"
	"testing"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
)

// Mock implementations for testing
type mockDropshipService struct {
	purchases []models.DropshipPurchase
	details   []models.DropshipPurchaseDetail
}

func (m *mockDropshipService) createPendingSalesJournal(ctx context.Context, jr DropshipJournalRepo, purchase *models.DropshipPurchase, prodTotal, pending float64) error {
	return nil
}

type mockDropshipRepo struct {
	purchases map[string]bool
}

func (m *mockDropshipRepo) InsertDropshipPurchase(ctx context.Context, p *models.DropshipPurchase) error {
	m.purchases[p.KodePesanan] = true
	return nil
}

func (m *mockDropshipRepo) InsertDropshipPurchaseDetail(ctx context.Context, d *models.DropshipPurchaseDetail) error {
	return nil
}

func (m *mockDropshipRepo) ExistsDropshipPurchase(ctx context.Context, kodePesanan string) (bool, error) {
	return m.purchases[kodePesanan], nil
}

func (m *mockDropshipRepo) ListExistingPurchases(ctx context.Context, ids []string) (map[string]bool, error) {
	return make(map[string]bool), nil
}

func (m *mockDropshipRepo) ListDropshipPurchases(ctx context.Context, channel, store, from, to, orderNo, sortBy, dir string, limit, offset int) ([]models.DropshipPurchase, int, error) {
	return []models.DropshipPurchase{}, 0, nil
}

func (m *mockDropshipRepo) ListDropshipPurchasesFiltered(ctx context.Context, params *models.FilterParams) (*models.QueryResult, error) {
	return &models.QueryResult{}, nil
}

func (m *mockDropshipRepo) SumDropshipPurchases(ctx context.Context, channel, store, from, to string) (float64, error) {
	return 0, nil
}

func (m *mockDropshipRepo) GetDropshipPurchaseByID(ctx context.Context, kodePesanan string) (*models.DropshipPurchase, error) {
	return &models.DropshipPurchase{}, nil
}

func (m *mockDropshipRepo) ListDropshipPurchaseDetails(ctx context.Context, kodePesanan string) ([]models.DropshipPurchaseDetail, error) {
	return []models.DropshipPurchaseDetail{}, nil
}

func (m *mockDropshipRepo) TopProducts(ctx context.Context, channel, store, from, to string, limit int) ([]models.ProductSales, error) {
	return []models.ProductSales{}, nil
}

func (m *mockDropshipRepo) DailyTotals(ctx context.Context, channel, store, from, to string) ([]repository.DailyPurchaseTotal, error) {
	return []repository.DailyPurchaseTotal{}, nil
}

func (m *mockDropshipRepo) MonthlyTotals(ctx context.Context, channel, store, from, to string) ([]repository.MonthlyPurchaseTotal, error) {
	return []repository.MonthlyPurchaseTotal{}, nil
}

func (m *mockDropshipRepo) CancelledSummary(ctx context.Context, channel, store, from, to string) (repository.CancelledSummary, error) {
	return repository.CancelledSummary{}, nil
}

type mockJournalRepo struct{}

func (m *mockJournalRepo) CreateJournalEntry(ctx context.Context, e *models.JournalEntry) (int64, error) {
	return 1, nil
}

func (m *mockJournalRepo) InsertJournalLine(ctx context.Context, l *models.JournalLine) error {
	return nil
}

func (m *mockJournalRepo) InsertJournalLines(ctx context.Context, lines []models.JournalLine) error {
	return nil
}

func TestStreamingImportProcessor_ChunkReading(t *testing.T) {
	// Create a mock service
	repo := &mockDropshipRepo{purchases: make(map[string]bool)}
	journalRepo := &mockJournalRepo{}
	
	service := &DropshipService{
		repo:        repo,
		journalRepo: journalRepo,
	}

	// Create processor
	config := DefaultStreamingImportConfig()
	config.ChunkSize = 2 // Small chunk size for testing
	processor := NewStreamingImportProcessor(service, config)

	// Test validateHeader
	validHeader := make([]string, 25)
	err := processor.validateHeader(validHeader)
	if err != nil {
		t.Errorf("Expected no error for valid header, got %v", err)
	}

	// Test invalid header
	invalidHeader := make([]string, 10)
	err = processor.validateHeader(invalidHeader)
	if err == nil {
		t.Error("Expected error for invalid header")
	}
}

func TestStreamingImportProcessor_MemoryOptimization(t *testing.T) {
	config := DefaultStreamingImportConfig()
	config.EnableMemoryOptimization = true
	config.ChunkSize = 1000

	processor := NewStreamingImportProcessor(nil, config)

	// Test memory stats
	stats := processor.GetStats()
	if stats.StartTime.IsZero() {
		t.Error("Expected non-zero start time")
	}
}

func TestStreamingImportConfig_Defaults(t *testing.T) {
	config := DefaultStreamingImportConfig()

	expectedChunkSize := DefaultChunkSize
	if config.ChunkSize != expectedChunkSize {
		t.Errorf("Expected chunk size %d, got %d", expectedChunkSize, config.ChunkSize)
	}

	expectedMaxFiles := DefaultMaxConcurrentFiles
	if config.MaxConcurrentFiles != expectedMaxFiles {
		t.Errorf("Expected max files %d, got %d", expectedMaxFiles, config.MaxConcurrentFiles)
	}

	expectedMaxFileSize := int64(DefaultMaxFileSize)
	if config.MaxFileSize != expectedMaxFileSize {
		t.Errorf("Expected max file size %d, got %d", expectedMaxFileSize, config.MaxFileSize)
	}
}

func TestStreamingImportProcessor_EstimateRemainingTime(t *testing.T) {
	processor := NewStreamingImportProcessor(nil, DefaultStreamingImportConfig())

	// Initially should return 0
	eta := processor.EstimateRemainingTime()
	if eta != 0 {
		t.Errorf("Expected 0 ETA initially, got %v", eta)
	}

	// Simulate some progress
	processor.mu.Lock()
	processor.stats.TotalRows = 1000
	processor.stats.ProcessedRows = 500
	processor.stats.StartTime = time.Now().Add(-1 * time.Minute)
	processor.mu.Unlock()

	eta = processor.EstimateRemainingTime()
	if eta <= 0 {
		t.Errorf("Expected positive ETA with progress, got %v", eta)
	}
}

func TestStreamingImportProcessor_ValidateHeader(t *testing.T) {
	processor := NewStreamingImportProcessor(nil, DefaultStreamingImportConfig())

	// Test valid header
	validHeader := make([]string, 25)
	err := processor.validateHeader(validHeader)
	if err != nil {
		t.Errorf("Expected no error for valid header, got %v", err)
	}

	// Test invalid header
	invalidHeader := make([]string, 10)
	err = processor.validateHeader(invalidHeader)
	if err == nil {
		t.Error("Expected error for invalid header")
	}
}