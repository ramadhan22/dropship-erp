package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
)

// WalletWithdrawalService handles listing withdrawal transactions from
// Shopee wallet and posting them to the journal.
type WalletWithdrawalService struct {
	walletSvc   *WalletTransactionService
	journalRepo *repository.JournalRepo
}

// NewWalletWithdrawalService creates a new WalletWithdrawalService.
func NewWalletWithdrawalService(w *WalletTransactionService, jr *repository.JournalRepo) *WalletWithdrawalService {
	return &WalletWithdrawalService{walletSvc: w, journalRepo: jr}
}

// List returns withdrawal transactions for the given store.
func (s *WalletWithdrawalService) List(ctx context.Context, store string, p WalletTransactionParams) ([]WalletTransaction, bool, error) {
	if s.walletSvc == nil {
		return nil, false, fmt.Errorf("wallet service nil")
	}
	p.TransactionType = "WITHDRAWAL_CREATED"
	if p.PageSize == 0 {
		p.PageSize = 25
	}
	txs, next, err := s.walletSvc.ListWalletTransactions(ctx, store, p)
	if err != nil {
		return nil, false, err
	}
	
	// Adjust withdrawal amounts by SPM_DISBURSE_ADD transactions on the same day
	for i := range txs {
		disburseAddAmount, err := s.findSpmDisburseAddAmount(ctx, store, txs[i].CreateTime)
		if err != nil {
			// Log the error but don't fail the entire request
			log.Printf("Warning: failed to check SPM_DISBURSE_ADD for transaction %d: %v", txs[i].TransactionID, err)
		} else if disburseAddAmount > 0 {
			txs[i].Amount = txs[i].Amount - disburseAddAmount
		}
	}
	
	if s.journalRepo != nil {
		for i := range txs {
			sid := fmt.Sprintf("%d", txs[i].TransactionID)
			if je, _ := s.journalRepo.GetJournalEntryBySource(ctx, "wallet_withdrawal", sid); je != nil {
				txs[i].Journaled = true
			}
		}
	}
	return txs, next, nil
}

// WalletTransactionLister interface for testing
type WalletTransactionLister interface {
	ListWalletTransactions(ctx context.Context, store string, p WalletTransactionParams) ([]WalletTransaction, bool, error)
}

// findSpmDisburseAddAmount finds SPM_DISBURSE_ADD transactions on the same day as the withdrawal
// and returns the total amount to be deducted from the withdrawal.
func (s *WalletWithdrawalService) findSpmDisburseAddAmount(ctx context.Context, store string, withdrawalTime int64) (float64, error) {
	return findSpmDisburseAddAmountWithService(ctx, s.walletSvc, store, withdrawalTime)
}

// findSpmDisburseAddAmountWithService is the testable version that accepts an interface
func findSpmDisburseAddAmountWithService(ctx context.Context, walletSvc WalletTransactionLister, store string, withdrawalTime int64) (float64, error) {
	if walletSvc == nil {
		return 0, fmt.Errorf("wallet service nil")
	}

	// Convert withdrawal time to start and end of day
	withdrawalDate := time.Unix(withdrawalTime, 0)
	startOfDay := time.Date(withdrawalDate.Year(), withdrawalDate.Month(), withdrawalDate.Day(), 0, 0, 0, 0, withdrawalDate.Location())
	endOfDay := time.Date(withdrawalDate.Year(), withdrawalDate.Month(), withdrawalDate.Day(), 23, 59, 59, 999999999, withdrawalDate.Location())

	fromUnix := startOfDay.Unix()
	toUnix := endOfDay.Unix()

	var totalDisburseAdd float64
	page := 0

	for {
		params := WalletTransactionParams{
			PageNo:          page,
			PageSize:        50,
			CreateTimeFrom:  &fromUnix,
			CreateTimeTo:    &toUnix,
			TransactionType: "SPM_DISBURSE_ADD",
		}

		txs, more, err := walletSvc.ListWalletTransactions(ctx, store, params)
		if err != nil {
			return 0, fmt.Errorf("failed to fetch SPM_DISBURSE_ADD transactions: %w", err)
		}

		for _, tx := range txs {
			totalDisburseAdd += tx.Amount
		}

		if !more {
			break
		}
		page++
	}

	return totalDisburseAdd, nil
}

// CreateJournal posts a journal entry for the given transaction.
func (s *WalletWithdrawalService) CreateJournal(ctx context.Context, store string, t WalletTransaction) error {
	if s.journalRepo == nil {
		return fmt.Errorf("journal repo nil")
	}
	sid := fmt.Sprintf("%d", t.TransactionID)
	if je, _ := s.journalRepo.GetJournalEntryBySource(ctx, "wallet_withdrawal", sid); je != nil {
		return nil
	}

	// Check for SPM_DISBURSE_ADD transactions on the same day
	disburseAddAmount, err := s.findSpmDisburseAddAmount(ctx, store, t.CreateTime)
	if err != nil {
		return fmt.Errorf("failed to check SPM_DISBURSE_ADD transactions: %w", err)
	}

	// Adjust withdrawal amount by deducting SPM_DISBURSE_ADD amount
	adjustedAmount := t.Amount - disburseAddAmount
	amt := -adjustedAmount

	// Create description that reflects whether amount was adjusted
	description := "Withdraw Shopee"
	if disburseAddAmount > 0 {
		description = fmt.Sprintf("Withdraw Shopee (adjusted by SPM_DISBURSE_ADD: %.2f)", disburseAddAmount)
	}

	je := &models.JournalEntry{
		EntryDate:    time.Unix(t.CreateTime, 0),
		Description:  stringPtr(description),
		SourceType:   "wallet_withdrawal",
		SourceID:     sid,
		ShopUsername: store,
		Store:        store,
		CreatedAt:    time.Now(),
	}
	jid, err := s.journalRepo.CreateJournalEntry(ctx, je)
	if err != nil {
		return err
	}
	lines := []models.JournalLine{
		{JournalID: jid, AccountID: 11014, IsDebit: true, Amount: amt},
		{JournalID: jid, AccountID: saldoShopeeAccountID(store), IsDebit: false, Amount: amt},
	}
	for i := range lines {
		if err := s.journalRepo.InsertJournalLine(ctx, &lines[i]); err != nil {
			return err
		}
	}
	return nil
}

// CreateAllJournal fetches withdrawal transactions backwards in time and posts
// them until two consecutive windows return no transactions.
func (s *WalletWithdrawalService) CreateAllJournal(ctx context.Context, store string) error {
	if s.walletSvc == nil {
		return fmt.Errorf("wallet service nil")
	}
	var empty int
	to := time.Now()
	for empty < 2 {
		from := to.AddDate(0, 0, -14)
		fromUnix := from.Unix()
		toUnix := to.Unix()
		page := 0
		total := 0
		for {
			params := WalletTransactionParams{
				PageNo:          page,
				PageSize:        50,
				CreateTimeFrom:  &fromUnix,
				CreateTimeTo:    &toUnix,
				TransactionType: "WITHDRAWAL_CREATED",
			}
			txs, more, err := s.walletSvc.ListWalletTransactions(ctx, store, params)
			if err != nil {
				return err
			}
			for _, tx := range txs {
				if err := s.CreateJournal(ctx, store, tx); err != nil {
					return err
				}
			}
			total += len(txs)
			if !more {
				break
			}
			page++
		}
		if total == 0 {
			empty++
		} else {
			empty = 0
		}
		to = from
	}
	return nil
}
