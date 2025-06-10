package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// MetricRepo handles CRUD for the cached_metrics table.
type MetricRepo struct {
	db *sqlx.DB
}

// NewMetricRepo constructs a MetricRepo.
func NewMetricRepo(db *sqlx.DB) *MetricRepo {
	return &MetricRepo{db: db}
}

// UpsertCachedMetric inserts or updates a CachedMetric record.
// Uses ON CONFLICT on (shop_username, period) to decide insert vs. update.
func (r *MetricRepo) UpsertCachedMetric(ctx context.Context, m *models.CachedMetric) error {
	query := `
        INSERT INTO cached_metrics
          (shop_username, period, sum_revenue, sum_cogs, sum_fees, net_profit, ending_cash_balance, updated_at)
        VALUES
          (:shop_username, :period, :sum_revenue, :sum_cogs, :sum_fees, :net_profit, :ending_cash_balance, :updated_at)
        ON CONFLICT (shop_username, period)
        DO UPDATE SET
          sum_revenue = EXCLUDED.sum_revenue,
          sum_cogs = EXCLUDED.sum_cogs,
          sum_fees = EXCLUDED.sum_fees,
          net_profit = EXCLUDED.net_profit,
          ending_cash_balance = EXCLUDED.ending_cash_balance,
          updated_at = EXCLUDED.updated_at;
    `
	_, err := r.db.NamedExecContext(ctx, query, m)
	return err
}

// GetCachedMetric retrieves a single CachedMetric by shop and period (YYYY-MM).
// This lets your service quickly return pre‐computed metrics instead of recalculating on the fly.
func (r *MetricRepo) GetCachedMetric(
	ctx context.Context,
	shop, period string,
) (*models.CachedMetric, error) {
	var cm models.CachedMetric
	err := r.db.GetContext(ctx, &cm,
		`SELECT * FROM cached_metrics
         WHERE shop_username = $1
           AND period = $2`,
		shop, period)
	if err != nil {
		return nil, err
	}
	return &cm, nil
}
