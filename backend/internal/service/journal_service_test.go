package service

import (
	"context"
	"testing"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
)

type fakeJournalRepo struct {
	entry *models.JournalEntry
	lines []models.JournalLine
	next  int64
}

func (f *fakeJournalRepo) CreateJournalEntry(ctx context.Context, e *models.JournalEntry) (int64, error) {
	f.next++
	e.JournalID = f.next
	f.entry = e
	return f.next, nil
}

func (f *fakeJournalRepo) InsertJournalLine(ctx context.Context, l *models.JournalLine) error {
	cp := *l
	f.lines = append(f.lines, cp)
	return nil
}

func (f *fakeJournalRepo) ListJournalEntries(ctx context.Context, from, to, desc string) ([]models.JournalEntry, error) {
	return nil, nil
}

func (f *fakeJournalRepo) GetJournalEntry(ctx context.Context, id int64) (*models.JournalEntry, error) {
	return nil, nil
}

func (f *fakeJournalRepo) DeleteJournalEntry(ctx context.Context, id int64) error {
	return nil
}

func (f *fakeJournalRepo) GetLinesByJournalID(ctx context.Context, id int64) ([]repository.JournalLineDetail, error) {
	return nil, nil
}

func (f *fakeJournalRepo) ListEntriesBySourceID(ctx context.Context, sourceID string) ([]models.JournalEntry, error) {
	return nil, nil
}

func TestJournalServiceCreate_Balance(t *testing.T) {
	repo := &fakeJournalRepo{}
	svc := NewJournalService(nil, repo)

	entry := &models.JournalEntry{SourceType: "manual", SourceID: "1"}
	lines := []models.JournalLine{
		{AccountID: 1, IsDebit: true, Amount: 50},
		{AccountID: 2, IsDebit: false, Amount: 50},
	}
	id, err := svc.Create(context.Background(), entry, lines)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if id == 0 || repo.entry == nil || len(repo.lines) != 2 {
		t.Fatalf("repo not called correctly")
	}
}

func TestJournalServiceCreate_Unbalanced(t *testing.T) {
	repo := &fakeJournalRepo{}
	svc := NewJournalService(nil, repo)

	entry := &models.JournalEntry{SourceType: "manual", SourceID: "1"}
	lines := []models.JournalLine{
		{AccountID: 1, IsDebit: true, Amount: 50},
		{AccountID: 2, IsDebit: false, Amount: 40},
	}
	if _, err := svc.Create(context.Background(), entry, lines); err == nil {
		t.Fatalf("expected error for unbalanced entry")
	}
}
