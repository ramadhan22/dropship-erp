package handlers

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

type WithdrawalSvc interface {
	Create(ctx context.Context, w *models.Withdrawal) error
	List(ctx context.Context) ([]models.Withdrawal, error)
	ImportXLSX(ctx context.Context, r io.Reader) (int, error)
}

type WithdrawalHandler struct{ svc WithdrawalSvc }

func NewWithdrawalHandler(s WithdrawalSvc) *WithdrawalHandler { return &WithdrawalHandler{svc: s} }

func (h *WithdrawalHandler) RegisterRoutes(r gin.IRouter) {
	grp := r.Group("/withdrawals")
	grp.POST("/", h.create)
	grp.GET("/", h.list)
	grp.POST("/import", h.importXLSX)
}

func (h *WithdrawalHandler) create(c *gin.Context) {
	var req struct {
		Store  string  `json:"store" binding:"required"`
		Date   string  `json:"date" binding:"required"`
		Amount float64 `json:"amount" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	t, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date"})
		return
	}
	w := &models.Withdrawal{Store: req.Store, Date: t, Amount: req.Amount}
	if err := h.svc.Create(c.Request.Context(), w); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, w)
}

func (h *WithdrawalHandler) list(c *gin.Context) {
	list, err := h.svc.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *WithdrawalHandler) importXLSX(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer f.Close()
	n, err := h.svc.ImportXLSX(c.Request.Context(), f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"inserted": n})
}
