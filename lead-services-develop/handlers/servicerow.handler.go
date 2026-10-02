// ServiceRows
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

// not using whole
func (h *Handler) CreateServiceRow(c *gin.Context) {
	db := middleware.GetDB(c)

	var input models.ServiceRow
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid input",
			Code:    "VALIDATION_ERROR",
			Details: err.Error(),
		})
		return
	}

	tx := db.Begin()
	if err := tx.Create(&input); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to create service",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to commit transaction",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "ServiceRow creation successfully",
		"data":    input,
	})
}

// GetResidentialEntry retrieves a ResidentialEntry by ID
func (h *Handler) GetServiceRow(c *gin.Context) {
	db := middleware.GetDB(c)

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Missing ID",
			Code:    "VALIDATION_ERROR",
			Details: "estimate_id is required in URL path",
		})
		return
	}

	var serviceRows []models.ServiceRow

	err := db.Where("estimate_id = ?", id).Find(&serviceRows)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Service rows not found",
				Code:    "NOT_FOUND",
				Details: err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to retrieve service rows",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
		}
		return
	}
	for i := range serviceRows {
		var ResidentialEntry models.ResidentialEntry
		err := db.Preload("ManagerResidentialEntry").Preload("ManagerResidentialType").Where("id = ?", serviceRows[i].ResidentialEntryID).Find(&ResidentialEntry)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, ErrorResponse{
					Error:   fmt.Sprintf("Residential entry not found for service row with residential_entry_id: %v", serviceRows[i].ResidentialEntryID),
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
		serviceRows[i].ResidentialEntry = ResidentialEntry
	}
	c.JSON(http.StatusOK, gin.H{
		"data": serviceRows,
	})
}

// UpdateResidentialEntry updates a ResidentialEntry by ID
func (h *Handler) UpdateServiceRow(c *gin.Context) {
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

	var input models.ServiceRow
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid input",
			Code:    "VALIDATION_ERROR",
			Details: err.Error(),
		})
		return
	}
	tx := db.Begin()
	var existing models.ServiceRow
	if err := tx.First(&existing, "id = ?", id); err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Residential entry not found",
				Code:    "NOT_FOUND",
				Details: err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to find residential entry",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
		}
		return
	}

	if err := tx.Save(&existing); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to update residential entry",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to commit transaction",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ServiceRow updated successfully",
		"data":    existing,
	})
}

// DeleteResidentialEntry deletes a ResidentialEntry by ID
func (h *Handler) DeleteServiceRow(c *gin.Context) {
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

	tx := db.Begin()
	var ServiceRow models.ServiceRow
	if err := tx.First(&ServiceRow, "id = ?", id); err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Residential entry not found",
				Code:    "NOT_FOUND",
				Details: err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to find residential entry",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
		}
		return
	}

	if err := tx.Delete(&ServiceRow); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to delete residential entry",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to commit transaction",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Residential entry deleted successfully",
	})
}
