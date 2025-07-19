package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/service"
)

// ForecastHandler handles forecast-related HTTP requests
type ForecastHandler struct {
	forecastService *service.ForecastService
}

// NewForecastHandler creates a new ForecastHandler
func NewForecastHandler(fs *service.ForecastService) *ForecastHandler {
	return &ForecastHandler{
		forecastService: fs,
	}
}

// HandleGenerateForecast handles POST /api/forecast/generate
func (fh *ForecastHandler) HandleGenerateForecast(c *gin.Context) {
	log.Printf("Processing forecast generation request")

	var req service.ForecastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Invalid forecast request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Validate required fields
	if req.Shop == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Shop parameter is required",
		})
		return
	}

	if req.Period == "" {
		req.Period = "monthly" // default to monthly
	}

	if req.Period != "monthly" && req.Period != "yearly" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Period must be 'monthly' or 'yearly'",
		})
		return
	}

	// Set default dates if not provided
	now := time.Now()
	if req.StartDate.IsZero() {
		if req.Period == "monthly" {
			// Default to start of current month minus 3 months for historical data
			req.StartDate = time.Date(now.Year(), now.Month()-3, 1, 0, 0, 0, 0, time.UTC)
		} else {
			// Default to start of current year minus 2 years for historical data
			req.StartDate = time.Date(now.Year()-2, 1, 1, 0, 0, 0, 0, time.UTC)
		}
	}

	if req.EndDate.IsZero() {
		req.EndDate = now
	}

	if req.ForecastTo.IsZero() {
		if req.Period == "monthly" {
			// Forecast to end of current month
			req.ForecastTo = time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, -1)
		} else {
			// Forecast to end of current year
			req.ForecastTo = time.Date(now.Year(), 12, 31, 0, 0, 0, 0, time.UTC)
		}
	}

	log.Printf("Generating forecast: shop=%s, period=%s, start=%s, end=%s, forecastTo=%s",
		req.Shop, req.Period, req.StartDate.Format("2006-01-02"), 
		req.EndDate.Format("2006-01-02"), req.ForecastTo.Format("2006-01-02"))

	// Generate forecast
	forecast, err := fh.forecastService.GenerateForecast(c.Request.Context(), req)
	if err != nil {
		log.Printf("Failed to generate forecast: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate forecast",
			"details": err.Error(),
		})
		return
	}

	log.Printf("Forecast generated successfully for shop %s", req.Shop)
	c.JSON(http.StatusOK, forecast)
}

// HandleGetForecastParams handles GET /api/forecast/params
func (fh *ForecastHandler) HandleGetForecastParams(c *gin.Context) {
	shop := c.Query("shop")
	period := c.Query("period")
	
	if period == "" {
		period = "monthly"
	}

	now := time.Now()
	var suggestedStart, suggestedEnd, suggestedForecastTo time.Time

	if period == "monthly" {
		// Suggest last 6 months for historical data
		suggestedStart = time.Date(now.Year(), now.Month()-6, 1, 0, 0, 0, 0, time.UTC)
		suggestedEnd = now
		// Forecast to end of current month
		suggestedForecastTo = time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, -1)
	} else {
		// Suggest last 3 years for historical data
		suggestedStart = time.Date(now.Year()-3, 1, 1, 0, 0, 0, 0, time.UTC)
		suggestedEnd = now
		// Forecast to end of current year
		suggestedForecastTo = time.Date(now.Year(), 12, 31, 0, 0, 0, 0, time.UTC)
	}

	c.JSON(http.StatusOK, gin.H{
		"shop":                shop,
		"period":              period,
		"suggestedStartDate":  suggestedStart.Format("2006-01-02"),
		"suggestedEndDate":    suggestedEnd.Format("2006-01-02"),
		"suggestedForecastTo": suggestedForecastTo.Format("2006-01-02"),
		"currentDate":         now.Format("2006-01-02"),
	})
}

// HandleGetForecastSummary handles GET /api/forecast/summary
func (fh *ForecastHandler) HandleGetForecastSummary(c *gin.Context) {
	shop := c.Query("shop")
	if shop == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Shop parameter is required",
		})
		return
	}

	period := c.Query("period")
	if period == "" {
		period = "monthly"
	}

	daysStr := c.Query("days")
	days := 30 // default to 30 days
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	now := time.Now()
	startDate := now.AddDate(0, 0, -days)
	
	// Create forecast request for summary
	req := service.ForecastRequest{
		Shop:       shop,
		Period:     period,
		StartDate:  startDate,
		EndDate:    now,
		ForecastTo: now.AddDate(0, 0, days), // Forecast same number of days into future
	}

	forecast, err := fh.forecastService.GenerateForecast(c.Request.Context(), req)
	if err != nil {
		log.Printf("Failed to generate forecast summary: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate forecast summary",
			"details": err.Error(),
		})
		return
	}

	// Create summary response
	summary := gin.H{
		"shop":                 shop,
		"period":               period,
		"days":                 days,
		"forecastSales":        forecast.Sales.TotalForecast,
		"forecastExpenses":     forecast.Expenses.TotalForecast,
		"forecastProfit":       forecast.Profit.TotalForecast,
		"historicalSales":      forecast.Sales.TotalHistorical,
		"historicalExpenses":   forecast.Expenses.TotalHistorical,
		"historicalProfit":     forecast.Profit.TotalHistorical,
		"salesGrowthRate":      forecast.Sales.GrowthRate,
		"expensesGrowthRate":   forecast.Expenses.GrowthRate,
		"profitGrowthRate":     forecast.Profit.GrowthRate,
		"salesConfidence":      forecast.Sales.Confidence,
		"expensesConfidence":   forecast.Expenses.Confidence,
		"profitConfidence":     forecast.Profit.Confidence,
		"generated":            forecast.Generated,
	}

	c.JSON(http.StatusOK, summary)
}