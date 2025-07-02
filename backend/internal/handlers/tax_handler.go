package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// TaxServiceInterface defines needed methods.
type TaxServiceInterface interface {
	ComputeTax(ctx context.Context, store, periodType, periodValue string) (*models.TaxPayment, error)
	PayTax(ctx context.Context, tp *models.TaxPayment) error
}

type TaxHandler struct{ svc TaxServiceInterface }

func NewTaxHandler(svc TaxServiceInterface) *TaxHandler { return &TaxHandler{svc: svc} }

func (h *TaxHandler) Register(r gin.IRouter) {
	grp := r.Group("/tax-payment")
	grp.GET("", h.Get)
	grp.POST("/pay", h.Pay)
}

func (h *TaxHandler) Get(c *gin.Context) {
	store := c.Query("store")
	pt := c.Query("type")
	pv := c.Query("period")
	tp, err := h.svc.ComputeTax(c.Request.Context(), store, pt, pv)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tp)
}

func (h *TaxHandler) Pay(c *gin.Context) {
	var tp models.TaxPayment
	if err := c.ShouldBindJSON(&tp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.PayTax(c.Request.Context(), &tp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}
