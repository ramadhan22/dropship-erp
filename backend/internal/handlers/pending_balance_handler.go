package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PendingBalanceService interface {
	GetPendingBalance(ctx context.Context, store string) (float64, error)
}

type PendingBalanceHandler struct{ svc PendingBalanceService }

func NewPendingBalanceHandler(s PendingBalanceService) *PendingBalanceHandler {
	return &PendingBalanceHandler{svc: s}
}

func (h *PendingBalanceHandler) RegisterRoutes(r gin.IRouter) {
	r.GET("/pending-balance", h.get)
}

func (h *PendingBalanceHandler) get(c *gin.Context) {
	store := c.Query("store")
	bal, err := h.svc.GetPendingBalance(context.Background(), store)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"pending_balance": bal})
}
