package handlers

import (
	"context"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// DropshipServiceInterface defines only the method the handler needs.
type DropshipServiceInterface interface {
	ImportFromCSV(ctx context.Context, r io.Reader) (int, error)
	ListDropshipPurchases(ctx context.Context, channel, store, date, month, year string, limit, offset int) ([]models.DropshipPurchase, int, error)
	SumDropshipPurchases(ctx context.Context, channel, store, date, month, year string) (float64, error)
	GetDropshipPurchaseByID(ctx context.Context, kodePesanan string) (*models.DropshipPurchase, error)
	ListDropshipPurchaseDetails(ctx context.Context, kodePesanan string) ([]models.DropshipPurchaseDetail, error)
	TopProducts(ctx context.Context, channel, store, from, to string, limit int) ([]models.ProductSales, error)
}

type DropshipHandler struct {
	svc DropshipServiceInterface
}

// Now accepts any DropshipServiceInterface
func NewDropshipHandler(svc DropshipServiceInterface) *DropshipHandler {
	return &DropshipHandler{svc: svc}
}

func (h *DropshipHandler) HandleImport(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	f, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer f.Close()

	count, err := h.svc.ImportFromCSV(context.Background(), f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"inserted": count})
}

// HandleList returns dropship purchases with optional filters and pagination.
func (h *DropshipHandler) HandleList(c *gin.Context) {
	channel := c.Query("channel")
	store := c.Query("store")
	date := c.Query("date")
	month := c.Query("month")
	year := c.Query("year")
	pageStr := c.DefaultQuery("page", "1")
	sizeStr := c.DefaultQuery("page_size", "10")
	page, _ := strconv.Atoi(pageStr)
	size, _ := strconv.Atoi(sizeStr)
	if page < 1 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	limit := size
	offset := (page - 1) * size

	list, total, err := h.svc.ListDropshipPurchases(context.Background(), channel, store, date, month, year, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list, "total": total})
}

// HandleSum returns the sum of total_transaksi for all data matching filters.
func (h *DropshipHandler) HandleSum(c *gin.Context) {
	channel := c.Query("channel")
	store := c.Query("store")
	date := c.Query("date")
	month := c.Query("month")
	year := c.Query("year")
	ctx := c.Request.Context()
	sum, err := h.svc.SumDropshipPurchases(ctx, channel, store, date, month, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"total": sum})
}

// HandleListDetails returns detail rows for a given kode_pesanan.
func (h *DropshipHandler) HandleListDetails(c *gin.Context) {
	kode := c.Param("id")
	details, err := h.svc.ListDropshipPurchaseDetails(context.Background(), kode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, details)
}

// HandleTopProducts returns aggregated sales by product.
func (h *DropshipHandler) HandleTopProducts(c *gin.Context) {
	channel := c.Query("channel")
	store := c.Query("store")
	from := c.Query("from")
	to := c.Query("to")
	limitStr := c.DefaultQuery("limit", "5")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 {
		limit = 5
	}
	res, err := h.svc.TopProducts(context.Background(), channel, store, from, to, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}
