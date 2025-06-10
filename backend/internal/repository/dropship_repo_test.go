// File: backend/internal/repository/dropship_repo_test.go

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// tempCleanupDropship deletes any row with the given purchaseID.
func tempCleanupDropship(t *testing.T, purchaseID string) {
	_, err := testDB.ExecContext(context.Background(),
		"DELETE FROM dropship_purchases WHERE purchase_id = $1", purchaseID)
	if err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}
}

func TestInsertAndGetDropshipPurchase(t *testing.T) {
	ctx := context.Background()
	repo := NewDropshipRepo(testDB)

	purchaseID := "TEST-DS-" + time.Now().Format("20060102150405")
	ds := &models.DropshipPurchase{
		SellerUsername: "TestShop",
		PurchaseID:     purchaseID,
		SKU:            "ABC123",
		Quantity:       1,
		PurchasePrice:  10.00,
		PurchaseFee:    0.50,
		Status:         "completed",
		PurchaseDate:   time.Now(),
		SupplierName:   ptrString("TestSupplier"),
	}
	if err := repo.InsertDropshipPurchase(ctx, ds); err != nil {
		t.Fatalf("InsertDropshipPurchase failed: %v", err)
	}
	t.Log("InsertDropshipPurchase succeeded")

	fetched, err := repo.GetDropshipPurchaseByID(ctx, purchaseID)
	if err != nil {
		t.Fatalf("GetDropshipPurchaseByID failed: %v", err)
	}
	if fetched.PurchaseID != purchaseID {
		t.Errorf("Expected PurchaseID %s, got %s", purchaseID, fetched.PurchaseID)
	}
	t.Log("GetDropshipPurchaseByID succeeded")

	// Cleanup
	tempCleanupDropship(t, purchaseID)
}
