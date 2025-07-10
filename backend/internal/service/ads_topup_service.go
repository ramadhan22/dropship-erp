package service

import (
	"context"
	"fmt"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
)

// AdsTopupService handles listing ads topup transactions and posting them to the journal.
type AdsTopupService struct {
	walletSvc   *WalletTransactionService
	journalRepo *repository.JournalRepo
}

func NewAdsTopupService(w *WalletTransactionService, jr *repository.JournalRepo) *AdsTopupService {
	return &AdsTopupService{walletSvc: w, journalRepo: jr}
}

func (s *AdsTopupService) List(ctx context.Context, store string, p WalletTransactionParams) ([]WalletTransaction, bool, error) {
	if s.walletSvc == nil {
		return nil, false, fmt.Errorf("wallet service nil")
	}
	p.TransactionType = "SPM_DEDUCT"
	if p.PageSize == 0 {
		p.PageSize = 25
	}
	return s.walletSvc.ListWalletTransactions(ctx, store, p)
}

func (s *AdsTopupService) CreateJournal(ctx context.Context, store string, t WalletTransaction) error {
	if s.journalRepo == nil {
		return fmt.Errorf("journal repo nil")
	}
	sid := fmt.Sprintf("%d", t.TransactionID)
	if je, _ := s.journalRepo.GetJournalEntryBySource(ctx, "ads_topup", sid); je != nil {
		return nil
	}
	amt := -t.Amount
	je := &models.JournalEntry{
		EntryDate:    time.Unix(t.CreateTime, 0),
		Description:  stringPtr("Shopee Ads Topup"),
		SourceType:   "ads_topup",
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
		{JournalID: jid, AccountID: adsSaldoShopeeAccountID(store), IsDebit: true, Amount: amt},
		{JournalID: jid, AccountID: saldoShopeeAccountID(store), IsDebit: false, Amount: amt},
	}
	for i := range lines {
		if err := s.journalRepo.InsertJournalLine(ctx, &lines[i]); err != nil {
			return err
		}
	}
	return nil
}

func stringPtr(s string) *string { return &s }
