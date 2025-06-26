package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ConfigProvider defines methods to expose config values needed by the UI.
type ConfigProvider interface {
	ShopeeAuthURL() string
}

type ConfigHandler struct{ cfg ConfigProvider }

func NewConfigHandler(c ConfigProvider) *ConfigHandler { return &ConfigHandler{cfg: c} }

func (h *ConfigHandler) RegisterRoutes(r gin.IRouter) {
	r.GET("/config/shopee-auth-url", h.handleAuthURL)
}

func (h *ConfigHandler) handleAuthURL(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"url": h.cfg.ShopeeAuthURL()})
}
