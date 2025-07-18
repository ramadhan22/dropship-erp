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
	Label          string  `json:"label"`
	Amount         float64 `json:"amount"`
	Percent        float64 `json:"percent,omitempty"`
	PreviousAmount float64 `json:"previousAmount,omitempty"`
	Change         float64 `json:"change,omitempty"`
	ChangePercent  float64 `json:"changePercent,omitempty"`
	Manual         bool    `json:"manual,omitempty"`
	Indent         int     `json:"indent,omitempty"`
	Group          bool    `json:"group,omitempty"`
}

// ProfitLoss aggregates profit and loss data for a period.
type ProfitLoss struct {
	PendapatanUsaha          []ProfitLossRow `json:"pendapatanUsaha"`
	TotalPendapatanUsaha     float64         `json:"totalPendapatanUsaha"`
	PrevTotalPendapatanUsaha float64         `json:"prevTotalPendapatanUsaha,omitempty"`
	HargaPokokPenjualan      []ProfitLossRow `json:"hargaPokokPenjualan"`
	TotalHargaPokokPenjualan float64         `json:"totalHargaPokokPenjualan"`
	PrevTotalHargaPokokPenjualan float64     `json:"prevTotalHargaPokokPenjualan,omitempty"`
	LabaKotor                ProfitLossRow   `json:"labaKotor"`
	BebanOperasional         []ProfitLossRow `json:"bebanOperasional"`
	TotalBebanOperasional    float64         `json:"totalBebanOperasional"`
	PrevTotalBebanOperasional float64        `json:"prevTotalBebanOperasional,omitempty"`
	BebanPemasaran           []ProfitLossRow `json:"bebanPemasaran"`
	TotalBebanPemasaran      float64         `json:"totalBebanPemasaran"`
	PrevTotalBebanPemasaran  float64         `json:"prevTotalBebanPemasaran,omitempty"`
	BebanAdministrasi        []ProfitLossRow `json:"bebanAdministrasi"`
	TotalBebanAdministrasi   float64         `json:"totalBebanAdministrasi"`
	PrevTotalBebanAdministrasi float64       `json:"prevTotalBebanAdministrasi,omitempty"`
	TotalBebanUsaha          ProfitLossRow   `json:"totalBebanUsaha"`
	LabaSebelumPajak         float64         `json:"labaSebelumPajak"`
	PrevLabaSebelumPajak     float64         `json:"prevLabaSebelumPajak,omitempty"`
	PajakPenghasilan         []ProfitLossRow `json:"pajakPenghasilan"`
	TotalPajakPenghasilan    float64         `json:"totalPajakPenghasilan"`
	PrevTotalPajakPenghasilan float64        `json:"prevTotalPajakPenghasilan,omitempty"`
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
// typ should be "Monthly" or "Yearly". The returned range always spans the
// entire calendar month or year specified by the arguments.
// If comparison is true, it also fetches previous period data for comparison.
func (s *ProfitLossReportService) GetProfitLoss(ctx context.Context, typ string, month, year int, store string, comparison bool) (*ProfitLoss, error) {
	var start, end time.Time
	var prevStart, prevEnd time.Time
	
	switch typ {
	case "Yearly":
		// Ignore the month argument and return data for the entire year.
		start = time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
		end = start.AddDate(1, 0, 0).Add(-time.Nanosecond)
		if comparison {
			prevStart = time.Date(year-1, 1, 1, 0, 0, 0, 0, time.UTC)
			prevEnd = prevStart.AddDate(1, 0, 0).Add(-time.Nanosecond)
		}
	case "Monthly":
		if month == 0 {
			return nil, fmt.Errorf("month required")
		}
		start = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
		end = start.AddDate(0, 1, 0).Add(-time.Nanosecond)
		if comparison {
			prevStart = start.AddDate(0, -1, 0)
			prevEnd = start.Add(-time.Nanosecond)
		}
	default:
		return nil, fmt.Errorf("invalid type")
	}

	balances, err := s.jr.GetAccountBalancesBetween(ctx, store, start, end)
	if err != nil {
		return nil, err
	}

	var prevBalances []repository.AccountBalance
	if comparison {
		prevBalances, err = s.jr.GetAccountBalancesBetween(ctx, store, prevStart, prevEnd)
		if err != nil {
			return nil, err
		}
	}

	return s.buildProfitLoss(balances, prevBalances, comparison)
}

func pct(v, base float64) float64 {
	if base == 0 {
		return 0
	}
	return v / base * 100
}

// buildProfitLoss constructs a ProfitLoss struct from current and previous period balances.
func (s *ProfitLossReportService) buildProfitLoss(balances, prevBalances []repository.AccountBalance, comparison bool) (*ProfitLoss, error) {
	var revRows, hppRows, opRows []ProfitLossRow
	var adminRows, taxRows []ProfitLossRow
	var marketingRows []ProfitLossRow
	var totalRev, totalHPP, totalOp, totalMarketing, totalAdmin, totalTax float64

	// Build maps for previous period data
	prevRevMap := make(map[string]float64)
	prevHppMap := make(map[string]float64)
	prevOpMap := make(map[string]float64)
	prevMarketingMap := make(map[string]float64)
	prevAdminMap := make(map[string]float64)
	prevTaxMap := make(map[string]float64)

	var prevTotalRev, prevTotalHPP, prevTotalOp, prevTotalMarketing, prevTotalAdmin, prevTotalTax float64

	if comparison {
		for _, ab := range prevBalances {
			code := ab.AccountCode
			switch {
			case strings.HasPrefix(code, "4."):
				amt := -ab.Balance
				prevRevMap[ab.AccountName] = amt
				prevTotalRev += amt
			case strings.HasPrefix(code, "5.1"):
				prevHppMap[ab.AccountName] = ab.Balance
				prevTotalHPP += ab.Balance
			case code == "5.5":
				// skip parent marketing account balance
			case strings.HasPrefix(code, "5.5."):
				prevMarketingMap[ab.AccountName] = ab.Balance
				prevTotalMarketing += ab.Balance
			case strings.HasPrefix(code, "5.2"):
				prevOpMap[ab.AccountName] = ab.Balance
				prevTotalOp += ab.Balance
			case strings.HasPrefix(code, "5.3"):
				prevAdminMap[ab.AccountName] = ab.Balance
				prevTotalAdmin += ab.Balance
			case strings.HasPrefix(code, "5.4"):
				prevTaxMap[ab.AccountName] = ab.Balance
				prevTotalTax += ab.Balance
			}
		}
	}

	// Process current period data
	for _, ab := range balances {
		code := ab.AccountCode
		switch {
		case strings.HasPrefix(code, "4."):
			amt := -ab.Balance
			row := ProfitLossRow{Label: ab.AccountName, Amount: amt}
			if comparison {
				prevAmt := prevRevMap[ab.AccountName]
				row.PreviousAmount = prevAmt
				row.Change = amt - prevAmt
				row.ChangePercent = changePct(amt, prevAmt)
			}
			revRows = append(revRows, row)
			totalRev += amt
		case strings.HasPrefix(code, "5.1"):
			row := ProfitLossRow{Label: ab.AccountName, Amount: ab.Balance}
			if comparison {
				prevAmt := prevHppMap[ab.AccountName]
				row.PreviousAmount = prevAmt
				row.Change = ab.Balance - prevAmt
				row.ChangePercent = changePct(ab.Balance, prevAmt)
			}
			hppRows = append(hppRows, row)
			totalHPP += ab.Balance
		case code == "5.5":
			// skip parent marketing account balance
		case strings.HasPrefix(code, "5.5."):
			row := ProfitLossRow{Label: ab.AccountName, Amount: ab.Balance, Indent: 1}
			if comparison {
				prevAmt := prevMarketingMap[ab.AccountName]
				row.PreviousAmount = prevAmt
				row.Change = ab.Balance - prevAmt
				row.ChangePercent = changePct(ab.Balance, prevAmt)
			}
			marketingRows = append(marketingRows, row)
			totalMarketing += ab.Balance
		case strings.HasPrefix(code, "5.2"):
			row := ProfitLossRow{Label: ab.AccountName, Amount: ab.Balance}
			if comparison {
				prevAmt := prevOpMap[ab.AccountName]
				row.PreviousAmount = prevAmt
				row.Change = ab.Balance - prevAmt
				row.ChangePercent = changePct(ab.Balance, prevAmt)
			}
			opRows = append(opRows, row)
			totalOp += ab.Balance
		case strings.HasPrefix(code, "5.3"):
			row := ProfitLossRow{Label: ab.AccountName, Amount: ab.Balance}
			if comparison {
				prevAmt := prevAdminMap[ab.AccountName]
				row.PreviousAmount = prevAmt
				row.Change = ab.Balance - prevAmt
				row.ChangePercent = changePct(ab.Balance, prevAmt)
			}
			adminRows = append(adminRows, row)
			totalAdmin += ab.Balance
		case strings.HasPrefix(code, "5.4"):
			row := ProfitLossRow{Label: ab.AccountName, Amount: ab.Balance}
			if comparison {
				prevAmt := prevTaxMap[ab.AccountName]
				row.PreviousAmount = prevAmt
				row.Change = ab.Balance - prevAmt
				row.ChangePercent = changePct(ab.Balance, prevAmt)
			}
			taxRows = append(taxRows, row)
			totalTax += ab.Balance
		}
	}

	labaKotor := totalRev - totalHPP
	totalBebanUsahaAmt := totalOp + totalMarketing + totalAdmin
	labaSebelumPajak := totalRev - totalHPP - totalOp - totalMarketing - totalAdmin
	labaBersih := labaSebelumPajak - totalTax

	labaKotorRow := ProfitLossRow{Amount: labaKotor, Percent: pct(labaKotor, totalRev)}
	totalBebanUsahaRow := ProfitLossRow{Amount: totalBebanUsahaAmt, Percent: pct(totalBebanUsahaAmt, totalRev)}
	labaRugiBersihRow := ProfitLossRow{Amount: labaBersih, Percent: pct(labaBersih, totalRev)}

	if comparison {
		prevLabaKotor := prevTotalRev - prevTotalHPP
		prevTotalBebanUsahaAmt := prevTotalOp + prevTotalMarketing + prevTotalAdmin
		prevLabaSebelumPajak := prevTotalRev - prevTotalHPP - prevTotalOp - prevTotalMarketing - prevTotalAdmin
		prevLabaBersih := prevLabaSebelumPajak - prevTotalTax

		labaKotorRow.PreviousAmount = prevLabaKotor
		labaKotorRow.Change = labaKotor - prevLabaKotor
		labaKotorRow.ChangePercent = changePct(labaKotor, prevLabaKotor)

		totalBebanUsahaRow.PreviousAmount = prevTotalBebanUsahaAmt
		totalBebanUsahaRow.Change = totalBebanUsahaAmt - prevTotalBebanUsahaAmt
		totalBebanUsahaRow.ChangePercent = changePct(totalBebanUsahaAmt, prevTotalBebanUsahaAmt)

		labaRugiBersihRow.PreviousAmount = prevLabaBersih
		labaRugiBersihRow.Change = labaBersih - prevLabaBersih
		labaRugiBersihRow.ChangePercent = changePct(labaBersih, prevLabaBersih)
	}

	res := &ProfitLoss{
		PendapatanUsaha:          revRows,
		TotalPendapatanUsaha:     totalRev,
		HargaPokokPenjualan:      hppRows,
		TotalHargaPokokPenjualan: totalHPP,
		LabaKotor:                labaKotorRow,
		BebanOperasional:         opRows,
		TotalBebanOperasional:    totalOp,
		BebanPemasaran:           marketingRows,
		TotalBebanPemasaran:      totalMarketing,
		BebanAdministrasi:        adminRows,
		TotalBebanAdministrasi:   totalAdmin,
		TotalBebanUsaha:          totalBebanUsahaRow,
		LabaSebelumPajak:         labaSebelumPajak,
		PajakPenghasilan:         taxRows,
		TotalPajakPenghasilan:    totalTax,
		LabaRugiBersih:           labaRugiBersihRow,
	}

	if comparison {
		res.PrevTotalPendapatanUsaha = prevTotalRev
		res.PrevTotalHargaPokokPenjualan = prevTotalHPP
		res.PrevTotalBebanOperasional = prevTotalOp
		res.PrevTotalBebanPemasaran = prevTotalMarketing
		res.PrevTotalBebanAdministrasi = prevTotalAdmin
		res.PrevLabaSebelumPajak = prevTotalRev - prevTotalHPP - prevTotalOp - prevTotalMarketing - prevTotalAdmin
		res.PrevTotalPajakPenghasilan = prevTotalTax
	}

	return res, nil
}

// changePct calculates percentage change between current and previous values
func changePct(current, previous float64) float64 {
	if previous == 0 {
		if current == 0 {
			return 0
		}
		return 100 // or could return a special value to indicate new item
	}
	return ((current - previous) / previous) * 100
}
