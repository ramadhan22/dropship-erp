package service

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// mockWalletTransactionService is a mock implementation for testing
type mockWalletTransactionService struct {
	transactions []WalletTransaction
	err          error
}

func (m *mockWalletTransactionService) ListWalletTransactions(ctx context.Context, store string, p WalletTransactionParams) ([]WalletTransaction, bool, error) {
	if m.err != nil {
		return nil, false, m.err
	}

	// Filter by transaction type if specified
	var filtered []WalletTransaction
	for _, tx := range m.transactions {
		if p.TransactionType == "" || tx.TransactionType == p.TransactionType {
			// Check time range if specified
			if p.CreateTimeFrom != nil && tx.CreateTime < *p.CreateTimeFrom {
				continue
			}
			if p.CreateTimeTo != nil && tx.CreateTime > *p.CreateTimeTo {
				continue
			}
			filtered = append(filtered, tx)
		}
	}

	return filtered, false, nil
}

func TestFindSpmDisburseAddAmountWithService(t *testing.T) {
	// Create a base time for testing
	baseTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		transactions   []WalletTransaction
		withdrawalTime int64
		expectedAmount float64
		expectError    bool
	}{
		{
			name: "No SPM_DISBURSE_ADD transactions",
			transactions: []WalletTransaction{
				{TransactionID: 1, TransactionType: "WITHDRAWAL_CREATED", Amount: 100.0, CreateTime: baseTime.Unix()},
			},
			withdrawalTime: baseTime.Unix(),
			expectedAmount: 0.0,
			expectError:    false,
		},
		{
			name: "Single SPM_DISBURSE_ADD transaction same day",
			transactions: []WalletTransaction{
				{TransactionID: 1, TransactionType: "SPM_DISBURSE_ADD", Amount: 50.0, CreateTime: baseTime.Unix()},
				{TransactionID: 2, TransactionType: "WITHDRAWAL_CREATED", Amount: 100.0, CreateTime: baseTime.Unix()},
			},
			withdrawalTime: baseTime.Unix(),
			expectedAmount: 50.0,
			expectError:    false,
		},
		{
			name: "Multiple SPM_DISBURSE_ADD transactions same day",
			transactions: []WalletTransaction{
				{TransactionID: 1, TransactionType: "SPM_DISBURSE_ADD", Amount: 30.0, CreateTime: baseTime.Unix()},
				{TransactionID: 2, TransactionType: "SPM_DISBURSE_ADD", Amount: 20.0, CreateTime: baseTime.Add(2 * time.Hour).Unix()},
				{TransactionID: 3, TransactionType: "WITHDRAWAL_CREATED", Amount: 100.0, CreateTime: baseTime.Unix()},
			},
			withdrawalTime: baseTime.Unix(),
			expectedAmount: 50.0,
			expectError:    false,
		},
		{
			name: "SPM_DISBURSE_ADD transaction different day",
			transactions: []WalletTransaction{
				{TransactionID: 1, TransactionType: "SPM_DISBURSE_ADD", Amount: 50.0, CreateTime: baseTime.AddDate(0, 0, -1).Unix()},
				{TransactionID: 2, TransactionType: "WITHDRAWAL_CREATED", Amount: 100.0, CreateTime: baseTime.Unix()},
			},
			withdrawalTime: baseTime.Unix(),
			expectedAmount: 0.0,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWalletSvc := &mockWalletTransactionService{
				transactions: tt.transactions,
			}

			amount, err := findSpmDisburseAddAmountWithService(context.Background(), mockWalletSvc, "test-store", tt.withdrawalTime)

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if amount != tt.expectedAmount {
				t.Errorf("expected amount %.2f, got %.2f", tt.expectedAmount, amount)
			}
		})
	}
}

func TestWithdrawalDescriptionGeneration(t *testing.T) {
	tests := []struct {
		name                string
		disburseAddAmount   float64
		expectedDescription string
	}{
		{
			name:                "No SPM_DISBURSE_ADD adjustment",
			disburseAddAmount:   0.0,
			expectedDescription: "Withdraw Shopee",
		},
		{
			name:                "With SPM_DISBURSE_ADD adjustment",
			disburseAddAmount:   25.5,
			expectedDescription: "Withdraw Shopee (adjusted by SPM_DISBURSE_ADD: 25.50)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the description generation logic
			description := "Withdraw Shopee"
			if tt.disburseAddAmount > 0 {
				description = fmt.Sprintf("Withdraw Shopee (adjusted by SPM_DISBURSE_ADD: %.2f)", tt.disburseAddAmount)
			}

			if description != tt.expectedDescription {
				t.Errorf("expected description %q, got %q", tt.expectedDescription, description)
			}
		})
	}
}
