package handlers

import (
	"context"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// ShopeeServiceInterface defines only the method the handler needs.
type ShopeeServiceInterface interface {
	ImportSettledOrdersXLSX(ctx context.Context, r io.Reader) (int, []string, error)
	ImportAffiliateCSV(ctx context.Context, r io.Reader) (int, error)
	ListSettled(ctx context.Context, channel, store, from, to, orderNo, sortBy, dir string, limit, offset int) ([]models.ShopeeSettled, int, error)
	SumShopeeSettled(ctx context.Context, channel, store, from, to string) (*models.ShopeeSummary, error)
	ListAffiliate(ctx context.Context, noPesanan, from, to string, limit, offset int) ([]models.ShopeeAffiliateSale, int, error)
	SumAffiliate(ctx context.Context, noPesanan, from, to string) (*models.ShopeeAffiliateSummary, error)
	ListSalesProfit(ctx context.Context, channel, store, from, to, orderNo, sortBy, dir string, limit, offset int) ([]models.SalesProfit, int, error)
	ConfirmSettle(ctx context.Context, orderSN string) error
	GetSettleDetail(ctx context.Context, orderSN string) (*models.ShopeeSettled, float64, error)
	GetReturnList(ctx context.Context, storeFilter, pageNo, pageSize, createTimeFrom, createTimeTo, updateTimeFrom, updateTimeTo, status, negotiationStatus, sellerProofStatus, sellerCompensationStatus string) ([]models.ShopeeOrderReturn, bool, error)
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
	mismatches := []string{}
	for i, fh := range files {
		log.Printf("HandleImport processing file %d of %d: %s", i+1, len(files), fh.Filename)
		f, err := fh.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		count, mis, err := h.svc.ImportSettledOrdersXLSX(ctx, f)
		f.Close()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		log.Printf("HandleImport finished file %d of %d: %s inserted=%d", i+1, len(files), fh.Filename, count)
		total += count
		mismatches = append(mismatches, mis...)
	}
	c.JSON(http.StatusOK, gin.H{"inserted": total, "mismatches": mismatches})
}

func (h *ShopeeHandler) HandleConfirmSettle(c *gin.Context) {
	sn := c.Param("order_sn")
	if err := h.svc.ConfirmSettle(c.Request.Context(), sn); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *ShopeeHandler) HandleGetSettleDetail(c *gin.Context) {
	sn := c.Param("order_sn")
	data, sum, err := h.svc.GetSettleDetail(c.Request.Context(), sn)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data, "dropship_total": sum})
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
	for i, fh := range files {
		log.Printf("HandleImportAffiliate processing file %d of %d: %s", i+1, len(files), fh.Filename)
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
		log.Printf("HandleImportAffiliate finished file %d of %d: %s inserted=%d", i+1, len(files), fh.Filename, count)
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

func (h *ShopeeHandler) HandleGetReturnList(c *gin.Context) {
	storeFilter := c.Query("store")
	pageNo := c.Query("page_no")
	pageSize := c.Query("page_size")
	createTimeFrom := c.Query("create_time_from")
	createTimeTo := c.Query("create_time_to")
	updateTimeFrom := c.Query("update_time_from")
	updateTimeTo := c.Query("update_time_to")
	status := c.Query("status")
	negotiationStatus := c.Query("negotiation_status")
	sellerProofStatus := c.Query("seller_proof_status")
	sellerCompensationStatus := c.Query("seller_compensation_status")

	ctx := c.Request.Context()
	returns, hasMore, err := h.svc.GetReturnList(ctx, storeFilter, pageNo, pageSize, createTimeFrom, createTimeTo, updateTimeFrom, updateTimeTo, status, negotiationStatus, sellerProofStatus, sellerCompensationStatus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":     returns,
		"has_more": hasMore,
		"total":    len(returns),
	})
}
