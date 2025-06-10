package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/service"
)

// BalanceServiceInterface defines the one method we need.
type BalanceServiceInterface interface {
	GetBalanceSheet(ctx context.Context, shop string, asOfDate time.Time) ([]service.CategoryBalance, error)
}

type BalanceHandler struct {
	svc BalanceServiceInterface
}

func NewBalanceHandler(svc BalanceServiceInterface) *BalanceHandler {
	return &BalanceHandler{svc: svc}
}

func (h *BalanceHandler) HandleGetBalanceSheet(c *gin.Context) {
	shop := c.Query("shop")
	period := c.Query("period")
	asOf, err := time.Parse("2006-01", period)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid period"})
		return
	}
	asOf = asOf.AddDate(0, 1, 0).Add(-time.Second)

	res, err := h.svc.GetBalanceSheet(context.Background(), shop, asOf)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}
