// File: backend/internal/service/metric_service_test.go

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
)

// fakes for each repo interface

type fakeDropRepoM struct {
	data map[string][]models.DropshipPurchase
}

func (f *fakeDropRepoM) ListDropshipPurchasesByShopAndDate(
	ctx context.Context, shop, from, to string,
) ([]models.DropshipPurchase, error) {
	if ds, ok := f.data[shop]; ok {
		return ds, nil
	}
	return nil, nil
}

type fakeShopeeRepoM struct {
	data map[string][]models.ShopeeSettledOrder
}

func (f *fakeShopeeRepoM) ListShopeeOrdersByShopAndDate(
	ctx context.Context, shop, from, to string,
) ([]models.ShopeeSettledOrder, error) {
	if so, ok := f.data[shop]; ok {
		return so, nil
	}
	return nil, nil
}

type fakeJournalRepoM struct {
	data map[string][]repository.AccountBalance
}

func (f *fakeJournalRepoM) GetAccountBalancesAsOf(
	ctx context.Context, shop string, asOfDate time.Time,
) ([]repository.AccountBalance, error) {
	if bs, ok := f.data[shop]; ok {
		return bs, nil
	}
	return nil, nil
}

type fakeMetricRepoM struct {
	saved []*models.CachedMetric
}

func (f *fakeMetricRepoM) UpsertCachedMetric(
	ctx context.Context, m *models.CachedMetric,
) error {
	f.saved = append(f.saved, m)
	return nil
}
func (f *fakeMetricRepoM) GetCachedMetric(
	ctx context.Context, shop, period string,
) (*models.CachedMetric, error) {
	return nil, errors.New("not needed")
}

func TestCalculateAndCacheMetrics(t *testing.T) {
	ctx := context.Background()

	// Prepare fake data: one Shopee order and one Dropship purchase
	shop := "ShopA"
	period := "2025-05"
	start := time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC)

	fDrop := &fakeDropRepoM{
		data: map[string][]models.DropshipPurchase{
			shop: {
				{
					KodePesanan:         "DP-1",
					NamaToko:            shop,
					TotalTransaksi:      52.00,
					WaktuPesananTerbuat: start.AddDate(0, 0, 5),
				},
			},
		},
	}
	fShopee := &fakeShopeeRepoM{
		data: map[string][]models.ShopeeSettledOrder{
			shop: {
				{
					OrderID:         "SO-1",
					NetIncome:       100.00,
					ServiceFee:      3.00,
					CampaignFee:     0.00,
					CreditCardFee:   1.50,
					ShippingSubsidy: 0.00,
					TaxImportFee:    0.00,
					SettledDate:     start.AddDate(0, 0, 10),
				},
			},
		},
	}
	fJournal := &fakeJournalRepoM{
		data: map[string][]repository.AccountBalance{
			shop: {
				{
					AccountID:   1001,
					AccountCode: "1001",
					AccountName: "Cash",
					AccountType: "Asset",
					Balance:     200.00,
				},
			},
		},
	}
	fMetric := &fakeMetricRepoM{}

	svc := NewMetricService(fDrop, fShopee, fJournal, fMetric)
	if err := svc.CalculateAndCacheMetrics(ctx, shop, period); err != nil {
		t.Fatalf("CalculateAndCacheMetrics failed: %v", err)
	}

	// Verify one CachedMetric was saved
	if len(fMetric.saved) != 1 {
		t.Fatalf("expected 1 CachedMetric, got %d", len(fMetric.saved))
	}
	cm := fMetric.saved[0]
	// sumRevenue=100.00, sumFees=3+1.5=4.5, sumCOGS=52, netProfit=100-4.5-52=43.5
	if cm.NetProfit != 43.5 {
		t.Errorf("unexpected NetProfit: %f", cm.NetProfit)
	}
	if cm.EndingCashBalance != 200.00 {
		t.Errorf("unexpected EndingCashBalance: %f", cm.EndingCashBalance)
	}
}
