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

// ReconcileService handles reconciliation operations between dropship purchases and Shopee escrow settlements.
type ReconcileService struct {
	db            *sqlx.DB
	reconcileRepo *repository.ReconcileRepo
	shopeeService *ShopeeService
}

// NewReconcileService creates a new reconcile service.
func NewReconcileService(db *sqlx.DB, reconcileRepo *repository.ReconcileRepo, shopeeService *ShopeeService) *ReconcileService {
	return &ReconcileService{
		db:            db,
		reconcileRepo: reconcileRepo,
		shopeeService: shopeeService,
	}
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
