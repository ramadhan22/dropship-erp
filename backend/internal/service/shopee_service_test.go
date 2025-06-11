// File: backend/internal/service/shopee_service_test.go

package service

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// fakeShopeeRepo captures calls to InsertShopeeOrder.
type fakeShopeeRepo struct {
	inserted []*models.ShopeeSettledOrder
	errOn    string // if OrderID equals this, return error
}

func (f *fakeShopeeRepo) InsertShopeeOrder(ctx context.Context, o *models.ShopeeSettledOrder) error {
	if o.OrderID == f.errOn {
		return errors.New("forced error")
	}
	f.inserted = append(f.inserted, o)
	return nil
}

func TestImportSettledOrdersCSV(t *testing.T) {
	// Create a fake CSV in memory (represent it as []byte)
	csvContent := []byte(`order_id,net_income,service_fee,campaign_fee,credit_card_fee,shipping_subsidy,tax_import_fee,settled_date,purchase_id,seller_username
SO-001,30.00,1.00,0.00,0.20,0.00,0.00,2025-05-15,DP-123,MyShop
`)
	// Write to a temp file
	tmp := t.TempDir() + "/shp.csv"
	if err := os.WriteFile(tmp, csvContent, 0644); err != nil {
		t.Fatal(err)
	}

	// Setup fake repos
	fakeS := &fakeShopeeRepo{}
	svc := NewShopeeService(fakeS)

	err := svc.ImportSettledOrdersCSV(context.Background(), tmp)
	if err != nil {
		t.Fatalf("ImportSettledOrdersCSV failed: %v", err)
	}

	// Verify Shopee insertion
	if len(fakeS.inserted) != 1 {
		t.Fatalf("expected 1 Shopee insert, got %d", len(fakeS.inserted))
	}
	ins := fakeS.inserted[0]
	if ins.OrderID != "SO-001" || ins.NetIncome != 30.00 {
		t.Errorf("unexpected inserted Shopee: %+v", ins)
	}

}
