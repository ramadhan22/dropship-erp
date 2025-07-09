package service

import (
	"context"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// OrderDetailRepo defines the subset of repository methods used by the service.
type OrderDetailRepo interface {
	ListOrderDetails(ctx context.Context, store, order string, limit, offset int) ([]models.ShopeeOrderDetailRow, int, error)
	GetOrderDetail(ctx context.Context, sn string) (*models.ShopeeOrderDetailRow, []models.ShopeeOrderItemRow, []models.ShopeeOrderPackageRow, error)
}

// OrderDetailService exposes listing and retrieval of stored Shopee order details.
type OrderDetailService struct {
	repo OrderDetailRepo
}

func NewOrderDetailService(r OrderDetailRepo) *OrderDetailService {
	return &OrderDetailService{repo: r}
}

func (s *OrderDetailService) ListOrderDetails(ctx context.Context, store, order string, limit, offset int) ([]models.ShopeeOrderDetailRow, int, error) {
	return s.repo.ListOrderDetails(ctx, store, order, limit, offset)
}

func (s *OrderDetailService) GetOrderDetail(ctx context.Context, sn string) (*models.ShopeeOrderDetailRow, []models.ShopeeOrderItemRow, []models.ShopeeOrderPackageRow, error) {
	return s.repo.GetOrderDetail(ctx, sn)
}
