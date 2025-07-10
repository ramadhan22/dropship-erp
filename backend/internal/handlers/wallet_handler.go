package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/service"
)

// WalletService defines the needed service methods.
type WalletService interface {
	ListWalletTransactions(ctx context.Context, store string, p service.WalletTransactionParams) ([]service.WalletTransaction, bool, error)
}

type WalletHandler struct{ svc WalletService }

func NewWalletHandler(s WalletService) *WalletHandler { return &WalletHandler{svc: s} }

func (h *WalletHandler) RegisterRoutes(r gin.IRouter) {
	r.GET("/wallet/transactions", h.list)
}

func (h *WalletHandler) list(c *gin.Context) {
	store := c.Query("store")
	if store == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "store is required"})
		return
	}
	pageNo, _ := strconv.Atoi(c.DefaultQuery("page_no", "0"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "25"))
	if pageSize > 100 {
		pageSize = 100
	}
	var createFromPtr, createToPtr *int64
	if v := c.Query("create_time_from"); v != "" {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			createFromPtr = &i
		}
	}
	if v := c.Query("create_time_to"); v != "" {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			createToPtr = &i
		}
	}
	params := service.WalletTransactionParams{
		PageNo:             pageNo,
		PageSize:           pageSize,
		CreateTimeFrom:     createFromPtr,
		CreateTimeTo:       createToPtr,
		WalletType:         c.Query("wallet_type"),
		TransactionType:    c.Query("transaction_type"),
		MoneyFlow:          c.Query("money_flow"),
		TransactionTabType: c.Query("transaction_tab_type"),
	}
	list, next, err := h.svc.ListWalletTransactions(c.Request.Context(), store, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list, "has_next_page": next})
}
