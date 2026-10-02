package handlers

import (
	"errors"
	"inventory/middleware"
	"inventory/models"

	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type MessageTemplateInput struct {
	Name     string `json:"name" binding:"required"`
	Subject  string `json:"subject" binding:"required"`
	Content  string `json:"content" binding:"required"`
	Category string `json:"category" binding:"required"`
}

func isRecordNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return true
	}

	errText := strings.ToLower(err.Error())
	return strings.Contains(errText, "record not found") || strings.Contains(errText, "not found")
}

// HandleListMessageTemplates retrieves all message templates for the tenant
func HandleListMessageTemplates() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)
		var templates []models.MessageTemplate

		if err := database.Model(&models.MessageTemplate{}).Order("created_at desc").Find(&templates); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve templates"})
			return
		}

		if templates == nil {
			templates = []models.MessageTemplate{}
		}

		c.JSON(http.StatusOK, templates)
	}
}

// HandleCreateMessageTemplate creates a new message template
func HandleCreateMessageTemplate() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)
		var input MessageTemplateInput

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
			return
		}

		// Check for duplicate name
		var existing models.MessageTemplate
		if err := database.Model(&models.MessageTemplate{}).Where("lower(name) = lower(?)", input.Name).First(&existing); err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Template name already exists"})
			return
		} else if !isRecordNotFoundError(err) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error checking for existing template"})
			return
		}

		template := models.MessageTemplate{
			Name:     input.Name,
			Subject:  input.Subject,
			Content:  input.Content,
			Category: input.Category,
		}

		if err := database.Create(&template); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create template"})
			return
		}

		c.JSON(http.StatusCreated, template)
	}
}

// HandleUpdateMessageTemplate updates an existing message template
func HandleUpdateMessageTemplate() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
			return
		}

		var input MessageTemplateInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
			return
		}

		// Check for duplicate name
		var existing models.MessageTemplate
		if err := database.Model(&models.MessageTemplate{}).Where("lower(name) = lower(?) AND id != ?", input.Name, id).First(&existing); err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Template name already exists"})
			return
		} else if !isRecordNotFoundError(err) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error checking for existing template"})
			return
		}

		var template models.MessageTemplate
		if err := database.Model(&models.MessageTemplate{}).Where("id = ?", id).First(&template); err != nil {
			if isRecordNotFoundError(err) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find template"})
			}
			return
		}

		template.Name = strings.TrimSpace(input.Name)
		template.Subject = input.Subject
		template.Content = input.Content
		template.Category = input.Category

		if err := database.Save(&template); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update template"})
			return
		}

		c.JSON(http.StatusOK, template)
	}
}

// HandleDeleteMessageTemplate deletes a message template
func HandleDeleteMessageTemplate() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
			return
		}

		var template models.MessageTemplate
		if err := database.Model(&models.MessageTemplate{}).Where("id = ?", id).Take(&template); err != nil {
			if isRecordNotFoundError(err) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find template"})
			}
			return
		}

		if err := database.Delete(&template); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete template"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Template deleted successfully"})
	}
}
