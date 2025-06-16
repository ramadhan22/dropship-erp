package handlers

import (
	"context"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

type AdInvoiceService interface {
	ImportInvoicePDF(ctx context.Context, r io.Reader) error
	ListInvoices(ctx context.Context, sortBy, dir string) ([]models.AdInvoice, error)
}

type AdInvoiceHandler struct{ svc AdInvoiceService }

func NewAdInvoiceHandler(svc AdInvoiceService) *AdInvoiceHandler { return &AdInvoiceHandler{svc: svc} }

func (h *AdInvoiceHandler) RegisterRoutes(r gin.IRouter) {
	grp := r.Group("/ad-invoices")
	grp.POST("/", h.handleImport)
	grp.GET("/", h.handleList)
}

func (h *AdInvoiceHandler) handleImport(c *gin.Context) {
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
	if err := h.svc.ImportInvoicePDF(c.Request.Context(), f); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusCreated)
}

func (h *AdInvoiceHandler) handleList(c *gin.Context) {
	sortBy := c.DefaultQuery("sort", "invoice_date")
	dir := c.DefaultQuery("dir", "desc")
	list, err := h.svc.ListInvoices(c.Request.Context(), sortBy, dir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}
