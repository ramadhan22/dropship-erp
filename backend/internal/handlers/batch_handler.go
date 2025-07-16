package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/service"
)

// BatchHandler exposes endpoints for batch history.
type BatchHandler struct{ svc *service.BatchService }

func NewBatchHandler(s *service.BatchService) *BatchHandler { return &BatchHandler{svc: s} }

func (h *BatchHandler) RegisterRoutes(r gin.IRouter) {
	grp := r.Group("/batches")
	grp.GET("/", h.list)
	grp.GET("/:id/details", h.details)
}

func (h *BatchHandler) list(c *gin.Context) {
	statusStr := c.DefaultQuery("status", "pending,processing")
	statuses := []string{}
	for _, s := range strings.Split(statusStr, ",") {
		s = strings.TrimSpace(s)
		if s != "" {
			statuses = append(statuses, s)
		}
	}
	typeStr := c.Query("type")
	types := []string{}
	if typeStr != "" {
		for _, t := range strings.Split(typeStr, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				types = append(types, t)
			}
		}
	}
	list, err := h.svc.ListFiltered(context.Background(), types, statuses)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *BatchHandler) details(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	list, err := h.svc.ListDetails(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}
