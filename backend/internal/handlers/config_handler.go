package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ConfigProvider defines methods to expose config values needed by the UI.
type ConfigProvider interface {
	ShopeeAuthURL(storeID int64) string
}

type ConfigHandler struct{ cfg ConfigProvider }

func NewConfigHandler(c ConfigProvider) *ConfigHandler { return &ConfigHandler{cfg: c} }

func (h *ConfigHandler) RegisterRoutes(r gin.IRouter) {
	r.GET("/config/shopee-auth-url", h.handleAuthURL)
}

func (h *ConfigHandler) handleAuthURL(c *gin.Context) {
	idStr := c.Query("store_id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "store_id is required"})
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid store_id"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"url": h.cfg.ShopeeAuthURL(id)})
}
