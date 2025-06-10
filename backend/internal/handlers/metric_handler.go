package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/service"
)

// MetricServiceInterface defines just the bits the handler needs.
type MetricServiceInterface interface {
	CalculateAndCacheMetrics(ctx context.Context, shop, period string) error
	MetricRepo() service.MetricRepoInterface
}

type MetricHandler struct {
	svc MetricServiceInterface
}

func NewMetricHandler(svc MetricServiceInterface) *MetricHandler {
	return &MetricHandler{svc: svc}
}

func (h *MetricHandler) HandleCalculateMetrics(c *gin.Context) {
	var req struct {
		Shop   string `json:"shop" binding:"required"`
		Period string `json:"period" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.CalculateAndCacheMetrics(context.Background(), req.Shop, req.Period); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "metrics computed"})
}

func (h *MetricHandler) HandleGetMetrics(c *gin.Context) {
	shop := c.Query("shop")
	period := c.Query("period")
	cm, err := h.svc.MetricRepo().GetCachedMetric(context.Background(), shop, period)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "metrics not found"})
		return
	}
	c.JSON(http.StatusOK, cm)
}
