package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// DashboardHandler serves mock dashboard data.
type DashboardHandler struct{}

func NewDashboardHandler() *DashboardHandler { return &DashboardHandler{} }

// RegisterRoutes registers the /dashboard endpoint under the given router group.
func (h *DashboardHandler) RegisterRoutes(r gin.IRouter) {
	r.GET("/dashboard", h.handleDashboard)
}

func (h *DashboardHandler) handleDashboard(c *gin.Context) {
	// In a real implementation these values would come from the database.
	c.JSON(http.StatusOK, gin.H{
		"summary": gin.H{
			"total_orders":       gin.H{"value": 1200, "change": 0.05},
			"avg_order_value":    gin.H{"value": 50000, "change": -0.03},
			"total_cancelled":    gin.H{"value": 20, "change": -0.1},
			"total_customers":    gin.H{"value": 1000, "change": 0.1},
			"total_price":        gin.H{"value": 1000000, "change": 0.2},
			"total_discounts":    gin.H{"value": 50000, "change": -0.05},
			"total_net_profit":   gin.H{"value": 300000, "change": 0.15},
			"outstanding_amount": gin.H{"value": 20000, "change": -0.02},
		},
		"charts": gin.H{
			"total_sales":         []gin.H{{"date": "2025-01", "value": 100000}},
			"avg_order_value":     []gin.H{{"date": "2025-01", "value": 50000}},
			"number_of_customers": []gin.H{{"date": "2025-01", "value": 300}},
			"number_of_orders":    []gin.H{{"date": "2025-01", "value": 400}},
		},
	})
}
