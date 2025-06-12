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
)

// fake service
type fakeShopeeService struct {
	err bool
}

func (f *fakeShopeeService) ImportSettledOrdersXLSX(ctx context.Context, r io.Reader) (int, error) {
	if f.err {
		return 0, errors.New("fail import")
	}
	return 1, nil
}

func TestShopeeHandleImport_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeShopeeService{}
	h := NewShopeeHandler(svc)

	rec := httptest.NewRecorder()
	r := gin.New()
	r.POST("/api/shopee/import", h.HandleImport)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, _ := writer.CreateFormFile("file", "ok.xlsx")
	part.Write([]byte("xlsx"))
	writer.Close()

	req := httptest.NewRequest("POST", "/api/shopee/import", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestShopeeHandleImport_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeShopeeService{err: true}
	h := NewShopeeHandler(svc)

	rec := httptest.NewRecorder()
	r := gin.New()
	r.POST("/api/shopee/import", h.HandleImport)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, _ := writer.CreateFormFile("file", "bad.xlsx")
	part.Write([]byte("xlsx"))
	writer.Close()

	req := httptest.NewRequest("POST", "/api/shopee/import", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}
