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
	count int
	fail  bool
}

func (f *fakeShopeeRepo) InsertShopeeSettled(ctx context.Context, s *models.ShopeeSettled) error {
	if f.fail {
		return errors.New("fail")
	}
	f.count++
	return nil
}

func TestImportSettledOrdersXLSX(t *testing.T) {
	f := excelize.NewFile()
	sheet, _ := f.NewSheet("Data")
	headers := []string{"No.", "No. Pesanan", "No. Pengajuan", "Username (Pembeli)", "Waktu Pesanan Dibuat", "Metode pembayaran pembeli", "Tanggal Dana Dilepaskan", "Harga Asli Produk", "Total Diskon Produk", "Jumlah Pengembalian Dana ke Pembeli", "Diskon Produk dari Shopee", "Diskon Voucher Ditanggung Penjual", "Cashback Koin yang Ditanggung Penjual", "Ongkir Dibayar Pembeli", "Diskon Ongkir Ditanggung Jasa Kirim", "Gratis Ongkir dari Shopee", "Ongkir yang Diteruskan oleh Shopee ke Jasa Kirim", "Ongkos Kirim Pengembalian Barang", "Pengembalian Biaya Kirim", "Biaya Komisi AMS", "Biaya Administrasi", "Biaya Layanan (termasuk PPN 11%)", "Premi", "Biaya Program", "Biaya Kartu Kredit", "Biaya Kampanye", "Bea Masuk, PPN & PPh", "Total Penghasilan", "Kode Voucher", "Kompensasi", "Promo Gratis Ongkir dari Penjual", "Jasa Kirim", "Nama Kurir", "Pengembalian Dana ke Pembeli", "Pro-rata Koin yang Ditukarkan untuk Pengembalian Barang", "Pro-rata Voucher Shopee untuk Pengembalian Barang", "Pro-rated Bank Payment Channel Promotion  for return refund Items", "Pro-rated Shopee Payment Channel Promotion  for return refund Items"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 6)
		f.SetCellValue("Data", cell, h)
	}
	data := []interface{}{1, "SO-1", "NG-1", "user", "2025-01-01", "COD", "2025-01-02", 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, "VCH", 1, 1, "jne", "kurir", 1, 1, 1, 1, 1}
	for i, v := range data {
		cell, _ := excelize.CoordinatesToCellName(i+1, 7)
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
	// Write an invalid header on row 6
	f.SetCellValue("Data", "B6", "WRONG")
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
