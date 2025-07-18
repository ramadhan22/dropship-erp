package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// Test the new error handling functionality

func TestBulkReconcileWithErrorHandling_Success(t *testing.T) {
	ctx := context.Background()

	// Prepare fake repos with preloaded data
	fDrop := &fakeDropRepoRec{
		data: map[string]*models.DropshipPurchase{
			"DP-111": {KodePesanan: "DP-111", TotalTransaksi: 50.00},
			"DP-222": {KodePesanan: "DP-222", TotalTransaksi: 75.00},
		},
	}
	fShopee := &fakeShopeeRepoRec{
		data: map[string]*models.ShopeeSettledOrder{
			"SO-111": {OrderID: "SO-111", NetIncome: 45.00, SettledDate: time.Now()},
			"SO-222": {OrderID: "SO-222", NetIncome: 70.00, SettledDate: time.Now()},
		},
	}
	fJournal := &fakeJournalRepoRec{nextID: 0}
	fRec := &fakeRecRepoRec{}
	fDetail := &fakeDetailRepo{}
	fFailed := &fakeFailedRecRepo{}

	svc := NewReconcileService(nil, fDrop, fShopee, fJournal, fRec, nil, fDetail, nil, nil, nil, fFailed, nil, 5, nil)

	pairs := [][2]string{
		{"DP-111", "SO-111"},
		{"DP-222", "SO-222"},
	}

	report, err := svc.BulkReconcileWithErrorHandling(ctx, pairs, "ShopA", nil)
	if err != nil {
		t.Fatalf("BulkReconcileWithErrorHandling error: %v", err)
	}

	// Verify report
	if report.TotalTransactions != 2 {
		t.Errorf("expected 2 total transactions, got %d", report.TotalTransactions)
	}
	if report.SuccessfulTransactions != 2 {
		t.Errorf("expected 2 successful transactions, got %d", report.SuccessfulTransactions)
	}
	if report.FailedTransactions != 0 {
		t.Errorf("expected 0 failed transactions, got %d", report.FailedTransactions)
	}
	if report.FailureRate != 0 {
		t.Errorf("expected 0%% failure rate, got %.2f%%", report.FailureRate)
	}
}

func TestBulkReconcileWithErrorHandling_WithFailures(t *testing.T) {
	ctx := context.Background()

	// Prepare fake repos with missing data to cause failures
	fDrop := &fakeDropRepoRec{
		data: map[string]*models.DropshipPurchase{
			"DP-111": {KodePesanan: "DP-111", TotalTransaksi: 50.00},
			// DP-222 is missing - will cause error
		},
	}
	fShopee := &fakeShopeeRepoRec{
		data: map[string]*models.ShopeeSettledOrder{
			"SO-111": {OrderID: "SO-111", NetIncome: 45.00, SettledDate: time.Now()},
			"SO-222": {OrderID: "SO-222", NetIncome: 70.00, SettledDate: time.Now()},
		},
	}
	fJournal := &fakeJournalRepoRec{nextID: 0}
	fRec := &fakeRecRepoRec{}
	fDetail := &fakeDetailRepo{}
	fFailed := &fakeFailedRecRepo{}

	svc := NewReconcileService(nil, fDrop, fShopee, fJournal, fRec, nil, fDetail, nil, nil, nil, fFailed, nil, 5, nil)

	pairs := [][2]string{
		{"DP-111", "SO-111"},
		{"DP-222", "SO-222"}, // This will fail because DP-222 is missing
	}

	report, err := svc.BulkReconcileWithErrorHandling(ctx, pairs, "ShopA", nil)
	if err != nil {
		t.Fatalf("BulkReconcileWithErrorHandling error: %v", err)
	}

	// Verify report
	if report.TotalTransactions != 2 {
		t.Errorf("expected 2 total transactions, got %d", report.TotalTransactions)
	}
	if report.SuccessfulTransactions != 1 {
		t.Errorf("expected 1 successful transaction, got %d", report.SuccessfulTransactions)
	}
	if report.FailedTransactions != 1 {
		t.Errorf("expected 1 failed transaction, got %d", report.FailedTransactions)
	}
	if report.FailureRate != 50.0 {
		t.Errorf("expected 50%% failure rate, got %.2f%%", report.FailureRate)
	}

	// Verify failed transaction was recorded
	if len(fFailed.failures) != 1 {
		t.Errorf("expected 1 recorded failure, got %d", len(fFailed.failures))
	}

	if len(fFailed.failures) > 0 {
		failure := fFailed.failures[0]
		if failure.PurchaseID != "DP-222" {
			t.Errorf("expected failed purchase ID DP-222, got %s", failure.PurchaseID)
		}
		if failure.Shop != "ShopA" {
			t.Errorf("expected failed shop ShopA, got %s", failure.Shop)
		}
		if failure.ErrorType != "purchase_not_found" {
			t.Errorf("expected error type purchase_not_found, got %s", failure.ErrorType)
		}
	}
}

func TestCategorizeError(t *testing.T) {
	svc := &ReconcileService{}

	tests := []struct {
		name         string
		errorMessage string
		expectedType string
	}{
		{
			name:         "purchase not found",
			errorMessage: "fetch DropshipPurchase DP-123: not found",
			expectedType: "purchase_not_found",
		},
		{
			name:         "shopee order not found",
			errorMessage: "fetch ShopeeOrder SO-456: not found",
			expectedType: "shopee_order_not_found",
		},
		{
			name:         "database error",
			errorMessage: "database error: connection failed",
			expectedType: "database_error",
		},
		{
			name:         "network error",
			errorMessage: "connection refused",
			expectedType: "network_error",
		},
		{
			name:         "timeout error",
			errorMessage: "context deadline exceeded",
			expectedType: "timeout_error",
		},
		{
			name:         "journal creation error",
			errorMessage: "create JournalEntry: failed",
			expectedType: "journal_creation_error",
		},
		{
			name:         "journal balance error",
			errorMessage: "unbalanced journal: debit 100 credit 90",
			expectedType: "journal_balance_error",
		},
		{
			name:         "authentication error",
			errorMessage: "invalid access_token",
			expectedType: "authentication_error",
		},
		{
			name:         "unknown error",
			errorMessage: "some unexpected error",
			expectedType: "unknown_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.categorizeError(errors.New(tt.errorMessage))
			if result != tt.expectedType {
				t.Errorf("categorizeError(%q) = %q, want %q", tt.errorMessage, result, tt.expectedType)
			}
		})
	}
}

func TestShouldHaltProcessing(t *testing.T) {
	config := &models.ReconciliationConfig{
		MaxAllowedFailures:      5,
		FailureThresholdPercent: 10.0,
		CriticalErrorTypes:      []string{"database_error", "critical_system_error"},
	}
	svc := &ReconcileService{config: config}

	tests := []struct {
		name                   string
		successfulTransactions int
		failedTransactions     int
		errorType              string
		expectedHalt           bool
	}{
		{
			name:                   "should not halt - low failure count",
			successfulTransactions: 18,
			failedTransactions:     1, // 1/19 = ~5% < 10%
			errorType:              "purchase_not_found",
			expectedHalt:           false,
		},
		{
			name:                   "should halt - max failures exceeded",
			successfulTransactions: 10,
			failedTransactions:     6,
			errorType:              "purchase_not_found",
			expectedHalt:           true,
		},
		{
			name:                   "should halt - critical error type",
			successfulTransactions: 10,
			failedTransactions:     1,
			errorType:              "database_error",
			expectedHalt:           true,
		},
		{
			name:                   "should halt - failure rate too high",
			successfulTransactions: 8,
			failedTransactions:     3, // 3/11 = ~27% > 10%
			errorType:              "purchase_not_found",
			expectedHalt:           true,
		},
		{
			name:                   "should not halt - low failure rate",
			successfulTransactions: 18,
			failedTransactions:     2, // 2/20 = 10%
			errorType:              "purchase_not_found",
			expectedHalt:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := &models.ReconciliationReport{
				TotalTransactions:      tt.successfulTransactions + tt.failedTransactions,
				SuccessfulTransactions: tt.successfulTransactions,
				FailedTransactions:     tt.failedTransactions,
			}

			result := svc.shouldHaltProcessing(report, tt.errorType)
			if result != tt.expectedHalt {
				t.Errorf("shouldHaltProcessing() = %v, want %v", result, tt.expectedHalt)
			}
		})
	}
}

func TestGenerateReconciliationReport(t *testing.T) {
	ctx := context.Background()

	fFailed := &fakeFailedRecRepo{
		failures: []models.FailedReconciliation{
			{
				PurchaseID: "DP-1",
				Shop:       "ShopA",
				ErrorType:  "purchase_not_found",
				FailedAt:   time.Now(),
			},
			{
				PurchaseID: "DP-2",
				Shop:       "ShopA",
				ErrorType:  "purchase_not_found",
				FailedAt:   time.Now(),
			},
			{
				PurchaseID: "DP-3",
				Shop:       "ShopA",
				ErrorType:  "network_error",
				FailedAt:   time.Now(),
			},
		},
	}

	config := &models.ReconciliationConfig{
		GenerateDetailedReport: true,
	}

	svc := &ReconcileService{
		failedRepo: fFailed,
		config:     config,
	}

	since := time.Now().Add(-24 * time.Hour)
	report, err := svc.GenerateReconciliationReport(ctx, "ShopA", since)
	if err != nil {
		t.Fatalf("GenerateReconciliationReport error: %v", err)
	}

	if report.FailedTransactions != 3 {
		t.Errorf("expected 3 failed transactions, got %d", report.FailedTransactions)
	}

	if len(report.FailureCategories) != 2 {
		t.Errorf("expected 2 failure categories, got %d", len(report.FailureCategories))
	}

	if report.FailureCategories["purchase_not_found"] != 2 {
		t.Errorf("expected 2 purchase_not_found failures, got %d", report.FailureCategories["purchase_not_found"])
	}

	if report.FailureCategories["network_error"] != 1 {
		t.Errorf("expected 1 network_error failure, got %d", report.FailureCategories["network_error"])
	}

	if len(report.FailedTransactionList) != 3 {
		t.Errorf("expected 3 failed transactions in detailed list, got %d", len(report.FailedTransactionList))
	}
}
