// File: backend/internal/service/dropship_service_test.go

package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/config"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
)

type testRoundTripper func(*http.Request) (*http.Response, error)

func (f testRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

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

func (f *fakeDropshipRepo) ListExistingPurchases(ctx context.Context, ids []string) (map[string]bool, error) {
	res := make(map[string]bool)
	for _, id := range ids {
		if f.existing[id] {
			res[id] = true
		}
	}
	return res, nil
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

type fakeStoreRepo struct{ store *models.Store }

func (f *fakeStoreRepo) GetStoreByName(ctx context.Context, name string) (*models.Store, error) {
	if f.store != nil && f.store.NamaToko == name {
		return f.store, nil
	}
	return nil, nil
}

func (f *fakeStoreRepo) UpdateStore(ctx context.Context, s *models.Store) error {
	f.store = s
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

	// fake Shopee detail server
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "access_token/get") {
			fmt.Fprint(w, `{"response":{"access_token":"tok","refresh_token":"ref","expire_in":3600,"request_id":"1"}}`)
			return
		}
		if strings.Contains(r.URL.Path, "get_order_detail") {
			fmt.Fprint(w, `{"response":{"order_list":[{"order_sn":"INV1","item_list":[{"model_original_price":"15.75","model_quantity_purchased":2}]}]}}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	oldTransport := http.DefaultTransport
	http.DefaultTransport = testRoundTripper(func(req *http.Request) (*http.Response, error) {
		req.URL.Scheme = "http"
		req.URL.Host = strings.TrimPrefix(srv.URL, "http://")
		return oldTransport.RoundTrip(req)
	})
	defer func() { http.DefaultTransport = oldTransport }()

	client := NewShopeeClient(config.ShopeeAPIConfig{BaseURLShopee: srv.URL, PartnerID: "1", PartnerKey: "key", ShopID: "2"})
	client.httpClient = srv.Client()

	now := time.Now()
	exp := 3600
	store := &models.Store{NamaToko: "MyShop", AccessToken: ptrString("tok"), RefreshToken: ptrString("ref"), ShopID: ptrString("2"), ExpireIn: &exp, LastUpdated: &now}

	fake := &fakeDropshipRepo{}
	jfake := &fakeJournalRepoDrop{}
	srepo := &fakeStoreRepo{store: store}
	svc := NewDropshipService(nil, fake, jfake, srepo, nil, nil, client, 5)

	ctx := context.Background()
	count, err := svc.ImportFromCSV(ctx, &buf, "", 0)
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

func TestImportFromCSV_CleanInvoiceQuote(t *testing.T) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	headers := []string{"No", "waktu", "status", "kode", "trx", "sku", "nama", "harga", "qty", "total_harga", "biaya_lain", "biaya_mitra", "total_transaksi", "harga_ch", "total_harga_ch", "potensi", "dibuat", "channel", "toko", "invoice", "gudang", "ekspedisi", "cashless", "resi", "waktu_kirim", "provinsi", "kota"}
	w.Write(headers)
	row := []string{"1", "01 January 2025, 10:00:00", "selesai", "PS-123Q", "TRX1", "SKU1", "ProdukA", "15.75", "2", "31.50", "1", "0.5", "33.0", "15.75", "31.50", "2.0", "user", "online", "MyShop", "'INVQ", "GudangA", "JNE", "Ya", "RESI1", "02 January 2025, 10:00:00", "Jawa", "Bandung"}
	w.Write(row)
	w.Flush()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "access_token/get") {
			fmt.Fprint(w, `{"response":{"access_token":"tok","refresh_token":"ref","expire_in":3600,"request_id":"1"}}`)
			return
		}
		if strings.Contains(r.URL.Path, "get_order_detail") {
			fmt.Fprint(w, `{"response":{"order_list":[{"order_sn":"INVQ","item_list":[{"model_original_price":"15.75","model_quantity_purchased":2}]}]}}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	oldTransport := http.DefaultTransport
	http.DefaultTransport = testRoundTripper(func(req *http.Request) (*http.Response, error) {
		req.URL.Scheme = "http"
		req.URL.Host = strings.TrimPrefix(srv.URL, "http://")
		return oldTransport.RoundTrip(req)
	})
	defer func() { http.DefaultTransport = oldTransport }()

	client := NewShopeeClient(config.ShopeeAPIConfig{BaseURLShopee: srv.URL, PartnerID: "1", PartnerKey: "key", ShopID: "2"})
	client.httpClient = srv.Client()

	now := time.Now()
	exp := 3600
	store := &models.Store{NamaToko: "MyShop", AccessToken: ptrString("tok"), RefreshToken: ptrString("ref"), ShopID: ptrString("2"), ExpireIn: &exp, LastUpdated: &now}

	fake := &fakeDropshipRepo{}
	jfake := &fakeJournalRepoDrop{}
	srepo := &fakeStoreRepo{store: store}
	svc := NewDropshipService(nil, fake, jfake, srepo, nil, nil, client, 5)

	ctx := context.Background()
	count, err := svc.ImportFromCSV(ctx, &buf, "", 0)
	if err != nil {
		t.Fatalf("ImportFromCSV error: %v", err)
	}
	if count != 1 {
		t.Errorf("expected count 1, got %d", count)
	}

	if len(fake.insertedHeader) != 1 {
		t.Fatalf("expected 1 header insert, got %d", len(fake.insertedHeader))
	}
	inserted := fake.insertedHeader[0]
	if inserted.KodeInvoiceChannel != "INVQ" {
		t.Errorf("expected invoice 'INVQ', got '%s'", inserted.KodeInvoiceChannel)
	}
}

func TestImportFromCSV_ParseError(t *testing.T) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	w.Write([]string{"No", "waktu", "status", "kode", "trx", "sku", "nama", "harga", "qty", "total_harga", "biaya_lain", "biaya_mitra", "total_transaksi", "harga_ch", "total_harga_ch", "potensi", "dibuat", "channel", "toko", "invoice", "gudang", "ekspedisi", "cashless", "resi", "waktu_kirim", "provinsi", "kota"})
	w.Write([]string{"1", "01 January 2025, 10:00:00", "selesai", "PS-456", "TRX1", "SKU2", "ProdukB", "15.00", "two", "30", "1", "0.5", "31.5", "15", "30", "2", "user", "online", "Shop", "INV", "G", "JNE", "Ya", "RESI", "02 January 2025, 10:00:00", "Jawa", "Bandung"})
	w.Flush()

	fake := &fakeDropshipRepo{}
	svc := NewDropshipService(nil, fake, nil, nil, nil, nil, nil, 5)
	count, err := svc.ImportFromCSV(context.Background(), &buf, "", 0)
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

func TestImportFromCSV_SkipExisting(t *testing.T) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	headers := []string{"No", "waktu", "status", "kode", "trx", "sku", "nama", "harga", "qty", "total_harga", "biaya_lain", "biaya_mitra", "total_transaksi", "harga_ch", "total_harga_ch", "potensi", "dibuat", "channel", "toko", "invoice", "gudang", "ekspedisi", "cashless", "resi", "waktu_kirim", "provinsi", "kota"}
	w.Write(headers)
	row := []string{"1", "01 January 2025, 10:00:00", "selesai", "PS-EXIST", "TRX1", "SKU1", "ProdukA", "15.75", "2", "31.50", "1", "0.5", "33.0", "15.75", "31.50", "2.0", "user", "online", "MyShop", "INV1", "GudangA", "JNE", "Ya", "RESI1", "02 January 2025, 10:00:00", "Jawa", "Bandung"}
	w.Write(row)
	w.Flush()

	fake := &fakeDropshipRepo{existing: map[string]bool{"PS-EXIST": true}}
	svc := NewDropshipService(nil, fake, nil, nil, nil, nil, nil, 5)
	count, err := svc.ImportFromCSV(context.Background(), &buf, "", 0)
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

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "access_token/get") {
			fmt.Fprint(w, `{"response":{"access_token":"tok","refresh_token":"ref","expire_in":3600,"request_id":"1"}}`)
			return
		}
		if strings.Contains(r.URL.Path, "get_order_detail") {
			fmt.Fprint(w, `{"response":{"order_list":[{"order_sn":"INV1","item_list":[{"model_original_price":"15.75","model_quantity_purchased":2},{"model_original_price":"20","model_quantity_purchased":1}]}]}}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	oldTransport := http.DefaultTransport
	http.DefaultTransport = testRoundTripper(func(req *http.Request) (*http.Response, error) {
		req.URL.Scheme = "http"
		req.URL.Host = strings.TrimPrefix(srv.URL, "http://")
		return oldTransport.RoundTrip(req)
	})
	defer func() { http.DefaultTransport = oldTransport }()

	client := NewShopeeClient(config.ShopeeAPIConfig{BaseURLShopee: srv.URL, PartnerID: "1", PartnerKey: "key", ShopID: "2"})
	client.httpClient = srv.Client()
	now := time.Now()
	exp := 3600
	store := &models.Store{NamaToko: "MyShop", AccessToken: ptrString("tok"), RefreshToken: ptrString("ref"), ShopID: ptrString("2"), ExpireIn: &exp, LastUpdated: &now}

	fake := &fakeDropshipRepo{}
	jfake := &fakeJournalRepoDrop{}
	srepo := &fakeStoreRepo{store: store}
	svc := NewDropshipService(nil, fake, jfake, srepo, nil, nil, client, 5)

	count, err := svc.ImportFromCSV(context.Background(), &buf, "", 0)
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
	svc := NewDropshipService(nil, fake, nil, nil, nil, nil, nil, 5)

	count, err := svc.ImportFromCSV(context.Background(), &buf, "Tokopedia", 0)
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

func TestImportFromCSV_SkipOnDetailError(t *testing.T) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	headers := []string{"No", "waktu", "status", "kode", "trx", "sku", "nama", "harga", "qty", "total_harga", "biaya_lain", "biaya_mitra", "total_transaksi", "harga_ch", "total_harga_ch", "potensi", "dibuat", "channel", "toko", "invoice", "gudang", "ekspedisi", "cashless", "resi", "waktu_kirim", "provinsi", "kota"}
	w.Write(headers)
	row := []string{"1", "01 January 2025, 10:00:00", "selesai", "PS-ERR", "TRX1", "SKU1", "ProdukA", "15", "1", "15", "0", "0", "15", "15", "15", "0", "user", "online", "MyShop", "INVERR", "Gudang", "JNE", "Ya", "RESI", "02 January 2025, 10:00:00", "Jawa", "Bandung"}
	w.Write(row)
	w.Flush()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "access_token/get") {
			fmt.Fprint(w, `{"response":{"access_token":"tok","refresh_token":"ref","expire_in":3600,"request_id":"1"}}`)
			return
		}
		if strings.Contains(r.URL.Path, "get_order_detail") {
			http.Error(w, "fail", http.StatusInternalServerError)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	oldTransport := http.DefaultTransport
	http.DefaultTransport = testRoundTripper(func(req *http.Request) (*http.Response, error) {
		req.URL.Scheme = "http"
		req.URL.Host = strings.TrimPrefix(srv.URL, "http://")
		return oldTransport.RoundTrip(req)
	})
	defer func() { http.DefaultTransport = oldTransport }()

	client := NewShopeeClient(config.ShopeeAPIConfig{BaseURLShopee: srv.URL, PartnerID: "1", PartnerKey: "key", ShopID: "2"})
	client.httpClient = srv.Client()

	now := time.Now()
	exp := 3600
	store := &models.Store{NamaToko: "MyShop", AccessToken: ptrString("tok"), RefreshToken: ptrString("ref"), ShopID: ptrString("2"), ExpireIn: &exp, LastUpdated: &now}

	fakeRepo := &fakeDropshipRepo{}
	srepo := &fakeStoreRepo{store: store}
	svc := NewDropshipService(nil, fakeRepo, nil, srepo, nil, nil, client, 5)

	count, err := svc.ImportFromCSV(context.Background(), &buf, "", 0)
	if err != nil {
		t.Fatalf("ImportFromCSV error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 rows, got %d", count)
	}
	if len(fakeRepo.insertedHeader) != 0 {
		t.Errorf("expected no inserts, got %d", len(fakeRepo.insertedHeader))
	}
}

func TestImportFromCSV_FreeSample(t *testing.T) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	headers := []string{"No", "waktu", "status", "kode", "trx", "sku", "nama", "harga", "qty", "total_harga", "biaya_lain", "biaya_mitra", "total_transaksi", "harga_ch", "total_harga_ch", "potensi", "dibuat", "channel", "toko", "invoice", "gudang", "ekspedisi", "cashless", "resi", "waktu_kirim", "provinsi", "kota"}
	w.Write(headers)
	row := []string{"1", "01 January 2025, 10:00:00", "selesai", "PS-FS", "TRX1", "SKU1", "ProdukA", "10", "1", "10", "0", "0", "10", "10", "10", "0", "user", "online", "MR eStore Free Sample", "INVFS", "Gudang", "JNE", "Ya", "RESI", "02 January 2025, 10:00:00", "Jawa", "Bandung"}
	w.Write(row)
	w.Flush()

	fakeRepo := &fakeDropshipRepo{}
	jfake := &fakeJournalRepoDrop{}
	svc := NewDropshipService(nil, fakeRepo, jfake, nil, nil, nil, nil, 5)

	count, err := svc.ImportFromCSV(context.Background(), &buf, "", 0)
	if err != nil {
		t.Fatalf("ImportFromCSV error: %v", err)
	}
	if count != 1 {
		t.Errorf("expected count 1, got %d", count)
	}
	if len(jfake.entries) != 1 || len(jfake.lines) != 2 {
		t.Fatalf("expected 1 journal entry and 2 lines, got %d/%d", len(jfake.entries), len(jfake.lines))
	}
}
