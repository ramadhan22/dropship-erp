package service

import (
	"context"
	"testing"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// Test the new async CreateReconcileBatchesAsync method
func TestCreateReconcileBatchesAsync(t *testing.T) {
	// Create fake repositories
	drop := &fakeDropRepoBatch{data: make(map[string]*models.DropshipPurchase)}
	jrepo := &fakeJournalRepoBatch{}
	batchSvc := &fakeBatchSvc{}

	// Create service
	svc := NewReconcileService(nil, drop, nil, jrepo, nil, nil, nil, nil, nil, batchSvc, nil, nil, 5, nil)

	// Test async batch creation
	result, err := svc.CreateReconcileBatchesAsync(context.Background(), "test_shop", "test_order", "pending", "2024-01-01", "2024-01-31")

	// Verify result
	if err != nil {
		t.Fatalf("CreateReconcileBatchesAsync failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result but got nil")
	}

	if result.BatchCount != 1 {
		t.Errorf("Expected BatchCount=1, got %d", result.BatchCount)
	}

	if result.MasterBatchID == nil {
		t.Error("Expected MasterBatchID to be set")
	}

	// Verify that a batch was created in the fake service
	if len(batchSvc.created) != 1 {
		t.Errorf("Expected 1 batch to be created, got %d", len(batchSvc.created))
	}

	createdBatch := batchSvc.created[0]
	if createdBatch.ProcessType != "reconcile_batch_creation" {
		t.Errorf("Expected ProcessType='reconcile_batch_creation', got '%s'", createdBatch.ProcessType)
	}

	if createdBatch.Status != "pending" {
		t.Errorf("Expected Status='pending', got '%s'", createdBatch.Status)
	}

	// Verify metadata is stored in ErrorMessage
	expectedMetadata := "shop=test_shop,order=test_order,status=pending,from=2024-01-01,to=2024-01-31"
	if createdBatch.ErrorMessage != expectedMetadata {
		t.Errorf("Expected ErrorMessage='%s', got '%s'", expectedMetadata, createdBatch.ErrorMessage)
	}
}

// Test the ProcessReconcileBatchCreation method
func TestProcessReconcileBatchCreation(t *testing.T) {
	// Create fake repositories
	drop := &fakeDropRepoBatch{data: make(map[string]*models.DropshipPurchase)}
	jrepo := &fakeJournalRepoBatch{}
	batchSvc := &fakeBatchSvc{}

	// Create service
	svc := NewReconcileService(nil, drop, nil, jrepo, nil, nil, nil, nil, nil, batchSvc, nil, nil, 5, nil)

	// Test processing a master batch
	masterBatchID := int64(1)
	svc.ProcessReconcileBatchCreation(context.Background(), masterBatchID)

	// Verify that GetByID was called (implicitly through the fake returning the expected data)
	// The fake GetByID returns a batch with the expected metadata format
	// Since this is a unit test with mocks, we're mainly testing that the flow doesn't crash
	// and the methods are called in the right order
}

// Test the parseMetadata helper function
func TestParseMetadata(t *testing.T) {
	tests := []struct {
		input    string
		expected map[string]string
	}{
		{
			input: "shop=test,order=ABC,status=pending,from=2024-01-01,to=2024-01-31",
			expected: map[string]string{
				"shop":   "test",
				"order":  "ABC",
				"status": "pending",
				"from":   "2024-01-01",
				"to":     "2024-01-31",
			},
		},
		{
			input: "key1=value1,key2=value2",
			expected: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			input:    "",
			expected: map[string]string{},
		},
		{
			input:    "invalid_format",
			expected: map[string]string{},
		},
	}

	for _, test := range tests {
		result := parseMetadata(test.input)

		if len(result) != len(test.expected) {
			t.Errorf("For input '%s', expected %d keys, got %d", test.input, len(test.expected), len(result))
			continue
		}

		for key, expectedValue := range test.expected {
			if actualValue, exists := result[key]; !exists {
				t.Errorf("For input '%s', expected key '%s' to exist", test.input, key)
			} else if actualValue != expectedValue {
				t.Errorf("For input '%s', expected key '%s' to have value '%s', got '%s'", test.input, key, expectedValue, actualValue)
			}
		}
	}
}
