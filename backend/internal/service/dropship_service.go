// File: backend/internal/service/dropship_service.go

package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// DropshipRepoInterface defines the subset of DropshipRepo methods that the service needs.
// In production, you pass in *repository.DropshipRepo; in tests you pass a fake implementing this.
type DropshipRepoInterface interface {
	InsertDropshipPurchase(ctx context.Context, p *models.DropshipPurchase) error
}

// DropshipService handles CSV‐import and any Dropship‐related business logic.
type DropshipService struct {
	repo DropshipRepoInterface
}

// NewDropshipService constructs a DropshipService with the given repository.
func NewDropshipService(repo DropshipRepoInterface) *DropshipService {
	return &DropshipService{repo: repo}
}

// ImportFromCSV reads a Dumpsihp CSV file (with a header row) and inserts each purchase row.
// Expected CSV columns (example):
//
//	0: seller_username
//	1: purchase_id
//	2: order_id         (can be empty string if not linked yet)
//	3: sku
//	4: qty
//	5: purchase_price
//	6: purchase_fee
//	7: status
//	8: purchase_date    (YYYY-MM-DD)
//	9: supplier_name
//
// Any parse error aborts the import and returns it.
func (s *DropshipService) ImportFromCSV(ctx context.Context, filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open CSV: %w", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	// First row is header; skip it.
	if _, err := reader.Read(); err != nil {
		return fmt.Errorf("read header: %w", err)
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("read row: %w", err)
		}
		// Parse fields
		quantity, err := strconv.Atoi(record[8])
		if err != nil {
			return fmt.Errorf("parse qty '%s': %w", record[8], err)
		}
		price, err := strconv.ParseFloat(record[7], 64)
		if err != nil {
			return fmt.Errorf("parse purchase_price '%s': %w", record[7], err)
		}
		fee, err := strconv.ParseFloat(record[11], 64)
		if err != nil {
			fee = 0.0
		}
		pDate, err := time.Parse("02 January 2006, 15:04:05", record[1])
		if err != nil {
			return fmt.Errorf("parse purchase_date '%s': %w", record[1], err)
		}

		// order_id might be empty; treat as nil
		var orderID *string
		if rec := record[19]; rec != "" {
			orderID = &rec
		}

		// supplier_name might be empty
		var supplier *string
		if rec := record[9]; rec != "" {
			supplier = &rec
		}

		p := &models.DropshipPurchase{
			SellerUsername: record[0],
			PurchaseID:     record[1],
			OrderID:        orderID,
			SKU:            record[3],
			Quantity:       quantity,
			PurchasePrice:  price,
			PurchaseFee:    fee,
			Status:         record[7],
			PurchaseDate:   pDate,
			SupplierName:   supplier,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		if err := s.repo.InsertDropshipPurchase(ctx, p); err != nil {
			return fmt.Errorf("insert purchase %s: %w", p.PurchaseID, err)
		}
	}
	return nil
}
