package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/tuliporg/agent-management-api/middleware"
	"github.com/tuliporg/agent-management-api/models"
)

type CreateServiceGroupInput struct {
	Name        string `json:"name" binding:"required"`
	ServiceIDs  []uint `json:"service_ids"` // optional
	Description string `json:"description"` // optional
}

type UpdateServiceGroupInput struct {
	Name        string `json:"name" binding:"required"`
	ServiceIDs  []uint `json:"service_ids"` // empty slice => clear
	Description string `json:"description"` // optional
}

// Response DTOs
type ServiceItem struct {
	ID      uint                  `json:"id"`
	Details models.ServiceDetails `json:"details"`
}

type ServiceGroupLite struct {
	ID          uint              `json:"id"`
	Name        string            `json:"name"`
	ServiceIDs  models.Int64Array `json:"service_ids"`
	Description string            `json:"description"`
}

type ServiceGroupWithServices struct {
	ID          uint              `json:"id"`
	Name        string            `json:"name"`
	ServiceIDs  models.Int64Array `json:"service_ids"`
	Services    []ServiceItem     `json:"services"`
	Description string            `json:"description"`
}

// CreateServiceGroup: save only id, name, service_ids
func (h *Handler) CreateServiceGroup(c *gin.Context) {
	db := middleware.GetDB(c)
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database not initialized"})
		return
	}

	var input CreateServiceGroupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	sg := models.ServiceGroup{Name: input.Name, Description: input.Description}

	if len(input.ServiceIDs) > 0 {
		seen := make(map[uint]struct{}, len(input.ServiceIDs))
		unique := make([]uint, 0, len(input.ServiceIDs))
		for _, id := range input.ServiceIDs {
			if _, ok := seen[id]; !ok {
				seen[id] = struct{}{}
				unique = append(unique, id)
			}
		}
		sg.ServiceIDs = make(models.Int64Array, 0, len(unique))
		for _, id := range unique {
			sg.ServiceIDs = append(sg.ServiceIDs, int64(id))
		}
	} else {
		sg.ServiceIDs = models.Int64Array{}
	}

	if err := tx.Create(&sg); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create service group"})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit transaction"})
		return
	}

	out := ServiceGroupLite{
		ID:          sg.ID,
		Name:        sg.Name,
		ServiceIDs:  sg.ServiceIDs,
		Description: sg.Description,
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "service group created",
		"data":    out,
	})
}

// GetAllServiceGroups: list groups and include details for services referenced in service_ids
func (h *Handler) GetAllServiceGroups(c *gin.Context) {
	db := middleware.GetDB(c)
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database not initialized"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "6"))
	all := c.Query("all") == "true"

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 6
	}

	var groups []ServiceGroupLite
	query := db.Model(&models.ServiceGroup{}).Select("id, name, service_ids, description")

	if all {
		if err := query.Find(&groups); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch service groups"})
			return
		}

		enriched := make([]ServiceGroupWithServices, 0, len(groups))

		// Collect unique service IDs
		idSet := make(map[uint]struct{})
		for _, g := range groups {
			for _, sid := range g.ServiceIDs {
				if sid > 0 {
					idSet[uint(sid)] = struct{}{}
				}
			}
		}
		uniqueIDs := make([]uint, 0, len(idSet))
		for id := range idSet {
			uniqueIDs = append(uniqueIDs, id)
		}

		// Load services (id, details)
		svcMap := make(map[uint]models.ServiceDetails, len(uniqueIDs))
		if len(uniqueIDs) > 0 {
			var services []models.Service
			if err := db.Model(&models.Service{}).
				Select("id, details").
				Where("id IN ?", uniqueIDs).
				Find(&services); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch services"})
				return
			}
			for _, s := range services {
				svcMap[s.ID] = s.Details
			}
		}

		// Build response
		for _, g := range groups {
			items := make([]ServiceItem, 0, len(g.ServiceIDs))
			for _, sid := range g.ServiceIDs {
				if d, ok := svcMap[uint(sid)]; ok {
					items = append(items, ServiceItem{
						ID:      uint(sid),
						Details: d,
					})
				}
			}
			enriched = append(enriched, ServiceGroupWithServices{
				ID:          g.ID,
				Name:        g.Name,
				ServiceIDs:  g.ServiceIDs,
				Services:    items,
				Description: g.Description,
			})
		}

		c.JSON(http.StatusOK, gin.H{"data": enriched})
		return
	}

	var total int64
	if err := db.Model(&models.ServiceGroup{}).Count(&total).Error(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count service groups"})
		return
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&groups); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch service groups"})
		return
	}

	enriched := make([]ServiceGroupWithServices, 0, len(groups))

	// Collect unique service IDs
	idSet := make(map[uint]struct{})
	for _, g := range groups {
		for _, sid := range g.ServiceIDs {
			if sid > 0 {
				idSet[uint(sid)] = struct{}{}
			}
		}
	}
	uniqueIDs := make([]uint, 0, len(idSet))
	for id := range idSet {
		uniqueIDs = append(uniqueIDs, id)
	}

	// Load services (id, details)
	svcMap := make(map[uint]models.ServiceDetails, len(uniqueIDs))
	if len(uniqueIDs) > 0 {
		var services []models.Service
		if err := db.Model(&models.Service{}).
			Select("id, details").
			Where("id IN ?", uniqueIDs).
			Find(&services); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch services"})
			return
		}
		for _, s := range services {
			svcMap[s.ID] = s.Details
		}
	}

	// Build response
	for _, g := range groups {
		items := make([]ServiceItem, 0, len(g.ServiceIDs))
		for _, sid := range g.ServiceIDs {
			if d, ok := svcMap[uint(sid)]; ok {
				items = append(items, ServiceItem{
					ID:      uint(sid),
					Details: d,
				})
			}
		}
		enriched = append(enriched, ServiceGroupWithServices{
			ID:          g.ID,
			Name:        g.Name,
			ServiceIDs:  g.ServiceIDs,
			Services:    items,
			Description: g.Description,
		})
	}

	totalPages := (int(total) + pageSize - 1) / pageSize

	c.JSON(http.StatusOK, gin.H{
		"data":        enriched,
		"page":        page,
		"page_size":   pageSize,
		"total":       total,
		"total_pages": totalPages,
	})
}

// UpdateServiceGroup: update only name and service_ids (no many2many)
func (h *Handler) UpdateServiceGroup(c *gin.Context) {
	db := middleware.GetDB(c)
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database not initialized"})
		return
	}

	id := c.Param("id")
	var input UpdateServiceGroupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var sg models.ServiceGroup
	if err := tx.First(&sg, id); err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "service group not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch service group"})
		return
	}

	sg.Name = input.Name
	sg.Description = input.Description

	// Deduplicate service_ids and store
	seen := make(map[uint]struct{}, len(input.ServiceIDs))
	unique := make([]uint, 0, len(input.ServiceIDs))
	for _, idVal := range input.ServiceIDs {
		if _, ok := seen[idVal]; !ok {
			seen[idVal] = struct{}{}
			unique = append(unique, idVal)
		}
	}

	sg.ServiceIDs = make(models.Int64Array, 0, len(unique))
	for _, idVal := range unique {
		sg.ServiceIDs = append(sg.ServiceIDs, int64(idVal))
	}

	if err := tx.Save(&sg); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update service group"})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit transaction"})
		return
	}

	out := ServiceGroupLite{
		ID:          sg.ID,
		Name:        sg.Name,
		ServiceIDs:  sg.ServiceIDs,
		Description: sg.Description,
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "service group updated",
		"data":    out,
	})
}

// DeleteServiceGroup
func (h *Handler) DeleteServiceGroup(c *gin.Context) {
	db := middleware.GetDB(c)
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database not initialized"})
		return
	}

	id := c.Param("id")

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var sg models.ServiceGroup
	if err := tx.First(&sg, id); err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "service group not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch service group"})
		return
	}

	if err := tx.Delete(&sg); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete service group"})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "service group deleted"})
}
