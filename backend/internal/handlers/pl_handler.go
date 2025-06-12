package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/service"
)

type PLHandler struct{ svc *service.PLService }

func NewPLHandler(s *service.PLService) *PLHandler { return &PLHandler{svc: s} }

func (h *PLHandler) RegisterRoutes(r gin.IRouter) {
	grp := r.Group("/pl")
	grp.GET("/", h.get)
}

func (h *PLHandler) get(c *gin.Context) {
	shop := c.Query("shop")
	period := c.Query("period")
	m, err := h.svc.ComputePL(context.Background(), shop, period)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, m)
}
