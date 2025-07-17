package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// ReconcileServiceInterface defines only what the handler calls.
type ReconcileServiceInterface interface {
	MatchAndJournal(ctx context.Context, purchaseID, orderID, shop string) error
	BulkReconcileWithErrorHandling(ctx context.Context, pairs [][2]string, shop string, batchID *int64) (*models.ReconciliationReport, error)
	GenerateReconciliationReport(ctx context.Context, shop string, since time.Time) (*models.ReconciliationReport, error)
	RetryFailedReconciliations(ctx context.Context, shop string, maxRetries int) (*models.ReconciliationReport, error)
	GetFailedReconciliationsSummary(ctx context.Context, shop string, days int) (map[string]interface{}, error)
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

// HandleBulkReconcileWithErrorHandling processes multiple reconciliation pairs with error handling
func (h *ReconcileHandler) HandleBulkReconcileWithErrorHandling(c *gin.Context) {
	var req struct {
		Shop  string      `json:"shop" binding:"required"`
		Pairs [][2]string `json:"pairs" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	report, err := h.svc.BulkReconcileWithErrorHandling(context.Background(), req.Pairs, req.Shop, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "bulk reconciliation completed",
		"report":  report,
	})
}

// HandleGenerateReconciliationReport generates a comprehensive reconciliation report
func (h *ReconcileHandler) HandleGenerateReconciliationReport(c *gin.Context) {
	shop := c.Query("shop")
	if shop == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "shop parameter is required"})
		return
	}

	daysParam := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid days parameter"})
		return
	}

	since := time.Now().AddDate(0, 0, -days)
	report, err := h.svc.GenerateReconciliationReport(context.Background(), shop, since)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"report": report,
	})
}

// HandleRetryFailedReconciliations retries failed reconciliation transactions
func (h *ReconcileHandler) HandleRetryFailedReconciliations(c *gin.Context) {
	var req struct {
		Shop       string `json:"shop" binding:"required"`
		MaxRetries int    `json:"max_retries"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.MaxRetries <= 0 {
		req.MaxRetries = 50 // Default max retries
	}

	report, err := h.svc.RetryFailedReconciliations(context.Background(), req.Shop, req.MaxRetries)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "retry reconciliation completed",
		"report":  report,
	})
}

// HandleGetFailedReconciliationsSummary provides a quick overview of failed reconciliations
func (h *ReconcileHandler) HandleGetFailedReconciliationsSummary(c *gin.Context) {
	shop := c.Query("shop")
	if shop == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "shop parameter is required"})
		return
	}

	daysParam := c.DefaultQuery("days", "7")
	days, err := strconv.Atoi(daysParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid days parameter"})
		return
	}

	summary, err := h.svc.GetFailedReconciliationsSummary(context.Background(), shop, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"summary": summary,
	})
}
