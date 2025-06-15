package service

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
)

// fakeShopeeRepo captures inserted rows.
type fakeShopeeRepo struct {
	count      int
	fail       bool
	existing   map[string]bool
	affExpense float64
}

type fakeJournalRepoS struct {
	entries []*models.JournalEntry
	lines   []*models.JournalLine
	nextID  int64
}

func (f *fakeJournalRepoS) CreateJournalEntry(ctx context.Context, e *models.JournalEntry) (int64, error) {
	f.nextID++
	e.JournalID = f.nextID
	f.entries = append(f.entries, e)
	return f.nextID, nil
}
func (f *fakeJournalRepoS) InsertJournalLine(ctx context.Context, l *models.JournalLine) error {
	f.lines = append(f.lines, l)
	return nil
}
func (f *fakeJournalRepoS) GetJournalEntryBySource(ctx context.Context, sourceType, sourceID string) (*models.JournalEntry, error) {
	for _, e := range f.entries {
		if e.SourceType == sourceType && e.SourceID == sourceID {
			return e, nil
		}
	}
	return nil, nil
}
func (f *fakeJournalRepoS) GetLinesByJournalID(ctx context.Context, id int64) ([]repository.JournalLineDetail, error) {
	res := []repository.JournalLineDetail{}
	for _, l := range f.lines {
		if l.JournalID == id {
			res = append(res, repository.JournalLineDetail{LineID: l.LineID, JournalID: l.JournalID, AccountID: l.AccountID, IsDebit: l.IsDebit, Amount: l.Amount})
		}
	}
	return res, nil
}
func (f *fakeJournalRepoS) UpdateJournalLineAmount(ctx context.Context, lineID int64, amount float64) error {
	for _, l := range f.lines {
		if l.LineID == lineID {
			l.Amount = amount
		}
	}
	return nil
}

func (f *fakeShopeeRepo) InsertShopeeSettled(ctx context.Context, s *models.ShopeeSettled) error {
	if f.fail {
		return errors.New("fail")
	}
	f.count++
	return nil
}

func (f *fakeShopeeRepo) InsertShopeeAffiliateSale(ctx context.Context, s *models.ShopeeAffiliateSale) error {
	if f.fail {
		return errors.New("fail")
	}
	f.count++
	return nil
}

func (f *fakeShopeeRepo) ExistsShopeeSettled(ctx context.Context, noPesanan string) (bool, error) {
	if f.existing == nil {
		return false, nil
	}
	return f.existing[noPesanan], nil
}

func (f *fakeShopeeRepo) ExistsShopeeAffiliateSale(ctx context.Context, orderID, productCode string) (bool, error) {
	if f.existing == nil {
		return false, nil
	}
	return f.existing[orderID], nil
}

func (f *fakeShopeeRepo) ListShopeeSettled(ctx context.Context, channel, store, from, to, orderNo, sortBy, dir string, limit, offset int) ([]models.ShopeeSettled, int, error) {
	return nil, 0, nil
}

func (f *fakeShopeeRepo) SumShopeeSettled(ctx context.Context, channel, store, from, to string) (*models.ShopeeSummary, error) {
	return &models.ShopeeSummary{}, nil
}

func (f *fakeShopeeRepo) ListShopeeAffiliateSales(ctx context.Context, date, month, year string, limit, offset int) ([]models.ShopeeAffiliateSale, int, error) {
	return nil, 0, nil
}

func (f *fakeShopeeRepo) SumShopeeAffiliateSales(ctx context.Context, date, month, year string) (*models.ShopeeAffiliateSummary, error) {
	return &models.ShopeeAffiliateSummary{}, nil
}

func (f *fakeShopeeRepo) GetAffiliateExpenseByOrder(ctx context.Context, kodePesanan string) (float64, error) {
	return f.affExpense, nil
}

func TestImportSettledOrdersXLSX(t *testing.T) {
	f := excelize.NewFile()
	sheet, _ := f.NewSheet("Data")
	headers := append([]string{"No."}, expectedHeaders...)
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 5)
		f.SetCellValue("Data", cell, h)
	}
	data := []interface{}{
		1, "SO-1", "NG-1", "user", "2025-01-01", "COD", "2025-01-02",
		1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
		1, // total penghasilan
		1, // kode voucher
		1, // kompensasi
		1, // promo gratis ongkir dari penjual
		"jne", "kurir", "",
		1, 1, 1, 1, 1,
	}
	for i, v := range data {
		cell, _ := excelize.CoordinatesToCellName(i+1, 6)
		f.SetCellValue("Data", cell, v)
	}
	f.SetActiveSheet(sheet)
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatal(err)
	}

	repo := &fakeShopeeRepo{}
	svc := NewShopeeService(nil, repo, nil, nil)
	inserted, err := svc.ImportSettledOrdersXLSX(context.Background(), bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("import error: %v", err)
	}
	if inserted != 1 || repo.count != 1 {
		t.Fatalf("expected 1 insert, got svc %d repo %d", inserted, repo.count)
	}
}

func TestImportSettledOrdersXLSX_HeaderMismatch(t *testing.T) {
	f := excelize.NewFile()
	sheet, _ := f.NewSheet("Data")
	// Write an invalid header on row 5
	f.SetCellValue("Data", "B5", "WRONG")
	f.SetActiveSheet(sheet)
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatal(err)
	}

	repo := &fakeShopeeRepo{}
	svc := NewShopeeService(nil, repo, nil, nil)
	_, err := svc.ImportSettledOrdersXLSX(context.Background(), bytes.NewReader(buf.Bytes()))
	if err == nil {
		t.Fatalf("expected error due to header mismatch")
	}
}

func TestImportSettledOrdersXLSX_SkipDuplicates(t *testing.T) {
	f := excelize.NewFile()
	sheet, _ := f.NewSheet("Data")
	headers := append([]string{"No."}, expectedHeaders...)
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 5)
		f.SetCellValue("Data", cell, h)
	}
	data := []interface{}{
		1, "SO-1", "NG-1", "user", "2025-01-01", "COD", "2025-01-02",
		1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
		1,
		1,
		1,
		1,
		"jne", "kurir", "",
		1, 1, 1, 1, 1,
	}
	for i, v := range data {
		cell, _ := excelize.CoordinatesToCellName(i+1, 6)
		f.SetCellValue("Data", cell, v)
	}
	f.SetActiveSheet(sheet)
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatal(err)
	}

	repo := &fakeShopeeRepo{existing: map[string]bool{"SO-1": true}}
	svc := NewShopeeService(nil, repo, nil, nil)
	inserted, err := svc.ImportSettledOrdersXLSX(context.Background(), bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("import error: %v", err)
	}
	if inserted != 0 || repo.count != 0 {
		t.Fatalf("expected 0 insert, got svc %d repo %d", inserted, repo.count)
	}
}

func TestImportAffiliateCSV(t *testing.T) {
	csvData := "Kode Pesanan,Status Pesanan,Status Terverifikasi,Waktu Pesanan,Waktu Pesanan Selesai,Waktu Pesanan Terverifikasi,Kode Produk,Nama Produk,ID Model,L1 Kategori Global,L2 Kategori Global,L3 Kategori Global,Kode Promo,Harga(Rp),Jumlah,Nama Affiliate,Username Affiliate,MCN Terhubung,ID Komisi Pesanan,Partner Promo,Jenis Promo,Nilai Pembelian(Rp),Jumlah Pengembalian(Rp),Tipe Pesanan,Estimasi Komisi per Produk(Rp),Estimasi Komisi Affiliate per Produk(Rp),Persentase Komisi Affiliate per Produk,Estimasi Komisi MCN per Produk(Rp),Persentase Komisi MCN per Produk,Estimasi Komisi per Pesanan(Rp),Estimasi Komisi Affiliate per Pesanan(Rp),Estimasi Komisi MCN per Pesanan(Rp),Catatan Produk,Platform,Tingkat Komisi,Pengeluaran(Rp),Status Pemotongan,Metode Pemotongan,Waktu Pemotongan\n" +
		"SO1,Selesai,Sah,2025-06-01 10:00:00,,,P1,Produk,ID1,Cat1,Cat2,Cat3,,1000,1,Aff,affuser,,1,,Promo,1000,0,Langsung,10,10,10%,0,0%,10,10,0,,IG,10%,0,,,"
	repo := &fakeShopeeRepo{}
	svc := NewShopeeService(nil, repo, nil, nil)
	inserted, err := svc.ImportAffiliateCSV(context.Background(), strings.NewReader(csvData))
	if err != nil {
		t.Fatalf("import error: %v", err)
	}
	if inserted != 1 || repo.count != 1 {
		t.Fatalf("expected 1 insert, got svc %d repo %d", inserted, repo.count)
	}
}
