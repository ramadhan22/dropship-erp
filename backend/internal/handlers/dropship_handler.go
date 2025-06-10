package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

// DropshipServiceInterface defines only the method the handler needs.
type DropshipServiceInterface interface {
	ImportFromCSV(ctx context.Context, filePath string) error
}

type DropshipHandler struct {
	svc DropshipServiceInterface
}

// Now accepts any DropshipServiceInterface
func NewDropshipHandler(svc DropshipServiceInterface) *DropshipHandler {
	return &DropshipHandler{svc: svc}
}

func (h *DropshipHandler) HandleImport(c *gin.Context) {
	var req struct {
		FilePath string `json:"file_path" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.ImportFromCSV(context.Background(), req.FilePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "dropship import successful"})
}
