package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/ramadhan22/dropship-erp/backend/internal/service"
)

type ReconcileExtraService interface {
	ListUnmatched(ctx context.Context, shop string) ([]models.ReconciledTransaction, error)
	ListCandidates(ctx context.Context, shop, order, status, from, to string, limit, offset int) ([]models.ReconcileCandidate, int, error)
	BulkReconcile(ctx context.Context, pairs [][2]string, shop string) error
	CheckAndMarkComplete(ctx context.Context, kodePesanan string) error
	GetShopeeOrderStatus(ctx context.Context, invoice string) (string, error)
	GetShopeeOrderDetail(ctx context.Context, invoice string) (*service.ShopeeOrderDetail, error)
	GetShopeeOrderDetailCached(ctx context.Context, invoice string) (*service.ShopeeOrderDetail, *int64, error)
	GetShopeeEscrowDetail(ctx context.Context, invoice string) (*service.ShopeeEscrowDetail, error)
	GetShopeeEscrowDetailCached(ctx context.Context, invoice string) (*service.ShopeeEscrowDetail, error)
	GetShopeeAccessToken(ctx context.Context, invoice string) (string, error)
	CancelPurchase(ctx context.Context, kodePesanan string) error
	UpdateShopeeStatus(ctx context.Context, invoice string) error
	UpdateShopeeStatuses(ctx context.Context, invoices []string) error
	CreateReconcileBatches(ctx context.Context, shop, order, status, from, to string) (*models.ReconcileBatchInfo, error)
	GetBackgroundJobStatus(ctx context.Context, batchID int64) (string, error)
}

type ReconcileExtraHandler struct{ svc ReconcileExtraService }

func NewReconcileExtraHandler(s ReconcileExtraService) *ReconcileExtraHandler {
	return &ReconcileExtraHandler{svc: s}
}

func (h *ReconcileExtraHandler) RegisterRoutes(r gin.IRouter) {
	grp := r.Group("/reconcile")
	grp.GET("/unmatched", h.list)
	grp.GET("/candidates", h.candidates)
	grp.POST("/bulk", h.bulk)
	grp.POST("/batch", h.createBatch)
	grp.POST("/check", h.check)
	grp.POST("/cancel", h.cancel)
	grp.POST("/update_status", h.updateStatus)
	grp.POST("/update_statuses", h.updateStatuses)
	grp.GET("/status", h.status)
	grp.GET("/escrow", h.escrow)
	grp.GET("/token", h.token)
	grp.GET("/job-status/:batchId", h.jobStatus)
}

func (h *ReconcileExtraHandler) list(c *gin.Context) {
	shop := c.Query("shop")
	res, err := h.svc.ListUnmatched(context.Background(), shop)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *ReconcileExtraHandler) candidates(c *gin.Context) {
	shop := c.Query("shop")
	order := c.Query("order")
	status := c.Query("status")
	from := c.Query("from")
	to := c.Query("to")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	offset := (page - 1) * size
	list, total, err := h.svc.ListCandidates(context.Background(), shop, order, status, from, to, size, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list, "total": total})
}

func (h *ReconcileExtraHandler) bulk(c *gin.Context) {
	var req struct {
		Pairs [][2]string `json:"pairs"`
		Shop  string      `json:"shop"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.BulkReconcile(context.Background(), req.Pairs, req.Shop); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (h *ReconcileExtraHandler) check(c *gin.Context) {
	var req struct {
		KodePesanan string `json:"kode_pesanan" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := h.svc.CheckAndMarkComplete(context.Background(), req.KodePesanan)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "status updated"})
}

func (h *ReconcileExtraHandler) cancel(c *gin.Context) {
	var req struct {
		KodePesanan string `json:"kode_pesanan" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.CancelPurchase(context.Background(), req.KodePesanan); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (h *ReconcileExtraHandler) status(c *gin.Context) {
	invoice := c.Query("invoice")
	if invoice == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing invoice"})
		return
	}

	// Use cached version first
	detail, batchID, err := h.svc.GetShopeeOrderDetailCached(context.Background(), invoice)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// If we have cached data, return it immediately
	if detail != nil {
		c.JSON(http.StatusOK, detail)
		return
	}

	// If we queued a background job, return job info
	if batchID != nil {
		c.JSON(http.StatusAccepted, gin.H{
			"status":   "processing",
			"batch_id": *batchID,
			"message":  "Order detail is being fetched in the background. Please check back in a moment.",
		})
		return
	}

	// Should not reach here, but fallback
	c.JSON(http.StatusInternalServerError, gin.H{"error": "unexpected error"})
}

func (h *ReconcileExtraHandler) escrow(c *gin.Context) {
	invoice := c.Query("invoice")
	if invoice == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing invoice"})
		return
	}
	detail, err := h.svc.GetShopeeEscrowDetailCached(context.Background(), invoice)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, detail)
}

func (h *ReconcileExtraHandler) token(c *gin.Context) {
	invoice := c.Query("invoice")
	if invoice == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing invoice"})
		return
	}
	tok, err := h.svc.GetShopeeAccessToken(context.Background(), invoice)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"access_token": tok})
}

func (h *ReconcileExtraHandler) updateStatus(c *gin.Context) {
	var req struct {
		Invoice string `json:"invoice" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.UpdateShopeeStatus(context.Background(), req.Invoice); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (h *ReconcileExtraHandler) updateStatuses(c *gin.Context) {
	var req struct {
		Invoices []string `json:"invoices"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.UpdateShopeeStatuses(context.Background(), req.Invoices); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (h *ReconcileExtraHandler) createBatch(c *gin.Context) {
	var req struct {
		Shop   string `json:"shop"`
		Order  string `json:"order"`
		Status string `json:"status"`
		From   string `json:"from"`
		To     string `json:"to"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call the service which will return batch information
	batchInfo, err := h.svc.CreateReconcileBatches(context.Background(), req.Shop, req.Order, req.Status, req.From, req.To)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":            "Reconcile batches created successfully",
		"batches_created":    batchInfo.BatchCount,
		"total_transactions": batchInfo.TotalTransactions,
		"status":             "processing will begin shortly",
	})
}

func (h *ReconcileExtraHandler) jobStatus(c *gin.Context) {
	batchIdStr := c.Param("batchId")
	batchID, err := strconv.ParseInt(batchIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid batch ID"})
		return
	}

	status, err := h.svc.GetBackgroundJobStatus(context.Background(), batchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"batch_id": batchID,
		"status":   status,
	})
}
