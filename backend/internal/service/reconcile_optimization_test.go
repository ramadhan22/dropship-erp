package service

import (
	"context"
	"testing"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// MockDropshipRepoOptimized is a mock that tracks whether bulk methods are called
type MockDropshipRepoOptimized struct {
	purchases   map[string]*models.DropshipPurchase
	bulkCalled  bool
	singleCalls int
}

func (m *MockDropshipRepoOptimized) GetDropshipPurchaseByInvoice(ctx context.Context, kodeInvoice string) (*models.DropshipPurchase, error) {
	m.singleCalls++
	return m.purchases[kodeInvoice], nil
}

func (m *MockDropshipRepoOptimized) GetDropshipPurchaseByID(ctx context.Context, kodePesanan string) (*models.DropshipPurchase, error) {
	for _, p := range m.purchases {
		if p.KodePesanan == kodePesanan {
			return p, nil
		}
	}
	return nil, nil
}

func (m *MockDropshipRepoOptimized) GetDropshipPurchasesByInvoices(ctx context.Context, invoices []string) ([]*models.DropshipPurchase, error) {
	m.bulkCalled = true
	result := make([]*models.DropshipPurchase, 0, len(invoices))
	for _, inv := range invoices {
		if p, exists := m.purchases[inv]; exists {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *MockDropshipRepoOptimized) UpdatePurchaseStatus(ctx context.Context, kodePesanan, status string) error {
	return nil
}

func (m *MockDropshipRepoOptimized) SumDetailByInvoice(ctx context.Context, kodeInvoice string) (float64, error) {
	return 100.0, nil
}

func (m *MockDropshipRepoOptimized) SumProductCostByInvoice(ctx context.Context, kodeInvoice string) (float64, error) {
	return 80.0, nil
}

// Test that the bulk optimization is working
func TestBulkOptimization(t *testing.T) {
	mockRepo := &MockDropshipRepoOptimized{
		purchases: map[string]*models.DropshipPurchase{
			"INV001": {KodePesanan: "P001", KodeInvoiceChannel: "INV001", NamaToko: "Store1"},
			"INV002": {KodePesanan: "P002", KodeInvoiceChannel: "INV002", NamaToko: "Store1"},
			"INV003": {KodePesanan: "P003", KodeInvoiceChannel: "INV003", NamaToko: "Store2"},
		},
	}

	svc := &ReconcileService{
		dropRepo: mockRepo,
	}

	invoices := []string{"INV001", "INV002", "INV003"}

	// Test the bulk fetch method
	purchases, err := svc.bulkGetDropshipPurchasesByInvoices(context.Background(), invoices)
	if err != nil {
		t.Fatalf("bulkGetDropshipPurchasesByInvoices failed: %v", err)
	}

	// Verify bulk method was called instead of individual calls
	if !mockRepo.bulkCalled {
		t.Error("Expected bulk method to be called")
	}

	if mockRepo.singleCalls > 0 {
		t.Errorf("Expected no single calls, but got %d", mockRepo.singleCalls)
	}

	// Verify we got the right data
	if len(purchases) != 3 {
		t.Errorf("Expected 3 purchases, got %d", len(purchases))
	}
}

// Test fallback when bulk method not available
func TestFallbackToSingleCalls(t *testing.T) {
	// Use a mock that doesn't implement the bulk method
	mockRepo := &BasicMockDropshipRepo{
		purchases: map[string]*models.DropshipPurchase{
			"INV001": {KodePesanan: "P001", KodeInvoiceChannel: "INV001", NamaToko: "Store1"},
		},
	}

	svc := &ReconcileService{
		dropRepo: mockRepo,
	}

	invoices := []string{"INV001"}

	// Should fallback to individual calls
	purchases, err := svc.bulkGetDropshipPurchasesByInvoices(context.Background(), invoices)
	if err != nil {
		t.Fatalf("bulkGetDropshipPurchasesByInvoices failed: %v", err)
	}

	if len(purchases) != 1 {
		t.Errorf("Expected 1 purchase, got %d", len(purchases))
	}
}

type BasicMockDropshipRepo struct {
	purchases map[string]*models.DropshipPurchase
}

func (m *BasicMockDropshipRepo) GetDropshipPurchaseByInvoice(ctx context.Context, kodeInvoice string) (*models.DropshipPurchase, error) {
	return m.purchases[kodeInvoice], nil
}

func (m *BasicMockDropshipRepo) GetDropshipPurchaseByID(ctx context.Context, kodePesanan string) (*models.DropshipPurchase, error) {
	for _, p := range m.purchases {
		if p.KodePesanan == kodePesanan {
			return p, nil
		}
	}
	return nil, nil
}

func (m *BasicMockDropshipRepo) UpdatePurchaseStatus(ctx context.Context, kodePesanan, status string) error {
	return nil
}

func (m *BasicMockDropshipRepo) SumDetailByInvoice(ctx context.Context, kodeInvoice string) (float64, error) {
	return 100.0, nil
}

func (m *BasicMockDropshipRepo) SumProductCostByInvoice(ctx context.Context, kodeInvoice string) (float64, error) {
	return 80.0, nil
}
