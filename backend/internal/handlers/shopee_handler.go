package handlers

import (
	"context"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ShopeeServiceInterface defines only the method the handler needs.
type ShopeeServiceInterface interface {
	ImportSettledOrdersXLSX(ctx context.Context, r io.Reader) (int, error)
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
