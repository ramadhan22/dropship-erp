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

func (f *fakeDropRepoRec) GetDropshipPurchaseByID(ctx context.Context, purchaseID string) (*models.DropshipPurchase, error) {
	if dp, ok := f.data[purchaseID]; ok {
		return dp, nil
	}
	return nil, errors.New("not found")
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

type fakeRecRepoRec struct {
	inserted []*models.ReconciledTransaction
}

func (f *fakeRecRepoRec) InsertReconciledTransaction(ctx context.Context, r *models.ReconciledTransaction) error {
	f.inserted = append(f.inserted, r)
	return nil
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

	svc := NewReconcileService(fDrop, fShopee, fJournal, fRec)
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
