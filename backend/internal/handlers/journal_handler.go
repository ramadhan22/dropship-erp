package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/service"
)

type JournalHandler struct{ svc *service.JournalService }

func NewJournalHandler(s *service.JournalService) *JournalHandler { return &JournalHandler{svc: s} }

func (h *JournalHandler) RegisterRoutes(r gin.IRouter) {
	grp := r.Group("/journal")
	grp.GET("/", h.list)
	grp.GET("/:id", h.get)
	grp.DELETE("/:id", h.del)
}

func (h *JournalHandler) list(c *gin.Context) {
	list, err := h.svc.List(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *JournalHandler) get(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	je, err := h.svc.Get(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, je)
}

func (h *JournalHandler) del(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.svc.Delete(context.Background(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}
