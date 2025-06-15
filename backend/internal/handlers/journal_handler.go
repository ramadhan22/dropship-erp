package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/ramadhan22/dropship-erp/backend/internal/service"
)

type JournalHandler struct{ svc *service.JournalService }

func NewJournalHandler(s *service.JournalService) *JournalHandler { return &JournalHandler{svc: s} }

func (h *JournalHandler) RegisterRoutes(r gin.IRouter) {
	grp := r.Group("/journal")
	grp.POST("/", h.create)
	grp.GET("/", h.list)
	grp.GET("/:id", h.get)
	grp.GET("/:id/lines", h.getLines)
	grp.DELETE("/:id", h.del)
}

func (h *JournalHandler) list(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")
	desc := c.Query("q")
	list, err := h.svc.List(context.Background(), from, to, desc)
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

func (h *JournalHandler) getLines(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	lines, err := h.svc.Lines(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, lines)
}

type journalCreateReq struct {
	Entry models.JournalEntry  `json:"entry"`
	Lines []models.JournalLine `json:"lines"`
}

func (h *JournalHandler) create(c *gin.Context) {
	var req journalCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id, err := h.svc.Create(context.Background(), &req.Entry, req.Lines)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"journal_id": id})
}
