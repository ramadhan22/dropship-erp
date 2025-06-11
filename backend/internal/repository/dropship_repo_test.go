// File: backend/internal/repository/dropship_repo_test.go

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// tempCleanupDropship deletes any row with the given purchaseID.
func tempCleanupDropship(t *testing.T, kodePesanan string) {
	_, err := testDB.ExecContext(context.Background(),
		"DELETE FROM dropship_purchases WHERE kode_pesanan = $1", kodePesanan)
	if err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}
}

func TestInsertAndGetDropshipPurchase(t *testing.T) {
	ctx := context.Background()
	repo := NewDropshipRepo(testDB)

	kode := "TEST-DS-" + time.Now().Format("20060102150405")
	ds := &models.DropshipPurchase{
		KodePesanan:           kode,
		KodeTransaksi:         "TRX-1",
		WaktuPesananTerbuat:   time.Now(),
		StatusPesananTerakhir: "baru",
		BiayaLainnya:          1.0,
		BiayaMitraJakmall:     0.5,
		TotalTransaksi:        10.5,
		DibuatOleh:            "user",
		JenisChannel:          "online",
		NamaToko:              "TestShop",
		KodeInvoiceChannel:    "INV-1",
		GudangPengiriman:      "gudang",
		JenisEkspedisi:        "kurir",
		Cashless:              "Ya",
		NomorResi:             "RESI1",
		WaktuPengiriman:       time.Now(),
		Provinsi:              "Jawa",
		Kota:                  "Bandung",
	}
	if err := repo.InsertDropshipPurchase(ctx, ds); err != nil {
		t.Fatalf("InsertDropshipPurchase failed: %v", err)
	}
	t.Log("InsertDropshipPurchase succeeded")

	fetched, err := repo.GetDropshipPurchaseByID(ctx, kode)
	if err != nil {
		t.Fatalf("GetDropshipPurchaseByID failed: %v", err)
	}
	if fetched.KodePesanan != kode {
		t.Errorf("Expected KodePesanan %s, got %s", kode, fetched.KodePesanan)
	}
	t.Log("GetDropshipPurchaseByID succeeded")

	// Cleanup
	tempCleanupDropship(t, kode)
}
