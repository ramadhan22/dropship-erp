package service

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
)

// PendingBalanceService exposes Shopee pending balance retrieval.
type PendingBalanceService struct {
	client    *ShopeeClient
	storeRepo *repository.ChannelRepo
}

// NewPendingBalanceService constructs a PendingBalanceService with the given client.
func NewPendingBalanceService(c *ShopeeClient, storeRepo *repository.ChannelRepo) *PendingBalanceService {
	return &PendingBalanceService{
		client:    c,
		storeRepo: storeRepo,
	}
}

// GetPendingBalance returns total pending balance for the store using the Shopee API.
func (s *PendingBalanceService) GetPendingBalance(ctx context.Context, store string) (float64, error) {
	if s.client == nil {
		return 0, nil
	}

	// If we have a store repository, validate the token proactively
	if s.storeRepo != nil {
		st, err := s.storeRepo.GetStoreByName(ctx, store)
		if err != nil || st == nil {
			log.Printf("Warning: could not fetch store %s for token validation: %v", store, err)
			// Continue without validation as fallback
		} else {
			// Ensure token is valid before making API call
			if err := s.ensureTokenValid(ctx, st); err != nil {
				return 0, fmt.Errorf("token validation failed for store %s: %w", store, err)
			}

			// Update client with fresh token
			if st.AccessToken != nil {
				s.client.AccessToken = *st.AccessToken
			}
			if st.ShopID != nil {
				s.client.ShopID = *st.ShopID
			}
		}
	}

	// Call the API with potentially refreshed token
	balance, err := s.client.GetPendingBalance(ctx, store)
	if err != nil && strings.Contains(err.Error(), "invalid_access_token") && s.storeRepo != nil {
		// Fallback: try to refresh token and retry if we get token error
		if st, stErr := s.storeRepo.GetStoreByName(ctx, store); stErr == nil && st != nil {
			if refreshErr := s.ensureTokenValid(ctx, st); refreshErr == nil {
				// Update client with fresh token
				if st.AccessToken != nil {
					s.client.AccessToken = *st.AccessToken
				}
				if st.ShopID != nil {
					s.client.ShopID = *st.ShopID
				}
				// Retry the API call
				balance, err = s.client.GetPendingBalance(ctx, store)
			}
		}
	}

	return balance, err
}

// ensureTokenValid reuses the same token validation logic from wallet_transaction_service.go
func (s *PendingBalanceService) ensureTokenValid(ctx context.Context, st *models.Store) error {
	if s.client == nil || s.storeRepo == nil {
		return fmt.Errorf("missing client or store repo")
	}

	// Use the same token validation logic as other services
	return ensureTokenValid(ctx, st, s.client, s.storeRepo)
}
