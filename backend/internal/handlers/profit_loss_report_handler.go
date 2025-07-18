package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/service"
)

// ProfitLossReportSvc defines required method for the handler.
type ProfitLossReportSvc interface {
	GetProfitLoss(ctx context.Context, typ string, month, year int, store string, comparison bool) (*service.ProfitLoss, error)
}

type ProfitLossReportHandler struct{ svc ProfitLossReportSvc }

func NewProfitLossReportHandler(s ProfitLossReportSvc) *ProfitLossReportHandler {
	return &ProfitLossReportHandler{svc: s}
}

func (h *ProfitLossReportHandler) RegisterRoutes(r gin.IRouter) {
	r.GET("/profitloss", h.handleGet)
}

func (h *ProfitLossReportHandler) handleGet(c *gin.Context) {
	typ := c.DefaultQuery("type", "Monthly")
	month, _ := strconv.Atoi(c.DefaultQuery("month", "0"))
	year, _ := strconv.Atoi(c.DefaultQuery("year", "0"))
	store := c.Query("store")
	comparison := c.DefaultQuery("comparison", "true") == "true"

	res, err := h.svc.GetProfitLoss(context.Background(), typ, month, year, store, comparison)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}
