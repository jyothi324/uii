// discounts
package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tuliporg/agent-management-api/internal/activity"
	"github.com/tuliporg/agent-management-api/middleware"
	"github.com/tuliporg/agent-management-api/models"
	"github.com/tuliporg/agent-management-api/services"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func (h *Handler) CreateTask(c *gin.Context) {
	db := middleware.GetDB(c)

	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database not initialized",
		})
		return
	}

	var request struct {
		models.Task
		JobID              *string                   `json:"jobId"`
		ResidentialEntryID *uint                     `json:"residentialEntryId"`
		StageName          *string                   `json:"stageName"`
		SalesPersonID      *string                   `json:"SalesPersonID"`
		StartTime          *string                   `json:"startTime"`
		EndTime            *string                   `json:"endTime"`
		Duration           *int                      `json:"duration"`
		Projects           []models.StageTaskProject `json:"projects"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid input",
			Code:    "VALIDATION_ERROR",
			Details: err.Error(),
		})
		return
	}

	tx := db.Begin()

	// ================================
	// IF JOB EXISTS → CREATE STAGE TASK ONLY
	// ================================
	if request.JobID != nil && *request.JobID != "" {

		if request.ResidentialEntryID != nil &&
			request.StageName != nil &&
			*request.StageName != "" {

			var stage models.Stage

			if err := tx.First(&stage,
				"residential_entry_id = ? AND name = ?",
				*request.ResidentialEntryID,
				*request.StageName); err == nil {

				parsedTime, err := time.Parse(time.RFC3339, request.DueDate)
				if err != nil {
					tx.Rollback()
					c.JSON(http.StatusBadRequest, ErrorResponse{
						Error:   "Invalid start date format",
						Code:    "VALIDATION_ERROR",
						Details: err.Error(),
					})
					return
				}

				startDate := parsedTime.UTC()
				startTime := parsedTime.UTC()

				// ✅ Determine duration and end time
				durationMinutes := 240
				if request.Duration != nil && *request.Duration > 0 {
					durationMinutes = *request.Duration
				}

				endTime := startTime.Add(time.Duration(durationMinutes) * time.Minute)

				// ✅ Use explicitly provided StartTime and EndTime if available
				if request.StartTime != nil && *request.StartTime != "" {
					if st, err := time.Parse(time.RFC3339, *request.StartTime); err == nil {
						startTime = st.UTC()
						startDate = startTime // Keep startDate in sync
					}
				}
				if request.EndTime != nil && *request.EndTime != "" {
					if et, err := time.Parse(time.RFC3339, *request.EndTime); err == nil {
						endTime = et.UTC()
						durationMinutes = int(endTime.Sub(startTime).Minutes())
					}
				}

				stageTask := models.StageTask{
					ID:        uuid.NewString(),
					StageID:   stage.ID,
					Name:      request.Title,
					DueDate:   endTime,
					StartDate: startDate,
					StartTime: &startTime,
					EndTime:   &endTime,
					Duration:  durationMinutes,
					AssignedTo: func() string {
						if request.SalesPersonID != nil && *request.SalesPersonID != "" {
							return *request.SalesPersonID
						}
						return request.AssigneeID
					}(),
					CreatedAt: time.Now().UTC(),
					UpdatedAt: time.Now().UTC(),
					Projects: func() datatypes.JSON {
						projectsJSON, _ := json.Marshal(request.Projects)
						return datatypes.JSON(projectsJSON)
					}(),
				}

				if err := tx.Create(&stageTask); err != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, ErrorResponse{
						Error:   "Failed to create stage task",
						Code:    "DATABASE_ERROR",
						Details: err.Error(),
					})
					return
				}

				// Activity Log
				userID, userName := middleware.GetRequestUser(c)
				event := activity.ActivityEvent{
					EntityType: "job",
					EntityID:   *request.JobID,
					Action:     fmt.Sprintf("task %s created", strings.ToLower(stageTask.Name)),
					UserID:     userID,
					UserName:   userName,
					Payload: map[string]any{
						"task_name":   stageTask.Name,
						"stage_name":  *request.StageName,
						"created_by":  userName,
						"assigned_to": stageTask.AssignedTo,
						"start_time":  stageTask.StartTime,
					},
				}

				if err := activity.InsertActivityEvent(c.Request.Context(), tx.GetUnderlying(), event); err != nil {
					log.Printf("activity event publish failed: %v", err)
				}

				// Refresh Jobs JSON
				if stage.ID != "" && stage.LeadID != "" {

					var lead models.Lead

					if err := tx.First(&lead, "id = ?", stage.LeadID); err == nil {
						jobService := services.NewJobService()
						_ = jobService.UpsertJobFromLead(tx, lead)
					}
				}
			}
		}

		// IMPORTANT: Stop here so calendar task is NOT created
		if err := tx.Commit(); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to commit transaction",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Stage task created successfully",
		})
		return
	}

	if request.EndTime != nil && *request.EndTime != "" {
		request.Task.EndTime = *request.EndTime
	}

	// ================================
	// IF NO JOB → CREATE CALENDAR TASK
	// ================================
	request.Task.ID = 0
	if err := tx.Create(&request.Task); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to create task",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	// Activity Log for Calendar Task
	userID, userName := middleware.GetRequestUser(c)
	event := activity.ActivityEvent{
		EntityType: "task",
		EntityID:   fmt.Sprintf("%d", request.Task.ID),
		Action:     "created",
		UserID:     userID,
		UserName:   userName,
		Payload: map[string]any{
			"task_name":   request.Task.Title,
			"created_by":  userName,
			"assigned_to": request.Task.AssigneeID,
		},
	}
	if err := activity.InsertActivityEvent(c.Request.Context(), tx.GetUnderlying(), event); err != nil {
		log.Printf("activity event publish failed: %v", err)
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

	c.JSON(http.StatusCreated, gin.H{
		"message": "Task created successfully",
		"data":    request.Task,
	})
}

// UpdateTask updates a Task by ID
func (h *Handler) UpdateTask(c *gin.Context) {
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

	var input models.Task
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid input",
			Code:    "VALIDATION_ERROR",
			Details: err.Error(),
		})
		return
	}

	tx := db.Begin()
	var existing models.Task
	if err := tx.First(&existing, "id = ?", id); err != nil {
		tx.Rollback()
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Task not found",
				Code:    "NOT_FOUND",
				Details: err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to find task",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
		}
		return
	}

	// Update only the status field
	if input.Status != "" {
		existing.Status = input.Status
	}
	// Update CompletedAt if status is being set to completed
	if input.Status == "completed" && input.CompletedAt != nil {
		existing.CompletedAt = input.CompletedAt
	}

	if input.EndTime != "" {
		existing.EndTime = input.EndTime
	}

	if input.Photos != nil {
		photosJSON, _ := json.Marshal(input.Photos)
		existing.Photos = datatypes.JSON(photosJSON)
	}

	if err := tx.Save(&existing); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to update task",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	// Activity Log
	userID, userName := middleware.GetRequestUser(c)
	event := activity.ActivityEvent{
		EntityType: "task",
		EntityID:   id,
		Action:     "updated",
		UserID:     userID,
		UserName:   userName,
		Payload: map[string]any{
			"task_name":   existing.Title,
			"new_status":  existing.Status,
			"updated_by":  userName,
			"assigned_to": existing.AssigneeID,
		},
	}
	if err := activity.InsertActivityEvent(c.Request.Context(), tx.GetUnderlying(), event); err != nil {
		log.Printf("activity event publish failed: %v", err)
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
		"message": "Task updated successfully",
		"data":    existing,
	})
}
func (h *FileHandler) getBucketName1(c *gin.Context) string {
	subdomain := middleware.GetRawSubdomain(c)
	if subdomain == "" || subdomain == "default" {
		return h.bucket
	}
	return fmt.Sprintf("tulip-%s-files", subdomain)
}
func (h *FileHandler) UploadTaskPhotos(c *gin.Context) {
	db := middleware.GetDB(c)
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database not initialized"})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Task ID required"})
		return
	}

	// Find task — try the Job tracker's StageTask first (the common case),
	// then fall back to Estimate Project Management's EstimateProjectTask,
	// since both share this same upload endpoint but live in separate tables.
	var task models.StageTask
	isEstimateTask := false
	if err := db.First(&task, "id = ?", id); err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find task"})
			return
		}

		var probe models.EstimateProjectTask
		if err := db.First(&probe, "id = ?", id); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find task"})
			}
			return
		}
		isEstimateTask = true
	}

	// Use same bucket logic as CreateFile
	bucket := h.getBucketName1(c)

	fmt.Printf("DEBUG: UploadTaskPhotos Content-Type: %s\n", c.Request.Header.Get("Content-Type"))
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No photos uploaded"})
		return
	}

	files := form.File["photos"]
	fmt.Printf("DEBUG: UploadTaskPhotos files found: %d\n", len(files))
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No photos found"})
		return
	}

	const maxFileSize = 10 * 1024 * 1024 // 10 MB

	allowedTypes := map[string]bool{
		"image/jpeg":    true,
		"image/jpg":     true,
		"image/png":     true,
		"image/webp":    true,
		"image/svg+xml": true,
		"image/heic":    true,
		"image/heif":    true,
	}

	var photoURLs []string

	for _, fileHeader := range files {

		// Validate file size
		if fileHeader.Size > maxFileSize {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("File %s exceeds 10MB limit", fileHeader.Filename),
			})
			return
		}

		contentType := fileHeader.Header.Get("Content-Type")

		// Validate file type
		if !allowedTypes[contentType] {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Unsupported file type: %s", contentType),
			})
			return
		}

		file, err := fileHeader.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to open file",
				"details": err.Error(),
			})
			return
		}

		fileExt := filepath.Ext(fileHeader.Filename)
		uniqueID := uuid.New().String()

		// Store inside tasks folder with unique filename
		s3Key := fmt.Sprintf(
			"tasks/%s/%s%s",
			id,
			uniqueID,
			fileExt,
		)

		_, err = h.s3Client.PutObject(&s3.PutObjectInput{
			Bucket:             aws.String(bucket),
			Key:                aws.String(s3Key),
			Body:               file,
			ContentType:        aws.String(contentType),
			ContentDisposition: aws.String("attachment"),
		})

		file.Close()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to upload file",
				"details": err.Error(),
			})
			return
		}

		// Save full URL (same behavior as current implementation)
		url := fmt.Sprintf(
			"https://%s.s3.%s.amazonaws.com/%s",
			bucket,
			h.region,
			s3Key,
		)

		photoURLs = append(photoURLs, url)
	}

	// Save URLs into task photos field
	photosJSON, err := json.Marshal(photoURLs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to marshal photos",
			"details": err.Error(),
		})
		return
	}

	if isEstimateTask {
		var estimateTask models.EstimateProjectTask
		if err := db.First(&estimateTask, "id = ?", id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reload task"})
			return
		}
		estimateTask.Photos = photoURLs
		if err := db.Save(&estimateTask); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to save task",
				"details": err.Error(),
			})
			return
		}
	} else {
		task.Photos = datatypes.JSON(photosJSON)
		if err := db.Save(&task); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to save task",
				"details": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Photos uploaded successfully",
		"photos":  photoURLs,
	})
}

func (h *Handler) DeleteTask(c *gin.Context) {
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
	var discount models.Task
	if err := tx.First(&discount, "id = ?", id); err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Residential entry not found",
				Code:    "NOT_FOUND",
				Details: err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to find residential entry",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
		}
		return
	}

	if err := tx.Delete(&discount); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to delete residential entry",
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
		"message": "Residential entry deleted successfully",
	})
}

func (h *Handler) GetAllTasks(c *gin.Context) {
	db := middleware.GetDB(c)

	var tasks []models.Task
	if err := db.Find(&tasks); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to retrieve discounts",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": tasks,
	})
}

func (h *Handler) TaskCompletionStatus(c *gin.Context) {
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

	var task models.StageTask

	if err := db.First(&task, "id = ?", id); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Task not found",
				Code:    "NOT_FOUND",
				Details: err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to retrieve task",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Completed": task.Completed,
		"Cancelled": task.Cancelled,
	})
}
