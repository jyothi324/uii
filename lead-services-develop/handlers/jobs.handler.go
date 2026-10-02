package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tuliporg/agent-management-api/internal/activity"
	"github.com/tuliporg/agent-management-api/middleware"
	"github.com/tuliporg/agent-management-api/services"
)

type ActivityEventRow struct {
	ID         uint      `json:"id"`
	EntityType string    `json:"entity_type"`
	EntityID   string    `json:"entity_id"`
	Action     string    `json:"action"`
	UserID     string    `json:"user_id"`
	UserName   string    `json:"user_name"`
	Payload    []byte    `json:"-"`
	CreatedAt  time.Time `json:"created_at"`
}

type ActivityEventResponse struct {
	ID         uint           `json:"id"`
	EntityType string         `json:"entity_type"`
	EntityID   string         `json:"entity_id"`
	Action     string         `json:"action"`
	UserID     string         `json:"user_id"`
	UserName   string         `json:"user_name"`
	Payload    map[string]any `json:"payload"`
	CreatedAt  time.Time      `json:"created_at"`
}

type JobHandler struct {
	jobService *services.JobService
}

func NewJobHandler() *JobHandler {
	return &JobHandler{
		jobService: services.NewJobService(),
	}
}

func (h *JobHandler) GetAllJobs(c *gin.Context) {
	db := middleware.GetDB(c)

	status := c.Query("status")

	pageStr := c.Query("page")
	limitStr := c.Query("limit")

	page := 1
	limit := 0 // 0 = no limit → return everything

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p >= 1 {
			page = p
		}
	}
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l >= 1 && l <= 1000 {
			limit = l
		}
	}

	search := c.Query("search")

	jobs, totalCount, err := h.jobService.GetAllJobs(db, status, page, limit, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch jobs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Jobs retrieved successfully",
		"data":        jobs,
		"total_count": totalCount,
		"page":        page,
		"limit":       limit,
	})
}

func (h *JobHandler) SyncJob(c *gin.Context) {
	db := middleware.GetDB(c)
	jobID := c.Param("id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "job id is required"})
		return
	}

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unexpected error during job sync"})
		}
	}()

	result, err := h.jobService.SyncJobByID(tx, jobID)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	userID, userName := middleware.GetRequestUser(c)
	event := activity.ActivityEvent{
		EntityType: "job",
		EntityID:   jobID,
		Action:     "updated",
		UserID:     userID,
		UserName:   userName,
		Payload: map[string]any{
			"updated_by": userName,
		},
	}
	if err := activity.InsertActivityEvent(c.Request.Context(), tx.GetUnderlying(), event); err != nil {
		log.Printf("activity event publish failed: %v", err)
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit job sync"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Job synced successfully",
		"data":    result,
	})
}

func (h *JobHandler) GetJobActivity(c *gin.Context) {
	db := middleware.GetDB(c)
	jobId := c.Param("id")
	if jobId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID",
			"code":    "VALIDATION_ERROR",
			"details": "Job ID parameter is required",
		})
		return
	}

	var rows []ActivityEventRow
	if err := db.GetUnderlying().
		Table("activity_events").
		Select("id, entity_type, entity_id, action, user_id, user_name, payload, created_at").
		Where("entity_type = ? AND entity_id = ?", "job", jobId).
		Order("created_at DESC").
		Scan(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch job activity events",
			"code":    "DATABASE_ERROR",
			"details": err.Error(),
		})
		return
	}

	events := make([]ActivityEventResponse, 0, len(rows))
	for _, row := range rows {
		var payload map[string]any
		if len(row.Payload) > 0 {
			if err := json.Unmarshal(row.Payload, &payload); err != nil {
				payload = nil
			}
		}
		events = append(events, ActivityEventResponse{
			ID:         row.ID,
			EntityType: row.EntityType,
			EntityID:   row.EntityID,
			Action:     row.Action,
			UserID:     row.UserID,
			UserName:   row.UserName,
			Payload:    payload,
			CreatedAt:  row.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Job activity events retrieved successfully",
		"data":    events,
	})
}
