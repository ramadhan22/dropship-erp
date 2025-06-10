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
