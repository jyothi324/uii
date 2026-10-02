package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tuliporg/agent-management-api/db"
	"github.com/tuliporg/agent-management-api/internal/activity"
	"github.com/tuliporg/agent-management-api/middleware"
	"github.com/tuliporg/agent-management-api/models"
	"github.com/tuliporg/agent-management-api/services"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type CreateTaskRequest struct {
	Name       string `json:"name" binding:"required"`
	DueDate    string `json:"dueDate" binding:"required"`
	Duration   int    `json:"duration" binding:"required,min=1"`
	AssignedTo string `json:"assignedTo"`
}

func marshalStageTaskProjects(projects []models.StageTaskProject) (datatypes.JSON, error) {
	if projects == nil {
		projects = []models.StageTaskProject{}
	}
	projectsJSON, err := json.Marshal(projects)
	if err != nil {
		return nil, err
	}
	return datatypes.JSON(projectsJSON), nil
}

func parseStageTaskProjects(raw datatypes.JSON) []models.StageTaskProject {
	if len(raw) == 0 {
		return []models.StageTaskProject{}
	}

	var projects []models.StageTaskProject
	if err := json.Unmarshal(raw, &projects); err != nil {
		return []models.StageTaskProject{}
	}
	if projects == nil {
		return []models.StageTaskProject{}
	}
	return projects
}

var stageOrder = []string{
	"Template",
	"CAD",
	"Cutting",
	"Fabrication",
	"Installation",
	"Service Call",
}

func (h *FileHandler) GetStagesByLeadID() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := middleware.GetDB(c)
		leadID := c.Param("lead_id")
		id, err := strconv.ParseUint(leadID, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lead ID"})
			return
		}

		if leadID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "lead_id is required"})
			return
		}
		var stages []models.Stage
		if err := db.Preload("Tasks").Preload("Users").
			Where("residential_entry_id = ?", id).Find(&stages); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stages: " + err.Error()})
			return
		}
		sortedStages := make([]map[string]interface{}, 0, len(stages))
		for _, stageName := range stageOrder {
			for _, stage := range stages {
				if stage.Name == stageName {
					stageResponse := map[string]interface{}{
						"ID":                 stage.ID,
						"LeadID":             stage.LeadID,
						"Name":               stage.Name,
						"Description":        stage.Description,
						"State":              stage.State,
						"Date":               stage.Date,
						"JobID":              stage.JobID,
						"Notes":              stage.Notes,
						"Team":               stage.Team,
						"CreatedAt":          stage.CreatedAt,
						"ResidentialEntryID": stage.ResidentialEntryID,
						"UpdatedAt":          stage.UpdatedAt,
						// "Materials":          stage.Materials,
						"Users": stage.Users,
						"Tasks": make([]map[string]interface{}, 0, len(stage.Tasks)),
					}

					for _, task := range stage.Tasks {
						var photoKeys []string
						var photoURLs []string

						if len(task.Photos) > 0 {
							if err := json.Unmarshal(task.Photos, &photoKeys); err == nil {
								for _, key := range photoKeys {
									url, err := h.generatePresignedURL(c, key)
									if err == nil {
										photoURLs = append(photoURLs, url)
									}
								}
							}
						}
						taskResponse := map[string]interface{}{
							"id":         task.ID,
							"stageId":    task.StageID,
							"name":       task.Name,
							"dueDate":    task.DueDate.UTC().Format("2006-01-02"),
							"startDate":  task.StartDate.UTC().Format("2006-01-02"),
							"startTime":  task.StartTime.UTC().Format(time.RFC3339), // ISO8601 format
							"endTime":    task.EndTime.UTC().Format(time.RFC3339),   // ISO8601 format
							"duration":   task.Duration,
							"completed":  task.Completed,
							"cancelled":  task.Cancelled,
							"assignedTo": task.AssignedTo,
							"createdAt":  task.CreatedAt.UTC().Format(time.RFC3339),
							"updatedAt":  task.UpdatedAt.UTC().Format(time.RFC3339),
							"isRunning":  task.IsRunning,
							"elapsed":    task.Elapsed,
							"notes":      task.NotesHistory,
							"photos":     photoURLs,
							"projects":   parseStageTaskProjects(task.Projects),
						}
						if task.LastResumedAt != nil {
							taskResponse["lastResumedAt"] = task.LastResumedAt.UTC().Format(time.RFC3339)
						}

						stageResponse["Tasks"] = append(stageResponse["Tasks"].([]map[string]interface{}), taskResponse)
					}

					sortedStages = append(sortedStages, stageResponse)
				}
			}
		}
		c.JSON(http.StatusOK, gin.H{"stages": sortedStages})
	}
}
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
func (h *Handler) UpdateStageStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := middleware.GetDB(c)

		leadID := c.Param("lead_id")
		stageID := c.Param("stage_id")
		if leadID == "" || stageID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "lead_id and stage_id are required"})
			return
		}

		// Read raw request body
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		var request struct {
			State string `json:"state" binding:"required"`
		}
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
			return
		}

		// Validate state
		validStates := []string{"Not Started", "In Progress", "Completed", "Cancelled"}
		if !contains(validStates, request.State) {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid state: %q", request.State)})
			return
		}
		tx := db.Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Unexpected error occurred"})
			}
		}()

		var stage models.Stage
		if err := tx.Where("id = ?", stageID).First(&stage); err != nil {
			tx.Rollback()
			c.JSON(http.StatusNotFound, gin.H{"error": "Stage not found"})
			return
		}

		// Capture old status for the activity log
		oldStatus := stage.State

		var stages []models.Stage
		if err := tx.Preload("Tasks").Where("residential_entry_id = ?", stage.ResidentialEntryID).Find(&stages); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stages: " + err.Error()})
			return
		}

		// Sort stages by stageOrder
		sortedStages := make([]models.Stage, 0, len(stages))
		for _, stageName := range stageOrder {
			for _, s := range stages {
				if s.Name == stageName {
					sortedStages = append(sortedStages, s)
				}
			}
		}

		currentIndex := -1
		for i, s := range sortedStages {
			if s.ID == stageID {
				currentIndex = i
				break
			}
		}

		if currentIndex == -1 {
			tx.Rollback()
			c.JSON(http.StatusNotFound, gin.H{"error": "Stage not found in ordered list"})
			return
		}
		stage.State = request.State
		if request.State != "Not Started" {
			now := time.Now().UTC()
			stage.Date = &now
		} else {
			stage.Date = nil
		}

		if err := tx.Save(&stage); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update stage status: " + err.Error()})
			return
		}

		if request.State == "Not Started" || request.State == "In Progress" {
			for i := currentIndex + 1; i < len(sortedStages); i++ {
				if sortedStages[i].State != "Not Started" {
					for _, task := range sortedStages[i].Tasks {
						// if len(task.Timers) > 0 {
						// 	if err := tx.Where("task_id = ?", task.ID).Delete(&models.StageTaskTimer{}); err != nil {
						// 		tx.Rollback()
						// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task timers"})
						// 		return
						// 	}
						// }
						if task.Completed {
							task.Completed = false
							if err := tx.Save(&task); err != nil {
								tx.Rollback()
								c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset task: " + err.Error()})
								return
							}
						}
						if task.Cancelled {
							task.Cancelled = false
							if err := tx.Save(&task); err != nil {
								tx.Rollback()
								c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset task: " + err.Error()})
								return
							}
						}
					}
					// Reset stage
					sortedStages[i].State = "Not Started"
					sortedStages[i].Date = nil
					if err := tx.Save(&sortedStages[i]); err != nil {
						tx.Rollback()
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset subsequent stage: " + err.Error()})
						return
					}
				}
			}
		}

		// Auto-progress next stage if current stage is Completed
		if request.State == "Completed" && currentIndex < len(sortedStages)-1 {
			nextStage := sortedStages[currentIndex+1]
			if nextStage.State == "Not Started" {
				nextStage.State = "In Progress"
				now := time.Now().UTC()
				nextStage.Date = &now
				if err := tx.Save(&nextStage); err != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update next stage: " + err.Error()})
					return
				}
			}
		}
		if request.State == "Cancelled" && currentIndex < len(sortedStages)-1 {
			nextStage := sortedStages[currentIndex+1]
			if nextStage.State == "Not Started" {
				nextStage.State = "In Progress"
				now := time.Now().UTC()
				nextStage.Date = &now
				if err := tx.Save(&nextStage); err != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update next stage: " + err.Error()})
					return
				}
			}
		}

		userID, userName := middleware.GetRequestUser(c)
		activityEvent := activity.ActivityEvent{
			EntityType: "job",
			EntityID:   stage.JobID,
			Action:     fmt.Sprintf("%s %s", strings.ToLower(stage.Name), strings.ToLower(request.State)),
			UserID:     userID,
			UserName:   userName,
			Payload: map[string]any{
				"stage_name": stage.Name,
				"stage_id":   stage.ID,
				"job_id":     stage.JobID,
				"old_status": oldStatus,
				"new_status": request.State,
				"updated_by": userName,
			},
		}

		if err := activity.InsertActivityEvent(c.Request.Context(), tx.GetUnderlying(), activityEvent); err != nil {
			log.Printf("stage activity event failed: %v", err)
		}

		if err := tx.Commit(); err != nil {
			return
		}

		// Sync job snapshot so reports reflect the stage state change immediately
		var lead models.Lead
		if err := db.Where("id = ?", stage.LeadID).First(&lead); err == nil {
			jobService := services.NewJobService()
			if syncErr := jobService.UpsertJobFromLead(db, lead); syncErr != nil {
				fmt.Printf("WARN: UpsertJobFromLead failed after stage status update for lead %s: %v\n", stage.LeadID, syncErr)
			}
		}
		c.JSON(http.StatusOK, gin.H{"message": "Stage status updated", "stage": stage})
	}
}

func (h *Handler) StartTaskTimer() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := middleware.GetDB(c)
		taskID := c.Param("id")

		now := time.Now().UTC()

		var task models.StageTask
		if err := db.Where("id = ?", taskID).First(&task); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}

		if task.Completed {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot start timer for completed task"})
			return
		}

		if task.IsRunning {
			elapsed := task.Elapsed
			if task.LastResumedAt != nil {
				elapsed += int(now.Sub(*task.LastResumedAt).Seconds())
			}
			c.JSON(http.StatusConflict, gin.H{ // ✅ Changed from 200 to 409
				"message": "Task timer already running",
				"task": gin.H{
					"id":            task.ID,
					"isRunning":     task.IsRunning,
					"elapsed":       elapsed,
					"lastResumedAt": task.LastResumedAt,
					"completed":     task.Completed,
				},
			})
			return
		}

		// ✅ Start or resume timer with server time
		task.IsRunning = true
		task.LastResumedAt = &now // ✅ Use server time

		// Only set StartTime if this is the first time starting
		if task.StartTime == nil {
			task.StartTime = &now
		}

		if err := db.Save(&task); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start task timer"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Task timer started/resumed",
			"task": gin.H{
				"id":            task.ID,
				"isRunning":     task.IsRunning,
				"elapsed":       task.Elapsed,
				"lastResumedAt": task.LastResumedAt.UTC().Format(time.RFC3339), // ✅ ISO8601 UTC
				"completed":     task.Completed,
			},
		})
	}
}

// ✅ UPDATE PauseTaskTimer() - Lines 443-506
func (h *Handler) PauseTaskTimer() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := middleware.GetDB(c)
		taskID := c.Param("id")

		// ✅ Use server time (UTC)
		now := time.Now().UTC()

		var task models.StageTask
		if err := db.Where("id = ?", taskID).First(&task); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}

		if task.Completed {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot pause completed task"})
			return
		}

		// ✅ If already paused, return 409 Conflict (not 200 OK)
		if !task.IsRunning {
			c.JSON(http.StatusConflict, gin.H{ // ✅ Changed from 200 to 409
				"message": "Task timer already paused",
				"task": gin.H{
					"id":            task.ID,
					"isRunning":     false,
					"elapsed":       task.Elapsed,
					"lastResumedAt": nil,
					"completed":     task.Completed,
				},
			})
			return
		}

		// ✅ Calculate elapsed time using server time
		if task.LastResumedAt != nil {
			additionalElapsed := int(now.Sub(*task.LastResumedAt).Seconds())
			task.Elapsed += additionalElapsed
		}

		task.IsRunning = false
		task.LastResumedAt = nil

		if err := db.Save(&task); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to pause task timer"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Task timer paused",
			"task": gin.H{
				"id":            task.ID,
				"isRunning":     false,
				"elapsed":       task.Elapsed,
				"lastResumedAt": nil,
				"completed":     task.Completed,
			},
		})
	}
}

// ✅ UPDATE CompleteTaskTimer() - Lines 508-557
func (h *Handler) CompleteTaskTimer() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := middleware.GetDB(c)
		taskID := c.Param("id")

		// ✅ Use server time (UTC)
		now := time.Now().UTC()

		tx := db.Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Unexpected error occurred"})
			}
		}()

		var task models.StageTask
		if err := tx.Preload("Stage").Where("id = ?", taskID).First(&task); err != nil {
			// Not a Job tracker task — fall back to Estimate Project
			var estTask models.EstimateProjectTask
			if estErr := tx.Where("id = ?", taskID).First(&estTask); estErr != nil {
				tx.Rollback()
				c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
				return
			}

			if estTask.Completed {
				tx.Rollback()
				c.JSON(http.StatusConflict, gin.H{
					"message":       "Task already completed",
					"total_elapsed": estTask.Elapsed,
					"task":          estTask,
				})
				return
			}

			if estTask.IsRunning && estTask.LastResumedAt != nil {
				estTask.Elapsed += int(now.Sub(*estTask.LastResumedAt).Seconds())
			}
			estTask.IsRunning = false
			estTask.Completed = true
			estTask.LastResumedAt = nil

			if err := tx.Save(&estTask); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete task"})
				return
			}

			if err := tx.Commit(); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit task completion"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"message":       "Task completed",
				"total_elapsed": estTask.Elapsed,
				"task": gin.H{
					"id":        estTask.ID,
					"completed": estTask.Completed,
					"elapsed":   estTask.Elapsed,
				},
			})
			return
		}

		// ✅ If already completed, return 409 Conflict (not 200 OK)
		if task.Completed {
			tx.Rollback()
			c.JSON(http.StatusConflict, gin.H{ // ✅ Changed from 200 to 409
				"message":       "Task already completed",
				"total_elapsed": task.Elapsed,
				"task":          task,
			})
			return
		}

		// ✅ Calculate final elapsed time using server time
		if task.IsRunning && task.LastResumedAt != nil {
			elapsed := int(now.Sub(*task.LastResumedAt).Seconds())
			task.Elapsed += elapsed
		}

		task.IsRunning = false
		task.Completed = true
		task.EndTime = &now // ✅ Use server time
		task.LastResumedAt = nil

		if err := tx.Save(&task); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete task"})
			return
		}

		userID, userName := middleware.GetRequestUser(c)
		event := activity.ActivityEvent{
			EntityType: "job",
			EntityID:   task.Stage.JobID,
			Action:     fmt.Sprintf("task completed"),
			UserID:     userID,
			UserName:   userName,
			Payload: map[string]any{
				"task_name":  task.Name,
				"completed":  true,
				"updated_by": userName,
			},
		}
		if err := activity.InsertActivityEvent(c.Request.Context(), tx.GetUnderlying(), event); err != nil {
			log.Printf("activity event publish failed: %v", err)
		}

		if err := tx.Commit(); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit task completion"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":       "Task completed",
			"total_elapsed": task.Elapsed,
			"task": gin.H{
				"id":        task.ID,
				"completed": task.Completed,
				"elapsed":   task.Elapsed,
				"endTime":   task.EndTime.UTC().Format(time.RFC3339), // ✅ ISO8601 UTC
			},
		})
	}
}

func (h *Handler) CancelTask() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := middleware.GetDB(c)
		taskID := c.Param("id")

		now := time.Now().UTC()

		tx := db.Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Unexpected error occurred"})
			}
		}()

		var task models.StageTask
		if err := tx.Preload("Stage").Where("id = ?", taskID).First(&task); err != nil {
			// Not a Job tracker task — fall back to Estimate Project
			var estTask models.EstimateProjectTask
			if estErr := tx.Where("id = ?", taskID).First(&estTask); estErr != nil {
				tx.Rollback()
				c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
				return
			}

			if estTask.Cancelled {
				tx.Rollback()
				c.JSON(http.StatusConflict, gin.H{
					"message":       "Task already cancelled",
					"total_elapsed": estTask.Elapsed,
					"task":          estTask,
				})
				return
			}

			if estTask.IsRunning && estTask.LastResumedAt != nil {
				estTask.Elapsed += int(now.Sub(*estTask.LastResumedAt).Seconds())
			}
			estTask.IsRunning = false
			estTask.Cancelled = true
			estTask.LastResumedAt = nil

			if err := tx.Save(&estTask); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel task"})
				return
			}

			if err := tx.Commit(); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit task cancellation"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"message":       "Task cancelled",
				"total_elapsed": estTask.Elapsed,
				"task": gin.H{
					"id":        estTask.ID,
					"cancelled": estTask.Cancelled,
					"elapsed":   estTask.Elapsed,
				},
			})
			return
		}

		if task.Cancelled {
			tx.Rollback()
			c.JSON(http.StatusConflict, gin.H{
				"message":       "Task already cancelled",
				"total_elapsed": task.Elapsed,
				"task":          task,
			})
			return
		}

		if task.IsRunning && task.LastResumedAt != nil {
			elapsed := int(now.Sub(*task.LastResumedAt).Seconds())
			task.Elapsed += elapsed
		}

		task.IsRunning = false
		task.Cancelled = true
		task.EndTime = &now
		task.LastResumedAt = nil

		if err := tx.Save(&task); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel task"})
			return
		}

		userID, userName := middleware.GetRequestUser(c)
		event := activity.ActivityEvent{
			EntityType: "job",
			EntityID:   task.Stage.JobID,
			Action:     fmt.Sprintf("task cancelled"),
			UserID:     userID,
			UserName:   userName,
			Payload: map[string]any{
				"task_name":  task.Name,
				"cancelled":  true,
				"updated_by": userName,
			},
		}
		if err := activity.InsertActivityEvent(c.Request.Context(), tx.GetUnderlying(), event); err != nil {
			log.Printf("activity event publish failed: %v", err)
		}

		if err := tx.Commit(); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit task cancellation"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":       "Task cancelled",
			"total_elapsed": task.Elapsed,
			"task": gin.H{
				"id":        task.ID,
				"cancelled": task.Cancelled,
				"elapsed":   task.Elapsed,
				"endTime":   task.EndTime.UTC().Format(time.RFC3339),
			},
		})
	}
}

// ✅ UPDATE GetTaskTimerStatus() - Lines 615-660
func (h *Handler) GetTaskTimerStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := middleware.GetDB(c)
		taskID := c.Param("id")

		// ✅ Always use server time (don't trust client timestamp)
		currentTime := time.Now().UTC()

		var task models.StageTask
		if err := db.Where("id = ?", taskID).First(&task); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}

		// ✅ Calculate elapsed time using server time
		elapsed := task.Elapsed
		if task.IsRunning && task.LastResumedAt != nil {
			currentElapsed := int(currentTime.Sub(*task.LastResumedAt).Seconds())
			elapsed += currentElapsed
		}

		response := gin.H{
			"task_id":      task.ID,
			"is_running":   task.IsRunning,
			"completed":    task.Completed,
			"elapsed_time": elapsed,
		}

		if task.StartTime != nil {
			response["start_time"] = task.StartTime.UTC().Format(time.RFC3339)
		}
		if task.EndTime != nil {
			response["end_time"] = task.EndTime.UTC().Format(time.RFC3339)
		}
		if task.LastResumedAt != nil {
			response["last_resumed_at"] = task.LastResumedAt.UTC().Format(time.RFC3339)
		}

		c.JSON(http.StatusOK, response)
	}
}
func (h *Handler) CreateStageTask() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := middleware.GetDB(c)
		if db == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database not initialized"})
			return
		}

		leadID := c.Param("lead_id")
		stageID := c.Param("stage_id")
		if leadID == "" || stageID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "lead_id and stage_id are required"})
			return
		}

		resEntryID, err := strconv.ParseUint(leadID, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "lead_id must be a numeric residential_entry_id"})
			return
		}

		body, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		var request struct {
			Name       string                     `json:"name" binding:"required"`
			StartDate  string                     `json:"startDate" binding:"required"`
			StartTime  string                     `json:"startTime" binding:"required"`
			EndTime    string                     `json:"endTime"`
			AssignedTo string                     `json:"assignedTo"`
			TaskID     string                     `json:"taskId"`
			Duration   *int                       `json:"duration,omitempty"`
			Completed  *bool                      `json:"completed,omitempty"`
			Cancelled  *bool                      `json:"cancelled,omitempty"`
			Notes      string                     `json:"notes,omitempty"`
			JobID      string                     `json:"jobId,omitempty"`
			Projects   *[]models.StageTaskProject `json:"projects,omitempty"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
			return
		}

		if strings.Contains(request.StartDate, "T") {
			request.StartDate = request.StartDate[:10]
		}

		startDate, err := time.Parse("2006-01-02", request.StartDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid startDate format '%s', use YYYY-MM-DD", request.StartDate)})
			return
		}

		startTimeUTC, err := time.Parse(time.RFC3339, request.StartTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid startTime format, use ISO8601 (e.g., 2024-01-15T09:00:00Z)"})
			return
		}
		startTimeUTC = startTimeUTC.UTC()

		tx := db.Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Unexpected error occurred"})
			}
		}()

		var stage models.Stage
		if err := tx.Where("id = ? AND residential_entry_id = ?", stageID, resEntryID).First(&stage); err != nil {
			tx.Rollback()
			c.JSON(http.StatusNotFound, gin.H{"error": "Stage or lead not found"})
			return
		}

		durationMinutes := 240
		if request.Duration != nil && *request.Duration > 0 {
			durationMinutes = *request.Duration
		}

		endTimeUTC := startTimeUTC.Add(time.Duration(durationMinutes) * time.Minute)

		if request.EndTime != "" {
			if parsedEnd, err := time.Parse(time.RFC3339, request.EndTime); err == nil {
				endTimeUTC = parsedEnd.UTC()
				durationMinutes = int(endTimeUTC.Sub(startTimeUTC).Minutes())
			}
		}

		dueDate := endTimeUTC
		now := time.Now().UTC()

		var noteEntries []models.NoteEntry
		if request.Notes != "" {
			noteEntries = append(noteEntries, models.NoteEntry{
				Text:      request.Notes,
				CreatedAt: now.Format(time.RFC3339),
			})
		}

		if request.TaskID != "" {
			var existingTask models.StageTask
			if err := tx.Where("id = ? AND stage_id = ?", request.TaskID, stageID).First(&existingTask); err != nil {
				tx.Rollback()
				c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
				return
			}

			oldName := existingTask.Name
			oldAssignedTo := existingTask.AssignedTo
			oldStartTime := existingTask.StartTime
			oldEndTime := existingTask.EndTime
			oldDuration := existingTask.Duration

			existingTask.Name = request.Name
			existingTask.StartDate = startDate
			existingTask.StartTime = &startTimeUTC
			existingTask.EndTime = &endTimeUTC
			existingTask.DueDate = dueDate
			existingTask.Duration = durationMinutes
			if strings.TrimSpace(request.AssignedTo) != "" && !strings.Contains(request.AssignedTo, ",") {
				existingTask.AssignedTo = replacePrimaryAssignee(oldAssignedTo, strings.TrimSpace(request.AssignedTo))
			} else {
				existingTask.AssignedTo = request.AssignedTo
			}
			existingTask.UpdatedAt = now

			if len(noteEntries) > 0 {
				existingTask.NotesHistory = append(existingTask.NotesHistory, noteEntries...)
			}

			if request.Projects != nil {
				projectsJSON, err := marshalStageTaskProjects(*request.Projects)
				if err != nil {
					tx.Rollback()
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid projects data"})
					return
				}
				existingTask.Projects = projectsJSON
			}

			if request.Completed != nil {
				existingTask.Completed = *request.Completed
				if *request.Completed {
					if existingTask.LastResumedAt != nil {
						additionalElapsed := int(time.Since(*existingTask.LastResumedAt).Seconds())
						existingTask.Elapsed += additionalElapsed
					}
					existingTask.IsRunning = false
					existingTask.LastResumedAt = nil
				}
			}
			if request.Cancelled != nil {
				existingTask.Cancelled = *request.Cancelled
				if *request.Cancelled {
					if existingTask.LastResumedAt != nil {
						additionalElapsed := int(time.Since(*existingTask.LastResumedAt).Seconds())
						existingTask.Elapsed += additionalElapsed
					}
					existingTask.IsRunning = false
					existingTask.LastResumedAt = nil
				}
			}

			if err := tx.Save(&existingTask); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
				return
			}

			if oldAssignedTo != existingTask.AssignedTo {
				syncStageUsersForPrimaryChange(tx, stageID, existingTask.Name, oldAssignedTo, existingTask.AssignedTo)
			}

			if strings.TrimSpace(existingTask.AssignedTo) == "" {
				if err := tx.Where("stage_id = ? AND task_name IN ?", stageID, []string{oldName, existingTask.Name}).Delete(&models.StageUser{}).Error; err != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear task users"})
					return
				}
			}

			userID, userName := middleware.GetRequestUser(c)
			payload := map[string]any{
				"task_name":  existingTask.Name,
				"stage_name": stage.Name,
				"updated_by": userName,
			}

			if oldName != existingTask.Name {
				payload["name_old"] = oldName
				payload["name_new"] = existingTask.Name
			}
			if oldAssignedTo != existingTask.AssignedTo {
				payload["assigned_to_old"] = oldAssignedTo
				payload["assigned_to_new"] = existingTask.AssignedTo
			}
			if oldStartTime != existingTask.StartTime {
				if oldStartTime != nil {
					payload["start_time_old"] = oldStartTime.UTC().Format(time.RFC3339)
				}
				if existingTask.StartTime != nil {
					payload["start_time_new"] = existingTask.StartTime.UTC().Format(time.RFC3339)
				}
			}
			if oldEndTime != existingTask.EndTime {
				if oldEndTime != nil {
					payload["end_time_old"] = oldEndTime.UTC().Format(time.RFC3339)
				}
				if existingTask.EndTime != nil {
					payload["end_time_new"] = existingTask.EndTime.UTC().Format(time.RFC3339)
				}
			}
			if oldDuration != existingTask.Duration {
				payload["duration_old"] = oldDuration
				payload["duration_new"] = existingTask.Duration
			}

			jobID := request.JobID
			if jobID == "" {
				jobID = stage.JobID
			}

			event := activity.ActivityEvent{
				EntityType: "job",
				EntityID:   jobID,
				Action:     "task updated",
				UserID:     userID,
				UserName:   userName,
				Payload:    payload,
			}
			if err := activity.InsertActivityEvent(c.Request.Context(), tx.GetUnderlying(), event); err != nil {
				log.Printf("activity event publish failed: %v", err)
			}

			if err := tx.Commit(); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
				return
			}

			var lead models.Lead
			if err := db.Where("id = ?", stage.LeadID).First(&lead); err == nil {
				jobService := services.NewJobService()
				if syncErr := jobService.UpsertJobFromLead(db, lead); syncErr != nil {
					fmt.Printf("WARN: UpsertJobFromLead failed after task update for lead %s: %v\n", stage.LeadID, syncErr)
				}
			}

			c.JSON(http.StatusOK, gin.H{
				"message": "Task updated successfully",
				"task": gin.H{
					"id":         existingTask.ID,
					"stageId":    existingTask.StageID,
					"name":       existingTask.Name,
					"dueDate":    existingTask.DueDate.UTC().Format("2006-01-02"),
					"startDate":  existingTask.StartDate.UTC().Format("2006-01-02"),
					"startTime":  existingTask.StartTime.UTC().Format(time.RFC3339),
					"endTime":    existingTask.EndTime.UTC().Format(time.RFC3339),
					"duration":   existingTask.Duration,
					"completed":  existingTask.Completed,
					"cancelled":  existingTask.Cancelled,
					"assignedTo": existingTask.AssignedTo,
					"isRunning":  existingTask.IsRunning,
					"elapsed":    existingTask.Elapsed,
					"updatedAt":  existingTask.UpdatedAt.UTC().Format(time.RFC3339),
					"notes":      existingTask.NotesHistory,
					"photos":     existingTask.Photos,
					"projects":   parseStageTaskProjects(existingTask.Projects),
				},
			})
		} else {
			taskID := uuid.NewString()
			completed := false
			if request.Completed != nil {
				completed = *request.Completed
			}
			cancelled := false
			if request.Cancelled != nil {
				cancelled = *request.Cancelled
			}
			projectsJSON, err := marshalStageTaskProjects(nil)
			if err != nil {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid projects data"})
				return
			}
			if request.Projects != nil {
				projectsJSON, err = marshalStageTaskProjects(*request.Projects)
				if err != nil {
					tx.Rollback()
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid projects data"})
					return
				}
			}
			// Snapshot the lead's shipping address as it stands right now,
			// so this task keeps pointing at the address it was created
			// against even if the lead's delivery type / shipping address
			// is changed later.
			var taskAddress datatypes.JSON
			var leadForAddress models.Lead
			if err := tx.Select("id, details").Where("id = ?", stage.LeadID).First(&leadForAddress); err == nil {
				if leadForAddress.Details.ShippingAddressID != nil {
					var shipAddr models.Address
					if err := tx.Where("id = ?", *leadForAddress.Details.ShippingAddressID).First(&shipAddr); err == nil {
						snapshot := map[string]interface{}{
							"street":    shipAddr.Street,
							"unitSuite": shipAddr.UnitSuite,
							"city":      shipAddr.City,
							"state":     shipAddr.State,
							"zip":       shipAddr.Zip,
							"county":    shipAddr.County,
							"lat":       shipAddr.Lat,
							"lon":       shipAddr.Lng,
							"formatted": shipAddr.Formatted,
						}
						if b, marshalErr := json.Marshal(snapshot); marshalErr == nil {
							taskAddress = datatypes.JSON(b)
						}
					}
				}
			}

			task := models.StageTask{
				ID:           taskID,
				StageID:      stageID,
				Name:         request.Name,
				DueDate:      dueDate,
				StartDate:    startDate,
				StartTime:    &startTimeUTC,
				EndTime:      &endTimeUTC,
				Duration:     durationMinutes,
				Completed:    completed,
				Address:      taskAddress,
				Cancelled:    cancelled,
				AssignedTo:   request.AssignedTo,
				NotesHistory: noteEntries,
				Projects:     projectsJSON,
				CreatedAt:    now,
				UpdatedAt:    now,
			}

			if err := tx.Create(&task); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
				return
			}

			userID, userName := middleware.GetRequestUser(c)
			event := activity.ActivityEvent{
				EntityType: "job",
				EntityID:   stage.JobID,
				Action:     "task created",
				UserID:     userID,
				UserName:   userName,
				Payload: map[string]any{
					"task_name":   task.Name,
					"created_by":  userName,
					"assigned_to": task.AssignedTo,
					"start_time":  task.StartTime,
					"stage_name":  stage.Name,
				},
			}
			if err := activity.InsertActivityEvent(c.Request.Context(), tx.GetUnderlying(), event); err != nil {
				log.Printf("activity event publish failed: %v", err)
			}

			if err := tx.Commit(); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
				return
			}

			var lead models.Lead
			if err := db.Where("id = ?", stage.LeadID).First(&lead); err == nil {
				jobService := services.NewJobService()
				if syncErr := jobService.UpsertJobFromLead(db, lead); syncErr != nil {
					fmt.Printf("WARN: UpsertJobFromLead failed after task create for lead %s: %v\n", stage.LeadID, syncErr)
				}
			}

			c.JSON(http.StatusCreated, gin.H{
				"message": "Task created successfully",
				"task": gin.H{
					"id":         task.ID,
					"stageId":    task.StageID,
					"name":       task.Name,
					"dueDate":    task.DueDate.UTC().Format("2006-01-02"),
					"startDate":  task.StartDate.UTC().Format("2006-01-02"),
					"startTime":  task.StartTime.UTC().Format(time.RFC3339),
					"endTime":    task.EndTime.UTC().Format(time.RFC3339),
					"duration":   task.Duration,
					"completed":  task.Completed,
					"cancelled":  task.Cancelled,
					"assignedTo": task.AssignedTo,
					"notes":      task.NotesHistory,
					"photos":     task.Photos,
					"projects":   parseStageTaskProjects(task.Projects),
				},
			})
		}
	}
}
func (h *Handler) UpdateStageTask() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := middleware.GetDB(c)

		leadID := c.Param("lead_id")
		stageID := c.Param("stage_id")
		taskID := c.Param("task_id")
		if leadID == "" || stageID == "" || taskID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "lead_id, stage_id, and task_id are required"})
			return
		}

		var request struct {
			Completed bool                       `json:"completed"`
			Cancelled bool                       `json:"cancelled"`
			Photos    []string                   `json:"photos"`
			Projects  *[]models.StageTaskProject `json:"projects,omitempty"`
		}

		if strings.HasPrefix(c.GetHeader("Content-Type"), "multipart/form-data") {
			request.Completed, _ = strconv.ParseBool(c.PostForm("completed"))
			request.Cancelled, _ = strconv.ParseBool(c.PostForm("cancelled"))

			form, err := c.MultipartForm()
			if err == nil && form != nil {
				files := form.File["photos"]
				if len(files) > 0 {
					sess, err := session.NewSession()
					if err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize AWS session"})
						return
					}
					s3Client := s3.New(sess)
					region := "us-east-1"
					if sess.Config.Region != nil && *sess.Config.Region != "" {
						region = *sess.Config.Region
					}

					tenant := c.GetString("tenant")
					bucket := "tulip-default-files"
					if tenant != "" && tenant != "default" {
						bucket = fmt.Sprintf("tulip-%s-files", tenant)
					}

					for _, fileHeader := range files {
						file, err := fileHeader.Open()
						if err != nil {
							c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open uploaded file"})
							return
						}

						fileExt := filepath.Ext(fileHeader.Filename)
						s3Key := fmt.Sprintf("tasks/%s%s", uuid.New().String(), fileExt)

						_, err = s3Client.PutObject(&s3.PutObjectInput{
							Bucket:             aws.String(bucket),
							Key:                aws.String(s3Key),
							Body:               file,
							ContentType:        aws.String(fileHeader.Header.Get("Content-Type")),
							ContentDisposition: aws.String("attachment"),
						})
						file.Close()

						if err != nil {
							c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file to S3"})
							return
						}

						url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucket, region, s3Key)
						request.Photos = append(request.Photos, url)
					}
				}
			}
		} else {
			if err := c.ShouldBindJSON(&request); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
				return
			}
		}

		tx := db.Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Unexpected error occurred"})
			}
		}()

		var task models.StageTask
		if err := tx.Where("id = ? AND stage_id = ?", taskID, stageID).First(&task); err != nil {
			// Estimate Project Management task fallback — same rationale as
			// the timer handlers above.
			var estTask models.EstimateProjectTask
			if estErr := tx.Where("id = ? AND stage_id = ?", taskID, stageID).First(&estTask); estErr != nil {
				tx.Rollback()
				c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
				return
			}

			estTask.Completed = request.Completed
			estTask.Cancelled = request.Cancelled
			if (request.Completed || request.Cancelled) && estTask.IsRunning {
				if estTask.LastResumedAt != nil {
					estTask.Elapsed += int(time.Now().UTC().Sub(*estTask.LastResumedAt).Seconds())
				}
				estTask.IsRunning = false
				estTask.LastResumedAt = nil
			}
			if request.Photos != nil {
				estTask.Photos = request.Photos
			}

			if err := tx.Save(&estTask); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task: " + err.Error()})
				return
			}

			if err := tx.Commit(); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "Task updated successfully", "task": estTask})
			return
		}

		var stage models.Stage
		if err := tx.Where("id = ?", stageID).First(&stage); err != nil {
			tx.Rollback()
			c.JSON(http.StatusNotFound, gin.H{"error": "Stage not found"})
			return
		}

		task.Completed = request.Completed
		if request.Completed && task.IsRunning {
			if task.LastResumedAt != nil {
				task.Elapsed += int(time.Now().UTC().Sub(*task.LastResumedAt).Seconds())
			}
			task.IsRunning = false
			task.LastResumedAt = nil
		}

		task.Cancelled = request.Cancelled
		if request.Cancelled && task.IsRunning {
			if task.LastResumedAt != nil {
				task.Elapsed += int(time.Now().UTC().Sub(*task.LastResumedAt).Seconds())
			}
			task.IsRunning = false
			task.LastResumedAt = nil
		}

		if request.Photos != nil {
			photosJSON, _ := json.Marshal(request.Photos)
			task.Photos = datatypes.JSON(photosJSON)
		}

		if request.Projects != nil {
			projectsJSON, err := marshalStageTaskProjects(*request.Projects)
			if err != nil {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid projects data"})
				return
			}
			task.Projects = projectsJSON
		}

		if err := tx.Save(&task); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task: " + err.Error()})
			return
		}

		// Activity Log
		userID, userName := middleware.GetRequestUser(c)
		action := "updated"
		if request.Completed {
			action = "completed"
		} else if request.Cancelled {
			action = "cancelled"
		}
		event := activity.ActivityEvent{
			EntityType: "job",
			EntityID:   stage.JobID,
			Action:     fmt.Sprintf("task %s", action),
			UserID:     userID,
			UserName:   userName,
			Payload: map[string]any{
				"task_name":  task.Name,
				"completed":  task.Completed,
				"cancelled":  task.Cancelled,
				"updated_by": userName,
			},
		}
		if err := activity.InsertActivityEvent(c.Request.Context(), tx.GetUnderlying(), event); err != nil {
			log.Printf("activity event publish failed: %v", err)
		}

		if err := tx.Commit(); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
			return
		}

		// Sync job snapshot so reports reflect the task completion/cancellation immediately
		var lead models.Lead
		if err := db.Where("id = ?", stage.LeadID).First(&lead); err == nil {
			jobService := services.NewJobService()
			if syncErr := jobService.UpsertJobFromLead(db, lead); syncErr != nil {
				fmt.Printf("WARN: UpsertJobFromLead failed after task status update for lead %s: %v\n", stage.LeadID, syncErr)
			}
		}

		c.JSON(http.StatusOK, gin.H{"message": "Task updated successfully", "task": task})
	}
}

// func (h *Handler) ToggleStageTaskTimer() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		db := middleware.GetDB(c)

// 		leadID := c.Param("lead_id")
// 		stageID := c.Param("stage_id")
// 		taskID := c.Param("task_id")

// 		var request struct {
// 			IsRunning bool `json:"isRunning" binding:"required"`
// 		}
// 		if err := c.ShouldBindJSON(&request); err != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
// 			return
// 		}

// 		tx := db.Begin()
// 		defer func() {
// 			if r := recover(); r != nil {
// 				tx.Rollback()
// 				c.JSON(http.StatusInternalServerError, gin.H{"error": "Unexpected error occurred"})
// 			}
// 		}()

// 		var task models.StageTask
// 		if err := tx.Where("id = ? AND stage_id = ?", taskID, stageID).First(&task); err != nil {
// 			tx.Rollback()
// 			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
// 			return
// 		}

// 		var stage models.Stage
// 		if err := tx.Where("id = ? AND lead_id = ?", stageID, leadID).First(&stage); err != nil {
// 			tx.Rollback()
// 			c.JSON(http.StatusNotFound, gin.H{"error": "Stage not found"})
// 			return
// 		}

// 		if task.Completed {
// 			tx.Rollback()
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot toggle timer for completed task"})
// 			return
// 		}

// 		var timer models.StageTaskTimer
// 		err := tx.Where("task_id = ? AND is_running = ?", taskID, true).First(&timer)
// 		if request.IsRunning {
// 			if err == nil {
// 				tx.Rollback()
// 				c.JSON(http.StatusBadRequest, gin.H{"error": "Timer is already running"})
// 				return
// 			}
// 			timer = models.StageTaskTimer{
// 				ID:        uuid.NewString(),
// 				TaskID:    taskID,
// 				IsRunning: true,
// 				StartTime: func() *time.Time { t := time.Now().UTC(); return &t }(),
// 			}
// 			if err := tx.Create(&timer); err != nil {
// 				tx.Rollback()
// 				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start timer"})
// 				return
// 			}
// 		} else {
// 			if err != nil {
// 				tx.Rollback()
// 				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch timer"})
// 				return
// 			}
// 			timer.IsRunning = false
// 			timer.ElapsedTime += int(time.Now().UTC().Sub(*timer.StartTime).Seconds())
// 			timer.StartTime = nil
// 			if err := tx.Save(&timer); err != nil {
// 				tx.Rollback()
// 				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stop timer"})
// 				return
// 			}
// 		}

// 		// Check if task duration is exceeded
// 		var totalElapsed int
// 		if err := tx.Model(&models.StageTaskTimer{}).Where("task_id = ?", taskID).Select("SUM(elapsed_time)").Scan(&totalElapsed); err != nil {
// 			tx.Rollback()
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate elapsed time"})
// 			return
// 		}
// 		if totalElapsed >= task.Duration*60 {
// 			task.Completed = true
// 			if err := tx.Save(&task); err != nil {
// 				tx.Rollback()
// 				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark task as completed"})
// 				return
// 			}
// 			timer.IsRunning = false
// 			timer.StartTime = nil
// 			if err := tx.Save(&timer); err != nil {
// 				tx.Rollback()
// 				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update timer"})
// 				return
// 			}
// 		}

// 		if err := tx.Commit(); err != nil {
// 			tx.Rollback()
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
// 			return
// 		}

// 		c.JSON(http.StatusOK, gin.H{"message": "Timer toggled successfully", "timer": timer})
// 	}
// }

//not using

// func (h *Handler) AddStageMaterial() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		db := middleware.GetDB(c)

// 		leadID := c.Param("lead_id")
// 		stageID := c.Param("stage_id")
// 		if leadID == "" || stageID == "" {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "lead_id and stage_id are required"})
// 			return
// 		}

// 		var request struct {
// 			Name string `json:"name" binding:"required"`
// 		}
// 		if err := c.ShouldBindJSON(&request); err != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
// 			return
// 		}

// 		tx := db.Begin()
// 		defer func() {
// 			if r := recover(); r != nil {
// 				tx.Rollback()
// 				c.JSON(http.StatusInternalServerError, gin.H{"error": "Unexpected error occurred"})
// 			}
// 		}()

// 		var stage models.Stage
// 		if err := tx.Where("id = ? AND lead_id = ?", stageID, leadID).First(&stage); err != nil {
// 			tx.Rollback()
// 			c.JSON(http.StatusNotFound, gin.H{"error": "Stage not found"})
// 			return
// 		}

// 		material := models.StageMaterial{
// 			ID:      uuid.NewString(),
// 			StageID: stageID,
// 			Name:    request.Name,
// 		}

// 		if err := tx.Create(&material); err != nil {
// 			tx.Rollback()
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add material: " + err.Error()})
// 			return
// 		}

// 		if err := tx.Commit().Error; err != nil {
// 			tx.Rollback()
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
// 			return
// 		}
// 		c.JSON(http.StatusCreated, gin.H{"message": "Material added successfully", "material": material})
// 	}
// }

//not using

func (h *Handler) AddStageUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := middleware.GetDB(c)
		if db == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection unavailable"})
			return
		}

		leadID := c.Param("lead_id")
		stageID := c.Param("stage_id")
		if leadID == "" || stageID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "lead_id and stage_id are required"})
			return
		}

		var request struct {
			UserID   string `json:"userId" binding:"required"`
			Name     string `json:"name" binding:"required"`
			TaskName string `json:"taskName"`
			TaskID   string `json:"taskId" binding:"required"`
		}
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
			return
		}

		// Check duplicate per (stage, user, task) using your wrapper's pattern
		var existing models.StageUser
		if err := db.Where(
			"stage_id = ? AND user_id = ? AND task_name = ?",
			stageID, request.UserID, request.TaskName,
		).First(&existing); err == nil {
			c.JSON(http.StatusOK, gin.H{"message": "User already in stage for this task", "user": existing})
			return
		}

		user := models.StageUser{
			ID:       uuid.NewString(),
			StageID:  stageID,
			UserID:   request.UserID,
			Name:     request.Name,
			TaskName: request.TaskName,
		}

		if err := db.Create(&user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add user: " + err.Error()})
			return
		}

		var task models.StageTask
		if err := db.Where("id = ? AND stage_id = ?", request.TaskID, stageID).First(&task); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Selected task was not found"})
			return
		}
		assignment := request.UserID + "-" + request.Name
		assignments := strings.Split(task.AssignedTo, ",")
		alreadyAssigned := false
		for _, existingAssignment := range assignments {
			if strings.TrimSpace(existingAssignment) == assignment {
				alreadyAssigned = true
				break
			}
		}
		if !alreadyAssigned {
			if strings.TrimSpace(task.AssignedTo) == "" {
				task.AssignedTo = assignment
			} else {
				task.AssignedTo += "," + assignment
			}
		}
		if err := db.Save(&task); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign user to task: " + err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "User added successfully", "user": user})
	}
}

func (h *Handler) GetStageUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := middleware.GetDB(c)
		if db == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection unavailable"})
			return
		}

		stageID := c.Param("stage_id")
		if stageID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "stage_id is required"})
			return
		}

		var users []models.StageUser
		if err := db.Where("stage_id = ?", stageID).Find(&users); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stage users: " + err.Error()})
			return
		}

		// Group by task name so the frontend can easily render per-task members
		taskGroups := make(map[string][]gin.H)
		for _, u := range users {
			taskGroups[u.TaskName] = append(taskGroups[u.TaskName], gin.H{
				"id":        u.ID,
				"stageId":   u.StageID,
				"userId":    u.UserID,
				"name":      u.Name,
				"taskName":  u.TaskName,
				"createdAt": u.CreatedAt.UTC().Format(time.RFC3339),
			})
		}

		// Build ordered slice so response is deterministic
		grouped := make([]gin.H, 0, len(taskGroups))
		for taskName, members := range taskGroups {
			grouped = append(grouped, gin.H{
				"taskName": taskName,
				"members":  members,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"stageId":    stageID,
			"total":      len(users),
			"taskGroups": grouped,
			"users":      users, // flat list also included for convenience
		})
	}
}
func (h *Handler) UpdateStageUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := middleware.GetDB(c)
		stageID := c.Param("stage_id")

		var request struct {
			UserID          string `json:"userId" binding:"required"`
			Name            string `json:"name" binding:"required"`
			TaskName        string `json:"taskName"`
			TaskID          string `json:"taskId"`
			ReplacedUserID  string `json:"replacedUserId"`
		}
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
			return
		}

		assignment := strings.TrimSpace(request.UserID + "-" + request.Name)

		if request.ReplacedUserID != "" && request.ReplacedUserID != request.UserID {
			_ = db.Where(
				"stage_id = ? AND user_id = ? AND task_name = ?",
				stageID, request.ReplacedUserID, request.TaskName,
			).Delete(&models.StageUser{})
		}

		var existing models.StageUser
		err := db.Where(
			"stage_id = ? AND user_id = ? AND task_name = ?",
			stageID, request.UserID, request.TaskName,
		).First(&existing)

		if err != nil {
			user := models.StageUser{
				ID:       uuid.NewString(),
				StageID:  stageID,
				UserID:   request.UserID,
				Name:     request.Name,
				TaskName: request.TaskName,
			}
			if err := db.Create(&user); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add user"})
				return
			}
			existing = user
		} else if existing.Name != request.Name {
			existing.Name = request.Name
			if err := db.Save(&existing); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
				return
			}
		}

		var tasks []models.StageTask
		taskQuery := db.Where("stage_id = ? AND name = ?", stageID, request.TaskName)
		if request.TaskID != "" {
			taskQuery = db.Where("id = ? AND stage_id = ?", request.TaskID, stageID)
		}
		if err := taskQuery.Find(&tasks); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find task assignment"})
			return
		}
		for _, task := range tasks {
			updatedAssignedTo := replacePrimaryAssignee(task.AssignedTo, assignment)
			if updatedAssignedTo == task.AssignedTo && !hasTaskAssignee(task.AssignedTo, request.UserID) {
				updatedAssignedTo = assignment
			}
			if updatedAssignedTo != task.AssignedTo {
				task.AssignedTo = updatedAssignedTo
				if err := db.Save(&task); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task assignment"})
					return
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{"message": "User updated", "user": existing})
	}
}

func (h *Handler) DeleteStageUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := middleware.GetDB(c)
		if db == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection unavailable"})
			return
		}
 
		stageID := c.Param("stage_id")
		userID := c.Param("user_id")
		if stageID == "" || userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "stage_id and user_id are required"})
			return
		}

		taskName := c.Query("task_name")
		if taskName == "" {
			var request struct {
				TaskName string `json:"taskName"`
			}
			_ = c.ShouldBindJSON(&request)
			taskName = request.TaskName
		}

		var existing models.StageUser
		query := db.Where("stage_id = ? AND user_id = ?", stageID, userID)
		if taskName != "" {
			query = query.Where("task_name = ?", taskName)
		}
		if err := query.First(&existing); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found in this stage"})
			return
		}

		var assignedTasks []models.StageTask
		if err := db.Where("stage_id = ? AND name = ?", stageID, existing.TaskName).Find(&assignedTasks); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check task assignment: " + err.Error()})
			return
		}
		for _, task := range assignedTasks {
			if isPrimaryTaskAssignee(task.AssignedTo, userID) {
				c.JSON(http.StatusConflict, gin.H{"error": "User is the primary assignee for this task. Reassign the task before deleting the user."})
				return
			}
		}

		if err := db.Delete(&existing); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove user: " + err.Error()})
			return
		}

		var tasks []models.StageTask
		if err := db.Where("stage_id = ? AND name = ?", stageID, existing.TaskName).Find(&tasks); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find task assignment: " + err.Error()})
			return
		}
		for _, task := range tasks {
			updatedAssignedTo := removeTaskAssignee(task.AssignedTo, userID)
			if updatedAssignedTo != task.AssignedTo {
				task.AssignedTo = updatedAssignedTo
				if err := db.Save(&task); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove user from task: " + err.Error()})
					return
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "User removed successfully",
			"userId":  userID,
			"stageId": stageID,
		})
	}
}

func removeTaskAssignee(existing, userID string) string {
	kept := make([]string, 0)
	for _, assignment := range strings.Split(existing, ",") {
		assignment = strings.TrimSpace(assignment)
		if assignment == "" || strings.HasPrefix(assignment, userID+"-") {
			continue
		}
		kept = append(kept, assignment)
	}
	return strings.Join(kept, ",")
}

// replacePrimaryAssignee updates the first assignee on a task while keeping
// additional team members that were added from the Users tab.
func replacePrimaryAssignee(existing, newPrimary string) string {
	newPrimary = strings.TrimSpace(newPrimary)
	if newPrimary == "" {
		return existing
	}
	newUserID := newPrimary
	if dash := strings.Index(newPrimary, "-"); dash > 0 {
		newUserID = strings.TrimSpace(newPrimary[:dash])
	}

	parts := strings.Split(existing, ",")
	secondaries := make([]string, 0, len(parts))
	for i, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if i == 0 {
			continue
		}
		if strings.HasPrefix(part, newUserID+"-") {
			continue
		}
		secondaries = append(secondaries, part)
	}
	if len(secondaries) == 0 {
		return newPrimary
	}
	return newPrimary + "," + strings.Join(secondaries, ",")
}

func primaryAssigneeFromAssignedTo(assignedTo string) (userID, name string) {
	parts := strings.Split(assignedTo, ",")
	if len(parts) == 0 {
		return "", ""
	}
	first := strings.TrimSpace(parts[0])
	dash := strings.Index(first, "-")
	if dash <= 0 {
		return "", ""
	}
	return strings.TrimSpace(first[:dash]), strings.TrimSpace(first[dash+1:])
}

// When the calendar or scheduling tab changes the primary assignee, swap that
// person in the stage Users table without removing additional team members.
func syncStageUsersForPrimaryChange(store db.Database, stageID, taskName, oldAssigned, newAssigned string) {
	oldID, _ := primaryAssigneeFromAssignedTo(oldAssigned)
	newID, newName := primaryAssigneeFromAssignedTo(newAssigned)
	if oldID == "" || newID == "" || oldID == newID {
		return
	}

	_ = store.Where(
		"stage_id = ? AND user_id = ? AND task_name = ?",
		stageID, oldID, taskName,
	).Delete(&models.StageUser{})

	var existing models.StageUser
	if err := store.Where(
		"stage_id = ? AND user_id = ? AND task_name = ?",
		stageID, newID, taskName,
	).First(&existing); err != nil {
		_ = store.Create(&models.StageUser{
			ID:       uuid.NewString(),
			StageID:  stageID,
			UserID:   newID,
			Name:     newName,
			TaskName: taskName,
		})
		return
	}
	if newName != "" && existing.Name != newName {
		existing.Name = newName
		_ = store.Save(&existing)
	}
}

func syncEstimateStageUsersForPrimaryChange(store db.Database, stageID string, entryID uint, taskName, oldAssigned, newAssigned string) {
	oldID, _ := primaryAssigneeFromAssignedTo(oldAssigned)
	newID, newName := primaryAssigneeFromAssignedTo(newAssigned)
	if oldID == "" || newID == "" || oldID == newID {
		return
	}

	_ = store.Where(
		"stage_id = ? AND user_id = ? AND task_name = ?",
		stageID, oldID, taskName,
	).Delete(&models.EstimateStageUser{})

	var existing models.EstimateStageUser
	if err := store.Where(
		"stage_id = ? AND user_id = ? AND task_name = ?",
		stageID, newID, taskName,
	).First(&existing); err != nil {
		_ = store.Create(&models.EstimateStageUser{
			ID:                 uuid.NewString(),
			StageID:            stageID,
			ResidentialEntryID: entryID,
			UserID:             newID,
			Name:               newName,
			TaskName:           taskName,
		})
		return
	}
	if newName != "" && existing.Name != newName {
		existing.Name = newName
		_ = store.Save(&existing)
	}
}

func hasTaskAssignee(existing, userID string) bool {
	for _, assignment := range strings.Split(existing, ",") {
		if strings.HasPrefix(strings.TrimSpace(assignment), userID+"-") {
			return true
		}
	}
	return false
}

func isPrimaryTaskAssignee(existing, userID string) bool {
	assignments := strings.Split(existing, ",")
	return len(assignments) > 0 && strings.HasPrefix(strings.TrimSpace(assignments[0]), userID+"-")
}

func (h *Handler) AgentWithStages() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := middleware.GetDB(c)
		var stages []models.Stage
		if err := db.Preload("Tasks").
			Preload("Users").
			Order("created_at DESC").
			Find(&stages); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch stages: " + err.Error(),
			})
			return
		}

		// Estimate Project Management tasks live in a separate table
		// (EstimateProjectTask) but share the same Stage row as the Job
		// tracker. Pull them in here so they surface on the calendar the
		// same way Job scheduler tasks do.
		stageIDs := make([]string, 0, len(stages))
		for _, stage := range stages {
			stageIDs = append(stageIDs, stage.ID)
		}
		estimateTasksByStage := make(map[string][]models.EstimateProjectTask)
		if len(stageIDs) > 0 {
			var estimateTasks []models.EstimateProjectTask
			if err := db.Where("stage_id IN ?", stageIDs).Find(&estimateTasks); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to fetch estimate project tasks: " + err.Error(),
				})
				return
			}
			for _, et := range estimateTasks {
				estimateTasksByStage[et.StageID] = append(estimateTasksByStage[et.StageID], et)
			}
		}

		leadIDMap := make(map[string]bool)
		for _, stage := range stages {
			if stage.LeadID != "" {
				leadIDMap[stage.LeadID] = true
			}
		}
		if len(leadIDMap) == 0 {
			c.JSON(http.StatusOK, gin.H{"data": []gin.H{}})
			return
		}
		leadIDs := make([]string, 0, len(leadIDMap))
		for id := range leadIDMap {
			leadIDs = append(leadIDs, id)
		}
		var leads []models.Lead
		if err := db.
			Select("id, job_name, status, sales1_person_id, details").
			Where("id IN ?", leadIDs).
			Find(&leads); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch leads: " + err.Error(),
			})
			return
		}
		shippingAddressIDs := make(map[uint]bool)
		for _, lead := range leads {
			if lead.Details.ShippingAddressID != nil {
				shippingAddressIDs[*lead.Details.ShippingAddressID] = true
			}
		}
		addressMap := make(map[uint]models.Address)
		if len(shippingAddressIDs) > 0 {
			addressIDs := make([]uint, 0, len(shippingAddressIDs))
			for id := range shippingAddressIDs {
				addressIDs = append(addressIDs, id)
			}
			var addresses []models.Address
			if err := db.Where("id IN ?", addressIDs).Find(&addresses); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to fetch shipping addresses: " + err.Error(),
				})
				return
			}
			for _, addr := range addresses {
				addressMap[addr.ID] = addr
			}
			for i := range leads {
				if lead := &leads[i]; lead.Details.ShippingAddressID != nil {
					if addr, exists := addressMap[*lead.Details.ShippingAddressID]; exists {
						lead.Details.ShippingAddress = &addr
					}
				}
			}
		}
		agentIDMap := make(map[string]bool)
		for _, stage := range stages {
			for _, task := range stage.Tasks {
				if task.AssignedTo != "" {
					parts := strings.Split(task.AssignedTo, "-")
					if len(parts) > 0 && parts[0] != "" {
						agentIDMap[parts[0]] = true
					}
				}
			}
			for _, et := range estimateTasksByStage[stage.ID] {
				if et.AssignedTo != "" {
					parts := strings.Split(et.AssignedTo, "-")
					if len(parts) > 0 && parts[0] != "" {
						agentIDMap[parts[0]] = true
					}
				}
			}
		}

		if len(agentIDMap) == 0 {
			c.JSON(http.StatusOK, gin.H{"data": []gin.H{}})
			return
		}

		agentIDs := make([]string, 0, len(agentIDMap))
		for id := range agentIDMap {
			agentIDs = append(agentIDs, id)
		}

		var agents []models.Agent
		if err := db.Where("id IN ?", agentIDs).Find(&agents); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch agents: " + err.Error(),
			})
			return
		}

		// FIXED: Changed to store all stages instead of just one per JobID
		leadStages := make(map[string][]models.Stage)
		for _, stage := range stages {
			if stage.LeadID == "" {
				continue
			}
			leadStages[stage.LeadID] = append(leadStages[stage.LeadID], stage)
		}

		leadMap := make(map[string]models.Lead)
		for _, lead := range leads {
			leadMap[lead.ID] = lead
		}

		var response []gin.H
		for _, agent := range agents {
			agentData := gin.H{"agent": agent}
			leadsData := []gin.H{}

			processedLeads := make(map[string]bool)

			for _, stage := range stages {
				if stage.LeadID == "" {
					continue
				}

				hasAssignedTasks := false
				for _, task := range stage.Tasks {
					// Skip StageTask rows that are copies of Estimate Project
					// Management tasks — the original EstimateProjectTask row
					// is merged in separately below, so counting both would
					// duplicate the same logical task on the calendar.
					if task.SourceEstimateTaskID != "" {
						continue
					}
					parts := strings.Split(task.AssignedTo, "-")
					assignedID := ""
					if len(parts) > 0 {
						assignedID = parts[0]
					}
					if assignedID == agent.ID {
						hasAssignedTasks = true
						break
					}
				}
				if !hasAssignedTasks {
					for _, et := range estimateTasksByStage[stage.ID] {
						parts := strings.Split(et.AssignedTo, "-")
						assignedID := ""
						if len(parts) > 0 {
							assignedID = parts[0]
						}
						if assignedID == agent.ID {
							hasAssignedTasks = true
							break
						}
					}
				}

				if !hasAssignedTasks {
					continue
				}

				if processedLeads[stage.LeadID] {
					continue
				}
				processedLeads[stage.LeadID] = true

				lead, exists := leadMap[stage.LeadID]
				if !exists {
					continue
				}

				shippingAddress := gin.H{}
				if lead.Details.ShippingAddress != nil {
					shippingAddress = gin.H{
						"id":        lead.Details.ShippingAddress.ID,
						"street":    lead.Details.ShippingAddress.Street,
						"city":      lead.Details.ShippingAddress.City,
						"state":     lead.Details.ShippingAddress.State,
						"zip":       lead.Details.ShippingAddress.Zip,
						"county":    lead.Details.ShippingAddress.County,
						"lat":       lead.Details.ShippingAddress.Lat,
						"lon":       lead.Details.ShippingAddress.Lng,
						"formatted": lead.Details.ShippingAddress.Formatted,
					}
				}

				leadData := gin.H{
					"id":       lead.ID,
					"name":     lead.JobName,
					"shipping": shippingAddress,
					"status":   lead.Status,
					"details":  lead.Details,
				}

				// FIXED: Now iterates through all stages for the lead
				stagesData := []gin.H{}
				if stagesList, stageExists := leadStages[lead.ID]; stageExists {
					for _, stage := range stagesList {
						tasksResponse := make([]map[string]interface{}, 0)
						for _, task := range stage.Tasks {
							// Skip estimate-origin copies — the original
							// EstimateProjectTask is appended below, so
							// emitting both produced the duplicate calendar
							// entry (same name, same completed/cancelled
							// status) after converting an estimate to a job.
							if task.SourceEstimateTaskID != "" {
								continue
							}
							parts := strings.Split(task.AssignedTo, "-")
							assignedID := ""
							if len(parts) > 0 {
								assignedID = parts[0]
							}

							if assignedID != agent.ID {
								continue
							}

							taskResponse := map[string]interface{}{
								"id":          task.ID,
								"stageId":     task.StageID,
								"name":        task.Name,
								"dueDate":     task.DueDate.Format("2006-01-02"),
								"startDate":   task.StartDate.Format("2006-01-02"),
								"duration":    task.Duration,
								"completed":   task.Completed,
								"cancelled":   task.Cancelled,
								"notes":       task.NotesHistory,
								"photos":      task.Photos,
								"createdAt":   task.CreatedAt.Format(time.RFC3339),
								"updatedAt":   task.UpdatedAt.Format(time.RFC3339),
								"assigned_to": task.AssignedTo,
								"projects":    task.Projects,
							}
							// The task's own address snapshot (taken when it was created)
							// takes priority over the lead's current shipping address, so
							// this task's map pin doesn't move when the lead's address is
							// changed later. Older tasks created before this snapshot
							// existed simply have no "address" key here, and the frontend
							// falls back to the lead's live shipping address for those.
							if len(task.Address) > 0 {
								var addrSnapshot map[string]interface{}
								if err := json.Unmarshal(task.Address, &addrSnapshot); err == nil {
									taskResponse["address"] = addrSnapshot
								}
							}
							// StartTime/EndTime are *time.Time and can be nil
							// (e.g. a task that was never scheduled). Previously
							// this loop only ran for leads with at least one
							// agent-assigned StageTask, which happened to avoid
							// nil rows in practice; now that leads can also
							// qualify via an assigned EstimateProjectTask, this
							// loop can reach StageTask rows it never hit before,
							// so the nil guard is required here.
							if task.StartTime != nil {
								taskResponse["startTime"] = task.StartTime.UTC().Format(time.RFC3339)
							}
							if task.EndTime != nil {
								taskResponse["endTime"] = task.EndTime.UTC().Format(time.RFC3339)
							}
							tasksResponse = append(tasksResponse, taskResponse)
						}

						// Merge in Estimate Project Management tasks for this
						// same stage (separate table, same Stage row).
						for _, et := range estimateTasksByStage[stage.ID] {
							parts := strings.Split(et.AssignedTo, "-")
							assignedID := ""
							if len(parts) > 0 {
								assignedID = parts[0]
							}
							if assignedID != agent.ID {
								continue
							}

							startTime, endTime, dueDate, startDate := parseEstimateTaskTimes(et)
							taskResponse := map[string]interface{}{
								"id":          et.ID,
								"stageId":     et.StageID,
								"name":        et.Name,
								"dueDate":     dueDate.Format("2006-01-02"),
								"startDate":   startDate.Format("2006-01-02"),
								"duration":    et.Duration,
								"completed":   et.Completed,
								"cancelled":   et.Cancelled,
								"notes":       et.NotesHistory,
								"photos":      et.Photos,
								"createdAt":   et.CreatedAt.Format(time.RFC3339),
								"updatedAt":   et.UpdatedAt.Format(time.RFC3339),
								"assigned_to": et.AssignedTo,
								"projects":    et.Projects,
								"source":      "estimate",
							}
							if startTime != nil {
								taskResponse["startTime"] = startTime.UTC().Format(time.RFC3339)
							}
							if endTime != nil {
								taskResponse["endTime"] = endTime.UTC().Format(time.RFC3339)
							}
							tasksResponse = append(tasksResponse, taskResponse)
						}
						stageResponse := gin.H{
							"ID":                 stage.ID,
							"LeadID":             stage.LeadID,
							"Name":               stage.Name,
							"Description":        stage.Description,
							"State":              stage.State,
							"Date":               stage.Date,
							"JobID":              stage.JobID,
							"Notes":              stage.Notes,
							"Team":               stage.Team,
							"CreatedAt":          stage.CreatedAt,
							"ResidentialEntryID": stage.ResidentialEntryID,
							"UpdatedAt":          stage.UpdatedAt,
							// "Materials":          stage.Materials,
							"Users": stage.Users,
							"Tasks": tasksResponse,
						}

						stagesData = append(stagesData, gin.H{
							"id":    stage.JobID,
							"title": stage.Name,
							"stage": stageResponse,
						})
					}
				}
				leadData["stages"] = stagesData
				leadsData = append(leadsData, leadData)
			}

			agentData["leads"] = leadsData
			response = append(response, agentData)
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "Agents with leads and stages retrieved successfully",
			"data":    response,
		})
	}
}
func (h *Handler) GetAllStageTasks(c *gin.Context) {
	db := middleware.GetDB(c)

	var tasks []models.StageTask
	if err := db.Preload("Stage").Find(&tasks); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to retrieve tasks",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": tasks,
	})
}
func (h *Handler) AddStageNote() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := middleware.GetDB(c)

		stageID := c.Param("stage_id")
		leadIDStr := c.Param("lead_id")
		// Parse request body
		var req struct {
			Note string `json:"note" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
			return
		}

		// Find stage by ID and lead ID
		var stage models.Stage
		if err := db.Where("id = ? AND residential_entry_id = ?", stageID, leadIDStr).First(&stage); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Stage not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error finding stage: " + err.Error()})
			return
		}

		// Use the model's helper method to add the note
		stage.AddNote(req.Note)

		// Save updated notes
		if err := db.Model(&stage).Update("notes", stage.Notes); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save note"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Note added successfully",
			"notes":   stage.GetNotesArray(),
		})
	}
}

func (h *Handler) DeleteStageNote() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := middleware.GetDB(c)

		stageID := c.Param("stage_id")
		leadIDStr := c.Param("lead_id")
		noteIndexStr := c.Param("note_id")
		var err error

		noteIndex, err := strconv.Atoi(noteIndexStr) // First declaration is fine
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid note index"})
			return
		}

		// Find stage by ID and lead ID
		var stage models.Stage
		if err = db.Where("id = ? AND residential_entry_id = ?", stageID, leadIDStr).First(&stage); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Stage not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
			return
		}

		// Remember the note text so the linked scratchpad can be removed too
		deletedText := ""
		if existingNotes := stage.GetNotesArray(); noteIndex >= 0 && noteIndex < len(existingNotes) {
			deletedText = existingNotes[noteIndex]
		}

		// Use the model's helper method to delete the note
		stage.DeleteNote(noteIndex)

		// Save the updated stage back to the database
		if err = db.Save(&stage); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save updated notes"})
			return
		}

		// Keep the scratchpad in sync
		if deletedText != "" {
			removeScratchpadLinks(db, "stage", stage.ID, deletedText, 0)
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Note deleted successfully",
			"notes":   stage.GetNotesArray(),
		})
	}
}

func (h *Handler) GetAllStages() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := middleware.GetDB(c)
		var stages []models.Stage
		if err := db.Preload("Tasks").
			Preload("Users").
			Find(&stages); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stages: " + err.Error()})
			return
		}

		// Format response
		formattedStages := make([]map[string]interface{}, 0, len(stages))
		for _, stage := range stages {
			stageResponse := map[string]interface{}{
				"ID":                 stage.ID,
				"LeadID":             stage.LeadID,
				"JobID":              stage.JobID,
				"Name":               stage.Name,
				"Description":        stage.Description,
				"State":              stage.State,
				"Date":               stage.Date,
				"Notes":              stage.GetNotesArray(),
				"Team":               stage.Team,
				"CreatedAt":          stage.CreatedAt,
				"UpdatedAt":          stage.UpdatedAt,
				"ResidentialEntryID": stage.ResidentialEntryID,
				// "Materials":          stage.Materials,
				"Users": stage.Users,
				"Tasks": make([]map[string]interface{}, 0, len(stage.Tasks)),
			}
			for _, task := range stage.Tasks {
				taskResponse := map[string]interface{}{
					"id":        task.ID,
					"stageId":   task.StageID,
					"name":      task.Name,
					"dueDate":   task.DueDate.Format("2006-01-02"),
					"startDate": task.StartDate.Format("2006-01-02"),
					"startTime": task.StartTime.UTC().Format("15:04Z"),
					"endTime":   task.EndTime.UTC().Format("15:04Z"),
					"duration":  task.Duration,
					"completed": task.Completed,
					"cancelled": task.Cancelled,
					// "timers":     task.Timers,
					"assignedTo": task.AssignedTo,
					"photos":     task.Photos,
					"createdAt":  task.CreatedAt.Format(time.RFC3339),
					"updatedAt":  task.UpdatedAt.Format(time.RFC3339),
				}
				stageResponse["Tasks"] = append(stageResponse["Tasks"].([]map[string]interface{}), taskResponse)
			}

			formattedStages = append(formattedStages, stageResponse)
		}

		c.JSON(http.StatusOK, gin.H{
			"stages": formattedStages,
			"total":  len(formattedStages),
		})
	}
}

func (h *Handler) DeleteTaskRecord() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := middleware.GetDB(c)
		taskID := c.Param("id")

		tx := db.Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Unexpected error occurred"})
			}
		}()

		var task models.StageTask
		if err := tx.Preload("Stage").Where("id = ?", taskID).First(&task); err != nil {
			// Not a Job tracker task — fall back to Estimate Project task
			var estTask models.EstimateProjectTask
			if estErr := tx.Where("id = ?", taskID).First(&estTask); estErr != nil {
				tx.Rollback()
				c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
				return
			}

			if !estTask.Cancelled {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{"error": "Only cancelled tasks can be deleted"})
				return
			}

			// Remove any StageTask copies created from this estimate task
			if err := tx.Where("source_estimate_task_id = ?", estTask.ID).Delete(&models.StageTask{}); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete linked tasks"})
				return
			}

			if err := tx.Delete(&estTask); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
				return
			}

			if err := tx.Where("stage_id = ? AND residential_entry_id = ? AND task_name = ?", estTask.StageID, estTask.ResidentialEntryID, estTask.Name).Delete(&models.EstimateStageUser{}); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove task assignees"})
				return
			}

			if err := tx.Commit(); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit task deletion"})
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "Task deleted", "taskId": taskID})
			return
		}

		if !task.Cancelled {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Only cancelled tasks can be deleted"})
			return
		}

		stageID := task.StageID
		taskName := task.Name
		jobID := task.Stage.JobID
		leadID := task.Stage.LeadID

		if err := tx.Delete(&task); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
			return
		}

		// Remove the stage-user assignment created for this task (keyed by stage + task name)
		if err := tx.Where("stage_id = ? AND task_name = ?", stageID, taskName).Delete(&models.StageUser{}); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove task assignees"})
			return
		}

		userID, userName := middleware.GetRequestUser(c)
		event := activity.ActivityEvent{
			EntityType: "job",
			EntityID:   jobID,
			Action:     "task deleted",
			UserID:     userID,
			UserName:   userName,
			Payload: map[string]any{
				"task_name":  taskName,
				"stage_name": task.Stage.Name,
				"deleted":    true,
				"updated_by": userName,
			},
		}
		if err := activity.InsertActivityEvent(c.Request.Context(), tx.GetUnderlying(), event); err != nil {
			log.Printf("activity event publish failed: %v", err)
		}

		if err := tx.Commit(); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit task deletion"})
			return
		}

		// Keep the job snapshot / reports in sync
		var lead models.Lead
		if err := db.Where("id = ?", leadID).First(&lead); err == nil {
			jobService := services.NewJobService()
			if syncErr := jobService.UpsertJobFromLead(db, lead); syncErr != nil {
				fmt.Printf("WARN: UpsertJobFromLead failed after task delete for lead %s: %v\n", leadID, syncErr)
			}
		}

		c.JSON(http.StatusOK, gin.H{"message": "Task deleted", "taskId": taskID})
	}
}
