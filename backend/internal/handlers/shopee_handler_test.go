package handlers

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// fake service
type fakeShopeeService struct {
	errOn string
}

func (f *fakeShopeeService) ImportSettledOrdersCSV(ctx context.Context, path string) error {
	if path == f.errOn {
		return errors.New("fail import")
	}
	return nil
}

func TestShopeeHandleImport_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeShopeeService{}
	h := NewShopeeHandler(svc)

	rec := httptest.NewRecorder()
	r := gin.New()
	r.POST("/api/shopee/import", h.HandleImport)

	req := httptest.NewRequest("POST", "/api/shopee/import", bytes.NewBufferString(`{"file_path":"ok.csv"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestShopeeHandleImport_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeShopeeService{errOn: "bad.csv"}
	h := NewShopeeHandler(svc)

	rec := httptest.NewRecorder()
	r := gin.New()
	r.POST("/api/shopee/import", h.HandleImport)

	req := httptest.NewRequest("POST", "/api/shopee/import", bytes.NewBufferString(`{"file_path":"bad.csv"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}
