package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/ramadhan22/dropship-erp/backend/internal/service"
)

type ExpenseHandler struct{ svc *service.ExpenseService }

func NewExpenseHandler(svc *service.ExpenseService) *ExpenseHandler { return &ExpenseHandler{svc: svc} }

func (h *ExpenseHandler) RegisterRoutes(r gin.IRouter) {
	grp := r.Group("/expenses")
	grp.POST("/", h.create)
	grp.GET("/", h.list)
	grp.GET("/:id", h.get)
	grp.PUT("/:id", h.update)
	grp.DELETE("/:id", h.delete)
}

func (h *ExpenseHandler) create(c *gin.Context) {
	var e models.Expense
	if err := c.ShouldBindJSON(&e); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if e.ID == "" {
		e.ID = ""
	} // placeholder, DB default
	var total float64
	for _, l := range e.Lines {
		total += l.Amount
	}
	if total != e.Amount && e.Amount != 0 {
		e.Amount = total
	}
	if total <= 0 || e.AssetAccountID == 0 || len(e.Lines) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid expense"})
		return
	}
	if err := h.svc.CreateExpense(context.Background(), &e); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, e)
}

func (h *ExpenseHandler) list(c *gin.Context) {
	accountID, _ := strconv.ParseInt(c.Query("account_id"), 10, 64)
	sortBy := c.DefaultQuery("sort", "date")
	dir := c.DefaultQuery("dir", "desc")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (page - 1) * size
	ex, total, err := h.svc.ListExpenses(context.Background(), accountID, sortBy, dir, size, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": ex, "total": total})
}

func (h *ExpenseHandler) get(c *gin.Context) {
	id := c.Param("id")
	ex, err := h.svc.GetExpense(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if ex == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, ex)
}

func (h *ExpenseHandler) update(c *gin.Context) {
	id := c.Param("id")
	var e models.Expense
	if err := c.ShouldBindJSON(&e); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	e.ID = id
	if err := h.svc.UpdateExpense(context.Background(), &e); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (h *ExpenseHandler) delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.DeleteExpense(context.Background(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}
