// File: backend/internal/service/reconcile_service.go

package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
)

// ReconcileRepoInterface defines just the methods needed from each repo.
type ReconcileServiceDropshipRepo interface {
	GetDropshipPurchaseByInvoice(ctx context.Context, kodeInvoice string) (*models.DropshipPurchase, error)
	GetDropshipPurchaseByID(ctx context.Context, kodePesanan string) (*models.DropshipPurchase, error)
	UpdatePurchaseStatus(ctx context.Context, kodePesanan, status string) error
	SumDetailByInvoice(ctx context.Context, kodeInvoice string) (float64, error)
	SumProductCostByInvoice(ctx context.Context, kodeInvoice string) (float64, error)
}
type ReconcileServiceShopeeRepo interface {
	// We only need to fetch the settled order.
	GetShopeeOrderByID(ctx context.Context, orderID string) (*models.ShopeeSettledOrder, error)
	ExistsShopeeSettled(ctx context.Context, noPesanan string) (bool, error)
}
type ReconcileServiceJournalRepo interface {
	CreateJournalEntry(ctx context.Context, e *models.JournalEntry) (int64, error)
	InsertJournalLine(ctx context.Context, l *models.JournalLine) error
}
type ReconcileServiceRecRepo interface {
	InsertReconciledTransaction(ctx context.Context, r *models.ReconciledTransaction) error
}
type ReconcileServiceStoreRepo interface {
	GetStoreByName(ctx context.Context, name string) (*models.Store, error)
}

// ReconcileService orchestrates matching Dropship + Shopee, creating journal entries + lines, and recording reconciliation.
type ReconcileService struct {
	db          *sqlx.DB
	dropRepo    ReconcileServiceDropshipRepo
	shopeeRepo  ReconcileServiceShopeeRepo
	journalRepo ReconcileServiceJournalRepo
	recRepo     ReconcileServiceRecRepo
	storeRepo   ReconcileServiceStoreRepo
	client      *ShopeeClient
}

// NewReconcileService constructs a ReconcileService.
func NewReconcileService(
	db *sqlx.DB,
	dr ReconcileServiceDropshipRepo,
	sr ReconcileServiceShopeeRepo,
	jr ReconcileServiceJournalRepo,
	rr ReconcileServiceRecRepo,
	srp ReconcileServiceStoreRepo,
	c *ShopeeClient,
) *ReconcileService {
	return &ReconcileService{
		db:          db,
		dropRepo:    dr,
		shopeeRepo:  sr,
		journalRepo: jr,
		recRepo:     rr,
		storeRepo:   srp,
		client:      c,
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
	log.Printf("Reconciling purchase %s with order %s for shop %s", purchaseID, orderID, shop)
	var tx *sqlx.Tx
	dropRepo := s.dropRepo
	jrRepo := s.journalRepo
	recRepo := s.recRepo
	if s.db != nil {
		var err error
		tx, err = s.db.BeginTxx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		dropRepo = repository.NewDropshipRepo(tx)
		jrRepo = repository.NewJournalRepo(tx)
		recRepo = repository.NewReconcileRepo(tx)
	}

	// 1. Fetch DropshipPurchase
	dp, err := dropRepo.GetDropshipPurchaseByInvoice(ctx, purchaseID)
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
	journalID, err := jrRepo.CreateJournalEntry(ctx, je)
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
	if err := jrRepo.InsertJournalLine(ctx, jl1); err != nil {
		return fmt.Errorf("insert JournalLine 1: %w", err)
	}
	jl2 := &models.JournalLine{
		JournalID: journalID,
		AccountID: 1001, // Cash
		IsDebit:   false,
		Amount:    so.NetIncome,
		Memo:      ptrString("Cash for " + orderID),
	}
	if err := jrRepo.InsertJournalLine(ctx, jl2); err != nil {
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
	if err := recRepo.InsertReconciledTransaction(ctx, rt); err != nil {
		return fmt.Errorf("insert ReconciledTransaction: %w", err)
	}
	if tx != nil {
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	log.Printf("ReconcileService completed purchase %s order %s", purchaseID, orderID)
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
func (s *ReconcileService) ListCandidates(ctx context.Context, shop, order string) ([]models.ReconcileCandidate, error) {
	if repo, ok := s.recRepo.(interface {
		ListCandidates(context.Context, string, string) ([]models.ReconcileCandidate, error)
	}); ok {
		return repo.ListCandidates(ctx, shop, order)
	}
	return nil, fmt.Errorf("not implemented")
}

// BulkReconcile simply loops MatchAndJournal over pairs.
func (s *ReconcileService) BulkReconcile(ctx context.Context, pairs [][2]string, shop string) error {
	log.Printf("BulkReconcile %d pairs for shop %s", len(pairs), shop)
	for _, p := range pairs {
		if err := s.MatchAndJournal(ctx, p[0], p[1], shop); err != nil {
			return err
		}
	}
	return nil
}

// CheckAndMarkComplete verifies a purchase has a corresponding shopee_settled
// entry and updates its status to "pesanan selesai" if found.
func (s *ReconcileService) CheckAndMarkComplete(ctx context.Context, kodePesanan string) error {
	log.Printf("CheckAndMarkComplete: %s", kodePesanan)
	var tx *sqlx.Tx
	dropRepo := s.dropRepo
	if s.db != nil {
		var err error
		tx, err = s.db.BeginTxx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		dropRepo = repository.NewDropshipRepo(tx)
	}

	dp, err := dropRepo.GetDropshipPurchaseByID(ctx, kodePesanan)
	if err != nil || dp == nil {
		return fmt.Errorf("fetch DropshipPurchase %s: %w", kodePesanan, err)
	}
	exists, err := s.shopeeRepo.ExistsShopeeSettled(ctx, dp.KodeInvoiceChannel)
	if err != nil {
		return fmt.Errorf("check shopee settled: %w", err)
	}
	if !exists {
		return fmt.Errorf("shopee settled order not found")
	}
	if err := dropRepo.UpdatePurchaseStatus(ctx, kodePesanan, "Pesanan selesai"); err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	if tx != nil {
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	log.Printf("CheckAndMarkComplete done: %s", kodePesanan)
	return nil
}

// CancelPurchase reverses pending sales journals for the given purchase except
// for the Biaya Mitra amount which remains recorded.
func (s *ReconcileService) CancelPurchase(ctx context.Context, kodePesanan string) error {
	log.Printf("CancelPurchase started: %s", kodePesanan)
	var tx *sqlx.Tx
	dropRepo := s.dropRepo
	jrRepo := s.journalRepo
	if s.db != nil {
		var err error
		tx, err = s.db.BeginTxx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		dropRepo = repository.NewDropshipRepo(tx)
		jrRepo = repository.NewJournalRepo(tx)
	}

	dp, err := dropRepo.GetDropshipPurchaseByID(ctx, kodePesanan)
	if err != nil || dp == nil {
		return fmt.Errorf("fetch DropshipPurchase %s: %w", kodePesanan, err)
	}

	prodCh, _ := dropRepo.SumDetailByInvoice(ctx, dp.KodeInvoiceChannel)
	prod, _ := dropRepo.SumProductCostByInvoice(ctx, dp.KodeInvoiceChannel)

	je := &models.JournalEntry{
		EntryDate:    time.Now(),
		Description:  ptrString("Cancel " + dp.KodeInvoiceChannel),
		SourceType:   "reconcile_cancel",
		SourceID:     dp.KodeInvoiceChannel,
		ShopUsername: dp.NamaToko,
		Store:        dp.NamaToko,
		CreatedAt:    time.Now(),
	}
	jid, err := jrRepo.CreateJournalEntry(ctx, je)
	if err != nil {
		return err
	}

	lines := []models.JournalLine{
		{JournalID: jid, AccountID: 11009, IsDebit: true, Amount: prod, Memo: ptrString("Saldo Jakmall " + dp.KodeInvoiceChannel)},
		{JournalID: jid, AccountID: 5001, IsDebit: false, Amount: prod, Memo: ptrString("HPP " + dp.KodeInvoiceChannel)},
		{JournalID: jid, AccountID: pendingAccountID(dp.NamaToko), IsDebit: false, Amount: prodCh, Memo: ptrString("Pending receivable " + dp.KodeInvoiceChannel)},
		{JournalID: jid, AccountID: 4001, IsDebit: true, Amount: prodCh, Memo: ptrString("Sales " + dp.KodeInvoiceChannel)},
	}
	for i := range lines {
		if lines[i].Amount == 0 {
			continue
		}
		if err := jrRepo.InsertJournalLine(ctx, &lines[i]); err != nil {
			return err
		}
	}

	if err := dropRepo.UpdatePurchaseStatus(ctx, kodePesanan, "Pesanan dibatalkan"); err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	if tx != nil {
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	log.Printf("CancelPurchase completed: %s", kodePesanan)
	return nil
}

// GetShopeeOrderStatus uses the Shopee API client to fetch current order status.
func (s *ReconcileService) GetShopeeOrderStatus(ctx context.Context, invoice string) (string, error) {
	if s.client == nil {
		return "", fmt.Errorf("shopee client not configured")
	}
	return s.client.GetOrderDetail(ctx, invoice)
}

// GetShopeeOrderDetail retrieves order detail using the store's saved access token.
func (s *ReconcileService) GetShopeeOrderDetail(ctx context.Context, invoice string) (*ShopeeOrderDetail, error) {
	if s.dropRepo == nil || s.storeRepo == nil {
		return nil, fmt.Errorf("repos not configured")
	}
	dp, err := s.dropRepo.GetDropshipPurchaseByInvoice(ctx, invoice)
	if err != nil || dp == nil {
		return nil, fmt.Errorf("fetch purchase %s: %w", invoice, err)
	}
	st, err := s.storeRepo.GetStoreByName(ctx, dp.NamaToko)
	if err != nil || st == nil {
		return nil, fmt.Errorf("fetch store %s: %w", dp.NamaToko, err)
	}
	if st.AccessToken == nil {
		return nil, fmt.Errorf("missing access token")
	}
	return FetchShopeeOrderDetail(ctx, *st.AccessToken, dp.KodeInvoiceChannel)
}

// GetShopeeAccessToken obtains an access token for the store related to the given invoice.
func (s *ReconcileService) GetShopeeAccessToken(ctx context.Context, invoice string) (string, error) {
	if s.client == nil {
		return "", fmt.Errorf("shopee client not configured")
	}
	dp, err := s.dropRepo.GetDropshipPurchaseByInvoice(ctx, invoice)
	if err != nil || dp == nil {
		return "", fmt.Errorf("fetch purchase %s: %w", invoice, err)
	}
	if s.storeRepo == nil {
		return "", fmt.Errorf("store repo not configured")
	}
	st, err := s.storeRepo.GetStoreByName(ctx, dp.NamaToko)
	if err != nil || st == nil {
		return "", fmt.Errorf("fetch store %s: %w", dp.NamaToko, err)
	}
	if st.CodeID == nil || st.ShopID == nil {
		return "", fmt.Errorf("store missing code or shop id")
	}
	tok, err := s.client.GetAccessToken(ctx, *st.CodeID, *st.ShopID)
	if err != nil {
		return "", err
	}
	return tok.AccessToken, nil
}
