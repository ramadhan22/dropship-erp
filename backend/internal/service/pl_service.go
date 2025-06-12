package service

import (
	"context"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

type PLService struct {
	metricRepo MetricRepoInterface
}

func NewPLService(m MetricRepoInterface) *PLService { return &PLService{metricRepo: m} }

func (s *PLService) ComputePL(ctx context.Context, shop, period string) (*models.CachedMetric, error) {
	return s.metricRepo.GetCachedMetric(ctx, shop, period)
}
