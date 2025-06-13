// File: backend/internal/repository/reconcile_repo_test.go

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// cleanupReconcile deletes a reconciled_transactions row by its ID.
func cleanupReconcile(t *testing.T, id int64) {
	_, err := testDB.ExecContext(context.Background(),
		"DELETE FROM reconciled_transactions WHERE id = $1", id)
	if err != nil {
		t.Fatalf("cleanupReconcile failed: %v", err)
	}
}

func TestInsertAndGetReconciledTransaction(t *testing.T) {
	ctx := context.Background()
	repo := NewReconcileRepo(testDB)

	rec := &models.ReconciledTransaction{
		ShopUsername: "TestShop",
		DropshipID:   ptrString("TEST-DS-0001"),
		ShopeeID:     ptrString("TEST-SP-0001"),
		Status:       "matched",
		MatchedAt:    time.Now(),
	}
	if err := repo.InsertReconciledTransaction(ctx, rec); err != nil {
		t.Fatalf("InsertReconciledTransaction failed: %v", err)
	}
	t.Log("InsertReconciledTransaction succeeded")

	period := time.Now().Format("2006-01")
	list, err := repo.GetReconciledTransactionsByShopAndPeriod(ctx, "TestShop", period)
	if err != nil {
		t.Fatalf("GetReconciledTransactionsByShopAndPeriod failed: %v", err)
	}
	found := false
	for _, r := range list {
		if r.DropshipID != nil && *r.DropshipID == "TEST-DS-0001" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected to find inserted ReconciledTransaction")
	}
	t.Log("GetReconciledTransactionsByShopAndPeriod succeeded")

	// Cleanup inserted rows
	for _, r := range list {
		if r.ShopUsername == "TestShop" {
			cleanupReconcile(t, r.ID)
		}
	}
}

func TestListCandidates(t *testing.T) {
	ctx := context.Background()
	recRepo := NewReconcileRepo(testDB)
	dropRepo := NewDropshipRepo(testDB)
	shopRepo := NewShopeeRepo(testDB)

	kode1 := "CAND-" + time.Now().Format("150405")
	dp1 := &models.DropshipPurchase{KodePesanan: kode1, NamaToko: "ShopA", StatusPesananTerakhir: "diproses", WaktuPesananTerbuat: time.Now()}
	_ = dropRepo.InsertDropshipPurchase(ctx, dp1)
	ss1 := &models.ShopeeSettled{NamaToko: "ShopA", NoPesanan: kode1, WaktuPesananDibuat: time.Now(), TanggalDanaDilepaskan: time.Now()}
	_ = shopRepo.InsertShopeeSettled(ctx, ss1)

	kode2 := "CAND-" + time.Now().Format("150405") + "b"
	dp2 := &models.DropshipPurchase{KodePesanan: kode2, NamaToko: "ShopA", StatusPesananTerakhir: "pesanan selesai", WaktuPesananTerbuat: time.Now()}
	_ = dropRepo.InsertDropshipPurchase(ctx, dp2)

	list, err := recRepo.ListCandidates(ctx, "ShopA")
	if err != nil {
		t.Fatalf("ListCandidates error: %v", err)
	}
	if len(list) < 2 {
		t.Errorf("expected at least 2 candidates, got %d", len(list))
	}

	// cleanup
	testDB.ExecContext(ctx, "DELETE FROM shopee_settled WHERE no_pesanan=$1", kode1)
	testDB.ExecContext(ctx, "DELETE FROM dropship_purchases WHERE kode_pesanan IN ($1,$2)", kode1, kode2)
}
