package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/logutil"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// ShopeeDetailBackgroundService handles background fetching of Shopee order details
type ShopeeDetailBackgroundService struct {
	reconcileService *ReconcileService
	batchService     ReconcileServiceBatchSvc
	detailRepo       ReconcileServiceDetailRepo
	dropRepo         ReconcileServiceDropshipRepo
	storeRepo        ReconcileServiceStoreRepo
	client           *ShopeeClient
}

// NewShopeeDetailBackgroundService creates a new background service for Shopee detail fetching
func NewShopeeDetailBackgroundService(
	rs *ReconcileService,
	bs ReconcileServiceBatchSvc,
	dr ReconcileServiceDetailRepo,
	drp ReconcileServiceDropshipRepo,
	sr ReconcileServiceStoreRepo,
	client *ShopeeClient,
) *ShopeeDetailBackgroundService {
	return &ShopeeDetailBackgroundService{
		reconcileService: rs,
		batchService:     bs,
		detailRepo:       dr,
		dropRepo:         drp,
		storeRepo:        sr,
		client:           client,
	}
}

// QueueOrderDetailFetch queues a background job to fetch Shopee order detail for an invoice
func (s *ShopeeDetailBackgroundService) QueueOrderDetailFetch(ctx context.Context, invoice string) (int64, error) {
	log.Printf("Queueing background fetch for order detail: %s", invoice)

	// Create a batch job for this fetch operation
	batch := &models.BatchHistory{
		ProcessType:  "shopee_order_detail_fetch",
		StartedAt:    time.Now(),
		TotalData:    1,
		DoneData:     0,
		Status:       "pending",
		ErrorMessage: "",
		FileName:     "",
		FilePath:     "",
	}

	batchID, err := s.batchService.Create(ctx, batch)
	if err != nil {
		return 0, fmt.Errorf("create batch job: %w", err)
	}

	// Create batch detail for tracking this specific invoice
	detail := &models.BatchHistoryDetail{
		BatchID:   batchID,
		Reference: invoice,
		Store:     "", // Will be filled when processing
		Status:    "pending",
		ErrorMsg:  "",
	}

	if err := s.batchService.CreateDetail(ctx, detail); err != nil {
		return 0, fmt.Errorf("create batch detail: %w", err)
	}

	log.Printf("Queued background fetch job %d for invoice: %s", batchID, invoice)
	return batchID, nil
}

// ProcessPendingOrderDetailFetches processes pending order detail fetch jobs
func (s *ShopeeDetailBackgroundService) ProcessPendingOrderDetailFetches(ctx context.Context) error {
	log.Printf("Processing pending order detail fetch jobs...")

	// Get pending batches
	pendingBatches, err := s.batchService.ListPendingByType(ctx, "shopee_order_detail_fetch")
	if err != nil {
		return fmt.Errorf("list pending batches: %w", err)
	}

	for _, batch := range pendingBatches {
		if err := s.processBatch(ctx, batch.ID); err != nil {
			logutil.Errorf("Failed to process batch %d: %v", batch.ID, err)
			continue
		}
	}

	return nil
}

// processBatch processes a single batch of order detail fetches
func (s *ShopeeDetailBackgroundService) processBatch(ctx context.Context, batchID int64) error {
	log.Printf("Processing order detail fetch batch: %d", batchID)

	// Update batch status to processing
	if err := s.batchService.UpdateStatus(ctx, batchID, "processing", ""); err != nil {
		return fmt.Errorf("update batch status: %w", err)
	}

	// Get batch details (invoices to process)
	details, err := s.batchService.ListDetails(ctx, batchID)
	if err != nil {
		return fmt.Errorf("list batch details: %w", err)
	}

	successCount := 0
	for _, detail := range details {
		if detail.Status != "pending" {
			continue // Skip already processed items
		}

		invoice := detail.Reference
		log.Printf("Fetching order detail for invoice: %s", invoice)

		// Attempt to fetch order detail from Shopee
		if err := s.fetchAndStoreOrderDetail(ctx, invoice, detail.ID); err != nil {
			logutil.Errorf("Failed to fetch order detail for %s: %v", invoice, err)
			s.batchService.UpdateDetailStatus(ctx, detail.ID, "failed", err.Error())
			continue
		}

		// Mark detail as completed
		if err := s.batchService.UpdateDetailStatus(ctx, detail.ID, "completed", ""); err != nil {
			logutil.Errorf("Failed to update detail status for %s: %v", invoice, err)
		}
		successCount++
	}

	// Update batch completion
	if err := s.batchService.UpdateDone(ctx, batchID, successCount); err != nil {
		logutil.Errorf("Failed to update batch done count: %v", err)
	}

	status := "completed"
	if successCount == 0 {
		status = "failed"
	}
	if err := s.batchService.UpdateStatus(ctx, batchID, status, ""); err != nil {
		logutil.Errorf("Failed to update batch final status: %v", err)
	}

	log.Printf("Completed processing batch %d: %d/%d successful", batchID, successCount, len(details))
	return nil
}

// fetchAndStoreOrderDetail fetches order detail from Shopee API and stores it in the database
func (s *ShopeeDetailBackgroundService) fetchAndStoreOrderDetail(ctx context.Context, invoice string, detailID int64) error {
	// First get dropship purchase to determine store
	dp, err := s.dropRepo.GetDropshipPurchaseByInvoice(ctx, invoice)
	if err != nil || dp == nil {
		return fmt.Errorf("get dropship purchase for invoice %s: %w", invoice, err)
	}

	// Get store information
	store, err := s.storeRepo.GetStoreByName(ctx, dp.NamaToko)
	if err != nil || store == nil {
		return fmt.Errorf("get store %s: %w", dp.NamaToko, err)
	}

	// Update batch detail with store name
	s.batchService.UpdateDetailStatus(ctx, detailID, "processing", "")

	// Ensure store token is valid
	if err := s.reconcileService.ensureStoreTokenValid(ctx, store); err != nil {
		return fmt.Errorf("ensure store token valid: %w", err)
	}

	// Fetch order detail from Shopee API
	orderSN := dp.KodeInvoiceChannel
	detail, err := s.client.FetchShopeeOrderDetail(ctx, *store.AccessToken, *store.ShopID, orderSN)
	if err != nil {
		return fmt.Errorf("fetch order detail from Shopee API: %w", err)
	}

	// Convert and save to database
	if err := s.saveOrderDetailToDatabase(ctx, orderSN, dp.NamaToko, detail); err != nil {
		return fmt.Errorf("save order detail to database: %w", err)
	}

	log.Printf("Successfully fetched and stored order detail for invoice: %s", invoice)
	return nil
}

// saveOrderDetailToDatabase converts and saves the Shopee API response to database
func (s *ShopeeDetailBackgroundService) saveOrderDetailToDatabase(ctx context.Context, orderSN, storeName string, detail *ShopeeOrderDetail) error {
	// Convert the ShopeeOrderDetail to database models
	// This logic should be similar to what's in order_detail_convert.go
	detailRow, items, packages, err := s.convertShopeeDetailToModels(orderSN, storeName, detail)
	if err != nil {
		return fmt.Errorf("convert detail to models: %w", err)
	}

	// Save to database
	if err := s.detailRepo.SaveOrderDetail(ctx, detailRow, items, packages); err != nil {
		return fmt.Errorf("save to database: %w", err)
	}

	return nil
}

// convertShopeeDetailToModels converts ShopeeOrderDetail to database models
func (s *ShopeeDetailBackgroundService) convertShopeeDetailToModels(orderSN, storeName string, detail *ShopeeOrderDetail) (*models.ShopeeOrderDetailRow, []models.ShopeeOrderItemRow, []models.ShopeeOrderPackageRow, error) {
	detailMap := map[string]any(*detail)

	// Create detail row
	detailRow := &models.ShopeeOrderDetailRow{
		OrderSN:   orderSN,
		NamaToko:  storeName,
		CreatedAt: time.Now(),
	}

	// Map common fields
	if status, ok := detailMap["order_status"]; ok {
		if statusStr, ok := status.(string); ok {
			detailRow.OrderStatus = &statusStr
		}
	}

	if totalAmount, ok := detailMap["total_amount"]; ok {
		if amount, ok := totalAmount.(float64); ok {
			detailRow.TotalAmount = &amount
		}
	}

	if currency, ok := detailMap["currency"]; ok {
		if currencyStr, ok := currency.(string); ok {
			detailRow.Currency = &currencyStr
		}
	}

	// Handle time fields
	if checkoutTime, ok := detailMap["checkout_time"]; ok {
		if checkoutInt, ok := checkoutTime.(float64); ok {
			t := time.Unix(int64(checkoutInt), 0)
			detailRow.CheckoutTime = &t
		}
	}

	if updateTime, ok := detailMap["update_time"]; ok {
		if updateInt, ok := updateTime.(float64); ok {
			t := time.Unix(int64(updateInt), 0)
			detailRow.UpdateTime = &t
		}
	}

	// Handle item list
	var items []models.ShopeeOrderItemRow
	if itemList, ok := detailMap["item_list"]; ok {
		if itemArray, ok := itemList.([]interface{}); ok {
			for _, item := range itemArray {
				if itemMap, ok := item.(map[string]interface{}); ok {
					itemRow := models.ShopeeOrderItemRow{
						OrderSN: orderSN,
					}

					if itemName, ok := itemMap["item_name"].(string); ok {
						itemRow.ItemName = &itemName
					}
					if modelSku, ok := itemMap["model_sku"].(string); ok {
						itemRow.ModelSKU = &modelSku
					}
					if qty, ok := itemMap["model_quantity_purchased"].(float64); ok {
						qtyInt := int(qty)
						itemRow.ModelQuantityPurchased = &qtyInt
					}
					if origPrice, ok := itemMap["model_original_price"].(float64); ok {
						itemRow.ModelOriginalPrice = &origPrice
					}
					if discPrice, ok := itemMap["model_discounted_price"].(float64); ok {
						itemRow.ModelDiscountedPrice = &discPrice
					}

					items = append(items, itemRow)
				}
			}
		}
	}

	// Handle package list (shipping info)
	var packages []models.ShopeeOrderPackageRow
	if packageList, ok := detailMap["package_list"]; ok {
		if packageArray, ok := packageList.([]interface{}); ok {
			for _, pkg := range packageArray {
				if pkgMap, ok := pkg.(map[string]interface{}); ok {
					packageRow := models.ShopeeOrderPackageRow{
						OrderSN: orderSN,
					}

					if logisticsStatus, ok := pkgMap["logistics_status"].(string); ok {
						packageRow.LogisticsStatus = &logisticsStatus
					}
					if shippingCarrier, ok := pkgMap["shipping_carrier"].(string); ok {
						packageRow.ShippingCarrier = &shippingCarrier
					}

					packages = append(packages, packageRow)
				}
			}
		}
	}

	return detailRow, items, packages, nil
}
