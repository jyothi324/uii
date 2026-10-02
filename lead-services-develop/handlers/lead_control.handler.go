package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tuliporg/agent-management-api/middleware"
	"github.com/tuliporg/agent-management-api/models"
	"gorm.io/gorm"
)

const (
	ctlDeliveryType        = "deliverytype"
	ctlLeadSource          = "leadsource"
	ctlAssociateType       = "associatetype"
	ctlResidentialType     = "residentialtype"
	ctlResidentialCategory = "residentialcategory"
)

type leadControlUpsertRequest struct {
	ID      uint                    `json:"id"`
	Name    string                  `json:"name"`
	Type    string                  `json:"type"`
	Address *models.CustomerAddress `json:"address"`
	IsPickup *bool 					`json:"isPickup"`
}

func (h *Handler) GetLeadControlAll(c *gin.Context) {
	db := middleware.GetDB(c)
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database not found in context"})
		return
	}

	switch c.Query("type") {
	case ctlResidentialType:
		var rows []models.ManagerResidentialType
		if err := db.Order("id desc").Find(&rows); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch types"})
			return
		}
		out := make([]gin.H, 0, len(rows))
		for _, r := range rows {
			out = append(out, gin.H{"ID": r.ID, "type": r.ResidentialEntryType, "length": r.Length})
		}
		c.JSON(http.StatusOK, out)

	case ctlResidentialCategory:
		var rows []models.ManagerResidentialEntry
		if err := db.Order("id desc").Find(&rows); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch categories"})
			return
		}
		out := make([]gin.H, 0, len(rows))
		for _, r := range rows {
			out = append(out, gin.H{"ID": r.ID, "name": r.ResidentialEntryName, "length": r.Length})
		}
		c.JSON(http.StatusOK, out)

	case ctlDeliveryType:
		var rows []models.DeliveryTypeOption
		if err := db.Order("id desc").Find(&rows); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch delivery types"})
			return
		}
		c.JSON(http.StatusOK, rows)

	case ctlLeadSource:
		var rows []models.LeadSourceOption
		if err := db.Order("id desc").Find(&rows); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch lead sources"})
			return
		}
		c.JSON(http.StatusOK, rows)

	case ctlAssociateType:
		var rows []models.AssociateTypeOption
		if err := db.Order("id desc").Find(&rows); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch associate types"})
			return
		}
		c.JSON(http.StatusOK, rows)

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown type"})
	}
}

func (h *Handler) CreateOrUpdateLeadControl(c *gin.Context) {
	db := middleware.GetDB(c)
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database not found in context"})
		return
	}

	typeParam := c.Query("type")
	var req leadControlUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	switch typeParam {
	case ctlResidentialType:
		name := strings.TrimSpace(req.Type)
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "type is required"})
			return
		}
		if req.ID != 0 {
			var row models.ManagerResidentialType
			if err := db.First(&row, req.ID); err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}
			row.ResidentialEntryType = name
			db.Save(&row)
			c.JSON(http.StatusOK, gin.H{"ID": row.ID, "type": row.ResidentialEntryType, "length": row.Length})
			return
		}
		row := models.ManagerResidentialType{ResidentialEntryType: name}
		if err := db.Create(&row); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"ID": row.ID, "type": row.ResidentialEntryType, "length": row.Length})

	case ctlResidentialCategory:
		name := strings.TrimSpace(req.Name)
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
			return
		}
		if req.ID != 0 {
			var row models.ManagerResidentialEntry
			if err := db.First(&row, req.ID); err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}
			row.ResidentialEntryName = name
			db.Save(&row)
			c.JSON(http.StatusOK, gin.H{"ID": row.ID, "name": row.ResidentialEntryName, "length": row.Length})
			return
		}
		row := models.ManagerResidentialEntry{ResidentialEntryName: name}
		if err := db.Create(&row); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"ID": row.ID, "name": row.ResidentialEntryName, "length": row.Length})

	case ctlDeliveryType:
		name := strings.TrimSpace(req.Name)
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
			return
		}
		if req.ID != 0 {
			var row models.DeliveryTypeOption
			if err := db.First(&row, req.ID); err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}
			row.Name = name
			if req.Address != nil {
				row.Address = *req.Address
			}
			if req.IsPickup != nil {
				row.IsPickup = req.IsPickup
			}
			db.Save(&row)
			c.JSON(http.StatusOK, row)
			return
		}
		var count int64
		db.Model(&models.DeliveryTypeOption{}).Where("LOWER(name) = LOWER(?)", name).Count(&count)
		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "already exists"})
			return
		}
		row := models.DeliveryTypeOption{Name: name}
		if req.Address != nil {
			row.Address = *req.Address
		}
		if req.IsPickup != nil {
			row.IsPickup = req.IsPickup
		}

		if err := db.Create(&row); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create"})
			return
		}
		c.JSON(http.StatusCreated, row)

	case ctlLeadSource:
		name := strings.TrimSpace(req.Name)
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
			return
		}
		if req.ID != 0 {
			var row models.LeadSourceOption
			if err := db.First(&row, req.ID); err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}
			row.Name = name
			db.Save(&row)
			c.JSON(http.StatusOK, row)
			return
		}
		var count int64
		db.Model(&models.LeadSourceOption{}).Where("LOWER(name) = LOWER(?)", name).Count(&count)
		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "already exists"})
			return
		}
		row := models.LeadSourceOption{Name: name}
		if err := db.Create(&row); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create"})
			return
		}
		c.JSON(http.StatusCreated, row)

	case ctlAssociateType:
		name := strings.TrimSpace(req.Name)
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
			return
		}
		if req.ID != 0 {
			var row models.AssociateTypeOption
			if err := db.First(&row, req.ID); err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}
			row.Name = name
			db.Save(&row)
			c.JSON(http.StatusOK, row)
			return
		}
		var count int64
		db.Model(&models.AssociateTypeOption{}).Where("LOWER(name) = LOWER(?)", name).Count(&count)
		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "already exists"})
			return
		}
		row := models.AssociateTypeOption{Name: name}
		if err := db.Create(&row); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create"})
			return
		}
		c.JSON(http.StatusCreated, row)

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown type"})
	}
}

func (h *Handler) DeleteLeadControl(c *gin.Context) {
	db := middleware.GetDB(c)
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database not found in context"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	switch c.Query("type") {
	case ctlResidentialType:
		var row models.ManagerResidentialType
		if err = db.First(&row, id); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete"})
			return
		}
		if row.Length > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "Residential type is already associated with estimate"})
			return
		}
		err = db.Delete(&models.ManagerResidentialType{}, id)
	case ctlResidentialCategory:
		var row models.ManagerResidentialEntry
		if err = db.First(&row, id); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete"})
			return
		}
		if row.Length > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "Residential category is already associated with estimate"})
			return
		}
		err = db.Delete(&models.ManagerResidentialEntry{}, id)
	case ctlDeliveryType:
		err = db.Delete(&models.DeliveryTypeOption{}, id)
	case ctlLeadSource:
		err = db.Delete(&models.LeadSourceOption{}, id)
	case ctlAssociateType:
		err = db.Delete(&models.AssociateTypeOption{}, id)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown type"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete"})
		return
	}
	c.Status(http.StatusNoContent)
}
