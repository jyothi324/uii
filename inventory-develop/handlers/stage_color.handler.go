package handlers

import (
	"errors"
	"inventory/middleware"
	"inventory/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func HandleSaveStageColors() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)
		if database == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection unavailable"})
			return
		}

		var req struct {
			Colors []struct {
				Stage string `json:"stage_name"`
				Color string `json:"color_hex"`
			} `json:"colors"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		if len(req.Colors) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "colors array is required"})
			return
		}

		for _, item := range req.Colors {
			if item.Stage == "" || item.Color == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "stage_name and color_hex are required for each color"})
				return
			}

			existing := models.StageColor{}
			err := database.Where("stage = ?", item.Stage).First(&existing)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					stageColor := models.StageColor{
						Stage: item.Stage,
						Color: item.Color,
					}
					if err := database.Create(&stageColor); err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save stage color"})
						return
					}
					continue
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query stage colors"})
				return
			}

			existing.Color = item.Color
			if err := database.Save(&existing); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update stage color"})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{"message": "Stage colors saved successfully"})
	}
}

func GetStageColors() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)
		if database == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection unavailable"})
			return
		}

		var colors []models.StageColor
		if err := database.Find(&colors); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stage colors"})
			return
		}

		c.JSON(http.StatusOK, colors)
	}
}
