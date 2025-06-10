// File: backend/internal/service/dropship_service_test.go

package service

import (
	"context"
	"encoding/csv"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// fakeDropshipRepo captures calls to InsertDropshipPurchase.
type fakeDropshipRepo struct {
	inserted []*models.DropshipPurchase
	errOn    string // if PurchaseID equals this, return error
}

func (f *fakeDropshipRepo) InsertDropshipPurchase(ctx context.Context, p *models.DropshipPurchase) error {
	if p.PurchaseID == f.errOn {
		return os.ErrInvalid
	}
	f.inserted = append(f.inserted, p)
	return nil
}

func TestImportFromCSV_Success(t *testing.T) {
	// Create a temporary CSV file
	tmp, err := ioutil.TempFile("", "dspurchases-*.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())

	// Prepare CSV content
	w := csv.NewWriter(tmp)
	headers := []string{
		"seller_username", "purchase_id", "order_id", "sku", "qty",
		"purchase_price", "purchase_fee", "status", "purchase_date", "supplier_name",
	}
	w.Write(headers)
	row := []string{
		"MyShop", "PS-123", "", "SKU1", "2",
		"15.75", "0.50", "completed", "2025-05-10", "SupplierA",
	}
	w.Write(row)
	w.Flush()
	tmp.Close()

	// Use fake repo
	fake := &fakeDropshipRepo{}
	svc := NewDropshipService(fake)

	ctx := context.Background()
	err = svc.ImportFromCSV(ctx, tmp.Name())
	if err != nil {
		t.Fatalf("ImportFromCSV error: %v", err)
	}

	// Expect exactly one inserted purchase
	if len(fake.inserted) != 1 {
		t.Fatalf("expected 1 insert, got %d", len(fake.inserted))
	}
	inserted := fake.inserted[0]
	if inserted.PurchaseID != "PS-123" {
		t.Errorf("expected PurchaseID 'PS-123', got '%s'", inserted.PurchaseID)
	}
	if inserted.SKU != "SKU1" || inserted.Quantity != 2 {
		t.Errorf("unexpected SKU/Quantity: %s/%d", inserted.SKU, inserted.Quantity)
	}
	// Check purchase_date was parsed
	if !inserted.PurchaseDate.Equal(time.Date(2025, 5, 10, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("unexpected PurchaseDate: %v", inserted.PurchaseDate)
	}
}

func TestImportFromCSV_ParseError(t *testing.T) {
	// CSV with invalid qty (not an integer)
	tmp, err := ioutil.TempFile("", "badqty-*.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())

	w := csv.NewWriter(tmp)
	w.Write([]string{
		"seller_username", "purchase_id", "order_id", "sku", "qty",
		"purchase_price", "purchase_fee", "status", "purchase_date", "supplier_name",
	})
	// qty = "two" (invalid)
	w.Write([]string{"MyShop", "PS-456", "", "SKU2", "two", "10.00", "0.25", "completed", "2025-05-11", ""})
	w.Flush()
	tmp.Close()

	fake := &fakeDropshipRepo{}
	svc := NewDropshipService(fake)
	err = svc.ImportFromCSV(context.Background(), tmp.Name())
	if err == nil {
		t.Fatal("expected parse error, got nil")
	}
	// The fake repo should not have been called
	if len(fake.inserted) != 0 {
		t.Errorf("expected no inserts, got %d", len(fake.inserted))
	}
}
