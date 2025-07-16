package repository

import (
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// AdsPerformanceRepository handles database operations for ads performance data.
type AdsPerformanceRepository struct {
	db *sqlx.DB
}

// NewAdsPerformanceRepository creates a new repository instance.
func NewAdsPerformanceRepository(db *sqlx.DB) *AdsPerformanceRepository {
	return &AdsPerformanceRepository{db: db}
}

// Create inserts a new ads performance record.
func (r *AdsPerformanceRepository) Create(ap *models.AdsPerformance) error {
	ap.CalculateMetrics()
	
	query := `
		INSERT INTO ads_performance (
			store_id, campaign_id, campaign_name, campaign_type, campaign_status,
			date_from, date_to, ads_viewed, total_clicks, orders_count,
			products_sold, sales_from_ads, ad_costs, click_rate, roas,
			daily_budget, target_roas, performance_change_percentage
		) VALUES (
			:store_id, :campaign_id, :campaign_name, :campaign_type, :campaign_status,
			:date_from, :date_to, :ads_viewed, :total_clicks, :orders_count,
			:products_sold, :sales_from_ads, :ad_costs, :click_rate, :roas,
			:daily_budget, :target_roas, :performance_change_percentage
		)
		ON CONFLICT (store_id, campaign_id, date_from, date_to)
		DO UPDATE SET
			campaign_name = EXCLUDED.campaign_name,
			campaign_type = EXCLUDED.campaign_type,
			campaign_status = EXCLUDED.campaign_status,
			ads_viewed = EXCLUDED.ads_viewed,
			total_clicks = EXCLUDED.total_clicks,
			orders_count = EXCLUDED.orders_count,
			products_sold = EXCLUDED.products_sold,
			sales_from_ads = EXCLUDED.sales_from_ads,
			ad_costs = EXCLUDED.ad_costs,
			click_rate = EXCLUDED.click_rate,
			roas = EXCLUDED.roas,
			daily_budget = EXCLUDED.daily_budget,
			target_roas = EXCLUDED.target_roas,
			performance_change_percentage = EXCLUDED.performance_change_percentage,
			updated_at = NOW()
		RETURNING id`
	
	return r.db.Get(&ap.ID, query, ap)
}

// List retrieves ads performance records with optional filtering.
func (r *AdsPerformanceRepository) List(filter *models.AdsPerformanceFilter, limit, offset int) ([]models.AdsPerformance, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	baseQuery := `
		SELECT ap.*, s.name as store_name
		FROM ads_performance ap
		LEFT JOIN stores s ON ap.store_id = s.id
		WHERE 1=1`

	if filter.StoreID != nil {
		conditions = append(conditions, fmt.Sprintf("ap.store_id = $%d", argIndex))
		args = append(args, *filter.StoreID)
		argIndex++
	}

	if filter.CampaignStatus != nil {
		conditions = append(conditions, fmt.Sprintf("ap.campaign_status = $%d", argIndex))
		args = append(args, *filter.CampaignStatus)
		argIndex++
	}

	if filter.CampaignType != nil {
		conditions = append(conditions, fmt.Sprintf("ap.campaign_type = $%d", argIndex))
		args = append(args, *filter.CampaignType)
		argIndex++
	}

	if filter.DateFrom != nil {
		conditions = append(conditions, fmt.Sprintf("ap.date_from >= $%d", argIndex))
		args = append(args, *filter.DateFrom)
		argIndex++
	}

	if filter.DateTo != nil {
		conditions = append(conditions, fmt.Sprintf("ap.date_to <= $%d", argIndex))
		args = append(args, *filter.DateTo)
		argIndex++
	}

	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	baseQuery += fmt.Sprintf(" ORDER BY ap.date_from DESC, ap.campaign_name LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	var results []models.AdsPerformance
	err := r.db.Select(&results, baseQuery, args...)
	return results, err
}

// GetSummary calculates aggregated metrics for the given filter criteria.
func (r *AdsPerformanceRepository) GetSummary(filter *models.AdsPerformanceFilter) (*models.AdsPerformanceSummary, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	baseQuery := `
		SELECT 
			COALESCE(SUM(ads_viewed), 0) as total_ads_viewed,
			COALESCE(SUM(total_clicks), 0) as total_clicks,
			COALESCE(SUM(orders_count), 0) as total_orders,
			COALESCE(SUM(products_sold), 0) as total_products_sold,
			COALESCE(SUM(sales_from_ads), 0) as total_sales_from_ads,
			COALESCE(SUM(ad_costs), 0) as total_ad_costs,
			COALESCE(AVG(click_rate), 0) as average_click_rate,
			COALESCE(AVG(roas), 0) as average_roas,
			MIN(date_from) as date_from,
			MAX(date_to) as date_to
		FROM ads_performance
		WHERE 1=1`

	if filter.StoreID != nil {
		conditions = append(conditions, fmt.Sprintf("store_id = $%d", argIndex))
		args = append(args, *filter.StoreID)
		argIndex++
	}

	if filter.CampaignStatus != nil {
		conditions = append(conditions, fmt.Sprintf("campaign_status = $%d", argIndex))
		args = append(args, *filter.CampaignStatus)
		argIndex++
	}

	if filter.DateFrom != nil {
		conditions = append(conditions, fmt.Sprintf("date_from >= $%d", argIndex))
		args = append(args, *filter.DateFrom)
		argIndex++
	}

	if filter.DateTo != nil {
		conditions = append(conditions, fmt.Sprintf("date_to <= $%d", argIndex))
		args = append(args, *filter.DateTo)
		argIndex++
	}

	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	var summary models.AdsPerformanceSummary
	err := r.db.Get(&summary, baseQuery, args...)
	return &summary, err
}

// GetByID retrieves a specific ads performance record.
func (r *AdsPerformanceRepository) GetByID(id int) (*models.AdsPerformance, error) {
	var ap models.AdsPerformance
	query := "SELECT * FROM ads_performance WHERE id = $1"
	err := r.db.Get(&ap, query, id)
	return &ap, err
}

// DeleteOldRecords removes records older than the specified date.
func (r *AdsPerformanceRepository) DeleteOldRecords(olderThan time.Time) (int64, error) {
	query := "DELETE FROM ads_performance WHERE created_at < $1"
	result, err := r.db.Exec(query, olderThan)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}