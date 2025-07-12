package service

import (
	"context"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
)

type DashboardRepo interface {
	CountOrders(ctx context.Context, channel, store, from, to string) (int, error)
	AvgOrderValue(ctx context.Context, channel, store, from, to string) (float64, error)
	CancelledSummary(ctx context.Context, channel, store, from, to string) (repository.CancelledSummary, error)
	DistinctCustomers(ctx context.Context, channel, store, from, to string) (int, error)
	DailyTotals(ctx context.Context, channel, store, from, to string) ([]repository.DailyPurchaseTotal, error)
	MonthlyTotals(ctx context.Context, channel, store, from, to string) ([]repository.MonthlyPurchaseTotal, error)
	SumDropshipPurchases(ctx context.Context, channel, store, from, to string) (float64, error)
}

type DashboardJournalRepo interface {
	GetAccountBalancesAsOf(ctx context.Context, shop string, asOfDate time.Time) ([]repository.AccountBalance, error)
}

type DashboardService struct {
	dropRepo    DashboardRepo
	journalRepo DashboardJournalRepo
	plSvc       *ProfitLossReportService
}

func NewDashboardService(dr DashboardRepo, jr DashboardJournalRepo, pl *ProfitLossReportService) *DashboardService {
	return &DashboardService{dropRepo: dr, journalRepo: jr, plSvc: pl}
}

type DashboardFilters struct {
	Channel string
	Store   string
	Period  string
	Month   int
	Year    int
}

type SummaryItem struct {
	Value  float64 `json:"value"`
	Change float64 `json:"change"`
}

type Point struct {
	Date  string  `json:"date"`
	Value float64 `json:"value"`
}

type DashboardData struct {
	Summary map[string]SummaryItem `json:"summary"`
	Charts  map[string][]Point     `json:"charts"`
}

func (s *DashboardService) GetDashboardData(ctx context.Context, f DashboardFilters) (*DashboardData, error) {
	var start, end time.Time
	if f.Period == "Yearly" {
		start = time.Date(f.Year, 1, 1, 0, 0, 0, 0, time.UTC)
		end = start.AddDate(1, 0, 0).Add(-time.Nanosecond)
	} else {
		if f.Month == 0 {
			f.Month = int(time.Now().Month())
		}
		start = time.Date(f.Year, time.Month(f.Month), 1, 0, 0, 0, 0, time.UTC)
		end = start.AddDate(0, 1, 0).Add(-time.Nanosecond)
	}
	from := start.Format("2006-01-02")
	to := end.Format("2006-01-02")

	totalOrders, err := s.dropRepo.CountOrders(ctx, f.Channel, f.Store, from, to)
	if err != nil {
		return nil, err
	}
	avgOrder, err := s.dropRepo.AvgOrderValue(ctx, f.Channel, f.Store, from, to)
	if err != nil {
		return nil, err
	}
	cancelled, err := s.dropRepo.CancelledSummary(ctx, f.Channel, f.Store, from, to)
	if err != nil {
		return nil, err
	}
	customers, err := s.dropRepo.DistinctCustomers(ctx, f.Channel, f.Store, from, to)
	if err != nil {
		return nil, err
	}
	totalPrice, err := s.dropRepo.SumDropshipPurchases(ctx, f.Channel, f.Store, from, to)
	if err != nil {
		return nil, err
	}

	// simple outstanding from account 11010
	balances, err := s.journalRepo.GetAccountBalancesAsOf(ctx, f.Store, end)
	var outstanding float64
	if err == nil {
		for _, ab := range balances {
			if ab.AccountID == 11010 {
				outstanding = ab.Balance
				break
			}
		}
	}

	pl, _ := s.plSvc.GetProfitLoss(ctx, f.Period, f.Month, f.Year, f.Store)
	netProfit := 0.0
	if pl != nil {
		netProfit = pl.LabaRugiBersih.Amount
	}

	charts := make(map[string][]Point)
	if f.Period == "Yearly" {
		m, err := s.dropRepo.MonthlyTotals(ctx, f.Channel, f.Store, from, to)
		if err == nil {
			for _, v := range m {
				charts["total_sales"] = append(charts["total_sales"], Point{Date: v.Month, Value: v.Total})
				avg := 0.0
				if v.Count > 0 {
					avg = v.Total / float64(v.Count)
				}
				charts["avg_order_value"] = append(charts["avg_order_value"], Point{Date: v.Month, Value: avg})
				charts["number_of_orders"] = append(charts["number_of_orders"], Point{Date: v.Month, Value: float64(v.Count)})
			}
		}
	} else {
		d, err := s.dropRepo.DailyTotals(ctx, f.Channel, f.Store, from, to)
		if err == nil {
			for _, v := range d {
				charts["total_sales"] = append(charts["total_sales"], Point{Date: v.Date, Value: v.Total})
				avg := 0.0
				if v.Count > 0 {
					avg = v.Total / float64(v.Count)
				}
				charts["avg_order_value"] = append(charts["avg_order_value"], Point{Date: v.Date, Value: avg})
				charts["number_of_orders"] = append(charts["number_of_orders"], Point{Date: v.Date, Value: float64(v.Count)})
			}
		}
	}

	data := &DashboardData{
		Summary: map[string]SummaryItem{
			"total_orders":       {Value: float64(totalOrders)},
			"avg_order_value":    {Value: avgOrder},
			"total_cancelled":    {Value: float64(cancelled.Count)},
			"total_customers":    {Value: float64(customers)},
			"total_price":        {Value: totalPrice},
			"total_discounts":    {Value: 0},
			"total_net_profit":   {Value: netProfit},
			"outstanding_amount": {Value: outstanding},
		},
		Charts: charts,
	}
	return data, nil
}
