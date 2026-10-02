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



func HandleListJobFormQuestions() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)
		if database == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection unavailable"})
			return
		}

		stage := strings.TrimSpace(c.Query("stage"))
		category := strings.TrimSpace(c.Query("category"))

		query := database.Model(&models.JobFormQuestion{})
		if stage != "" {
			query = query.Where("stage = ?", stage)
		}
		if category != "" {
			query = query.Where("category = ?", category)
		}

		var questions []models.JobFormQuestion
		if err := query.Order("sort_order ASC, id ASC").Find(&questions); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch questions"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": questions})
	}
}

type saveJobFormQuestionRequest struct {
	Stage     string `json:"stage"`
	Category  string `json:"category"`
	Text      string `json:"text"`
	SortOrder int    `json:"sort_order"`
}

func HandleCreateJobFormQuestion() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)
		if database == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection unavailable"})
			return
		}

		var req saveJobFormQuestionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		req.Stage = strings.TrimSpace(req.Stage)
		req.Category = strings.TrimSpace(req.Category)
		req.Text = strings.TrimSpace(req.Text)

		if req.Stage == "" || req.Category == "" || req.Text == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "stage, category and text are required"})
			return
		}

		question := models.JobFormQuestion{
			Stage:     req.Stage,
			Category:  req.Category,
			Text:      req.Text,
			SortOrder: req.SortOrder,
		}

		if err := database.Create(&question); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create question"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Question created successfully",
			"data":    question,
		})
	}
}

func HandleUpdateJobFormQuestion() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)
		if database == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection unavailable"})
			return
		}

		id := strings.TrimSpace(c.Param("id"))
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
			return
		}
		questionID, err := strconv.Atoi(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		question := models.JobFormQuestion{}
		if err := database.Model(&models.JobFormQuestion{}).
			Where("id = ?", questionID).
			First(&question); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query question"})
			return
		}

		var req saveJobFormQuestionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		req.Text = strings.TrimSpace(req.Text)
		if req.Text == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "text is required"})
			return
		}

		question.Text = req.Text
		question.SortOrder = req.SortOrder

		if err := database.Save(&question); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update question"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Question updated successfully",
			"data":    question,
		})
	}
}
func HandleDeleteJobFormQuestion() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)
		if database == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection unavailable"})
			return
		}

		id := strings.TrimSpace(c.Param("id"))
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
			return
		}
		questionID, err := strconv.Atoi(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		question := models.JobFormQuestion{}
		if err := database.Model(&models.JobFormQuestion{}).
			Where("id = ?", questionID).
			First(&question); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query question"})
			return
		}

		if err := database.Delete(&question); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete question"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Question deleted successfully",
		})
	}
}





func HandleGetJobFormSectionTitles() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)
		if database == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection unavailable"})
			return
		}

		stage := strings.TrimSpace(c.Query("stage"))
		if stage == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "stage is required"})
			return
		}

		titles := models.JobFormSectionTitles{}
		err := database.Model(&models.JobFormSectionTitles{}).
			Where("stage = ?", stage).
			First(&titles)

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusOK, gin.H{"data": models.JobFormSectionTitles{
					Stage:                    stage,
					MaterialsTitle:           "MATERIALS",
					ServicesTitle:            "SERVICES",
					FabricatorChecklistTitle: "Fabricator's checklist After Fabrication Complete (circle one)",
				}})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch titles"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": titles})
	}
}

type saveJobFormSectionTitlesRequest struct {
	Stage                    string `json:"stage"`
	MaterialsTitle           string `json:"materials_title"`
	ServicesTitle            string `json:"services_title"`
	FabricatorChecklistTitle string `json:"fabricator_checklist_title"`
}

func HandleUpsertJobFormSectionTitles() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)
		if database == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection unavailable"})
			return
		}

		var req saveJobFormSectionTitlesRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		req.Stage = strings.TrimSpace(req.Stage)
		if req.Stage == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "stage is required"})
			return
		}

		titles := models.JobFormSectionTitles{}
		err := database.Model(&models.JobFormSectionTitles{}).
			Where("stage = ?", req.Stage).
			First(&titles)

		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query titles"})
				return
			}
			titles = models.JobFormSectionTitles{
				Stage:                    req.Stage,
				MaterialsTitle:           req.MaterialsTitle,
				ServicesTitle:            req.ServicesTitle,
				FabricatorChecklistTitle: req.FabricatorChecklistTitle,
			}
			if err := database.Create(&titles); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create titles"})
				return
			}
			c.JSON(http.StatusCreated, gin.H{"message": "Titles created successfully", "data": titles})
			return
		}

		titles.MaterialsTitle = req.MaterialsTitle
		titles.ServicesTitle = req.ServicesTitle
		titles.FabricatorChecklistTitle = req.FabricatorChecklistTitle

		if err := database.Save(&titles); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update titles"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Titles updated successfully", "data": titles})
	}
}