// File: backend/internal/service/reconcile_service_test.go

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// fake implementations for each repo interface:

type fakeDropRepoRec struct {
	data map[string]*models.DropshipPurchase
}

func (f *fakeDropRepoRec) GetDropshipPurchaseByInvoice(ctx context.Context, invoice string) (*models.DropshipPurchase, error) {
	if dp, ok := f.data[invoice]; ok {
		return dp, nil
	}
	return nil, errors.New("not found")
}

func (f *fakeDropRepoRec) GetDropshipPurchaseByID(ctx context.Context, kode string) (*models.DropshipPurchase, error) {
	if dp, ok := f.data[kode]; ok {
		return dp, nil
	}
	return nil, errors.New("not found")
}

func (f *fakeDropRepoRec) UpdatePurchaseStatus(ctx context.Context, kode, status string) error {
	if dp, ok := f.data[kode]; ok {
		dp.StatusPesananTerakhir = status
		return nil
	}
	return errors.New("not found")
}

func (f *fakeDropRepoRec) SumDetailByInvoice(ctx context.Context, inv string) (float64, error) {
	if dp, ok := f.data[inv]; ok {
		return dp.TotalTransaksi, nil
	}
	return 0, nil
}

func (f *fakeDropRepoRec) SumProductCostByInvoice(ctx context.Context, inv string) (float64, error) {
	if dp, ok := f.data[inv]; ok {
		return dp.TotalTransaksi, nil
	}
	return 0, nil
}

type fakeShopeeRepoRec struct {
	data map[string]*models.ShopeeSettledOrder
}

func (f *fakeShopeeRepoRec) GetShopeeOrderByID(ctx context.Context, orderID string) (*models.ShopeeSettledOrder, error) {
	if so, ok := f.data[orderID]; ok {
		return so, nil
	}
	return nil, errors.New("not found")
}

func (f *fakeShopeeRepoRec) ExistsShopeeSettled(ctx context.Context, no string) (bool, error) {
	_, ok := f.data[no]
	return ok, nil
}

type fakeJournalRepoRec struct {
	entries []*models.JournalEntry
	lines   []*models.JournalLine
	nextID  int64
}

func (f *fakeJournalRepoRec) CreateJournalEntry(ctx context.Context, e *models.JournalEntry) (int64, error) {
	f.nextID++
	e.JournalID = f.nextID
	f.entries = append(f.entries, e)
	return f.nextID, nil
}
func (f *fakeJournalRepoRec) InsertJournalLine(ctx context.Context, l *models.JournalLine) error {
	f.lines = append(f.lines, l)
	return nil
}
func (f *fakeJournalRepoRec) InsertJournalLines(ctx context.Context, lines []models.JournalLine) error {
	for i := range lines {
		f.lines = append(f.lines, &lines[i])
	}
	return nil
}

type fakeRecRepoRec struct {
	inserted []*models.ReconciledTransaction
}

func (f *fakeRecRepoRec) InsertReconciledTransaction(ctx context.Context, r *models.ReconciledTransaction) error {
	f.inserted = append(f.inserted, r)
	return nil
}

type fakeDetailRepo struct {
	saved []*models.ShopeeOrderDetailRow
}

func (f *fakeDetailRepo) SaveOrderDetail(ctx context.Context, d *models.ShopeeOrderDetailRow, items []models.ShopeeOrderItemRow, pkgs []models.ShopeeOrderPackageRow) error {
	f.saved = append(f.saved, d)
	return nil
}

func (f *fakeDetailRepo) UpdateOrderDetailStatus(ctx context.Context, sn, status, orderStatus string, updateTime time.Time) error {
	return nil
}

func (f *fakeDetailRepo) GetOrderDetail(ctx context.Context, sn string) (*models.ShopeeOrderDetailRow, []models.ShopeeOrderItemRow, []models.ShopeeOrderPackageRow, error) {
	return nil, nil, nil, errors.New("not found")
}

type fakeFailedRecRepo struct {
	failures []models.FailedReconciliation
}

func (f *fakeFailedRecRepo) InsertFailedReconciliation(ctx context.Context, failed *models.FailedReconciliation) error {
	f.failures = append(f.failures, *failed)
	return nil
}

func (f *fakeFailedRecRepo) GetFailedReconciliationsByShop(ctx context.Context, shop string, limit, offset int) ([]models.FailedReconciliation, error) {
	return f.failures, nil
}

func (f *fakeFailedRecRepo) GetFailedReconciliationsByBatch(ctx context.Context, batchID int64) ([]models.FailedReconciliation, error) {
	return f.failures, nil
}

func (f *fakeFailedRecRepo) CountFailedReconciliationsByErrorType(ctx context.Context, shop string, since time.Time) (map[string]int, error) {
	counts := make(map[string]int)
	for _, fail := range f.failures {
		if fail.Shop == shop && fail.FailedAt.After(since) {
			counts[fail.ErrorType]++
		}
	}
	return counts, nil
}

func (f *fakeFailedRecRepo) MarkAsRetried(ctx context.Context, id int64) error {
	return nil
}

func (f *fakeFailedRecRepo) GetUnretriedFailedReconciliations(ctx context.Context, shop string, limit int) ([]models.FailedReconciliation, error) {
	return f.failures, nil
}

func TestMatchAndJournal_Success(t *testing.T) {
	ctx := context.Background()

	// Prepare fake repos with preloaded data
	fDrop := &fakeDropRepoRec{
		data: map[string]*models.DropshipPurchase{
			"DP-111": {KodePesanan: "DP-111", TotalTransaksi: 50.00},
		},
	}
	fShopee := &fakeShopeeRepoRec{
		data: map[string]*models.ShopeeSettledOrder{
			"SO-222": {OrderID: "SO-222", NetIncome: 80.00, SettledDate: time.Now()},
		},
	}
	fJournal := &fakeJournalRepoRec{nextID: 0}
	fRec := &fakeRecRepoRec{}
	fDetail := &fakeDetailRepo{}
	fFailed := &fakeFailedRecRepo{}

	svc := NewReconcileService(nil, fDrop, fShopee, fJournal, fRec, nil, fDetail, nil, nil, nil, fFailed, nil, 5, nil)
	err := svc.MatchAndJournal(ctx, "DP-111", "SO-222", "ShopA")
	if err != nil {
		t.Fatalf("MatchAndJournal error: %v", err)
	}

	// Verify one JournalEntry was created
	if len(fJournal.entries) != 1 {
		t.Errorf("expected 1 JournalEntry, got %d", len(fJournal.entries))
	}
	je := fJournal.entries[0]
	if je.SourceType != "reconcile" || je.SourceID != "SO-222" {
		t.Errorf("unexpected JournalEntry: %+v", je)
	}

	// Verify two JournalLines
	if len(fJournal.lines) != 2 {
		t.Errorf("expected 2 JournalLines, got %d", len(fJournal.lines))
	}
	// Check debit (COGS) and credit (Cash)
	var foundDebit, foundCredit bool
	for _, l := range fJournal.lines {
		if l.AccountID == 5001 && l.IsDebit && l.Amount == 50.00 {
			foundDebit = true
		}
		if l.AccountID == 1001 && !l.IsDebit && l.Amount == 80.00 {
			foundCredit = true
		}
	}
	if !foundDebit || !foundCredit {
		t.Errorf("did not find correct debit/credit lines: %+v", fJournal.lines)
	}

	// Verify a ReconciledTransaction was inserted
	if len(fRec.inserted) != 1 {
		t.Errorf("expected 1 ReconciledTransaction, got %d", len(fRec.inserted))
	}
	rt := fRec.inserted[0]
	if rt.DropshipID == nil || *rt.DropshipID != "DP-111" || rt.ShopeeID == nil || *rt.ShopeeID != "SO-222" {
		t.Errorf("unexpected ReconciledTransaction: %+v", rt)
	}
}
