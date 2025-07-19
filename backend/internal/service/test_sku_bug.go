// Test to reproduce the SKU bug reported by the user
package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"testing"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// This test reproduces the issue where two different SKUs for the same order
// result in duplicate details with the same SKU instead of preserving both SKUs
func TestImportFromCSV_SamePesananDifferentSKUs_Bug(t *testing.T) {
	// Setup: Create CSV data similar to the user's example
	csvData := [][]string{
		// Headers (would be skipped in real processing)
		{"no", "waktu_pesanan_terbuat", "status_pesanan_terakhir", "kode_pesanan", "kode_transaksi", "sku", "nama_produk", "harga_produk", "qty", "total_harga_produk", "biaya_lainnya", "biaya_mitra_jakmall", "total_transaksi", "harga_produk_channel", "total_harga_produk_channel", "potensi_keuntungan", "dibuat_oleh", "jenis_channel", "nama_toko", "kode_invoice_channel", "gudang_pengiriman", "jenis_ekspedisi", "cashless", "nomor_resi", "waktu_pengiriman", "provinsi", "kota"},
		// Two rows with same kode_pesanan but different SKUs
		{"1", "10 January 2025, 23:38:50", "Pesanan selesai", "26137342285", "1777530433", "7CHZ16BK", "NIKTO Lakban Sticker Magnetic Strong Tape Self Adhesive 1M - TMK92 Hitam 10MM", "3700", "2", "7400", "0", "2200", "24000", "7600", "15200", "7800", "Muhammad Ramadhan", "Shopee", "MR eStore Shopee", "2501116HYUP1VC", "Gudang Online / COD DKI Jakarta", "Shopee-Xpress STD", "Ya", "SPXID052542447131", "11 January 2025, 06:29:01", "Jawa Barat", "Kab. Indramayu"},
		{"2", "10 January 2025, 23:38:50", "Pesanan selesai", "26137342285", "1777530433", "7CHZ6NBR", "TaffPACK Tape Lakban PTFE Heat High Temperature Insulation 10M - TF10M Coklat 20 mm", "14400", "1", "14400", "0", "0", "22100", "22100", "7700", "0", "Muhammad Ramadhan", "Shopee", "MR eStore Shopee", "2501116HYUP1VC", "Gudang Online / COD DKI Jakarta", "Shopee-Xpress STD", "Ya", "SPXID052542447131", "11 January 2025, 06:29:01", "Jawa Barat", "Kab. Indramayu"},
	}

	// Create CSV content
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	for _, record := range csvData {
		writer.Write(record)
	}
	writer.Flush()

	// Setup fake repository
	repo := &fakeDropshipRepoWithDetailFailure{
		insertedHeader: []*models.DropshipPurchase{},
		insertedDetail: []*models.DropshipPurchaseDetail{},
		existing:       make(map[string]bool),
	}

	// Setup dropship service
	journalRepo := &fakeJournalRepoDrop{}
	service := NewDropshipService(nil, repo, journalRepo, nil, nil, nil, nil, nil, 5, 100)

	// Execute import
	ctx := context.Background()
	count, err := service.ImportFromCSV(ctx, bytes.NewReader(buf.Bytes()), "", 0)
	if err != nil {
		t.Fatalf("ImportFromCSV failed: %v", err)
	}

	t.Logf("Import count: %d", count)

	// Verify results
	t.Logf("Number of headers inserted: %d", len(repo.insertedHeader))
	t.Logf("Number of details inserted: %d", len(repo.insertedDetail))

	// Should have 1 header (one order)
	if len(repo.insertedHeader) != 1 {
		t.Errorf("Expected 1 header, got %d", len(repo.insertedHeader))
	}

	// Should have 2 details (two different products)
	if len(repo.insertedDetail) != 2 {
		t.Errorf("Expected 2 details, got %d", len(repo.insertedDetail))
	}

	// Verify that the SKUs are different
	if len(repo.insertedDetail) >= 2 {
		sku1 := repo.insertedDetail[0].SKU
		sku2 := repo.insertedDetail[1].SKU
		
		t.Logf("Detail 1 SKU: %s", sku1)
		t.Logf("Detail 2 SKU: %s", sku2)
		
		// This is the bug: both details might have the same SKU
		if sku1 == sku2 {
			t.Errorf("BUG REPRODUCED: Both details have the same SKU (%s), expected different SKUs (7CHZ16BK and 7CHZ6NBR)", sku1)
		}
		
		// Expected: different SKUs
		expectedSKUs := map[string]bool{"7CHZ16BK": false, "7CHZ6NBR": false}
		for _, detail := range repo.insertedDetail {
			if _, exists := expectedSKUs[detail.SKU]; exists {
				expectedSKUs[detail.SKU] = true
			}
		}
		
		for sku, found := range expectedSKUs {
			if !found {
				t.Errorf("Expected SKU %s not found in details", sku)
			}
		}
	}
}