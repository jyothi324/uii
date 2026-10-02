package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tuliporg/agent-management-api/middleware"
	"github.com/tuliporg/agent-management-api/models"
)

// DeleteResidentialEntry deletes a ResidentialEntry by ID.
func (h *Handler) DeleteResidentialEntry(c *gin.Context) {
	db := middleware.GetDB(c)

	id := c.Param("id")
	if err := db.Delete(&models.ResidentialEntry{}, id); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to delete residential entry",
			Code:    "DATABASE_ERROR",
			Details: "An error occurred while deleting the record",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Residential entry deleted successfully",
	})
}

func (h *Handler) GetAllResidentialEntry(c *gin.Context) {
	db := middleware.GetDB(c)

	var customers []models.ResidentialEntry
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

func (h *Handler) GetAllByLead(c *gin.Context) {
	db := middleware.GetDB(c)

	id := c.Param("id")
	var customers []models.ResidentialEntry
	if err := db.
		Preload("ManagerResidentialEntry").
		Preload("ManagerResidentialType").
		Preload("Lead").
		Preload("Estimate").
		Where("estimate_id = ?", id).
		Find(&customers); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Database error",
			Code:    "DATABASE_ERROR",
			Details: "An error occurred while querying the database: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": customers,
	})
}

func (h *Handler) GetOnlyByLead(c *gin.Context) {
	db := middleware.GetDB(c)

	id := c.Param("id")
	var customers []models.ResidentialEntry
	if err := db.
		Preload("ManagerResidentialEntry").
		Preload("ManagerResidentialType").
		Preload("Lead").
		Preload("Estimate").
		Where("lead_id = ?", id).
		Find(&customers); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Database error",
			Code:    "DATABASE_ERROR",
			Details: "An error occurred while querying the database: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": customers,
	})
}
