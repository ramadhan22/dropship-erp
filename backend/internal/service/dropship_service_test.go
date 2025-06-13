// File: backend/internal/service/dropship_service_test.go

package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"os"
	"testing"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// fakeDropshipRepo captures calls to InsertDropshipPurchase.
type fakeDropshipRepo struct {
        insertedHeader []*models.DropshipPurchase
        insertedDetail []*models.DropshipPurchaseDetail
        errOn          string
        existing       map[string]bool
}

func (f *fakeDropshipRepo) InsertDropshipPurchase(ctx context.Context, p *models.DropshipPurchase) error {
	if p.KodePesanan == f.errOn {
		return os.ErrInvalid
	}
	f.insertedHeader = append(f.insertedHeader, p)
	return nil
}

func (f *fakeDropshipRepo) InsertDropshipPurchaseDetail(ctx context.Context, d *models.DropshipPurchaseDetail) error {
	f.insertedDetail = append(f.insertedDetail, d)
	return nil
}

func (f *fakeDropshipRepo) ExistsDropshipPurchase(ctx context.Context, kode string) (bool, error) {
	return f.existing[kode], nil
}

func (f *fakeDropshipRepo) ListDropshipPurchases(ctx context.Context, channel, store, date, month, year string, limit, offset int) ([]models.DropshipPurchase, int, error) {
        return nil, 0, nil
}

func (f *fakeDropshipRepo) GetDropshipPurchaseByID(ctx context.Context, kode string) (*models.DropshipPurchase, error) {
        return nil, nil
}

func (f *fakeDropshipRepo) ListDropshipPurchaseDetails(ctx context.Context, kode string) ([]models.DropshipPurchaseDetail, error) {
        return nil, nil
}

func (f *fakeDropshipRepo) SumDropshipPurchases(ctx context.Context, channel, store, date, month, year string) (float64, error) {
        return 0, nil
}

func TestImportFromCSV_Success(t *testing.T) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	headers := []string{"No", "waktu", "status", "kode", "trx", "sku", "nama", "harga", "qty", "total_harga", "biaya_lain", "biaya_mitra", "total_transaksi", "harga_ch", "total_harga_ch", "potensi", "dibuat", "channel", "toko", "invoice", "gudang", "ekspedisi", "cashless", "resi", "waktu_kirim", "provinsi", "kota"}
	w.Write(headers)
	row := []string{"1", "01 January 2025, 10:00:00", "selesai", "PS-123", "TRX1", "SKU1", "ProdukA", "15.75", "2", "31.50", "1", "0.5", "33.0", "15.75", "31.50", "2.0", "user", "online", "MyShop", "INV1", "GudangA", "JNE", "Ya", "RESI1", "02 January 2025, 10:00:00", "Jawa", "Bandung"}
	w.Write(row)
	w.Flush()

	// Use fake repo
	fake := &fakeDropshipRepo{}
	svc := NewDropshipService(fake)

	ctx := context.Background()
	count, err := svc.ImportFromCSV(ctx, &buf)
	if err != nil {
		t.Fatalf("ImportFromCSV error: %v", err)
	}
	if count != 1 {
		t.Errorf("expected count 1, got %d", count)
	}

	if len(fake.insertedHeader) != 1 || len(fake.insertedDetail) != 1 {
		t.Fatalf("expected 1 header and 1 detail insert, got %d/%d", len(fake.insertedHeader), len(fake.insertedDetail))
	}
	inserted := fake.insertedHeader[0]
	if inserted.KodePesanan != "PS-123" {
		t.Errorf("expected KodePesanan 'PS-123', got '%s'", inserted.KodePesanan)
	}
	d := fake.insertedDetail[0]
	if d.SKU != "SKU1" || d.Qty != 2 {
		t.Errorf("unexpected SKU/Qty: %s/%d", d.SKU, d.Qty)
	}
}

func TestImportFromCSV_ParseError(t *testing.T) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	w.Write([]string{"No", "waktu", "status", "kode", "trx", "sku", "nama", "harga", "qty", "total_harga", "biaya_lain", "biaya_mitra", "total_transaksi", "harga_ch", "total_harga_ch", "potensi", "dibuat", "channel", "toko", "invoice", "gudang", "ekspedisi", "cashless", "resi", "waktu_kirim", "provinsi", "kota"})
	w.Write([]string{"1", "01 January 2025, 10:00:00", "selesai", "PS-456", "TRX1", "SKU2", "ProdukB", "15.00", "two", "30", "1", "0.5", "31.5", "15", "30", "2", "user", "online", "Shop", "INV", "G", "JNE", "Ya", "RESI", "02 January 2025, 10:00:00", "Jawa", "Bandung"})
	w.Flush()

	fake := &fakeDropshipRepo{}
	svc := NewDropshipService(fake)
	count, err := svc.ImportFromCSV(context.Background(), &buf)
	if err == nil {
		t.Fatal("expected parse error, got nil")
	}
	if count != 0 {
		t.Errorf("expected count 0, got %d", count)
	}
	// The fake repo should not have been called
	if len(fake.insertedHeader) != 0 {
		t.Errorf("expected no inserts, got %d", len(fake.insertedHeader))
	}
}

func TestImportFromCSV_SkipExisting(t *testing.T) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	headers := []string{"No", "waktu", "status", "kode", "trx", "sku", "nama", "harga", "qty", "total_harga", "biaya_lain", "biaya_mitra", "total_transaksi", "harga_ch", "total_harga_ch", "potensi", "dibuat", "channel", "toko", "invoice", "gudang", "ekspedisi", "cashless", "resi", "waktu_kirim", "provinsi", "kota"}
	w.Write(headers)
	row := []string{"1", "01 January 2025, 10:00:00", "selesai", "PS-EXIST", "TRX1", "SKU1", "ProdukA", "15.75", "2", "31.50", "1", "0.5", "33.0", "15.75", "31.50", "2.0", "user", "online", "MyShop", "INV1", "GudangA", "JNE", "Ya", "RESI1", "02 January 2025, 10:00:00", "Jawa", "Bandung"}
	w.Write(row)
	w.Flush()

	fake := &fakeDropshipRepo{existing: map[string]bool{"PS-EXIST": true}}
	svc := NewDropshipService(fake)
	count, err := svc.ImportFromCSV(context.Background(), &buf)
	if err != nil {
		t.Fatalf("ImportFromCSV error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected count 0, got %d", count)
	}
	if len(fake.insertedHeader) != 0 || len(fake.insertedDetail) != 0 {
		t.Fatalf("expected no inserts, got %d/%d", len(fake.insertedHeader), len(fake.insertedDetail))
	}
}
