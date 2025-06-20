package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
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

// ProfitLossJournalRepo defines the journal repo method needed for PL reports.
type ProfitLossJournalRepo interface {
	GetAccountBalancesBetween(ctx context.Context, shop string, from, to time.Time) ([]repository.AccountBalance, error)
}

// ProfitLossReportService computes profit and loss data using journal entries.
type ProfitLossReportService struct {
	jr ProfitLossJournalRepo
}

// NewProfitLossReportService constructs a ProfitLossReportService.
func NewProfitLossReportService(jr ProfitLossJournalRepo) *ProfitLossReportService {
	return &ProfitLossReportService{jr: jr}
}

// GetProfitLoss returns profit and loss information for the given period.
// typ should be "Monthly" or "Yearly". If typ is "Yearly" and month > 0,
// the report is calculated year-to-date up to the end of the given month.
func (s *ProfitLossReportService) GetProfitLoss(ctx context.Context, typ string, month, year int, store string) (*ProfitLoss, error) {
	var start, end time.Time
	switch typ {
	case "Yearly":
		start = time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
		if month > 0 && month <= 12 {
			end = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC).
				AddDate(0, 1, 0).Add(-time.Nanosecond)
		} else {
			end = start.AddDate(1, 0, 0).Add(-time.Nanosecond)
		}
	case "Monthly":
		if month == 0 {
			return nil, fmt.Errorf("month required")
		}
		start = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
		end = start.AddDate(0, 1, 0).Add(-time.Nanosecond)
	default:
		return nil, fmt.Errorf("invalid type")
	}

	balances, err := s.jr.GetAccountBalancesBetween(ctx, store, start, end)
	if err != nil {
		return nil, err
	}

	var revRows, hppRows, opRows, adminRows, taxRows []ProfitLossRow
	var totalRev, totalHPP, totalOp, totalAdmin, totalTax float64

	for _, ab := range balances {
		code := ab.AccountCode
		switch {
		case strings.HasPrefix(code, "4."):
			amt := -ab.Balance
			revRows = append(revRows, ProfitLossRow{Label: ab.AccountName, Amount: amt})
			totalRev += amt
		case strings.HasPrefix(code, "5.1"):
			hppRows = append(hppRows, ProfitLossRow{Label: ab.AccountName, Amount: ab.Balance})
			totalHPP += ab.Balance
		case strings.HasPrefix(code, "5.2"):
			opRows = append(opRows, ProfitLossRow{Label: ab.AccountName, Amount: ab.Balance})
			totalOp += ab.Balance
		case strings.HasPrefix(code, "5.3"):
			adminRows = append(adminRows, ProfitLossRow{Label: ab.AccountName, Amount: ab.Balance})
			totalAdmin += ab.Balance
		case strings.HasPrefix(code, "5.4"):
			taxRows = append(taxRows, ProfitLossRow{Label: ab.AccountName, Amount: ab.Balance})
			totalTax += ab.Balance
		}
	}

	labaKotor := totalRev - totalHPP
	totalBebanUsahaAmt := totalOp + totalAdmin
	labaSebelumPajak := totalRev - totalHPP - totalOp - totalAdmin
	labaBersih := labaSebelumPajak - totalTax

	res := &ProfitLoss{
		PendapatanUsaha:          revRows,
		TotalPendapatanUsaha:     totalRev,
		HargaPokokPenjualan:      hppRows,
		TotalHargaPokokPenjualan: totalHPP,
		LabaKotor:                ProfitLossRow{Amount: labaKotor, Percent: pct(labaKotor, totalRev)},
		BebanOperasional:         opRows,
		TotalBebanOperasional:    totalOp,
		BebanAdministrasi:        adminRows,
		TotalBebanAdministrasi:   totalAdmin,
		TotalBebanUsaha:          ProfitLossRow{Amount: totalBebanUsahaAmt, Percent: pct(totalBebanUsahaAmt, totalRev)},
		LabaSebelumPajak:         labaSebelumPajak,
		PajakPenghasilan:         taxRows,
		TotalPajakPenghasilan:    totalTax,
		LabaRugiBersih:           ProfitLossRow{Amount: labaBersih, Percent: pct(labaBersih, totalRev)},
	}
	return res, nil
}

func pct(v, base float64) float64 {
	if base == 0 {
		return 0
	}
	return v / base * 100
}
