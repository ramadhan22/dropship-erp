// File: backend/internal/service/reconcile_service.go

package service

import (
	"context"
	"fmt"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// ReconcileRepoInterface defines just the methods needed from each repo.
type ReconcileServiceDropshipRepo interface {
	GetDropshipPurchaseByInvoice(ctx context.Context, kodeInvoice string) (*models.DropshipPurchase, error)
}
type ReconcileServiceShopeeRepo interface {
	// We only need to fetch the settled order.
	GetShopeeOrderByID(ctx context.Context, orderID string) (*models.ShopeeSettledOrder, error)
}
type ReconcileServiceJournalRepo interface {
	CreateJournalEntry(ctx context.Context, e *models.JournalEntry) (int64, error)
	InsertJournalLine(ctx context.Context, l *models.JournalLine) error
}
type ReconcileServiceRecRepo interface {
	InsertReconciledTransaction(ctx context.Context, r *models.ReconciledTransaction) error
}

// ReconcileService orchestrates matching Dropship + Shopee, creating journal entries + lines, and recording reconciliation.
type ReconcileService struct {
	dropRepo    ReconcileServiceDropshipRepo
	shopeeRepo  ReconcileServiceShopeeRepo
	journalRepo ReconcileServiceJournalRepo
	recRepo     ReconcileServiceRecRepo
}

// NewReconcileService constructs a ReconcileService.
func NewReconcileService(
	dr ReconcileServiceDropshipRepo,
	sr ReconcileServiceShopeeRepo,
	jr ReconcileServiceJournalRepo,
	rr ReconcileServiceRecRepo,
) *ReconcileService {
	return &ReconcileService{
		dropRepo:    dr,
		shopeeRepo:  sr,
		journalRepo: jr,
		recRepo:     rr,
	}
}

// MatchAndJournal does the following:
//  1. Ensure both DropshipPurchase and ShopeeSettledOrder exist,
//  2. Create a JournalEntry (header),
//  3. Insert two JournalLines (debit COGS, credit Cash),
//  4. Insert a ReconciledTransaction record.
func (s *ReconcileService) MatchAndJournal(
	ctx context.Context,
	purchaseID, orderID, shop string,
) error {
	// 1. Fetch DropshipPurchase
	dp, err := s.dropRepo.GetDropshipPurchaseByInvoice(ctx, purchaseID)
	if err != nil || dp == nil {
		return fmt.Errorf("fetch DropshipPurchase %s: %w", purchaseID, err)
	}

	// 2. Fetch ShopeeSettledOrder
	so, err := s.shopeeRepo.GetShopeeOrderByID(ctx, orderID)
	if err != nil || so == nil {
		return fmt.Errorf("fetch ShopeeOrder %s: %w", orderID, err)
	}

	// 3. Create JournalEntry
	je := &models.JournalEntry{
		EntryDate:    so.SettledDate,
		Description:  ptrString(fmt.Sprintf("Reconcile %s / %s", purchaseID, orderID)),
		SourceType:   "reconcile",
		SourceID:     orderID,
		ShopUsername: shop,
		Store:        shop,
		CreatedAt:    time.Now(),
	}
	journalID, err := s.journalRepo.CreateJournalEntry(ctx, je)
	if err != nil {
		return fmt.Errorf("create JournalEntry: %w", err)
	}

	// 4. Debit COGS (account_id=5001) and credit Cash (account_id=1001)
	//    Amounts: dp.TotalTransaksi debited, so.NetIncome credited
	jl1 := &models.JournalLine{
		JournalID: journalID,
		AccountID: 5001, // COGS
		IsDebit:   true,
		Amount:    dp.TotalTransaksi,
		Memo:      ptrString("COGS for " + purchaseID),
	}
	if err := s.journalRepo.InsertJournalLine(ctx, jl1); err != nil {
		return fmt.Errorf("insert JournalLine 1: %w", err)
	}
	jl2 := &models.JournalLine{
		JournalID: journalID,
		AccountID: 1001, // Cash
		IsDebit:   false,
		Amount:    so.NetIncome,
		Memo:      ptrString("Cash for " + orderID),
	}
	if err := s.journalRepo.InsertJournalLine(ctx, jl2); err != nil {
		return fmt.Errorf("insert JournalLine 2: %w", err)
	}

	// 5. Insert into reconciled_transactions
	rt := &models.ReconciledTransaction{
		ShopUsername: shop,
		DropshipID:   &dp.KodePesanan,
		ShopeeID:     &so.OrderID,
		Status:       "matched",
		MatchedAt:    time.Now(),
	}
	if err := s.recRepo.InsertReconciledTransaction(ctx, rt); err != nil {
		return fmt.Errorf("insert ReconciledTransaction: %w", err)
	}

	return nil
}

// ptrString helper
func ptrString(s string) *string { return &s }

// ListUnmatched delegates to repo to list unmatched rows.
func (s *ReconcileService) ListUnmatched(ctx context.Context, shop string) ([]models.ReconciledTransaction, error) {
	if repo, ok := s.recRepo.(interface {
		ListUnmatched(context.Context, string) ([]models.ReconciledTransaction, error)
	}); ok {
		return repo.ListUnmatched(ctx, shop)
	}
	return nil, fmt.Errorf("not implemented")
}

// ListCandidates proxies to the repo to fetch transactions that need attention.
func (s *ReconcileService) ListCandidates(ctx context.Context, shop string) ([]models.ReconcileCandidate, error) {
	if repo, ok := s.recRepo.(interface {
		ListCandidates(context.Context, string) ([]models.ReconcileCandidate, error)
	}); ok {
		return repo.ListCandidates(ctx, shop)
	}
	return nil, fmt.Errorf("not implemented")
}

// BulkReconcile simply loops MatchAndJournal over pairs.
func (s *ReconcileService) BulkReconcile(ctx context.Context, pairs [][2]string, shop string) error {
	for _, p := range pairs {
		if err := s.MatchAndJournal(ctx, p[0], p[1], shop); err != nil {
			return err
		}
	}
	return nil
}
