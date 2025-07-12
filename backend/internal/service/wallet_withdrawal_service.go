package service

import (
	"context"
	"fmt"
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

// CreateJournal posts a journal entry for the given transaction.
func (s *WalletWithdrawalService) CreateJournal(ctx context.Context, store string, t WalletTransaction) error {
	if s.journalRepo == nil {
		return fmt.Errorf("journal repo nil")
	}
	sid := fmt.Sprintf("%d", t.TransactionID)
	if je, _ := s.journalRepo.GetJournalEntryBySource(ctx, "wallet_withdrawal", sid); je != nil {
		return nil
	}
	amt := -t.Amount
	je := &models.JournalEntry{
		EntryDate:    time.Unix(t.CreateTime, 0),
		Description:  stringPtr("Withdraw Shopee"),
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
