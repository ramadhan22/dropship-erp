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
		{AccountCode: "5.2.3.1", AccountName: "Voucher", Balance: 5},
	}}
	svc := NewProfitLossReportService(repo)

	pl, err := svc.GetProfitLoss(context.Background(), "Monthly", 5, 2025, "ShopX")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pl.TotalPendapatanUsaha != 200 {
		t.Errorf("got %f want 200", pl.TotalPendapatanUsaha)
	}
	if pl.LabaRugiBersih.Amount != 95 {
		t.Errorf("got %f want 95", pl.LabaRugiBersih.Amount)
	}
	if len(pl.BebanOperasional) != 2 {
		t.Errorf("expected 2 operasional rows, got %d", len(pl.BebanOperasional))
	}
	if pl.BebanOperasional[0].Label != "Beban Pemasaran" {
		t.Errorf("got label %s want Beban Pemasaran", pl.BebanOperasional[0].Label)
	}
	if !pl.BebanOperasional[0].Group {
		t.Errorf("expected marketing header to be group")
	}
	if pl.BebanOperasional[1].Label != "Voucher" {
		t.Errorf("got label %s want Voucher", pl.BebanOperasional[1].Label)
	}
	if pl.BebanOperasional[1].Indent != 1 {
		t.Errorf("expected voucher indent 1, got %d", pl.BebanOperasional[1].Indent)
	}
}

func TestProfitLossReportService_SkipMarketingParentAccount(t *testing.T) {
	repo := &fakeJournalRepoPL{balances: []repository.AccountBalance{
		{AccountCode: "4.1", AccountName: "Penjualan", Balance: -100},
		{AccountCode: "5.1", AccountName: "HPP", Balance: 50},
		{AccountCode: "5.2.3", AccountName: "Beban Pemasaran", Balance: 10},
		{AccountCode: "5.2.3.1", AccountName: "Voucher", Balance: 5},
	}}
	svc := NewProfitLossReportService(repo)

	pl, err := svc.GetProfitLoss(context.Background(), "Monthly", 5, 2025, "ShopX")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(pl.BebanOperasional) != 2 {
		t.Fatalf("expected 2 operasional rows, got %d", len(pl.BebanOperasional))
	}
	if pl.BebanOperasional[0].Label != "Beban Pemasaran" {
		t.Errorf("got label %s want Beban Pemasaran", pl.BebanOperasional[0].Label)
	}
	if !pl.BebanOperasional[0].Group {
		t.Errorf("expected marketing header to be group")
	}
	if pl.BebanOperasional[1].Label != "Voucher" {
		t.Errorf("got label %s want Voucher", pl.BebanOperasional[1].Label)
	}
	if pl.BebanOperasional[1].Indent != 1 {
		t.Errorf("expected voucher indent 1, got %d", pl.BebanOperasional[1].Indent)
	}
}

func TestProfitLossReportService_GetProfitLoss_MonthlyPeriod(t *testing.T) {
	repo := &fakeJournalRepoPL{balances: []repository.AccountBalance{}}
	svc := NewProfitLossReportService(repo)

	_, err := svc.GetProfitLoss(context.Background(), "Monthly", 3, 2025, "ShopX")
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

	_, err := svc.GetProfitLoss(context.Background(), "Yearly", 5, 2025, "ShopX")
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
