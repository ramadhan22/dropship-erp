package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ReconcileServiceInterface defines only what the handler calls.
type ReconcileServiceInterface interface {
	MatchAndJournal(ctx context.Context, purchaseID, orderID, shop string) error
}

type ReconcileHandler struct {
	svc ReconcileServiceInterface
}

// Now accepts any ReconcileServiceInterface
func NewReconcileHandler(svc ReconcileServiceInterface) *ReconcileHandler {
	return &ReconcileHandler{svc: svc}
}

func (h *ReconcileHandler) HandleMatchAndJournal(c *gin.Context) {
	var req struct {
		Shop       string `json:"shop" binding:"required"`
		PurchaseID string `json:"purchase_id" binding:"required"`
		OrderID    string `json:"order_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.MatchAndJournal(context.Background(), req.PurchaseID, req.OrderID, req.Shop); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "reconciliation successful"})
}
