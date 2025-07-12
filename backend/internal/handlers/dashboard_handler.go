package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/service"
)

type DashboardServiceInterface interface {
	GetDashboardData(ctx context.Context, f service.DashboardFilters) (*service.DashboardData, error)
}

// DashboardHandler serves dashboard data from the service.
type DashboardHandler struct {
	svc DashboardServiceInterface
}

func NewDashboardHandler(svc DashboardServiceInterface) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

// RegisterRoutes registers the /dashboard endpoint under the given router group.
func (h *DashboardHandler) RegisterRoutes(r gin.IRouter) {
	r.GET("/dashboard", h.handleDashboard)
}

func (h *DashboardHandler) handleDashboard(c *gin.Context) {
	month, _ := strconv.Atoi(c.Query("month"))
	year, _ := strconv.Atoi(c.Query("year"))
	data, err := h.svc.GetDashboardData(c.Request.Context(), service.DashboardFilters{
		Channel: c.Query("channel"),
		Store:   c.Query("store"),
		Period:  c.Query("period"),
		Month:   month,
		Year:    year,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}
