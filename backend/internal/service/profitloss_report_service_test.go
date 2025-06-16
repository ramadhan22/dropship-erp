package service

import (
	"context"
	"testing"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
)

type fakeJournalRepoPL struct{ balances []repository.AccountBalance }

func (f *fakeJournalRepoPL) GetAccountBalancesBetween(ctx context.Context, shop string, from, to time.Time) ([]repository.AccountBalance, error) {
	return f.balances, nil
}

func TestProfitLossReportService_GetProfitLoss(t *testing.T) {
	repo := &fakeJournalRepoPL{balances: []repository.AccountBalance{
		{AccountCode: "4.1", AccountName: "Penjualan", Balance: -200},
		{AccountCode: "5.1", AccountName: "HPP", Balance: 100},
		{AccountCode: "5.2.3", AccountName: "Voucher", Balance: 5},
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
	if len(pl.BebanOperasional) != 1 {
		t.Errorf("expected 1 operasional row, got %d", len(pl.BebanOperasional))
	}
}
