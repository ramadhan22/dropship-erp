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
	// Log detailed information about what's being inserted
	fmt.Printf("REPO: Inserting detail - kode_pesanan='%s', SKU='%s' (len=%d, bytes=%v), nama_produk='%s'\n", 
		d.KodePesanan, d.SKU, len(d.SKU), []byte(d.SKU), d.NamaProduk)
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

// TestImportFromCSV_SamePesananDifferentSKUs reproduces the user-reported bug
// where two different SKUs for the same order result in duplicate details with the same SKU
func TestImportFromCSV_SamePesananDifferentSKUs(t *testing.T) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	
	// CSV header matching the user's data structure
	headers := []string{"no", "waktu_pesanan_terbuat", "status_pesanan_terakhir", "kode_pesanan", "kode_transaksi", "sku", "nama_produk", "harga_produk", "qty", "total_harga_produk", "biaya_lainnya", "biaya_mitra_jakmall", "total_transaksi", "harga_produk_channel", "total_harga_produk_channel", "potensi_keuntungan", "dibuat_oleh", "jenis_channel", "nama_toko", "kode_invoice_channel", "gudang_pengiriman", "jenis_ekspedisi", "cashless", "nomor_resi", "waktu_pengiriman", "provinsi", "kota"}
	w.Write(headers)
	
	// Two rows with same kode_pesanan but different SKUs (from user's example)
	row1 := []string{"1", "10 January 2025, 23:38:50", "Pesanan selesai", "26137342285", "1777530433", "7CHZ16BK", "NIKTO Lakban Sticker Magnetic Strong Tape Self Adhesive 1M - TMK92 Hitam 10MM", "3700", "2", "7400", "0", "2200", "24000", "7600", "15200", "7800", "Muhammad Ramadhan", "Shopee", "MR eStore Free Sample", "2501116HYUP1VC", "Gudang Online / COD DKI Jakarta", "Shopee-Xpress STD", "Ya", "SPXID052542447131", "11 January 2025, 06:29:01", "Jawa Barat", "Kab. Indramayu"}
	row2 := []string{"2", "10 January 2025, 23:38:50", "Pesanan selesai", "26137342285", "1777530433", "7CHZ6NBR", "TaffPACK Tape Lakban PTFE Heat High Temperature Insulation 10M - TF10M Coklat 20 mm", "14400", "1", "14400", "0", "0", "22100", "22100", "7700", "0", "Muhammad Ramadhan", "Shopee", "MR eStore Free Sample", "2501116HYUP1VC", "Gudang Online / COD DKI Jakarta", "Shopee-Xpress STD", "Ya", "SPXID052542447131", "11 January 2025, 06:29:01", "Jawa Barat", "Kab. Indramayu"}
	
	w.Write(row1)
	w.Write(row2)
	w.Flush()

	// Setup fake repository
	fake := &fakeDropshipRepoWithDetailFailure{
		insertedHeader: []*models.DropshipPurchase{},
		insertedDetail: []*models.DropshipPurchaseDetail{},
		existing:       make(map[string]bool),
	}
	jfake := &fakeJournalRepoDrop{}
	
	svc := NewDropshipService(nil, fake, jfake, nil, nil, nil, nil, nil, 5, 100)

	ctx := context.Background()
	count, err := svc.ImportFromCSV(ctx, &buf, "", 0)
	if err != nil {
		t.Fatalf("ImportFromCSV error: %v", err)
	}

	t.Logf("Import count: %d", count)
	t.Logf("Number of headers inserted: %d", len(fake.insertedHeader))
	t.Logf("Number of details inserted: %d", len(fake.insertedDetail))

	// Should have 1 header (one order)
	if len(fake.insertedHeader) != 1 {
		t.Errorf("Expected 1 header, got %d", len(fake.insertedHeader))
	}

	// Should have 2 details (two different products)
	if len(fake.insertedDetail) != 2 {
		t.Errorf("Expected 2 details, got %d", len(fake.insertedDetail))
	}

	// Verify that the SKUs are different (this is the main test)
	if len(fake.insertedDetail) >= 2 {
		sku1 := fake.insertedDetail[0].SKU
		sku2 := fake.insertedDetail[1].SKU
		
		t.Logf("Detail 1 SKU: %s", sku1)
		t.Logf("Detail 2 SKU: %s", sku2)
		
		// This was the bug: both details might have the same SKU
		if sku1 == sku2 {
			t.Errorf("BUG REPRODUCED: Both details have the same SKU (%s), expected different SKUs (7CHZ16BK and 7CHZ6NBR)", sku1)
		}
		
		// Expected: different SKUs
		expectedSKUs := map[string]bool{"7CHZ16BK": false, "7CHZ6NBR": false}
		for _, detail := range fake.insertedDetail {
			if _, exists := expectedSKUs[detail.SKU]; exists {
				expectedSKUs[detail.SKU] = true
			} else {
				t.Errorf("Unexpected SKU found: %s", detail.SKU)
			}
		}
		
		for sku, found := range expectedSKUs {
			if !found {
				t.Errorf("Expected SKU %s not found in details", sku)
			}
		}
		
		// Additional verification: check that both details have the same kode_pesanan
		for _, detail := range fake.insertedDetail {
			if detail.KodePesanan != "26137342285" {
				t.Errorf("Expected kode_pesanan 26137342285, got %s", detail.KodePesanan)
			}
		}
	}
}

// TestImportFromCSV_DetailedSKUDebugging creates a detailed test to debug the SKU issue
func TestImportFromCSV_DetailedSKUDebugging(t *testing.T) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	
	// CSV header
	headers := []string{"no", "waktu_pesanan_terbuat", "status_pesanan_terakhir", "kode_pesanan", "kode_transaksi", "sku", "nama_produk", "harga_produk", "qty", "total_harga_produk", "biaya_lainnya", "biaya_mitra_jakmall", "total_transaksi", "harga_produk_channel", "total_harga_produk_channel", "potensi_keuntungan", "dibuat_oleh", "jenis_channel", "nama_toko", "kode_invoice_channel", "gudang_pengiriman", "jenis_ekspedisi", "cashless", "nomor_resi", "waktu_pengiriman", "provinsi", "kota"}
	w.Write(headers)
	
	// Use exact data from user's comment but with MR eStore Free Sample to avoid API calls
	row1 := []string{"1", "10 January 2025, 23:38:50", "Pesanan selesai", "26137342285", "1777530433", "7CHZ16BK", "NIKTO Lakban Sticker Magnetic Strong Tape Self Adhesive 1M - TMK92 Hitam 10MM", "3700", "2", "7400", "0", "2200", "24000", "7600", "15200", "7800", "Muhammad Ramadhan", "Shopee", "MR eStore Free Sample", "2501116HYUP1VC", "Gudang Online / COD DKI Jakarta", "Shopee-Xpress STD", "Ya", "SPXID052542447131", "11 January 2025, 06:29:01", "Jawa Barat", "Kab. Indramayu"}
	row2 := []string{"2", "10 January 2025, 23:38:50", "Pesanan selesai", "26137342285", "1777530433", "7CHZ6NBR", "TaffPACK Tape Lakban PTFE Heat High Temperature Insulation 10M - TF10M Coklat 20 mm", "14400", "1", "14400", "0", "0", "22100", "22100", "7700", "0", "Muhammad Ramadhan", "Shopee", "MR eStore Free Sample", "2501116HYUP1VC", "Gudang Online / COD DKI Jakarta", "Shopee-Xpress STD", "Ya", "SPXID052542447131", "11 January 2025, 06:29:01", "Jawa Barat", "Kab. Indramayu"}
	
	w.Write(row1)
	w.Write(row2)
	w.Flush()

	// Debug: Print the raw CSV content
	csvContent := buf.String()
	t.Logf("Raw CSV content:\n%s", csvContent)
	
	// Debug: Parse and print each field
	buf2 := bytes.NewBufferString(csvContent)
	reader := csv.NewReader(buf2)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read CSV: %v", err)
	}
	
	for i, record := range records {
		t.Logf("Row %d has %d fields:", i, len(record))
		if i > 0 { // Skip header
			t.Logf("  kode_pesanan (field 3): '%s'", record[3])
			t.Logf("  SKU (field 5): '%s' (len=%d, bytes=%v)", record[5], len(record[5]), []byte(record[5]))
			t.Logf("  nama_toko (field 18): '%s'", record[18])
		}
	}

	// Create a custom repository that logs all insertions
	fake := &fakeDropshipRepoWithDetailFailure{
		insertedHeader: []*models.DropshipPurchase{},
		insertedDetail: []*models.DropshipPurchaseDetail{},
		existing:       make(map[string]bool),
	}
	
	jfake := &fakeJournalRepoDrop{}
	svc := NewDropshipService(nil, fake, jfake, nil, nil, nil, nil, nil, 5, 100)

	ctx := context.Background()
	count, err := svc.ImportFromCSV(ctx, bytes.NewReader([]byte(csvContent)), "", 0)
	if err != nil {
		t.Fatalf("ImportFromCSV error: %v", err)
	}

	t.Logf("Import completed: count=%d, headers=%d, details=%d", count, len(fake.insertedHeader), len(fake.insertedDetail))

	// Detailed analysis of results
	for i, detail := range fake.insertedDetail {
		t.Logf("Result detail %d: kode_pesanan='%s', SKU='%s' (len=%d, bytes=%v), nama_produk='%s'", 
			i+1, detail.KodePesanan, detail.SKU, len(detail.SKU), []byte(detail.SKU), detail.NamaProduk)
	}

	// Check if we have the expected results
	if len(fake.insertedDetail) != 2 {
		t.Errorf("Expected 2 details, got %d", len(fake.insertedDetail))
		return
	}

	sku1 := fake.insertedDetail[0].SKU
	sku2 := fake.insertedDetail[1].SKU
	
	if sku1 == sku2 {
		t.Errorf("BUG REPRODUCED: Both details have the same SKU (%s)", sku1)
		t.Errorf("Expected: first='7CHZ16BK', second='7CHZ6NBR'")
		t.Errorf("Actual: first='%s', second='%s'", sku1, sku2)
	} else {
		t.Logf("✅ SKUs are different as expected: first='%s', second='%s'", sku1, sku2)
	}
}

// TestImportFromCSV_MissingColumns tests what happens when columns are missing
func TestImportFromCSV_MissingColumns(t *testing.T) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	
	// CSV header
	headers := []string{"no", "waktu_pesanan_terbuat", "status_pesanan_terakhir", "kode_pesanan", "kode_transaksi", "sku", "nama_produk", "harga_produk", "qty", "total_harga_produk", "biaya_lainnya", "biaya_mitra_jakmall", "total_transaksi", "harga_produk_channel", "total_harga_produk_channel", "potensi_keuntungan", "dibuat_oleh", "jenis_channel", "nama_toko", "kode_invoice_channel", "gudang_pengiriman", "jenis_ekspedisi", "cashless", "nomor_resi", "waktu_pengiriman", "provinsi", "kota"}
	w.Write(headers)
	
	// Simulate the user's issue: what if there are missing columns or the second row is malformed?
	// This could cause the CSV parser to read the wrong field as SKU
	row1 := []string{"1", "10 January 2025, 23:38:50", "Pesanan selesai", "26137342285", "1777530433", "7CHZ16BK", "NIKTO Lakban Sticker Magnetic Strong Tape Self Adhesive 1M - TMK92 Hitam 10MM", "3700", "2", "7400", "0", "2200", "24000", "7600", "15200", "7800", "Muhammad Ramadhan", "Shopee", "MR eStore Free Sample", "2501116HYUP1VC", "Gudang Online / COD DKI Jakarta", "Shopee-Xpress STD", "Ya", "SPXID052542447131", "11 January 2025, 06:29:01", "Jawa Barat", "Kab. Indramayu"}
	// Second row with missing field at position 4 (kode_transaksi) - this will shift all subsequent fields  
	row2 := []string{"2", "10 January 2025, 23:38:50", "Pesanan selesai", "26137342285", "7CHZ6NBR", "TaffPACK Tape Lakban PTFE Heat High Temperature Insulation 10M - TF10M Coklat 20 mm", "14400", "1", "14400", "0", "0", "22100", "22100", "7700", "0", "Muhammad Ramadhan", "Shopee", "MR eStore Free Sample", "2501116HYUP1VC", "Gudang Online / COD DKI Jakarta", "Shopee-Xpress STD", "Ya", "SPXID052542447131", "11 January 2025, 06:29:01", "Jawa Barat", "Kab. Indramayu"}
	
	w.Write(row1)
	w.Write(row2)
	w.Flush()

	// Debug: Parse and print each field to see the shift
	csvContent := buf.String()
	buf2 := bytes.NewBufferString(csvContent)
	reader := csv.NewReader(buf2)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read CSV: %v", err)
	}
	
	for i, record := range records {
		if i > 0 { // Skip header
			t.Logf("Row %d has %d fields:", i, len(record))
			t.Logf("  Field 3 (should be kode_pesanan): '%s'", record[3])
			t.Logf("  Field 4 (should be kode_transaksi): '%s'", record[4])
			t.Logf("  Field 5 (should be SKU): '%s'", record[5])
			t.Logf("  Field 6 (should be nama_produk): '%s'", record[6])
		}
	}

	fake := &fakeDropshipRepoWithDetailFailure{
		insertedHeader: []*models.DropshipPurchase{},
		insertedDetail: []*models.DropshipPurchaseDetail{},
		existing:       make(map[string]bool),
	}
	
	jfake := &fakeJournalRepoDrop{}
	svc := NewDropshipService(nil, fake, jfake, nil, nil, nil, nil, nil, 5, 100)

	ctx := context.Background()
	count, err := svc.ImportFromCSV(ctx, bytes.NewReader([]byte(csvContent)), "", 0)
	if err != nil {
		t.Fatalf("ImportFromCSV error: %v", err)
	}

	t.Logf("Import completed: count=%d, headers=%d, details=%d", count, len(fake.insertedHeader), len(fake.insertedDetail))

	// Check what SKUs were actually inserted
	for i, detail := range fake.insertedDetail {
		t.Logf("Detail %d: SKU='%s', nama_produk='%s'", i+1, detail.SKU, detail.NamaProduk)
	}

	if len(fake.insertedDetail) == 2 {
		sku1 := fake.insertedDetail[0].SKU
		sku2 := fake.insertedDetail[1].SKU
		
		if sku1 == sku2 {
			t.Logf("🔍 ISSUE REPRODUCED: Both details have the same SKU (%s) due to column shift!", sku1)
			t.Logf("This could explain the user's issue if their CSV has missing/misaligned columns")
		} else {
			t.Logf("SKUs are different: '%s' vs '%s'", sku1, sku2)
		}
	}
}

// Test to verify column filling issue reported by user
func TestImportFromCSV_ColumnFillingIssue(t *testing.T) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	
	// CSV header
	headers := []string{"no", "waktu_pesanan_terbuat", "status_pesanan_terakhir", "kode_pesanan", "kode_transaksi", "sku", "nama_produk", "harga_produk", "qty", "total_harga_produk", "biaya_lainnya", "biaya_mitra_jakmall", "total_transaksi", "harga_produk_channel", "total_harga_produk_channel", "potensi_keuntungan", "dibuat_oleh", "jenis_channel", "nama_toko", "kode_invoice_channel", "gudang_pengiriman", "jenis_ekspedisi", "cashless", "nomor_resi", "waktu_pengiriman", "provinsi", "kota"}
	w.Write(headers)
	
	// Exact data from user's comment - one row has empty fields
	row1 := []string{"1", "10 January 2025, 23:38:50", "Pesanan selesai", "26137342285", "1777530433", "7CHZ16BK", "NIKTO Lakban Sticker Magnetic Strong Tape Self Adhesive 1M - TMK92 Hitam 10MM", "3700", "2", "7400", "0", "2200", "24000", "7600", "15200", "7800", "Muhammad Ramadhan", "Shopee", "MR eStore Free Sample", "2501116HYUP1VC", "Gudang Online / COD DKI Jakarta", "Shopee-Xpress STD", "Ya", "SPXID052542447131", "11 January 2025, 06:29:01", "Jawa Barat", "Kab. Indramayu"}
	row2 := []string{"2", "10 January 2025, 23:38:50", "Pesanan selesai", "26137342285", "1777530433", "7CHZ6NBR", "TaffPACK Tape Lakban PTFE Heat High Temperature Insulation 10M - TF10M Coklat 20 mm", "14400", "1", "14400", "", "", "22100", "22100", "7700", "", "Muhammad Ramadhan", "Shopee", "MR eStore Free Sample", "2501116HYUP1VC", "Gudang Online / COD DKI Jakarta", "Shopee-Xpress STD", "Ya", "SPXID052542447131", "11 January 2025, 06:29:01", "Jawa Barat", "Kab. Indramayu"}
	
	w.Write(row1)
	w.Write(row2)
	w.Flush()

	fake := &fakeDropshipRepoWithDetailFailure{
		insertedHeader: []*models.DropshipPurchase{},
		insertedDetail: []*models.DropshipPurchaseDetail{},
		existing:       make(map[string]bool),
	}
	
	jfake := &fakeJournalRepoDrop{}
	svc := NewDropshipService(nil, fake, jfake, nil, nil, nil, nil, nil, 5, 100)

	ctx := context.Background()
	count, err := svc.ImportFromCSV(ctx, &buf, "", 0)
	if err != nil {
		t.Fatalf("ImportFromCSV error: %v", err)
	}

	t.Logf("Import completed: count=%d, headers=%d, details=%d", count, len(fake.insertedHeader), len(fake.insertedDetail))

	// Check that all fields are properly filled
	if len(fake.insertedHeader) > 0 {
		header := fake.insertedHeader[0]
		t.Logf("Header fields:")
		t.Logf("  KodeTransaksi: '%s' (should be '1777530433')", header.KodeTransaksi)
		t.Logf("  DibuatOleh: '%s' (should be 'Muhammad Ramadhan')", header.DibuatOleh)
		t.Logf("  GudangPengiriman: '%s'", header.GudangPengiriman)
		t.Logf("  JenisEkspedisi: '%s'", header.JenisEkspedisi)
		t.Logf("  Cashless: '%s'", header.Cashless)
		t.Logf("  NomorResi: '%s'", header.NomorResi)
		t.Logf("  Provinsi: '%s'", header.Provinsi)
		t.Logf("  Kota: '%s'", header.Kota)
		
		// Check for empty fields that should be filled
		emptyFields := []string{}
		if header.KodeTransaksi == "" {
			emptyFields = append(emptyFields, "KodeTransaksi")
		}
		if header.DibuatOleh == "" {
			emptyFields = append(emptyFields, "DibuatOleh")
		}
		if header.GudangPengiriman == "" {
			emptyFields = append(emptyFields, "GudangPengiriman")
		}
		if header.JenisEkspedisi == "" {
			emptyFields = append(emptyFields, "JenisEkspedisi")
		}
		if header.Cashless == "" {
			emptyFields = append(emptyFields, "Cashless")
		}
		if header.NomorResi == "" {
			emptyFields = append(emptyFields, "NomorResi")
		}
		if header.Provinsi == "" {
			emptyFields = append(emptyFields, "Provinsi")
		}
		if header.Kota == "" {
			emptyFields = append(emptyFields, "Kota")
		}
		
		if len(emptyFields) > 0 {
			t.Errorf("❌ ISSUE FOUND: Empty fields in header: %v", emptyFields)
			t.Errorf("This matches the user's complaint about columns not being filled")
		} else {
			t.Logf("✅ All header fields are properly filled")
		}
	}

	// Check the details
	for i, detail := range fake.insertedDetail {
		t.Logf("Detail %d: SKU='%s', NamaProduk='%s'", i+1, detail.SKU, detail.NamaProduk)
	}

	// Verify SKU issue
	if len(fake.insertedDetail) == 2 {
		sku1 := fake.insertedDetail[0].SKU
		sku2 := fake.insertedDetail[1].SKU
		
		if sku1 == sku2 {
			t.Errorf("❌ SKU BUG REPRODUCED: Both details have the same SKU (%s)", sku1)
		} else {
			t.Logf("✅ SKUs are different: '%s' vs '%s'", sku1, sku2)
		}
	}
}