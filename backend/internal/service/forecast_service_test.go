package service

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
)

// Mock repositories for testing
type mockForecastDropshipRepo struct{}

func (m *mockForecastDropshipRepo) ListDropshipPurchasesByShopAndDate(ctx context.Context, shop, from, to string) ([]models.DropshipPurchase, error) {
	// Return some sample data for testing
	return []models.DropshipPurchase{
		{
			KodePesanan:           "ORDER1",
			TotalTransaksi:        1000000, // 1M IDR
			WaktuPesananTerbuat:   time.Now().AddDate(0, -1, 0), // 1 month ago
		},
		{
			KodePesanan:           "ORDER2", 
			TotalTransaksi:        1200000, // 1.2M IDR
			WaktuPesananTerbuat:   time.Now().AddDate(0, 0, -15), // 15 days ago
		},
	}, nil
}

type mockForecastJournalRepo struct{}

func (m *mockForecastJournalRepo) GetAccountBalancesAsOf(ctx context.Context, shop string, asOfDate time.Time) ([]repository.AccountBalance, error) {
	// Return some sample expense account balances
	return []repository.AccountBalance{
		{
			AccountCode: "5000", // Expense account
			Balance:     500000,  // 500K IDR
		},
		{
			AccountCode: "5100", // Another expense account
			Balance:     300000,  // 300K IDR
		},
	}, nil
}

func TestForecastService_GenerateForecast(t *testing.T) {
	// Initialize the forecast service with mock repositories (without obsolete shopee repo)
	dropshipRepo := &mockForecastDropshipRepo{}
	journalRepo := &mockForecastJournalRepo{}
	
	forecastSvc := NewForecastService(dropshipRepo, journalRepo)
	
	// Create a forecast request
	now := time.Now()
	req := ForecastRequest{
		Shop:       "testshop",
		Period:     "monthly",
		StartDate:  now.AddDate(0, -3, 0), // 3 months ago
		EndDate:    now,
		ForecastTo: now.AddDate(0, 1, 0), // 1 month into future
	}
	
	// Generate forecast
	ctx := context.Background()
	forecast, err := forecastSvc.GenerateForecast(ctx, req)
	
	// Validate results
	if err != nil {
		t.Fatalf("GenerateForecast failed: %v", err)
	}
	
	if forecast == nil {
		t.Fatal("Forecast result is nil")
	}
	
	// Check that all forecast components are present
	if forecast.Sales.Metric != "sales" {
		t.Errorf("Expected sales metric, got %s", forecast.Sales.Metric)
	}
	
	if forecast.Expenses.Metric != "expenses" {
		t.Errorf("Expected expenses metric, got %s", forecast.Expenses.Metric)
	}
	
	if forecast.Profit.Metric != "profit" {
		t.Errorf("Expected profit metric, got %s", forecast.Profit.Metric)
	}
	
	// Check that forecast has reasonable data
	if forecast.Sales.TotalHistorical <= 0 {
		t.Errorf("Expected positive historical sales, got %f", forecast.Sales.TotalHistorical)
	}
	
	if forecast.Expenses.TotalHistorical <= 0 {
		t.Errorf("Expected positive historical expenses, got %f", forecast.Expenses.TotalHistorical)
	}
	
	// Validate that profit = sales - expenses (approximately)
	expectedProfit := forecast.Sales.TotalHistorical - forecast.Expenses.TotalHistorical
	if math.Abs(forecast.Profit.TotalHistorical-expectedProfit) > 1 {
		t.Errorf("Profit calculation incorrect. Expected %f, got %f", expectedProfit, forecast.Profit.TotalHistorical)
	}
	
	// Check that confidence is within valid range
	if forecast.Sales.Confidence < 0 || forecast.Sales.Confidence > 1 {
		t.Errorf("Sales confidence out of range: %f", forecast.Sales.Confidence)
	}
	
	if forecast.Expenses.Confidence < 0 || forecast.Expenses.Confidence > 1 {
		t.Errorf("Expenses confidence out of range: %f", forecast.Expenses.Confidence)
	}
	
	if forecast.Profit.Confidence < 0 || forecast.Profit.Confidence > 1 {
		t.Errorf("Profit confidence out of range: %f", forecast.Profit.Confidence)
	}
}

func TestForecastService_EmptyData(t *testing.T) {
	// Create empty mock repositories (without obsolete shopee repo)
	dropshipRepo := &mockForecastDropshipRepo{}
	journalRepo := &mockForecastJournalRepo{}
	
	forecastSvc := NewForecastService(dropshipRepo, journalRepo)
	
	// Create a forecast request for an empty shop
	now := time.Now()
	req := ForecastRequest{
		Shop:       "", // Empty shop should result in no data
		Period:     "monthly",
		StartDate:  now.AddDate(0, -1, 0),
		EndDate:    now,
		ForecastTo: now.AddDate(0, 1, 0),
	}
	
	// Generate forecast
	ctx := context.Background()
	forecast, err := forecastSvc.GenerateForecast(ctx, req)
	
	// Should still work but with limited data
	if err != nil {
		t.Fatalf("GenerateForecast with empty data failed: %v", err)
	}
	
	if forecast == nil {
		t.Fatal("Forecast result is nil")
	}
	
	// With empty data, totals should be low or zero, but structure should be intact
	if forecast.Sales.Method == "" {
		t.Error("Sales method should be set even with empty data")
	}
	
	if forecast.Expenses.Method == "" {
		t.Error("Expenses method should be set even with empty data")
	}
	
	if forecast.Profit.Method == "" {
		t.Error("Profit method should be set even with empty data")
	}
}