package handlers

import (
	"context"
	"encoding/csv"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
	"github.com/ramadhan22/dropship-erp/backend/internal/service"
)

// DropshipServiceInterface defines only the method the handler needs.
type DropshipServiceInterface interface {
	ImportFromCSV(ctx context.Context, r io.Reader, channel string, batchID int64) (int, error)
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
	channel := c.PostForm("channel")
	queued := len(files)
	for _, fh := range files {
		dir := filepath.Join("backend", "uploads", "dropship")
		os.MkdirAll(dir, 0o755)
		filename := time.Now().Format("20060102150405") + "_" + fh.Filename
		path := filepath.Join(dir, filename)
		if err := c.SaveUploadedFile(fh, path); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		var id int64
		if h.batch != nil {
			batch := &models.BatchHistory{ProcessType: "dropship_import", TotalData: 0, DoneData: 0, Status: "pending", FileName: fh.Filename, FilePath: path}
			var err error
			id, err = h.batch.Create(context.Background(), batch)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
		go h.processFile(id, path, channel)
	}
	c.JSON(http.StatusOK, gin.H{"queued": queued})
}

func (h *DropshipHandler) processFile(batchID int64, path, channel string) {
	ctx := context.Background()
	f, err := os.Open(path)
	if err != nil {
		if h.batch != nil {
			h.batch.UpdateStatus(ctx, batchID, "failed", err.Error())
		}
		return
	}
	var total int
	if h.batch != nil {
		total, err = countCSVRows(f)
		if err != nil {
			f.Close()
			h.batch.UpdateStatus(ctx, batchID, "failed", err.Error())
			return
		}
		h.batch.UpdateTotal(ctx, batchID, total)
		h.batch.UpdateStatus(ctx, batchID, "processing", "")
		if _, err := f.Seek(0, io.SeekStart); err != nil {
			f.Close()
			h.batch.UpdateStatus(ctx, batchID, "failed", err.Error())
			return
		}
	}
	count, err := h.svc.ImportFromCSV(ctx, f, channel, batchID)
	f.Close()
	if h.batch != nil {
		h.batch.UpdateDone(ctx, batchID, count)
		if err != nil {
			h.batch.UpdateStatus(ctx, batchID, "failed", err.Error())
		} else {
			h.batch.UpdateStatus(ctx, batchID, "completed", "")
		}
	}
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

// countCSVRows returns the number of data rows in a CSV reader.
// It expects the first line to be a header and ignores it.
func countCSVRows(r io.Reader) (int, error) {
	reader := csv.NewReader(r)
	if _, err := reader.Read(); err != nil {
		return 0, err
	}
	n := 0
	for {
		if _, err := reader.Read(); err == io.EOF {
			break
		} else if err != nil {
			return n, err
		} else {
			n++
		}
	}
	return n, nil
}
