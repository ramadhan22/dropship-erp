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

type fakeDashService struct{}

func (f *fakeDashService) GetDashboardData(ctx context.Context, _ service.DashboardFilters) (*service.DashboardData, error) {
	return &service.DashboardData{Summary: map[string]service.SummaryItem{"total_orders": {Value: 1200}}}, nil
}

func TestDashboardHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDashboardHandler(&fakeDashService{})
	r := gin.New()
	h.RegisterRoutes(r)

	req := httptest.NewRequest("GET", "/dashboard", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rec.Code)
	}
	var resp struct {
		Summary map[string]any `json:"summary"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json: %v", err)
	}
	if resp.Summary["total_orders"].(map[string]any)["value"] != 1200.0 {
		t.Errorf("unexpected summary data: %v", resp.Summary)
	}
}
