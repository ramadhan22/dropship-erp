package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/service"
)

type fakePLReportSvc struct {
	data *service.ProfitLoss
	err  error
}

func (f *fakePLReportSvc) GetProfitLoss(ctx context.Context, typ string, month, year int, store string) (*service.ProfitLoss, error) {
	return f.data, f.err
}

func TestProfitLossReportHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakePLReportSvc{data: &service.ProfitLoss{TotalPendapatanUsaha: 100}}
	h := NewProfitLossReportHandler(svc)

	rec := httptest.NewRecorder()
	router := gin.New()
	h.RegisterRoutes(router)

	req := httptest.NewRequest("GET", "/profitloss?type=Monthly&month=5&year=2025", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rec.Code)
	}
	var got service.ProfitLoss
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.TotalPendapatanUsaha != 100 {
		t.Errorf("unexpected body: %+v", got)
	}
}
