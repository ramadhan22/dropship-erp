package service

import (
	"context"
	"testing"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

type fakePLComputer struct{ cm *models.CachedMetric }

func (f *fakePLComputer) ComputePL(ctx context.Context, shop, period string) (*models.CachedMetric, error) {
	return f.cm, nil
}

func TestProfitLossReportService_GetProfitLoss(t *testing.T) {
	fake := &fakePLComputer{cm: &models.CachedMetric{SumRevenue: 100, SumCOGS: 60, SumFees: 10, NetProfit: 30}}
	svc := NewProfitLossReportService(fake)

	pl, err := svc.GetProfitLoss(context.Background(), "Monthly", 5, 2025, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pl.TotalPendapatanUsaha != 100 {
		t.Errorf("got %f want 100", pl.TotalPendapatanUsaha)
	}
	if pl.LabaRugiBersih.Amount != 30 {
		t.Errorf("got %f want 30", pl.LabaRugiBersih.Amount)
	}
}
