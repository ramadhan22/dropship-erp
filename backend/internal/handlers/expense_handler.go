package handlers

import (
	"context"
	"net/http"

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
	if err := h.svc.CreateExpense(context.Background(), &e); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, e)
}

func (h *ExpenseHandler) list(c *gin.Context) {
	ex, err := h.svc.ListExpenses(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ex)
}

func (h *ExpenseHandler) delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.DeleteExpense(context.Background(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}
