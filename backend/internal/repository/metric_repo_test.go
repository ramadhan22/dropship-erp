// File: backend/internal/repository/metric_repo_test.go

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// cleanupMetric deletes a cached_metrics row by shop and period.
func cleanupMetric(t *testing.T, shop, period string) {
	_, err := testDB.ExecContext(context.Background(),
		"DELETE FROM cached_metrics WHERE shop_username = $1 AND period = $2", shop, period)
	if err != nil {
		t.Fatalf("cleanupMetric failed: %v", err)
	}
}

func TestUpsertAndGetCachedMetric(t *testing.T) {
	ctx := context.Background()
	repo := NewMetricRepo(testDB)

	shop := "TestShop"
	period := time.Now().Format("2006-01")
	cm := &models.CachedMetric{
		ShopUsername:      shop,
		Period:            period,
		SumRevenue:        100.00,
		SumCOGS:           50.00,
		SumFees:           5.00,
		NetProfit:         45.00,
		EndingCashBalance: 1000.00,
		UpdatedAt:         time.Now(),
	}
	if err := repo.UpsertCachedMetric(ctx, cm); err != nil {
		t.Fatalf("UpsertCachedMetric failed: %v", err)
	}
	t.Log("UpsertCachedMetric succeeded")

	fetched, err := repo.GetCachedMetric(ctx, shop, period)
	if err != nil {
		t.Fatalf("GetCachedMetric failed: %v", err)
	}
	if fetched.NetProfit != 45.00 {
		t.Errorf("Expected NetProfit 45.00, got %f", fetched.NetProfit)
	}
	t.Log("GetCachedMetric succeeded")

	// Cleanup
	cleanupMetric(t, shop, period)
}
