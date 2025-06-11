// File: backend/internal/service/shopee_service.go

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

// ShopeeRepoInterface defines the subset of ShopeeRepo & DropshipRepo methods needed.
type ShopeeRepoInterface interface {
	InsertShopeeOrder(ctx context.Context, o *models.ShopeeSettledOrder) error
}

// ShopeeService handles CSV import of settled Shopee orders and links to dropship purchases.
type ShopeeService struct {
	shopeeRepo ShopeeRepoInterface
}

// NewShopeeService constructs a ShopeeService with the given repos.
func NewShopeeService(sr ShopeeRepoInterface) *ShopeeService {
	return &ShopeeService{shopeeRepo: sr}
}

// ImportSettledOrdersCSV reads a Shopee CSV and:
//  1. inserts each settled order into shopee_settled_orders,
//  2. if the CSV includes a purchase_id at column 8, updates that DropshipPurchase.OrderID.
func (s *ShopeeService) ImportSettledOrdersCSV(ctx context.Context, filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open CSV: %w", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	// Skip header
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
		// Parse numeric/string fields
		netIncome, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			return fmt.Errorf("parse net_income '%s': %w", record[1], err)
		}
		serviceFee, _ := strconv.ParseFloat(record[2], 64)
		campaignFee, _ := strconv.ParseFloat(record[3], 64)
		creditCardFee, _ := strconv.ParseFloat(record[4], 64)
		shippingSubsidy, _ := strconv.ParseFloat(record[5], 64)
		taxImportFee, _ := strconv.ParseFloat(record[6], 64)
		settledDate, err := time.Parse("2006-01-02", record[7])
		if err != nil {
			return fmt.Errorf("parse settled_date '%s': %w", record[7], err)
		}

		// purchase_id at column 8 is ignored with new schema

		so := &models.ShopeeSettledOrder{
			OrderID:         record[0],
			NetIncome:       netIncome,
			ServiceFee:      serviceFee,
			CampaignFee:     campaignFee,
			CreditCardFee:   creditCardFee,
			ShippingSubsidy: shippingSubsidy,
			TaxImportFee:    taxImportFee,
			SettledDate:     settledDate,
			SellerUsername:  record[9],
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}
		if err := s.shopeeRepo.InsertShopeeOrder(ctx, so); err != nil {
			return fmt.Errorf("insert order %s: %w", so.OrderID, err)
		}

		// linking to dropship purchases is no longer performed
	}
	return nil
}
