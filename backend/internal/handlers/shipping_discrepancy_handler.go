package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/service"
)

type ShippingDiscrepancyHandler struct {
	svc *service.ShippingDiscrepancyService
}

func NewShippingDiscrepancyHandler(svc *service.ShippingDiscrepancyService) *ShippingDiscrepancyHandler {
	return &ShippingDiscrepancyHandler{svc: svc}
}

func (h *ShippingDiscrepancyHandler) RegisterRoutes(r gin.IRouter) {
	grp := r.Group("/shipping-discrepancies")
	grp.GET("/", h.list)
	grp.GET("/stats", h.stats)
	grp.GET("/invoice/:invoice", h.getByInvoice)
}

// list retrieves shipping discrepancies with pagination and optional filtering
func (h *ShippingDiscrepancyHandler) list(c *gin.Context) {
	storeName := c.Query("store_name")
	discrepancyType := c.Query("type")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	offset := (page - 1) * size

	discrepancies, err := h.svc.GetShippingDiscrepancies(context.Background(), storeName, discrepancyType, size, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        discrepancies,
		"page":        page,
		"page_size":   size,
		"total_count": len(discrepancies),
		"store_name":  storeName,
		"type":        discrepancyType,
	})
}

// stats retrieves shipping discrepancy statistics for a date range
func (h *ShippingDiscrepancyHandler) stats(c *gin.Context) {
	startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, -1, 0).Format("2006-01-02"))
	endDateStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))
	statsType := c.DefaultQuery("type", "amounts") // "amounts" or "counts"

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format, use YYYY-MM-DD"})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format, use YYYY-MM-DD"})
		return
	}
	// Include the whole end date
	endDate = endDate.Add(24*time.Hour - time.Second)

	if statsType == "amounts" {
		sums, err := h.svc.GetShippingDiscrepancySums(context.Background(), startDate, endDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"start_date": startDateStr,
			"end_date":   endDateStr,
			"type":       "amounts",
			"stats":      sums,
		})
	} else {
		stats, err := h.svc.GetShippingDiscrepancyStats(context.Background(), startDate, endDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"start_date": startDateStr,
			"end_date":   endDateStr,
			"type":       "counts",
			"stats":      stats,
		})
	}
}

// getByInvoice retrieves a shipping discrepancy by invoice number
func (h *ShippingDiscrepancyHandler) getByInvoice(c *gin.Context) {
	invoice := c.Param("invoice")
	if invoice == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invoice number is required"})
		return
	}

	discrepancy, err := h.svc.GetShippingDiscrepancyByInvoice(context.Background(), invoice)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Shipping discrepancy not found"})
		return
	}

	c.JSON(http.StatusOK, discrepancy)
}
