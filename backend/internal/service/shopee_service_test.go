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
	headers := []string{"No.", "No. Pesanan", "No. Pengajuan", "Username (Pembeli)", "Waktu Pesanan Dibuat", "Metode pembayaran pembeli", "Tanggal Dana Dilepaskan", "Harga Asli Produk", "Total Diskon Produk", "Jumlah Pengembalian Dana ke Pembeli", "Komisi Shopee", "Biaya Admin Shopee", "Biaya Layanan", "Biaya Layanan Ekstra", "Biaya Penyedia Pembayaran", "Asuransi", "Total Biaya Transaksi", "Biaya Pengiriman", "Total Diskon Pengiriman", "Promo Gratis Ongkir Shopee", "Promo Gratis Ongkir dari Penjual", "Promo Diskon Shopee", "Promo Diskon Penjual", "Cashback Shopee", "Cashback Penjual", "Koin Shopee", "Potongan Lainnya", "Total Penerimaan", "Kompensasi", "Promo Gratis Ongkir dari Penjual", "Jasa Kirim", "Nama Kurir", "Pengembalian Dana ke Pembeli", "Pro-rata Koin yang Ditukarkan untuk Pengembalian Barang", "Pro-rata Voucher Shopee untuk Pengembalian Barang", "Pro-rated Bank Payment Channel Promotion for returns", "Pro-rated Shopee Payment Channel Promotion for returns"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 5)
		f.SetCellValue("Data", cell, h)
	}
	data := []interface{}{1, "SO-1", "NG-1", "user", "2025-01-01", "COD", "2025-01-02", 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, "jne", "kurir", 1, 1, 1, 1, 1}
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
