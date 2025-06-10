// File: backend/internal/repository/shopee_repo_test.go

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// tempCleanupShopee deletes any row with the given orderID.
func tempCleanupShopee(t *testing.T, orderID string) {
	_, err := testDB.ExecContext(context.Background(),
		"DELETE FROM shopee_settled_orders WHERE order_id = $1", orderID)
	if err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}
}

func TestInsertAndGetShopeeOrder(t *testing.T) {
	ctx := context.Background()
	repo := NewShopeeRepo(testDB)

	orderID := "TEST-SP-" + time.Now().Format("20060102150405")
	so := &models.ShopeeSettledOrder{
		OrderID:         orderID,
		NetIncome:       20.00,
		ServiceFee:      1.00,
		CampaignFee:     0.00,
		CreditCardFee:   0.20,
		ShippingSubsidy: 0.00,
		TaxImportFee:    0.00,
		SettledDate:     time.Now(),
		SellerUsername:  "TestShop",
	}
	if err := repo.InsertShopeeOrder(ctx, so); err != nil {
		t.Fatalf("InsertShopeeOrder failed: %v", err)
	}
	t.Log("InsertShopeeOrder succeeded")

	fetched, err := repo.GetShopeeOrderByID(ctx, orderID)
	if err != nil {
		t.Fatalf("GetShopeeOrderByID failed: %v", err)
	}
	if fetched.OrderID != orderID {
		t.Errorf("Expected OrderID %s, got %s", orderID, fetched.OrderID)
	}
	t.Log("GetShopeeOrderByID succeeded")

	// Cleanup
	tempCleanupShopee(t, orderID)
}
