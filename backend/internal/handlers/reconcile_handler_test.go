package handlers

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// fake reconcile service
type fakeReconcileService struct {
	err bool
}

func (f *fakeReconcileService) MatchAndJournal(ctx context.Context, purchaseID, orderID, shop string) error {
	if f.err {
		return errors.New("fail reconcile")
	}
	return nil
}

func (f *fakeReconcileService) BulkReconcileWithErrorHandling(ctx context.Context, pairs [][2]string, shop string, batchID *int64) (*models.ReconciliationReport, error) {
	if f.err {
		return nil, errors.New("bulk reconcile error")
	}
	return &models.ReconciliationReport{
		TotalTransactions:      len(pairs),
		SuccessfulTransactions: len(pairs),
		FailedTransactions:     0,
		FailureRate:           0,
	}, nil
}

func (f *fakeReconcileService) GenerateReconciliationReport(ctx context.Context, shop string, since time.Time) (*models.ReconciliationReport, error) {
	if f.err {
		return nil, errors.New("report generation error")
	}
	return &models.ReconciliationReport{
		TotalTransactions:      10,
		SuccessfulTransactions: 8,
		FailedTransactions:     2,
		FailureRate:           20.0,
		FailureCategories:     map[string]int{"purchase_not_found": 2},
	}, nil
}

func (f *fakeReconcileService) RetryFailedReconciliations(ctx context.Context, shop string, maxRetries int) (*models.ReconciliationReport, error) {
	if f.err {
		return nil, errors.New("retry error")
	}
	return &models.ReconciliationReport{
		TotalTransactions:      5,
		SuccessfulTransactions: 3,
		FailedTransactions:     2,
		FailureRate:           40.0,
	}, nil
}

func (f *fakeReconcileService) GetFailedReconciliationsSummary(ctx context.Context, shop string, days int) (map[string]interface{}, error) {
	if f.err {
		return nil, errors.New("summary error")
	}
	return map[string]interface{}{
		"shop":               shop,
		"period_days":        days,
		"total_failed":       5,
		"failure_categories": map[string]int{"purchase_not_found": 3, "network_error": 2},
	}, nil
}

func TestReconcileHandleMatchAndJournal_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeReconcileService{}
	h := NewReconcileHandler(svc)

	rec := httptest.NewRecorder()
	r := gin.New()
	r.POST("/api/reconcile", h.HandleMatchAndJournal)

	body := `{"shop":"S","purchase_id":"P","order_id":"O"}`
	req := httptest.NewRequest("POST", "/api/reconcile", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestReconcileHandleMatchAndJournal_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeReconcileService{err: true}
	h := NewReconcileHandler(svc)

	rec := httptest.NewRecorder()
	r := gin.New()
	r.POST("/api/reconcile", h.HandleMatchAndJournal)

	body := `{"shop":"S","purchase_id":"P","order_id":"O"}`
	req := httptest.NewRequest("POST", "/api/reconcile", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestReconcileHandleBulkReconcileWithErrorHandling_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeReconcileService{}
	h := NewReconcileHandler(svc)

	rec := httptest.NewRecorder()
	r := gin.New()
	r.POST("/api/reconcile/bulk", h.HandleBulkReconcileWithErrorHandling)

	body := `{"shop":"S","pairs":[["P1","O1"],["P2","O2"]]}`
	req := httptest.NewRequest("POST", "/api/reconcile/bulk", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestReconcileHandleGenerateReconciliationReport_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeReconcileService{}
	h := NewReconcileHandler(svc)

	rec := httptest.NewRecorder()
	r := gin.New()
	r.GET("/api/reconcile/report", h.HandleGenerateReconciliationReport)

	req := httptest.NewRequest("GET", "/api/reconcile/report?shop=TestShop&days=30", nil)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestReconcileHandleRetryFailedReconciliations_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeReconcileService{}
	h := NewReconcileHandler(svc)

	rec := httptest.NewRecorder()
	r := gin.New()
	r.POST("/api/reconcile/retry", h.HandleRetryFailedReconciliations)

	body := `{"shop":"S","max_retries":10}`
	req := httptest.NewRequest("POST", "/api/reconcile/retry", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestReconcileHandleGetFailedReconciliationsSummary_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeReconcileService{}
	h := NewReconcileHandler(svc)

	rec := httptest.NewRecorder()
	r := gin.New()
	r.GET("/api/reconcile/summary", h.HandleGetFailedReconciliationsSummary)

	req := httptest.NewRequest("GET", "/api/reconcile/summary?shop=TestShop&days=7", nil)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
