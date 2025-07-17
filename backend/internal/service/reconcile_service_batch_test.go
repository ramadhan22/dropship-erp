package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/config"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

type fakeDropRepoBatch struct {
	data    map[string]*models.DropshipPurchase
	lookups []string
}

func (f *fakeDropRepoBatch) GetDropshipPurchaseByInvoice(ctx context.Context, inv string) (*models.DropshipPurchase, error) {
	f.lookups = append(f.lookups, inv)
	if dp, ok := f.data[inv]; ok {
		return dp, nil
	}
	return nil, fmt.Errorf("not found")
}

func (f *fakeDropRepoBatch) GetDropshipPurchaseByID(ctx context.Context, kode string) (*models.DropshipPurchase, error) {
	if dp, ok := f.data[kode]; ok {
		return dp, nil
	}
	return nil, fmt.Errorf("not found")
}

func (f *fakeDropRepoBatch) UpdatePurchaseStatus(ctx context.Context, kode, status string) error {
	return nil
}
func (f *fakeDropRepoBatch) SumDetailByInvoice(ctx context.Context, inv string) (float64, error) {
	return 0, nil
}
func (f *fakeDropRepoBatch) SumProductCostByInvoice(ctx context.Context, inv string) (float64, error) {
	return 0, nil
}

type fakeStoreRepoBatch struct{ store *models.Store }

func (f *fakeStoreRepoBatch) GetStoreByName(ctx context.Context, name string) (*models.Store, error) {
	if f.store != nil && f.store.NamaToko == name {
		return f.store, nil
	}
	return nil, fmt.Errorf("not found")
}
func (f *fakeStoreRepoBatch) UpdateStore(ctx context.Context, s *models.Store) error {
	f.store = s
	return nil
}

type fakeJournalRepoBatch struct {
	entries []*models.JournalEntry
	lines   []*models.JournalLine
	nextID  int64
}

func (f *fakeJournalRepoBatch) CreateJournalEntry(ctx context.Context, e *models.JournalEntry) (int64, error) {
	f.nextID++
	e.JournalID = f.nextID
	f.entries = append(f.entries, e)
	return f.nextID, nil
}
func (f *fakeJournalRepoBatch) InsertJournalLine(ctx context.Context, l *models.JournalLine) error {
	f.lines = append(f.lines, l)
	return nil
}
func (f *fakeJournalRepoBatch) InsertJournalLines(ctx context.Context, lines []models.JournalLine) error {
	for i := range lines {
		f.lines = append(f.lines, &lines[i])
	}
	return nil
}

type fakeBatchSvc struct {
	created []*models.BatchHistory
	updated []int64
}

func (f *fakeBatchSvc) Create(ctx context.Context, b *models.BatchHistory) (int64, error) {
	f.created = append(f.created, b)
	return int64(len(f.created)), nil
}
func (f *fakeBatchSvc) UpdateDone(ctx context.Context, id int64, done int) error {
	f.updated = append(f.updated, id)
	return nil
}
func (f *fakeBatchSvc) UpdateStatus(ctx context.Context, id int64, status, msg string) error {
	return nil
}

func (f *fakeBatchSvc) CreateDetail(ctx context.Context, d *models.BatchHistoryDetail) error {
	return nil
}
func (f *fakeBatchSvc) ListDetails(ctx context.Context, id int64) ([]models.BatchHistoryDetail, error) {
	return []models.BatchHistoryDetail{}, nil
}
func (f *fakeBatchSvc) UpdateDetailStatus(ctx context.Context, id int64, status, msg string) error {
	return nil
}
func (f *fakeBatchSvc) ListPendingByType(ctx context.Context, typ string) ([]models.BatchHistory, error) {
	return []models.BatchHistory{}, nil
}

func TestProcessShopeeStatusBatch_Escrow(t *testing.T) {
	dp1 := &models.DropshipPurchase{KodePesanan: "DP1", KodeInvoiceChannel: "INV1", NamaToko: "ShopA"}
	dp2 := &models.DropshipPurchase{KodePesanan: "DP2", KodeInvoiceChannel: "INV2", NamaToko: "ShopA"}
	drop := &fakeDropRepoBatch{data: map[string]*models.DropshipPurchase{"INV1": dp1, "INV2": dp2}}

	now := time.Now()
	exp := 3600 * 24
	store := &models.Store{NamaToko: "ShopA", AccessToken: ptrString("tok"), RefreshToken: ptrString("ref"), ShopID: ptrString("2"), ExpireIn: &exp, LastUpdated: &now}
	srepo := &fakeStoreRepoBatch{store: store}
	jrepo := &fakeJournalRepoBatch{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/order/get_order_detail":
			fmt.Fprint(w, `{"response":{"order_list":[{"order_sn":"INV1","order_status":"COMPLETED","update_time":1},{"order_sn":"INV2","order_status":"COMPLETED","update_time":1}]}}`)
		case "/api/v2/payment/get_escrow_detail_batch":
			var req struct {
				OrderSNList []string `json:"order_sn_list"`
			}
			json.NewDecoder(r.Body).Decode(&req)
			fmt.Fprint(w, `{"response":[{"order_sn":"INV1","escrow_detail":{"order_income":{"order_original_price":1,"escrow_amount":1}}},{"order_sn":"INV2","escrow_detail":{"order_income":{"order_original_price":2,"escrow_amount":2}}}]}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	client := NewShopeeClient(config.ShopeeAPIConfig{BaseURLShopee: srv.URL, PartnerID: "1", PartnerKey: "key"})
	client.httpClient = srv.Client()

	svc := NewReconcileService(nil, drop, nil, jrepo, nil, srepo, nil, nil, client, nil, nil, 5, nil)

	svc.processShopeeStatusBatch(context.Background(), "ShopA", []*models.DropshipPurchase{dp1, dp2})

	if len(drop.lookups) != 2 {
		t.Fatalf("expected 2 lookups, got %d", len(drop.lookups))
	}
	found1, found2 := false, false
	for _, v := range drop.lookups {
		if v == "INV1" {
			found1 = true
		}
		if v == "INV2" {
			found2 = true
		}
	}
	if !found1 || !found2 {
		t.Fatalf("unexpected lookups %v", drop.lookups)
	}
}

func TestProcessShopeeStatusBatch_Escrow_OrderSNMismatch(t *testing.T) {
	dp1 := &models.DropshipPurchase{KodePesanan: "DP1", KodeInvoiceChannel: "INV1", NamaToko: "ShopA"}
	dp2 := &models.DropshipPurchase{KodePesanan: "DP2", KodeInvoiceChannel: "INV2", NamaToko: "ShopA"}
	drop := &fakeDropRepoBatch{data: map[string]*models.DropshipPurchase{"INV1": dp1, "INV2": dp2}}

	now := time.Now()
	exp := 3600 * 24
	store := &models.Store{NamaToko: "ShopA", AccessToken: ptrString("tok"), RefreshToken: ptrString("ref"), ShopID: ptrString("2"), ExpireIn: &exp, LastUpdated: &now}
	srepo := &fakeStoreRepoBatch{store: store}
	jrepo := &fakeJournalRepoBatch{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/order/get_order_detail":
			fmt.Fprint(w, `{"response":{"order_list":[{"order_sn":"DP1","order_status":"COMPLETED","update_time":1},{"order_sn":"DP2","order_status":"COMPLETED","update_time":1}]}}`)
		case "/api/v2/payment/get_escrow_detail_batch":
			fmt.Fprint(w, `{"response":[{"order_sn":"DP1","escrow_detail":{"order_income":{"order_original_price":1,"escrow_amount":1}}},{"order_sn":"DP2","escrow_detail":{"order_income":{"order_original_price":2,"escrow_amount":2}}}]}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	client := NewShopeeClient(config.ShopeeAPIConfig{BaseURLShopee: srv.URL, PartnerID: "1", PartnerKey: "key"})
	client.httpClient = srv.Client()

	svc := NewReconcileService(nil, drop, nil, jrepo, nil, srepo, nil, nil, client, nil, nil, 5, nil)

	svc.processShopeeStatusBatch(context.Background(), "ShopA", []*models.DropshipPurchase{dp1, dp2})

	if len(drop.lookups) != 2 {
		t.Fatalf("expected 2 lookups, got %d", len(drop.lookups))
	}
	found1, found2 := false, false
	for _, v := range drop.lookups {
		if v == "INV1" {
			found1 = true
		}
		if v == "INV2" {
			found2 = true
		}
	}
	if !found1 || !found2 {
		t.Fatalf("unexpected lookups %v", drop.lookups)
	}
}

func TestUpdateShopeeStatuses_BatchHistory(t *testing.T) {
	dp := &models.DropshipPurchase{KodePesanan: "DP1", KodeInvoiceChannel: "INV1", NamaToko: "ShopA"}
	drop := &fakeDropRepoBatch{data: map[string]*models.DropshipPurchase{"INV1": dp}}

	now := time.Now()
	exp := 3600 * 24
	store := &models.Store{NamaToko: "ShopA", AccessToken: ptrString("tok"), RefreshToken: ptrString("ref"), ShopID: ptrString("2"), ExpireIn: &exp, LastUpdated: &now}
	srepo := &fakeStoreRepoBatch{store: store}
	jrepo := &fakeJournalRepoBatch{}
	batchSvc := &fakeBatchSvc{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/order/get_order_detail":
			fmt.Fprint(w, `{"response":{"order_list":[{"order_sn":"INV1","order_status":"COMPLETED","update_time":1}]}}`)
		case "/api/v2/payment/get_escrow_detail_batch":
			fmt.Fprint(w, `{"response":[{"order_sn":"INV1","escrow_detail":{"order_income":{"order_original_price":1,"escrow_amount":1}}}]}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	client := NewShopeeClient(config.ShopeeAPIConfig{BaseURLShopee: srv.URL, PartnerID: "1", PartnerKey: "key"})
	client.httpClient = srv.Client()

	svc := NewReconcileService(nil, drop, nil, jrepo, nil, srepo, nil, nil, client, batchSvc, nil, 5, nil)

	svc.UpdateShopeeStatuses(context.Background(), []string{"INV1"})

	if len(batchSvc.created) != 1 {
		t.Fatalf("expected 1 batch record, got %d", len(batchSvc.created))
	}
	if batchSvc.created[0].TotalData != 1 {
		t.Fatalf("unexpected TotalData %d", batchSvc.created[0].TotalData)
	}
	if len(batchSvc.updated) != 1 {
		t.Fatalf("expected UpdateDone to be called once, got %d", len(batchSvc.updated))
	}
}
