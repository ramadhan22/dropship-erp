package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

type ReconcileExtraService interface {
	ListUnmatched(ctx context.Context, shop string) ([]models.ReconciledTransaction, error)
	ListCandidates(ctx context.Context, shop, order string) ([]models.ReconcileCandidate, error)
	BulkReconcile(ctx context.Context, pairs [][2]string, shop string) error
	CheckAndMarkComplete(ctx context.Context, kodePesanan string) error
	GetShopeeOrderStatus(ctx context.Context, invoice string) (string, error)
	CancelPurchase(ctx context.Context, kodePesanan string) error
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
	grp.POST("/check", h.check)
	grp.POST("/cancel", h.cancel)
	grp.GET("/status", h.status)
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
	res, err := h.svc.ListCandidates(context.Background(), shop, order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
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
	status, err := h.svc.GetShopeeOrderStatus(context.Background(), invoice)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": status})
}
