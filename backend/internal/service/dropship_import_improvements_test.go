// File: backend/internal/service/dropship_import_improvements_test.go

package service

import (
	"testing"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

func TestGroupRecordsByOrder(t *testing.T) {
	// Test CSV records with multiple products for the same order
	records := [][]string{
		// Header row would already be consumed by this point
		{"1", "01 January 2025, 10:00:00", "selesai", "PS-123", "TRX1", "SKU1", "Product A", "100.00", "1", "100.00", "5.00", "10.00", "315.00", "100.00", "100.00", "0.00", "user1", "Shopee", "TestShop", "INV123", "Gudang1", "JNE", "Ya", "RESI1", "02 January 2025, 10:00:00", "Jawa", "Bandung"},
		{"2", "01 January 2025, 10:00:00", "selesai", "PS-123", "TRX1", "SKU2", "Product B", "100.00", "2", "200.00", "5.00", "10.00", "315.00", "100.00", "200.00", "0.00", "user1", "Shopee", "TestShop", "INV123", "Gudang1", "JNE", "Ya", "RESI1", "02 January 2025, 10:00:00", "Jawa", "Bandung"},
		{"3", "01 January 2025, 10:00:00", "selesai", "PS-456", "TRX2", "SKU3", "Product C", "50.00", "1", "50.00", "2.50", "5.00", "57.50", "50.00", "50.00", "0.00", "user2", "Shopee", "TestShop", "INV456", "Gudang1", "JNE", "Ya", "RESI2", "02 January 2025, 10:00:00", "Jawa", "Bandung"},
	}

	groups, err := groupRecordsByOrder(records, "")
	if err != nil {
		t.Fatalf("groupRecordsByOrder failed: %v", err)
	}

	// Should have 2 orders
	if len(groups) != 2 {
		t.Errorf("Expected 2 orders, got %d", len(groups))
	}

	// Check PS-123 order (2 products)
	order123 := groups["PS-123"]
	if order123 == nil {
		t.Fatal("PS-123 order not found")
	}
	if len(order123.Details) != 2 {
		t.Errorf("Expected 2 details for PS-123, got %d", len(order123.Details))
	}
	if order123.Header.KodePesanan != "PS-123" {
		t.Errorf("Expected header kode_pesanan PS-123, got %s", order123.Header.KodePesanan)
	}
	if order123.Header.TotalTransaksi != 315.00 {
		t.Errorf("Expected total_transaksi 315.00, got %.2f", order123.Header.TotalTransaksi)
	}

	// Check PS-456 order (1 product)
	order456 := groups["PS-456"]
	if order456 == nil {
		t.Fatal("PS-456 order not found")
	}
	if len(order456.Details) != 1 {
		t.Errorf("Expected 1 detail for PS-456, got %d", len(order456.Details))
	}

	// Verify product details
	skus := make(map[string]bool)
	for _, detail := range order123.Details {
		skus[detail.SKU] = true
	}
	if !skus["SKU1"] || !skus["SKU2"] {
		t.Errorf("Expected SKU1 and SKU2 for PS-123, got %v", skus)
	}
}

func TestGroupRecordsByOrder_ChannelFilter(t *testing.T) {
	records := [][]string{
		{"1", "01 January 2025, 10:00:00", "selesai", "PS-SHOPEE", "TRX1", "SKU1", "Product A", "100.00", "1", "100.00", "5.00", "10.00", "115.00", "100.00", "100.00", "0.00", "user1", "Shopee", "TestShop", "INV1", "Gudang1", "JNE", "Ya", "RESI1", "02 January 2025, 10:00:00", "Jawa", "Bandung"},
		{"2", "01 January 2025, 10:00:00", "selesai", "PS-TOKPED", "TRX2", "SKU2", "Product B", "100.00", "1", "100.00", "5.00", "10.00", "115.00", "100.00", "100.00", "0.00", "user2", "Tokopedia", "TestShop", "INV2", "Gudang1", "JNE", "Ya", "RESI2", "02 January 2025, 10:00:00", "Jawa", "Bandung"},
	}

	// Filter by Shopee channel
	groups, err := groupRecordsByOrder(records, "Shopee")
	if err != nil {
		t.Fatalf("groupRecordsByOrder failed: %v", err)
	}

	// Should only have Shopee order
	if len(groups) != 1 {
		t.Errorf("Expected 1 order with Shopee filter, got %d", len(groups))
	}
	if groups["PS-SHOPEE"] == nil {
		t.Error("Expected PS-SHOPEE order to be included")
	}
	if groups["PS-TOKPED"] != nil {
		t.Error("Expected PS-TOKPED order to be filtered out")
	}
}

func TestValidateTransactionTotals(t *testing.T) {
	tests := []struct {
		name          string
		header        *models.DropshipPurchase
		details       []*models.DropshipPurchaseDetail
		shouldPass    bool
		expectedError string
	}{
		{
			name: "valid totals",
			header: &models.DropshipPurchase{
				TotalTransaksi:    315.00,
				BiayaLainnya:     5.00,
				BiayaMitraJakmall: 10.00,
			},
			details: []*models.DropshipPurchaseDetail{
				{TotalHargaProduk: 100.00},
				{TotalHargaProduk: 200.00},
			},
			shouldPass: true,
		},
		{
			name: "invalid totals - too high",
			header: &models.DropshipPurchase{
				TotalTransaksi:    999.00,
				BiayaLainnya:     5.00,
				BiayaMitraJakmall: 10.00,
			},
			details: []*models.DropshipPurchaseDetail{
				{TotalHargaProduk: 100.00},
				{TotalHargaProduk: 200.00},
			},
			shouldPass:    false,
			expectedError: "total transaction validation failed",
		},
		{
			name: "invalid totals - too low",
			header: &models.DropshipPurchase{
				TotalTransaksi:    100.00,
				BiayaLainnya:     5.00,
				BiayaMitraJakmall: 10.00,
			},
			details: []*models.DropshipPurchaseDetail{
				{TotalHargaProduk: 100.00},
				{TotalHargaProduk: 200.00},
			},
			shouldPass:    false,
			expectedError: "total transaction validation failed",
		},
		{
			name: "rounding tolerance",
			header: &models.DropshipPurchase{
				TotalTransaksi:    315.001, // within 0.01 tolerance
				BiayaLainnya:     5.00,
				BiayaMitraJakmall: 10.00,
			},
			details: []*models.DropshipPurchaseDetail{
				{TotalHargaProduk: 100.00},
				{TotalHargaProduk: 200.00},
			},
			shouldPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTransactionTotals(tt.header, tt.details)
			if tt.shouldPass {
				if err != nil {
					t.Errorf("Expected validation to pass, got error: %v", err)
				}
			} else {
				if err == nil {
					t.Error("Expected validation to fail, got no error")
				} else if tt.expectedError != "" && !contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s', got: %v", tt.expectedError, err)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || 
		   len(s) > len(substr) && s[len(s)-len(substr):] == substr ||
		   len(s) > len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}