// MaterialRows
package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tuliporg/agent-management-api/middleware"
	"github.com/tuliporg/agent-management-api/models"
	"gorm.io/gorm"
)

// GetResidentialEntry retrieves a ResidentialEntry by ID
func (h *Handler) GetMaterialRow(c *gin.Context) {
	db := middleware.GetDB(c)

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Missing ID",
			Code:    "VALIDATION_ERROR",
			Details: "ID is required in URL path",
		})
		return
	}

	var MaterialRow []models.MaterialRow

	if err := db.Find(&MaterialRow, "estimate_id = ?", id); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Residential entry not found",
				Code:    "NOT_FOUND",
				Details: err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to retrieve residential entry",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
		}
		return
	}
	for i := range MaterialRow {
		var ResidentialEntry models.ResidentialEntry
		err := db.Preload("ManagerResidentialEntry").Preload("ManagerResidentialType").Where("id = ?", MaterialRow[i].ResidentialEntryID).Find(&ResidentialEntry)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, ErrorResponse{
					Error:   fmt.Sprintf("Residential entry not found for service row with residential_entry_id: %v", MaterialRow[i].ResidentialEntryID),
					Code:    "NOT_FOUND",
					Details: err.Error(),
				})
				return
			} else {
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "Failed to retrieve residential entry",
					Code:    "DATABASE_ERROR",
					Details: err.Error(),
				})
				return
			}
		}
		MaterialRow[i].ResidentialEntry = ResidentialEntry
	}

	c.JSON(http.StatusOK, gin.H{
		"data": MaterialRow,
	})
}
