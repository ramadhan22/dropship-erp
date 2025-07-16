package models

import (
	"fmt"
	"strings"
	"time"
)

// FilterOperator represents the type of filter operation
type FilterOperator string

const (
	FilterOpEquals      FilterOperator = "eq"
	FilterOpNotEquals   FilterOperator = "neq"
	FilterOpContains    FilterOperator = "contains"
	FilterOpStartsWith  FilterOperator = "starts_with"
	FilterOpEndsWith    FilterOperator = "ends_with"
	FilterOpGreaterThan FilterOperator = "gt"
	FilterOpLessThan    FilterOperator = "lt"
	FilterOpGreaterEq   FilterOperator = "gte"
	FilterOpLessEq      FilterOperator = "lte"
	FilterOpBetween     FilterOperator = "between"
	FilterOpIn          FilterOperator = "in"
	FilterOpNotIn       FilterOperator = "not_in"
	FilterOpIsNull      FilterOperator = "is_null"
	FilterOpIsNotNull   FilterOperator = "is_not_null"
)

// FilterCondition represents a single filter condition
type FilterCondition struct {
	Field    string         `json:"field"`
	Operator FilterOperator `json:"operator"`
	Value    interface{}    `json:"value"`
	Values   []interface{}  `json:"values,omitempty"` // For IN, NOT_IN, BETWEEN operations
}

// FilterGroup represents a group of filter conditions with AND/OR logic
type FilterGroup struct {
	Logic      string            `json:"logic"` // "AND" or "OR"
	Conditions []FilterCondition `json:"conditions"`
	Groups     []FilterGroup     `json:"groups,omitempty"` // Nested groups
}

// SortCondition represents a sorting condition
type SortCondition struct {
	Field     string `json:"field"`
	Direction string `json:"direction"` // "asc" or "desc"
}

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Offset   int `json:"offset"`
	Limit    int `json:"limit"`
}

// FilterParams represents the complete filter, sort, and pagination parameters
type FilterParams struct {
	Filters    *FilterGroup      `json:"filters,omitempty"`
	Sort       []SortCondition   `json:"sort,omitempty"`
	Pagination *PaginationParams `json:"pagination,omitempty"`
}

// QueryResult represents the result of a filtered query
type QueryResult struct {
	Data       interface{} `json:"data"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// NewPaginationParams creates pagination parameters with defaults
func NewPaginationParams(page, pageSize int) *PaginationParams {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 1000 {
		pageSize = 1000 // Limit max page size
	}

	offset := (page - 1) * pageSize
	return &PaginationParams{
		Page:     page,
		PageSize: pageSize,
		Offset:   offset,
		Limit:    pageSize,
	}
}

// NewQueryResult creates a query result with pagination metadata
func NewQueryResult(data interface{}, total, page, pageSize int) *QueryResult {
	totalPages := (total + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}

	return &QueryResult{
		Data:       data,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}

// Validate checks if the filter condition is valid
func (fc *FilterCondition) Validate() error {
	if fc.Field == "" {
		return fmt.Errorf("field is required")
	}

	switch fc.Operator {
	case FilterOpEquals, FilterOpNotEquals, FilterOpContains, FilterOpStartsWith, FilterOpEndsWith,
		FilterOpGreaterThan, FilterOpLessThan, FilterOpGreaterEq, FilterOpLessEq:
		if fc.Value == nil {
			return fmt.Errorf("value is required for operator %s", fc.Operator)
		}
	case FilterOpBetween:
		if len(fc.Values) != 2 {
			return fmt.Errorf("between operator requires exactly 2 values")
		}
	case FilterOpIn, FilterOpNotIn:
		if len(fc.Values) == 0 {
			return fmt.Errorf("in/not_in operator requires at least 1 value")
		}
	case FilterOpIsNull, FilterOpIsNotNull:
		// No value required
	default:
		return fmt.Errorf("unsupported operator: %s", fc.Operator)
	}

	return nil
}

// Validate checks if the filter group is valid
func (fg *FilterGroup) Validate() error {
	if fg.Logic != "AND" && fg.Logic != "OR" {
		return fmt.Errorf("logic must be AND or OR")
	}

	for _, condition := range fg.Conditions {
		if err := condition.Validate(); err != nil {
			return fmt.Errorf("invalid condition: %w", err)
		}
	}

	for _, group := range fg.Groups {
		if err := group.Validate(); err != nil {
			return fmt.Errorf("invalid nested group: %w", err)
		}
	}

	return nil
}

// Validate checks if the sort condition is valid
func (sc *SortCondition) Validate() error {
	if sc.Field == "" {
		return fmt.Errorf("sort field is required")
	}
	if sc.Direction != "asc" && sc.Direction != "desc" {
		return fmt.Errorf("sort direction must be asc or desc")
	}
	return nil
}

// Validate checks if the filter parameters are valid
func (fp *FilterParams) Validate() error {
	if fp.Filters != nil {
		if err := fp.Filters.Validate(); err != nil {
			return fmt.Errorf("invalid filters: %w", err)
		}
	}

	for _, sort := range fp.Sort {
		if err := sort.Validate(); err != nil {
			return fmt.Errorf("invalid sort: %w", err)
		}
	}

	return nil
}

// ConvertValue converts a string value to the appropriate type for SQL operations
func ConvertValue(value interface{}, field string) interface{} {
	str, ok := value.(string)
	if !ok {
		return value
	}

	// Try to convert common types
	if strings.Contains(strings.ToLower(field), "date") || strings.Contains(strings.ToLower(field), "time") {
		if t, err := time.Parse("2006-01-02", str); err == nil {
			return t
		}
		if t, err := time.Parse("2006-01-02T15:04:05Z", str); err == nil {
			return t
		}
		if t, err := time.Parse("2006-01-02 15:04:05", str); err == nil {
			return t
		}
	}

	// Return as string for text operations
	return str
}