// File: backend/internal/service/reconcile_service.go

package service

import (
	"context"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/jmoiron/sqlx"
	"github.com/ramadhan22/dropship-erp/backend/internal/logutil"
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
	UpdateStore(ctx context.Context, s *models.Store) error
}
type ReconcileServiceDetailRepo interface {
	SaveOrderDetail(ctx context.Context, detail *models.ShopeeOrderDetailRow, items []models.ShopeeOrderItemRow, packages []models.ShopeeOrderPackageRow) error
	UpdateOrderDetailStatus(ctx context.Context, sn, status, orderStatus string, updateTime time.Time) error
	GetOrderDetail(ctx context.Context, sn string) (*models.ShopeeOrderDetailRow, []models.ShopeeOrderItemRow, []models.ShopeeOrderPackageRow, error)
}

type ReconcileServiceBatchSvc interface {
	Create(ctx context.Context, b *models.BatchHistory) (int64, error)
	UpdateDone(ctx context.Context, id int64, done int) error
	UpdateStatus(ctx context.Context, id int64, status, msg string) error
}

// ReconcileService orchestrates matching Dropship + Shopee, creating journal entries + lines, and recording reconciliation.
type ReconcileService struct {
	db          *sqlx.DB
	dropRepo    ReconcileServiceDropshipRepo
	shopeeRepo  ReconcileServiceShopeeRepo
	journalRepo ReconcileServiceJournalRepo
	adjRepo     *repository.ShopeeAdjustmentRepo
	recRepo     ReconcileServiceRecRepo
	storeRepo   ReconcileServiceStoreRepo
	detailRepo  ReconcileServiceDetailRepo
	client      *ShopeeClient
	batchSvc    ReconcileServiceBatchSvc
	maxThreads  int
}

// NewReconcileService constructs a ReconcileService.
func NewReconcileService(
	db *sqlx.DB,
	dr ReconcileServiceDropshipRepo,
	sr ReconcileServiceShopeeRepo,
	jr ReconcileServiceJournalRepo,
	rr ReconcileServiceRecRepo,
	srp ReconcileServiceStoreRepo,
	drp ReconcileServiceDetailRepo,
	ar *repository.ShopeeAdjustmentRepo,
	c *ShopeeClient,
	b ReconcileServiceBatchSvc,
	maxThreads int,
) *ReconcileService {
	return &ReconcileService{
		db:          db,
		dropRepo:    dr,
		shopeeRepo:  sr,
		journalRepo: jr,
		adjRepo:     ar,
		recRepo:     rr,
		storeRepo:   srp,
		detailRepo:  drp,
		client:      c,
		batchSvc:    b,
		maxThreads:  maxThreads,
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

func (s *ReconcileService) createAdjustmentJournal(ctx context.Context, jr ReconcileServiceJournalRepo, a *models.ShopeeAdjustment) error {
	je := &models.JournalEntry{
		EntryDate:    a.TanggalPenyesuaian,
		Description:  ptrString("Shopee adjustment " + a.NoPesanan),
		SourceType:   "shopee_adjustment",
		SourceID:     fmt.Sprintf("%s-%s-%s", a.NoPesanan, a.TanggalPenyesuaian.Format("20060102"), sanitizeID(a.TipePenyesuaian)),
		ShopUsername: a.NamaToko,
		Store:        a.NamaToko,
		CreatedAt:    time.Now(),
	}
	jid, err := jr.CreateJournalEntry(ctx, je)
	if err != nil {
		return err
	}
	amt := a.BiayaPenyesuaian
	saldoAcc := saldoShopeeAccountID(a.NamaToko)
	if amt >= 0 {
		lines := []models.JournalLine{
			{JournalID: jid, AccountID: saldoAcc, IsDebit: true, Amount: amt},
			{JournalID: jid, AccountID: 4001, IsDebit: false, Amount: amt},
		}
		for i := range lines {
			if err := jr.InsertJournalLine(ctx, &lines[i]); err != nil {
				return err
			}
		}
	} else {
		aamt := -amt
		acc := int64(55005)
		if strings.EqualFold(a.TipePenyesuaian, "Shipping Fee Discrepancy") {
			acc = 52010
		}
		lines := []models.JournalLine{
			{JournalID: jid, AccountID: acc, IsDebit: true, Amount: aamt},
			{JournalID: jid, AccountID: saldoAcc, IsDebit: false, Amount: aamt},
		}
		for i := range lines {
			if err := jr.InsertJournalLine(ctx, &lines[i]); err != nil {
				return err
			}
		}
	}
	return nil
}

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
func (s *ReconcileService) ListCandidates(ctx context.Context, shop, order, from, to string, limit, offset int) ([]models.ReconcileCandidate, int, error) {
	if repo, ok := s.recRepo.(interface {
		ListCandidates(context.Context, string, string, string, string, int, int) ([]models.ReconcileCandidate, int, error)
	}); ok {
		list, total, err := repo.ListCandidates(ctx, shop, order, from, to, limit, offset)
		if err != nil {
			return nil, 0, err
		}
		for i := range list {
			log.Printf("Fetching Shopee order detail for %s", list[i].KodeInvoiceChannel)

			var statusStr string
			if s.detailRepo != nil {
				if row, _, _, err := s.detailRepo.GetOrderDetail(ctx, list[i].KodeInvoiceChannel); err == nil && row != nil {
					if row.OrderStatus != nil {
						statusStr = *row.OrderStatus
					} else if row.Status != nil {
						statusStr = *row.Status
					}
				}
			}

			if statusStr == "" {
				detail, err := s.GetShopeeOrderDetail(ctx, list[i].KodeInvoiceChannel)
				if err != nil {
					logutil.Errorf("GetShopeeOrderDetail %s: %v", list[i].KodeInvoiceChannel, err)
					list[i].ShopeeOrderStatus = "Not Found"
					continue
				}
				status := (*detail)["order_status"]
				if status == nil {
					status = (*detail)["status"]
				}
				if str, ok := status.(string); ok {
					statusStr = str
				}
			}

			if statusStr == "" {
				statusStr = "Not Found"
			}
			list[i].ShopeeOrderStatus = statusStr
		}
		return list, total, nil
	}
	return nil, 0, fmt.Errorf("not implemented")
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
	return s.CancelPurchaseAt(ctx, kodePesanan, time.Now(), "Pesanan dibatalkan")
}

// CancelPurchaseAt performs the cancel journal reversal at the given entry date
// and updates the purchase status to the provided status string.
func (s *ReconcileService) CancelPurchaseAt(ctx context.Context, kodePesanan string, entryDate time.Time, status string) error {
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
		EntryDate:    entryDate,
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

	if err := dropRepo.UpdatePurchaseStatus(ctx, kodePesanan, status); err != nil {
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
	if st.AccessToken == nil || st.ShopID == nil {
		return nil, fmt.Errorf("missing access token or shop id")
	}
	if s.client == nil {
		return nil, fmt.Errorf("shopee client not configured")
	}

	if err := s.ensureStoreTokenValid(ctx, st); err != nil {
		return nil, err
	}

	detail, err := s.client.FetchShopeeOrderDetail(ctx, *st.AccessToken, *st.ShopID, dp.KodeInvoiceChannel)
	if err != nil && strings.Contains(err.Error(), "invalid_access_token") {
		if err := s.ensureStoreTokenValid(ctx, st); err != nil {
			return nil, err
		}
		detail, err = s.client.FetchShopeeOrderDetail(ctx, *st.AccessToken, *st.ShopID, dp.KodeInvoiceChannel)
	}
	if err == nil && s.detailRepo != nil {
		row, items, packages := normalizeOrderDetail(invoice, dp.NamaToko, *detail)
		if err := s.detailRepo.SaveOrderDetail(ctx, row, items, packages); err != nil {
			log.Printf("save order detail %s: %v", invoice, err)
		}
	}
	return detail, err
}

// GetShopeeEscrowDetail returns escrow information for the given invoice using
// the store's saved access token. It refreshes the token when needed.
func (s *ReconcileService) GetShopeeEscrowDetail(ctx context.Context, invoice string) (*ShopeeEscrowDetail, error) {
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
	if st.AccessToken == nil || st.ShopID == nil {
		return nil, fmt.Errorf("missing access token or shop id")
	}
	if s.client == nil {
		return nil, fmt.Errorf("shopee client not configured")
	}

	if err := s.ensureStoreTokenValid(ctx, st); err != nil {
		return nil, err
	}

	detail, err := s.client.GetEscrowDetail(ctx, *st.AccessToken, *st.ShopID, dp.KodeInvoiceChannel)
	if err != nil && strings.Contains(err.Error(), "invalid_access_token") {
		if err := s.ensureStoreTokenValid(ctx, st); err != nil {
			return nil, err
		}
		detail, err = s.client.GetEscrowDetail(ctx, *st.AccessToken, *st.ShopID, dp.KodeInvoiceChannel)
	}
	log.Printf("GetShopeeEscrowDetail: detail: %+v, err: %v", detail, err)
	return detail, err
}

func (s *ReconcileService) ensureStoreTokenValid(ctx context.Context, st *models.Store) error {
	log.Printf("TEST: ensureStoreTokenValid for store %d", st.StoreID)
	// New location (e.g., Asia/Jakarta)
	loc, _ := time.LoadLocation("Asia/Jakarta")

	// Reinterpret the time as if it were in the new location
	reinterpreted := time.Date(
		st.LastUpdated.Year(), st.LastUpdated.Month(), st.LastUpdated.Day(),
		st.LastUpdated.Hour(), st.LastUpdated.Minute(), st.LastUpdated.Second(), st.LastUpdated.Nanosecond(),
		loc,
	)
	exp := reinterpreted.Add(time.Duration(*st.ExpireIn) * time.Second)
	if st.RefreshToken == nil {
		log.Fatalf("ensureStoreTokenValid: missing refresh token for store %d", st.StoreID)
		return fmt.Errorf("missing refresh token")
	}
	if st.ShopID == nil || *st.ShopID == "" {
		log.Fatalf("ensureStoreTokenValid: missing shop id for store %d", st.StoreID)
		return fmt.Errorf("missing shop id")
	}
	if st.ExpireIn != nil && st.LastUpdated != nil {
		if time.Now().Before(exp.Local()) {
			log.Printf("Token for store %d is still valid until %v and current time is %v", st.StoreID, exp, time.Now())
			log.Printf("current time: %v", time.Now())
			return nil
		}
	}
	log.Printf("exp: %v", exp)
	s.client.ShopID = *st.ShopID
	s.client.RefreshToken = *st.RefreshToken
	resp, err := s.client.RefreshAccessToken(ctx)
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
	if uerr := s.storeRepo.UpdateStore(ctx, st); uerr != nil {
		log.Printf("update store token: %v", uerr)
	}
	return nil
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

// UpdateShopeeStatus checks the current Shopee order status and performs
// cancellation logic when the status is cancelled.
func (s *ReconcileService) UpdateShopeeStatus(ctx context.Context, invoice string) error {
	detail, err := s.GetShopeeOrderDetail(ctx, invoice)
	if err != nil {
		return err
	}
	statusVal, ok := (*detail)["order_status"]
	if !ok {
		statusVal = (*detail)["status"]
	}
	statusStr, _ := statusVal.(string)
	status := strings.ToLower(statusStr)
	var updateTime time.Time
	if ts, ok := (*detail)["update_time"]; ok {
		switch v := ts.(type) {
		case float64:
			updateTime = time.Unix(int64(v), 0)
		case int64:
			updateTime = time.Unix(v, 0)
		case int:
			updateTime = time.Unix(int64(v), 0)
		case string:
			if t, err := strconv.ParseInt(v, 10, 64); err == nil {
				updateTime = time.Unix(t, 0)
			}
		}
	}
	if updateTime.IsZero() {
		updateTime = time.Now()
	}
	if status == "completed" {
		if err := s.createEscrowSettlementJournal(ctx, invoice, statusStr, updateTime, nil); err != nil {
			return err
		}
		return nil
	}
	if status != "cancelled" {
		return nil
	}
	dp, err := s.dropRepo.GetDropshipPurchaseByInvoice(ctx, invoice)
	if err != nil || dp == nil {
		return fmt.Errorf("fetch purchase %s: %w", invoice, err)
	}
	return s.CancelPurchaseAt(ctx, dp.KodePesanan, updateTime, "Cancelled Shopee")
}

// createEscrowSettlementJournal posts journal entries based on escrow detail and
// marks the purchase as complete.
func (s *ReconcileService) createEscrowSettlementJournal(ctx context.Context, invoice, status string, updateTime time.Time, escDetail *ShopeeEscrowDetail) error {
	log.Printf("createEscrowSettlementJournal %s", invoice)

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

	log.Printf("Fetching DropshipPurchase for invoice %s", invoice)
	dp, err := dropRepo.GetDropshipPurchaseByInvoice(ctx, invoice)
	if err != nil || dp == nil {
		return fmt.Errorf("fetch purchase %s: %w", invoice, err)
	}

	if escDetail == nil {
		escDetail, err = s.GetShopeeEscrowDetail(ctx, invoice)
		if err != nil {
			return err
		}
	}
	m := map[string]any(*escDetail)
	income, _ := m["order_income"].(map[string]any)
	orderPrice := 0.0
	if orderPrice == 0 {
		if v := asFloat64(income, "order_original_price"); v != nil {
			orderPrice = *v
		}
	}
	commission := 0.0
	if v := asFloat64(income, "commission_fee"); v != nil {
		commission = *v
	}
	service := 0.0
	if v := asFloat64(income, "service_fee"); v != nil {
		service = *v
	}
	voucher := 0.0
	if v := asFloat64(income, "voucher_from_seller"); v != nil {
		voucher = *v
	}
	if v := asFloat64(income, "seller_coin_cash_back"); v != nil {
		voucher += *v
	}
	discount := 0.0
	if v := asFloat64(income, "order_seller_discount"); v != nil {
		discount = *v
	}
	affiliate := 0.0
	if v := asFloat64(income, "order_ams_commission_fee"); v != nil {
		affiliate = *v
	}

	logistikAmt := 0.0
	if adjList, ok := income["order_adjustment"].([]any); ok {
		for _, a := range adjList {
			am, ok := a.(map[string]any)
			if !ok {
				continue
			}
			reason, _ := am["adjustment_reason"].(string)
			if strings.EqualFold(reason, "BD Marketing") {
				if v := asFloat64(am, "amount"); v != nil {
					affiliate += math.Abs(*v)
				}
			}
			if strings.Contains(strings.ToLower(reason), "logistik") {
				if v := asFloat64(am, "amount"); v != nil {
					logistikAmt += math.Abs(*v)
				}
			}
		}
	}
	shipDisc := 0.0
	if v := asFloat64(income, "seller_shipping_discount"); v != nil {
		shipDisc = *v
	}
	escrowAmt := 0.0
	if v := asFloat64(income, "escrow_amount"); v != nil {
		escrowAmt = *v
	}
	actShip := 0.0
	if v := asFloat64(income, "actual_shipping_fee"); v != nil {
		actShip = *v
	}
	buyerShip := 0.0
	if v := asFloat64(income, "buyer_paid_shipping_fee"); v != nil {
		buyerShip = *v
	}
	shopeeRebate := 0.0
	if v := asFloat64(income, "shopee_shipping_rebate"); v != nil {
		shopeeRebate = *v
	}
	diff := actShip - buyerShip - shopeeRebate - shipDisc

	// Logistic compensation occurs when the item is lost in transit and the
	// logistic provider reimburses the seller.  Shopee records this as an
	// order adjustment with reason "Logistik".  In this scenario Shopee does
	// not charge any fees and the seller receives the reimbursed amount
	// directly in escrow.  When such an adjustment exists we simply transfer
	// the escrow amount from the pending account to the Shopee balance.
	logistikCase := logistikAmt > 0

	escrowAmt = escrowAmt - shipDisc

	debitTotal := commission + service + voucher + discount + shipDisc + affiliate + diff + escrowAmt
	if !logistikCase && math.Abs(debitTotal-orderPrice) > 0.01 {
		log.Printf("unbalanced journal for %s: debit %.2f credit %.2f", invoice, debitTotal, orderPrice)
		log.Printf("  commission: %.2f, service: %.2f, voucher: %.2f, discount: %.2f, shipDisc: %.2f, affiliate: %.2f, escrowAmt: %.2f, diff: %.2f",
			commission, service, voucher, discount, shipDisc, affiliate, escrowAmt, diff)
		log.Printf("  actShip: %.2f, buyerShip: %.2f, rebate: %.2f", actShip, buyerShip, shopeeRebate)
		log.Printf("  orderPrice: %.2f", orderPrice)
		return fmt.Errorf("unbalanced journal: debit %.2f credit %.2f", debitTotal, orderPrice)
	}

	je := &models.JournalEntry{
		EntryDate:    updateTime,
		Description:  ptrString("Shopee escrow " + invoice),
		SourceType:   "shopee_escrow",
		SourceID:     invoice,
		ShopUsername: dp.NamaToko,
		Store:        dp.NamaToko,
		CreatedAt:    time.Now(),
	}
	jid, err := jrRepo.CreateJournalEntry(ctx, je)
	if err != nil {
		return err
	}
	var lines []models.JournalLine
	if logistikCase {
		lines = []models.JournalLine{
			{JournalID: jid, AccountID: pendingAccountID(dp.NamaToko), IsDebit: false, Amount: escrowAmt, Memo: ptrString("Pending " + invoice)},
			{JournalID: jid, AccountID: saldoShopeeAccountID(dp.NamaToko), IsDebit: true, Amount: escrowAmt, Memo: ptrString("Saldo Shopee " + invoice)},
		}
	} else {
		lines = []models.JournalLine{
			{JournalID: jid, AccountID: pendingAccountID(dp.NamaToko), IsDebit: false, Amount: orderPrice, Memo: ptrString("Pending " + invoice)},
			{JournalID: jid, AccountID: 52006, IsDebit: true, Amount: commission, Memo: ptrString("Biaya Administrasi " + invoice)},
			{JournalID: jid, AccountID: 52004, IsDebit: true, Amount: service, Memo: ptrString("Biaya Layanan " + invoice)},
			{JournalID: jid, AccountID: 55001, IsDebit: true, Amount: voucher, Memo: ptrString("Voucher " + invoice)},
			{JournalID: jid, AccountID: 55004, IsDebit: true, Amount: discount, Memo: ptrString("Discount " + invoice)},
			{JournalID: jid, AccountID: 55006, IsDebit: true, Amount: shipDisc, Memo: ptrString("Diskon Ongkir " + invoice)},
			{JournalID: jid, AccountID: 55002, IsDebit: true, Amount: affiliate, Memo: ptrString("Biaya Affiliate " + invoice)},
			{JournalID: jid, AccountID: saldoShopeeAccountID(dp.NamaToko), IsDebit: true, Amount: escrowAmt, Memo: ptrString("Saldo Shopee " + invoice)},
		}
	}
	if diff > 0 {
		lines = append(lines,
			models.JournalLine{JournalID: jid, AccountID: saldoShopeeAccountID(dp.NamaToko), IsDebit: true, Amount: diff, Memo: ptrString("Selisih Ongkir " + invoice)},
			models.JournalLine{JournalID: jid, AccountID: 4001, IsDebit: false, Amount: diff, Memo: ptrString("Selisih Ongkir " + invoice)},
		)
	} else if diff < 0 {
		aamt := -diff
		lines = append(lines,
			models.JournalLine{JournalID: jid, AccountID: 52010, IsDebit: true, Amount: aamt, Memo: ptrString("Selisih Ongkir " + invoice)},
			models.JournalLine{JournalID: jid, AccountID: saldoShopeeAccountID(dp.NamaToko), IsDebit: false, Amount: aamt, Memo: ptrString("Selisih Ongkir " + invoice)},
		)
	}
	for i := range lines {
		if lines[i].Amount == 0 {
			continue
		}
		if err := jrRepo.InsertJournalLine(ctx, &lines[i]); err != nil {
			return err
		}
	}
	if math.Abs(diff) > 0.01 && s.adjRepo != nil {
		adj := &models.ShopeeAdjustment{
			NamaToko:           dp.NamaToko,
			TanggalPenyesuaian: updateTime,
			TipePenyesuaian:    "Shipping Fee Discrepancy",
			AlasanPenyesuaian:  "Auto from escrow",
			BiayaPenyesuaian:   diff,
			NoPesanan:          invoice,
			CreatedAt:          time.Now(),
		}
		if err := s.adjRepo.Delete(ctx, adj.NoPesanan, adj.TanggalPenyesuaian, adj.TipePenyesuaian); err != nil {
			return err
		}
		if err := s.adjRepo.Insert(ctx, adj); err != nil {
			return err
		}
		// Journal lines for the shipping fee discrepancy have already
		// been created above as part of the escrow settlement journal,
		// so we only persist the adjustment record here to avoid
		// double posting.
	}

	if s.detailRepo != nil {
		if err := s.detailRepo.UpdateOrderDetailStatus(ctx, dp.KodeInvoiceChannel, status, status, updateTime); err != nil {
			log.Printf("update order detail %s: %v", invoice, err)
		}
	}

	if err := dropRepo.UpdatePurchaseStatus(ctx, dp.KodePesanan, "Pesanan selesai"); err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	if tx != nil {
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}
func (s *ReconcileService) UpdateShopeeStatuses(ctx context.Context, invoices []string) error {
	if len(invoices) == 0 {
		return nil
	}
	batches := make(map[string][]*models.DropshipPurchase)
	for _, inv := range invoices {
		dp, err := s.dropRepo.GetDropshipPurchaseByInvoice(ctx, inv)
		if err != nil || dp == nil {
			log.Printf("fetch purchase %s: %v", inv, err)
			continue
		}
		batches[dp.NamaToko] = append(batches[dp.NamaToko], dp)
	}

	g, ctx := errgroup.WithContext(ctx)
	for store, list := range batches {
		store := store
		list := list
		var batchID int64
		if s.batchSvc != nil {
			bh := &models.BatchHistory{ProcessType: "shopee_status_batch", TotalData: len(list), DoneData: 0}
			var err error
			batchID, err = s.batchSvc.Create(ctx, bh)
			if err != nil {
				log.Printf("create batch history %s: %v", store, err)
			}
		}
		g.Go(func() error {
			s.processShopeeStatusBatch(ctx, store, list)
			if s.batchSvc != nil && batchID != 0 {
				s.batchSvc.UpdateDone(ctx, batchID, len(list))
				s.batchSvc.UpdateStatus(ctx, batchID, "completed", "")
			}
			return nil
		})
	}
	return g.Wait()
}

func (s *ReconcileService) processShopeeStatusBatch(ctx context.Context, store string, list []*models.DropshipPurchase) {
	st, err := s.storeRepo.GetStoreByName(ctx, store)
	if err != nil || st == nil || st.AccessToken == nil || st.ShopID == nil {
		log.Printf("fetch store %s: %v", store, err)
		return
	}
	if err := s.ensureStoreTokenValid(ctx, st); err != nil {
		log.Printf("ensure token %s: %v", store, err)
		return
	}
	sns := make([]string, len(list))
	dpMap := make(map[string]*models.DropshipPurchase, len(list))
	for i, dp := range list {
		sns[i] = dp.KodeInvoiceChannel
		dpMap[dp.KodeInvoiceChannel] = dp
	}
	details, err := s.client.FetchShopeeOrderDetails(ctx, *st.AccessToken, *st.ShopID, sns)
	if err != nil && strings.Contains(err.Error(), "invalid_access_token") {
		if e := s.ensureStoreTokenValid(ctx, st); e == nil {
			details, err = s.client.FetchShopeeOrderDetails(ctx, *st.AccessToken, *st.ShopID, sns)
		}
	}
	if err != nil {
		log.Printf("batch fetch detail %s: %v", store, err)
		return
	}
	completed := []string{}
	timeMap := make(map[string]time.Time)
	for _, det := range details {
		sn, _ := det["order_sn"].(string)
		statusVal, ok := det["order_status"]
		if !ok {
			statusVal = det["status"]
		}
		statusStr, _ := statusVal.(string)
		var updateTime time.Time
		if ts, ok := det["update_time"]; ok {
			switch v := ts.(type) {
			case float64:
				updateTime = time.Unix(int64(v), 0)
			case int64:
				updateTime = time.Unix(v, 0)
			case int:
				updateTime = time.Unix(int64(v), 0)
			case string:
				if t, err := strconv.ParseInt(v, 10, 64); err == nil {
					updateTime = time.Unix(t, 0)
				}
			}
		}
		if updateTime.IsZero() {
			updateTime = time.Now()
		}
		dp := dpMap[sn]
		if dp == nil {
			continue
		}
		status := strings.ToLower(statusStr)
		if status == "completed" {
			completed = append(completed, sn)
			timeMap[sn] = updateTime
			continue
		}
		if status == "cancelled" {
			if err := s.CancelPurchaseAt(ctx, dp.KodePesanan, updateTime, "Cancelled Shopee"); err != nil {
				log.Printf("cancel purchase %s: %v", dp.KodePesanan, err)
			}
		}
	}
	if len(completed) > 0 {
		escMap, err := s.client.FetchShopeeEscrowDetails(ctx, *st.AccessToken, *st.ShopID, completed)
		if err != nil && strings.Contains(err.Error(), "invalid_access_token") {
			if e := s.ensureStoreTokenValid(ctx, st); e == nil {
				escMap, err = s.client.FetchShopeeEscrowDetails(ctx, *st.AccessToken, *st.ShopID, completed)
			}
		}
		if err != nil {
			log.Printf("batch escrow detail %s: %v", store, err)
		} else {
			var wg sync.WaitGroup
			limit := s.maxThreads
			if limit <= 0 {
				limit = 5
			}
			sem := make(chan struct{}, limit)

			log.Printf("Processing %s", escMap)

			for sn, esc := range escMap {
				wg.Add(1)
				sem <- struct{}{}
				go func(sn string, esc ShopeeEscrowDetail) {
					defer func() { <-sem; wg.Done() }()
					log.Printf("Processing escrow settlement for %s", esc)
					inv := sn
					if dp, ok := dpMap[sn]; ok {
						log.Printf("Found DropshipPurchase for %s", dp.KodeInvoiceChannel)
						inv = dp.KodeInvoiceChannel
					}
					log.Printf("Processing escrow settlement for %s", inv)
					if err := s.createEscrowSettlementJournal(ctx, inv, "completed", timeMap[sn], &esc); err != nil {
						log.Printf("escrow settlement %s: %v", sn, err)
					}
				}(sn, esc)
			}
			wg.Wait()
		}
	}
}
