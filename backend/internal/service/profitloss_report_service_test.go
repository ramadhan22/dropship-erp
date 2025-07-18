package service

import (
	"context"
	"testing"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
)

type fakeJournalRepoPL struct {
	balances []repository.AccountBalance
	lastFrom time.Time
	lastTo   time.Time
}

func (f *fakeJournalRepoPL) GetAccountBalancesBetween(ctx context.Context, shop string, from, to time.Time) ([]repository.AccountBalance, error) {
	f.lastFrom = from
	f.lastTo = to
	return f.balances, nil
}

func TestProfitLossReportService_GetProfitLoss(t *testing.T) {
	repo := &fakeJournalRepoPL{balances: []repository.AccountBalance{
		{AccountCode: "4.1", AccountName: "Penjualan", Balance: -200},
		{AccountCode: "5.1", AccountName: "HPP", Balance: 100},
		{AccountCode: "5.5.1", AccountName: "Voucher", Balance: 5},
	}}
	svc := NewProfitLossReportService(repo)

	pl, err := svc.GetProfitLoss(context.Background(), "Monthly", 5, 2025, "ShopX", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pl.TotalPendapatanUsaha != 200 {
		t.Errorf("got %f want 200", pl.TotalPendapatanUsaha)
	}
	if pl.LabaRugiBersih.Amount != 95 {
		t.Errorf("got %f want 95", pl.LabaRugiBersih.Amount)
	}
	if len(pl.BebanPemasaran) != 1 {
		t.Errorf("expected 1 pemasaran row, got %d", len(pl.BebanPemasaran))
	}
	if pl.BebanPemasaran[0].Label != "Voucher" {
		t.Errorf("got label %s want Voucher", pl.BebanPemasaran[0].Label)
	}
	if pl.BebanPemasaran[0].Indent != 1 {
		t.Errorf("expected voucher indent 1, got %d", pl.BebanPemasaran[0].Indent)
	}
}

func TestProfitLossReportService_SkipMarketingParentAccount(t *testing.T) {
	repo := &fakeJournalRepoPL{balances: []repository.AccountBalance{
		{AccountCode: "4.1", AccountName: "Penjualan", Balance: -100},
		{AccountCode: "5.1", AccountName: "HPP", Balance: 50},
		{AccountCode: "5.5", AccountName: "Beban Pemasaran", Balance: 10},
		{AccountCode: "5.5.1", AccountName: "Voucher", Balance: 5},
	}}
	svc := NewProfitLossReportService(repo)

	pl, err := svc.GetProfitLoss(context.Background(), "Monthly", 5, 2025, "ShopX", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(pl.BebanPemasaran) != 1 {
		t.Fatalf("expected 1 pemasaran row, got %d", len(pl.BebanPemasaran))
	}
	if pl.BebanPemasaran[0].Label != "Voucher" {
		t.Errorf("got label %s want Voucher", pl.BebanPemasaran[0].Label)
	}
	if pl.BebanPemasaran[0].Indent != 1 {
		t.Errorf("expected voucher indent 1, got %d", pl.BebanPemasaran[0].Indent)
	}
}

func TestProfitLossReportService_GetProfitLoss_MonthlyPeriod(t *testing.T) {
	repo := &fakeJournalRepoPL{balances: []repository.AccountBalance{}}
	svc := NewProfitLossReportService(repo)

	_, err := svc.GetProfitLoss(context.Background(), "Monthly", 3, 2025, "ShopX", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantStart := time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)
	wantEnd := wantStart.AddDate(0, 1, 0).Add(-time.Nanosecond)
	if !repo.lastFrom.Equal(wantStart) {
		t.Errorf("from date got %v want %v", repo.lastFrom, wantStart)
	}
	if !repo.lastTo.Equal(wantEnd) {
		t.Errorf("to date got %v want %v", repo.lastTo, wantEnd)
	}
}

func TestProfitLossReportService_GetProfitLoss_Yearly(t *testing.T) {
	repo := &fakeJournalRepoPL{balances: []repository.AccountBalance{}}
	svc := NewProfitLossReportService(repo)

	_, err := svc.GetProfitLoss(context.Background(), "Yearly", 5, 2025, "ShopX", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantStart := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	wantEnd := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(1, 0, 0).Add(-time.Nanosecond)
	if !repo.lastFrom.Equal(wantStart) {
		t.Errorf("from date got %v want %v", repo.lastFrom, wantStart)
	}
	if !repo.lastTo.Equal(wantEnd) {
		t.Errorf("to date got %v want %v", repo.lastTo, wantEnd)
	}
}

// fakeJournalRepoPLMap allows returning different balances for each date key.
type fakeJournalRepoPLMap struct {
	data map[string][]repository.AccountBalance
}

func (f *fakeJournalRepoPLMap) GetAccountBalancesBetween(ctx context.Context, shop string, from, to time.Time) ([]repository.AccountBalance, error) {
	key := from.Format("2006-01-02") + "_" + to.Format("2006-01-02")
	return f.data[key], nil
}

func TestProfitLossReportService_DifferentPeriods(t *testing.T) {
	startMay := time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC)
	endMay := startMay.AddDate(0, 1, 0).Add(-time.Nanosecond)
	startJun := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	endJun := startJun.AddDate(0, 1, 0).Add(-time.Nanosecond)

	repo := &fakeJournalRepoPLMap{data: map[string][]repository.AccountBalance{
		startMay.Format("2006-01-02") + "_" + endMay.Format("2006-01-02"): {
			{AccountCode: "4.1", AccountName: "Penjualan", Balance: -200},
			{AccountCode: "5.1", AccountName: "HPP", Balance: 100},
		},
		startJun.Format("2006-01-02") + "_" + endJun.Format("2006-01-02"): {
			{AccountCode: "4.1", AccountName: "Penjualan", Balance: -100},
			{AccountCode: "5.1", AccountName: "HPP", Balance: 50},
		},
	}}

	svc := NewProfitLossReportService(repo)
	may, err := svc.GetProfitLoss(context.Background(), "Monthly", 5, 2025, "ShopX", false)
	if err != nil {
		t.Fatalf("error may: %v", err)
	}
	jun, err := svc.GetProfitLoss(context.Background(), "Monthly", 6, 2025, "ShopX", false)
	if err != nil {
		t.Fatalf("error jun: %v", err)
	}
	if may.TotalPendapatanUsaha == jun.TotalPendapatanUsaha {
		t.Errorf("expected different revenue, got both %v", may.TotalPendapatanUsaha)
	}
	if may.LabaRugiBersih.Amount == jun.LabaRugiBersih.Amount {
		t.Errorf("expected different profit, got %v", may.LabaRugiBersih.Amount)
	}
}

func TestProfitLossReportService_WithComparison(t *testing.T) {
	startMay := time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC)
	endMay := startMay.AddDate(0, 1, 0).Add(-time.Nanosecond)
	startApr := time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC)
	endApr := startApr.AddDate(0, 1, 0).Add(-time.Nanosecond)

	repo := &fakeJournalRepoPLMap{data: map[string][]repository.AccountBalance{
		startMay.Format("2006-01-02") + "_" + endMay.Format("2006-01-02"): {
			{AccountCode: "4.1", AccountName: "Penjualan", Balance: -200},
			{AccountCode: "5.1", AccountName: "HPP", Balance: 100},
		},
		startApr.Format("2006-01-02") + "_" + endApr.Format("2006-01-02"): {
			{AccountCode: "4.1", AccountName: "Penjualan", Balance: -100},
			{AccountCode: "5.1", AccountName: "HPP", Balance: 50},
		},
	}}

	svc := NewProfitLossReportService(repo)
	may, err := svc.GetProfitLoss(context.Background(), "Monthly", 5, 2025, "ShopX", true)
	if err != nil {
		t.Fatalf("error may: %v", err)
	}

	// Check current period data
	if may.TotalPendapatanUsaha != 200 {
		t.Errorf("expected current revenue 200, got %v", may.TotalPendapatanUsaha)
	}
	if may.PrevTotalPendapatanUsaha != 100 {
		t.Errorf("expected previous revenue 100, got %v", may.PrevTotalPendapatanUsaha)
	}

	// Check comparison data in rows
	if len(may.PendapatanUsaha) != 1 {
		t.Fatalf("expected 1 revenue row, got %d", len(may.PendapatanUsaha))
	}
	
	revRow := may.PendapatanUsaha[0]
	if revRow.Amount != 200 {
		t.Errorf("expected current amount 200, got %v", revRow.Amount)
	}
	if revRow.PreviousAmount != 100 {
		t.Errorf("expected previous amount 100, got %v", revRow.PreviousAmount)
	}
	if revRow.Change != 100 {
		t.Errorf("expected change 100, got %v", revRow.Change)
	}
	if revRow.ChangePercent != 100 {
		t.Errorf("expected change percent 100, got %v", revRow.ChangePercent)
	}
}
