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
