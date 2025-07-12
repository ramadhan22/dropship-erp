package service

import (
	"context"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
)

// BatchService provides operations on batch_history.
type BatchService struct{ repo *repository.BatchRepo }

func NewBatchService(r *repository.BatchRepo) *BatchService { return &BatchService{repo: r} }

func (s *BatchService) Create(ctx context.Context, b *models.BatchHistory) (int64, error) {
	return s.repo.Insert(ctx, b)
}

func (s *BatchService) UpdateDone(ctx context.Context, id int64, done int) error {
	return s.repo.UpdateDone(ctx, id, done)
}

func (s *BatchService) UpdateStatus(ctx context.Context, id int64, status, msg string) error {
	return s.repo.UpdateStatus(ctx, id, status, msg)
}

func (s *BatchService) List(ctx context.Context) ([]models.BatchHistory, error) {
	return s.repo.List(ctx)
}
