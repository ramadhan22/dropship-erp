package service

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/xuri/excelize/v2"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// fakeShopeeRepo captures inserted rows.
type fakeShopeeRepo struct {
	count    int
	fail     bool
	existing map[string]bool
}

func (f *fakeShopeeRepo) InsertShopeeSettled(ctx context.Context, s *models.ShopeeSettled) error {
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

func (f *fakeShopeeRepo) ListShopeeSettled(ctx context.Context, channel, store, date, month, year string, limit, offset int) ([]models.ShopeeSettled, int, error) {
	return nil, 0, nil
}

func (f *fakeShopeeRepo) SumShopeeSettled(ctx context.Context, channel, store, date, month, year string) (float64, error) {
	return 0, nil
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
	svc := NewShopeeService(repo)
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
	svc := NewShopeeService(repo)
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
	svc := NewShopeeService(repo)
	inserted, err := svc.ImportSettledOrdersXLSX(context.Background(), bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("import error: %v", err)
	}
	if inserted != 0 || repo.count != 0 {
		t.Fatalf("expected 0 insert, got svc %d repo %d", inserted, repo.count)
	}
}
