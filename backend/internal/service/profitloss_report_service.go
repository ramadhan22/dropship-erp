package service

import (
	"context"
	"fmt"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// ProfitLossRow represents a line in the profit loss statement.
type ProfitLossRow struct {
	Label   string  `json:"label"`
	Amount  float64 `json:"amount"`
	Percent float64 `json:"percent,omitempty"`
	Manual  bool    `json:"manual,omitempty"`
}

// ProfitLoss aggregates profit and loss data for a period.
type ProfitLoss struct {
	PendapatanUsaha          []ProfitLossRow `json:"pendapatanUsaha"`
	TotalPendapatanUsaha     float64         `json:"totalPendapatanUsaha"`
	HargaPokokPenjualan      []ProfitLossRow `json:"hargaPokokPenjualan"`
	TotalHargaPokokPenjualan float64         `json:"totalHargaPokokPenjualan"`
	LabaKotor                ProfitLossRow   `json:"labaKotor"`
	BebanOperasional         []ProfitLossRow `json:"bebanOperasional"`
	TotalBebanOperasional    float64         `json:"totalBebanOperasional"`
	BebanAdministrasi        []ProfitLossRow `json:"bebanAdministrasi"`
	TotalBebanAdministrasi   float64         `json:"totalBebanAdministrasi"`
	TotalBebanUsaha          ProfitLossRow   `json:"totalBebanUsaha"`
	LabaSebelumPajak         float64         `json:"labaSebelumPajak"`
	PajakPenghasilan         []ProfitLossRow `json:"pajakPenghasilan"`
	TotalPajakPenghasilan    float64         `json:"totalPajakPenghasilan"`
	LabaRugiBersih           ProfitLossRow   `json:"labaRugiBersih"`
}

// ProfitLossReportService computes profit and loss reports using CachedMetric data.
type plComputer interface {
	ComputePL(ctx context.Context, shop, period string) (*models.CachedMetric, error)
}

type ProfitLossReportService struct {
	pl plComputer
}

// NewProfitLossReportService constructs a ProfitLossReportService.
func NewProfitLossReportService(pl plComputer) *ProfitLossReportService {
	return &ProfitLossReportService{pl: pl}
}

// GetProfitLoss returns profit and loss information for the given period.
// typ should be "Monthly" or "Yearly". Month may be ignored for yearly reports.
func (s *ProfitLossReportService) GetProfitLoss(ctx context.Context, typ string, month, year int, store string) (*ProfitLoss, error) {
	var metric models.CachedMetric
	switch typ {
	case "Yearly":
		for m := 1; m <= 12; m++ {
			per := fmt.Sprintf("%04d-%02d", year, m)
			cm, err := s.pl.ComputePL(ctx, store, per)
			if err != nil {
				return nil, err
			}
			metric.SumRevenue += cm.SumRevenue
			metric.SumCOGS += cm.SumCOGS
			metric.SumFees += cm.SumFees
		}
		metric.NetProfit = metric.SumRevenue - metric.SumCOGS - metric.SumFees
	default:
		if month == 0 {
			return nil, fmt.Errorf("month required")
		}
		per := fmt.Sprintf("%04d-%02d", year, month)
		cm, err := s.pl.ComputePL(ctx, store, per)
		if err != nil {
			return nil, err
		}
		metric = *cm
	}

	rev := metric.SumRevenue
	cogs := metric.SumCOGS
	fees := metric.SumFees
	labaKotor := rev - cogs
	labaBersih := metric.NetProfit

	res := &ProfitLoss{
		PendapatanUsaha:          []ProfitLossRow{{Label: "Penjualan", Amount: rev}},
		TotalPendapatanUsaha:     rev,
		HargaPokokPenjualan:      []ProfitLossRow{{Label: "HPP", Amount: cogs}},
		TotalHargaPokokPenjualan: cogs,
		LabaKotor:                ProfitLossRow{Amount: labaKotor, Percent: pct(labaKotor, rev)},
		BebanOperasional:         []ProfitLossRow{{Label: "Marketplace Fees", Amount: fees}},
		TotalBebanOperasional:    fees,
		BebanAdministrasi:        []ProfitLossRow{},
		TotalBebanAdministrasi:   0,
		TotalBebanUsaha:          ProfitLossRow{Amount: fees, Percent: pct(fees, rev)},
		LabaSebelumPajak:         labaBersih,
		PajakPenghasilan:         []ProfitLossRow{},
		TotalPajakPenghasilan:    0,
		LabaRugiBersih:           ProfitLossRow{Amount: labaBersih, Percent: pct(labaBersih, rev)},
	}
	return res, nil
}

func pct(v, base float64) float64 {
	if base == 0 {
		return 0
	}
	return v / base * 100
}
