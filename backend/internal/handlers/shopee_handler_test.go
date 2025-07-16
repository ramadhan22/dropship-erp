package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// fake service
type fakeShopeeService struct {
	err bool
}

func (f *fakeShopeeService) ImportAffiliateCSV(ctx context.Context, r io.Reader) (int, error) {
	if f.err {
		return 0, errors.New("fail import")
	}
	return 1, nil
}

func (f *fakeShopeeService) ListAffiliate(ctx context.Context, date, month, year string, limit, offset int) ([]models.ShopeeAffiliateSale, int, error) {
	return nil, 0, nil
}

func (f *fakeShopeeService) SumAffiliate(ctx context.Context, date, month, year string) (*models.ShopeeAffiliateSummary, error) {
	return &models.ShopeeAffiliateSummary{}, nil
}

func (f *fakeShopeeService) ListSalesProfit(ctx context.Context, channel, store, from, to, orderNo, sortBy, dir string, limit, offset int) ([]models.SalesProfit, int, error) {
	return nil, 0, nil
}

func TestShopeeHandleImportAffiliate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeShopeeService{}
	h := NewShopeeHandler(svc)

	rec := httptest.NewRecorder()
	r := gin.New()
	r.POST("/api/shopee/affiliate", h.HandleImportAffiliate)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, _ := writer.CreateFormFile("file", "a.csv")
	part.Write([]byte("csv"))
	part, _ = writer.CreateFormFile("file", "b.csv")
	part.Write([]byte("csv"))
	writer.Close()

	req := httptest.NewRequest("POST", "/api/shopee/affiliate", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp struct {
		Inserted int `json:"inserted"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}
	if resp.Inserted != 2 {
		t.Fatalf("expected inserted 2, got %d", resp.Inserted)
	}
}
