package handlers

import (
	"context"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// ShopeeServiceInterface defines only the method the handler needs.
type ShopeeServiceInterface interface {
	ImportSettledOrdersXLSX(ctx context.Context, r io.Reader) (int, error)
	ImportAffiliateCSV(ctx context.Context, r io.Reader) (int, error)
	ListSettled(ctx context.Context, channel, store, from, to, orderNo, sortBy, dir string, limit, offset int) ([]models.ShopeeSettled, int, error)
	SumShopeeSettled(ctx context.Context, channel, store, from, to string) (*models.ShopeeSummary, error)
	ListAffiliate(ctx context.Context, noPesanan, from, to string, limit, offset int) ([]models.ShopeeAffiliateSale, int, error)
	SumAffiliate(ctx context.Context, noPesanan, from, to string) (*models.ShopeeAffiliateSummary, error)
	ListSalesProfit(ctx context.Context, channel, store, from, to, orderNo, sortBy, dir string, limit, offset int) ([]models.SalesProfit, int, error)
}

type ShopeeHandler struct {
	svc ShopeeServiceInterface
}

// Now accepts any ShopeeServiceInterface
func NewShopeeHandler(svc ShopeeServiceInterface) *ShopeeHandler {
	return &ShopeeHandler{svc: svc}
}

func (h *ShopeeHandler) HandleImport(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	files := form.File["file"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	ctx := c.Request.Context()
	total := 0
	for _, fh := range files {
		f, err := fh.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		count, err := h.svc.ImportSettledOrdersXLSX(ctx, f)
		f.Close()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		total += count
	}
	c.JSON(http.StatusOK, gin.H{"inserted": total})
}

func (h *ShopeeHandler) HandleImportAffiliate(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	files := form.File["file"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	ctx := c.Request.Context()
	total := 0
	for _, fh := range files {
		f, err := fh.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		count, err := h.svc.ImportAffiliateCSV(ctx, f)
		f.Close()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		total += count
	}
	c.JSON(http.StatusOK, gin.H{"inserted": total})
}

// HandleListSettled returns paginated shopee settled data with optional filters.
func (h *ShopeeHandler) HandleListSettled(c *gin.Context) {
	channel := c.Query("channel")
	store := c.Query("store")
	from := c.Query("from")
	to := c.Query("to")
	orderNo := c.Query("order")
	sortBy := c.DefaultQuery("sort", "waktu_pesanan_dibuat")
	dir := c.DefaultQuery("dir", "desc")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (page - 1) * size
	ctx := c.Request.Context()
	list, total, err := h.svc.ListSettled(ctx, channel, store, from, to, orderNo, sortBy, dir, size, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list, "total": total})
}

// HandleSumSettled returns the total penerimaan for all filtered rows.
func (h *ShopeeHandler) HandleSumSettled(c *gin.Context) {
	channel := c.Query("channel")
	store := c.Query("store")
	from := c.Query("from")
	to := c.Query("to")
	ctx := c.Request.Context()
	sum, err := h.svc.SumShopeeSettled(ctx, channel, store, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sum)
}

// HandleListAffiliate returns paginated affiliate sales data with optional filters.
func (h *ShopeeHandler) HandleListAffiliate(c *gin.Context) {
	noPesanan := c.Query("no_pesanan")
	from := c.Query("from")
	to := c.Query("to")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (page - 1) * size
	ctx := c.Request.Context()
	list, total, err := h.svc.ListAffiliate(ctx, noPesanan, from, to, size, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list, "total": total})
}

// HandleSumAffiliate returns total values for filtered affiliate rows.
func (h *ShopeeHandler) HandleSumAffiliate(c *gin.Context) {
	noPesanan := c.Query("no_pesanan")
	from := c.Query("from")
	to := c.Query("to")
	ctx := c.Request.Context()
	sum, err := h.svc.SumAffiliate(ctx, noPesanan, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sum)
}

// HandleListSalesProfit returns sales profit rows with pagination and filters.
func (h *ShopeeHandler) HandleListSalesProfit(c *gin.Context) {
	channel := c.Query("channel")
	store := c.Query("store")
	from := c.Query("from")
	to := c.Query("to")
	orderNo := c.Query("order")
	sortBy := c.DefaultQuery("sort", "tanggal_pesanan")
	dir := c.DefaultQuery("dir", "desc")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (page - 1) * size
	ctx := c.Request.Context()
	list, total, err := h.svc.ListSalesProfit(ctx, channel, store, from, to, orderNo, sortBy, dir, size, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list, "total": total})
}
