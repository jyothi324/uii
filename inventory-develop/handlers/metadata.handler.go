package handlers

import (
	"errors"
	"inventory/middleware"
	"inventory/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func resolve(typeParam string) (model models.MetadataModel, slice any, label string, ok bool) {
	switch typeParam {
	case "type":
		return &models.MaterialType{}, &[]models.MaterialType{}, "Material Type", true
	case "category":
		return &models.MaterialCategory{}, &[]models.MaterialCategory{}, "Material Category", true
	case "subcategory":
		return &models.MaterialSubCategory{}, &[]models.MaterialSubCategory{}, "Material SubCategory", true
	case "origin":
		return &models.MaterialOrigin{}, &[]models.MaterialOrigin{}, "Material Origin", true
	case "finish":
		return &models.MaterialFinish{}, &[]models.MaterialFinish{}, "Material Finish", true
	case "basecolor":
		return &models.MaterialBaseColor{}, &[]models.MaterialBaseColor{}, "Material Base Color", true
	case "level":
		return &models.MaterialLevel{}, &[]models.MaterialLevel{}, "Material Level", true
	}
	return nil, nil, "", false
}

func HandleCreateMetadata() gin.HandlerFunc {
	return func(c *gin.Context) {
		typeParam := c.Query("type")
		model, _, label, ok := resolve(typeParam)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or missing 'type' query param"})
			return
		}

		database := middleware.GetDB(c)
		if database == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection unavailable"})
			return
		}

		var req struct {
			ID   uint   `json:"id"`
			Name string `json:"name"`
			Unit string `json:"unit"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		now := time.Now()

		if req.ID == 0 {
			model.SetName(req.Name)
			model.SetCreatedAt(now)
			model.SetUpdatedAt(now)
			if mf, ok := model.(*models.MaterialType); ok {
				mf.Unit = req.Unit
			}

			if err := database.Create(model); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create " + label})
				return
			}
			c.JSON(http.StatusCreated, gin.H{"message": label + " created successfully"})
		} else {
			if err := database.First(model, req.ID); err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					c.JSON(http.StatusNotFound, gin.H{"error": label + " not found"})
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch " + label})
				return
			}

			model.SetName(req.Name)
			model.SetUpdatedAt(now)
			if mf, ok := model.(*models.MaterialType); ok {
				mf.Unit = req.Unit
			}

			if err := database.Save(model); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update " + label})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": label + " updated successfully"})
		}
	}
}

func GetAllMetadata() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, slice, label, ok := resolve(c.Query("type"))
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or missing 'type' query param"})
			return
		}

		database := middleware.GetDB(c)
		if database == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection unavailable"})
			return
		}

		if err := database.Order("updated_at desc").Find(slice); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch " + label})
			return
		}

		c.JSON(http.StatusOK, slice)
	}
}

func DeleteMetadataByID() gin.HandlerFunc {
	return func(c *gin.Context) {
		model, _, label, ok := resolve(c.Query("type"))
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or missing 'type' query param"})
			return
		}

		database := middleware.GetDB(c)
		if database == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection unavailable"})
			return
		}

		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing id parameter"})
			return
		}

		if err := database.First(model, id); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusOK, gin.H{"message": label + " not found", "data": []any{}})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch " + label})
			return
		}

		if err := database.Delete(model); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete " + label})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": label + " deleted successfully"})
	}
}
