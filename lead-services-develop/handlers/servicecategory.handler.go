package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tuliporg/agent-management-api/middleware"
	"github.com/tuliporg/agent-management-api/models"
	"gorm.io/gorm"
)

func (h *Handler) CreateServiceCategory(c *gin.Context) {
	db := middleware.GetDB(c)

	var category models.ServiceCategory
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Code:    "VALIDATION_ERROR",
			Details: err.Error(),
		})
		return
	}

	category.Name = strings.TrimSpace(category.Name)
	if category.ProCost < 0 || category.RetailCost < 0 {
    c.JSON(http.StatusBadRequest, ErrorResponse{
        Error:   "Costs must be non-negative",
        Code:    "VALIDATION_ERROR",
        Details: "pro_cost and retail_cost must be >= 0",
    })
    return
}
	if category.Name == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Name is required",
			Code:    "VALIDATION_ERROR",
			Details: "category name cannot be empty",
		})
		return
	}

	// Check duplicate
	var existing models.ServiceCategory
	if err := db.Where("LOWER(name) = LOWER(?)", category.Name).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, ErrorResponse{
			Error:   "Category already exists",
			Code:    "DUPLICATE_ERROR",
			Details: "a service category with this name already exists",
		})
		return
	}

	tx := db.Begin()
	if err := tx.Create(&category); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to create service category",
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

	c.JSON(http.StatusCreated, gin.H{"data": category, "message": "service category created successfully"})
}

func (h *Handler) GetAllServiceCategories(c *gin.Context) {
	db := middleware.GetDB(c)

	all := c.Query("all")
	search := strings.TrimSpace(c.Query("search"))

	page, pageSize := h.getPaginationParams(c)
	offset := (page - 1) * pageSize

	var categories []models.ServiceCategory
	var total int64

	query := db.Model(&models.ServiceCategory{})
	if search != "" {
		query = query.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(search)+"%")
	}

	if err := query.Count(&total).Error(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to count service categories",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	if all == "true" {
		if err := query.Find(&categories); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to retrieve service categories",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": categories, "total": total, "all": true})
		return
	}

	if err := query.Offset(offset).Limit(pageSize).Find(&categories); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to retrieve service categories",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	c.JSON(http.StatusOK, gin.H{
		"data":        categories,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

func (h *Handler) UpdateServiceCategory(c *gin.Context) {
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

	var body models.ServiceCategory
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Code:    "VALIDATION_ERROR",
			Details: err.Error(),
		})
		return
	}

	body.Name = strings.TrimSpace(body.Name)
	if body.Name == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Name is required",
			Code:    "VALIDATION_ERROR",
			Details: "category name cannot be empty",
		})
		return
	}
	if body.ProCost < 0 || body.RetailCost < 0 {
    c.JSON(http.StatusBadRequest, ErrorResponse{
        Error:   "Costs must be non-negative",
        Code:    "VALIDATION_ERROR",
        Details: "pro_cost and retail_cost must be >= 0",
    })
    return
}

	tx := db.Begin()

	var category models.ServiceCategory
	if err := tx.First(&category, "id = ?", id); err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Service category not found",
				Code:    "NOT_FOUND",
				Details: err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to find service category",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
		}
		return
	}

	// Check duplicate name (excluding self)
	var existing models.ServiceCategory
	if err := tx.Where("LOWER(name) = LOWER(?) AND id != ?", body.Name, id).First(&existing).Error; err == nil {
		tx.Rollback()
		c.JSON(http.StatusConflict, ErrorResponse{
			Error:   "Category name already exists",
			Code:    "DUPLICATE_ERROR",
			Details: "another service category with this name already exists",
		})
		return
	}

	category.Name = body.Name
	category.ProCost = body.ProCost
    category.RetailCost = body.RetailCost

	if err := tx.Save(&category); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to update service category",
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

	c.JSON(http.StatusOK, gin.H{"data": category, "message": "service category updated successfully"})
}

func (h *Handler) DeleteServiceCategory(c *gin.Context) {
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

	var category models.ServiceCategory
	if err := tx.First(&category, "id = ?", id); err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Service category not found",
				Code:    "NOT_FOUND",
				Details: err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to find service category",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
		}
		return
	}

	if err := tx.Delete(&category); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to delete service category",
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

	c.JSON(http.StatusOK, gin.H{"message": "service category deleted successfully"})
}