package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/ramadhan22/dropship-erp/backend/internal/service"
)

// ReconcileServiceInterface defines the interface for reconcile operations.
type ReconcileServiceInterface interface {
	ReconcileAll(ctx context.Context, req service.ReconcileAllRequest) (*service.ReconcileAllResponse, error)
	GetReconcileCandidates(ctx context.Context, shop, order, from, to string, limit, offset int) ([]models.ReconcileCandidate, int, error)
	GetReconciledTransactions(ctx context.Context, shop, period string) ([]models.ReconciledTransaction, error)
	UpdateEscrowStatus(ctx context.Context, orderSN string, status string) error
}

// ReconcileHandler handles HTTP requests for reconcile operations.
type ReconcileHandler struct {
	svc ReconcileServiceInterface
}

// NewReconcileHandler creates a new reconcile handler.
func NewReconcileHandler(svc ReconcileServiceInterface) *ReconcileHandler {
	return &ReconcileHandler{svc: svc}
}

// RegisterRoutes registers reconcile endpoints.
func (h *ReconcileHandler) RegisterRoutes(r gin.IRouter) {
	grp := r.Group("/reconcile")
	grp.POST("/all", h.handleReconcileAll)
	grp.GET("/candidates", h.handleGetCandidates)
	grp.GET("/transactions", h.handleGetTransactions)
	grp.PUT("/escrow/:orderSN/status", h.handleUpdateEscrowStatus)
}

// handleReconcileAll handles the "reconcile all" operation.
func (h *ReconcileHandler) handleReconcileAll(c *gin.Context) {
	var req service.ReconcileAllRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	response, err := h.svc.ReconcileAll(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// handleGetCandidates returns paginated reconcile candidates.
func (h *ReconcileHandler) handleGetCandidates(c *gin.Context) {
	shop := c.Query("shop")
	order := c.Query("order")
	from := c.Query("from")
	to := c.Query("to")

	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	candidates, total, err := h.svc.GetReconcileCandidates(c.Request.Context(), shop, order, from, to, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   candidates,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// handleGetTransactions returns reconciled transactions for a shop and period.
func (h *ReconcileHandler) handleGetTransactions(c *gin.Context) {
	shop := c.Query("shop")
	period := c.Query("period")

	if shop == "" || period == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "shop and period parameters are required"})
		return
	}

	transactions, err := h.svc.GetReconciledTransactions(c.Request.Context(), shop, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": transactions})
}

// handleUpdateEscrowStatus updates the escrow processing status for an order.
func (h *ReconcileHandler) handleUpdateEscrowStatus(c *gin.Context) {
	orderSN := c.Param("orderSN")
	if orderSN == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "orderSN parameter is required"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	if err := h.svc.UpdateEscrowStatus(c.Request.Context(), orderSN, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Escrow status updated successfully"})
}
