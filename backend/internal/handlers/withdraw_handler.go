package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type WithdrawService interface {
	WithdrawShopeeBalance(ctx context.Context, store string, amount float64) error
}

type WithdrawHandler struct{ svc WithdrawService }

func NewWithdrawHandler(svc WithdrawService) *WithdrawHandler { return &WithdrawHandler{svc: svc} }

func (h *WithdrawHandler) RegisterRoutes(r gin.IRouter) { r.POST("/withdraw", h.handleWithdraw) }

func (h *WithdrawHandler) handleWithdraw(c *gin.Context) {
	var req struct {
		Store  string  `json:"store" binding:"required"`
		Amount float64 `json:"amount" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount must be positive"})
		return
	}
	if err := h.svc.WithdrawShopeeBalance(c.Request.Context(), req.Store, req.Amount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}
