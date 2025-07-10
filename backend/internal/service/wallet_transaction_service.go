package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
)

// WalletTransactionParams holds optional filters for listing wallet transactions.
type WalletTransactionParams struct {
	PageNo             int
	PageSize           int
	CreateTimeFrom     *int64
	CreateTimeTo       *int64
	WalletType         string
	TransactionType    string
	MoneyFlow          string
	TransactionTabType string
}

// WalletTransaction represents a single wallet transaction row.
type WalletTransaction struct {
	TransactionID      int64   `json:"transaction_id"`
	Status             string  `json:"status"`
	TransactionType    string  `json:"transaction_type"`
	Amount             float64 `json:"amount"`
	CurrentBalance     float64 `json:"current_balance"`
	CreateTime         int64   `json:"create_time"`
	OrderSN            string  `json:"order_sn,omitempty"`
	RefundSN           string  `json:"refund_sn,omitempty"`
	WithdrawalType     string  `json:"withdrawal_type,omitempty"`
	TransactionFee     float64 `json:"transaction_fee,omitempty"`
	Description        string  `json:"description,omitempty"`
	BuyerName          string  `json:"buyer_name,omitempty"`
	WithdrawalID       int64   `json:"withdrawal_id,omitempty"`
	Reason             string  `json:"reason,omitempty"`
	RootWithdrawalID   int64   `json:"root_withdrawal_id,omitempty"`
	TransactionTabType string  `json:"transaction_tab_type,omitempty"`
	MoneyFlow          string  `json:"money_flow,omitempty"`
	Journaled          bool    `json:"journaled,omitempty"`
}

// WalletTransactionList contains transaction rows and pagination info.
type WalletTransactionList struct {
	Transactions []WalletTransaction
	more         bool
}

// WalletTransactionService fetches wallet transactions from Shopee API.
type WalletTransactionService struct {
	storeRepo *repository.ChannelRepo
	client    *ShopeeClient
}

func NewWalletTransactionService(r *repository.ChannelRepo, c *ShopeeClient) *WalletTransactionService {
	return &WalletTransactionService{storeRepo: r, client: c}
}

// ListWalletTransactions retrieves transactions for the given store with filters.
func (s *WalletTransactionService) ListWalletTransactions(ctx context.Context, store string, p WalletTransactionParams) ([]WalletTransaction, bool, error) {
	if s.storeRepo == nil || s.client == nil {
		return nil, false, fmt.Errorf("service not configured")
	}
	st, err := s.storeRepo.GetStoreByName(ctx, store)
	if err != nil || st == nil {
		return nil, false, fmt.Errorf("fetch store %s: %w", store, err)
	}
	if st.AccessToken == nil || st.ShopID == nil {
		return nil, false, fmt.Errorf("missing access token or shop id")
	}
	if err := ensureTokenValid(ctx, st, s.client, s.storeRepo); err != nil {
		return nil, false, err
	}
	resp, err := s.client.GetWalletTransactionList(ctx, *st.AccessToken, *st.ShopID, p)
	if err != nil && strings.Contains(err.Error(), "invalid_access_token") {
		if err2 := ensureTokenValid(ctx, st, s.client, s.storeRepo); err2 == nil {
			resp, err = s.client.GetWalletTransactionList(ctx, *st.AccessToken, *st.ShopID, p)
		}
	}
	if err != nil {
		return nil, false, err
	}
	log.Printf("Fetched %d wallet transactions for store %s", len(resp.Transactions), store)
	log.Printf("More transactions available: %t", resp.more)
	return resp.Transactions, resp.more, nil
}

// ensureTokenValid refreshes the store token if needed.
func ensureTokenValid(ctx context.Context, st *models.Store, c *ShopeeClient, repo *repository.ChannelRepo) error {
	if c == nil || repo == nil {
		return fmt.Errorf("missing client or store repo")
	}
	loc, _ := time.LoadLocation("Asia/Jakarta")
	reinterpreted := time.Date(
		st.LastUpdated.Year(), st.LastUpdated.Month(), st.LastUpdated.Day(),
		st.LastUpdated.Hour(), st.LastUpdated.Minute(), st.LastUpdated.Second(), st.LastUpdated.Nanosecond(),
		loc,
	)
	exp := reinterpreted.Add(time.Duration(*st.ExpireIn) * time.Second)
	if st.RefreshToken == nil {
		return fmt.Errorf("missing refresh token")
	}
	if st.ShopID == nil || *st.ShopID == "" {
		return fmt.Errorf("missing shop id")
	}
	if st.ExpireIn != nil && st.LastUpdated != nil {
		if time.Now().Before(exp.Local()) {
			return nil
		}
	}
	c.ShopID = *st.ShopID
	c.RefreshToken = *st.RefreshToken
	resp, err := c.RefreshAccessToken(ctx)
	if err != nil {
		return err
	}
	st.AccessToken = &resp.Response.AccessToken
	if resp.Response.RefreshToken != "" {
		st.RefreshToken = &resp.Response.RefreshToken
	}
	st.ExpireIn = &resp.Response.ExpireIn
	st.RequestID = &resp.Response.RequestID
	now := time.Now()
	st.LastUpdated = &now
	if err := repo.UpdateStore(ctx, st); err != nil {
		return err
	}
	return nil
}
