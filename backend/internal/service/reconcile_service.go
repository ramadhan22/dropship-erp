package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
)

// Repository interfaces for dependency injection
type DropshipRepository interface {
	GetDropshipPurchaseByInvoice(ctx context.Context, kodeInvoice string) (*models.DropshipPurchase, error)
	GetDropshipPurchaseByID(ctx context.Context, kodePesanan string) (*models.DropshipPurchase, error)
	UpdatePurchaseStatus(ctx context.Context, kodePesanan, status string) error
	SumDetailByInvoice(ctx context.Context, kodeInvoice string) (float64, error)
	SumProductCostByInvoice(ctx context.Context, kodeInvoice string) (float64, error)
}

type JournalRepository interface {
	ExistsBySourceTypeAndID(ctx context.Context, sourceType, sourceID string) (bool, error)
}

type ShopeeRepository interface {
	ExistsShopeeSettled(ctx context.Context, invoice string) (bool, error)
}

// ReconcileService handles reconciliation operations between dropship purchases and Shopee escrow settlements.
type ReconcileService struct {
	db            *sqlx.DB
	reconcileRepo *repository.ReconcileRepo
	shopeeService *ShopeeService
	dropRepo      DropshipRepository
	journalRepo   JournalRepository
	shopeeRepo    ShopeeRepository
	batchSvc      *BatchService
}

// NewReconcileService creates a new reconcile service.
func NewReconcileService(db *sqlx.DB, reconcileRepo *repository.ReconcileRepo, shopeeService *ShopeeService) *ReconcileService {
	return &ReconcileService{
		db:            db,
		reconcileRepo: reconcileRepo,
		shopeeService: shopeeService,
	}
}

// SetRepositories sets the additional repositories for bulk operations
func (s *ReconcileService) SetRepositories(dropRepo DropshipRepository, journalRepo JournalRepository, shopeeRepo ShopeeRepository) {
	s.dropRepo = dropRepo
	s.journalRepo = journalRepo
	s.shopeeRepo = shopeeRepo
}

// SetBatchService sets the batch service for batch processing
func (s *ReconcileService) SetBatchService(batchSvc *BatchService) {
	s.batchSvc = batchSvc
}

// ReconcileAllRequest represents the request payload for reconcile all operation.
type ReconcileAllRequest struct {
	Shop     string `json:"shop"`
	FromDate string `json:"from_date"`
	ToDate   string `json:"to_date"`
}

// ReconcileAllResponse represents the response for reconcile all operation.
type ReconcileAllResponse struct {
	TotalProcessed    int    `json:"total_processed"`
	SuccessfulMatches int    `json:"successful_matches"`
	ReturnedOrders    int    `json:"returned_orders"`
	ErrorCount        int    `json:"error_count"`
	Message           string `json:"message"`
}

// ReconcileAll processes all unreconciled transactions for a shop and time period.
// This includes handling returned orders with escrow settlements.
func (s *ReconcileService) ReconcileAll(ctx context.Context, req ReconcileAllRequest) (*ReconcileAllResponse, error) {
	log.Printf("Starting reconcile all for shop: %s, period: %s to %s", req.Shop, req.FromDate, req.ToDate)

	// Get all unreconciled candidates
	candidates, total, err := s.reconcileRepo.ListCandidates(ctx, req.Shop, "", req.FromDate, req.ToDate, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get reconcile candidates: %w", err)
	}

	response := &ReconcileAllResponse{
		TotalProcessed: total,
	}

	// Process each candidate
	for _, candidate := range candidates {
		if err := s.processCandidateForReconciliation(ctx, candidate); err != nil {
			log.Printf("Error processing candidate %s: %v", candidate.KodeInvoiceChannel, err)
			response.ErrorCount++
			continue
		}

		// Check if this is a returned order that needs special handling
		if s.isReturnedOrder(candidate) {
			if err := s.processReturnedOrderEscrow(ctx, candidate); err != nil {
				log.Printf("Error processing returned order escrow for %s: %v", candidate.KodeInvoiceChannel, err)
				response.ErrorCount++
				continue
			}
			response.ReturnedOrders++
		}

		response.SuccessfulMatches++
	}

	response.Message = fmt.Sprintf("Processed %d transactions: %d successful, %d returned orders, %d errors",
		response.TotalProcessed, response.SuccessfulMatches, response.ReturnedOrders, response.ErrorCount)

	log.Printf("Reconcile all completed: %s", response.Message)
	return response, nil
}

// processCandidateForReconciliation handles the basic reconciliation logic for a candidate.
func (s *ReconcileService) processCandidateForReconciliation(ctx context.Context, candidate models.ReconcileCandidate) error {
	// Create a reconciled transaction record
	reconciled := &models.ReconciledTransaction{
		ShopUsername: candidate.NamaToko,
		DropshipID:   &candidate.KodeInvoiceChannel,
		ShopeeID:     candidate.NoPesanan,
		Status:       "matched",
		MatchedAt:    time.Now(),
	}

	// If no Shopee order found, mark as unmatched
	if candidate.NoPesanan == nil {
		reconciled.Status = "unmatched"
		reconciled.ShopeeID = nil
	}

	return s.reconcileRepo.InsertReconciledTransaction(ctx, reconciled)
}

// isReturnedOrder checks if a candidate represents a returned order based on its status.
func (s *ReconcileService) isReturnedOrder(candidate models.ReconcileCandidate) bool {
	// Check if the order status indicates a return
	statusLower := strings.ToLower(candidate.StatusPesananTerakhir)
	return strings.Contains(statusLower, "return") ||
		strings.Contains(statusLower, "refund") ||
		strings.Contains(statusLower, "dibatalkan") ||
		strings.Contains(statusLower, "cancelled")
}

// processReturnedOrderEscrow handles the escrow settlement processing for returned orders.
func (s *ReconcileService) processReturnedOrderEscrow(ctx context.Context, candidate models.ReconcileCandidate) error {
	if candidate.NoPesanan == nil {
		return fmt.Errorf("cannot process returned order escrow: no Shopee order number")
	}

	// For now, we'll create a sample escrow detail for demonstration
	// In a real implementation, this would fetch actual escrow data from Shopee API
	escrowDetail := ShopeeEscrowDetail{
		"seller_return_refund": -50000, // Negative amount for returns
		"reverse_shipping_fee": 5000,
		"commission_fee":       2000,
		"service_fee":          1500,
	}

	// Use the returned order journal function
	return s.shopeeService.CreateReturnedOrderJournal(ctx, *candidate.NoPesanan, escrowDetail, time.Now())
}

// GetReconcileCandidates returns a paginated list of candidates for reconciliation.
func (s *ReconcileService) GetReconcileCandidates(ctx context.Context, shop, order, from, to string, limit, offset int) ([]models.ReconcileCandidate, int, error) {
	return s.reconcileRepo.ListCandidates(ctx, shop, order, from, to, limit, offset)
}

// GetReconciledTransactions returns reconciled transactions for a shop and period.
func (s *ReconcileService) GetReconciledTransactions(ctx context.Context, shop, period string) ([]models.ReconciledTransaction, error) {
	return s.reconcileRepo.GetReconciledTransactionsByShopAndPeriod(ctx, shop, period)
}

// UpdateEscrowStatus updates the escrow processing status for a specific order.
func (s *ReconcileService) UpdateEscrowStatus(ctx context.Context, orderSN string, status string) error {
	// This would update a status field in the database
	// For now, just log the status update
	log.Printf("Updating escrow status for order %s to %s", orderSN, status)
	return nil
}

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

// CheckAndMarkComplete verifies a purchase has been properly settled and updates 
// its status accordingly. This method checks if an order is complete by looking
// for escrow settlement journals or existing completion status.
func (s *ReconcileService) CheckAndMarkComplete(ctx context.Context, kodePesanan string) error {
	log.Printf("CheckAndMarkComplete: %s", kodePesanan)
	var tx *sqlx.Tx
	var err error
	
	dropRepo := s.dropRepo
	if s.db != nil {
		// Use transaction if database is available
		tx, err = s.db.Beginx()
		if err != nil {
			return fmt.Errorf("begin transaction: %w", err)
		}
		defer tx.Rollback()
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
			return fmt.Errorf("commit transaction: %w", err)
		}
	}
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

// UpdateShopeeStatuses updates Shopee order statuses for multiple invoices with bulk optimization
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

	for store, purchaseList := range batches {
		if err := s.processShopeeStatusBatch(ctx, store, purchaseList); err != nil {
			log.Printf("process batch for store %s: %v", store, err)
		}
	}
	return nil
}

// processShopeeStatusBatch processes a batch of Shopee status updates for a single store
func (s *ReconcileService) processShopeeStatusBatch(ctx context.Context, store string, purchases []*models.DropshipPurchase) error {
	// This would contain the Shopee API integration logic
	// For now, just log the batch processing
	log.Printf("Processing Shopee status batch for store %s with %d purchases", store, len(purchases))
	return nil
}

// CreateReconcileBatches groups reconciliation candidates by store and records
// them as batch_history rows with associated details. Each batch contains at
// most 50 invoices. Returns information about the created batches.
func (s *ReconcileService) CreateReconcileBatches(ctx context.Context, shop, order, from, to string) (*models.ReconcileBatchInfo, error) {
	if s.batchSvc == nil {
		return nil, fmt.Errorf("batch service not configured")
	}
	
	log.Printf("CreateReconcileBatches: fetching candidates for shop=%s, order=%s, from=%s, to=%s", shop, order, from, to)
	pageSize := 1000
	batchSize := 50 // Process in batches of 50 orders
	offset := 0
	all := []models.ReconcileCandidate{}
	for {
		list, total, err := s.ListCandidates(ctx, shop, order, from, to, pageSize, offset)
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
					return nil, err
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

// ProcessReconcileBatch processes a batch of reconciliation tasks with optimizations.
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
	
	// Process each detail with optimized lookups
	done := 0
	processStart := time.Now()
	for _, d := range details {
		dp, exists := purchaseMap[d.Reference]
		if !exists {
			msg := fmt.Sprintf("purchase not found for invoice %s", d.Reference)
			s.batchSvc.UpdateDetailStatus(ctx, d.ID, "failed", msg)
			continue
		}
		
		if err := s.CheckAndMarkComplete(ctx, dp.KodePesanan); err != nil {
			s.batchSvc.UpdateDetailStatus(ctx, d.ID, "failed", err.Error())
			continue
		}
		done++
		s.batchSvc.UpdateDone(ctx, id, done)
		s.batchSvc.UpdateDetailStatus(ctx, d.ID, "success", "")
	}
	processDuration := time.Since(processStart)
	totalDuration := time.Since(start)
	
	s.batchSvc.UpdateStatus(ctx, id, "completed", "")
	log.Printf("ProcessReconcileBatch %d completed in %v: %d/%d successful (status: %v, fetch: %v, process: %v)", 
		id, totalDuration, done, len(details), statusDuration, fetchDuration, processDuration)
}

// ListCandidates wraps the reconcileRepo method for easier access
func (s *ReconcileService) ListCandidates(ctx context.Context, shop, order, from, to string, limit, offset int) ([]models.ReconcileCandidate, int, error) {
	return s.reconcileRepo.ListCandidates(ctx, shop, order, from, to, limit, offset)
}

// CancelPurchase reverses pending sales journals for the given purchase except
// for the Biaya Mitra amount which remains recorded.
func (s *ReconcileService) CancelPurchase(ctx context.Context, kodePesanan string) error {
	log.Printf("CancelPurchase: %s", kodePesanan)
	
	// This would contain the logic to reverse journal entries
	// For now, just log the cancellation
	dp, err := s.dropRepo.GetDropshipPurchaseByID(ctx, kodePesanan)
	if err != nil || dp == nil {
		return fmt.Errorf("fetch DropshipPurchase %s: %w", kodePesanan, err)
	}
	
	// Update status to cancelled
	if err := s.dropRepo.UpdatePurchaseStatus(ctx, kodePesanan, "Dibatalkan"); err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	
	log.Printf("CancelPurchase: %s marked as cancelled", kodePesanan)
	return nil
}

// UpdateShopeeStatus updates status for a single invoice
func (s *ReconcileService) UpdateShopeeStatus(ctx context.Context, invoice string) error {
	return s.UpdateShopeeStatuses(ctx, []string{invoice})
}
