package handlers

import (
	"errors"
	"inventory/middleware"
	"inventory/models"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func HandleListEmailTemplates() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := middleware.GetDB(c)
		if db == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection unavailable"})
			return
		}

		var templates []models.EmailTemplate
		if err := db.Model(&models.EmailTemplate{}).Order("created_at desc").Find(&templates); err != nil {
			log.Println("DB error listing email templates:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch templates"})
			return
		}

		c.JSON(http.StatusOK, templates)
	}
}

func HandleCreateEmailTemplate() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := middleware.GetDB(c)
		if db == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection unavailable"})
			return
		}

		var req struct {
			Name    string `json:"name"`
			Subject string `json:"subject"`
			Content string `json:"content"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		req.Name = stringsTrim(req.Name)
		if req.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
			return
		}

		// Check uniqueness per-tenant
		var existing models.EmailTemplate
		err := db.Model(&models.EmailTemplate{}).Where("lower(name) = lower(?)", req.Name).First(&existing)
		if err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "template name already exists"})
			return
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Println("DB error checking existing template:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
			return
		}

		tmpl := models.EmailTemplate{
			Name:    req.Name,
			Subject: req.Subject,
			Content: req.Content,
		}
		if err := db.Create(&tmpl); err != nil {
			log.Println("DB create error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create template"})
			return
		}

		c.JSON(http.StatusCreated, tmpl)
	}
}

func HandleUpdateEmailTemplate() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := middleware.GetDB(c)
		if db == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection unavailable"})
			return
		}
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		var req struct {
			Name    string `json:"name"`
			Subject string `json:"subject"`
			Content string `json:"content"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		tmpl := models.EmailTemplate{}
		if err := db.Model(&models.EmailTemplate{}).Where("id = ?", id).First(&tmpl); err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
				return
			}
			log.Println("DB error fetching template:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
			return
		}

		// If name changed, ensure uniqueness
		newName := stringsTrim(req.Name)
		if newName != "" && newName != tmpl.Name {
			var other models.EmailTemplate
			err := db.Model(&models.EmailTemplate{}).Where("lower(name) = lower(?) AND id != ?", newName, id).First(&other)
			if err == nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "template name already exists"})
				return
			}
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				log.Println("DB error checking name uniqueness:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
				return
			}
			tmpl.Name = newName
		}

		tmpl.Subject = req.Subject
		tmpl.Content = req.Content

		if err := db.Save(&tmpl); err != nil {
			log.Println("DB save error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update template"})
			return
		}

		c.JSON(http.StatusOK, tmpl)
	}
}

func HandleDeleteEmailTemplate() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := middleware.GetDB(c)
		if db == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection unavailable"})
			return
		}
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		if err := db.Model(&models.EmailTemplate{}).Where("id = ?", id).Delete(&models.EmailTemplate{}); err != nil {
			log.Println("DB delete error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete template"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	}
}

// helper: lightweight trim without importing strings repeatedly
func stringsTrim(s string) string {
	// simple manual trim to avoid extra imports in small handler file
	// but use strings package for robustness — implement minimal here
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\n' || s[0] == '\t' || s[0] == '\r') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\n' || s[len(s)-1] == '\t' || s[len(s)-1] == '\r') {
		s = s[:len(s)-1]
	}
	return s
}
