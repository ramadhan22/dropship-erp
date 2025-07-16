package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/ramadhan22/dropship-erp/backend/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock for ReconcileExtraService interface
type MockReconcileExtraService struct {
	mock.Mock
}

func (m *MockReconcileExtraService) ListUnmatched(ctx context.Context, shop string) ([]models.ReconciledTransaction, error) {
	args := m.Called(ctx, shop)
	return args.Get(0).([]models.ReconciledTransaction), args.Error(1)
}

func (m *MockReconcileExtraService) ListCandidates(ctx context.Context, shop, order, from, to string, limit, offset int) ([]models.ReconcileCandidate, int, error) {
	args := m.Called(ctx, shop, order, from, to, limit, offset)
	return args.Get(0).([]models.ReconcileCandidate), args.Get(1).(int), args.Error(2)
}

func (m *MockReconcileExtraService) BulkReconcile(ctx context.Context, pairs [][2]string, shop string) error {
	args := m.Called(ctx, pairs, shop)
	return args.Error(0)
}

func (m *MockReconcileExtraService) CheckAndMarkComplete(ctx context.Context, kodePesanan string) error {
	args := m.Called(ctx, kodePesanan)
	return args.Error(0)
}

func (m *MockReconcileExtraService) GetShopeeOrderStatus(ctx context.Context, invoice string) (string, error) {
	args := m.Called(ctx, invoice)
	return args.Get(0).(string), args.Error(1)
}

func (m *MockReconcileExtraService) GetShopeeOrderDetail(ctx context.Context, invoice string) (*service.ShopeeOrderDetail, error) {
	args := m.Called(ctx, invoice)
	return args.Get(0).(*service.ShopeeOrderDetail), args.Error(1)
}

func (m *MockReconcileExtraService) GetShopeeEscrowDetail(ctx context.Context, invoice string) (*service.ShopeeEscrowDetail, error) {
	args := m.Called(ctx, invoice)
	return args.Get(0).(*service.ShopeeEscrowDetail), args.Error(1)
}

func (m *MockReconcileExtraService) GetShopeeAccessToken(ctx context.Context, invoice string) (string, error) {
	args := m.Called(ctx, invoice)
	return args.Get(0).(string), args.Error(1)
}

func (m *MockReconcileExtraService) CancelPurchase(ctx context.Context, kodePesanan string) error {
	args := m.Called(ctx, kodePesanan)
	return args.Error(0)
}

func (m *MockReconcileExtraService) UpdateShopeeStatus(ctx context.Context, invoice string) error {
	args := m.Called(ctx, invoice)
	return args.Error(0)
}

func (m *MockReconcileExtraService) UpdateShopeeStatuses(ctx context.Context, invoices []string) error {
	args := m.Called(ctx, invoices)
	return args.Error(0)
}

func (m *MockReconcileExtraService) CreateReconcileBatches(ctx context.Context, shop, order, from, to string) (*models.ReconcileBatchInfo, error) {
	args := m.Called(ctx, shop, order, from, to)
	return args.Get(0).(*models.ReconcileBatchInfo), args.Error(1)
}

func (m *MockReconcileExtraService) ProcessReturnedOrder(ctx context.Context, invoice string, isPartialReturn bool, returnAmount float64) error {
	args := m.Called(ctx, invoice, isPartialReturn, returnAmount)
	return args.Error(0)
}

func (m *MockReconcileExtraService) HasReturnJournal(ctx context.Context, invoice string) bool {
	args := m.Called(ctx, invoice)
	return args.Get(0).(bool)
}

func TestReconcileExtraHandler_processReturn_FullReturn(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockService := &MockReconcileExtraService{}
	handler := NewReconcileExtraHandler(mockService)

	// Mock expectations
	mockService.On("HasReturnJournal", mock.Anything, "TEST123").Return(false)
	mockService.On("ProcessReturnedOrder", mock.Anything, "TEST123", false, float64(0)).Return(nil)

	// Create request
	reqBody := map[string]interface{}{
		"invoice":           "TEST123",
		"is_partial_return": false,
		"return_amount":     0,
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Setup HTTP test
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/reconcile/return", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute
	handler.processReturn(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Equal(t, "Return processed successfully", response["message"])
	assert.Equal(t, "TEST123", response["invoice"])
	assert.Equal(t, "full", response["return_type"])
	assert.Equal(t, float64(0), response["amount"])

	mockService.AssertExpectations(t)
}

func TestReconcileExtraHandler_processReturn_PartialReturn(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockService := &MockReconcileExtraService{}
	handler := NewReconcileExtraHandler(mockService)

	// Mock expectations
	mockService.On("HasReturnJournal", mock.Anything, "TEST456").Return(false)
	mockService.On("ProcessReturnedOrder", mock.Anything, "TEST456", true, float64(50000)).Return(nil)

	// Create request
	reqBody := map[string]interface{}{
		"invoice":           "TEST456",
		"is_partial_return": true,
		"return_amount":     50000,
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Setup HTTP test
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/reconcile/return", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute
	handler.processReturn(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Equal(t, "Return processed successfully", response["message"])
	assert.Equal(t, "TEST456", response["invoice"])
	assert.Equal(t, "partial", response["return_type"])
	assert.Equal(t, float64(50000), response["amount"])

	mockService.AssertExpectations(t)
}

func TestReconcileExtraHandler_processReturn_AlreadyProcessed(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockService := &MockReconcileExtraService{}
	handler := NewReconcileExtraHandler(mockService)

	// Mock expectations - return already exists
	mockService.On("HasReturnJournal", mock.Anything, "TEST789").Return(true)

	// Create request
	reqBody := map[string]interface{}{
		"invoice":           "TEST789",
		"is_partial_return": false,
		"return_amount":     0,
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Setup HTTP test
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/reconcile/return", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute
	handler.processReturn(c)

	// Assert
	assert.Equal(t, http.StatusConflict, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Equal(t, "Return journal already exists for this invoice", response["error"])

	mockService.AssertExpectations(t)
}

func TestReconcileExtraHandler_processReturn_InvalidPartialReturn(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockService := &MockReconcileExtraService{}
	handler := NewReconcileExtraHandler(mockService)

	// Create request with invalid partial return (amount <= 0)
	reqBody := map[string]interface{}{
		"invoice":           "TEST999",
		"is_partial_return": true,
		"return_amount":     0, // Invalid for partial return
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Setup HTTP test
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/reconcile/return", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute
	handler.processReturn(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Equal(t, "Return amount must be greater than 0 for partial returns", response["error"])

	// Should not call any service methods for invalid requests
	mockService.AssertNotCalled(t, "HasReturnJournal")
	mockService.AssertNotCalled(t, "ProcessReturnedOrder")
}