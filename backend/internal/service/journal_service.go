package service

import (
	"context"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
)

type JournalService struct {
	repo *repository.JournalRepo
}

func NewJournalService(r *repository.JournalRepo) *JournalService { return &JournalService{repo: r} }

func (s *JournalService) List(ctx context.Context) ([]models.JournalEntry, error) {
	return s.repo.ListJournalEntries(ctx)
}

func (s *JournalService) Get(ctx context.Context, id int64) (*models.JournalEntry, error) {
	return s.repo.GetJournalEntry(ctx, id)
}

func (s *JournalService) Delete(ctx context.Context, id int64) error {
	return s.repo.DeleteJournalEntry(ctx, id)
}
