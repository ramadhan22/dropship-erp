package service

import (
	"context"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

type PLService struct {
	plReportSvc *ProfitLossReportService
}

func NewPLService(plReportSvc *ProfitLossReportService) *PLService {
	return &PLService{plReportSvc: plReportSvc}
}

func (s *PLService) ComputePL(ctx context.Context, shop, period string) (*models.CachedMetric, error) {
	// Parse period (e.g., "2025-01" -> year=2025, month=1)
	date, err := time.Parse("2006-01", period)
	if err != nil {
		return nil, err
	}
	year := date.Year()
	month := int(date.Month())

	// Get P&L from journal-based service
	pl, err := s.plReportSvc.GetProfitLoss(ctx, "Monthly", month, year, shop)
	if err != nil {
		return nil, err
	}

	// Convert ProfitLoss to CachedMetric format for backward compatibility
	cm := &models.CachedMetric{
		ShopUsername:      shop,
		Period:            period,
		SumRevenue:        pl.TotalPendapatanUsaha,
		SumCOGS:           pl.TotalHargaPokokPenjualan,
		SumFees:           pl.TotalBebanPemasaran,
		NetProfit:         pl.LabaRugiBersih.Amount,
		EndingCashBalance: 0, // This would require additional logic if needed
		UpdatedAt:         time.Now(),
	}
	return cm, nil
}
