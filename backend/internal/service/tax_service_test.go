package service

import (
	"context"
	"testing"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
)

type fakeMetricSvc struct{ rev float64 }

func (f *fakeMetricSvc) GetRevenue(ctx context.Context, store, pt, pv string) (float64, error) {
	return f.rev, nil
}

func (f *fakeMetricSvc) CalculateAndCacheMetrics(ctx context.Context, shop, period string) error {
	return nil
}
func (f *fakeMetricSvc) MetricRepo() MetricRepoInterface { return nil }

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
func (f *fakeJournalRepoT) InsertJournalLines(ctx context.Context, lines []models.JournalLine) error {
	for _, l := range lines {
		f.lines = append(f.lines, l)
	}
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
	svc := NewTaxService(nil, &fakeTaxRepo{}, &fakeJournalRepoT{}, &fakeMetricSvc{rev: 1000})
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
	svc := NewTaxService(nil, repo, jr, &fakeMetricSvc{})
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
