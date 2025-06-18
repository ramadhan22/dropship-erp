package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/service"
)

type AssetAccountService interface {
	ListBalances(ctx context.Context) ([]service.AssetAccountBalance, error)
	AdjustBalance(ctx context.Context, id int64, newBal float64) error
}

type AssetAccountHandler struct{ svc AssetAccountService }

func NewAssetAccountHandler(svc AssetAccountService) *AssetAccountHandler {
	return &AssetAccountHandler{svc: svc}
}

func (h *AssetAccountHandler) RegisterRoutes(r gin.IRouter) {
	grp := r.Group("/asset-accounts")
	grp.GET("/", h.list)
	grp.PUT("/:id/balance", h.adjust)
}

func (h *AssetAccountHandler) list(c *gin.Context) {
	res, err := h.svc.ListBalances(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *AssetAccountHandler) adjust(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req struct {
		Balance float64 `json:"balance"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.AdjustBalance(context.Background(), id, req.Balance); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "adjusted"})
}
