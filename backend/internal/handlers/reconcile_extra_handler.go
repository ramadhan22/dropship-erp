package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// ReconcileExtraService defines the interface for additional reconcile operations
type ReconcileExtraService interface {
	CheckAndMarkComplete(ctx context.Context, kodePesanan string) error
	CancelPurchase(ctx context.Context, kodePesanan string) error
	UpdateShopeeStatus(ctx context.Context, invoice string) error
	UpdateShopeeStatuses(ctx context.Context, invoices []string) error
	CreateReconcileBatches(ctx context.Context, shop, order, from, to string) (*models.ReconcileBatchInfo, error)
}

type ReconcileExtraHandler struct{ svc ReconcileExtraService }

func NewReconcileExtraHandler(svc ReconcileExtraService) *ReconcileExtraHandler {
	return &ReconcileExtraHandler{svc: svc}
}

func (h *ReconcileExtraHandler) RegisterRoutes(r gin.IRouter) {
	grp := r.Group("/reconcile")
	grp.POST("/check/:kodePesanan", h.checkComplete)
	grp.POST("/cancel/:kodePesanan", h.cancel)
	grp.POST("/status/:invoice", h.updateStatus)
	grp.POST("/statuses", h.updateStatuses)
	grp.POST("/batch", h.createBatch)
}

func (h *ReconcileExtraHandler) checkComplete(c *gin.Context) {
	kodePesanan := c.Param("kodePesanan")
	if kodePesanan == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "kodePesanan required"})
		return
	}
	if err := h.svc.CheckAndMarkComplete(c.Request.Context(), kodePesanan); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "completed"})
}

func (h *ReconcileExtraHandler) cancel(c *gin.Context) {
	kodePesanan := c.Param("kodePesanan")
	if kodePesanan == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "kodePesanan required"})
		return
	}
	if err := h.svc.CancelPurchase(c.Request.Context(), kodePesanan); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "cancelled"})
}

func (h *ReconcileExtraHandler) updateStatus(c *gin.Context) {
	invoice := c.Param("invoice")
	if invoice == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invoice required"})
		return
	}
	if err := h.svc.UpdateShopeeStatus(c.Request.Context(), invoice); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func (h *ReconcileExtraHandler) updateStatuses(c *gin.Context) {
	var req struct {
		Invoices []string `json:"invoices" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.UpdateShopeeStatuses(c.Request.Context(), req.Invoices); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func (h *ReconcileExtraHandler) createBatch(c *gin.Context) {
	var req struct {
		Shop  string `json:"shop"`
		Order string `json:"order"`
		From  string `json:"from"`
		To    string `json:"to"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Call the service which will return batch information
	batchInfo, err := h.svc.CreateReconcileBatches(context.Background(), req.Shop, req.Order, req.From, req.To)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Reconcile batches created successfully",
		"batches_created": batchInfo.BatchCount,
		"total_transactions": batchInfo.TotalTransactions,
		"status": "processing will begin shortly",
	})
}