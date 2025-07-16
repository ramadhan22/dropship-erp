package service

import (
	"context"
	"testing"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock for DropshipRepo interface
type MockDropshipRepo struct {
	mock.Mock
}

func (m *MockDropshipRepo) GetDropshipPurchaseByInvoice(ctx context.Context, kodeInvoice string) (*models.DropshipPurchase, error) {
	args := m.Called(ctx, kodeInvoice)
	return args.Get(0).(*models.DropshipPurchase), args.Error(1)
}

func (m *MockDropshipRepo) GetDropshipPurchaseByID(ctx context.Context, kodePesanan string) (*models.DropshipPurchase, error) {
	args := m.Called(ctx, kodePesanan)
	return args.Get(0).(*models.DropshipPurchase), args.Error(1)
}

func (m *MockDropshipRepo) UpdatePurchaseStatus(ctx context.Context, kodePesanan, status string) error {
	args := m.Called(ctx, kodePesanan, status)
	return args.Error(0)
}

func (m *MockDropshipRepo) SumDetailByInvoice(ctx context.Context, kodeInvoice string) (float64, error) {
	args := m.Called(ctx, kodeInvoice)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockDropshipRepo) SumProductCostByInvoice(ctx context.Context, kodeInvoice string) (float64, error) {
	args := m.Called(ctx, kodeInvoice)
	return args.Get(0).(float64), args.Error(1)
}

// Mock for JournalRepo interface
type MockJournalRepo struct {
	mock.Mock
}

func (m *MockJournalRepo) CreateJournalEntry(ctx context.Context, e *models.JournalEntry) (int64, error) {
	args := m.Called(ctx, e)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockJournalRepo) InsertJournalLine(ctx context.Context, l *models.JournalLine) error {
	args := m.Called(ctx, l)
	return args.Error(0)
}

// Mock for DetailRepo interface  
type MockDetailRepo struct {
	mock.Mock
}

func (m *MockDetailRepo) UpdateOrderDetailStatus(ctx context.Context, sn, status, orderStatus string, updateTime time.Time) error {
	args := m.Called(ctx, sn, status, orderStatus, updateTime)
	return args.Error(0)
}

func (m *MockDetailRepo) SaveOrderDetail(ctx context.Context, detail *models.ShopeeOrderDetailRow, items []models.ShopeeOrderItemRow, packages []models.ShopeeOrderPackageRow) error {
	args := m.Called(ctx, detail, items, packages)
	return args.Error(0)
}

func (m *MockDetailRepo) GetOrderDetail(ctx context.Context, sn string) (*models.ShopeeOrderDetailRow, []models.ShopeeOrderItemRow, []models.ShopeeOrderPackageRow, error) {
	args := m.Called(ctx, sn)
	return args.Get(0).(*models.ShopeeOrderDetailRow), 
		   args.Get(1).([]models.ShopeeOrderItemRow), 
		   args.Get(2).([]models.ShopeeOrderPackageRow), 
		   args.Error(3)
}

func TestReconcileService_createReturnedOrderJournal_FullReturn(t *testing.T) {
	// Setup mocks
	mockDropRepo := &MockDropshipRepo{}
	mockJournalRepo := &MockJournalRepo{}
	mockDetailRepo := &MockDetailRepo{}

	// Create service instance
	service := &ReconcileService{
		dropRepo:   mockDropRepo,
		journalRepo: mockJournalRepo,
		detailRepo: mockDetailRepo,
	}

	ctx := context.Background()
	invoice := "TEST123"
	status := "returned"
	updateTime := time.Now()
	isPartialReturn := false
	returnAmount := 0.0

	// Mock data
	dp := &models.DropshipPurchase{
		KodePesanan:        "ORD123",
		KodeInvoiceChannel: invoice,
		NamaToko:          "TestStore",
	}

	escDetail := ShopeeEscrowDetail{
		"order_income": map[string]any{
			"order_original_price": 100000.0,
			"commission_fee":       5000.0,
			"service_fee":         2000.0,
			"voucher_from_seller": 1000.0,
			"order_seller_discount": 500.0,
			"seller_shipping_discount": 1500.0,
			"order_ams_commission_fee": 3000.0,
		},
	}

	// Setup expectations
	mockDropRepo.On("GetDropshipPurchaseByInvoice", ctx, invoice).Return(dp, nil)
	mockJournalRepo.On("CreateJournalEntry", ctx, mock.AnythingOfType("*models.JournalEntry")).Return(int64(1), nil)
	mockJournalRepo.On("InsertJournalLine", ctx, mock.AnythingOfType("*models.JournalLine")).Return(nil)
	mockDetailRepo.On("UpdateOrderDetailStatus", ctx, invoice, "returned", "returned", updateTime).Return(nil)
	mockDropRepo.On("UpdatePurchaseStatus", ctx, "ORD123", "Pesanan dikembalikan").Return(nil)

	// Execute
	err := service.createReturnedOrderJournal(ctx, invoice, status, updateTime, &escDetail, isPartialReturn, returnAmount)

	// Assert
	assert.NoError(t, err)
	
	// Verify journal entry was created
	mockJournalRepo.AssertCalled(t, "CreateJournalEntry", ctx, mock.MatchedBy(func(entry *models.JournalEntry) bool {
		return entry.SourceType == "shopee_return" && 
			   entry.SourceID == invoice &&
			   entry.Store == "TestStore"
	}))

	// Verify journal lines were created (should have 8 lines for full return)
	assert.Equal(t, 8, len(mockJournalRepo.Calls)-1) // -1 for CreateJournalEntry call
	
	// Verify status updates
	mockDetailRepo.AssertCalled(t, "UpdateOrderDetailStatus", ctx, invoice, "returned", "returned", updateTime)
	mockDropRepo.AssertCalled(t, "UpdatePurchaseStatus", ctx, "ORD123", "Pesanan dikembalikan")

	mockDropRepo.AssertExpectations(t)
	mockJournalRepo.AssertExpectations(t)
	mockDetailRepo.AssertExpectations(t)
}

func TestReconcileService_createReturnedOrderJournal_PartialReturn(t *testing.T) {
	// Setup mocks
	mockDropRepo := &MockDropshipRepo{}
	mockJournalRepo := &MockJournalRepo{}
	mockDetailRepo := &MockDetailRepo{}

	// Create service instance
	service := &ReconcileService{
		dropRepo:   mockDropRepo,
		journalRepo: mockJournalRepo,
		detailRepo: mockDetailRepo,
	}

	ctx := context.Background()
	invoice := "TEST456"
	status := "partial_return"
	updateTime := time.Now()
	isPartialReturn := true
	returnAmount := 50000.0 // Half of original order

	// Mock data
	dp := &models.DropshipPurchase{
		KodePesanan:        "ORD456",
		KodeInvoiceChannel: invoice,
		NamaToko:          "TestStore",
	}

	escDetail := ShopeeEscrowDetail{
		"order_income": map[string]any{
			"order_original_price": 100000.0,
			"commission_fee":       5000.0,
			"service_fee":         2000.0,
			"voucher_from_seller": 1000.0,
			"order_seller_discount": 500.0,
			"seller_shipping_discount": 1500.0,
			"order_ams_commission_fee": 3000.0,
		},
	}

	// Setup expectations
	mockDropRepo.On("GetDropshipPurchaseByInvoice", ctx, invoice).Return(dp, nil)
	mockJournalRepo.On("CreateJournalEntry", ctx, mock.AnythingOfType("*models.JournalEntry")).Return(int64(2), nil)
	mockJournalRepo.On("InsertJournalLine", ctx, mock.AnythingOfType("*models.JournalLine")).Return(nil)
	mockDetailRepo.On("UpdateOrderDetailStatus", ctx, invoice, "partial_return", "partial_return", updateTime).Return(nil)
	mockDropRepo.On("UpdatePurchaseStatus", ctx, "ORD456", "Sebagian dikembalikan").Return(nil)

	// Execute
	err := service.createReturnedOrderJournal(ctx, invoice, status, updateTime, &escDetail, isPartialReturn, returnAmount)

	// Assert
	assert.NoError(t, err)
	
	// Verify journal entry description includes partial info
	mockJournalRepo.AssertCalled(t, "CreateJournalEntry", ctx, mock.MatchedBy(func(entry *models.JournalEntry) bool {
		return entry.SourceType == "shopee_return" && 
			   entry.SourceID == invoice &&
			   entry.Description != nil &&
			   *entry.Description == "Shopee return TEST456 (partial 50000.00)"
	}))

	// Verify status updates for partial return
	mockDetailRepo.AssertCalled(t, "UpdateOrderDetailStatus", ctx, invoice, "partial_return", "partial_return", updateTime)
	mockDropRepo.AssertCalled(t, "UpdatePurchaseStatus", ctx, "ORD456", "Sebagian dikembalikan")

	mockDropRepo.AssertExpectations(t)
	mockJournalRepo.AssertExpectations(t)
	mockDetailRepo.AssertExpectations(t)
}

func TestReconcileService_UpdateShopeeStatus_ReturnedOrder(t *testing.T) {
	// This test would require more complex mocking of the ShopeeClient
	// For now, we'll focus on the main returned order journal logic
	t.Skip("Integration test - requires full mock setup of ShopeeClient and GetShopeeEscrowDetail")
}