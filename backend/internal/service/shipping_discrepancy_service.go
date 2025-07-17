package service

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
)

type ShippingDiscrepancyService struct {
	db       *sqlx.DB
	discRepo *repository.ShippingDiscrepancyRepo
}

func NewShippingDiscrepancyService(db *sqlx.DB, dr *repository.ShippingDiscrepancyRepo) *ShippingDiscrepancyService {
	return &ShippingDiscrepancyService{db: db, discRepo: dr}
}

// GetShippingDiscrepancies retrieves shipping discrepancies with pagination and optional filtering
func (s *ShippingDiscrepancyService) GetShippingDiscrepancies(
	ctx context.Context,
	storeName string,
	discrepancyType string,
	limit, offset int,
) ([]models.ShippingDiscrepancy, error) {
	if storeName != "" {
		return s.discRepo.GetShippingDiscrepanciesByStore(ctx, storeName, limit, offset)
	}
	if discrepancyType != "" {
		return s.discRepo.GetShippingDiscrepanciesByType(ctx, discrepancyType, limit, offset)
	}
	return s.discRepo.GetAllShippingDiscrepancies(ctx, limit, offset)
}

// GetShippingDiscrepancyStats retrieves shipping discrepancy statistics for a date range
func (s *ShippingDiscrepancyService) GetShippingDiscrepancyStats(
	ctx context.Context,
	startDate, endDate time.Time,
) (map[string]int, error) {
	return s.discRepo.CountShippingDiscrepanciesByDateRange(ctx, startDate, endDate)
}

// GetShippingDiscrepancyByInvoice retrieves a shipping discrepancy by invoice number
func (s *ShippingDiscrepancyService) GetShippingDiscrepancyByInvoice(
	ctx context.Context,
	invoiceNumber string,
) (*models.ShippingDiscrepancy, error) {
	return s.discRepo.GetShippingDiscrepancyByInvoice(ctx, invoiceNumber)
}