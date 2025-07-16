package service

import (
	"context"
	"time"
)

// PLSummary represents profit & loss summary data without database dependency
type PLSummary struct {
	ShopUsername      string    `json:"shop_username"`
	Period            string    `json:"period"`
	SumRevenue        float64   `json:"sum_revenue"`
	SumCOGS           float64   `json:"sum_cogs"`
	SumFees           float64   `json:"sum_fees"`
	NetProfit         float64   `json:"net_profit"`
	EndingCashBalance float64   `json:"ending_cash_balance"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type PLService struct {
	plReportSvc *ProfitLossReportService
}

func NewPLService(plReportSvc *ProfitLossReportService) *PLService {
	return &PLService{plReportSvc: plReportSvc}
}

func (s *PLService) ComputePL(ctx context.Context, shop, period string) (*PLSummary, error) {
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

	// Convert ProfitLoss to PLSummary format
	summary := &PLSummary{
		ShopUsername:      shop,
		Period:            period,
		SumRevenue:        pl.TotalPendapatanUsaha,
		SumCOGS:           pl.TotalHargaPokokPenjualan,
		SumFees:           pl.TotalBebanPemasaran,
		NetProfit:         pl.LabaRugiBersih.Amount,
		EndingCashBalance: 0, // This would require additional logic if needed
		UpdatedAt:         time.Now(),
	}
	return summary, nil
}
