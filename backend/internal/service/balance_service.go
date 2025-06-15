// File: backend/internal/service/balance_service.go

package service

import (
	"context"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
)

// BalanceServiceJournalRepo defines the method needed from JournalRepo for balance sheet.
type BalanceServiceJournalRepo interface {
	GetAccountBalancesAsOf(ctx context.Context, shop string, asOfDate time.Time) ([]repository.AccountBalance, error)
}

// CategoryBalance groups account balances under a category (Assets, Liabilities, Equity).
type CategoryBalance struct {
	Category string                      `json:"category"` // e.g. "Assets"
	Accounts []repository.AccountBalance `json:"accounts"` // list of account balances in this category
	Total    float64                     `json:"total"`    // sum of balances
}

// BalanceService provides balance sheet data by category.
type BalanceService struct {
	journalRepo BalanceServiceJournalRepo
}

// NewBalanceService constructs a BalanceService with the given JournalRepo.
func NewBalanceService(jr BalanceServiceJournalRepo) *BalanceService {
	return &BalanceService{journalRepo: jr}
}

// GetBalanceSheet returns balances grouped into Assets, Liabilities, Equity as of asOfDate.
func (s *BalanceService) GetBalanceSheet(
	ctx context.Context,
	shop string,
	asOfDate time.Time,
) ([]CategoryBalance, error) {
	// Fetch raw account balances
	accBalances, err := s.journalRepo.GetAccountBalancesAsOf(ctx, shop, asOfDate)
	if err != nil {
		return nil, err
	}

	// Initialize category groups
	groups := map[string]*CategoryBalance{
		"Asset":     {Category: "Assets", Accounts: []repository.AccountBalance{}},
		"Liability": {Category: "Liabilities", Accounts: []repository.AccountBalance{}},
		"Equity":    {Category: "Equity", Accounts: []repository.AccountBalance{}},
	}

	// Assign each account balance to its category
	for _, ab := range accBalances {
		if grp, ok := groups[ab.AccountType]; ok {
			grp.Accounts = append(grp.Accounts, ab)
			grp.Total += ab.Balance
		}
	}

	// Return groups in a consistent order: Assets, Liabilities, Equity
	return []CategoryBalance{
		*groups["Asset"],
		*groups["Liability"],
		*groups["Equity"],
	}, nil
}
