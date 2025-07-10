package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/service"
)

// AdsTopupService defines needed methods for ads topup operations.
type AdsTopupService interface {
	List(ctx context.Context, store string, p service.WalletTransactionParams) ([]service.WalletTransaction, bool, error)
	CreateJournal(ctx context.Context, store string, t service.WalletTransaction) error
}

type AdsTopupHandler struct{ svc AdsTopupService }

func NewAdsTopupHandler(s AdsTopupService) *AdsTopupHandler { return &AdsTopupHandler{svc: s} }

func (h *AdsTopupHandler) RegisterRoutes(r gin.IRouter) {
	grp := r.Group("/ads-topups")
	grp.GET("", h.list)
	grp.POST("/journal", h.journal)
}

func (h *AdsTopupHandler) list(c *gin.Context) {
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
	var fromPtr, toPtr *int64
	if v := c.Query("create_time_from"); v != "" {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			fromPtr = &i
		}
	}
	if v := c.Query("create_time_to"); v != "" {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			toPtr = &i
		}
	}
	params := service.WalletTransactionParams{
		PageNo:         pageNo,
		PageSize:       pageSize,
		CreateTimeFrom: fromPtr,
		CreateTimeTo:   toPtr,
	}
	list, next, err := h.svc.List(c.Request.Context(), store, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list, "has_next_page": next})
}

func (h *AdsTopupHandler) journal(c *gin.Context) {
	var req struct {
		Store         string  `json:"store"`
		TransactionID int64   `json:"transaction_id"`
		CreateTime    int64   `json:"create_time"`
		Amount        float64 `json:"amount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	txn := service.WalletTransaction{
		TransactionID: req.TransactionID,
		CreateTime:    req.CreateTime,
		Amount:        req.Amount,
	}
	if err := h.svc.CreateJournal(c.Request.Context(), req.Store, txn); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "journaled"})
}
