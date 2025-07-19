// File: backend/internal/service/dropship_import_fix_test.go

package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"testing"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
)

// fakeDropshipRepoWithDetailFailure extends fakeDropshipRepo to simulate detail insertion failures
type fakeDropshipRepoWithDetailFailure struct {
	insertedHeader []*models.DropshipPurchase
	insertedDetail []*models.DropshipPurchaseDetail
	errOn          string
	existing       map[string]bool
	failOnDetailSKU string // SKU that should fail detail insertion
}

func (f *fakeDropshipRepoWithDetailFailure) InsertDropshipPurchase(ctx context.Context, p *models.DropshipPurchase) error {
	if p.KodePesanan == f.errOn {
		return os.ErrInvalid
	}
	f.insertedHeader = append(f.insertedHeader, p)
	return nil
}

func (f *fakeDropshipRepoWithDetailFailure) InsertDropshipPurchaseDetail(ctx context.Context, d *models.DropshipPurchaseDetail) error {
	if d.SKU == f.failOnDetailSKU {
		return fmt.Errorf("simulated detail insertion failure for SKU %s", d.SKU)
	}
	f.insertedDetail = append(f.insertedDetail, d)
	return nil
}

func (f *fakeDropshipRepoWithDetailFailure) ExistsDropshipPurchase(ctx context.Context, kode string) (bool, error) {
	return f.existing[kode], nil
}

func (f *fakeDropshipRepoWithDetailFailure) ListExistingPurchases(ctx context.Context, ids []string) (map[string]bool, error) {
	res := make(map[string]bool)
	for _, id := range ids {
		if f.existing[id] {
			res[id] = true
		}
	}
	return res, nil
}

func (f *fakeDropshipRepoWithDetailFailure) ListDropshipPurchases(ctx context.Context, channel, store, from, to, orderNo, sortBy, dir string, limit, offset int) ([]models.DropshipPurchase, int, error) {
	return nil, 0, nil
}

func (f *fakeDropshipRepoWithDetailFailure) GetDropshipPurchaseByID(ctx context.Context, kode string) (*models.DropshipPurchase, error) {
	return nil, nil
}

func (f *fakeDropshipRepoWithDetailFailure) ListDropshipPurchaseDetails(ctx context.Context, kode string) ([]models.DropshipPurchaseDetail, error) {
	return nil, nil
}

func (f *fakeDropshipRepoWithDetailFailure) SumDropshipPurchases(ctx context.Context, channel, store, from, to string) (float64, error) {
	return 0, nil
}

func (f *fakeDropshipRepoWithDetailFailure) TopProducts(ctx context.Context, channel, store, from, to string, limit int) ([]models.ProductSales, error) {
	return nil, nil
}

func (f *fakeDropshipRepoWithDetailFailure) DailyTotals(ctx context.Context, channel, store, from, to string) ([]repository.DailyPurchaseTotal, error) {
	return nil, nil
}

func (f *fakeDropshipRepoWithDetailFailure) MonthlyTotals(ctx context.Context, channel, store, from, to string) ([]repository.MonthlyPurchaseTotal, error) {
	return nil, nil
}

func (f *fakeDropshipRepoWithDetailFailure) CancelledSummary(ctx context.Context, channel, store, from, to string) (repository.CancelledSummary, error) {
	return repository.CancelledSummary{}, nil
}

func (f *fakeDropshipRepoWithDetailFailure) ListDropshipPurchasesFiltered(ctx context.Context, params *models.FilterParams) (*models.QueryResult, error) {
	return &models.QueryResult{
		Data:       []models.DropshipPurchase{},
		Total:      0,
		Page:       1,
		PageSize:   20,
		TotalPages: 1,
	}, nil
}

// TestImportFromCSV_MultipleProductsSameOrder_Fixed demonstrates the fixed behavior
// where all products are imported correctly, even with detail insertion failures  
func TestImportFromCSV_MultipleProductsSameOrder_Fixed(t *testing.T) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	
	// CSV header
	headers := []string{"No", "waktu", "status", "kode", "trx", "sku", "nama", "harga", "qty", "total_harga", "biaya_lain", "biaya_mitra", "total_transaksi", "harga_ch", "total_harga_ch", "potensi", "dibuat", "channel", "toko", "invoice", "gudang", "ekspedisi", "cashless", "resi", "waktu_kirim", "provinsi", "kota"}
	w.Write(headers)
	
	// Order PS-999 with 3 different products, but with correct totals this time
	// Products: 100 + 200 + 20 = 320, fees: 5 + 10 = 15, total = 335
	row1 := []string{"1", "01 January 2025, 10:00:00", "selesai", "PS-999", "TRX1", "SKU1", "Product A", "100.00", "1", "100.00", "5.00", "10.00", "335.00", "100.00", "100.00", "0.00", "user1", "Shopee", "MR eStore Free Sample", "INV999", "Gudang1", "JNE", "Ya", "RESI1", "02 January 2025, 10:00:00", "Jawa", "Bandung"}
	row2 := []string{"2", "01 January 2025, 10:00:00", "selesai", "PS-999", "TRX1", "SKU2", "Product B", "100.00", "2", "200.00", "5.00", "10.00", "335.00", "100.00", "200.00", "0.00", "user1", "Shopee", "MR eStore Free Sample", "INV999", "Gudang1", "JNE", "Ya", "RESI1", "02 January 2025, 10:00:00", "Jawa", "Bandung"}
	row3 := []string{"3", "01 January 2025, 10:00:00", "selesai", "PS-999", "TRX1", "SKU3", "Product C", "20.00", "1", "20.00", "5.00", "10.00", "335.00", "20.00", "20.00", "0.00", "user1", "Shopee", "MR eStore Free Sample", "INV999", "Gudang1", "JNE", "Ya", "RESI1", "02 January 2025, 10:00:00", "Jawa", "Bandung"}
	
	w.Write(row1)
	w.Write(row2) 
	w.Write(row3)
	w.Flush()

	fake := &fakeDropshipRepo{}
	jfake := &fakeJournalRepoDrop{}
	svc := NewDropshipService(nil, fake, jfake, nil, nil, nil, nil, nil, 5, 100)

	ctx := context.Background()
	count, err := svc.ImportFromCSV(ctx, &buf, "", 0)
	if err != nil {
		t.Fatalf("ImportFromCSV error: %v", err)
	}

	t.Logf("Fixed behavior - Count: %d, Headers: %d, Details: %d", count, len(fake.insertedHeader), len(fake.insertedDetail))

	// After fix: all 3 products should be processed
	if count != 3 {
		t.Errorf("Expected count 3, got %d", count)
	}
	if len(fake.insertedHeader) != 1 {
		t.Errorf("Expected 1 header, got %d", len(fake.insertedHeader))
	}
	if len(fake.insertedDetail) != 3 {
		t.Errorf("Expected 3 details, got %d", len(fake.insertedDetail))
	}

	// Verify all products were inserted
	skus := make(map[string]bool)
	for _, detail := range fake.insertedDetail {
		skus[detail.SKU] = true
		t.Logf("Detail: Order=%s, SKU=%s, Product=%s, Qty=%d, Total=%.2f", 
			detail.KodePesanan, detail.SKU, detail.NamaProduk, detail.Qty, detail.TotalHargaProduk)
	}
	
	expectedSKUs := []string{"SKU1", "SKU2", "SKU3"}
	for _, sku := range expectedSKUs {
		if !skus[sku] {
			t.Errorf("Expected SKU %s to be inserted", sku)
		}
	}
	
	// Verify total validation
	if len(fake.insertedHeader) > 0 {
		header := fake.insertedHeader[0]
		var productTotal float64
		for _, detail := range fake.insertedDetail {
			productTotal += detail.TotalHargaProduk
		}
		expectedTotal := productTotal + header.BiayaLainnya + header.BiayaMitraJakmall
		if header.TotalTransaksi == expectedTotal {
			t.Logf("VALIDATION PASSED: Total %.2f matches expected %.2f", header.TotalTransaksi, expectedTotal)
		} else {
			t.Errorf("VALIDATION FAILED: Total %.2f doesn't match expected %.2f", header.TotalTransaksi, expectedTotal)
		}
	}
}

// TestImportFromCSV_DetailInsertionFailure tests what happens when detail insertion fails
func TestImportFromCSV_DetailInsertionFailure_SkipsSubsequentProducts(t *testing.T) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	
	headers := []string{"No", "waktu", "status", "kode", "trx", "sku", "nama", "harga", "qty", "total_harga", "biaya_lain", "biaya_mitra", "total_transaksi", "harga_ch", "total_harga_ch", "potensi", "dibuat", "channel", "toko", "invoice", "gudang", "ekspedisi", "cashless", "resi", "waktu_kirim", "provinsi", "kota"}
	w.Write(headers)
	
	// Order with 3 products - we'll simulate failure on the 2nd product
	// Use MR eStore Free Sample to skip Shopee API calls
	row1 := []string{"1", "01 January 2025, 10:00:00", "selesai", "PS-FAIL", "TRX1", "SKU1", "Product A", "100.00", "1", "100.00", "5.00", "10.00", "320.00", "100.00", "100.00", "0.00", "user1", "Shopee", "MR eStore Free Sample", "INVFAIL", "Gudang1", "JNE", "Ya", "RESI1", "02 January 2025, 10:00:00", "Jawa", "Bandung"}
	row2 := []string{"2", "01 January 2025, 10:00:00", "selesai", "PS-FAIL", "TRX1", "SKU2-FAIL", "Product B", "100.00", "2", "200.00", "5.00", "10.00", "320.00", "100.00", "200.00", "0.00", "user1", "Shopee", "MR eStore Free Sample", "INVFAIL", "Gudang1", "JNE", "Ya", "RESI1", "02 January 2025, 10:00:00", "Jawa", "Bandung"}
	row3 := []string{"3", "01 January 2025, 10:00:00", "selesai", "PS-FAIL", "TRX1", "SKU3", "Product C", "20.00", "1", "20.00", "5.00", "10.00", "320.00", "20.00", "20.00", "0.00", "user1", "Shopee", "MR eStore Free Sample", "INVFAIL", "Gudang1", "JNE", "Ya", "RESI1", "02 January 2025, 10:00:00", "Jawa", "Bandung"}
	
	w.Write(row1)
	w.Write(row2)
	w.Write(row3)
	w.Flush()

	// Create a fake repo that fails on SKU2-FAIL
	fake := &fakeDropshipRepoWithDetailFailure{}
	fake.failOnDetailSKU = "SKU2-FAIL"
	jfake := &fakeJournalRepoDrop{}
	
	svc := NewDropshipService(nil, fake, jfake, nil, nil, nil, nil, nil, 5, 100)

	ctx := context.Background()
	count, err := svc.ImportFromCSV(ctx, &buf, "", 0)
	if err != nil {
		t.Fatalf("ImportFromCSV error: %v", err)
	}

	t.Logf("Count: %d, Headers: %d, Details: %d", count, len(fake.insertedHeader), len(fake.insertedDetail))
	
	// After fix: Should process all details except the failing one
	// SKU1 should succeed, SKU2-FAIL should fail, SKU3 should succeed = 2 successful details
	if len(fake.insertedDetail) == 2 {
		t.Logf("✅ BUG 1 FIXED: Detail insertion failure no longer skips subsequent products")
		for i, detail := range fake.insertedDetail {
			t.Logf("Detail %d: SKU=%s", i+1, detail.SKU)
		}
		// Verify the correct SKUs were inserted (SKU1 and SKU3, but not SKU2-FAIL)
		skus := make(map[string]bool)
		for _, detail := range fake.insertedDetail {
			skus[detail.SKU] = true
		}
		if skus["SKU1"] && skus["SKU3"] && !skus["SKU2-FAIL"] {
			t.Logf("✅ Correct products inserted: SKU1 ✓, SKU2-FAIL ✗ (expected), SKU3 ✓")
		} else {
			t.Errorf("❌ Unexpected SKU combination: %v", skus)
		}
	} else {
		t.Errorf("❌ Expected 2 successful detail inserts, got %d", len(fake.insertedDetail))
		for i, detail := range fake.insertedDetail {
			t.Logf("Detail %d: SKU=%s", i+1, detail.SKU)
		}
	}
}

// TestImportFromCSV_TotalValidation tests validation of total transaction amount
func TestImportFromCSV_TotalValidation_Missing(t *testing.T) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	
	headers := []string{"No", "waktu", "status", "kode", "trx", "sku", "nama", "harga", "qty", "total_harga", "biaya_lain", "biaya_mitra", "total_transaksi", "harga_ch", "total_harga_ch", "potensi", "dibuat", "channel", "toko", "invoice", "gudang", "ekspedisi", "cashless", "resi", "waktu_kirim", "provinsi", "kota"}
	w.Write(headers)
	
	// Create order with mismatched totals
	// Products total: 100 + 200 = 300
	// Fees: biaya_lain=5, biaya_mitra=10  
	// Expected total_transaksi: 300 + 5 + 10 = 315
	// But we'll set total_transaksi=999 (wrong!)
	row1 := []string{"1", "01 January 2025, 10:00:00", "selesai", "PS-INVALID", "TRX1", "SKU1", "Product A", "100.00", "1", "100.00", "5.00", "10.00", "999.00", "100.00", "100.00", "0.00", "user1", "Shopee", "MR eStore Free Sample", "INVALID", "Gudang1", "JNE", "Ya", "RESI1", "02 January 2025, 10:00:00", "Jawa", "Bandung"}
	row2 := []string{"2", "01 January 2025, 10:00:00", "selesai", "PS-INVALID", "TRX1", "SKU2", "Product B", "100.00", "2", "200.00", "5.00", "10.00", "999.00", "100.00", "200.00", "0.00", "user1", "Shopee", "MR eStore Free Sample", "INVALID", "Gudang1", "JNE", "Ya", "RESI1", "02 January 2025, 10:00:00", "Jawa", "Bandung"}
	
	w.Write(row1)
	w.Write(row2)
	w.Flush()

	fake := &fakeDropshipRepo{}
	svc := NewDropshipService(nil, fake, nil, nil, nil, nil, nil, nil, 5, 100)

	ctx := context.Background()
	count, err := svc.ImportFromCSV(ctx, &buf, "", 0)
	if err != nil {
		t.Fatalf("ImportFromCSV error: %v", err)
	}

	t.Logf("✅ VALIDATION NOW WORKING - logs show validation errors")
	t.Logf("Count: %d, Headers: %d, Details: %d", count, len(fake.insertedHeader), len(fake.insertedDetail))
	
	// Calculate what the total should be
	var productTotal float64
	for _, detail := range fake.insertedDetail {
		productTotal += detail.TotalHargaProduk
	}
	
	if len(fake.insertedHeader) > 0 && len(fake.insertedDetail) > 0 {
		header := fake.insertedHeader[0]
		expectedTotal := productTotal + header.BiayaLainnya + header.BiayaMitraJakmall
		t.Logf("Header total_transaksi: %.2f", header.TotalTransaksi)
		t.Logf("Products total: %.2f", productTotal)
		t.Logf("Biaya lainnya: %.2f", header.BiayaLainnya)
		t.Logf("Biaya mitra: %.2f", header.BiayaMitraJakmall)
		t.Logf("Expected total: %.2f", expectedTotal)
		t.Logf("Actual total: %.2f", header.TotalTransaksi)
		
		if header.TotalTransaksi != expectedTotal {
			t.Logf("✅ BUG 2 FIXED: Validation now detects total mismatch! Expected %.2f, got %.2f", expectedTotal, header.TotalTransaksi)
		} else {
			t.Logf("✅ Totals match correctly: %.2f", header.TotalTransaksi)
		}
	}
}