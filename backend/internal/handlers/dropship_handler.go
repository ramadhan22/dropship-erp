package handlers

import (
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
	"github.com/ramadhan22/dropship-erp/backend/internal/service"
)

// DropshipServiceInterface defines only the method the handler needs.
type DropshipServiceInterface interface {
	ImportFromCSV(ctx context.Context, r io.Reader, channel string) (int, error)
	ListDropshipPurchases(ctx context.Context, channel, store, from, to, orderNo, sortBy, dir string, limit, offset int) ([]models.DropshipPurchase, int, error)
	SumDropshipPurchases(ctx context.Context, channel, store, from, to string) (float64, error)
	GetDropshipPurchaseByID(ctx context.Context, kodePesanan string) (*models.DropshipPurchase, error)
	ListDropshipPurchaseDetails(ctx context.Context, kodePesanan string) ([]models.DropshipPurchaseDetail, error)
	TopProducts(ctx context.Context, channel, store, from, to string, limit int) ([]models.ProductSales, error)
	DailyTotals(ctx context.Context, channel, store, from, to string) ([]repository.DailyPurchaseTotal, error)
	MonthlyTotals(ctx context.Context, channel, store, from, to string) ([]repository.MonthlyPurchaseTotal, error)
	CancelledSummary(ctx context.Context, channel, store, from, to string) (repository.CancelledSummary, error)
}

type DropshipHandler struct {
	svc   DropshipServiceInterface
	batch *service.BatchService
}

// Now accepts any DropshipServiceInterface
func NewDropshipHandler(svc DropshipServiceInterface, batch *service.BatchService) *DropshipHandler {
	return &DropshipHandler{svc: svc, batch: batch}
}

func (h *DropshipHandler) HandleImport(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	channel := c.PostForm("channel")

	var id int64
	if h.batch != nil {
		batch := &models.BatchHistory{ProcessType: "dropship_import", TotalData: 1, DoneData: 0}
		var err error
		id, err = h.batch.Create(context.Background(), batch)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	go func(fh *multipart.FileHeader, ch string, batchID int64) {
		f, err := fh.Open()
		if err != nil {
			if h.batch != nil {
				h.batch.UpdateStatus(context.Background(), batchID, "failed", err.Error())
			}
			return
		}
		defer f.Close()
		if _, err := h.svc.ImportFromCSV(context.Background(), f, ch); err != nil {
			if h.batch != nil {
				h.batch.UpdateStatus(context.Background(), batchID, "failed", err.Error())
			}
			return
		}
		if h.batch != nil {
			h.batch.UpdateDone(context.Background(), batchID, 1)
			h.batch.UpdateStatus(context.Background(), batchID, "completed", "")
		}
	}(fileHeader, channel, id)

	c.JSON(http.StatusOK, gin.H{"message": "processing in background", "batch_id": id})
}

// HandleList returns dropship purchases with optional filters and pagination.
func (h *DropshipHandler) HandleList(c *gin.Context) {
	channel := c.Query("channel")
	store := c.Query("store")
	from := c.Query("from")
	to := c.Query("to")
	orderNo := c.Query("order")
	sortBy := c.DefaultQuery("sort", "waktu_pesanan_terbuat")
	dir := c.DefaultQuery("dir", "desc")
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

	list, total, err := h.svc.ListDropshipPurchases(context.Background(), channel, store, from, to, orderNo, sortBy, dir, limit, offset)
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
	from := c.Query("from")
	to := c.Query("to")
	ctx := c.Request.Context()
	sum, err := h.svc.SumDropshipPurchases(ctx, channel, store, from, to)
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

// HandleDailyTotals returns daily aggregated purchase totals and counts.
func (h *DropshipHandler) HandleDailyTotals(c *gin.Context) {
	channel := c.Query("channel")
	store := c.Query("store")
	from := c.Query("from")
	to := c.Query("to")
	res, err := h.svc.DailyTotals(c.Request.Context(), channel, store, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// HandleMonthlyTotals returns monthly aggregated purchase totals and counts.
func (h *DropshipHandler) HandleMonthlyTotals(c *gin.Context) {
	channel := c.Query("channel")
	store := c.Query("store")
	from := c.Query("from")
	to := c.Query("to")
	res, err := h.svc.MonthlyTotals(c.Request.Context(), channel, store, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// HandleCancelledSummary returns cancellation count and Biaya Mitra totals.
func (h *DropshipHandler) HandleCancelledSummary(c *gin.Context) {
	channel := c.Query("channel")
	store := c.Query("store")
	from := c.Query("from")
	to := c.Query("to")
	res, err := h.svc.CancelledSummary(c.Request.Context(), channel, store, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}
