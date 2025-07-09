package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// getIDParam parses the path parameter named "id" as int64.
func getIDParam(c *gin.Context) (int64, error) {
	return strconv.ParseInt(c.Param("id"), 10, 64)
}
