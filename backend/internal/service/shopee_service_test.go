package service

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/xuri/excelize/v2"

	"github.com/ramadhan22/dropship-erp/backend/internal/config"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
)

// fakeShopeeRepo captures inserted rows.
type fakeShopeeRepo struct {
	count             int
	fail              bool
	existingSettled   map[string]bool
	existingAffiliate map[string]bool
	deletedAffiliate  []string
	affExpense        float64
	order             *models.ShopeeSettled
	confirmed         []string
}

func (f *fakeShopeeRepo) MarkMismatch(ctx context.Context, orderSN string, mismatch bool) error {
	return nil
}

func (f *fakeShopeeRepo) ConfirmSettle(ctx context.Context, orderSN string) error {
	f.confirmed = append(f.confirmed, orderSN)
	return nil
}

func (f *fakeShopeeRepo) GetBySN(ctx context.Context, orderSN string) (*models.ShopeeSettled, error) {
	if f.order != nil {
		return f.order, nil
	}
	return &models.ShopeeSettled{NamaToko: "TOKO", NoPesanan: orderSN, HargaAsliProduk: 1}, nil
}

type fakeJournalRepoS struct {
	entries []*models.JournalEntry
	lines   []*models.JournalLine
	nextID  int64
}

type fakeDropRepoA struct {
	byInvoice map[string]*models.DropshipPurchase
	byTrans   map[string]*models.DropshipPurchase
	updated   map[string]string
}

func (f *fakeDropRepoA) GetDropshipPurchaseByInvoice(ctx context.Context, inv string) (*models.DropshipPurchase, error) {
	if dp, ok := f.byInvoice[inv]; ok {
		return dp, nil
	}
	return nil, nil
}

func (f *fakeDropRepoA) GetDropshipPurchaseByID(ctx context.Context, kode string) (*models.DropshipPurchase, error) {
	return nil, nil
}

func (f *fakeDropRepoA) GetDropshipPurchaseByTransaction(ctx context.Context, trx string) (*models.DropshipPurchase, error) {
	if f.byTrans == nil {
		return nil, nil
	}
	if dp, ok := f.byTrans[trx]; ok {
		return dp, nil
	}
	return nil, nil
}

func (f *fakeDropRepoA) SumDetailByInvoice(ctx context.Context, inv string) (float64, error) {
	if dp, ok := f.byInvoice[inv]; ok {
		return dp.TotalTransaksi, nil
	}
	return 0, nil
}

func (f *fakeDropRepoA) UpdateDropshipStatus(ctx context.Context, kode, status string) error {
	if f.updated == nil {
		f.updated = map[string]string{}
	}
	f.updated[kode] = status
	return nil
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

func (f *fakeJournalRepoS) DeleteJournalEntry(ctx context.Context, id int64) error {
	for i, e := range f.entries {
		if e.JournalID == id {
			f.entries = append(f.entries[:i], f.entries[i+1:]...)
			break
		}
	}
	filtered := []*models.JournalLine{}
	for _, l := range f.lines {
		if l.JournalID != id {
			filtered = append(filtered, l)
		}
	}
	f.lines = filtered
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
	if f.existingSettled == nil {
		return false, nil
	}
	return f.existingSettled[noPesanan], nil
}

func (f *fakeShopeeRepo) ExistsShopeeAffiliateSale(ctx context.Context, orderID, productCode, komisiID string) (bool, error) {
	if f.existingAffiliate == nil {
		return false, nil
	}
	return f.existingAffiliate[orderID+"|"+productCode+"|"+komisiID], nil
}

func (f *fakeShopeeRepo) DeleteShopeeAffiliateSale(ctx context.Context, orderID, productCode, komisiID string) error {
	f.deletedAffiliate = append(f.deletedAffiliate, orderID+"|"+productCode+"|"+komisiID)
	return nil
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

func (f *fakeShopeeRepo) ListSalesProfit(ctx context.Context, channel, store, from, to, orderNo, sortBy, dir string, limit, offset int) ([]models.SalesProfit, int, error) {
	return nil, 0, nil
}

func TestImportSettledOrdersXLSX(t *testing.T) {
	f := excelize.NewFile()
	sheet, _ := f.NewSheet("Data")
	headers := append([]string{"No."}, expectedHeadersOld...)
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 6)
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
		cell, _ := excelize.CoordinatesToCellName(i+1, 7)
		f.SetCellValue("Data", cell, v)
	}
	f.SetActiveSheet(sheet)
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatal(err)
	}

	repo := &fakeShopeeRepo{existingSettled: map[string]bool{"SO1": true}}
	svc := NewShopeeService(nil, repo, nil, nil, nil, config.ShopeeAPIConfig{})
	inserted, _, err := svc.ImportSettledOrdersXLSX(context.Background(), bytes.NewReader(buf.Bytes()))
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

	repo := &fakeShopeeRepo{existingSettled: map[string]bool{"SO1": true}}
	svc := NewShopeeService(nil, repo, nil, nil, nil, config.ShopeeAPIConfig{})
	_, _, err := svc.ImportSettledOrdersXLSX(context.Background(), bytes.NewReader(buf.Bytes()))
	if err == nil {
		t.Fatalf("expected error due to header mismatch")
	}
}

func TestImportSettledOrdersXLSX_SkipDuplicates(t *testing.T) {
	f := excelize.NewFile()
	sheet, _ := f.NewSheet("Data")
	headers := append([]string{"No."}, expectedHeadersOld...)
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 6)
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
		cell, _ := excelize.CoordinatesToCellName(i+1, 7)
		f.SetCellValue("Data", cell, v)
	}
	f.SetActiveSheet(sheet)
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatal(err)
	}

	repo := &fakeShopeeRepo{existingSettled: map[string]bool{"SO-1": true}}
	svc := NewShopeeService(nil, repo, nil, nil, nil, config.ShopeeAPIConfig{})
	inserted, _, err := svc.ImportSettledOrdersXLSX(context.Background(), bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("import error: %v", err)
	}
	if inserted != 0 || repo.count != 0 {
		t.Fatalf("expected 0 insert, got svc %d repo %d", inserted, repo.count)
	}
}

func TestImportSettledOrdersXLSX_UpdateDropshipStatus(t *testing.T) {
	f := excelize.NewFile()
	sheet, _ := f.NewSheet("Data")
	headers := append([]string{"No."}, expectedHeadersOld...)
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 6)
		f.SetCellValue("Data", cell, h)
	}
	data := []interface{}{
		1, "SO-2", "TRX-1", "user", "2025-01-01", "COD", "2025-01-02",
		1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
		1,
		1,
		1,
		1,
		"jne", "kurir", "",
		1, 1, 1, 1, 1,
	}
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
	drop := &fakeDropRepoA{byTrans: map[string]*models.DropshipPurchase{
		"TRX-1": {KodePesanan: "DP1", StatusPesananTerakhir: "Diproses"},
	}}
	svc := NewShopeeService(nil, repo, drop, nil, nil, config.ShopeeAPIConfig{})
	inserted, _, err := svc.ImportSettledOrdersXLSX(context.Background(), bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("import error: %v", err)
	}
	if inserted != 1 || repo.count != 1 {
		t.Fatalf("expected 1 insert, got svc %d repo %d", inserted, repo.count)
	}
	if drop.updated["DP1"] != "Pesanan selesai" {
		t.Fatalf("expected status update, got %v", drop.updated)
	}
}

func TestImportAffiliateCSV(t *testing.T) {
	csvData := "Kode Pesanan,Status Pesanan,Status Terverifikasi,Waktu Pesanan,Waktu Pesanan Selesai,Waktu Pesanan Terverifikasi,Kode Produk,Nama Produk,ID Model,L1 Kategori Global,L2 Kategori Global,L3 Kategori Global,Kode Promo,Harga(Rp),Jumlah,Nama Affiliate,Username Affiliate,MCN Terhubung,ID Komisi Pesanan,Partner Promo,Jenis Promo,Nilai Pembelian(Rp),Jumlah Pengembalian(Rp),Tipe Pesanan,Estimasi Komisi per Produk(Rp),Estimasi Komisi Affiliate per Produk(Rp),Persentase Komisi Affiliate per Produk,Estimasi Komisi MCN per Produk(Rp),Persentase Komisi MCN per Produk,Estimasi Komisi per Pesanan(Rp),Estimasi Komisi Affiliate per Pesanan(Rp),Estimasi Komisi MCN per Pesanan(Rp),Catatan Produk,Platform,Tingkat Komisi,Pengeluaran(Rp),Status Pemotongan,Metode Pemotongan,Waktu Pemotongan\n" +
		"SO1,Selesai,Sah,2025-06-01 10:00:00,,,P1,Produk,ID1,Cat1,Cat2,Cat3,,1000,1,Aff,affuser,,1,,Promo,1000,0,Langsung,10,10,10%,0,0%,10,10,0,,IG,10%,0,,,"
	repo := &fakeShopeeRepo{existingSettled: map[string]bool{"SO1": true}}
	svc := NewShopeeService(nil, repo, nil, nil, nil, config.ShopeeAPIConfig{})
	inserted, err := svc.ImportAffiliateCSV(context.Background(), strings.NewReader(csvData))
	if err != nil {
		t.Fatalf("import error: %v", err)
	}
	if inserted != 1 || repo.count != 1 {
		t.Fatalf("expected 1 insert, got svc %d repo %d", inserted, repo.count)
	}
}

func TestImportAffiliateCSV_JournalEntry(t *testing.T) {
	csvData := "Kode Pesanan,Status Pesanan,Status Terverifikasi,Waktu Pesanan,Waktu Pesanan Selesai,Waktu Pesanan Terverifikasi,Kode Produk,Nama Produk,ID Model,L1 Kategori Global,L2 Kategori Global,L3 Kategori Global,Kode Promo,Harga(Rp),Jumlah,Nama Affiliate,Username Affiliate,MCN Terhubung,ID Komisi Pesanan,Partner Promo,Jenis Promo,Nilai Pembelian(Rp),Jumlah Pengembalian(Rp),Tipe Pesanan,Estimasi Komisi per Produk(Rp),Estimasi Komisi Affiliate per Produk(Rp),Persentase Komisi Affiliate per Produk,Estimasi Komisi MCN per Produk(Rp),Persentase Komisi MCN per Produk,Estimasi Komisi per Pesanan(Rp),Estimasi Komisi Affiliate per Pesanan(Rp),Estimasi Komisi MCN per Pesanan(Rp),Catatan Produk,Platform,Tingkat Komisi,Pengeluaran(Rp),Status Pemotongan,Metode Pemotongan,Waktu Pemotongan\n" +
		"SO1,Selesai,Sah,2025-06-01 10:00:00,,,P1,Produk,ID1,Cat1,Cat2,Cat3,,1000,1,Aff,affuser,,1,,Promo,1000,0,Langsung,10,10,10%,0,0%,10,10,0,,IG,10%,5,,,"
	repo := &fakeShopeeRepo{existingSettled: map[string]bool{"SO1": true}}
	jr := &fakeJournalRepoS{}
	svc := NewShopeeService(nil, repo, nil, jr, nil, config.ShopeeAPIConfig{})
	inserted, err := svc.ImportAffiliateCSV(context.Background(), strings.NewReader(csvData))
	if err != nil {
		t.Fatalf("import error: %v", err)
	}
	if inserted != 1 || repo.count != 1 {
		t.Fatalf("expected 1 insert, got svc %d repo %d", inserted, repo.count)
	}
	if len(jr.entries) != 1 {
		t.Fatalf("expected 1 journal entry, got %d", len(jr.entries))
	}
	if jr.entries[0].SourceType != "shopee_affiliate" {
		t.Fatalf("wrong source type %s", jr.entries[0].SourceType)
	}
	if len(jr.lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(jr.lines))
	}
}

func TestImportAffiliateCSV_SkipDuplicate(t *testing.T) {
	csvData := "Kode Pesanan,Status Pesanan,Status Terverifikasi,Waktu Pesanan,Waktu Pesanan Selesai,Waktu Pesanan Terverifikasi,Kode Produk,Nama Produk,ID Model,L1 Kategori Global,L2 Kategori Global,L3 Kategori Global,Kode Promo,Harga(Rp),Jumlah,Nama Affiliate,Username Affiliate,MCN Terhubung,ID Komisi Pesanan,Partner Promo,Jenis Promo,Nilai Pembelian(Rp),Jumlah Pengembalian(Rp),Tipe Pesanan,Estimasi Komisi per Produk(Rp),Estimasi Komisi Affiliate per Produk(Rp),Persentase Komisi Affiliate per Produk,Estimasi Komisi MCN per Produk(Rp),Persentase Komisi MCN per Produk,Estimasi Komisi per Pesanan(Rp),Estimasi Komisi Affiliate per Pesanan(Rp),Estimasi Komisi MCN per Pesanan(Rp),Catatan Produk,Platform,Tingkat Komisi,Pengeluaran(Rp),Status Pemotongan,Metode Pemotongan,Waktu Pemotongan\n" +
		"SO1,Selesai,Sah,2025-06-01 10:00:00,,,P1,Produk,ID1,Cat1,Cat2,Cat3,,1000,1,Aff,affuser,,1,,Promo,1000,0,Langsung,10,10,10%,0,0%,10,10,0,,IG,10%,0,,,"
	repo := &fakeShopeeRepo{
		existingSettled:   map[string]bool{"SO1": true},
		existingAffiliate: map[string]bool{"SO1|P1|1": true},
	}
	jr := &fakeJournalRepoS{}
	svc := NewShopeeService(nil, repo, nil, jr, nil, config.ShopeeAPIConfig{})
	inserted, err := svc.ImportAffiliateCSV(context.Background(), strings.NewReader(csvData))
	if err != nil {
		t.Fatalf("import error: %v", err)
	}
	if inserted != 1 || repo.count != 1 {
		t.Fatalf("expected 1 insert, got svc %d repo %d", inserted, repo.count)
	}
	if len(repo.deletedAffiliate) != 1 {
		t.Fatalf("expected delete called")
	}
}

func TestImportAffiliateCSV_SkipMissingOrder(t *testing.T) {
	csvData := "Kode Pesanan,Status Pesanan,Status Terverifikasi,Waktu Pesanan,Waktu Pesanan Selesai,Waktu Pesanan Terverifikasi,Kode Produk,Nama Produk,ID Model,L1 Kategori Global,L2 Kategori Global,L3 Kategori Global,Kode Promo,Harga(Rp),Jumlah,Nama Affiliate,Username Affiliate,MCN Terhubung,ID Komisi Pesanan,Partner Promo,Jenis Promo,Nilai Pembelian(Rp),Jumlah Pengembalian(Rp),Tipe Pesanan,Estimasi Komisi per Produk(Rp),Estimasi Komisi Affiliate per Produk(Rp),Persentase Komisi Affiliate per Produk,Estimasi Komisi MCN per Produk(Rp),Persentase Komisi MCN per Produk,Estimasi Komisi per Pesanan(Rp),Estimasi Komisi Affiliate per Pesanan(Rp),Estimasi Komisi MCN per Pesanan(Rp),Catatan Produk,Platform,Tingkat Komisi,Pengeluaran(Rp),Status Pemotongan,Metode Pemotongan,Waktu Pemotongan\n" +
		"SO1,Selesai,Sah,2025-06-01 10:00:00,,,P1,Produk,ID1,Cat1,Cat2,Cat3,,1000,1,Aff,affuser,,1,,Promo,1000,0,Langsung,10,10,10%,0,0%,10,10,0,,IG,10%,0,,,"
	repo := &fakeShopeeRepo{}
	drop := &fakeDropRepoA{byInvoice: map[string]*models.DropshipPurchase{
		"SO1": {NamaToko: "TOKO"},
	}}
	jr := &fakeJournalRepoS{}
	svc := NewShopeeService(nil, repo, drop, jr, nil, config.ShopeeAPIConfig{})
	inserted, err := svc.ImportAffiliateCSV(context.Background(), strings.NewReader(csvData))
	if err != nil {
		t.Fatalf("import error: %v", err)
	}
	if inserted != 1 || repo.count != 1 {
		t.Fatalf("expected 1 insert, got svc %d repo %d", inserted, repo.count)
	}
	if len(jr.entries) != 0 {
		t.Fatalf("expected no journal entries, got %d", len(jr.entries))
	}
}

func TestImportAffiliateCSV_FilterStatus(t *testing.T) {
	csvData := "Kode Pesanan,Status Pesanan,Status Terverifikasi,Waktu Pesanan,Waktu Pesanan Selesai,Waktu Pesanan Terverifikasi,Kode Produk,Nama Produk,IDModel,L1 Kategori Global,L2 Kategori Global,L3 Kategori Global,Kode Promo,Harga(Rp),Jumlah,Nama Affiliate,Username Affiliate,MCN Terhubung,ID Komisi Pesanan,Partner Promo,Jenis Promo,Nilai Pembelian(Rp),Jumlah Pengembalian(Rp),Tipe Pesanan,Estimasi Komisi per Produk(Rp),Estimasi Komisi Affiliate per Produk(Rp),Persentase Komisi Affiliate per Produk,Estimasi Komisi MCN per Produk(Rp),Persentase Komisi MCN per Produk,Estimasi Komisi per Pesanan(Rp),Estimasi Komisi Affiliate per Pesanan(Rp),Estimasi Komisi MCN per Pesanan(Rp),Catatan Produk,Platform,Tingkat Komisi,Pengeluaran(Rp),Status Pemotongan,Metode Pemotongan,Waktu Pemotongan\n" +
		"SO1,Sedang Dikirim,Sah,2025-06-01 10:00:00,,,P1,Produk,ID1,Cat1,Cat2,Cat3,,1000,1,Aff,affuser,,1,,Promo,1000,0,Langsung,10,10,10%,0,0%,10,10,0,,IG,10%,0,,,"
	repo := &fakeShopeeRepo{}
	svc := NewShopeeService(nil, repo, nil, nil, nil, config.ShopeeAPIConfig{})
	inserted, err := svc.ImportAffiliateCSV(context.Background(), strings.NewReader(csvData))
	if err != nil {
		t.Fatalf("import error: %v", err)
	}
	if inserted != 0 || repo.count != 0 {
		t.Fatalf("expected 0 insert, got svc %d repo %d", inserted, repo.count)
	}
}

func TestImportAffiliateCSV_NonSahNoJournal(t *testing.T) {
	csvData := "Kode Pesanan,Status Pesanan,Status Terverifikasi,Waktu Pesanan,Waktu Pesanan Selesai,Waktu Pesanan Terverifikasi,Kode Produk,Nama Produk,IDModel,L1 Kategori Global,L2 Kategori Global,L3 Kategori Global,Kode Promo,Harga(Rp),Jumlah,Nama Affiliate,Username Affiliate,MCN Terhubung,ID Komisi Pesanan,Partner Promo,Jenis Promo,Nilai Pembelian(Rp),Jumlah Pengembalian(Rp),Tipe Pesanan,Estimasi Komisi per Produk(Rp),Estimasi Komisi Affiliate per Produk(Rp),Persentase Komisi Affiliate per Produk,Estimasi Komisi MCN per Produk(Rp),Persentase Komisi MCN per Produk,Estimasi Komisi per Pesanan(Rp),Estimasi Komisi Affiliate per Pesanan(Rp),Estimasi Komisi MCN per Pesanan(Rp),Catatan Produk,Platform,Tingkat Komisi,Pengeluaran(Rp),Status Pemotongan,Metode Pemotongan,Waktu Pemotongan\n" +
		"SO1,Selesai,Tidak,2025-06-01 10:00:00,,,P1,Produk,ID1,Cat1,Cat2,Cat3,,1000,1,Aff,affuser,,1,,Promo,1000,0,Langsung,10,10,10%,0,0%,10,10,0,,IG,10%,5,,,"
	repo := &fakeShopeeRepo{existingSettled: map[string]bool{"SO1": true}}
	jr := &fakeJournalRepoS{}
	svc := NewShopeeService(nil, repo, nil, jr, nil, config.ShopeeAPIConfig{})
	inserted, err := svc.ImportAffiliateCSV(context.Background(), strings.NewReader(csvData))
	if err != nil {
		t.Fatalf("import error: %v", err)
	}
	if inserted != 1 || repo.count != 1 {
		t.Fatalf("expected 1 insert, got svc %d repo %d", inserted, repo.count)
	}
	if len(jr.entries) != 0 {
		t.Fatalf("expected no journal entries, got %d", len(jr.entries))
	}
}

func TestConfirmSettleCreatesFeeLines(t *testing.T) {
	repo := &fakeShopeeRepo{
		order: &models.ShopeeSettled{
			NamaToko:                 "MR eStore Shopee",
			NoPesanan:                "SO1",
			WaktuPesananDibuat:       time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
			TanggalDanaDilepaskan:    time.Date(2025, 6, 2, 0, 0, 0, 0, time.UTC),
			HargaAsliProduk:          100,
			TotalDiskonProduk:        -10,
			BiayaAdminShopee:         -5,
			PromoGratisOngkirPenjual: -2,
			PromoDiskonShopee:        -3,
		},
		affExpense: 1,
	}
	jr := &fakeJournalRepoS{}
	svc := NewShopeeService(nil, repo, nil, jr, nil, config.ShopeeAPIConfig{})

	if err := svc.ConfirmSettle(context.Background(), "SO1"); err != nil {
		t.Fatalf("confirm error: %v", err)
	}
	if len(jr.entries) != 1 {
		t.Fatalf("expected 1 journal entry, got %d", len(jr.entries))
	}
	if jr.entries[0].EntryDate != repo.order.TanggalDanaDilepaskan {
		t.Fatalf("wrong entry date")
	}
	if len(jr.lines) != 7 {
		t.Fatalf("expected 7 journal lines, got %d", len(jr.lines))
	}
	if len(repo.confirmed) != 1 || repo.confirmed[0] != "SO1" {
		t.Fatalf("confirm not recorded: %v", repo.confirmed)
	}
}

func TestConfirmSettleMismatchCreatesAdjustment(t *testing.T) {
	repo := &fakeShopeeRepo{
		order: &models.ShopeeSettled{
			NamaToko:              "TOKO",
			NoPesanan:             "SO2",
			TanggalDanaDilepaskan: time.Date(2025, 6, 2, 0, 0, 0, 0, time.UTC),
			HargaAsliProduk:       110,
			TotalDiskonProduk:     -10,
			IsDataMismatch:        true,
		},
	}
	drop := &fakeDropRepoA{byInvoice: map[string]*models.DropshipPurchase{
		"SO2": {TotalTransaksi: 100},
	}}
	jr := &fakeJournalRepoS{}
	svc := NewShopeeService(nil, repo, drop, jr, nil, config.ShopeeAPIConfig{})

	if err := svc.ConfirmSettle(context.Background(), "SO2"); err != nil {
		t.Fatalf("confirm error: %v", err)
	}
	if len(jr.entries) != 3 {
		t.Fatalf("expected 3 journal entries, got %d", len(jr.entries))
	}
	if jr.entries[0].SourceType != "shopee_grossup" {
		t.Fatalf("first entry not gross up")
	}
	if jr.entries[1].SourceType != "shopee_discount" {
		t.Fatalf("second entry not discount")
	}
	if len(jr.lines) != 6 {
		t.Fatalf("expected 6 journal lines, got %d", len(jr.lines))
	}
}

func TestImportSettledOrdersXLSX_AutoSettle(t *testing.T) {
	f := excelize.NewFile()
	sheet, _ := f.NewSheet("Data")
	headers := append([]string{"No."}, expectedHeadersOld...)
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 6)
		f.SetCellValue("Data", cell, h)
	}
	data := []interface{}{
		1, "SO-3", "TRX", "user", "2025-01-01", "COD", "2025-01-02",
		1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
		1,
		1,
		1,
		1,
		"jne", "kurir", "",
		1, 1, 1, 1, 1,
	}
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
	drop := &fakeDropRepoA{byInvoice: map[string]*models.DropshipPurchase{
		"SO-3": {TotalTransaksi: 1},
	}}
	jr := &fakeJournalRepoS{}
	svc := NewShopeeService(nil, repo, drop, jr, nil, config.ShopeeAPIConfig{})
	inserted, mis, err := svc.ImportSettledOrdersXLSX(context.Background(), bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("import error: %v", err)
	}
	if inserted != 1 || len(mis) != 0 {
		t.Fatalf("unexpected insert %d mismatches %v", inserted, mis)
	}
	if len(repo.confirmed) != 1 {
		t.Fatalf("auto settle not confirmed")
	}
	if len(jr.entries) != 1 {
		t.Fatalf("expected 1 journal entry, got %d", len(jr.entries))
	}
}

func TestImportSettledOrdersXLSX_AutoAdjustMismatch(t *testing.T) {
	f := excelize.NewFile()
	sheet, _ := f.NewSheet("Data")
	headers := append([]string{"No."}, expectedHeadersOld...)
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 6)
		f.SetCellValue("Data", cell, h)
	}
	data := []interface{}{
		1, "SO-4", "TRX", "user", "2025-01-01", "COD", "2025-01-02",
		110, -10, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
		1,
		1,
		1,
		1,
		"jne", "kurir", "",
		1, 1, 1, 1, 1,
	}
	for i, v := range data {
		cell, _ := excelize.CoordinatesToCellName(i+1, 7)
		f.SetCellValue("Data", cell, v)
	}
	f.SetActiveSheet(sheet)
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatal(err)
	}

	repo := &fakeShopeeRepo{order: &models.ShopeeSettled{
		NamaToko:              "TOKO",
		NoPesanan:             "SO-4",
		TanggalDanaDilepaskan: time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
		HargaAsliProduk:       110,
		TotalDiskonProduk:     -10,
		IsDataMismatch:        true,
	}}
	drop := &fakeDropRepoA{byInvoice: map[string]*models.DropshipPurchase{
		"SO-4": {TotalTransaksi: 100},
	}}
	jr := &fakeJournalRepoS{}
	svc := NewShopeeService(nil, repo, drop, jr, nil, config.ShopeeAPIConfig{})
	inserted, mis, err := svc.ImportSettledOrdersXLSX(context.Background(), bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("import error: %v", err)
	}
	if inserted != 1 || len(mis) != 0 {
		t.Fatalf("unexpected insert %d mismatches %v", inserted, mis)
	}
	if len(repo.confirmed) != 1 || repo.confirmed[0] != "SO-4" {
		t.Fatalf("auto settle not confirmed: %v", repo.confirmed)
	}
	if len(jr.entries) != 3 {
		t.Fatalf("expected 3 journal entries, got %d", len(jr.entries))
	}
}

func TestCreateReturnedOrderJournal(t *testing.T) {
	repo := &fakeShopeeRepo{
		order: &models.ShopeeSettled{
			NamaToko:              "MR eStore Shopee",
			NoPesanan:             "SO-RET-1",
			TanggalDanaDilepaskan: time.Date(2025, 6, 2, 0, 0, 0, 0, time.UTC),
			HargaAsliProduk:       200000,
		},
	}

	// Create an original settlement journal entry for jakmall calculation
	jr := &fakeJournalRepoS{nextID: 1}
	originalEntry := &models.JournalEntry{
		JournalID:    1,
		EntryDate:    time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
		SourceType:   "shopee_settled",
		SourceID:     "SO-RET-1",
		ShopUsername: "MR eStore Shopee",
		Store:        "MR eStore Shopee",
	}
	jr.entries = append(jr.entries, originalEntry)

	// Add original settlement lines: pending credit 200000, saldo shopee debit 150000 (75% jakmall)
	jr.lines = append(jr.lines,
		&models.JournalLine{LineID: 1, JournalID: 1, AccountID: 11010, IsDebit: false, Amount: 200000}, // Pending credit
		&models.JournalLine{LineID: 2, JournalID: 1, AccountID: 11011, IsDebit: true, Amount: 150000},  // Saldo shopee (jakmall)
		&models.JournalLine{LineID: 3, JournalID: 1, AccountID: 55004, IsDebit: true, Amount: 50000},   // Other expenses
	)

	svc := NewShopeeService(nil, repo, nil, jr, nil, config.ShopeeAPIConfig{})

	// Create escrow detail for returned order
	escrowDetail := ShopeeEscrowDetail{
		"seller_return_refund": -113400, // Negative as per sample
		"reverse_shipping_fee": 5000,
		"commission_fee":       4536,
		"service_fee":          3686,
	}

	settlementDate := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)

	err := svc.CreateReturnedOrderJournal(context.Background(), "SO-RET-1", escrowDetail, settlementDate)
	if err != nil {
		t.Fatalf("CreateReturnedOrderJournal error: %v", err)
	}

	// Check that a new journal entry was created
	if len(jr.entries) != 2 {
		t.Fatalf("expected 2 journal entries (original + return), got %d", len(jr.entries))
	}

	returnEntry := jr.entries[1]
	if returnEntry.SourceType != "shopee_return" {
		t.Fatalf("expected return entry source type shopee_return, got %s", returnEntry.SourceType)
	}
	if returnEntry.SourceID != "SO-RET-1-return" {
		t.Fatalf("expected return entry source ID SO-RET-1-return, got %s", returnEntry.SourceID)
	}
	if returnEntry.EntryDate != settlementDate {
		t.Fatalf("expected return entry date %v, got %v", settlementDate, returnEntry.EntryDate)
	}

	// Count lines for the return entry (journalID = 2, which is the second entry created)
	returnJournalID := returnEntry.JournalID
	returnLines := 0
	var pendingCreditFound, salesDebitFound, jakmallDebitFound, shippingDebitFound, commissionDebitFound, serviceDebitFound bool

	for _, line := range jr.lines {
		if line.JournalID == returnJournalID { // Return entry
			returnLines++
			switch line.AccountID {
			case 11010: // Pending account
				if !line.IsDebit && line.Amount == 113400 {
					pendingCreditFound = true
				}
			case 4001: // Sales account
				if line.IsDebit && line.Amount == 113400 {
					salesDebitFound = true
				}
			case 11011: // Saldo Shopee (jakmall adjustment)
				if line.IsDebit && line.Amount > 0 {
					jakmallDebitFound = true
					// Should be 113400 * (150000/200000) = 113400 * 0.75 = 85050
					expectedJakmall := 113400 * 0.75
					if line.Amount != expectedJakmall {
						t.Fatalf("expected jakmall adjustment %.2f, got %.2f", expectedJakmall, line.Amount)
					}
				}
			case 52010: // Shipping expense
				if line.IsDebit && line.Amount == 5000 {
					shippingDebitFound = true
				}
			case 52006: // Commission expense
				if line.IsDebit && line.Amount == 4536 {
					commissionDebitFound = true
				}
			case 52004: // Service expense
				if line.IsDebit && line.Amount == 3686 {
					serviceDebitFound = true
				}
			}
		}
	}

	if returnLines != 6 {
		t.Fatalf("expected 6 return journal lines, got %d", returnLines)
	}

	if !pendingCreditFound {
		t.Fatalf("pending balance reduction (credit) not found")
	}
	if !salesDebitFound {
		t.Fatalf("sales reduction (debit) not found")
	}
	if !jakmallDebitFound {
		t.Fatalf("jakmall adjustment (debit) not found")
	}
	if !shippingDebitFound {
		t.Fatalf("shipping fee (debit) not found")
	}
	if !commissionDebitFound {
		t.Fatalf("commission fee (debit) not found")
	}
	if !serviceDebitFound {
		t.Fatalf("service fee (debit) not found")
	}
}

func TestCreateReturnedOrderJournal_NoOriginalJournal(t *testing.T) {
	repo := &fakeShopeeRepo{
		order: &models.ShopeeSettled{
			NamaToko:  "MR eStore Shopee",
			NoPesanan: "SO-RET-NO-ORIG",
		},
	}

	jr := &fakeJournalRepoS{} // No original journal entries
	svc := NewShopeeService(nil, repo, nil, jr, nil, config.ShopeeAPIConfig{})

	escrowDetail := ShopeeEscrowDetail{
		"seller_return_refund": -50000,
		"reverse_shipping_fee": 2000,
	}

	settlementDate := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)

	// Should still work but without jakmall adjustment
	err := svc.CreateReturnedOrderJournal(context.Background(), "SO-RET-NO-ORIG", escrowDetail, settlementDate)
	if err != nil {
		t.Fatalf("CreateReturnedOrderJournal error: %v", err)
	}

	// Should create return entry without jakmall adjustment
	if len(jr.entries) != 1 {
		t.Fatalf("expected 1 journal entry, got %d", len(jr.entries))
	}

	// Should have 3 lines: pending credit, sales debit, shipping debit (no jakmall)
	returnLines := 0
	for _, line := range jr.lines {
		if line.JournalID == 1 {
			returnLines++
		}
	}
	if returnLines != 3 {
		t.Fatalf("expected 3 return journal lines (no jakmall), got %d", returnLines)
	}
}

func TestCreateReturnedOrderJournal_ZeroRefund(t *testing.T) {
	repo := &fakeShopeeRepo{
		order: &models.ShopeeSettled{
			NamaToko:  "MR eStore Shopee",
			NoPesanan: "SO-RET-ZERO",
		},
	}

	jr := &fakeJournalRepoS{}
	svc := NewShopeeService(nil, repo, nil, jr, nil, config.ShopeeAPIConfig{})

	escrowDetail := ShopeeEscrowDetail{
		"seller_return_refund": 0, // No refund
		"reverse_shipping_fee": 1000,
	}

	settlementDate := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)

	err := svc.CreateReturnedOrderJournal(context.Background(), "SO-RET-ZERO", escrowDetail, settlementDate)
	if err == nil || !strings.Contains(err.Error(), "no return refund amount") {
		t.Fatalf("expected error for zero refund amount, got %v", err)
	}
}
