// File: backend/internal/handlers/dropship_handler_test.go

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

// fakeDropshipService implements the DropshipServiceInterface for testing.
type fakeDropshipService struct {
	errOn string
}

func (f *fakeDropshipService) ImportFromCSV(ctx context.Context, path string) error {
	if path == f.errOn {
		return errors.New("fail import")
	}
	return nil
}

func TestHandleImport_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeDropshipService{}
	// This is the real constructor from dropship_handler.go
	h := NewDropshipHandler(svc)

	rec := httptest.NewRecorder()
	router := gin.New()
	router.POST("/api/dropship/import", h.HandleImport)

	body := `{"file_path":"good.csv"}`
	req := httptest.NewRequest("POST", "/api/dropship/import", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
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

	// Missing file_path
	req := httptest.NewRequest("POST", "/api/dropship/import", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandleImport_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeDropshipService{errOn: "bad.csv"}
	h := NewDropshipHandler(svc)

	rec := httptest.NewRecorder()
	router := gin.New()
	router.POST("/api/dropship/import", h.HandleImport)

	body := `{"file_path":"bad.csv"}`
	req := httptest.NewRequest("POST", "/api/dropship/import", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}
