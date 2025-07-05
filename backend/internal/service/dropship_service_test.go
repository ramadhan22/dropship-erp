// File: backend/internal/service/dropship_service_test.go

package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"os"
	"testing"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
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

func (f *fakeDropshipRepo) ListDropshipPurchases(ctx context.Context, channel, store, from, to, orderNo, sortBy, dir string, limit, offset int) ([]models.DropshipPurchase, int, error) {
	return nil, 0, nil
}

func (f *fakeDropshipRepo) GetDropshipPurchaseByID(ctx context.Context, kode string) (*models.DropshipPurchase, error) {
	return nil, nil
}

func (f *fakeDropshipRepo) ListDropshipPurchaseDetails(ctx context.Context, kode string) ([]models.DropshipPurchaseDetail, error) {
	return nil, nil
}

func (f *fakeDropshipRepo) SumDropshipPurchases(ctx context.Context, channel, store, from, to string) (float64, error) {
	return 0, nil
}

func (f *fakeDropshipRepo) TopProducts(ctx context.Context, channel, store, from, to string, limit int) ([]models.ProductSales, error) {
	return nil, nil
}

func (f *fakeDropshipRepo) DailyTotals(ctx context.Context, channel, store, from, to string) ([]repository.DailyPurchaseTotal, error) {
	return nil, nil
}

func (f *fakeDropshipRepo) MonthlyTotals(ctx context.Context, channel, store, from, to string) ([]repository.MonthlyPurchaseTotal, error) {
	return nil, nil
}

func (f *fakeDropshipRepo) CancelledSummary(ctx context.Context, channel, store, from, to string) (repository.CancelledSummary, error) {
	return repository.CancelledSummary{}, nil
}

type fakeJournalRepoDrop struct {
	entries []*models.JournalEntry
	lines   []*models.JournalLine
	nextID  int64
}

func (f *fakeJournalRepoDrop) CreateJournalEntry(ctx context.Context, e *models.JournalEntry) (int64, error) {
	f.nextID++
	e.JournalID = f.nextID
	f.entries = append(f.entries, e)
	return f.nextID, nil
}

func (f *fakeJournalRepoDrop) InsertJournalLine(ctx context.Context, l *models.JournalLine) error {
	f.lines = append(f.lines, l)
	return nil
}
func (f *fakeJournalRepoDrop) GetJournalEntryBySource(ctx context.Context, sourceType, sourceID string) (*models.JournalEntry, error) {
	for _, e := range f.entries {
		if e.SourceType == sourceType && e.SourceID == sourceID {
			return e, nil
		}
	}
	return nil, nil
}
func (f *fakeJournalRepoDrop) GetLinesByJournalID(ctx context.Context, id int64) ([]repository.JournalLineDetail, error) {
	res := []repository.JournalLineDetail{}
	for _, l := range f.lines {
		if l.JournalID == id {
			res = append(res, repository.JournalLineDetail{LineID: l.LineID, JournalID: l.JournalID, AccountID: l.AccountID, IsDebit: l.IsDebit, Amount: l.Amount})
		}
	}
	return res, nil
}
func (f *fakeJournalRepoDrop) UpdateJournalLineAmount(ctx context.Context, lineID int64, amount float64) error {
	for _, l := range f.lines {
		if l.LineID == lineID {
			l.Amount = amount
		}
	}
	return nil
}

func TestImportFromCSV_Success(t *testing.T) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	headers := []string{"No", "waktu", "status", "kode", "trx", "sku", "nama", "harga", "qty", "total_harga", "biaya_lain", "biaya_mitra", "total_transaksi", "harga_ch", "total_harga_ch", "potensi", "dibuat", "channel", "toko", "invoice", "gudang", "ekspedisi", "cashless", "resi", "waktu_kirim", "provinsi", "kota"}
	w.Write(headers)
	row := []string{"1", "01 January 2025, 10:00:00", "selesai", "PS-123", "TRX1", "SKU1", "ProdukA", "15.75", "2", "31.50", "1", "0.5", "33.0", "15.75", "31.50", "2.0", "user", "online", "MyShop", "INV1", "GudangA", "JNE", "Ya", "RESI1", "02 January 2025, 10:00:00", "Jawa", "Bandung"}
	w.Write(row)
	w.Flush()

	// Use fake repos
	fake := &fakeDropshipRepo{}
	jfake := &fakeJournalRepoDrop{}
	svc := NewDropshipService(nil, fake, jfake)

	ctx := context.Background()
	count, err := svc.ImportFromCSV(ctx, &buf, "")
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

	if len(jfake.entries) != 1 || len(jfake.lines) != 5 {
		t.Fatalf("expected 1 journal entry and 5 lines, got %d/%d", len(jfake.entries), len(jfake.lines))
	}
}

func TestImportFromCSV_ParseError(t *testing.T) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	w.Write([]string{"No", "waktu", "status", "kode", "trx", "sku", "nama", "harga", "qty", "total_harga", "biaya_lain", "biaya_mitra", "total_transaksi", "harga_ch", "total_harga_ch", "potensi", "dibuat", "channel", "toko", "invoice", "gudang", "ekspedisi", "cashless", "resi", "waktu_kirim", "provinsi", "kota"})
	w.Write([]string{"1", "01 January 2025, 10:00:00", "selesai", "PS-456", "TRX1", "SKU2", "ProdukB", "15.00", "two", "30", "1", "0.5", "31.5", "15", "30", "2", "user", "online", "Shop", "INV", "G", "JNE", "Ya", "RESI", "02 January 2025, 10:00:00", "Jawa", "Bandung"})
	w.Flush()

	fake := &fakeDropshipRepo{}
	svc := NewDropshipService(nil, fake, nil)
	count, err := svc.ImportFromCSV(context.Background(), &buf, "")
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
	svc := NewDropshipService(nil, fake, nil)
	count, err := svc.ImportFromCSV(context.Background(), &buf, "")
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

func TestImportFromCSV_JournalSumsProducts(t *testing.T) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	headers := []string{"No", "waktu", "status", "kode", "trx", "sku", "nama", "harga", "qty", "total_harga", "biaya_lain", "biaya_mitra", "total_transaksi", "harga_ch", "total_harga_ch", "potensi", "dibuat", "channel", "toko", "invoice", "gudang", "ekspedisi", "cashless", "resi", "waktu_kirim", "provinsi", "kota"}
	w.Write(headers)
	row1 := []string{"1", "01 January 2025, 10:00:00", "selesai", "PS-200", "TRX1", "SKU1", "ProdukA", "15.75", "2", "31.50", "1", "0.5", "33.0", "15.75", "31.50", "2.0", "user", "online", "MyShop", "INV1", "GudangA", "JNE", "Ya", "RESI1", "02 January 2025, 10:00:00", "Jawa", "Bandung"}
	row2 := []string{"2", "01 January 2025, 10:00:00", "selesai", "PS-200", "TRX1", "SKU2", "ProdukB", "20.00", "1", "20.00", "1", "0.5", "21.0", "20.00", "20.00", "1.0", "user", "online", "MyShop", "INV1", "GudangA", "JNE", "Ya", "RESI1", "02 January 2025, 10:00:00", "Jawa", "Bandung"}
	w.Write(row1)
	w.Write(row2)
	w.Flush()

	fake := &fakeDropshipRepo{}
	jfake := &fakeJournalRepoDrop{}
	svc := NewDropshipService(nil, fake, jfake)

	count, err := svc.ImportFromCSV(context.Background(), &buf, "")
	if err != nil {
		t.Fatalf("ImportFromCSV error: %v", err)
	}
	if count != 2 {
		t.Errorf("expected count 2, got %d", count)
	}
	if len(jfake.entries) != 1 {
		t.Fatalf("expected 1 journal entry, got %d", len(jfake.entries))
	}

	var hpp, sales float64
	for _, l := range jfake.lines {
		if l.AccountID == 5001 {
			hpp = l.Amount
		}
		if l.AccountID == 4001 {
			sales = l.Amount
		}
	}
	if hpp != 51.5 {
		t.Errorf("expected HPP 51.5, got %.2f", hpp)
	}
	if sales != 51.5 {
		t.Errorf("expected sales 51.5, got %.2f", sales)
	}
}

func TestImportFromCSV_ChannelFilter(t *testing.T) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	headers := []string{"No", "waktu", "status", "kode", "trx", "sku", "nama", "harga", "qty", "total_harga", "biaya_lain", "biaya_mitra", "total_transaksi", "harga_ch", "total_harga_ch", "potensi", "dibuat", "channel", "toko", "invoice", "gudang", "ekspedisi", "cashless", "resi", "waktu_kirim", "provinsi", "kota"}
	w.Write(headers)
	row := []string{"1", "01 January 2025, 10:00:00", "selesai", "PS-1234", "TRX1", "SKU1", "ProdukA", "15.75", "1", "15.75", "1", "0.5", "17.25", "15.75", "15.75", "1.0", "user", "Shopee", "Shop1", "INV", "Gudang", "JNE", "Ya", "RESI", "02 January 2025, 10:00:00", "Jawa", "Bandung"}
	w.Write(row)
	w.Flush()

	fake := &fakeDropshipRepo{}
	svc := NewDropshipService(nil, fake, nil)

	count, err := svc.ImportFromCSV(context.Background(), &buf, "Tokopedia")
	if err != nil {
		t.Fatalf("ImportFromCSV error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected count 0, got %d", count)
	}
	if len(fake.insertedHeader) != 0 {
		t.Errorf("expected no inserts, got %d", len(fake.insertedHeader))
	}
}
