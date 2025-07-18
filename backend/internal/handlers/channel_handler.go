package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// ChannelServiceInterface defines methods required by the handler.
type ChannelServiceInterface interface {
	CreateJenisChannel(ctx context.Context, jenisChannel string) (int64, error)
	CreateStore(ctx context.Context, channelID int64, namaToko string) (int64, error)
	ListJenisChannels(ctx context.Context) ([]models.JenisChannel, error)
	ListStoresByChannel(ctx context.Context, channelID int64) ([]models.Store, error)
	ListStoresByChannelName(ctx context.Context, channelName string) ([]models.Store, error)
	GetStore(ctx context.Context, id int64) (*models.Store, error)
	ListAllStores(ctx context.Context) ([]models.StoreWithChannel, error)
	UpdateStore(ctx context.Context, st *models.Store) error
	DeleteStore(ctx context.Context, id int64) error
}

type ChannelHandler struct {
	svc ChannelServiceInterface
}

func NewChannelHandler(s ChannelServiceInterface) *ChannelHandler {
	return &ChannelHandler{svc: s}
}

func (h *ChannelHandler) HandleCreateJenisChannel(c *gin.Context) {
	var req struct {
		JenisChannel string `json:"jenis_channel" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id, err := h.svc.CreateJenisChannel(context.Background(), req.JenisChannel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"jenis_channel_id": id})
}

func (h *ChannelHandler) HandleCreateStore(c *gin.Context) {
	var req struct {
		JenisChannelID int64  `json:"jenis_channel_id" binding:"required"`
		NamaToko       string `json:"nama_toko" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id, err := h.svc.CreateStore(context.Background(), req.JenisChannelID, req.NamaToko)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"store_id": id})
}

func (h *ChannelHandler) HandleListJenisChannels(c *gin.Context) {
	list, err := h.svc.ListJenisChannels(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *ChannelHandler) HandleListStores(c *gin.Context) {
	cidStr := c.Param("id")
	cid, err := strconv.ParseInt(cidStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid jenis_channel id"})
		return
	}
	list, err := h.svc.ListStoresByChannel(context.Background(), cid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// HandleListStoresByName returns stores filtered by channel name provided as query param "channel".
func (h *ChannelHandler) HandleListStoresByName(c *gin.Context) {
	channel := c.Query("channel")
	if channel == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "channel is required"})
		return
	}
	list, err := h.svc.ListStoresByChannelName(context.Background(), channel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// HandleGetStore returns a single store by ID.
func (h *ChannelHandler) HandleGetStore(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid store id"})
		return
	}
	st, err := h.svc.GetStore(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, st)
}

// HandleListAllStores returns all stores joined with channel names.
func (h *ChannelHandler) HandleListAllStores(c *gin.Context) {
	list, err := h.svc.ListAllStores(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// HandleUpdateStore updates a store row.
func (h *ChannelHandler) HandleUpdateStore(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid store id"})
		return
	}
	var req struct {
		JenisChannelID int64   `json:"jenis_channel_id"`
		NamaToko       string  `json:"nama_toko"`
		CodeID         *string `json:"code_id"`
		ShopID         *string `json:"shop_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	st := &models.Store{
		StoreID:        id,
		JenisChannelID: req.JenisChannelID,
		NamaToko:       req.NamaToko,
		CodeID:         req.CodeID,
		ShopID:         req.ShopID,
	}
	if err := h.svc.UpdateStore(context.Background(), st); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// HandleDeleteStore removes a store.
func (h *ChannelHandler) HandleDeleteStore(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid store id"})
		return
	}
	if err := h.svc.DeleteStore(context.Background(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}
