package service

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
)

type JournalRepoInterface interface {
	CreateJournalEntry(ctx context.Context, e *models.JournalEntry) (int64, error)
	InsertJournalLine(ctx context.Context, l *models.JournalLine) error
	ListJournalEntries(ctx context.Context, from, to, desc string) ([]models.JournalEntry, error)
	GetJournalEntry(ctx context.Context, id int64) (*models.JournalEntry, error)
	GetLinesByJournalID(ctx context.Context, id int64) ([]repository.JournalLineDetail, error)
	ListEntriesBySourceID(ctx context.Context, sourceID string) ([]models.JournalEntry, error)
	DeleteJournalEntry(ctx context.Context, id int64) error
}

type JournalService struct {
	db   *sqlx.DB
	repo JournalRepoInterface
}

func NewJournalService(db *sqlx.DB, r JournalRepoInterface) *JournalService {
	return &JournalService{db: db, repo: r}
}

func (s *JournalService) List(ctx context.Context, from, to, desc string) ([]models.JournalEntry, error) {
	return s.repo.ListJournalEntries(ctx, from, to, desc)
}

func (s *JournalService) Get(ctx context.Context, id int64) (*models.JournalEntry, error) {
	return s.repo.GetJournalEntry(ctx, id)
}

func (s *JournalService) Delete(ctx context.Context, id int64) error {
	return s.repo.DeleteJournalEntry(ctx, id)
}

func (s *JournalService) Lines(ctx context.Context, id int64) ([]repository.JournalLineDetail, error) {
	return s.repo.GetLinesByJournalID(ctx, id)
}

// EntryWithLines bundles a journal entry with its lines.
type EntryWithLines struct {
	Entry models.JournalEntry            `json:"entry"`
	Lines []repository.JournalLineDetail `json:"lines"`
}

// LinesBySource returns all journal entries matching the source ID along with their lines.
func (s *JournalService) LinesBySource(ctx context.Context, sourceID string) ([]EntryWithLines, error) {
	entries, err := s.repo.ListEntriesBySourceID(ctx, sourceID)
	if err != nil {
		return nil, err
	}
	result := make([]EntryWithLines, 0, len(entries))
	for _, e := range entries {
		lines, err := s.repo.GetLinesByJournalID(ctx, e.JournalID)
		if err != nil {
			return nil, err
		}
		result = append(result, EntryWithLines{Entry: e, Lines: lines})
	}
	return result, nil
}

// Create inserts a JournalEntry along with its lines. It ensures total debits
// equal total credits otherwise returns an error.
func (s *JournalService) Create(
	ctx context.Context,
	e *models.JournalEntry,
	lines []models.JournalLine,
) (int64, error) {
	var debit, credit float64
	for _, l := range lines {
		if l.IsDebit {
			debit += l.Amount
		} else {
			credit += l.Amount
		}
	}
	if debit != credit {
		return 0, fmt.Errorf("debits %.2f do not equal credits %.2f", debit, credit)
	}
	if s.db == nil {
		id, err := s.repo.CreateJournalEntry(ctx, e)
		if err != nil {
			return 0, err
		}
		for i := range lines {
			lines[i].JournalID = id
			if err := s.repo.InsertJournalLine(ctx, &lines[i]); err != nil {
				return 0, err
			}
		}
		return id, nil
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	repoTx := repository.NewJournalRepo(tx)
	id, err := repoTx.CreateJournalEntry(ctx, e)
	if err != nil {
		return 0, err
	}
	for i := range lines {
		lines[i].JournalID = id
		if err := repoTx.InsertJournalLine(ctx, &lines[i]); err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}
