package handlers

import (
	"context"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// ShopeeServiceInterface defines only the method the handler needs.
type ShopeeServiceInterface interface {
	ImportSettledOrdersXLSX(ctx context.Context, r io.Reader) (int, error)
	ListSettled(ctx context.Context, channel, store, date, month, year string, limit, offset int) ([]models.ShopeeSettled, int, error)
}

type ShopeeHandler struct {
	svc ShopeeServiceInterface
}

// Now accepts any ShopeeServiceInterface
func NewShopeeHandler(svc ShopeeServiceInterface) *ShopeeHandler {
	return &ShopeeHandler{svc: svc}
}

func (h *ShopeeHandler) HandleImport(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	f, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer f.Close()

	ctx := c.Request.Context()
	count, err := h.svc.ImportSettledOrdersXLSX(ctx, f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"inserted": count})
}

// HandleListSettled returns paginated shopee settled data with optional filters.
func (h *ShopeeHandler) HandleListSettled(c *gin.Context) {
	channel := c.Query("channel")
	store := c.Query("store")
	date := c.Query("date")
	month := c.Query("month")
	year := c.Query("year")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (page - 1) * size
	ctx := c.Request.Context()
	list, total, err := h.svc.ListSettled(ctx, channel, store, date, month, year, size, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list, "total": total})
}
