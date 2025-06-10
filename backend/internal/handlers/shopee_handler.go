package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ShopeeServiceInterface defines only the method the handler needs.
type ShopeeServiceInterface interface {
	ImportSettledOrdersCSV(ctx context.Context, filePath string) error
}

type ShopeeHandler struct {
	svc ShopeeServiceInterface
}

// Now accepts any ShopeeServiceInterface
func NewShopeeHandler(svc ShopeeServiceInterface) *ShopeeHandler {
	return &ShopeeHandler{svc: svc}
}

func (h *ShopeeHandler) HandleImport(c *gin.Context) {
	var req struct {
		FilePath string `json:"file_path" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.ImportSettledOrdersCSV(context.Background(), req.FilePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "shopee import successful"})
}
