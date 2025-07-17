package middleware

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

const FilterParamsKey = "filter_params"

// FilterMiddleware parses filter, sort, and pagination parameters from the request
func FilterMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		params := &models.FilterParams{}

		// Parse pagination parameters
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
		params.Pagination = models.NewPaginationParams(page, pageSize)

		// Parse filters
		if filtersJSON := c.Query("filters"); filtersJSON != "" {
			var filters models.FilterGroup
			if err := json.Unmarshal([]byte(filtersJSON), &filters); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid filters format: " + err.Error()})
				c.Abort()
				return
			}
			params.Filters = &filters
		}

		// Parse legacy individual filter parameters for backward compatibility
		if params.Filters == nil {
			filters := parseLegacyFilters(c)
			if filters != nil {
				params.Filters = filters
			}
		}

		// Parse sort parameters
		if sortJSON := c.Query("sort"); sortJSON != "" {
			var sort []models.SortCondition
			if err := json.Unmarshal([]byte(sortJSON), &sort); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sort format: " + err.Error()})
				c.Abort()
				return
			}
			params.Sort = sort
		} else {
			// Parse legacy sort parameters
			if sortBy := c.Query("sort_by"); sortBy != "" {
				direction := c.DefaultQuery("sort_dir", "asc")
				if direction != "asc" && direction != "desc" {
					direction = "asc"
				}
				params.Sort = []models.SortCondition{{
					Field:     sortBy,
					Direction: direction,
				}}
			}
		}

		// Validate parameters
		if err := params.Validate(); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid filter parameters: " + err.Error()})
			c.Abort()
			return
		}

		// Store in context
		c.Set(FilterParamsKey, params)
		c.Next()
	}
}

// GetFilterParams retrieves filter parameters from the Gin context
func GetFilterParams(c *gin.Context) *models.FilterParams {
	if params, exists := c.Get(FilterParamsKey); exists {
		return params.(*models.FilterParams)
	}
	return &models.FilterParams{
		Pagination: models.NewPaginationParams(1, 20),
	}
}

// parseLegacyFilters parses legacy individual query parameters into a filter group
func parseLegacyFilters(c *gin.Context) *models.FilterGroup {
	var conditions []models.FilterCondition

	// Common legacy filter patterns
	legacyFilters := map[string]models.FilterOperator{
		"channel":  models.FilterOpEquals,
		"store":    models.FilterOpEquals,
		"order":    models.FilterOpContains,
		"order_no": models.FilterOpContains,
		"status":   models.FilterOpEquals,
		"from":     models.FilterOpGreaterEq,
		"to":       models.FilterOpLessEq,
	}

	for param, operator := range legacyFilters {
		if value := c.Query(param); value != "" {
			// Handle date fields specially
			field := param
			if param == "from" {
				field = "created_at" // Default date field, can be overridden
			} else if param == "to" {
				field = "created_at"
			}

			conditions = append(conditions, models.FilterCondition{
				Field:    field,
				Operator: operator,
				Value:    value,
			})
		}
	}

	// Parse text search parameters
	if search := c.Query("search"); search != "" {
		// Create a group for text search across multiple fields
		searchFields := []string{"name", "description", "email"} // Default search fields
		if searchFieldsParam := c.Query("search_fields"); searchFieldsParam != "" {
			searchFields = strings.Split(searchFieldsParam, ",")
		}

		if len(searchFields) > 0 {
			searchConditions := make([]models.FilterCondition, len(searchFields))
			for i, field := range searchFields {
				searchConditions[i] = models.FilterCondition{
					Field:    strings.TrimSpace(field),
					Operator: models.FilterOpContains,
					Value:    search,
				}
			}

			// If we have other conditions, we need to create a nested structure
			if len(conditions) > 0 {
				return &models.FilterGroup{
					Logic:      "AND",
					Conditions: conditions,
					Groups: []models.FilterGroup{
						{
							Logic:      "OR",
							Conditions: searchConditions,
						},
					},
				}
			} else {
				return &models.FilterGroup{
					Logic:      "OR",
					Conditions: searchConditions,
				}
			}
		}
	}

	if len(conditions) == 0 {
		return nil
	}

	return &models.FilterGroup{
		Logic:      "AND",
		Conditions: conditions,
	}
}

// ParseDateRange parses date range parameters with flexible formats
func ParseDateRange(from, to string) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	if from != "" {
		result["from"] = from
	}
	if to != "" {
		result["to"] = to
	}

	return result, nil
}

// QuickFilters provides common filter presets
type QuickFilter struct {
	Name    string                 `json:"name"`
	Label   string                 `json:"label"`
	Filters *models.FilterGroup    `json:"filters"`
	Sort    []models.SortCondition `json:"sort,omitempty"`
}

// GetQuickFilters returns predefined quick filters for common use cases
func GetQuickFilters() []QuickFilter {
	return []QuickFilter{
		{
			Name:  "recent",
			Label: "Recent (Last 7 days)",
			Filters: &models.FilterGroup{
				Logic: "AND",
				Conditions: []models.FilterCondition{
					{
						Field:    "created_at",
						Operator: models.FilterOpGreaterEq,
						Value:    "now() - interval '7 days'",
					},
				},
			},
			Sort: []models.SortCondition{
				{Field: "created_at", Direction: "desc"},
			},
		},
		{
			Name:  "this_month",
			Label: "This Month",
			Filters: &models.FilterGroup{
				Logic: "AND",
				Conditions: []models.FilterCondition{
					{
						Field:    "created_at",
						Operator: models.FilterOpGreaterEq,
						Value:    "date_trunc('month', current_date)",
					},
				},
			},
		},
		{
			Name:  "active",
			Label: "Active Only",
			Filters: &models.FilterGroup{
				Logic: "AND",
				Conditions: []models.FilterCondition{
					{
						Field:    "status",
						Operator: models.FilterOpEquals,
						Value:    "active",
					},
				},
			},
		},
	}
}
