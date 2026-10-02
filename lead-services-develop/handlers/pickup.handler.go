// discounts
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	// "github.com/kr/pretty"
	"github.com/tuliporg/agent-management-api/middleware"
	"github.com/tuliporg/agent-management-api/models"
	"gorm.io/gorm"
)

func (h *Handler) CreatePickup(c *gin.Context) {
	db := middleware.GetDB(c)

	var input models.PickupLocations
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
			Error:   "Failed to create Pickup",
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
		"message": "location creation successfully",
		"data":    input,
	})
}

// UpdateResidentialEntry updates a ResidentialEntry by ID
func (h *Handler) UpdatePickup(c *gin.Context) {
	db := middleware.GetDB(c)

	// Get ID from URL
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Missing ID",
			Code:    "VALIDATION_ERROR",
			Details: "ID is required in URL path",
		})
		return
	}

	// Parse JSON input
	var input models.PickupLocations
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid input",
			Code:    "VALIDATION_ERROR",
			Details: err.Error(),
		})
		return
	}

	// Basic validation: Ensure at least one field is provided
	if input.Name == "" && input.Street == "" && input.City == "" && input.State == "" && input.Zip == "" && input.Country == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid input",
			Code:    "VALIDATION_ERROR",
			Details: "At least one field must be provided for update",
		})
		return
	}

	// Start transaction
	tx := db.Begin()
	if tx.Error() != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to start transaction",
			Code:    "DATABASE_ERROR",
			Details: "asdfasd",
		})
		return
	}

	// Find existing record
	var existing models.PickupLocations
	if err := tx.Where("id = ?", id).First(&existing); err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Pickup location not found",
				Code:    "NOT_FOUND",
				Details: "No pickup location found with the provided ID",
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to find pickup location",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
		}
		return
	}

	// Set ID to ensure correct record is updated
	input.ID = existing.ID

	// Save the updated data
	if err := tx.Save(&input); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to update pickup location",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to commit transaction",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	// Fetch updated record
	var updated models.PickupLocations
	if err := db.Where("id = ?", id).First(&updated); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to fetch updated pickup location",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	// Return success response with updated data
	c.JSON(http.StatusOK, gin.H{
		"message": "Pickup location updated successfully",
		"data":    updated,
	})
}

// DeleteResidentialEntry deletes a ResidentialEntry by ID
func (h *Handler) DeletePickup(c *gin.Context) {
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
	var discount models.PickupLocations
	if err := tx.First(&discount, "id = ?", id); err != nil {
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

	if err := tx.Delete(&discount); err != nil {
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
		"message": "location deleted successfully",
	})
}
func (h *Handler) ListPickupLocations(c *gin.Context) {
	db := middleware.GetDB(c)

	all := c.Query("all") // check if ?all=true

	page, pageSize := h.getPaginationParams(c)
	offset := (page - 1) * pageSize

	var pickupLocations []models.PickupLocations
	var total int64

	pickupQuery := db.Model(&models.PickupLocations{})

	// --- Count total records ---
	if err := pickupQuery.Count(&total).Error(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to count pickup locations",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	// --- If frontend requests all pickup locations ---
	if all == "true" {
		if err := pickupQuery.Find(&pickupLocations); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to retrieve all pickup locations",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":  pickupLocations,
			"total": total,
			"all":   true,
		})
		return
	}

	// --- Normal paginated response (default) ---
	if err := pickupQuery.Offset(offset).Limit(pageSize).Find(&pickupLocations); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to retrieve pickup locations",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	c.JSON(http.StatusOK, gin.H{
		"data":        pickupLocations,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}
