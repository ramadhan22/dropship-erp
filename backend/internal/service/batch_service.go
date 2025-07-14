package service

import (
	"context"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
)

// BatchService provides operations on batch_history.
type BatchService struct {
	repo       *repository.BatchRepo
	detailRepo *repository.BatchDetailRepo
}

func NewBatchService(r *repository.BatchRepo, d *repository.BatchDetailRepo) *BatchService {
	return &BatchService{repo: r, detailRepo: d}
}

func (s *BatchService) Create(ctx context.Context, b *models.BatchHistory) (int64, error) {
	return s.repo.Insert(ctx, b)
}

func (s *BatchService) UpdateDone(ctx context.Context, id int64, done int) error {
	return s.repo.UpdateDone(ctx, id, done)
}

// UpdateTotal sets the total number of rows expected for a batch process.
func (s *BatchService) UpdateTotal(ctx context.Context, id int64, total int) error {
	return s.repo.UpdateTotal(ctx, id, total)
}

func (s *BatchService) UpdateStatus(ctx context.Context, id int64, status, msg string) error {
	return s.repo.UpdateStatus(ctx, id, status, msg)
}

func (s *BatchService) List(ctx context.Context) ([]models.BatchHistory, error) {
	return s.repo.List(ctx)
}

func (s *BatchService) CreateDetail(ctx context.Context, d *models.BatchHistoryDetail) error {
	if s.detailRepo == nil {
		return nil
	}
	return s.detailRepo.Insert(ctx, d)
}

func (s *BatchService) ListDetails(ctx context.Context, batchID int64) ([]models.BatchHistoryDetail, error) {
	if s.detailRepo == nil {
		return []models.BatchHistoryDetail{}, nil
	}
	return s.detailRepo.ListByBatchID(ctx, batchID)
}

func (s *BatchService) UpdateDetailStatus(ctx context.Context, id int64, status, msg string) error {
	if s.detailRepo == nil {
		return nil
	}
	return s.detailRepo.UpdateStatus(ctx, id, status, msg)
}

// ListPendingByType returns batches with the given process type and status 'pending'.
func (s *BatchService) ListPendingByType(ctx context.Context, typ string) ([]models.BatchHistory, error) {
	return s.repo.ListByProcessAndStatus(ctx, typ, "pending")
}
