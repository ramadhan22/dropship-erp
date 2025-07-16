package service

import (
	"context"
	"testing"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
)

type fakeTaxJournalRepo struct {
	balances []repository.AccountBalance
}

func (f *fakeTaxJournalRepo) GetAccountBalancesBetween(ctx context.Context, shop string, from, to time.Time) ([]repository.AccountBalance, error) {
	if f.balances == nil {
		// Return some revenue accounts with negative balances (since they're credit accounts)
		f.balances = []repository.AccountBalance{
			{AccountCode: "4.1", AccountName: "Revenue", Balance: -1000}, // Revenue should be negative
		}
	}
	return f.balances, nil
}

type fakeTaxRepo struct {
	paid map[string]time.Time
}

func (f *fakeTaxRepo) Get(ctx context.Context, store, pt, pv string) (*models.TaxPayment, error) {
	return nil, nil
}
func (f *fakeTaxRepo) Create(ctx context.Context, tp *models.TaxPayment) error { return nil }
func (f *fakeTaxRepo) MarkPaid(ctx context.Context, id string, paidAt time.Time) error {
	if f.paid == nil {
		f.paid = make(map[string]time.Time)
	}
	f.paid[id] = paidAt
	return nil
}

type fakeJournalRepoT struct {
	entries []*models.JournalEntry
	lines   []models.JournalLine
}

func (f *fakeJournalRepoT) CreateJournalEntry(ctx context.Context, e *models.JournalEntry) (int64, error) {
	e.JournalID = int64(len(f.entries) + 1)
	f.entries = append(f.entries, e)
	return e.JournalID, nil
}
func (f *fakeJournalRepoT) InsertJournalLine(ctx context.Context, l *models.JournalLine) error {
	f.lines = append(f.lines, *l)
	return nil
}
func (f *fakeJournalRepoT) ListJournalEntries(ctx context.Context, from, to, desc string) ([]models.JournalEntry, error) {
	return nil, nil
}
func (f *fakeJournalRepoT) GetJournalEntry(ctx context.Context, id int64) (*models.JournalEntry, error) {
	return nil, nil
}
func (f *fakeJournalRepoT) GetLinesByJournalID(ctx context.Context, id int64) ([]repository.JournalLineDetail, error) {
	return nil, nil
}
func (f *fakeJournalRepoT) ListEntriesBySourceID(ctx context.Context, sourceID string) ([]models.JournalEntry, error) {
	return nil, nil
}
func (f *fakeJournalRepoT) DeleteJournalEntry(ctx context.Context, id int64) error { return nil }

func TestComputeTax(t *testing.T) {
	taxJournalRepo := &fakeTaxJournalRepo{}
	svc := NewTaxService(nil, &fakeTaxRepo{}, &fakeJournalRepoT{}, taxJournalRepo)
	tp, err := svc.ComputeTax(context.Background(), "Store", "monthly", "2025-06")
	if err != nil {
		t.Fatalf("err %v", err)
	}
	if tp.TaxAmount != 5 {
		t.Errorf("expected 5 got %v", tp.TaxAmount)
	}
}

func TestPayTax(t *testing.T) {
	repo := &fakeTaxRepo{}
	jr := &fakeJournalRepoT{}
	taxJournalRepo := &fakeTaxJournalRepo{}
	svc := NewTaxService(nil, repo, jr, taxJournalRepo)
	tp := &models.TaxPayment{ID: "1", TaxAmount: 5, Store: "S"}
	if err := svc.PayTax(context.Background(), tp); err != nil {
		t.Fatalf("err %v", err)
	}
	if !tp.IsPaid || repo.paid["1"].IsZero() {
		t.Errorf("not marked paid")
	}
	if len(jr.entries) != 1 || len(jr.lines) != 2 {
		t.Errorf("journal not created")
	}
}
