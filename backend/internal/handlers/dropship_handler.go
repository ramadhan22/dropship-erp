package handlers

import (
	"context"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// DropshipServiceInterface defines only the method the handler needs.
type DropshipServiceInterface interface {
	ImportFromCSV(ctx context.Context, r io.Reader) error
}

type DropshipHandler struct {
	svc DropshipServiceInterface
}

// Now accepts any DropshipServiceInterface
func NewDropshipHandler(svc DropshipServiceInterface) *DropshipHandler {
	return &DropshipHandler{svc: svc}
}

func (h *DropshipHandler) HandleImport(c *gin.Context) {
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

	if err := h.svc.ImportFromCSV(context.Background(), f); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "dropship import successful"})
}
