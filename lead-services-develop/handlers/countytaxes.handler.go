package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tuliporg/agent-management-api/middleware"
	"github.com/tuliporg/agent-management-api/models"
	"github.com/tuliporg/agent-management-api/utils"
	"gorm.io/gorm"
)

func (h *Handler) CreateCountyTax(c *gin.Context) {
	db := middleware.GetDB(c)
	utils.Log(c, "County Created successfully")
	var input models.CountyTax
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
			Error:   "Failed to create county taxes",
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
		"message": "CountyTax creation successfully",
		"data":    input,
	})
}

func (h *Handler) UpdateCountyTax(c *gin.Context) {
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

	var input models.CountyTax
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid input",
			Code:    "VALIDATION_ERROR",
			Details: err.Error(),
		})
		return
	}
	tx := db.Begin()
	var existing models.CountyTax
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
	existing.Name = input.Name
	existing.TaxPercentage = input.TaxPercentage
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
		"message": "CountyTax updated successfully",
		"data":    existing,
	})
}
func (h *Handler) DeleteCountyTax(c *gin.Context) {
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
	var CountyTax models.CountyTax
	if err := tx.First(&CountyTax, "id = ?", id); err != nil {
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

	if err := tx.Delete(&CountyTax); err != nil {
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
func (h *Handler) GetAllCountyTaxes(c *gin.Context) {
	db := middleware.GetDB(c)

	all := c.Query("all")

	page, pageSize := h.getPaginationParams(c)
	offset := (page - 1) * pageSize

	var countyTaxes []models.CountyTax
	var total int64

	countQuery := db.Model(&models.CountyTax{})
	if err := countQuery.Count(&total).Error(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to count county taxes",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	if all == "true" {
		if err := countQuery.Find(&countyTaxes); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to retrieve all county taxes",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":  countyTaxes,
			"total": total,
			"all":   true,
		})
		return
	}

	if err := countQuery.Offset(offset).Limit(pageSize).Find(&countyTaxes); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to retrieve county taxes",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	c.JSON(http.StatusOK, gin.H{
		"data":        countyTaxes,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}
