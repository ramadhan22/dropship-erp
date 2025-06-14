package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

type PLService struct {
	metricRepo MetricRepoInterface
	metricSvc  *MetricService
}

func NewPLService(m MetricRepoInterface, svc *MetricService) *PLService {
	return &PLService{metricRepo: m, metricSvc: svc}
}

func (s *PLService) ComputePL(ctx context.Context, shop, period string) (*models.CachedMetric, error) {
	cm, err := s.metricRepo.GetCachedMetric(ctx, shop, period)
	if err != nil && errors.Is(err, sql.ErrNoRows) && s.metricSvc != nil {
		if err := s.metricSvc.CalculateAndCacheMetrics(ctx, shop, period); err != nil {
			return nil, err
		}
		cm, err = s.metricRepo.GetCachedMetric(ctx, shop, period)
	}
	return cm, err
}
