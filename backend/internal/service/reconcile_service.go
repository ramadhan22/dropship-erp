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
	InsertJournalLines(ctx context.Context, lines []models.JournalLine) error
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
	CreateDetail(ctx context.Context, d *models.BatchHistoryDetail) error
	ListDetails(ctx context.Context, batchID int64) ([]models.BatchHistoryDetail, error)
	UpdateDetailStatus(ctx context.Context, id int64, status, msg string) error
}

type ReconcileServiceFailedRepo interface {
	InsertFailedReconciliation(ctx context.Context, failed *models.FailedReconciliation) error
	GetFailedReconciliationsByShop(ctx context.Context, shop string, limit, offset int) ([]models.FailedReconciliation, error)
	GetFailedReconciliationsByBatch(ctx context.Context, batchID int64) ([]models.FailedReconciliation, error)
	CountFailedReconciliationsByErrorType(ctx context.Context, shop string, since time.Time) (map[string]int, error)
	MarkAsRetried(ctx context.Context, id int64) error
	GetUnretriedFailedReconciliations(ctx context.Context, shop string, limit int) ([]models.FailedReconciliation, error)
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
	failedRepo  ReconcileServiceFailedRepo
	maxThreads  int
	config      *models.ReconciliationConfig
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
	fr ReconcileServiceFailedRepo,
	maxThreads int,
	config *models.ReconciliationConfig,
) *ReconcileService {
	// Set default config if not provided
	if config == nil {
		config = &models.ReconciliationConfig{
			MaxAllowedFailures:      100,
			FailureThresholdPercent: 5.0,
			CriticalErrorTypes:      []string{"database_error", "critical_system_error"},
			RetryFailedTransactions: false,
			GenerateDetailedReport:  true,
		}
	}
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
		failedRepo:  fr,
		maxThreads:  maxThreads,
		config:      config,
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

// bulkGetDropshipPurchasesByInvoices fetches multiple purchases efficiently
func (s *ReconcileService) bulkGetDropshipPurchasesByInvoices(ctx context.Context, invoices []string) ([]*models.DropshipPurchase, error) {
	// Check if the repository supports bulk operations
	if bulkRepo, ok := s.dropRepo.(interface {
		GetDropshipPurchasesByInvoices(ctx context.Context, invoices []string) ([]*models.DropshipPurchase, error)
	}); ok {
		return bulkRepo.GetDropshipPurchasesByInvoices(ctx, invoices)
	}

	// Fallback to individual calls if bulk method not available
	purchases := make([]*models.DropshipPurchase, 0, len(invoices))
	for _, inv := range invoices {
		dp, err := s.dropRepo.GetDropshipPurchaseByInvoice(ctx, inv)
		if err != nil {
			log.Printf("fetch purchase %s: %v", inv, err)
			continue
		}
		if dp != nil {
			purchases = append(purchases, dp)
		}
	}
	return purchases, nil
}

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
func (s *ReconcileService) ListCandidates(ctx context.Context, shop, order, status, from, to string, limit, offset int) ([]models.ReconcileCandidate, int, error) {
	if repo, ok := s.recRepo.(interface {
		ListCandidates(context.Context, string, string, string, string, string, int, int) ([]models.ReconcileCandidate, int, error)
	}); ok {
		list, total, err := repo.ListCandidates(ctx, shop, order, status, from, to, limit, offset)
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

// BulkReconcileWithErrorHandling processes reconciliation pairs with robust error handling.
// It continues processing even when individual transactions fail and provides a detailed report.
func (s *ReconcileService) BulkReconcileWithErrorHandling(ctx context.Context, pairs [][2]string, shop string, batchID *int64) (*models.ReconciliationReport, error) {
	startTime := time.Now()
	log.Printf("BulkReconcileWithErrorHandling %d pairs for shop %s", len(pairs), shop)

	report := &models.ReconciliationReport{
		TotalTransactions:      len(pairs),
		SuccessfulTransactions: 0,
		FailedTransactions:     0,
		ProcessingStartTime:    startTime,
		FailureCategories:      make(map[string]int),
		FailedTransactionList:  []models.FailedReconciliation{},
	}

	for _, p := range pairs {
		err := s.MatchAndJournal(ctx, p[0], p[1], shop)
		if err != nil {
			// Handle the error gracefully
			if failErr := s.recordFailedReconciliation(ctx, p[0], &p[1], shop, err, batchID); failErr != nil {
				log.Printf("Failed to record failure for %s: %v", p[0], failErr)
			}

			// Update report
			report.FailedTransactions++
			errorType := s.categorizeError(err)
			report.FailureCategories[errorType]++

			// Check if we should halt processing
			if s.shouldHaltProcessing(report, errorType) {
				log.Printf("Halting reconciliation due to critical error or failure threshold")
				break
			}

			continue
		}

		report.SuccessfulTransactions++
	}

	// Finalize report
	endTime := time.Now()
	report.ProcessingEndTime = endTime
	report.Duration = endTime.Sub(startTime).String()
	if report.TotalTransactions > 0 {
		report.FailureRate = float64(report.FailedTransactions) / float64(report.TotalTransactions) * 100
	}

	// Add failed transaction details if configured
	if s.config.GenerateDetailedReport && report.FailedTransactions > 0 && batchID != nil {
		if failedList, err := s.failedRepo.GetFailedReconciliationsByBatch(ctx, *batchID); err == nil {
			report.FailedTransactionList = failedList
		}
	}

	log.Printf("BulkReconcileWithErrorHandling completed: %d successful, %d failed, %.2f%% failure rate",
		report.SuccessfulTransactions, report.FailedTransactions, report.FailureRate)

	return report, nil
}

// recordFailedReconciliation stores details of a failed reconciliation transaction.
func (s *ReconcileService) recordFailedReconciliation(ctx context.Context, purchaseID string, orderID *string, shop string, err error, batchID *int64) error {
	if s.failedRepo == nil {
		return nil // Skip if failed repo not configured
	}

	failed := &models.FailedReconciliation{
		PurchaseID: purchaseID,
		OrderID:    orderID,
		Shop:       shop,
		ErrorType:  s.categorizeError(err),
		ErrorMsg:   err.Error(),
		FailedAt:   time.Now(),
		BatchID:    batchID,
	}

	return s.failedRepo.InsertFailedReconciliation(ctx, failed)
}

// categorizeError classifies errors into categories for reporting and handling.
func (s *ReconcileService) categorizeError(err error) string {
	errStr := strings.ToLower(err.Error())

	if strings.Contains(errStr, "context canceled") || strings.Contains(errStr, "context deadline") {
		return "timeout_error"
	}
	if strings.Contains(errStr, "database") || strings.Contains(errStr, "sql") {
		return "database_error"
	}
	if strings.Contains(errStr, "connection") || strings.Contains(errStr, "network") {
		return "network_error"
	}
	if strings.Contains(errStr, "fetch dropshippurchase") {
		return "purchase_not_found"
	}
	if strings.Contains(errStr, "fetch shopeeorder") {
		return "shopee_order_not_found"
	}
	if strings.Contains(errStr, "create journalentry") {
		return "journal_creation_error"
	}
	if strings.Contains(errStr, "unbalanced journal") {
		return "journal_balance_error"
	}
	if strings.Contains(errStr, "access_token") || strings.Contains(errStr, "authentication") {
		return "authentication_error"
	}

	return "unknown_error"
}

// shouldHaltProcessing determines if the reconciliation process should stop based on error conditions.
func (s *ReconcileService) shouldHaltProcessing(report *models.ReconciliationReport, errorType string) bool {
	// Check for critical error types
	for _, criticalType := range s.config.CriticalErrorTypes {
		if errorType == criticalType {
			return true
		}
	}

	// Check failure count threshold
	if s.config.MaxAllowedFailures > 0 && report.FailedTransactions >= s.config.MaxAllowedFailures {
		return true
	}

	// Check failure rate threshold (only if we have processed enough transactions)
	if report.TotalTransactions >= 10 { // Only check rate after processing at least 10 transactions
		processed := report.SuccessfulTransactions + report.FailedTransactions
		if processed > 0 {
			currentFailureRate := float64(report.FailedTransactions) / float64(processed) * 100
			if currentFailureRate > s.config.FailureThresholdPercent {
				return true
			}
		}
	}

	return false
}

// CheckAndMarkComplete verifies a purchase has been properly settled and updates
// its status accordingly. This method checks if an order is complete by looking
// for escrow settlement journals or existing completion status.
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

	// If already marked as complete, no need to check further
	if dp.StatusPesananTerakhir == "Pesanan selesai" {
		log.Printf("CheckAndMarkComplete: %s already marked as complete", kodePesanan)
		return nil
	}

	// Check if escrow settlement journal exists (primary indicator of completion)
	isSettled := s.hasEscrowSettlement(ctx, dp.KodeInvoiceChannel)

	if !isSettled {
		return fmt.Errorf("order not yet settled - no escrow settlement journal found")
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

// hasEscrowSettlement checks if an escrow settlement journal exists for the given invoice
func (s *ReconcileService) hasEscrowSettlement(ctx context.Context, invoice string) bool {
	if journalRepo, ok := s.journalRepo.(interface {
		ExistsBySourceTypeAndID(ctx context.Context, sourceType, sourceID string) (bool, error)
	}); ok {
		exists, err := journalRepo.ExistsBySourceTypeAndID(ctx, "shopee_escrow", invoice)
		if err != nil {
			log.Printf("check escrow journal for %s: %v", invoice, err)
			return false
		}
		return exists
	}
	return false
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
	// Filter out lines with zero amounts and use bulk insert
	validLines := make([]models.JournalLine, 0, len(lines))
	for i := range lines {
		if lines[i].Amount != 0 {
			validLines = append(validLines, lines[i])
		}
	}
	if len(validLines) > 0 {
		if err := jrRepo.InsertJournalLines(ctx, validLines); err != nil {
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
	log.Printf("ensureStoreTokenValid store=%d", st.StoreID)
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

	// Handle returned orders - both full and partial returns
	if status == "returned" || status == "partial_return" || strings.Contains(strings.ToLower(statusStr), "return") {
		// Check if return journal already exists to avoid duplicates
		if s.HasReturnJournal(ctx, invoice) {
			log.Printf("Return journal already exists for %s, skipping", invoice)
			return nil
		}

		// Get escrow detail to determine return amounts
		escDetail, err := s.GetShopeeEscrowDetail(ctx, invoice)
		if err != nil {
			return fmt.Errorf("get escrow detail for return %s: %w", invoice, err)
		}

		// Determine if this is a partial return and extract return amount
		isPartialReturn := status == "partial_return" || strings.Contains(strings.ToLower(statusStr), "partial")
		returnAmount := 0.0

		// For partial returns, extract the actual return amount from escrow detail
		if isPartialReturn {
			m := map[string]any(*escDetail)
			if income, ok := m["order_income"].(map[string]any); ok {
				if refundAmt, ok := income["refund_amount"]; ok {
					if v := asFloat64(map[string]any{"refund_amount": refundAmt}, "refund_amount"); v != nil {
						returnAmount = *v
					}
				}
				// Fallback: calculate from order adjustments if refund_amount not available
				if returnAmount == 0 {
					if adjList, ok := income["order_adjustment"].([]any); ok {
						for _, a := range adjList {
							am, ok := a.(map[string]any)
							if !ok {
								continue
							}
							reason, _ := am["adjustment_reason"].(string)
							if strings.Contains(strings.ToLower(reason), "return") || strings.Contains(strings.ToLower(reason), "refund") {
								if v := asFloat64(am, "amount"); v != nil {
									returnAmount += math.Abs(*v)
								}
							}
						}
					}
				}
			}
		}

		if err := s.createReturnedOrderJournal(ctx, invoice, statusStr, updateTime, escDetail, isPartialReturn, returnAmount); err != nil {
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

	// escrowAmt = escrowAmt - shipDisc

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
	if diff < 0 {
		aamt := -diff
		lines = append(lines,
			models.JournalLine{JournalID: jid, AccountID: 4001, IsDebit: false, Amount: aamt, Memo: ptrString("Selisih Ongkir Lebih" + invoice)},
		)
	} else if diff > 0 {
		lines = append(lines,
			models.JournalLine{JournalID: jid, AccountID: 52010, IsDebit: true, Amount: diff, Memo: ptrString("Selisih Ongkir Kurang" + invoice)},
		)
	}
	// Calculate total debits and credits to ensure the journal is balanced
	totalDebits := 0.0
	totalCredits := 0.0
	for _, line := range lines {
		if line.Amount == 0 {
			continue
		}
		if line.IsDebit {
			totalDebits += line.Amount
		} else {
			totalCredits += line.Amount
		}
	}

	// Check if journal is balanced
	if math.Abs(totalDebits-totalCredits) > 0.01 {
		log.Printf("unbalanced escrow settlement journal for %s: debits %.2f != credits %.2f", invoice, totalDebits, totalCredits)
		for i, line := range lines {
			if line.Amount == 0 {
				continue
			}
			debitStr := "credit"
			if line.IsDebit {
				debitStr = "debit"
			}
			log.Printf("  line %d: account %d, %s %.2f", i, line.AccountID, debitStr, line.Amount)
		}
		if tx != nil {
			tx.Rollback()
		}
		return fmt.Errorf("unbalanced journal entries: debits %.2f != credits %.2f", totalDebits, totalCredits)
	}

	// Filter out lines with zero amounts and use bulk insert
	validLines := make([]models.JournalLine, 0, len(lines))
	for i := range lines {
		if lines[i].Amount != 0 {
			validLines = append(validLines, lines[i])
		}
	}
	if len(validLines) > 0 {
		if err := jrRepo.InsertJournalLines(ctx, validLines); err != nil {
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

// createReturnedOrderJournal handles journal entries for returned orders in escrow settlements.
// It reverses the original escrow settlement and records appropriate refund entries.
func (s *ReconcileService) createReturnedOrderJournal(ctx context.Context, invoice, status string, updateTime time.Time, escDetail *ShopeeEscrowDetail, isPartialReturn bool, returnAmount float64) error {
	log.Printf("createReturnedOrderJournal %s (partial: %t, amount: %.2f)", invoice, isPartialReturn, returnAmount)

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

	log.Printf("Fetching DropshipPurchase for returned invoice %s", invoice)
	dp, err := dropRepo.GetDropshipPurchaseByInvoice(ctx, invoice)
	if err != nil || dp == nil {
		return fmt.Errorf("fetch purchase %s: %w", invoice, err)
	}

	// Get escrow detail if not provided
	if escDetail == nil {
		escDetail, err = s.GetShopeeEscrowDetail(ctx, invoice)
		if err != nil {
			return err
		}
	}

	m := map[string]any(*escDetail)
	income, _ := m["order_income"].(map[string]any)

	// Extract amounts from escrow detail
	orderPrice := 0.0
	if v := asFloat64(income, "order_original_price"); v != nil {
		orderPrice = *v
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
	shipDisc := 0.0
	if v := asFloat64(income, "seller_shipping_discount"); v != nil {
		shipDisc = *v
	}

	// Handle BD Marketing adjustments for affiliate fees
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
		}
	}

	// Calculate return proportion for partial returns
	returnProportion := 1.0
	if isPartialReturn && orderPrice > 0 {
		returnProportion = returnAmount / orderPrice
	}

	// Scale amounts proportionally for partial returns
	returnCommission := commission * returnProportion
	returnService := service * returnProportion
	returnVoucher := voucher * returnProportion
	returnDiscount := discount * returnProportion
	returnShipDisc := shipDisc * returnProportion
	returnAffiliate := affiliate * returnProportion
	actualReturnAmount := orderPrice * returnProportion

	je := &models.JournalEntry{
		EntryDate: updateTime,
		Description: ptrString(fmt.Sprintf("Shopee return %s%s", invoice, func() string {
			if isPartialReturn {
				return fmt.Sprintf(" (partial %.2f)", actualReturnAmount)
			}
			return ""
		}())),
		SourceType:   "shopee_return",
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

	// For returned orders, we need to reverse the escrow settlement journal entries
	// and record the refund to the customer
	lines = []models.JournalLine{
		// Reverse the pending account credit (now debit)
		{JournalID: jid, AccountID: pendingAccountID(dp.NamaToko), IsDebit: true, Amount: actualReturnAmount, Memo: ptrString("Reverse pending " + invoice)},

		// Reverse the expense account debits (now credits)
		{JournalID: jid, AccountID: 52006, IsDebit: false, Amount: returnCommission, Memo: ptrString("Reverse commission " + invoice)},
		{JournalID: jid, AccountID: 52004, IsDebit: false, Amount: returnService, Memo: ptrString("Reverse service fee " + invoice)},
		{JournalID: jid, AccountID: 55001, IsDebit: false, Amount: returnVoucher, Memo: ptrString("Reverse voucher " + invoice)},
		{JournalID: jid, AccountID: 55004, IsDebit: false, Amount: returnDiscount, Memo: ptrString("Reverse discount " + invoice)},
		{JournalID: jid, AccountID: 55006, IsDebit: false, Amount: returnShipDisc, Memo: ptrString("Reverse shipping discount " + invoice)},
		{JournalID: jid, AccountID: 55002, IsDebit: false, Amount: returnAffiliate, Memo: ptrString("Reverse affiliate " + invoice)},

		// Record refund to customer using refund account
		{JournalID: jid, AccountID: 52009, IsDebit: true, Amount: actualReturnAmount, Memo: ptrString("Refund " + invoice)},
	}

	// Filter out lines with zero amounts and use bulk insert
	validLines := make([]models.JournalLine, 0, len(lines))
	for i := range lines {
		if lines[i].Amount != 0 {
			validLines = append(validLines, lines[i])
		}
	}
	if len(validLines) > 0 {
		if err := jrRepo.InsertJournalLines(ctx, validLines); err != nil {
			return err
		}
	}

	// Update order detail status if available
	if s.detailRepo != nil {
		returnStatus := "returned"
		if isPartialReturn {
			returnStatus = "partial_return"
		}
		if err := s.detailRepo.UpdateOrderDetailStatus(ctx, dp.KodeInvoiceChannel, returnStatus, returnStatus, updateTime); err != nil {
			log.Printf("update order detail status %s: %v", invoice, err)
		}
	}

	// Update purchase status
	purchaseStatus := "Pesanan dikembalikan"
	if isPartialReturn {
		purchaseStatus = "Sebagian dikembalikan"
	}
	if err := dropRepo.UpdatePurchaseStatus(ctx, dp.KodePesanan, purchaseStatus); err != nil {
		return fmt.Errorf("update purchase status: %w", err)
	}

	if tx != nil {
		if err := tx.Commit(); err != nil {
			return err
		}
	}

	log.Printf("createReturnedOrderJournal completed: %s", invoice)
	return nil
}

// ProcessReturnedOrder handles manual return processing from the reconcile dashboard.
// This method allows updating escrow status for returned orders with proper journal entries.

// HasReturnJournal checks if a return journal entry already exists for the given invoice
func (s *ReconcileService) HasReturnJournal(ctx context.Context, invoice string) bool {
	if journalRepo, ok := s.journalRepo.(interface {
		ExistsBySourceTypeAndID(ctx context.Context, sourceType, sourceID string) (bool, error)
	}); ok {
		exists, err := journalRepo.ExistsBySourceTypeAndID(ctx, "shopee_return", invoice)
		if err != nil {
			log.Printf("check return journal for %s: %v", invoice, err)
			return false
		}
		return exists
	}
	return false
}

func (s *ReconcileService) UpdateShopeeStatuses(ctx context.Context, invoices []string) error {
	if len(invoices) == 0 {
		return nil
	}

	start := time.Now()
	// Optimize: Bulk fetch all DropshipPurchases instead of individual calls
	log.Printf("UpdateShopeeStatuses: fetching %d purchases in bulk", len(invoices))
	purchases, err := s.bulkGetDropshipPurchasesByInvoices(ctx, invoices)
	if err != nil {
		log.Printf("bulk fetch purchases: %v", err)
		return err
	}
	fetchDuration := time.Since(start)
	log.Printf("UpdateShopeeStatuses: bulk fetch completed in %v for %d purchases", fetchDuration, len(purchases))

	batches := make(map[string][]*models.DropshipPurchase)
	for _, dp := range purchases {
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
	dpMap := make(map[string]*models.DropshipPurchase, len(list)*2)
	for i, dp := range list {
		sns[i] = dp.KodeInvoiceChannel
		dpMap[dp.KodeInvoiceChannel] = dp
		dpMap[dp.KodePesanan] = dp
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
	returned := []string{} // Track returned orders for separate processing
	timeMap := make(map[string]time.Time)
	statusMap := make(map[string]string) // Track status strings for returned orders
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
		if dp == nil && s.dropRepo != nil {
			if d, err := s.dropRepo.GetDropshipPurchaseByID(ctx, sn); err == nil && d != nil {
				dp = d
				dpMap[sn] = d
				dpMap[d.KodeInvoiceChannel] = d
			}
		}
		if dp == nil {
			continue
		}
		status := strings.ToLower(statusStr)
		if status == "completed" {
			completed = append(completed, sn)
			timeMap[sn] = updateTime
			continue
		}

		// Handle returned orders
		if status == "returned" || status == "partial_return" || strings.Contains(strings.ToLower(statusStr), "return") {
			returned = append(returned, sn)
			timeMap[sn] = updateTime
			statusMap[sn] = statusStr
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

			log.Printf("processing %d escrow settlements", len(escMap))

			for sn, esc := range escMap {
				wg.Add(1)
				sem <- struct{}{}
				go func(sn string, esc ShopeeEscrowDetail) {
					defer func() { <-sem; wg.Done() }()
					log.Printf("processing escrow for order %s", sn)
					inv := sn
					dp, ok := dpMap[sn]
					if !ok && s.dropRepo != nil {
						if d, err := s.dropRepo.GetDropshipPurchaseByID(ctx, sn); err == nil && d != nil {
							dp = d
							dpMap[sn] = d
							dpMap[d.KodeInvoiceChannel] = d
						}
					}
					if dp != nil {
						log.Printf("found DropshipPurchase invoice %s", dp.KodeInvoiceChannel)
						inv = dp.KodeInvoiceChannel
					}
					log.Printf("creating escrow settlement journal for %s", inv)
					if err := s.createEscrowSettlementJournal(ctx, inv, "completed", timeMap[sn], &esc); err != nil {
						log.Printf("escrow settlement %s: %v", sn, err)
					}
				}(sn, esc)
			}
			wg.Wait()
		}
	}

	// Process returned orders if any
	if len(returned) > 0 {
		log.Printf("processing %d returned orders", len(returned))
		escMap, err := s.client.FetchShopeeEscrowDetails(ctx, *st.AccessToken, *st.ShopID, returned)
		if err != nil && strings.Contains(err.Error(), "invalid_access_token") {
			if e := s.ensureStoreTokenValid(ctx, st); e == nil {
				escMap, err = s.client.FetchShopeeEscrowDetails(ctx, *st.AccessToken, *st.ShopID, returned)
			}
		}
		if err != nil {
			log.Printf("batch escrow detail for returns %s: %v", store, err)
		} else {
			var wg sync.WaitGroup
			limit := s.maxThreads
			if limit <= 0 {
				limit = 5
			}
			sem := make(chan struct{}, limit)

			log.Printf("processing %d returned order escrow details", len(escMap))

			for sn, esc := range escMap {
				wg.Add(1)
				sem <- struct{}{}
				go func(sn string, esc ShopeeEscrowDetail) {
					defer func() { <-sem; wg.Done() }()
					log.Printf("processing return escrow for order %s", sn)
					inv := sn
					dp, ok := dpMap[sn]
					if !ok && s.dropRepo != nil {
						if d, err := s.dropRepo.GetDropshipPurchaseByID(ctx, sn); err == nil && d != nil {
							dp = d
							dpMap[sn] = d
							dpMap[d.KodeInvoiceChannel] = d
						}
					}
					if dp != nil {
						log.Printf("found DropshipPurchase invoice %s for return", dp.KodeInvoiceChannel)
						inv = dp.KodeInvoiceChannel
					}

					// Determine if partial return and extract return amount
					statusStr := statusMap[sn]
					isPartialReturn := strings.Contains(strings.ToLower(statusStr), "partial") || strings.Contains(strings.ToLower(statusStr), "partial_return")
					returnAmount := 0.0

					if isPartialReturn {
						m := map[string]any(esc)
						if income, ok := m["order_income"].(map[string]any); ok {
							if refundAmt, ok := income["refund_amount"]; ok {
								if v := asFloat64(map[string]any{"refund_amount": refundAmt}, "refund_amount"); v != nil {
									returnAmount = *v
								}
							}
						}
					}

					log.Printf("creating returned order journal for %s (partial: %t, amount: %.2f)", inv, isPartialReturn, returnAmount)
					if err := s.createReturnedOrderJournal(ctx, inv, statusStr, timeMap[sn], &esc, isPartialReturn, returnAmount); err != nil {
						log.Printf("returned order journal %s: %v", sn, err)
					}
				}(sn, esc)
			}
			wg.Wait()
		}
	}
}

// CreateReconcileBatches groups reconciliation candidates by store and records
// them as batch_history rows with associated details. Each batch contains at
// most 50 invoices. Returns information about the created batches.
func (s *ReconcileService) CreateReconcileBatches(ctx context.Context, shop, order, status, from, to string) (*models.ReconcileBatchInfo, error) {
	if s.batchSvc == nil {
		return nil, fmt.Errorf("batch service not configured")
	}

	log.Printf("CreateReconcileBatches: fetching candidates for shop=%s, order=%s, status=%s, from=%s, to=%s", shop, order, status, from, to)
	pageSize := 1000
	batchSize := 50 // Process in batches of 50 orders
	offset := 0
	all := []models.ReconcileCandidate{}
	for {
		list, total, err := s.ListCandidates(ctx, shop, order, status, from, to, pageSize, offset)
		if err != nil {
			return nil, err
		}
		all = append(all, list...)
		if len(all) >= total {
			break
		}
		offset += pageSize
	}

	log.Printf("CreateReconcileBatches: found %d total candidates", len(all))

	batches := make(map[string][]models.ReconcileCandidate)
	for _, c := range all {
		batches[c.NamaToko] = append(batches[c.NamaToko], c)
	}

	batchCount := 0
	for store, list := range batches {
		log.Printf("CreateReconcileBatches: processing %d candidates for store %s", len(list), store)
		for i := 0; i < len(list); i += batchSize {
			end := i + batchSize
			if end > len(list) {
				end = len(list)
			}
			subset := list[i:end]
			bh := &models.BatchHistory{ProcessType: "reconcile_batch", TotalData: len(subset), DoneData: 0, Status: "pending"}
			batchID, err := s.batchSvc.Create(ctx, bh)
			if err != nil {
				return nil, err
			}
			batchCount++

			for _, cand := range subset {
				d := &models.BatchHistoryDetail{BatchID: batchID, Reference: cand.KodeInvoiceChannel, Store: store, Status: "pending"}
				if err := s.batchSvc.CreateDetail(ctx, d); err != nil {
					log.Printf("create batch detail %s: %v", store, err)
				}
			}
		}
	}

	result := &models.ReconcileBatchInfo{
		BatchCount:        batchCount,
		TotalTransactions: len(all),
	}

	log.Printf("CreateReconcileBatches: created %d batches for %d total transactions", result.BatchCount, result.TotalTransactions)
	return result, nil
}

// ProcessReconcileBatch processes a batch of reconciliation tasks with optimizations and robust error handling.
// Shopee statuses are updated in bulk before marking each purchase complete.
func (s *ReconcileService) ProcessReconcileBatch(ctx context.Context, id int64) {
	if s.batchSvc == nil {
		return
	}

	start := time.Now()
	log.Printf("ProcessReconcileBatch %d: starting batch processing", id)

	details, err := s.batchSvc.ListDetails(ctx, id)
	if err != nil {
		log.Printf("list batch details %d: %v", id, err)
		s.batchSvc.UpdateStatus(ctx, id, "failed", err.Error())
		return
	}
	s.batchSvc.UpdateStatus(ctx, id, "processing", "")

	invoices := make([]string, len(details))
	for i, d := range details {
		invoices[i] = d.Reference
	}

	// Update Shopee statuses in bulk first
	log.Printf("ProcessReconcileBatch %d: updating %d statuses", id, len(invoices))
	statusStart := time.Now()
	if err := s.UpdateShopeeStatuses(ctx, invoices); err != nil {
		log.Printf("update statuses batch %d: %v", id, err)
	}
	statusDuration := time.Since(statusStart)
	log.Printf("ProcessReconcileBatch %d: status update completed in %v", id, statusDuration)

	// Bulk fetch purchases to reduce database calls
	log.Printf("ProcessReconcileBatch %d: bulk fetching purchases", id)
	fetchStart := time.Now()
	purchases, err := s.bulkGetDropshipPurchasesByInvoices(ctx, invoices)
	if err != nil {
		log.Printf("bulk fetch purchases batch %d: %v", id, err)
		s.batchSvc.UpdateStatus(ctx, id, "failed", err.Error())
		return
	}
	fetchDuration := time.Since(fetchStart)
	log.Printf("ProcessReconcileBatch %d: bulk fetch completed in %v for %d purchases", id, fetchDuration, len(purchases))

	// Create lookup map for faster access
	purchaseMap := make(map[string]*models.DropshipPurchase)
	for _, dp := range purchases {
		purchaseMap[dp.KodeInvoiceChannel] = dp
	}

	// Process each detail with robust error handling
	done := 0
	failed := 0
	processStart := time.Now()

	for _, d := range details {
		dp, exists := purchaseMap[d.Reference]
		if !exists {
			msg := fmt.Sprintf("purchase not found for invoice %s", d.Reference)
			log.Printf("ProcessReconcileBatch %d: %s", id, msg)

			// Record the failure
			if s.failedRepo != nil {
				failedRec := &models.FailedReconciliation{
					PurchaseID: d.Reference,
					Shop:       d.Store,
					ErrorType:  "purchase_not_found",
					ErrorMsg:   msg,
					FailedAt:   time.Now(),
					BatchID:    &id,
				}
				if err := s.failedRepo.InsertFailedReconciliation(ctx, failedRec); err != nil {
					log.Printf("Failed to record failure for %s: %v", d.Reference, err)
				}
			}

			s.batchSvc.UpdateDetailStatus(ctx, d.ID, "failed", msg)
			failed++
			continue
		}

		if err := s.CheckAndMarkComplete(ctx, dp.KodePesanan); err != nil {
			log.Printf("ProcessReconcileBatch %d: CheckAndMarkComplete failed for %s: %v", id, dp.KodePesanan, err)

			// Record the failure
			if s.failedRepo != nil {
				failedRec := &models.FailedReconciliation{
					PurchaseID: dp.KodePesanan,
					Shop:       d.Store,
					ErrorType:  s.categorizeError(err),
					ErrorMsg:   err.Error(),
					FailedAt:   time.Now(),
					BatchID:    &id,
				}
				if err := s.failedRepo.InsertFailedReconciliation(ctx, failedRec); err != nil {
					log.Printf("Failed to record failure for %s: %v", dp.KodePesanan, err)
				}
			}

			s.batchSvc.UpdateDetailStatus(ctx, d.ID, "failed", err.Error())
			failed++
			continue
		}

		done++
		s.batchSvc.UpdateDone(ctx, id, done)
		s.batchSvc.UpdateDetailStatus(ctx, d.ID, "success", "")
	}

	processDuration := time.Since(processStart)
	totalDuration := time.Since(start)

	// Calculate failure rate and log comprehensive results
	total := len(details)
	failureRate := 0.0
	if total > 0 {
		failureRate = float64(failed) / float64(total) * 100
	}

	status := "completed"
	statusMsg := ""

	// Check if we should mark the batch as failed due to high failure rate
	if s.config != nil && failureRate > s.config.FailureThresholdPercent && total >= 10 {
		status = "completed_with_warnings"
		statusMsg = fmt.Sprintf("High failure rate: %.2f%%", failureRate)
	}

	s.batchSvc.UpdateStatus(ctx, id, status, statusMsg)
	log.Printf("ProcessReconcileBatch %d completed in %v: %d successful, %d failed, %.2f%% failure rate (status: %v, fetch: %v, process: %v)",
		id, totalDuration, done, failed, failureRate, statusDuration, fetchDuration, processDuration)
}

// GenerateReconciliationReport creates a comprehensive report for a given shop and time period.
func (s *ReconcileService) GenerateReconciliationReport(ctx context.Context, shop string, since time.Time) (*models.ReconciliationReport, error) {
	if s.failedRepo == nil {
		return nil, fmt.Errorf("failed reconciliation repository not configured")
	}

	// Get failure counts by error type
	failureCategories, err := s.failedRepo.CountFailedReconciliationsByErrorType(ctx, shop, since)
	if err != nil {
		return nil, fmt.Errorf("count failures by error type: %w", err)
	}

	// Calculate totals
	totalFailed := 0
	for _, count := range failureCategories {
		totalFailed += count
	}

	// Get detailed failed transaction list if configured
	var failedList []models.FailedReconciliation
	if s.config.GenerateDetailedReport {
		failedList, err = s.failedRepo.GetFailedReconciliationsByShop(ctx, shop, 100, 0) // Get last 100 failures
		if err != nil {
			log.Printf("Failed to get detailed failed transactions: %v", err)
		}
	}

	report := &models.ReconciliationReport{
		TotalTransactions:      0, // This would need to be calculated separately based on business logic
		SuccessfulTransactions: 0, // This would need to be calculated separately
		FailedTransactions:     totalFailed,
		FailureRate:            0, // Will be calculated if total is available
		ProcessingStartTime:    since,
		ProcessingEndTime:      time.Now(),
		Duration:               time.Since(since).String(),
		FailureCategories:      failureCategories,
		FailedTransactionList:  failedList,
	}

	return report, nil
}

// RetryFailedReconciliations attempts to reprocess failed reconciliation transactions.
func (s *ReconcileService) RetryFailedReconciliations(ctx context.Context, shop string, maxRetries int) (*models.ReconciliationReport, error) {
	if s.failedRepo == nil {
		return nil, fmt.Errorf("failed reconciliation repository not configured")
	}

	log.Printf("RetryFailedReconciliations: starting retry for shop %s, max %d retries", shop, maxRetries)
	startTime := time.Now()

	// Get unretried failed reconciliations
	failedList, err := s.failedRepo.GetUnretriedFailedReconciliations(ctx, shop, maxRetries)
	if err != nil {
		return nil, fmt.Errorf("get unretried failed reconciliations: %w", err)
	}

	report := &models.ReconciliationReport{
		TotalTransactions:      len(failedList),
		SuccessfulTransactions: 0,
		FailedTransactions:     0,
		ProcessingStartTime:    startTime,
		FailureCategories:      make(map[string]int),
		FailedTransactionList:  []models.FailedReconciliation{},
	}

	// Retry each failed transaction
	for _, failed := range failedList {
		var err error
		if failed.OrderID != nil {
			err = s.MatchAndJournal(ctx, failed.PurchaseID, *failed.OrderID, failed.Shop)
		} else {
			// For transactions without order ID, try to complete the purchase
			err = s.CheckAndMarkComplete(ctx, failed.PurchaseID)
		}

		// Mark as retried regardless of outcome
		if markErr := s.failedRepo.MarkAsRetried(ctx, failed.ID); markErr != nil {
			log.Printf("Failed to mark transaction %d as retried: %v", failed.ID, markErr)
		}

		if err != nil {
			// Still failed after retry
			report.FailedTransactions++
			errorType := s.categorizeError(err)
			report.FailureCategories[errorType]++

			// Record the new failure
			newFailed := &models.FailedReconciliation{
				PurchaseID: failed.PurchaseID,
				OrderID:    failed.OrderID,
				Shop:       failed.Shop,
				ErrorType:  errorType,
				ErrorMsg:   err.Error(),
				FailedAt:   time.Now(),
			}
			if recordErr := s.failedRepo.InsertFailedReconciliation(ctx, newFailed); recordErr != nil {
				log.Printf("Failed to record retry failure: %v", recordErr)
			}
		} else {
			// Retry succeeded
			report.SuccessfulTransactions++
		}
	}

	// Finalize report
	endTime := time.Now()
	report.ProcessingEndTime = endTime
	report.Duration = endTime.Sub(startTime).String()
	if report.TotalTransactions > 0 {
		report.FailureRate = float64(report.FailedTransactions) / float64(report.TotalTransactions) * 100
	}

	log.Printf("RetryFailedReconciliations completed: %d retried, %d successful, %d still failed, %.2f%% failure rate",
		report.TotalTransactions, report.SuccessfulTransactions, report.FailedTransactions, report.FailureRate)

	return report, nil
}

// GetFailedReconciliationsSummary provides a quick overview of failed reconciliations for a shop.
func (s *ReconcileService) GetFailedReconciliationsSummary(ctx context.Context, shop string, days int) (map[string]interface{}, error) {
	if s.failedRepo == nil {
		return nil, fmt.Errorf("failed reconciliation repository not configured")
	}

	since := time.Now().AddDate(0, 0, -days)

	// Get failure categories
	categories, err := s.failedRepo.CountFailedReconciliationsByErrorType(ctx, shop, since)
	if err != nil {
		return nil, err
	}

	// Calculate totals
	totalFailed := 0
	for _, count := range categories {
		totalFailed += count
	}

	summary := map[string]interface{}{
		"shop":               shop,
		"period_days":        days,
		"total_failed":       totalFailed,
		"failure_categories": categories,
		"since":              since,
	}

	return summary, nil
}
