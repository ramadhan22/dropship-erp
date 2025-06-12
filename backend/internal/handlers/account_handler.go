package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// AccountServiceInterface defines the service methods needed by the handler.
type AccountServiceInterface interface {
	CreateAccount(ctx context.Context, a *models.Account) (int64, error)
	GetAccount(ctx context.Context, id int64) (*models.Account, error)
	ListAccounts(ctx context.Context) ([]models.Account, error)
	UpdateAccount(ctx context.Context, a *models.Account) error
	DeleteAccount(ctx context.Context, id int64) error
}

// AccountHandler exposes HTTP handlers for account CRUD.
type AccountHandler struct {
	svc AccountServiceInterface
}

func NewAccountHandler(s AccountServiceInterface) *AccountHandler {
	return &AccountHandler{svc: s}
}

func (h *AccountHandler) HandleCreateAccount(c *gin.Context) {
	var req models.Account
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id, err := h.svc.CreateAccount(context.Background(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"account_id": id})
}

func (h *AccountHandler) HandleListAccounts(c *gin.Context) {
	list, err := h.svc.ListAccounts(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *AccountHandler) HandleGetAccount(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
		return
	}
	acc, err := h.svc.GetAccount(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if acc == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, acc)
}

func (h *AccountHandler) HandleUpdateAccount(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
		return
	}
	var req models.Account
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.AccountID = id
	if err := h.svc.UpdateAccount(context.Background(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func (h *AccountHandler) HandleDeleteAccount(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
		return
	}
	if err := h.svc.DeleteAccount(context.Background(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
