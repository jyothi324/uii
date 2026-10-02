package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tuliporg/agent-management-api/middleware"
	"github.com/tuliporg/agent-management-api/models"
)

func (h *Handler) GetAllAssociateCustomers(c *gin.Context) {
	db := middleware.GetDB(c)

	var customers []models.AssociateCustomer
	if err := db.Find(&customers); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to retrieve discounts",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": customers,
	})
}
