// File: backend/internal/handlers/dropship_handler_test.go

package handlers

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// fakeDropshipService implements the DropshipServiceInterface for testing.
type fakeDropshipService struct {
	fail bool
}

func (f *fakeDropshipService) ImportFromCSV(ctx context.Context, r io.Reader) (int, error) {
	if f.fail {
		return 0, errors.New("fail import")
	}
	return 1, nil
}

func (f *fakeDropshipService) ListDropshipPurchases(ctx context.Context, channel, store, from, to string, limit, offset int) ([]models.DropshipPurchase, int, error) {
	return nil, 0, nil
}

func (f *fakeDropshipService) GetDropshipPurchaseByID(ctx context.Context, kode string) (*models.DropshipPurchase, error) {
	return nil, nil
}

func (f *fakeDropshipService) ListDropshipPurchaseDetails(ctx context.Context, kode string) ([]models.DropshipPurchaseDetail, error) {
	return nil, nil
}

func (f *fakeDropshipService) SumDropshipPurchases(ctx context.Context, channel, store, from, to string) (float64, error) {
	return 0, nil
}

func (f *fakeDropshipService) TopProducts(ctx context.Context, channel, store, from, to string, limit int) ([]models.ProductSales, error) {
	return nil, nil
}

func TestHandleImport_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeDropshipService{}
	// This is the real constructor from dropship_handler.go
	h := NewDropshipHandler(svc)

	rec := httptest.NewRecorder()
	router := gin.New()
	router.POST("/api/dropship/import", h.HandleImport)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, _ := writer.CreateFormFile("file", "good.csv")
	part.Write([]byte("csv"))
	writer.Close()

	req := httptest.NewRequest("POST", "/api/dropship/import", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestHandleImport_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeDropshipService{}
	h := NewDropshipHandler(svc)

	rec := httptest.NewRecorder()
	router := gin.New()
	router.POST("/api/dropship/import", h.HandleImport)

	req := httptest.NewRequest("POST", "/api/dropship/import", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandleImport_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeDropshipService{fail: true}
	h := NewDropshipHandler(svc)

	rec := httptest.NewRecorder()
	router := gin.New()
	router.POST("/api/dropship/import", h.HandleImport)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, _ := writer.CreateFormFile("file", "bad.csv")
	part.Write([]byte("csv"))
	writer.Close()

	req := httptest.NewRequest("POST", "/api/dropship/import", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}
