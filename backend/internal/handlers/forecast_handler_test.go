package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
	"github.com/ramadhan22/dropship-erp/backend/internal/service"
)

// Mock repositories for testing handlers
type mockHandlerDropshipRepo struct{}

func (m *mockHandlerDropshipRepo) ListDropshipPurchasesByShopAndDate(ctx context.Context, shop, from, to string) ([]models.DropshipPurchase, error) {
	return []models.DropshipPurchase{
		{
			KodePesanan:    "ORDER1",
			TotalTransaksi: 1000000,
		},
	}, nil
}

type mockHandlerJournalRepo struct{}

func (m *mockHandlerJournalRepo) GetAccountBalancesAsOf(ctx context.Context, shop string, asOfDate time.Time) ([]repository.AccountBalance, error) {
	return []repository.AccountBalance{
		{
			AccountCode: "5000",
			Balance:     500000,
		},
	}, nil
}

func TestForecastHandler_HandleGenerateForecast(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	
	// Create mock services (without obsolete shopee repo)
	dropshipRepo := &mockHandlerDropshipRepo{}
	journalRepo := &mockHandlerJournalRepo{}
	forecastSvc := service.NewForecastService(dropshipRepo, journalRepo)
	handler := NewForecastHandler(forecastSvc)
	
	// Create test request
	now := time.Now()
	req := service.ForecastRequest{
		Shop:       "testshop",
		Period:     "monthly",
		StartDate:  now.AddDate(0, -1, 0),
		EndDate:    now,
		ForecastTo: now.AddDate(0, 1, 0),
	}
	
	reqBody, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}
	
	// Create HTTP request
	httpReq, err := http.NewRequest("POST", "/api/forecast/generate", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to create HTTP request: %v", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	
	// Create response recorder
	w := httptest.NewRecorder()
	
	// Create Gin context
	router := gin.New()
	router.POST("/api/forecast/generate", handler.HandleGenerateForecast)
	
	// Perform request
	router.ServeHTTP(w, httpReq)
	
	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Response: %s", w.Code, w.Body.String())
	}
	
	// Parse response
	var response service.ForecastResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	
	// Validate response structure
	if response.Sales.Metric != "sales" {
		t.Errorf("Expected sales metric, got %s", response.Sales.Metric)
	}
	
	if response.Expenses.Metric != "expenses" {
		t.Errorf("Expected expenses metric, got %s", response.Expenses.Metric)
	}
	
	if response.Profit.Metric != "profit" {
		t.Errorf("Expected profit metric, got %s", response.Profit.Metric)
	}
	
	if response.Period != "monthly" {
		t.Errorf("Expected monthly period, got %s", response.Period)
	}
}

func TestForecastHandler_HandleGetForecastParams(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	
	// Create mock services (without obsolete shopee repo)
	dropshipRepo := &mockHandlerDropshipRepo{}
	journalRepo := &mockHandlerJournalRepo{}
	forecastSvc := service.NewForecastService(dropshipRepo, journalRepo)
	handler := NewForecastHandler(forecastSvc)
	
	// Create HTTP request
	httpReq, err := http.NewRequest("GET", "/api/forecast/params?shop=testshop&period=monthly", nil)
	if err != nil {
		t.Fatalf("Failed to create HTTP request: %v", err)
	}
	
	// Create response recorder
	w := httptest.NewRecorder()
	
	// Create Gin context
	router := gin.New()
	router.GET("/api/forecast/params", handler.HandleGetForecastParams)
	
	// Perform request
	router.ServeHTTP(w, httpReq)
	
	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Response: %s", w.Code, w.Body.String())
	}
	
	// Parse response
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	
	// Validate response has expected fields
	expectedFields := []string{"shop", "period", "suggestedStartDate", "suggestedEndDate", "suggestedForecastTo", "currentDate"}
	for _, field := range expectedFields {
		if _, ok := response[field]; !ok {
			t.Errorf("Response missing field: %s", field)
		}
	}
	
	if response["shop"] != "testshop" {
		t.Errorf("Expected shop 'testshop', got %v", response["shop"])
	}
	
	if response["period"] != "monthly" {
		t.Errorf("Expected period 'monthly', got %v", response["period"])
	}
}

func TestForecastHandler_HandleGetForecastSummary(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	
	// Create mock services (without obsolete shopee repo)
	dropshipRepo := &mockHandlerDropshipRepo{}
	journalRepo := &mockHandlerJournalRepo{}
	forecastSvc := service.NewForecastService(dropshipRepo, journalRepo)
	handler := NewForecastHandler(forecastSvc)
	
	// Create HTTP request
	httpReq, err := http.NewRequest("GET", "/api/forecast/summary?shop=testshop&period=monthly&days=30", nil)
	if err != nil {
		t.Fatalf("Failed to create HTTP request: %v", err)
	}
	
	// Create response recorder
	w := httptest.NewRecorder()
	
	// Create Gin context
	router := gin.New()
	router.GET("/api/forecast/summary", handler.HandleGetForecastSummary)
	
	// Perform request
	router.ServeHTTP(w, httpReq)
	
	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Response: %s", w.Code, w.Body.String())
	}
	
	// Parse response
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	
	// Validate response has expected fields
	expectedFields := []string{"shop", "period", "days", "forecastSales", "forecastExpenses", "forecastProfit"}
	for _, field := range expectedFields {
		if _, ok := response[field]; !ok {
			t.Errorf("Response missing field: %s", field)
		}
	}
	
	if response["shop"] != "testshop" {
		t.Errorf("Expected shop 'testshop', got %v", response["shop"])
	}
	
	if response["period"] != "monthly" {
		t.Errorf("Expected period 'monthly', got %v", response["period"])
	}
}

func TestForecastHandler_BadRequest(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	
	// Create mock services (without obsolete shopee repo)
	dropshipRepo := &mockHandlerDropshipRepo{}
	journalRepo := &mockHandlerJournalRepo{}
	forecastSvc := service.NewForecastService(dropshipRepo, journalRepo)
	handler := NewForecastHandler(forecastSvc)
	
	// Create invalid JSON request
	httpReq, err := http.NewRequest("POST", "/api/forecast/generate", bytes.NewBuffer([]byte("invalid json")))
	if err != nil {
		t.Fatalf("Failed to create HTTP request: %v", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	
	// Create response recorder
	w := httptest.NewRecorder()
	
	// Create Gin context
	router := gin.New()
	router.POST("/api/forecast/generate", handler.HandleGenerateForecast)
	
	// Perform request
	router.ServeHTTP(w, httpReq)
	
	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d. Response: %s", w.Code, w.Body.String())
	}
}