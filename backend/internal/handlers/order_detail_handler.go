package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// OrderDetailService defines methods used by the handler.
type OrderDetailService interface {
	ListOrderDetails(ctx context.Context, store, order string, limit, offset int) ([]models.ShopeeOrderDetailRow, int, error)
	GetOrderDetail(ctx context.Context, sn string) (*models.ShopeeOrderDetailRow, []models.ShopeeOrderItemRow, []models.ShopeeOrderPackageRow, error)
}

// OrderDetailHandler handles HTTP requests for stored order details.
type OrderDetailHandler struct{ svc OrderDetailService }

func NewOrderDetailHandler(s OrderDetailService) *OrderDetailHandler {
	return &OrderDetailHandler{svc: s}
}

func (h *OrderDetailHandler) RegisterRoutes(r gin.IRouter) {
	grp := r.Group("/order-details")
	grp.GET("", h.list)
	grp.GET(":sn", h.get)
}

func (h *OrderDetailHandler) list(c *gin.Context) {
	store := c.Query("store")
	order := c.Query("order")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	limit := size
	offset := (page - 1) * size
	list, total, err := h.svc.ListOrderDetails(context.Background(), store, order, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list, "total": total})
}

func (h *OrderDetailHandler) get(c *gin.Context) {
	sn := c.Param("sn")
	det, items, packs, err := h.svc.GetOrderDetail(context.Background(), sn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"detail": det, "items": items, "packages": packs})
}
