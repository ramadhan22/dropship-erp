package service

import (
	"context"
)

// PendingBalanceService exposes Shopee pending balance retrieval.
type PendingBalanceService struct {
	client *ShopeeClient
}

// NewPendingBalanceService constructs a PendingBalanceService with the given client.
func NewPendingBalanceService(c *ShopeeClient) *PendingBalanceService {
	return &PendingBalanceService{client: c}
}

// GetPendingBalance returns total pending balance for the store using the Shopee API.
func (s *PendingBalanceService) GetPendingBalance(ctx context.Context, store string) (float64, error) {
	if s.client == nil {
		return 0, nil
	}
	return s.client.GetPendingBalance(ctx, store)
}
