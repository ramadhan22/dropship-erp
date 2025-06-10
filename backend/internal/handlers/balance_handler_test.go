// File: backend/internal/handlers/balance_handler_test.go

package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/service"
)

// fakeBalanceService returns a canned balance sheet.
type fakeBalanceService struct {
	data []service.CategoryBalance
	err  error
}

func (f *fakeBalanceService) GetBalanceSheet(ctx context.Context, shop string, asOfDate time.Time) ([]service.CategoryBalance, error) {
	return f.data, f.err
}

func TestHandleGetBalanceSheet_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	expected := []service.CategoryBalance{
		{Category: "Assets", Total: 500},
		{Category: "Liabilities", Total: -200},
		{Category: "Equity", Total: 300},
	}
	svc := &fakeBalanceService{data: expected}
	h := NewBalanceHandler(svc)

	rec := httptest.NewRecorder()
	router := gin.New()
	router.GET("/api/balancesheet", h.HandleGetBalanceSheet)

	req := httptest.NewRequest("GET", "/api/balancesheet?shop=ShopA&period=2025-05", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var got []service.CategoryBalance
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}
	if len(got) != 3 || got[0].Total != 500 {
		t.Errorf("unexpected response: %+v", got)
	}
}

func TestHandleGetBalanceSheet_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeBalanceService{}
	h := NewBalanceHandler(svc)

	rec := httptest.NewRecorder()
	router := gin.New()
	router.GET("/api/balancesheet", h.HandleGetBalanceSheet)

	// invalid period format
	req := httptest.NewRequest("GET", "/api/balancesheet?shop=ShopA&period=bad", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandleGetBalanceSheet_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeBalanceService{err: errors.New("fail")}
	h := NewBalanceHandler(svc)

	rec := httptest.NewRecorder()
	router := gin.New()
	router.GET("/api/balancesheet", h.HandleGetBalanceSheet)

	req := httptest.NewRequest("GET", "/api/balancesheet?shop=ShopA&period=2025-05", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}
