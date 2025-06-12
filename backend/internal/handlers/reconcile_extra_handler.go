package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

type ReconcileExtraService interface {
	ListUnmatched(ctx context.Context, shop string) ([]models.ReconciledTransaction, error)
	BulkReconcile(ctx context.Context, pairs [][2]string, shop string) error
}

type ReconcileExtraHandler struct{ svc ReconcileExtraService }

func NewReconcileExtraHandler(s ReconcileExtraService) *ReconcileExtraHandler {
	return &ReconcileExtraHandler{svc: s}
}

func (h *ReconcileExtraHandler) RegisterRoutes(r gin.IRouter) {
	grp := r.Group("/reconcile")
	grp.GET("/unmatched", h.list)
	grp.POST("/bulk", h.bulk)
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
