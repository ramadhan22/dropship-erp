package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestDashboardHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDashboardHandler()
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
