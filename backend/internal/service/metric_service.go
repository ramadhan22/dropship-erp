// File: backend/internal/service/metric_service.go

package service

import (
	"context"
	"fmt"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
)

// MetricRepoInterface defines methods needed by MetricService.
type MetricRepoInterface interface {
	UpsertCachedMetric(ctx context.Context, m *models.CachedMetric) error
	GetCachedMetric(ctx context.Context, shop, period string) (*models.CachedMetric, error)
}

// MetricServiceDropshipRepo defines the subset of DropshipRepo methods required.
type MetricServiceDropshipRepo interface {
	ListDropshipPurchasesByShopAndDate(ctx context.Context, shop, from, to string) ([]models.DropshipPurchase, error)
}

// MetricServiceShopeeRepo defines the subset of ShopeeRepo methods required.
type MetricServiceShopeeRepo interface {
	ListShopeeOrdersByShopAndDate(ctx context.Context, shop, from, to string) ([]models.ShopeeSettledOrder, error)
}

// MetricServiceJournalRepo defines the subset of JournalRepo methods required.
type MetricServiceJournalRepo interface {
	GetAccountBalancesAsOf(ctx context.Context, shop string, asOfDate time.Time) ([]repository.AccountBalance, error)
}

// MetricService calculates financial metrics and caches them.
type MetricService struct {
	dropRepo    MetricServiceDropshipRepo
	shopeeRepo  MetricServiceShopeeRepo
	journalRepo MetricServiceJournalRepo
	metricRepo  MetricRepoInterface
}

// NewMetricService constructs a MetricService with the provided repositories.
func NewMetricService(
	dr MetricServiceDropshipRepo,
	sr MetricServiceShopeeRepo,
	jr MetricServiceJournalRepo,
	mr MetricRepoInterface,
) *MetricService {
	return &MetricService{
		dropRepo:    dr,
		shopeeRepo:  sr,
		journalRepo: jr,
		metricRepo:  mr,
	}
}

// CalculateAndCacheMetrics computes revenue, COGS, fees, net profit, and ending cash, then caches them.
func (s *MetricService) CalculateAndCacheMetrics(
	ctx context.Context,
	shop, period string,
) error {
	// Parse period into start and end dates
	start, err := time.Parse("2006-01-02", period+"-01")
	if err != nil {
		return fmt.Errorf("invalid period format: %w", err)
	}
	end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)

	// Format for repo queries
	fromDate := start.Format("2006-01-02")
	toDate := end.Format("2006-01-02")

	// Aggregate Shopee orders
	orders, err := s.shopeeRepo.ListShopeeOrdersByShopAndDate(ctx, shop, fromDate, toDate)
	if err != nil {
		return fmt.Errorf("list shopee orders: %w", err)
	}
	var sumRevenue, sumFees float64
	for _, o := range orders {
		sumRevenue += o.NetIncome
		sumFees += o.ServiceFee + o.CampaignFee + o.CreditCardFee + o.ShippingSubsidy + o.TaxImportFee
	}

	// Aggregate Dropship costs
	purchases, err := s.dropRepo.ListDropshipPurchasesByShopAndDate(ctx, shop, fromDate, toDate)
	if err != nil {
		return fmt.Errorf("list dropship purchases: %w", err)
	}
	var sumCOGS float64
	for _, dp := range purchases {
		sumCOGS += dp.TotalTransaksi
	}

	// Calculate net profit
	netProfit := sumRevenue - sumCOGS - sumFees

	// Get ending cash balance
	balances, err := s.journalRepo.GetAccountBalancesAsOf(ctx, shop, end)
	if err != nil {
		return fmt.Errorf("get account balances: %w", err)
	}
	var endingCash float64
	for _, ab := range balances {
		if ab.AccountCode == "1001" {
			endingCash = ab.Balance
			break
		}
	}

	// Upsert to cache
	cm := &models.CachedMetric{
		ShopUsername:      shop,
		Period:            period,
		SumRevenue:        sumRevenue,
		SumCOGS:           sumCOGS,
		SumFees:           sumFees,
		NetProfit:         netProfit,
		EndingCashBalance: endingCash,
		UpdatedAt:         time.Now(),
	}
	if err := s.metricRepo.UpsertCachedMetric(ctx, cm); err != nil {
		return fmt.Errorf("upsert cached metric: %w", err)
	}
	return nil
}

// MetricRepo exposes the underlying MetricRepoInterface for handler GET.
func (s *MetricService) MetricRepo() MetricRepoInterface {
	return s.metricRepo
}
