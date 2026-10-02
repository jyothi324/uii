package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/tuliporg/agent-management-api/db"        
	"github.com/gin-gonic/gin"
	"github.com/tuliporg/agent-management-api/middleware"
	"github.com/tuliporg/agent-management-api/models"
	"gorm.io/gorm"
)

// refreshServiceCategories replaces the category snapshot stored in each
// service's Details.ServiceCategory slice with the current live data from
// the service_categories table, so price updates are always reflected.
func (h *Handler) refreshServiceCategories(conn db.Database, services []models.Service) error {
    idSet := map[uint]struct{}{}
    for _, svc := range services {
        for _, cat := range svc.Details.ServiceCategory {
            if cat.ID != 0 {
                idSet[cat.ID] = struct{}{}
            }
        }
    }
    if len(idSet) == 0 {
        return nil
    }

    ids := make([]uint, 0, len(idSet))
    for id := range idSet {
        ids = append(ids, id)
    }

    var freshCategories []models.ServiceCategory
    if err := conn.Where("id IN ?", ids).Find(&freshCategories); err != nil {
        return err
    }

    catMap := make(map[uint]models.ServiceCategory, len(freshCategories))
    for _, cat := range freshCategories {
        catMap[cat.ID] = cat
    }

    for i := range services {
        for j, cat := range services[i].Details.ServiceCategory {
            if fresh, ok := catMap[cat.ID]; ok {
                services[i].Details.ServiceCategory[j] = fresh
            }
        }
    }
    return nil
}

// CreateService creates a new service
func (h *Handler) CreateService(c *gin.Context) {
	db := middleware.GetDB(c)

	if db == nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Database connection not found",
			Code:    "DATABASE_ERROR",
			Details: "DB instance is nil",
		})
		return
	}

	var service models.Service
	if err := c.ShouldBindJSON(&service); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Code:    "VALIDATION_ERROR",
			Details: err.Error(),
		})
		return
	}

	validatedCategories := []models.ServiceCategory{}
	for _, cat := range service.Details.ServiceCategory {
		if cat.ID == 0 {
			continue
		}
		var category models.ServiceCategory
		if err := db.First(&category, "id = ?", cat.ID); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Invalid Service Category ID",
				Code:    "VALIDATION_ERROR",
				Details: fmt.Sprintf("Service Category ID %d not found", cat.ID),
			})
			return
		}
		validatedCategories = append(validatedCategories, category)
	}
	service.Details.ServiceCategory = validatedCategories

	if service.Details.ServiceGroup.ID != 0 {
		var group models.ServiceGroup
		if err := db.First(&group, "id = ?", service.Details.ServiceGroup.ID); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Invalid Service Group ID",
				Code:    "VALIDATION_ERROR",
				Details: "Service Group not found",
			})
			return
		}
		service.Details.ServiceGroup = group
	}

	// if service.Details.ServiceType.ID != 0 {
	// 	var stype models.ServiceType
	// 	if err := db.First(&stype, "id = ?", service.Details.ServiceType.ID); err != nil {
	// 		c.JSON(http.StatusBadRequest, ErrorResponse{
	// 			Error:   "Invalid Service Type ID",
	// 			Code:    "VALIDATION_ERROR",
	// 			Details: "Service Type not found",
	// 		})
	// 		return
	// 	}
	// 	service.Details.ServiceType = stype
	// }

	tx := db.Begin()
	if err := tx.Create(&service); err != nil {
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

	c.JSON(http.StatusCreated, service)
}

// GetAllServices retrieves all services
func (h *Handler) GetAllServices(c *gin.Context) {
	db := middleware.GetDB(c)

	all := c.Query("all")
	search := strings.TrimSpace(c.Query("search"))

	page, pageSize := h.getPaginationParams(c)
	offset := (page - 1) * pageSize

	var services []models.Service
	var total int64

	serviceQuery := db.Model(&models.Service{})

	if search != "" {
		serviceQuery = serviceQuery.Where(
			"LOWER(details->>'Name') LIKE ?",
			"%"+strings.ToLower(search)+"%",
		)
	}

	serviceQuery.Count(&total)
	if err := serviceQuery.Error(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to count services",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	if all == "true" {
		if err := serviceQuery.Find(&services); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to retrieve all services",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
			return
		}
		// Refresh category prices from DB
		if err := h.refreshServiceCategories(db, services); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to refresh service categories",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"data":  services,
			"total": total,
			"all":   true,
		})
		return
	}

	if err := serviceQuery.Offset(offset).Limit(pageSize).Find(&services); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to retrieve services",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	// Refresh category prices from DB
	if err := h.refreshServiceCategories(db, services); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to refresh service categories",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	c.JSON(http.StatusOK, gin.H{
		"data":        services,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

// UpdateService updates a service by ID
func (h *Handler) UpdateService(c *gin.Context) {
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

	var updatedService models.Service
	if err := c.ShouldBindJSON(&updatedService); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Code:    "VALIDATION_ERROR",
			Details: err.Error(),
		})
		return
	}

	tx := db.Begin()

	var service models.Service
	if err := tx.First(&service, "id = ?", id); err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Service not found",
				Code:    "NOT_FOUND",
				Details: err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to find service",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
		}
		return
	}

	if updatedService.Details.ServiceGroup.ID != 0 {
		var group models.ServiceGroup
		if err := tx.First(&group, "id = ?", updatedService.Details.ServiceGroup.ID); err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Invalid Service Group ID",
				Code:    "VALIDATION_ERROR",
				Details: "Service Group not found",
			})
			return
		}
		updatedService.Details.ServiceGroup = group
	}

	validatedCategories := []models.ServiceCategory{}
	for _, cat := range updatedService.Details.ServiceCategory {
		if cat.ID == 0 {
			continue
		}
		var category models.ServiceCategory
		if err := tx.First(&category, "id = ?", cat.ID); err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Invalid Service Category ID",
				Code:    "VALIDATION_ERROR",
				Details: fmt.Sprintf("Service Category ID %d not found", cat.ID),
			})
			return
		}
		validatedCategories = append(validatedCategories, category)
	}
	updatedService.Details.ServiceCategory = validatedCategories

	// if updatedService.Details.ServiceType.ID != 0 {
	// 	var stype models.ServiceType
	// 	if err := tx.First(&stype, "id = ?", updatedService.Details.ServiceType.ID); err != nil {
	// 		tx.Rollback()
	// 		c.JSON(http.StatusBadRequest, ErrorResponse{
	// 			Error:   "Invalid Service Type ID",
	// 			Code:    "VALIDATION_ERROR",
	// 			Details: "Service Type not found",
	// 		})
	// 		return
	// 	}
	// 	updatedService.Details.ServiceType = stype
	// }

	service.Details = updatedService.Details

	if err := tx.Save(&service); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to update service",
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

	c.JSON(http.StatusOK, service)
}

// DeleteService deletes a service by ID
func (h *Handler) DeleteService(c *gin.Context) {
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

	var service models.Service
	if err := tx.First(&service, "id = ?", id); err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Service not found",
				Code:    "NOT_FOUND",
				Details: err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to find service",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
		}
		return
	}

	if err := tx.Delete(&service); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to delete service",
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
		"message": "service deleted successfully",
	})
}