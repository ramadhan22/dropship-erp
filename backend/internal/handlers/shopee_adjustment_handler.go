package handlers

import (
	"context"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

type ShopeeAdjustmentSvc interface {
	ImportXLSX(ctx context.Context, r io.Reader) (int, error)
	List(ctx context.Context, from, to string) ([]models.ShopeeAdjustment, error)
	Update(ctx context.Context, a *models.ShopeeAdjustment) error
	Delete(ctx context.Context, id int64) error
}

type ShopeeAdjustmentHandler struct{ svc ShopeeAdjustmentSvc }

func NewShopeeAdjustmentHandler(s ShopeeAdjustmentSvc) *ShopeeAdjustmentHandler {
	return &ShopeeAdjustmentHandler{svc: s}
}

func (h *ShopeeAdjustmentHandler) RegisterRoutes(r gin.IRouter) {
	grp := r.Group("/shopee/adjustments")
	grp.POST("/import", h.importXLSX)
	grp.GET("/", h.list)
	grp.PUT("/:id", h.update)
	grp.DELETE("/:id", h.delete)
}

func (h *ShopeeAdjustmentHandler) importXLSX(c *gin.Context) {
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

func (h *ShopeeAdjustmentHandler) list(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")
	list, err := h.svc.List(c.Request.Context(), from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *ShopeeAdjustmentHandler) update(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var a models.ShopeeAdjustment
	if err := c.ShouldBindJSON(&a); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	a.ID = id
	if err := h.svc.Update(c.Request.Context(), &a); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (h *ShopeeAdjustmentHandler) delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}
