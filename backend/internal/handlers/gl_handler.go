package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/service"
)

type GLHandler struct{ svc *service.GLService }

func NewGLHandler(s *service.GLService) *GLHandler { return &GLHandler{svc: s} }

func (h *GLHandler) RegisterRoutes(r gin.IRouter) {
	grp := r.Group("/generalledger")
	grp.GET("/", h.fetch)
}

func (h *GLHandler) fetch(c *gin.Context) {
	shop := c.Query("shop")
	fromStr := c.Query("from")
	toStr := c.Query("to")
	from, _ := time.Parse("2006-01-02", fromStr)
	to, _ := time.Parse("2006-01-02", toStr)
	res, err := h.svc.FetchGeneralLedger(context.Background(), shop, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}
