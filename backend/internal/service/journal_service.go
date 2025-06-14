package service

import (
	"context"
	"fmt"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
)

type JournalRepoInterface interface {
	CreateJournalEntry(ctx context.Context, e *models.JournalEntry) (int64, error)
	InsertJournalLine(ctx context.Context, l *models.JournalLine) error
	ListJournalEntries(ctx context.Context) ([]models.JournalEntry, error)
	GetJournalEntry(ctx context.Context, id int64) (*models.JournalEntry, error)
	GetLinesByJournalID(ctx context.Context, id int64) ([]repository.JournalLineDetail, error)
	DeleteJournalEntry(ctx context.Context, id int64) error
}

type JournalService struct {
	repo JournalRepoInterface
}

func NewJournalService(r JournalRepoInterface) *JournalService { return &JournalService{repo: r} }

func (s *JournalService) List(ctx context.Context) ([]models.JournalEntry, error) {
	return s.repo.ListJournalEntries(ctx)
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
