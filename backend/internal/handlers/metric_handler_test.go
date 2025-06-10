// File: backend/internal/handlers/metric_handler_test.go

package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/ramadhan22/dropship-erp/backend/internal/service"
)

// fakeMetricService implements just the methods MetricHandler needs.
type fakeMetricService struct {
	calcErr   error
	returnCM  *models.CachedMetric
	returnErr error
}

func (f *fakeMetricService) CalculateAndCacheMetrics(ctx context.Context, shop, period string) error {
	return f.calcErr
}
func (f *fakeMetricService) MetricRepo() service.MetricRepoInterface {
	return f
}
func (f *fakeMetricService) GetCachedMetric(ctx context.Context, shop, period string) (*models.CachedMetric, error) {
	return f.returnCM, f.returnErr
}

func TestHandleCalculateMetrics_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeMetricService{}
	h := NewMetricHandler(svc)

	rec := httptest.NewRecorder()
	router := gin.New()
	router.POST("/api/metrics", h.HandleCalculateMetrics)

	body := `{"shop":"ShopX","period":"2025-05"}`
	req := httptest.NewRequest("POST", "/api/metrics", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestHandleCalculateMetrics_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeMetricService{}
	h := NewMetricHandler(svc)

	rec := httptest.NewRecorder()
	router := gin.New()
	router.POST("/api/metrics", h.HandleCalculateMetrics)

	// missing fields
	req := httptest.NewRequest("POST", "/api/metrics", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandleCalculateMetrics_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeMetricService{calcErr: errors.New("fail")}
	h := NewMetricHandler(svc)

	rec := httptest.NewRecorder()
	router := gin.New()
	router.POST("/api/metrics", h.HandleCalculateMetrics)

	body := `{"shop":"ShopX","period":"2025-05"}`
	req := httptest.NewRequest("POST", "/api/metrics", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestHandleGetMetrics_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	expected := &models.CachedMetric{
		ShopUsername:      "ShopX",
		Period:            "2025-05",
		SumRevenue:        100,
		SumCOGS:           50,
		SumFees:           5,
		NetProfit:         45,
		EndingCashBalance: 200,
	}
	svc := &fakeMetricService{returnCM: expected}
	h := NewMetricHandler(svc)

	rec := httptest.NewRecorder()
	router := gin.New()
	router.GET("/api/metrics", h.HandleGetMetrics)

	req := httptest.NewRequest("GET", "/api/metrics?shop=ShopX&period=2025-05", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var got models.CachedMetric
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}
	if got.NetProfit != expected.NetProfit {
		t.Errorf("expected NetProfit %f, got %f", expected.NetProfit, got.NetProfit)
	}
}

func TestHandleGetMetrics_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeMetricService{returnErr: errors.New("missing")}
	h := NewMetricHandler(svc)

	rec := httptest.NewRecorder()
	router := gin.New()
	router.GET("/api/metrics", h.HandleGetMetrics)

	req := httptest.NewRequest("GET", "/api/metrics?shop=ShopX&period=2025-05", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

// inside fakeMetricService in metric_handler_test.go
func (f *fakeMetricService) UpsertCachedMetric(ctx context.Context, m *models.CachedMetric) error {
	return nil
}
