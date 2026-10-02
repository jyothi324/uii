package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tuliporg/agent-management-api/db"
	"github.com/tuliporg/agent-management-api/internal/activity"
	"github.com/tuliporg/agent-management-api/middleware"
	"github.com/tuliporg/agent-management-api/models"
	"github.com/tuliporg/agent-management-api/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// presignEstimateTaskPhotos turns each stored S3 URL/key on an
// EstimateProjectTask into a short-lived presigned GET URL, mirroring how
// GetStagesByLeadID already does this for Job scheduler tasks. Without this,
// the raw https://...amazonaws.com/... URL stored by UploadTaskPhotos 403s
// when the frontend requests it directly, since the bucket isn't public.
// Uses the same env vars and per-tenant bucket naming as FileHandler.
func presignEstimateTaskPhotos(c *gin.Context, keys []string) []string {
	if len(keys) == 0 {
		return keys
	}

	awsRegion := os.Getenv("AWS_REGION")
	if awsRegion == "" {
		awsRegion = "us-east-2"
	}
	bucket := os.Getenv("S3_BUCKET")
	if bucket == "" {
		bucket = "tulipfiles"
	}
	if subdomain := middleware.GetRawSubdomain(c); subdomain != "" && subdomain != "default" {
		bucket = fmt.Sprintf("tulip-%s-files", subdomain)
	}

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(awsRegion),
		Credentials: credentials.NewStaticCredentials(
			os.Getenv("AWS_ACCESS_KEY_ID"),
			os.Getenv("AWS_SECRET_ACCESS_KEY"),
			"",
		),
	})
	if err != nil {
		log.Printf("presignEstimateTaskPhotos: failed to create AWS session: %v", err)
		return keys
	}
	s3Client := s3.New(sess)

	presigned := make([]string, 0, len(keys))
	for _, key := range keys {
		s3Key := key
		// Handle full URLs already stored (current UploadTaskPhotos behavior)
		// by stripping down to the bare S3 key, same trick FileHandler's own
		// generatePresignedURL uses for legacy stored values.
		if strings.HasPrefix(s3Key, "https://") || strings.HasPrefix(s3Key, "http://") {
			parts := strings.Split(s3Key, "/")
			if len(parts) >= 4 {
				s3Key = strings.Join(parts[3:], "/")
			}
		}
		req, _ := s3Client.GetObjectRequest(&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(s3Key),
		})
		url, err := req.Presign(1 * time.Hour)
		if err != nil {
			log.Printf("presignEstimateTaskPhotos: failed to presign %q: %v", s3Key, err)
			continue
		}
		presigned = append(presigned, url)
	}
	return presigned
}

// pmContext bundles the resolved Estimate, target ResidentialEntry, and its
// "Template" stage — the only stage Estimate Project Management may touch.
type pmContext struct {
	Estimate models.Estimate
	EntryID  uint
	Stage    models.Stage
}

func loadPMContext(store db.Database, estimateID string, residentialEntryIDParam string) (*pmContext, int, string) {
	var estimate models.Estimate
	if err := store.Preload("ResidentialEntries").Preload("Lead").
		Where("id = ?", estimateID).First(&estimate); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("[pm-debug] estimate %q not found", estimateID)
			return nil, http.StatusNotFound, "Estimate not found"
		}
		log.Printf("[pm-debug] estimate %q fetch error: %v", estimateID, err)
		return nil, http.StatusInternalServerError, "Failed to fetch estimate"
	}

	if len(estimate.ResidentialEntries) == 0 {
		log.Printf("[pm-debug] estimate %q has zero residential entries", estimateID)
		return nil, http.StatusBadRequest, "Estimate has no residential entries"
	}

	var entryID uint
	if residentialEntryIDParam != "" {
		parsed, err := strconv.ParseUint(residentialEntryIDParam, 10, 64)
		if err != nil {
			log.Printf("[pm-debug] invalid residential_entry_id param %q: %v", residentialEntryIDParam, err)
			return nil, http.StatusBadRequest, "Invalid residential_entry_id"
		}
		entryID = uint(parsed)
	} else {
		entryID = estimate.ResidentialEntries[0].ID
	}

	found := false
	for _, e := range estimate.ResidentialEntries {
		if e.ID == entryID {
			found = true
			break
		}
	}
	if !found {
		// The requested entryID doesn't match — this can happen if the
		// frontend cached a stale ID (e.g. the residential entry was
		// recreated on a subsequent save with a new ID). Rather than hard
		// failing, fall back to the estimate's first residential entry so
		// Project Management still opens, matching how Jobs never blocks
		// on a stale ID.
		availableIDs := make([]uint, 0, len(estimate.ResidentialEntries))
		for _, e := range estimate.ResidentialEntries {
			availableIDs = append(availableIDs, e.ID)
		}
		log.Printf("[pm-debug] requested entryID %d not found on estimate %q; available IDs: %v — falling back to first entry", entryID, estimateID, availableIDs)
		entryID = estimate.ResidentialEntries[0].ID
	}

	var stage models.Stage
	err := store.Where("residential_entry_id = ? AND name = ?", entryID, "Template").First(&stage)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("[pm-debug] stage lookup error for entryID %d: %v", entryID, err)
		return nil, http.StatusInternalServerError, "Failed to fetch Template stage"
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// No Template stage exists yet for this residential entry (e.g. it
		// was created through a path that skips default stage creation).
		// Auto-provision it here so Project Management behaves the same as
		// Jobs, which always guarantee stages exist — without needing to
		// touch the main estimate handler.
		//
		// Reuse the lead's job ID first: a lead has one job, so this entry's
		// Template stage must not land on a different job than its siblings.
		// Then fall back to a sibling entry's job ID, and only mint a new one
		// when the lead has no job at all.
		jobID := resolveLeadJobID(store, estimate.Lead.ID)
		if jobID == "" {
			siblingEntryIDs := make([]uint, 0, len(estimate.ResidentialEntries))
			for _, e := range estimate.ResidentialEntries {
				if e.ID != entryID {
					siblingEntryIDs = append(siblingEntryIDs, e.ID)
				}
			}
			if len(siblingEntryIDs) > 0 {
				var siblingStage models.Stage
				if err := store.Where("residential_entry_id IN ? AND job_id != ?", siblingEntryIDs, "").First(&siblingStage); err == nil && siblingStage.JobID != "" {
					jobID = siblingStage.JobID
				}
			}
		}
		if jobID == "" {
			jobID = utils.GenerateEmployeeID(store)
		}

		log.Printf("[pm-debug] no Template stage for entryID %d — creating one now with jobID %s", entryID, jobID)
		newStage := models.Stage{
			ID:                 uuid.NewString(),
			LeadID:             estimate.Lead.ID,
			JobID:              jobID,
			Name:               "Template",
			State:              "In Progress",
			Description:        "Prepare and finalize the template for the project",
			ResidentialEntryID: entryID,
		}
		if err := store.Create(&newStage); err != nil {
			log.Printf("[pm-debug] failed to auto-create Template stage for entryID %d: %v", entryID, err)
			return nil, http.StatusInternalServerError, "Failed to create Template stage"
		}
		stage = newStage
	}

	return &pmContext{Estimate: estimate, EntryID: entryID, Stage: stage}, 0, ""
}

// GET /estimates/:id/project-management?residential_entry_id=123
func (h *Handler) GetEstimateProjectManagement(c *gin.Context) {
	db := middleware.GetDB(c)
	estimateID := c.Param("id")

	ctx, status, msg := loadPMContext(db, estimateID, c.Query("residential_entry_id"))
	if ctx == nil {
		c.JSON(status, ErrorResponse{Error: msg, Code: "VALIDATION_ERROR"})
		return
	}

	var tasks []models.EstimateProjectTask
	if err := db.Where("residential_entry_id = ?", ctx.EntryID).Find(&tasks); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch tasks", Code: "DATABASE_ERROR", Details: err.Error()})
		return
	}
	for i := range tasks {
		tasks[i].Photos = presignEstimateTaskPhotos(c, tasks[i].Photos)
	}

	var users []models.EstimateStageUser
	if err := db.Where("residential_entry_id = ?", ctx.EntryID).Find(&users); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch stage users", Code: "DATABASE_ERROR", Details: err.Error()})
		return
	}

	var notes []models.EstimateStageNote
	if err := db.Where("residential_entry_id = ?", ctx.EntryID).Find(&notes); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch notes", Code: "DATABASE_ERROR", Details: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Estimate project management data retrieved successfully",
		"data": gin.H{
			"residentialEntryId": ctx.EntryID,
			"stage":              ctx.Stage,
			"tasks":              tasks,
			"users":              users,
			"notes":              notes,
		},
	})
}

// pmMutationRequest is the single request shape for every Estimate Project
// Management write. The "action" field selects which fields are read —
// this keeps Project Management down to two HTTP endpoints (GET to fetch,
// POST to mutate) instead of one endpoint per action, with no change to
// what each action actually validates or writes.
type pmMutationRequest struct {
	Action             string `json:"action"` // "status" | "task" | "timer" | "user" | "note_add" | "note_delete"
	ResidentialEntryID uint   `json:"residentialEntryId"`

	// action = "status"
	Status string `json:"status"`

	// action = "task"
	TaskID     string                    `json:"taskId"`
	Name       string                    `json:"name"`
	StartDate  string                    `json:"startDate"`
	StartTime  string                    `json:"startTime"`
	EndTime    string                    `json:"endTime"`
	Duration   int                       `json:"duration"`
	AssignedTo string                    `json:"assignedTo"`
	Notes      string                    `json:"notes"`
	Projects   []models.StageTaskProject `json:"projects"`

	// action = "timer"
	TimerTaskID string `json:"timerTaskId"`
	TimerAction string `json:"timerAction"` // start | pause | complete | cancel

	// action = "user"
	UserID   string `json:"userId"`
	UserName string `json:"userName"`
	TaskName string `json:"taskName"`

	// action = "note_add"
	Note string `json:"note"`

	// action = "note_delete"
	NoteID string `json:"noteId"`

	// action = "user_delete"
	// (reuses UserID field above)
}

// MutateEstimateProjectManagement is the single write endpoint for Estimate
// Project Management. It replaces six separate routes (status update, task
// save, timer start/pause/complete/cancel, user upsert, note add, note
// delete) with one POST that dispatches on body.action.
func (h *Handler) MutateEstimateProjectManagement(c *gin.Context) {
	store := middleware.GetDB(c)
	estimateID := c.Param("id")

	var body pmMutationRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid input", Code: "VALIDATION_ERROR", Details: err.Error()})
		return
	}

	switch body.Action {
	case "status":
		h.pmUpdateStatus(c, store, estimateID, body)
	case "task":
		h.pmSaveTask(c, store, estimateID, body)
	case "timer":
		h.pmTaskTimerAction(c, store, body)
	case "user":
		h.pmUpsertUser(c, store, estimateID, body)
	case "note_add":
		h.pmAddNote(c, store, estimateID, body)
	case "note_delete":
		h.pmDeleteNote(c, store, estimateID, body)
	case "user_delete":
		h.pmDeleteUser(c, store, estimateID, body)
	default:
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Unknown or missing action", Code: "VALIDATION_ERROR"})
	}
}

func (h *Handler) pmUpdateStatus(c *gin.Context, store db.Database, estimateID string, body pmMutationRequest) {
	allowed := map[string]bool{"Not Started": true, "In Progress": true, "Completed": true, "Cancelled": true}
	if !allowed[body.Status] {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid status value", Code: "VALIDATION_ERROR"})
		return
	}

	ctx, status, msg := loadPMContext(store, estimateID, strconv.FormatUint(uint64(body.ResidentialEntryID), 10))
	if ctx == nil {
		c.JSON(status, ErrorResponse{Error: msg, Code: "VALIDATION_ERROR"})
		return
	}

	if err := store.Model(&models.Stage{}).Where("id = ?", ctx.Stage.ID).Update("state", body.Status); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update stage status", Code: "DATABASE_ERROR", Details: err.Error()})
		return
	}

	userID, userName := middleware.GetRequestUser(c)
	event := activity.ActivityEvent{
		EntityType: "estimate",
		EntityID:   ctx.Estimate.ID,
		Action:     "pm_status_updated",
		UserID:     userID,
		UserName:   userName,
		Payload:    map[string]any{"lead_name": ctx.Estimate.Lead.JobName, "status": body.Status},
	}
	if err := activity.InsertActivityEvent(c.Request.Context(), store.GetUnderlying(), event); err != nil {
		log.Printf("activity event publish failed: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Template stage status updated successfully"})
}

func (h *Handler) pmSaveTask(c *gin.Context, store db.Database, estimateID string, body pmMutationRequest) {
	ctx, status, msg := loadPMContext(store, estimateID, strconv.FormatUint(uint64(body.ResidentialEntryID), 10))
	if ctx == nil {
		c.JSON(status, ErrorResponse{Error: msg, Code: "VALIDATION_ERROR"})
		return
	}

	if body.TaskID != "" {
		var task models.EstimateProjectTask
		if err := store.Where("id = ? AND residential_entry_id = ?", body.TaskID, ctx.EntryID).First(&task); err != nil {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Task not found", Code: "NOT_FOUND"})
			return
		}
		oldAssignedTo := task.AssignedTo
		task.Name = body.Name
		task.StartDate = body.StartDate
		task.StartTime = body.StartTime
		task.EndTime = body.EndTime
		task.Duration = body.Duration
		if strings.TrimSpace(body.AssignedTo) != "" && !strings.Contains(body.AssignedTo, ",") {
			task.AssignedTo = replacePrimaryAssignee(task.AssignedTo, strings.TrimSpace(body.AssignedTo))
		} else {
			task.AssignedTo = body.AssignedTo
		}
		task.Projects = body.Projects
		if body.Notes != "" {
			task.NotesHistory = append(task.NotesHistory, models.NoteEntry{
				Text:      body.Notes,
				CreatedAt: time.Now().UTC().Format(time.RFC3339),
			})
		}
		if err := store.Save(&task); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update task", Code: "DATABASE_ERROR", Details: err.Error()})
			return
		}
		if oldAssignedTo != task.AssignedTo {
			syncEstimateStageUsersForPrimaryChange(store, ctx.Stage.ID, ctx.EntryID, task.Name, oldAssignedTo, task.AssignedTo)
		}
		c.JSON(http.StatusOK, gin.H{"message": "Task updated successfully", "task": task})
		return
	}

	newTask := models.EstimateProjectTask{
		ID:                 uuid.NewString(),
		StageID:            ctx.Stage.ID,
		ResidentialEntryID: ctx.EntryID,
		Name:               body.Name,
		StartDate:          body.StartDate,
		StartTime:          body.StartTime,
		EndTime:            body.EndTime,
		Duration:           body.Duration,
		AssignedTo:         body.AssignedTo,
		Projects:           body.Projects,
	}
	if body.Notes != "" {
		newTask.NotesHistory = []models.NoteEntry{{Text: body.Notes, CreatedAt: time.Now().UTC().Format(time.RFC3339)}}
	}
	if err := store.Create(&newTask); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create task", Code: "DATABASE_ERROR", Details: err.Error()})
		return
	}

	userID, userName := middleware.GetRequestUser(c)
	event := activity.ActivityEvent{
		EntityType: "estimate",
		EntityID:   ctx.Estimate.ID,
		Action:     "pm_task_added",
		UserID:     userID,
		UserName:   userName,
		Payload:    map[string]any{"lead_name": ctx.Estimate.Lead.JobName, "task_name": newTask.Name},
	}
	if err := activity.InsertActivityEvent(c.Request.Context(), store.GetUnderlying(), event); err != nil {
		log.Printf("activity event publish failed: %v", err)
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Task created successfully", "task": newTask})
}

func (h *Handler) pmTaskTimerAction(c *gin.Context, store db.Database, body pmMutationRequest) {
	if body.TimerTaskID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "timerTaskId is required", Code: "VALIDATION_ERROR"})
		return
	}

	var task models.EstimateProjectTask
	if err := store.First(&task, "id = ?", body.TimerTaskID); err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Task not found", Code: "NOT_FOUND"})
		return
	}

	now := time.Now()
	switch body.TimerAction {
	case "start":
		task.IsRunning = true
		task.LastResumedAt = &now
	case "pause":
		if task.IsRunning && task.LastResumedAt != nil {
			task.Elapsed += int(now.Sub(*task.LastResumedAt).Seconds())
		}
		task.IsRunning = false
		task.LastResumedAt = nil
	case "complete":
		if task.IsRunning && task.LastResumedAt != nil {
			task.Elapsed += int(now.Sub(*task.LastResumedAt).Seconds())
		}
		task.IsRunning = false
		task.Completed = true
		task.LastResumedAt = nil
	case "cancel":
		if task.IsRunning && task.LastResumedAt != nil {
			task.Elapsed += int(now.Sub(*task.LastResumedAt).Seconds())
		}
		task.IsRunning = false
		task.Cancelled = true
		task.LastResumedAt = nil
	default:
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid timer action", Code: "VALIDATION_ERROR"})
		return
	}

	if err := store.Save(&task); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update task timer", Code: "DATABASE_ERROR", Details: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       fmt.Sprintf("Task %s successful", body.TimerAction),
		"task":          task,
		"total_elapsed": task.Elapsed,
	})
}

func (h *Handler) pmUpsertUser(c *gin.Context, store db.Database, estimateID string, body pmMutationRequest) {
	ctx, status, msg := loadPMContext(store, estimateID, strconv.FormatUint(uint64(body.ResidentialEntryID), 10))
	if ctx == nil {
		c.JSON(status, ErrorResponse{Error: msg, Code: "VALIDATION_ERROR"})
		return
	}

	var existing models.EstimateStageUser
	err := store.Where("stage_id = ? AND user_id = ? AND task_name = ?", ctx.Stage.ID, body.UserID, body.TaskName).First(&existing)
	if err == nil {
		existing.Name = body.UserName
		existing.TaskName = body.TaskName
		if err := store.Save(&existing); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update stage user", Code: "DATABASE_ERROR", Details: err.Error()})
			return
		}
		if body.TaskID != "" {
			var task models.EstimateProjectTask
			if err := store.Where("id = ? AND stage_id = ?", body.TaskID, ctx.Stage.ID).First(&task); err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Selected task was not found", Code: "VALIDATION_ERROR"})
				return
			}
			task.AssignedTo = appendTaskAssignee(task.AssignedTo, body.UserID, body.UserName)
			if err := store.Save(&task); err != nil {
				c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to assign user to task", Code: "DATABASE_ERROR", Details: err.Error()})
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{"message": "Stage user updated successfully", "user": existing})
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Database error", Code: "DATABASE_ERROR", Details: err.Error()})
		return
	}

	newUser := models.EstimateStageUser{
		ID:                 uuid.NewString(),
		StageID:            ctx.Stage.ID,
		ResidentialEntryID: ctx.EntryID,
		UserID:             body.UserID,
		Name:               body.UserName,
		TaskName:           body.TaskName,
	}
	if err := store.Create(&newUser); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to add stage user", Code: "DATABASE_ERROR", Details: err.Error()})
		return
	}
	if body.TaskID != "" {
		var task models.EstimateProjectTask
		if err := store.Where("id = ? AND stage_id = ?", body.TaskID, ctx.Stage.ID).First(&task); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Selected task was not found", Code: "VALIDATION_ERROR"})
			return
		}
		task.AssignedTo = appendTaskAssignee(task.AssignedTo, body.UserID, body.UserName)
		if err := store.Save(&task); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to assign user to task", Code: "DATABASE_ERROR", Details: err.Error()})
			return
		}
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Stage user added successfully", "user": newUser})
}

func appendTaskAssignee(existing, userID, name string) string {
	assignment := userID + "-" + name
	for _, current := range strings.Split(existing, ",") {
		if strings.TrimSpace(current) == assignment {
			return existing
		}
	}
	if strings.TrimSpace(existing) == "" {
		return assignment
	}
	return existing + "," + assignment
}

func (h *Handler) pmDeleteUser(c *gin.Context, store db.Database, estimateID string, body pmMutationRequest) {
	ctx, status, msg := loadPMContext(store, estimateID, strconv.FormatUint(uint64(body.ResidentialEntryID), 10))
	if ctx == nil {
		c.JSON(status, ErrorResponse{Error: msg, Code: "VALIDATION_ERROR"})
		return
	}

	var tasks []models.EstimateProjectTask
	if err := store.Where("stage_id = ? AND name = ? AND residential_entry_id = ?", ctx.Stage.ID, body.TaskName, ctx.EntryID).Find(&tasks); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to check task assignment", Code: "DATABASE_ERROR", Details: err.Error()})
		return
	}
	for _, task := range tasks {
		if isPrimaryTaskAssignee(task.AssignedTo, body.UserID) {
			c.JSON(http.StatusConflict, ErrorResponse{Error: "User is the primary assignee for this task. Reassign the task before deleting the user.", Code: "PRIMARY_USER_ASSIGNED_TO_TASK"})
			return
		}
	}

	if err := store.Delete(&models.EstimateStageUser{}, "stage_id = ? AND user_id = ? AND task_name = ?", ctx.Stage.ID, body.UserID, body.TaskName); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete stage user", Code: "DATABASE_ERROR", Details: err.Error()})
		return
	}

	for _, task := range tasks {
		updatedAssignedTo := removeTaskAssignee(task.AssignedTo, body.UserID)
		if updatedAssignedTo != task.AssignedTo {
			task.AssignedTo = updatedAssignedTo
			if err := store.Save(&task); err != nil {
				c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to remove user from task", Code: "DATABASE_ERROR", Details: err.Error()})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Stage user deleted successfully"})
}

func (h *Handler) pmAddNote(c *gin.Context, store db.Database, estimateID string, body pmMutationRequest) {
	ctx, status, msg := loadPMContext(store, estimateID, strconv.FormatUint(uint64(body.ResidentialEntryID), 10))
	if ctx == nil {
		c.JSON(status, ErrorResponse{Error: msg, Code: "VALIDATION_ERROR"})
		return
	}

	note := models.EstimateStageNote{
		StageID:            ctx.Stage.ID,
		ResidentialEntryID: ctx.EntryID,
		Text:               body.Note,
	}
	if err := store.Create(&note); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to add note", Code: "DATABASE_ERROR", Details: err.Error()})
		return
	}

	var notes []models.EstimateStageNote
	if err := store.Where("residential_entry_id = ?", ctx.EntryID).Find(&notes); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to reload notes", Code: "DATABASE_ERROR", Details: err.Error()})
		return
	}

	userID, userName := middleware.GetRequestUser(c)
	event := activity.ActivityEvent{
		EntityType: "estimate",
		EntityID:   ctx.Estimate.ID,
		Action:     "pm_note_added",
		UserID:     userID,
		UserName:   userName,
		Payload:    map[string]any{"lead_name": ctx.Estimate.Lead.JobName, "note_text": body.Note},
	}
	if err := activity.InsertActivityEvent(c.Request.Context(), store.GetUnderlying(), event); err != nil {
		log.Printf("activity event publish failed: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Note added successfully", "notes": notes})
}

func (h *Handler) pmDeleteNote(c *gin.Context, store db.Database, estimateID string, body pmMutationRequest) {
	ctx, status, msg := loadPMContext(store, estimateID, strconv.FormatUint(uint64(body.ResidentialEntryID), 10))
	if ctx == nil {
		c.JSON(status, ErrorResponse{Error: msg, Code: "VALIDATION_ERROR"})
		return
	}

	// Load the note first so the linked scratchpad can be removed too
	var deletedNote models.EstimateStageNote
	_ = store.Where("id = ? AND residential_entry_id = ?", body.NoteID, ctx.EntryID).First(&deletedNote)

	if err := store.Delete(&models.EstimateStageNote{}, "id = ? AND residential_entry_id = ?", body.NoteID, ctx.EntryID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete note", Code: "DATABASE_ERROR", Details: err.Error()})
		return
	}

	if deletedNote.ID != 0 {
		removeScratchpadLinks(store, "estimate_stage", deletedNote.StageID, deletedNote.Text, deletedNote.ID)
	}

	var notes []models.EstimateStageNote
	if err := store.Where("residential_entry_id = ?", ctx.EntryID).Find(&notes); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to reload notes", Code: "DATABASE_ERROR", Details: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Note deleted successfully", "notes": notes})
}

// migrateEstimateProjectDataToJobStage copies tasks and user assignments
// created via Estimate Project Management (EstimateProjectTask /
// EstimateStageUser) onto the Job tracker's own tables (StageTask /
// StageUser) for the given Template stage, the moment "Create a Job"
// turns that residential entry into a Job. The Stage row itself is
// shared (Project Management writes to the same Stage the Job tracker
// later reads), so only the task/user rows — which live in separate,
// intentionally-isolated tables — need to be carried over.
//
// Idempotent: if StageTask/StageUser rows already exist for this stage
// (e.g. because "Create a Job" was triggered more than once), nothing is
// copied again.
func migrateEstimateProjectDataToJobStage(tx db.Database, residentialEntryID uint, stage models.Stage) error {
	var existingTaskCount int64
	if err := tx.GetUnderlying().Model(&models.StageTask{}).Where("stage_id = ?", stage.ID).Count(&existingTaskCount).Error; err != nil {
		return err
	}
	var existingUserCount int64
	if err := tx.GetUnderlying().Model(&models.StageUser{}).Where("stage_id = ?", stage.ID).Count(&existingUserCount).Error; err != nil {
		return err
	}

	if existingTaskCount == 0 {
		var estimateTasks []models.EstimateProjectTask
		if err := tx.Where("residential_entry_id = ?", residentialEntryID).Find(&estimateTasks); err != nil {
			return err
		}
		for _, et := range estimateTasks {
			startTime, endTime, dueDate, startDate := parseEstimateTaskTimes(et)

			photosJSON, err := json.Marshal(et.Photos)
			if err != nil {
				return err
			}
			projectsJSON, err := json.Marshal(et.Projects)
			if err != nil {
				return err
			}

			newTask := models.StageTask{
				ID:                   uuid.NewString(),
				StageID:              stage.ID,
				Name:                 et.Name,
				DueDate:              dueDate,
				StartDate:            startDate,
				Duration:             et.Duration,
				Completed:            et.Completed,
				Cancelled:            et.Cancelled,
				IsRunning:            et.IsRunning,
				Elapsed:              et.Elapsed,
				StartTime:            startTime,
				LastResumedAt:        et.LastResumedAt,
				EndTime:              endTime,
				AssignedTo:           et.AssignedTo,
				NotesHistory:         models.NoteEntries(et.NotesHistory),
				Photos:               datatypes.JSON(photosJSON),
				Projects:             datatypes.JSON(projectsJSON),
				SourceEstimateTaskID: et.ID,
			}
			if err := tx.Create(&newTask); err != nil {
				return err
			}
		}
	}

	if existingUserCount == 0 {
		var estimateUsers []models.EstimateStageUser
		if err := tx.Where("residential_entry_id = ?", residentialEntryID).Find(&estimateUsers); err != nil {
			return err
		}
		for _, eu := range estimateUsers {
			newUser := models.StageUser{
				ID:       uuid.NewString(),
				StageID:  stage.ID,
				UserID:   eu.UserID,
				Name:     eu.Name,
				TaskName: eu.TaskName,
			}
			if err := tx.Create(&newUser); err != nil {
				return err
			}
		}
	}

	return nil
}

// parseEstimateTaskTimes converts EstimateProjectTask's string-based date
// fields (as sent by the frontend: "YYYY-MM-DD" for StartDate, RFC3339
// for StartTime/EndTime) into the time.Time / *time.Time fields StageTask
// expects. Falls back to "now" for StartDate if parsing fails, mirroring
// how CreateStageTask defaults missing/invalid values elsewhere.
func parseEstimateTaskTimes(et models.EstimateProjectTask) (startTime, endTime *time.Time, dueDate, startDate time.Time) {
	if t, err := time.Parse(time.RFC3339, et.StartTime); err == nil {
		tt := t.UTC()
		startTime = &tt
	}
	if t, err := time.Parse(time.RFC3339, et.EndTime); err == nil {
		tt := t.UTC()
		endTime = &tt
	}
	if t, err := time.Parse("2006-01-02", et.StartDate); err == nil {
		startDate = t
	} else {
		startDate = time.Now().UTC()
	}
	if endTime != nil {
		dueDate = *endTime
	} else {
		dueDate = startDate
	}
	return
}
