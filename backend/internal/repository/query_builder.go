package repository

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// QueryBuilder builds SQL queries from filter parameters
type QueryBuilder struct {
	baseQuery     string
	whereClause   strings.Builder
	args          []interface{}
	argIndex      int
	allowedFields map[string]string // field name -> column name mapping
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder(baseQuery string) *QueryBuilder {
	return &QueryBuilder{
		baseQuery:     baseQuery,
		allowedFields: make(map[string]string),
		argIndex:      0,
	}
}

// SetAllowedFields sets the mapping of filter field names to database column names
func (qb *QueryBuilder) SetAllowedFields(fields map[string]string) *QueryBuilder {
	qb.allowedFields = fields
	return qb
}

// BuildQuery builds the complete SQL query from filter parameters
func (qb *QueryBuilder) BuildQuery(params *models.FilterParams) (string, []interface{}, error) {
	var query strings.Builder
	query.WriteString(qb.baseQuery)

	// Build WHERE clause from filters
	if params != nil && params.Filters != nil {
		whereSQL, err := qb.buildWhereClause(params.Filters)
		if err != nil {
			return "", nil, err
		}
		if whereSQL != "" {
			if strings.Contains(strings.ToUpper(qb.baseQuery), "WHERE") {
				query.WriteString(" AND (")
				query.WriteString(whereSQL)
				query.WriteString(")")
			} else {
				query.WriteString(" WHERE ")
				query.WriteString(whereSQL)
			}
		}
	}

	// Build ORDER BY clause
	if params != nil && len(params.Sort) > 0 {
		orderBy, err := qb.buildOrderBy(params.Sort)
		if err != nil {
			return "", nil, err
		}
		query.WriteString(" ORDER BY ")
		query.WriteString(orderBy)
	}

	// Build LIMIT and OFFSET
	if params != nil && params.Pagination != nil {
		query.WriteString(fmt.Sprintf(" LIMIT %d OFFSET %d", params.Pagination.Limit, params.Pagination.Offset))
	}

	return query.String(), qb.args, nil
}

// BuildCountQuery builds a count query for pagination
func (qb *QueryBuilder) BuildCountQuery(params *models.FilterParams) (string, []interface{}, error) {
	// Reset args for count query
	qb.args = []interface{}{}
	qb.argIndex = 0

	var query strings.Builder

	// Extract the base table/joins part from the base query
	baseQuery := qb.baseQuery
	if strings.Contains(strings.ToUpper(baseQuery), "SELECT") {
		// Replace SELECT ... FROM with SELECT COUNT(*) FROM
		parts := strings.SplitN(strings.ToUpper(baseQuery), "FROM", 2)
		if len(parts) == 2 {
			query.WriteString("SELECT COUNT(*) FROM")
			query.WriteString(parts[1])
		} else {
			return "", nil, fmt.Errorf("invalid base query for count")
		}
	} else {
		// Assume it's already a FROM clause
		query.WriteString("SELECT COUNT(*) ")
		query.WriteString(baseQuery)
	}

	// Build WHERE clause from filters (same as regular query)
	if params != nil && params.Filters != nil {
		whereSQL, err := qb.buildWhereClause(params.Filters)
		if err != nil {
			return "", nil, err
		}
		if whereSQL != "" {
			if strings.Contains(strings.ToUpper(query.String()), "WHERE") {
				query.WriteString(" AND (")
				query.WriteString(whereSQL)
				query.WriteString(")")
			} else {
				query.WriteString(" WHERE ")
				query.WriteString(whereSQL)
			}
		}
	}

	return query.String(), qb.args, nil
}

// buildWhereClause builds the WHERE clause from a filter group
func (qb *QueryBuilder) buildWhereClause(group *models.FilterGroup) (string, error) {
	var conditions []string

	// Process individual conditions
	for _, condition := range group.Conditions {
		sql, err := qb.buildCondition(&condition)
		if err != nil {
			return "", err
		}
		if sql != "" {
			conditions = append(conditions, sql)
		}
	}

	// Process nested groups
	for _, nestedGroup := range group.Groups {
		sql, err := qb.buildWhereClause(&nestedGroup)
		if err != nil {
			return "", err
		}
		if sql != "" {
			conditions = append(conditions, fmt.Sprintf("(%s)", sql))
		}
	}

	if len(conditions) == 0 {
		return "", nil
	}

	return strings.Join(conditions, fmt.Sprintf(" %s ", group.Logic)), nil
}

// buildCondition builds SQL for a single filter condition
func (qb *QueryBuilder) buildCondition(condition *models.FilterCondition) (string, error) {
	// Get the column name from the field mapping
	columnName, ok := qb.allowedFields[condition.Field]
	if !ok {
		return "", fmt.Errorf("field '%s' is not allowed for filtering", condition.Field)
	}

	switch condition.Operator {
	case models.FilterOpEquals:
		return qb.buildSimpleCondition(columnName, "=", condition.Value)
	case models.FilterOpNotEquals:
		return qb.buildSimpleCondition(columnName, "!=", condition.Value)
	case models.FilterOpGreaterThan:
		return qb.buildSimpleCondition(columnName, ">", condition.Value)
	case models.FilterOpLessThan:
		return qb.buildSimpleCondition(columnName, "<", condition.Value)
	case models.FilterOpGreaterEq:
		return qb.buildSimpleCondition(columnName, ">=", condition.Value)
	case models.FilterOpLessEq:
		return qb.buildSimpleCondition(columnName, "<=", condition.Value)
	case models.FilterOpContains:
		return qb.buildLikeCondition(columnName, condition.Value, "%%%s%%")
	case models.FilterOpStartsWith:
		return qb.buildLikeCondition(columnName, condition.Value, "%s%%")
	case models.FilterOpEndsWith:
		return qb.buildLikeCondition(columnName, condition.Value, "%%%s")
	case models.FilterOpBetween:
		return qb.buildBetweenCondition(columnName, condition.Values)
	case models.FilterOpIn:
		return qb.buildInCondition(columnName, condition.Values, "IN")
	case models.FilterOpNotIn:
		return qb.buildInCondition(columnName, condition.Values, "NOT IN")
	case models.FilterOpIsNull:
		return fmt.Sprintf("%s IS NULL", columnName), nil
	case models.FilterOpIsNotNull:
		return fmt.Sprintf("%s IS NOT NULL", columnName), nil
	default:
		return "", fmt.Errorf("unsupported operator: %s", condition.Operator)
	}
}

// buildSimpleCondition builds a simple comparison condition
func (qb *QueryBuilder) buildSimpleCondition(columnName, operator string, value interface{}) (string, error) {
	qb.argIndex++
	placeholder := fmt.Sprintf("$%d", qb.argIndex)
	qb.args = append(qb.args, models.ConvertValue(value, columnName))
	return fmt.Sprintf("%s %s %s", columnName, operator, placeholder), nil
}

// buildLikeCondition builds a LIKE condition
func (qb *QueryBuilder) buildLikeCondition(columnName string, value interface{}, pattern string) (string, error) {
	qb.argIndex++
	placeholder := fmt.Sprintf("$%d", qb.argIndex)
	likeValue := fmt.Sprintf(pattern, value)
	qb.args = append(qb.args, likeValue)
	return fmt.Sprintf("%s ILIKE %s", columnName, placeholder), nil
}

// buildBetweenCondition builds a BETWEEN condition
func (qb *QueryBuilder) buildBetweenCondition(columnName string, values []interface{}) (string, error) {
	if len(values) != 2 {
		return "", fmt.Errorf("between requires exactly 2 values")
	}

	qb.argIndex++
	placeholder1 := fmt.Sprintf("$%d", qb.argIndex)
	qb.args = append(qb.args, models.ConvertValue(values[0], columnName))

	qb.argIndex++
	placeholder2 := fmt.Sprintf("$%d", qb.argIndex)
	qb.args = append(qb.args, models.ConvertValue(values[1], columnName))

	return fmt.Sprintf("%s BETWEEN %s AND %s", columnName, placeholder1, placeholder2), nil
}

// buildInCondition builds an IN or NOT IN condition
func (qb *QueryBuilder) buildInCondition(columnName string, values []interface{}, operator string) (string, error) {
	if len(values) == 0 {
		return "", fmt.Errorf("in condition requires at least 1 value")
	}

	var placeholders []string
	for _, value := range values {
		qb.argIndex++
		placeholder := fmt.Sprintf("$%d", qb.argIndex)
		qb.args = append(qb.args, models.ConvertValue(value, columnName))
		placeholders = append(placeholders, placeholder)
	}

	return fmt.Sprintf("%s %s (%s)", columnName, operator, strings.Join(placeholders, ", ")), nil
}

// buildOrderBy builds the ORDER BY clause
func (qb *QueryBuilder) buildOrderBy(sorts []models.SortCondition) (string, error) {
	var orderClauses []string

	for _, sort := range sorts {
		columnName, ok := qb.allowedFields[sort.Field]
		if !ok {
			return "", fmt.Errorf("field '%s' is not allowed for sorting", sort.Field)
		}
		orderClauses = append(orderClauses, fmt.Sprintf("%s %s", columnName, strings.ToUpper(sort.Direction)))
	}

	return strings.Join(orderClauses, ", "), nil
}

// ParseValue parses a string value to the appropriate type
func ParseValue(value string, fieldType reflect.Type) (interface{}, error) {
	switch fieldType.Kind() {
	case reflect.Int, reflect.Int64:
		return strconv.ParseInt(value, 10, 64)
	case reflect.Float64:
		return strconv.ParseFloat(value, 64)
	case reflect.Bool:
		return strconv.ParseBool(value)
	default:
		return value, nil
	}
}
