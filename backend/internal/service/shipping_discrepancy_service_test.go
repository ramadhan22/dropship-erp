package service

import (
	"context"
	"testing"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// mockShippingDiscrepancyRepo is a mock repository for testing shipping discrepancy service
type mockShippingDiscrepancyRepo struct {
	discrepancies []models.ShippingDiscrepancy
	insertCount   int
	statsResult   map[string]int
	sumsResult    map[string]float64
}

func (m *mockShippingDiscrepancyRepo) InsertShippingDiscrepancy(ctx context.Context, discrepancy *models.ShippingDiscrepancy) error {
	m.insertCount++
	m.discrepancies = append(m.discrepancies, *discrepancy)
	return nil
}

func (m *mockShippingDiscrepancyRepo) GetShippingDiscrepanciesByStore(ctx context.Context, storeName string, limit, offset int) ([]models.ShippingDiscrepancy, error) {
	return m.discrepancies, nil
}

func (m *mockShippingDiscrepancyRepo) GetAllShippingDiscrepancies(ctx context.Context, limit, offset int) ([]models.ShippingDiscrepancy, error) {
	return m.discrepancies, nil
}

func (m *mockShippingDiscrepancyRepo) GetShippingDiscrepanciesByType(ctx context.Context, discrepancyType string, limit, offset int) ([]models.ShippingDiscrepancy, error) {
	return m.discrepancies, nil
}

func (m *mockShippingDiscrepancyRepo) CountShippingDiscrepanciesByDateRange(ctx context.Context, startDate, endDate time.Time) (map[string]int, error) {
	return m.statsResult, nil
}

func (m *mockShippingDiscrepancyRepo) GetShippingDiscrepancySumsByDateRange(ctx context.Context, startDate, endDate time.Time) (map[string]float64, error) {
	return m.sumsResult, nil
}

func (m *mockShippingDiscrepancyRepo) GetShippingDiscrepancyByInvoice(ctx context.Context, invoiceNumber string) (*models.ShippingDiscrepancy, error) {
	for _, disc := range m.discrepancies {
		if disc.InvoiceNumber == invoiceNumber {
			return &disc, nil
		}
	}
	return nil, nil
}

func TestShippingDiscrepancyService_GetShippingDiscrepancySums(t *testing.T) {
	mockRepo := &mockShippingDiscrepancyRepo{
		sumsResult: map[string]float64{
			"selisih_ongkir":       50000.0,
			"reverse_shipping_fee": 25000.0,
		},
		statsResult: map[string]int{
			"selisih_ongkir":       10,
			"reverse_shipping_fee": 5,
		},
	}

	service := &ShippingDiscrepancyService{
		discRepo: mockRepo,
	}

	startDate := time.Now().AddDate(0, -1, 0)
	endDate := time.Now()

	// Test GetShippingDiscrepancySums
	sums, err := service.GetShippingDiscrepancySums(context.Background(), startDate, endDate)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if sums["selisih_ongkir"] != 50000.0 {
		t.Errorf("Expected selisih_ongkir sum to be 50000.0, got %f", sums["selisih_ongkir"])
	}

	if sums["reverse_shipping_fee"] != 25000.0 {
		t.Errorf("Expected reverse_shipping_fee sum to be 25000.0, got %f", sums["reverse_shipping_fee"])
	}

	// Test GetShippingDiscrepancyStats still works
	stats, err := service.GetShippingDiscrepancyStats(context.Background(), startDate, endDate)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if stats["selisih_ongkir"] != 10 {
		t.Errorf("Expected selisih_ongkir count to be 10, got %d", stats["selisih_ongkir"])
	}

	if stats["reverse_shipping_fee"] != 5 {
		t.Errorf("Expected reverse_shipping_fee count to be 5, got %d", stats["reverse_shipping_fee"])
	}
}

func TestShippingDiscrepancyService_ReturnIDField(t *testing.T) {
	mockRepo := &mockShippingDiscrepancyRepo{}
	service := &ShippingDiscrepancyService{
		discRepo: mockRepo,
	}

	// Test that we can retrieve discrepancies and they have return_id field
	returnID := "RET123"
	discrepancy := models.ShippingDiscrepancy{
		InvoiceNumber:     "INV123",
		ReturnID:          &returnID,
		DiscrepancyType:   "selisih_ongkir",
		DiscrepancyAmount: 10000.0,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	mockRepo.discrepancies = append(mockRepo.discrepancies, discrepancy)

	results, err := service.GetShippingDiscrepancies(context.Background(), "", "", 10, 0)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 discrepancy, got %d", len(results))
	}

	if results[0].ReturnID == nil || *results[0].ReturnID != "RET123" {
		t.Errorf("Expected return_id to be 'RET123', got %v", results[0].ReturnID)
	}
}
