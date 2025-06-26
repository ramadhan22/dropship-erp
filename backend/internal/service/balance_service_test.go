// File: backend/internal/service/balance_service_test.go

package service

import (
	"context"
	"testing"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
)

// fakeJournalRepoB simulates GetAccountBalancesAsOf
type fakeJournalRepoB struct {
	data map[string][]repository.AccountBalance
}

func (f *fakeJournalRepoB) GetAccountBalancesAsOf(ctx context.Context, shop string, asOfDate time.Time) ([]repository.AccountBalance, error) {
	return f.data[shop], nil
}

func TestGetBalanceSheet(t *testing.T) {
	ctx := context.Background()
	shop := "ShopX"
	// Prepare fake balances: one asset and one liability
	fFake := &fakeJournalRepoB{
		data: map[string][]repository.AccountBalance{
			shop: {
				{AccountID: 1001, AccountCode: "1001", AccountName: "Cash", AccountType: "Asset", Balance: 500.0},
				{AccountID: 2001, AccountCode: "2001", AccountName: "Payable", AccountType: "Liability", Balance: -200.0},
				{AccountID: 3001, AccountCode: "3001", AccountName: "Equity", AccountType: "Equity", Balance: 300.0},
			},
		},
	}
	svc := NewBalanceService(fFake)
	asOf := time.Now()
	cats, err := svc.GetBalanceSheet(ctx, shop, asOf)
	if err != nil {
		t.Fatalf("GetBalanceSheet failed: %v", err)
	}

	if len(cats) != 3 {
		t.Fatalf("expected 3 categories, got %d", len(cats))
	}
	// Check Asset category
	if cats[0].Category != "Assets" || cats[0].Total != 500.0 {
		t.Errorf("unexpected Assets group: %+v", cats[0])
	}
	// Check Liability category
	if cats[1].Category != "Liabilities" || cats[1].Total != -200.0 {
		t.Errorf("unexpected Liabilities group: %+v", cats[1])
	}
	// Check Equity category
	if cats[2].Category != "Equity" || cats[2].Total != 300.0 {
		t.Errorf("unexpected Equity group: %+v", cats[2])
	}
}

// fakeJournalRepoBMap returns balances based on asOfDate.
type fakeJournalRepoBMap struct {
	data map[string]map[string][]repository.AccountBalance
}

func (f *fakeJournalRepoBMap) GetAccountBalancesAsOf(ctx context.Context, shop string, asOfDate time.Time) ([]repository.AccountBalance, error) {
	return f.data[shop][asOfDate.Format("2006-01-02")], nil
}

func TestGetBalanceSheet_DifferentDates(t *testing.T) {
	shop := "ShopX"
	d1 := time.Date(2025, 5, 31, 0, 0, 0, 0, time.UTC)
	d2 := time.Date(2025, 6, 30, 0, 0, 0, 0, time.UTC)
	repo := &fakeJournalRepoBMap{data: map[string]map[string][]repository.AccountBalance{
		shop: {
			d1.Format("2006-01-02"): {
				{AccountID: 1001, AccountCode: "1001", AccountName: "Cash", AccountType: "Asset", Balance: 100},
			},
			d2.Format("2006-01-02"): {
				{AccountID: 1001, AccountCode: "1001", AccountName: "Cash", AccountType: "Asset", Balance: 200},
			},
		},
	}}
	svc := NewBalanceService(repo)
	c1, err := svc.GetBalanceSheet(context.Background(), shop, d1)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	c2, err := svc.GetBalanceSheet(context.Background(), shop, d2)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if c1[0].Total == c2[0].Total {
		t.Errorf("expected different asset totals")
	}
}
