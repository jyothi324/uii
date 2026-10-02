package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"inventory/middleware"
	"inventory/models"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// type saveJobFormTemplateRequest struct {
// 	Stage        string          `json:"stage"`
// 	Name         string          `json:"name"`
// 	Description  string          `json:"description"`
// 	TemplateData json.RawMessage `json:"template_data"`
// 	IsActive     *bool           `json:"is_active"`
// }

type saveJobTaskFormRequest struct {
	JobID            string          `json:"job_id"`
	TaskID           string          `json:"task_id"`
	TaskName         string          `json:"task_name"`
	Stage            string          `json:"stage"`
	SalesPerson      string          `json:"salesperson"`
	FormData         json.RawMessage `json:"form_data"`
	TemplateSnapshot json.RawMessage `json:"template_snapshot"`
	TaskSnapshot     json.RawMessage `json:"task_snapshot"`
}

func normalizeJSON(raw json.RawMessage) datatypes.JSON {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return datatypes.JSON([]byte("{}"))
	}
	return datatypes.JSON(trimmed)
}

func isBlankJSON(raw json.RawMessage) bool {
	trimmed := bytes.TrimSpace(raw)
	return len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null"))
}

// func HandleUpsertJobFormTemplate() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		database := middleware.GetDB(c)
// 		if database == nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection unavailable"})
// 			return
// 		}

// 		var req saveJobFormTemplateRequest
// 		if err := c.ShouldBindJSON(&req); err != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
// 			return
// 		}

// 		req.Stage = strings.TrimSpace(req.Stage)
// 		req.Name = strings.TrimSpace(req.Name)
// 		req.Description = strings.TrimSpace(req.Description)

// 		if req.Stage == "" {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "stage is required"})
// 			return
// 		}
// 		if isBlankJSON(req.TemplateData) {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "template_data is required"})
// 			return
// 		}
// 		if req.Name == "" {
// 			req.Name = req.Stage + " Form"
// 		}

// 		template := models.JobFormTemplate{}
// 		err := database.Model(&models.JobFormTemplate{}).
// 			Where("stage = ?", req.Stage).
// 			First(&template)

// 		if err != nil {
// 			if !errors.Is(err, gorm.ErrRecordNotFound) {
// 				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query form template"})
// 				return
// 			}

// 			isActive := true
// 			if req.IsActive != nil {
// 				isActive = *req.IsActive
// 			}

// 			template = models.JobFormTemplate{
// 				Stage:        req.Stage,
// 				Name:         req.Name,
// 				Description:  req.Description,
// 				TemplateData: normalizeJSON(req.TemplateData),
// 				IsActive:     isActive,
// 			}

// 			if err := database.Create(&template); err != nil {
// 				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create form template"})
// 				return
// 			}

// 			c.JSON(http.StatusCreated, gin.H{
// 				"message": "Form template created successfully",
// 				"data":    template,
// 			})
// 			return
// 		}

// 		template.Name = req.Name
// 		template.Description = req.Description
// 		template.TemplateData = normalizeJSON(req.TemplateData)
// 		if req.IsActive != nil {
// 			template.IsActive = *req.IsActive
// 		}

// 		if err := database.Save(&template); err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update form template"})
// 			return
// 		}

// 		c.JSON(http.StatusOK, gin.H{
// 			"message": "Form template updated successfully",
// 			"data":    template,
// 		})
// 	}
// }

// func HandleListJobFormTemplates() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		database := middleware.GetDB(c)
// 		if database == nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection unavailable"})
// 			return
// 		}

// 		stage := strings.TrimSpace(c.Query("stage"))
// 		query := database.Model(&models.JobFormTemplate{})
// 		if stage != "" {
// 			query = query.Where("stage = ?", stage)
// 		}

// 		var templates []models.JobFormTemplate
// 		if err := query.Order("stage ASC").Find(&templates); err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch form templates"})
// 			return
// 		}

// 		c.JSON(http.StatusOK, templates)
// 	}
// }

// func HandleGetJobFormTemplateByStage() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		database := middleware.GetDB(c)
// 		if database == nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection unavailable"})
// 			return
// 		}

// 		stage := strings.TrimSpace(c.Param("stage"))
// 		if stage == "" {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "stage is required"})
// 			return
// 		}

// 		template := models.JobFormTemplate{}
// 		err := database.Model(&models.JobFormTemplate{}).
// 			Where("stage = ?", stage).
// 			First(&template)
// 		if err != nil {
// 			if errors.Is(err, gorm.ErrRecordNotFound) {
// 				c.JSON(http.StatusNotFound, gin.H{"error": "Form template not found"})
// 				return
// 			}
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch form template"})
// 			return
// 		}

// 		c.JSON(http.StatusOK, template)
// 	}
// }

func HandleUpsertJobTaskForm() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)
		if database == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection unavailable"})
			return
		}

		var req saveJobTaskFormRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		req.JobID = strings.TrimSpace(req.JobID)
		req.TaskID = strings.TrimSpace(req.TaskID)
		req.TaskName = strings.TrimSpace(req.TaskName)
		req.Stage = strings.TrimSpace(req.Stage)
		req.SalesPerson = strings.TrimSpace(req.SalesPerson)

		if req.JobID == "" || req.TaskID == "" || req.Stage == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "job_id, task_id and stage are required"})
			return
		}

		existing := models.JobTaskForm{}
		err := database.Model(&models.JobTaskForm{}).
			Where("job_id = ? AND task_id = ? AND stage = ?", req.JobID, req.TaskID, req.Stage).
			First(&existing)

		if err == nil {
			existing.TaskName = req.TaskName
			existing.SalesPerson = req.SalesPerson
			if !isBlankJSON(req.FormData) {
				existing.FormData = normalizeJSON(req.FormData)
			}
			if !isBlankJSON(req.TemplateSnapshot) {
				existing.TemplateSnapshot = normalizeJSON(req.TemplateSnapshot)
			}
			if !isBlankJSON(req.TaskSnapshot) {
				existing.TaskSnapshot = normalizeJSON(req.TaskSnapshot)
			}

			if err := database.Save(&existing); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task form"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"message": "Task form updated successfully",
				"data":    existing,
			})
			return
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query task form"})
			return
		}

templateSnapshot := normalizeJSON(req.TemplateSnapshot)

		formData := normalizeJSON(req.FormData)
		if isBlankJSON(req.FormData) && len(templateSnapshot) > 0 {
			formData = templateSnapshot
		}

		taskForm := models.JobTaskForm{
			JobID:            req.JobID,
			TaskID:           req.TaskID,
			TaskName:         req.TaskName,
			Stage:            req.Stage,
			SalesPerson:      req.SalesPerson,
			FormData:         formData,
			TemplateSnapshot: templateSnapshot,
			TaskSnapshot:     normalizeJSON(req.TaskSnapshot),
		}

		if err := database.Create(&taskForm); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task form"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Task form created successfully",
			"data":    taskForm,
		})
	}
}

func HandleListJobTaskForms() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)
		if database == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection unavailable"})
			return
		}

		jobID := strings.TrimSpace(c.Query("job_id"))
		taskID := strings.TrimSpace(c.Query("task_id"))
		stage := strings.TrimSpace(c.Query("stage"))

		query := database.Model(&models.JobTaskForm{})
		if jobID != "" {
			query = query.Where("job_id = ?", jobID)
		}
		if taskID != "" {
			query = query.Where("task_id = ?", taskID)
		}
		if stage != "" {
			query = query.Where("stage = ?", stage)
		}

		var forms []models.JobTaskForm
		if err := query.Order("updated_at DESC").Find(&forms); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch task forms"})
			return
		}

		c.JSON(http.StatusOK, forms)
	}
}

// func HandleGetJobTaskFormByID() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		database := middleware.GetDB(c)
// 		if database == nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection unavailable"})
// 			return
// 		}

// 		id := strings.TrimSpace(c.Param("id"))
// 		if id == "" {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
// 			return
// 		}

// 		form := models.JobTaskForm{}
// 		err := database.Model(&models.JobTaskForm{}).First(&form, "id = ?", id)
// 		if err != nil {
// 			if errors.Is(err, gorm.ErrRecordNotFound) {
// 				c.JSON(http.StatusNotFound, gin.H{"error": "Task form not found"})
// 				return
// 			}
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch task form"})
// 			return
// 		}

// 		c.JSON(http.StatusOK, form)
// 	}
// }
