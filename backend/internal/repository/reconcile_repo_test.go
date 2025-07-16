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
	jrepo := NewJournalRepo(testDB)

	kode1 := "CAND-" + time.Now().Format("150405")
	dp1 := &models.DropshipPurchase{KodePesanan: kode1, KodeInvoiceChannel: kode1, NamaToko: "ShopA", StatusPesananTerakhir: "diproses", WaktuPesananTerbuat: time.Now()}
	_ = dropRepo.InsertDropshipPurchase(ctx, dp1)
	ss1 := &models.ShopeeSettled{NamaToko: "ShopA", NoPesanan: kode1, WaktuPesananDibuat: time.Now(), TanggalDanaDilepaskan: time.Now()}
	_ = shopRepo.InsertShopeeSettled(ctx, ss1)
	je := &models.JournalEntry{
		EntryDate:    time.Now(),
		Description:  ptrString("pending"),
		SourceType:   "pending_sales",
		SourceID:     kode1,
		ShopUsername: "ShopA",
		Store:        "ShopA",
		CreatedAt:    time.Now(),
	}
	_, _ = jrepo.CreateJournalEntry(ctx, je)

	kode2 := "CAND-" + time.Now().Format("150405") + "b"
	dp2 := &models.DropshipPurchase{KodePesanan: kode2, KodeInvoiceChannel: kode2, NamaToko: "ShopA", StatusPesananTerakhir: "diproses", WaktuPesananTerbuat: time.Now()}
	_ = dropRepo.InsertDropshipPurchase(ctx, dp2)

	je2 := &models.JournalEntry{
		EntryDate:    time.Now(),
		Description:  ptrString("pending"),
		SourceType:   "pending_sales",
		SourceID:     kode2,
		ShopUsername: "ShopA",
		Store:        "ShopA",
		CreatedAt:    time.Now(),
	}
	_, _ = jrepo.CreateJournalEntry(ctx, je2)

	escrow := &models.JournalEntry{
		EntryDate:    time.Now(),
		Description:  ptrString("escrow"),
		SourceType:   "shopee_escrow",
		SourceID:     kode2,
		ShopUsername: "ShopA",
		Store:        "ShopA",
		CreatedAt:    time.Now(),
	}
	_, _ = jrepo.CreateJournalEntry(ctx, escrow)

	list, total, err := recRepo.ListCandidates(ctx, "ShopA", "", "", "", "", 10, 0)
	if err != nil {
		t.Fatalf("ListCandidates error: %v", err)
	}
	if total != 1 || len(list) != 1 {
		t.Errorf("expected 1 candidate, got %d total %d", len(list), total)
	}

	list2, total2, err := recRepo.ListCandidates(ctx, "", kode1, "", "", "", 10, 0)
	if err != nil {
		t.Fatalf("ListCandidates by order error: %v", err)
	}
	if len(list2) != 1 || total2 != 1 {
		t.Errorf("expected 1 candidate, got %d total %d", len(list2), total2)
	}

	list3, total3, err := recRepo.ListCandidates(ctx, "", "", "diproses", "", "", 10, 0)
	if err != nil {
		t.Fatalf("ListCandidates by status error: %v", err)
	}
	if len(list3) == 0 || total3 == 0 {
		t.Errorf("expected candidates by status, got %d total %d", len(list3), total3)
	}

	// cleanup
	testDB.ExecContext(ctx, "DELETE FROM shopee_settled WHERE no_pesanan=$1", kode1)
	testDB.ExecContext(ctx, "DELETE FROM dropship_purchases WHERE kode_pesanan IN ($1,$2)", kode1, kode2)
	testDB.ExecContext(ctx, "DELETE FROM journal_entries WHERE source_id IN ($1,$2)", kode1, kode2)
}
