package service

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/logutil"
	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
)

// ForecastDataPoint represents a single data point in time series
type ForecastDataPoint struct {
	Date   time.Time `json:"date"`
	Value  float64   `json:"value"`
	Source string    `json:"source"` // "historical" or "forecast"
}

// ForecastResult contains forecast data for a specific metric
type ForecastResult struct {
	Metric            string               `json:"metric"`
	HistoricalData    []ForecastDataPoint  `json:"historicalData"`
	ForecastData      []ForecastDataPoint  `json:"forecastData"`
	TotalForecast     float64              `json:"totalForecast"`
	TotalHistorical   float64              `json:"totalHistorical"`
	GrowthRate        float64              `json:"growthRate"`
	Confidence        float64              `json:"confidence"`
	Method            string               `json:"method"`
}

// ForecastRequest contains parameters for forecast generation
type ForecastRequest struct {
	Shop        string    `json:"shop"`
	Period      string    `json:"period"` // "monthly" or "yearly"
	StartDate   time.Time `json:"startDate"`
	EndDate     time.Time `json:"endDate"`
	ForecastTo  time.Time `json:"forecastTo"`
}

// ForecastResponse contains all forecast results
type ForecastResponse struct {
	Sales    ForecastResult `json:"sales"`
	Expenses ForecastResult `json:"expenses"`
	Profit   ForecastResult `json:"profit"`
	Period   string         `json:"period"`
	Generated time.Time      `json:"generated"`
}

// ForecastService provides forecasting capabilities using historical data
type ForecastService struct {
	dropshipRepo MetricServiceDropshipRepo
	shopeeRepo   MetricServiceShopeeRepo
	journalRepo  MetricServiceJournalRepo
}

// NewForecastService creates a new ForecastService
func NewForecastService(
	dr MetricServiceDropshipRepo,
	sr MetricServiceShopeeRepo,
	jr MetricServiceJournalRepo,
) *ForecastService {
	return &ForecastService{
		dropshipRepo: dr,
		shopeeRepo:   sr,
		journalRepo:  jr,
	}
}

// GenerateForecast creates forecasts for sales, expenses, and profit
func (fs *ForecastService) GenerateForecast(ctx context.Context, req ForecastRequest) (*ForecastResponse, error) {
	logutil.Printf("Generating forecast for shop=%s, period=%s, from=%s to=%s", 
		req.Shop, req.Period, req.StartDate.Format("2006-01-02"), req.ForecastTo.Format("2006-01-02"))

	// Generate sales forecast
	salesForecast, err := fs.forecastSales(ctx, req)
	if err != nil {
		logutil.Errorf("Failed to forecast sales: %v", err)
		return nil, fmt.Errorf("failed to forecast sales: %w", err)
	}

	// Generate expenses forecast
	expensesForecast, err := fs.forecastExpenses(ctx, req)
	if err != nil {
		logutil.Errorf("Failed to forecast expenses: %v", err)
		return nil, fmt.Errorf("failed to forecast expenses: %w", err)
	}

	// Generate profit forecast based on sales and expenses
	profitForecast := fs.calculateProfitForecast(salesForecast, expensesForecast)

	response := &ForecastResponse{
		Sales:     *salesForecast,
		Expenses:  *expensesForecast,
		Profit:    *profitForecast,
		Period:    req.Period,
		Generated: time.Now(),
	}

	logutil.Printf("Forecast generated: Sales=%.2f, Expenses=%.2f, Profit=%.2f", 
		salesForecast.TotalForecast, expensesForecast.TotalForecast, profitForecast.TotalForecast)

	return response, nil
}

// forecastSales generates sales forecast using historical dropship and Shopee data
func (fs *ForecastService) forecastSales(ctx context.Context, req ForecastRequest) (*ForecastResult, error) {
	// Get historical sales data
	historicalData, err := fs.getHistoricalSalesData(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get historical sales data: %w", err)
	}

	if len(historicalData) == 0 {
		return &ForecastResult{
			Metric:            "sales",
			HistoricalData:    []ForecastDataPoint{},
			ForecastData:      []ForecastDataPoint{},
			TotalForecast:     0,
			TotalHistorical:   0,
			GrowthRate:        0,
			Confidence:        0.5,
			Method:            "insufficient_data",
		}, nil
	}

	// Apply forecasting algorithm
	forecastData := fs.applyLinearTrendForecast(historicalData, req.ForecastTo)
	
	totalHistorical := fs.sumValues(historicalData)
	totalForecast := fs.sumValues(forecastData)
	growthRate := fs.calculateGrowthRate(historicalData)
	confidence := fs.calculateConfidence(historicalData)

	return &ForecastResult{
		Metric:            "sales",
		HistoricalData:    historicalData,
		ForecastData:      forecastData,
		TotalForecast:     totalForecast,
		TotalHistorical:   totalHistorical,
		GrowthRate:        growthRate,
		Confidence:        confidence,
		Method:            "linear_trend",
	}, nil
}

// forecastExpenses generates expenses forecast using historical journal data
func (fs *ForecastService) forecastExpenses(ctx context.Context, req ForecastRequest) (*ForecastResult, error) {
	// Get historical expenses data
	historicalData, err := fs.getHistoricalExpensesData(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get historical expenses data: %w", err)
	}

	if len(historicalData) == 0 {
		return &ForecastResult{
			Metric:            "expenses",
			HistoricalData:    []ForecastDataPoint{},
			ForecastData:      []ForecastDataPoint{},
			TotalForecast:     0,
			TotalHistorical:   0,
			GrowthRate:        0,
			Confidence:        0.5,
			Method:            "insufficient_data",
		}, nil
	}

	// Apply forecasting algorithm
	forecastData := fs.applyMovingAverageForecast(historicalData, req.ForecastTo)
	
	totalHistorical := fs.sumValues(historicalData)
	totalForecast := fs.sumValues(forecastData)
	growthRate := fs.calculateGrowthRate(historicalData)
	confidence := fs.calculateConfidence(historicalData)

	return &ForecastResult{
		Metric:            "expenses",
		HistoricalData:    historicalData,
		ForecastData:      forecastData,
		TotalForecast:     totalForecast,
		TotalHistorical:   totalHistorical,
		GrowthRate:        growthRate,
		Confidence:        confidence,
		Method:            "moving_average",
	}, nil
}

// calculateProfitForecast calculates profit forecast from sales and expenses
func (fs *ForecastService) calculateProfitForecast(sales, expenses *ForecastResult) *ForecastResult {
	// Create profit forecast data points
	var historicalData []ForecastDataPoint
	var forecastData []ForecastDataPoint

	// Calculate historical profit
	for i := 0; i < len(sales.HistoricalData) && i < len(expenses.HistoricalData); i++ {
		profit := sales.HistoricalData[i].Value - expenses.HistoricalData[i].Value
		historicalData = append(historicalData, ForecastDataPoint{
			Date:   sales.HistoricalData[i].Date,
			Value:  profit,
			Source: "historical",
		})
	}

	// Calculate forecast profit
	for i := 0; i < len(sales.ForecastData) && i < len(expenses.ForecastData); i++ {
		profit := sales.ForecastData[i].Value - expenses.ForecastData[i].Value
		forecastData = append(forecastData, ForecastDataPoint{
			Date:   sales.ForecastData[i].Date,
			Value:  profit,
			Source: "forecast",
		})
	}

	totalHistorical := fs.sumValues(historicalData)
	totalForecast := fs.sumValues(forecastData)
	growthRate := fs.calculateGrowthRate(historicalData)
	confidence := math.Min(sales.Confidence, expenses.Confidence)

	return &ForecastResult{
		Metric:            "profit",
		HistoricalData:    historicalData,
		ForecastData:      forecastData,
		TotalForecast:     totalForecast,
		TotalHistorical:   totalHistorical,
		GrowthRate:        growthRate,
		Confidence:        confidence,
		Method:            "calculated",
	}
}

// getHistoricalSalesData aggregates sales data from dropship and Shopee sources
func (fs *ForecastService) getHistoricalSalesData(ctx context.Context, req ForecastRequest) ([]ForecastDataPoint, error) {
	var dataPoints []ForecastDataPoint

	// Get data point by point based on period
	current := req.StartDate
	for current.Before(req.EndDate) {
		var periodEnd time.Time
		if req.Period == "monthly" {
			periodEnd = current.AddDate(0, 1, 0)
		} else {
			periodEnd = current.AddDate(1, 0, 0)
		}
		if periodEnd.After(req.EndDate) {
			periodEnd = req.EndDate
		}

		// Get dropship sales for this period
		dropshipSales, err := fs.getDropshipSalesForPeriod(ctx, req.Shop, current, periodEnd)
		if err != nil {
			return nil, fmt.Errorf("failed to get dropship sales: %w", err)
		}

		// Get Shopee sales for this period
		shopeeSales, err := fs.getShopeeSalesForPeriod(ctx, req.Shop, current, periodEnd)
		if err != nil {
			return nil, fmt.Errorf("failed to get Shopee sales: %w", err)
		}

		totalSales := dropshipSales + shopeeSales
		if totalSales > 0 {
			dataPoints = append(dataPoints, ForecastDataPoint{
				Date:   current,
				Value:  totalSales,
				Source: "historical",
			})
		}

		current = periodEnd
	}

	return dataPoints, nil
}

// getHistoricalExpensesData gets expenses data from journal entries
func (fs *ForecastService) getHistoricalExpensesData(ctx context.Context, req ForecastRequest) ([]ForecastDataPoint, error) {
	var dataPoints []ForecastDataPoint

	// Get expenses data point by point based on period
	current := req.StartDate
	for current.Before(req.EndDate) {
		var periodEnd time.Time
		if req.Period == "monthly" {
			periodEnd = current.AddDate(0, 1, 0)
		} else {
			periodEnd = current.AddDate(1, 0, 0)
		}
		if periodEnd.After(req.EndDate) {
			periodEnd = req.EndDate
		}

		// Get expenses for this period from journal entries
		expenses, err := fs.getExpensesForPeriod(ctx, req.Shop, current, periodEnd)
		if err != nil {
			return nil, fmt.Errorf("failed to get expenses: %w", err)
		}

		if expenses > 0 {
			dataPoints = append(dataPoints, ForecastDataPoint{
				Date:   current,
				Value:  expenses,
				Source: "historical",
			})
		}

		current = periodEnd
	}

	return dataPoints, nil
}

// getDropshipSalesForPeriod gets dropship sales total for a period
func (fs *ForecastService) getDropshipSalesForPeriod(ctx context.Context, shop string, start, end time.Time) (float64, error) {
	purchases, err := fs.dropshipRepo.ListDropshipPurchasesByShopAndDate(ctx, shop, 
		start.Format("2006-01-02"), end.Format("2006-01-02"))
	if err != nil {
		return 0, err
	}

	var total float64
	for _, purchase := range purchases {
		total += purchase.TotalTransaksi
	}

	return total, nil
}

// getShopeeSalesForPeriod gets Shopee sales total for a period
func (fs *ForecastService) getShopeeSalesForPeriod(ctx context.Context, shop string, start, end time.Time) (float64, error) {
	orders, err := fs.shopeeRepo.ListShopeeOrdersByShopAndDate(ctx, shop, 
		start.Format("2006-01-02"), end.Format("2006-01-02"))
	if err != nil {
		return 0, err
	}

	var total float64
	for _, order := range orders {
		total += order.NetIncome
	}

	return total, nil
}

// getExpensesForPeriod gets expenses total for a period from journal entries
func (fs *ForecastService) getExpensesForPeriod(ctx context.Context, shop string, start, end time.Time) (float64, error) {
	// Get account balances (expenses are typically accounts starting with 5)
	balances, err := fs.journalRepo.GetAccountBalancesAsOf(ctx, shop, end)
	if err != nil {
		return 0, err
	}

	var total float64
	for _, balance := range balances {
		// Sum expense accounts (codes starting with 5)
		if len(balance.AccountCode) > 0 && balance.AccountCode[0] == '5' {
			total += balance.Balance
		}
	}

	return total, nil
}

// applyLinearTrendForecast applies linear trend forecasting
func (fs *ForecastService) applyLinearTrendForecast(historical []ForecastDataPoint, forecastTo time.Time) []ForecastDataPoint {
	if len(historical) < 2 {
		return []ForecastDataPoint{}
	}

	// Calculate linear trend
	n := float64(len(historical))
	var sumX, sumY, sumXY, sumX2 float64

	for i, point := range historical {
		x := float64(i)
		y := point.Value
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	// Linear regression: y = a + bx
	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	intercept := (sumY - slope*sumX) / n

	var forecast []ForecastDataPoint
	lastDate := historical[len(historical)-1].Date
	
	// Generate forecast points
	i := len(historical)
	for current := lastDate; current.Before(forecastTo); {
		// Advance to next period
		if len(historical) > 1 {
			period := historical[1].Date.Sub(historical[0].Date)
			current = current.Add(period)
		} else {
			current = current.AddDate(0, 0, 1) // default to daily
		}
		
		if current.After(forecastTo) {
			break
		}

		// Calculate forecast value
		value := intercept + slope*float64(i)
		if value < 0 {
			value = 0 // Don't forecast negative values
		}

		forecast = append(forecast, ForecastDataPoint{
			Date:   current,
			Value:  value,
			Source: "forecast",
		})

		i++
	}

	return forecast
}

// applyMovingAverageForecast applies moving average forecasting
func (fs *ForecastService) applyMovingAverageForecast(historical []ForecastDataPoint, forecastTo time.Time) []ForecastDataPoint {
	if len(historical) == 0 {
		return []ForecastDataPoint{}
	}

	// Calculate moving average (use last 3 periods or all if less)
	window := 3
	if len(historical) < window {
		window = len(historical)
	}

	var sum float64
	for i := len(historical) - window; i < len(historical); i++ {
		sum += historical[i].Value
	}
	average := sum / float64(window)

	var forecast []ForecastDataPoint
	lastDate := historical[len(historical)-1].Date
	
	// Generate forecast points with moving average
	for current := lastDate; current.Before(forecastTo); {
		// Advance to next period
		if len(historical) > 1 {
			period := historical[1].Date.Sub(historical[0].Date)
			current = current.Add(period)
		} else {
			current = current.AddDate(0, 0, 1) // default to daily
		}
		
		if current.After(forecastTo) {
			break
		}

		forecast = append(forecast, ForecastDataPoint{
			Date:   current,
			Value:  average,
			Source: "forecast",
		})
	}

	return forecast
}

// sumValues calculates the sum of all values in data points
func (fs *ForecastService) sumValues(dataPoints []ForecastDataPoint) float64 {
	var sum float64
	for _, point := range dataPoints {
		sum += point.Value
	}
	return sum
}

// calculateGrowthRate calculates the growth rate from historical data
func (fs *ForecastService) calculateGrowthRate(historical []ForecastDataPoint) float64 {
	if len(historical) < 2 {
		return 0
	}

	// Calculate average growth rate
	var totalGrowth float64
	validPeriods := 0

	for i := 1; i < len(historical); i++ {
		if historical[i-1].Value > 0 {
			growth := (historical[i].Value - historical[i-1].Value) / historical[i-1].Value
			totalGrowth += growth
			validPeriods++
		}
	}

	if validPeriods == 0 {
		return 0
	}

	return totalGrowth / float64(validPeriods)
}

// calculateConfidence calculates confidence level based on data consistency
func (fs *ForecastService) calculateConfidence(historical []ForecastDataPoint) float64 {
	if len(historical) < 2 {
		return 0.5
	}

	// Calculate coefficient of variation
	mean := fs.sumValues(historical) / float64(len(historical))
	if mean == 0 {
		return 0.5
	}

	var variance float64
	for _, point := range historical {
		variance += math.Pow(point.Value-mean, 2)
	}
	variance /= float64(len(historical))
	stdDev := math.Sqrt(variance)
	
	cv := stdDev / mean
	
	// Convert coefficient of variation to confidence (inverse relationship)
	confidence := 1.0 / (1.0 + cv)
	
	// Cap confidence between 0.1 and 0.9
	if confidence < 0.1 {
		confidence = 0.1
	}
	if confidence > 0.9 {
		confidence = 0.9
	}

	return confidence
}