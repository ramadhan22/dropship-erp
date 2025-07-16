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
	txs, next, err := s.walletSvc.ListWalletTransactions(ctx, store, p)
	if err != nil {
		return nil, false, err
	}
	if s.journalRepo != nil {
		for i := range txs {
			sid := fmt.Sprintf("%d", txs[i].TransactionID)
			if je, _ := s.journalRepo.GetJournalEntryBySource(ctx, "ads_topup", sid); je != nil {
				txs[i].Journaled = true
			}
		}
	}
	return txs, next, nil
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
		{JournalID: jid, AccountID: 55003, IsDebit: true, Amount: amt},
		{JournalID: jid, AccountID: saldoShopeeAccountID(store), IsDebit: false, Amount: amt},
	}
	for i := range lines {
		if err := s.journalRepo.InsertJournalLine(ctx, &lines[i]); err != nil {
			return err
		}
	}
	return nil
}

// CreateAllJournal fetches wallet transactions in 15-day windows going backwards
// in time and posts them to the journal until two consecutive windows return no
// transactions.
func (s *AdsTopupService) CreateAllJournal(ctx context.Context, store string) error {
	if s.walletSvc == nil {
		return fmt.Errorf("wallet service nil")
	}
	var emptyRanges int
	to := time.Now()
	for emptyRanges < 2 {
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
				TransactionType: "SPM_DEDUCT",
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
			emptyRanges++
		} else {
			emptyRanges = 0
		}
		to = from
	}
	return nil
}
