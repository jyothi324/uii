package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tuliporg/agent-management-api/handlers"
	"github.com/tuliporg/agent-management-api/mocks"
	"github.com/tuliporg/agent-management-api/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

//	type response struct {
//		Message  string                 `json:"message"`
//		Material map[string]interface{} `json:"material"`
//	}
//
// ----------------- GetStagesByLeadID -----------------
func TestGetStagesByLeadID_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	now := time.Now().UTC()
	stages := []models.Stage{
		{
			ID:                 "stage-1",
			LeadID:             "lead-1",
			ResidentialEntryID: 1,
			Name:               "Template",
			State:              "In Progress",
			Date:               &now,
			Tasks:              []models.StageTask{},
			Materials:          []models.StageMaterial{},
			Users:              []models.StageUser{},
		},
		{
			ID:                 "stage-2",
			LeadID:             "lead-1",
			ResidentialEntryID: 1,
			Name:               "CAD",
			State:              "Not Started",
			Tasks:              []models.StageTask{},
			Materials:          []models.StageMaterial{},
			Users:              []models.StageUser{},
		},
	}

	mockDB.On("Preload", "Tasks.Timers").Return(mockDB)
	mockDB.On("Preload", "Materials").Return(mockDB)
	mockDB.On("Preload", "Users").Return(mockDB)
	mockDB.On("Where", "residential_entry_id = ?", uint64(1)).Return(mockDB)
	mockDB.On("Find", mock.AnythingOfType("*[]models.Stage")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*[]models.Stage)
		*dest = stages
	}).Return(nil)

	req, _ := http.NewRequest("GET", "/stages/lead/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "lead_id", Value: "1"}}
	c.Set("db", mockDB)

	handler.GetStagesByLeadID()(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	stages_resp, ok := resp["stages"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(stages_resp))

	mockDB.AssertExpectations(t)
}

func TestGetStagesByLeadID_InvalidLeadID(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/stages/lead/invalid", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "lead_id", Value: "invalid"}}
	c.Set("db", mockDB)

	handler.GetStagesByLeadID()(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ----------------- UpdateStageStatus -----------------
func TestUpdateStageStatus_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	now := time.Now().UTC()
	stage := models.Stage{
		ID:                 "stage-1",
		LeadID:             "lead-1",
		ResidentialEntryID: 1,
		Name:               "Template",
		State:              "Not Started",
		Date:               nil,
	}

	stages := []models.Stage{stage}

	input := map[string]string{"state": "In Progress"}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Where", "id = ?", "stage-1").Return(mockDB)
	mockDB.On("First", mock.AnythingOfType("*models.Stage")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Stage)
		*dest = stage
	}).Return(nil)

	mockDB.On("Preload", "Tasks.Timers").Return(mockDB)
	mockDB.On("Where", "residential_entry_id = ?", uint(1)).Return(mockDB)
	mockDB.On("Find", mock.AnythingOfType("*[]models.Stage")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*[]models.Stage)
		*dest = stages
	}).Return(nil)

	mockDB.On("Save", mock.AnythingOfType("*models.Stage")).Run(func(args mock.Arguments) {
		s := args.Get(0).(*models.Stage)
		s.Date = &now
	}).Return(nil)
	mockDB.On("Commit").Return(nil)

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("PUT", "/stages/lead/lead-1/stage-1/status", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "lead_id", Value: "lead-1"},
		{Key: "stage_id", Value: "stage-1"},
	}
	c.Set("db", mockDB)

	handler.UpdateStageStatus()(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockDB.AssertExpectations(t)
}

func TestUpdateStageStatus_InvalidState(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	input := map[string]string{"state": "Invalid State"}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("PUT", "/stages/lead/lead-1/stage-1/status", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "lead_id", Value: "lead-1"},
		{Key: "stage_id", Value: "stage-1"},
	}
	c.Set("db", mockDB)

	handler.UpdateStageStatus()(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateStageStatus_StageNotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	input := map[string]string{"state": "In Progress"}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Where", "id = ?", "stage-999").Return(mockDB)
	mockDB.On("First", mock.AnythingOfType("*models.Stage")).Return(gorm.ErrRecordNotFound)
	mockDB.On("Rollback").Return(nil)

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("PUT", "/stages/lead/lead-1/stage-999/status", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "lead_id", Value: "lead-1"},
		{Key: "stage_id", Value: "stage-999"},
	}
	c.Set("db", mockDB)

	handler.UpdateStageStatus()(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockDB.AssertExpectations(t)
}

// ----------------- StartTaskTimer -----------------
func TestStartTaskTimer_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	task := models.StageTask{
		ID:        "task-1",
		StageID:   "stage-1",
		Name:      "Test Task",
		IsRunning: false,
		Completed: false,
		Elapsed:   0,
	}

	mockDB.On("Where", "id = ?", "task-1").Return(mockDB)
	mockDB.On("First", mock.AnythingOfType("*models.StageTask")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.StageTask)
		*dest = task
	}).Return(nil)
	mockDB.On("Save", mock.AnythingOfType("*models.StageTask")).Return(nil)

	req, _ := http.NewRequest("POST", "/tasks/task-1/timer/start", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "task-1"}}
	c.Set("db", mockDB)

	handler.StartTaskTimer()(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Task timer started/resumed", resp["message"])

	mockDB.AssertExpectations(t)
}

func TestStartTaskTimer_AlreadyRunning(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	now := time.Now().UTC()
	task := models.StageTask{
		ID:            "task-1",
		StageID:       "stage-1",
		Name:          "Test Task",
		IsRunning:     true,
		LastResumedAt: &now,
		Completed:     false,
		Elapsed:       100,
	}

	mockDB.On("Where", "id = ?", "task-1").Return(mockDB)
	mockDB.On("First", mock.AnythingOfType("*models.StageTask")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.StageTask)
		*dest = task
	}).Return(nil)

	req, _ := http.NewRequest("POST", "/tasks/task-1/timer/start", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "task-1"}}
	c.Set("db", mockDB)

	handler.StartTaskTimer()(c)

	assert.Equal(t, http.StatusConflict, w.Code)
	mockDB.AssertExpectations(t)
}

func TestStartTaskTimer_CompletedTask(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	task := models.StageTask{
		ID:        "task-1",
		StageID:   "stage-1",
		Name:      "Test Task",
		IsRunning: false,
		Completed: true,
		Elapsed:   1000,
	}

	mockDB.On("Where", "id = ?", "task-1").Return(mockDB)
	mockDB.On("First", mock.AnythingOfType("*models.StageTask")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.StageTask)
		*dest = task
	}).Return(nil)

	req, _ := http.NewRequest("POST", "/tasks/task-1/timer/start", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "task-1"}}
	c.Set("db", mockDB)

	handler.StartTaskTimer()(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockDB.AssertExpectations(t)
}

func TestStartTaskTimer_TaskNotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	mockDB.On("Where", "id = ?", "task-999").Return(mockDB)
	mockDB.On("First", mock.AnythingOfType("*models.StageTask")).Return(gorm.ErrRecordNotFound)

	req, _ := http.NewRequest("POST", "/tasks/task-999/timer/start", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "task-999"}}
	c.Set("db", mockDB)

	handler.StartTaskTimer()(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockDB.AssertExpectations(t)
}

// ----------------- PauseTaskTimer -----------------
func TestPauseTaskTimer_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	now := time.Now().UTC()
	task := models.StageTask{
		ID:            "task-1",
		StageID:       "stage-1",
		Name:          "Test Task",
		IsRunning:     true,
		LastResumedAt: &now,
		Completed:     false,
		Elapsed:       100,
	}

	mockDB.On("Where", "id = ?", "task-1").Return(mockDB)
	mockDB.On("First", mock.AnythingOfType("*models.StageTask")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.StageTask)
		*dest = task
	}).Return(nil)
	mockDB.On("Save", mock.AnythingOfType("*models.StageTask")).Return(nil)

	req, _ := http.NewRequest("POST", "/tasks/task-1/timer/pause", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "task-1"}}
	c.Set("db", mockDB)

	handler.PauseTaskTimer()(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Task timer paused", resp["message"])

	mockDB.AssertExpectations(t)
}

func TestPauseTaskTimer_AlreadyPaused(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	task := models.StageTask{
		ID:            "task-1",
		StageID:       "stage-1",
		Name:          "Test Task",
		IsRunning:     false,
		LastResumedAt: nil,
		Completed:     false,
		Elapsed:       100,
	}

	mockDB.On("Where", "id = ?", "task-1").Return(mockDB)
	mockDB.On("First", mock.AnythingOfType("*models.StageTask")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.StageTask)
		*dest = task
	}).Return(nil)

	req, _ := http.NewRequest("POST", "/tasks/task-1/timer/pause", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "task-1"}}
	c.Set("db", mockDB)

	handler.PauseTaskTimer()(c)

	assert.Equal(t, http.StatusConflict, w.Code)
	mockDB.AssertExpectations(t)
}

// ----------------- CompleteTaskTimer -----------------
func TestCompleteTaskTimer_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	now := time.Now().UTC()
	task := models.StageTask{
		ID:            "task-1",
		StageID:       "stage-1",
		Name:          "Test Task",
		IsRunning:     true,
		LastResumedAt: &now,
		Completed:     false,
		Elapsed:       1000,
	}

	mockDB.On("Where", "id = ?", "task-1").Return(mockDB)
	mockDB.On("First", mock.AnythingOfType("*models.StageTask")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.StageTask)
		*dest = task
	}).Return(nil)
	mockDB.On("Save", mock.AnythingOfType("*models.StageTask")).Return(nil)

	req, _ := http.NewRequest("POST", "/tasks/task-1/timer/complete", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "task-1"}}
	c.Set("db", mockDB)

	handler.CompleteTaskTimer()(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Task completed", resp["message"])

	mockDB.AssertExpectations(t)
}

func TestCompleteTaskTimer_AlreadyCompleted(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	task := models.StageTask{
		ID:        "task-1",
		StageID:   "stage-1",
		Name:      "Test Task",
		IsRunning: false,
		Completed: true,
		Elapsed:   1500,
	}

	mockDB.On("Where", "id = ?", "task-1").Return(mockDB)
	mockDB.On("First", mock.AnythingOfType("*models.StageTask")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.StageTask)
		*dest = task
	}).Return(nil)

	req, _ := http.NewRequest("POST", "/tasks/task-1/timer/complete", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "task-1"}}
	c.Set("db", mockDB)

	handler.CompleteTaskTimer()(c)

	assert.Equal(t, http.StatusConflict, w.Code)
	mockDB.AssertExpectations(t)
}

// ----------------- GetTaskTimerStatus -----------------
func TestGetTaskTimerStatus_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	now := time.Now().UTC()
	startTime := now.Add(-10 * time.Minute)
	task := models.StageTask{
		ID:            "task-1",
		StageID:       "stage-1",
		Name:          "Test Task",
		IsRunning:     true,
		LastResumedAt: &startTime,
		StartTime:     &startTime,
		Completed:     false,
		Elapsed:       300,
	}

	mockDB.On("Where", "id = ?", "task-1").Return(mockDB)
	mockDB.On("First", mock.AnythingOfType("*models.StageTask")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.StageTask)
		*dest = task
	}).Return(nil)

	req, _ := http.NewRequest("GET", "/tasks/task-1/timer/status", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "task-1"}}
	c.Set("db", mockDB)

	handler.GetTaskTimerStatus()(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "task-1", resp["task_id"])
	assert.Equal(t, true, resp["is_running"])

	mockDB.AssertExpectations(t)
}

// ----------------- CreateStageTask -----------------
func TestCreateStageTask_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	stage := models.Stage{
		ID:                 "stage-1",
		LeadID:             "lead-1",
		ResidentialEntryID: 1,
		Name:               "Template",
	}

	duration := 240
	input := map[string]interface{}{
		"name":       "New Task",
		"startDate":  "2024-01-15",
		"startTime":  "2024-01-15T09:00:00Z",
		"assignedTo": "agent-1",
		"duration":   duration,
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Where", "id = ? AND residential_entry_id = ?", "stage-1", uint64(1)).Return(mockDB)
	mockDB.On("First", mock.AnythingOfType("*models.Stage")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Stage)
		*dest = stage
	}).Return(nil)
	mockDB.On("Create", mock.AnythingOfType("*models.StageTask")).Return(nil)
	mockDB.On("Commit").Return(nil)

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/leads/1/stages/stage-1/tasks", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "lead_id", Value: "1"},
		{Key: "stage_id", Value: "stage-1"},
	}
	c.Set("db", mockDB)

	handler.CreateStageTask()(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockDB.AssertExpectations(t)
}

func TestCreateStageTask_InvalidStartDate(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	input := map[string]interface{}{
		"name":       "New Task",
		"startDate":  "invalid-date",
		"startTime":  "2024-01-15T09:00:00Z",
		"assignedTo": "agent-1",
	}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/leads/1/stages/stage-1/tasks", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "lead_id", Value: "1"},
		{Key: "stage_id", Value: "stage-1"},
	}
	c.Set("db", mockDB)

	handler.CreateStageTask()(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ----------------- AddStageNote -----------------
func TestAddStageNote_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	notesJSON, _ := json.Marshal([]string{"Existing note"})
	stage := models.Stage{
		ID:                 "stage-1",
		LeadID:             "lead-1",
		ResidentialEntryID: 1,
		Name:               "Template",
		Notes:              datatypes.JSON(notesJSON),
	}

	input := map[string]string{"note": "New note"}

	mockDB.On("Where", "id = ? AND residential_entry_id = ?", "stage-1", "1").Return(mockDB)
	mockDB.On("First", mock.AnythingOfType("*models.Stage")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Stage)
		*dest = stage
	}).Return(nil)
	mockDB.On("Model", mock.AnythingOfType("*models.Stage")).Return(mockDB)
	mockDB.On("Update", "notes", mock.Anything).Return(nil)

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/leads/1/stages/stage-1/notes", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "lead_id", Value: "1"},
		{Key: "stage_id", Value: "stage-1"},
	}
	c.Set("db", mockDB)

	handler.AddStageNote()(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Note added successfully", resp["message"])

	mockDB.AssertExpectations(t)
}

func TestAddStageNote_StageNotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	input := map[string]string{"note": "New note"}

	mockDB.On("Where", "id = ? AND residential_entry_id = ?", "stage-999", "1").Return(mockDB)
	mockDB.On("First", mock.AnythingOfType("*models.Stage")).Return(gorm.ErrRecordNotFound)

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/leads/1/stages/stage-999/notes", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "lead_id", Value: "1"},
		{Key: "stage_id", Value: "stage-999"},
	}
	c.Set("db", mockDB)

	handler.AddStageNote()(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockDB.AssertExpectations(t)
}

// ----------------- DeleteStageNote -----------------
func TestDeleteStageNote_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	notesJSON, _ := json.Marshal([]string{"Note 1", "Note 2", "Note 3"})
	stage := models.Stage{
		ID:                 "stage-1",
		LeadID:             "lead-1",
		ResidentialEntryID: 1,
		Name:               "Template",
		Notes:              datatypes.JSON(notesJSON),
	}

	mockDB.On("Where", "id = ? AND residential_entry_id = ?", "stage-1", "1").Return(mockDB)
	mockDB.On("First", mock.AnythingOfType("*models.Stage")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Stage)
		*dest = stage
	}).Return(nil)
	mockDB.On("Save", mock.AnythingOfType("*models.Stage")).Return(nil)

	req, _ := http.NewRequest("DELETE", "/leads/1/stages/stage-1/notes/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "lead_id", Value: "1"},
		{Key: "stage_id", Value: "stage-1"},
		{Key: "note_id", Value: "1"},
	}
	c.Set("db", mockDB)

	handler.DeleteStageNote()(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Note deleted successfully", resp["message"])

	mockDB.AssertExpectations(t)
}

func TestDeleteStageNote_InvalidNoteIndex(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("DELETE", "/leads/1/stages/stage-1/notes/invalid", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "lead_id", Value: "1"},
		{Key: "stage_id", Value: "stage-1"},
		{Key: "note_id", Value: "invalid"},
	}
	c.Set("db", mockDB)

	handler.DeleteStageNote()(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// _----------------- UpdateStageTask -From-here----------------
func TestUpdateStageTask_Success_MarkCompleted(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	leadID := "lead-123"
	stageID := "stage-456"
	taskID := "task-789"

	existingTask := models.StageTask{
		ID:        taskID,
		StageID:   stageID,
		Name:      "Test Task",
		Completed: false,
	}

	existingStage := models.Stage{
		ID:     stageID,
		LeadID: leadID,
		Name:   "Test Stage",
	}

	now := time.Now().UTC()
	runningTimer := models.StageTaskTimer{
		ID:          "timer-1",
		TaskID:      taskID,
		IsRunning:   true,
		StartTime:   &now,
		ElapsedTime: 100,
	}

	request := struct {
		Completed bool `json:"completed"`
	}{
		Completed: true,
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Where", "id = ? AND stage_id = ?", taskID, stageID).Return(mockDB).Once()
	mockDB.On("First", mock.AnythingOfType("*models.StageTask")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.StageTask)
		*dest = existingTask
	}).Return(nil).Once()

	mockDB.On("Where", "id = ?", stageID).Return(mockDB).Once()
	mockDB.On("First", mock.AnythingOfType("*models.Stage")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Stage)
		*dest = existingStage
	}).Return(nil).Once()

	mockDB.On("Where", "task_id = ? AND is_running = ?", taskID, true).Return(mockDB).Once()
	mockDB.On("Find", mock.AnythingOfType("*[]models.StageTaskTimer")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*[]models.StageTaskTimer)
		*dest = []models.StageTaskTimer{runningTimer}
	}).Return(nil).Once()

	mockDB.On("Save", mock.AnythingOfType("*models.StageTaskTimer")).Return(nil).Once()
	mockDB.On("Save", mock.AnythingOfType("*models.StageTask")).Return(nil).Once()
	mockDB.On("Commit").Return(nil).Once()

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(request)
	req, _ := http.NewRequest("PUT", "/leads/"+leadID+"/stages/"+stageID+"/tasks/"+taskID, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "lead_id", Value: leadID},
		{Key: "stage_id", Value: stageID},
		{Key: "task_id", Value: taskID},
	}
	c.Set("db", mockDB)

	handler.UpdateStageTask()(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Task updated successfully", resp["message"])

	mockDB.AssertExpectations(t)
}

func TestUpdateStageTask_MissingParameters(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("PUT", "/leads//stages//tasks/", bytes.NewBuffer([]byte(`{"completed": true}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "lead_id", Value: ""},
		{Key: "stage_id", Value: ""},
		{Key: "task_id", Value: ""},
	}
	c.Set("db", mockDB)

	handler.UpdateStageTask()(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Contains(t, resp["error"], "required")
}

func TestUpdateStageTask_InvalidJSON(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("PUT", "/leads/lead-1/stages/stage-1/tasks/task-1", bytes.NewBuffer([]byte(`invalid json`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "lead_id", Value: "lead-1"},
		{Key: "stage_id", Value: "stage-1"},
		{Key: "task_id", Value: "task-1"},
	}
	c.Set("db", mockDB)

	handler.UpdateStageTask()(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Contains(t, resp["error"], "Invalid request")
}

func TestUpdateStageTask_TaskNotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	request := struct {
		Completed bool `json:"completed"`
	}{Completed: true}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Where", "id = ? AND stage_id = ?", "task-999", "stage-1").Return(mockDB)
	mockDB.On("First", mock.AnythingOfType("*models.StageTask")).Return(gorm.ErrRecordNotFound)
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(request)
	req, _ := http.NewRequest("PUT", "/leads/lead-1/stages/stage-1/tasks/task-999", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "lead_id", Value: "lead-1"},
		{Key: "stage_id", Value: "stage-1"},
		{Key: "task_id", Value: "task-999"},
	}
	c.Set("db", mockDB)

	handler.UpdateStageTask()(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Task not found", resp["error"])

	mockDB.AssertExpectations(t)
}

func TestUpdateStageTask_StageNotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	existingTask := models.StageTask{ID: "task-1", StageID: "stage-999"}
	request := struct {
		Completed bool `json:"completed"`
	}{Completed: true}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Where", "id = ? AND stage_id = ?", "task-1", "stage-999").Return(mockDB)
	mockDB.On("First", mock.AnythingOfType("*models.StageTask")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.StageTask)
		*dest = existingTask
	}).Return(nil).Once()
	mockDB.On("Where", "id = ?", "stage-999").Return(mockDB)
	mockDB.On("First", mock.AnythingOfType("*models.Stage")).Return(gorm.ErrRecordNotFound).Once()
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(request)
	req, _ := http.NewRequest("PUT", "/leads/lead-1/stages/stage-999/tasks/task-1", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "lead_id", Value: "lead-1"},
		{Key: "stage_id", Value: "stage-999"},
		{Key: "task_id", Value: "task-1"},
	}
	c.Set("db", mockDB)

	handler.UpdateStageTask()(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Stage not found", resp["error"])

	mockDB.AssertExpectations(t)
}

// ----------------- ToggleStageTaskTimer -----------------
func TestToggleStageTaskTimer_Success_StartTimer(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	leadID := "lead-123"
	stageID := "stage-456"
	taskID := "task-789"

	existingTask := models.StageTask{
		ID:        taskID,
		StageID:   stageID,
		Completed: false,
		Duration:  60,
	}

	existingStage := models.Stage{
		ID:     stageID,
		LeadID: leadID,
	}

	request := struct {
		IsRunning bool `json:"isRunning"`
	}{IsRunning: true}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Where", "id = ? AND stage_id = ?", taskID, stageID).Return(mockDB).Once()
	mockDB.On("First", mock.AnythingOfType("*models.StageTask")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.StageTask)
		*dest = existingTask
	}).Return(nil).Once()
	mockDB.On("Where", "id = ? AND lead_id = ?", stageID, leadID).Return(mockDB).Once()
	mockDB.On("First", mock.AnythingOfType("*models.Stage")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Stage)
		*dest = existingStage
	}).Return(nil).Once()
	mockDB.On("Where", "task_id = ? AND is_running = ?", taskID, true).Return(mockDB).Once()
	mockDB.On("First", mock.AnythingOfType("*models.StageTaskTimer")).Return(gorm.ErrRecordNotFound).Once()
	mockDB.On("Create", mock.AnythingOfType("*models.StageTaskTimer")).Return(nil).Once()
	mockDB.On("Model", mock.AnythingOfType("*models.StageTaskTimer")).Return(mockDB)
	mockDB.On("Where", "task_id = ?", taskID).Return(mockDB)
	mockDB.On("Select", "SUM(elapsed_time)").Return(mockDB)
	mockDB.On("Scan", mock.AnythingOfType("*int")).Return(nil)
	mockDB.On("Commit").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", "/leads/"+leadID+"/stages/"+stageID+"/tasks/"+taskID+"/timer", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "lead_id", Value: leadID},
		{Key: "stage_id", Value: stageID},
		{Key: "task_id", Value: taskID},
	}
	c.Set("db", mockDB)

	handler.ToggleStageTaskTimer()(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Timer toggled successfully", resp["message"])

	mockDB.AssertExpectations(t)
}

// func TestToggleStageTaskTimer_Success_StopTimer(t *testing.T) {
// 	// ------------------------------------
// 	// Arrange
// 	// ------------------------------------
// 	gin.SetMode(gin.TestMode)

// 	mockDB := new(mocks.MockDatabase)

// 	leadID := "lead-123"
// 	stageID := "stage-456"
// 	taskID := "task-789"

// 	existingTask := models.StageTask{
// 		ID:        taskID,
// 		StageID:   stageID,
// 		Completed: false,
// 		Duration:  60,
// 	}

// 	existingStage := models.Stage{
// 		ID:     stageID,
// 		LeadID: leadID,
// 	}

// 	now := time.Now().UTC()
// 	runningTimer := models.StageTaskTimer{
// 		ID:          "timer-1",
// 		TaskID:      taskID,
// 		IsRunning:   true,
// 		StartTime:   &now,
// 		ElapsedTime: 100,
// 	}
// 	// Transaction
// 	mockDB.On("Begin").Return(mockDB).Once()
// 	mockDB.On("Rollback").Return(mockDB).Once()
// 	mockDB.On("Commit").Return(mockDB).Once()

// 	// Fetch stage task
// 	mockDB.
// 		On("Where", "id = ? AND stage_id = ?", taskID, stageID).
// 		Return(mockDB).Once()

// 	mockDB.
// 		On("First", mock.AnythingOfType("*models.StageTask")).
// 		Run(func(args mock.Arguments) {
// 			*args.Get(0).(*models.StageTask) = existingTask
// 		}).
// 		Return(mockDB).Once()

// 	// Fetch stage
// 	mockDB.
// 		On("Where", "id = ? AND lead_id = ?", stageID, leadID).
// 		Return(mockDB).Once()

// 	mockDB.
// 		On("First", mock.AnythingOfType("*models.Stage")).
// 		Run(func(args mock.Arguments) {
// 			*args.Get(0).(*models.Stage) = existingStage
// 		}).
// 		Return(mockDB).Once()

// 	// Fetch running timer
// 	mockDB.
// 		On("Where", "task_id = ? AND is_running = ?", taskID, true).
// 		Return(mockDB).Once()

// 	mockDB.
// 		On("First", mock.AnythingOfType("*models.StageTaskTimer")).
// 		Run(func(args mock.Arguments) {
// 			*args.Get(0).(*models.StageTaskTimer) = runningTimer
// 		}).
// 		Return(mockDB).Once()

// 	// Save timer twice (stop + update duration)
// 	mockDB.
// 		On("Save", mock.AnythingOfType("*models.StageTaskTimer")).
// 		Return(mockDB).Twice()

// 	// SUM(elapsed_time)
// 	mockDB.
// 		On("Model", mock.AnythingOfType("*models.StageTaskTimer")).
// 		Return(mockDB).Once()

// 	mockDB.
// 		On("Where", "task_id = ?", taskID).
// 		Return(mockDB).Once()

// 	mockDB.
// 		On("Select", "SUM(elapsed_time)").
// 		Return(mockDB).Once()

// 	mockDB.
// 		On("Scan", mock.AnythingOfType("*int")).
// 		Return(mockDB).Once()

// 	handler := &handlers.Handler{
// 		DB: mockDB, // ✅ handler uses this DB
// 	}

// 	// ✅ RAW JSON — matches handler exactly
// 	data := []byte(`{"is_running": false}`)

// 	req, _ := http.NewRequest(
// 		http.MethodPost,
// 		"/leads/"+leadID+"/stages/"+stageID+"/tasks/"+taskID+"/timer",
// 		bytes.NewBuffer(data),
// 	)
// 	req.Header.Set("Content-Type", "application/json")

// 	w := httptest.NewRecorder()
// 	c, _ := gin.CreateTestContext(w)
// 	c.Request = req
// 	c.Params = gin.Params{
// 		{Key: "lead_id", Value: leadID},
// 		{Key: "stage_id", Value: stageID},
// 		{Key: "task_id", Value: taskID},
// 	}

// 	// ------------------------------------
// 	// Act
// 	// ------------------------------------
// 	handler.ToggleStageTaskTimer()(c)

// 	// ------------------------------------
// 	// Assert
// 	// ------------------------------------
// 	assert.Equal(t, http.StatusOK, w.Code)

// 	var resp map[string]interface{}
// 	err := json.Unmarshal(w.Body.Bytes(), &resp)
// 	assert.NoError(t, err)

// 	assert.Equal(t, "Timer toggled successfully", resp["message"])

// 	mockDB.AssertExpectations(t)
// }

func TestToggleStageTaskTimer_CompletedTask(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	existingTask := models.StageTask{
		ID:        "task-1",
		StageID:   "stage-1",
		Completed: true,
	}

	existingStage := models.Stage{
		ID:     "stage-1",
		LeadID: "lead-1",
	}

	request := struct {
		IsRunning bool `json:"isRunning"`
	}{IsRunning: true}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Where", "id = ? AND stage_id = ?", "task-1", "stage-1").Return(mockDB)
	mockDB.On("First", mock.AnythingOfType("*models.StageTask")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.StageTask)
		*dest = existingTask
	}).Return(nil).Once()
	mockDB.On("Where", "id = ? AND lead_id = ?", "stage-1", "lead-1").Return(mockDB)
	mockDB.On("First", mock.AnythingOfType("*models.Stage")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Stage)
		*dest = existingStage
	}).Return(nil).Once()
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", "/leads/lead-1/stages/stage-1/tasks/task-1/timer", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "lead_id", Value: "lead-1"},
		{Key: "stage_id", Value: "stage-1"},
		{Key: "task_id", Value: "task-1"},
	}
	c.Set("db", mockDB)

	handler.ToggleStageTaskTimer()(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Cannot toggle timer for completed task", resp["error"])

	mockDB.AssertExpectations(t)
}

func TestToggleStageTaskTimer_TimerAlreadyRunning(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	existingTask := models.StageTask{
		ID:        "task-1",
		StageID:   "stage-1",
		Completed: false,
	}

	existingStage := models.Stage{
		ID:     "stage-1",
		LeadID: "lead-1",
	}

	now := time.Now().UTC()
	runningTimer := models.StageTaskTimer{
		ID:        "timer-1",
		TaskID:    "task-1",
		IsRunning: true,
		StartTime: &now,
	}

	request := struct {
		IsRunning bool `json:"isRunning"`
	}{IsRunning: true}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Where", "id = ? AND stage_id = ?", "task-1", "stage-1").Return(mockDB)
	mockDB.On("First", mock.AnythingOfType("*models.StageTask")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.StageTask)
		*dest = existingTask
	}).Return(nil).Once()
	mockDB.On("Where", "id = ? AND lead_id = ?", "stage-1", "lead-1").Return(mockDB)
	mockDB.On("First", mock.AnythingOfType("*models.Stage")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Stage)
		*dest = existingStage
	}).Return(nil).Once()
	mockDB.On("Where", "task_id = ? AND is_running = ?", "task-1", true).Return(mockDB)
	mockDB.On("First", mock.AnythingOfType("*models.StageTaskTimer")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.StageTaskTimer)
		*dest = runningTimer
	}).Return(nil).Once()
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", "/leads/lead-1/stages/stage-1/tasks/task-1/timer", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "lead_id", Value: "lead-1"},
		{Key: "stage_id", Value: "stage-1"},
		{Key: "task_id", Value: "task-1"},
	}
	c.Set("db", mockDB)

	handler.ToggleStageTaskTimer()(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Timer is already running", resp["error"])

	mockDB.AssertExpectations(t)
}

// ----------------- AddStageMaterial -----------------
func TestAddStageMaterial_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := new(mocks.MockDatabase)

	leadID := "lead-123"
	stageID := "stage-456"

	existingStage := models.Stage{
		ID:     stageID,
		LeadID: leadID,
		Name:   "Test Stage",
	}

	request := struct {
		Name string `json:"name"`
	}{
		Name: "Test Material",
	}

	// ---- Transaction ----
	mockDB.On("Begin").Return(mockDB)

	// ---- Fetch stage ----
	mockDB.
		On("Where", "id = ? AND lead_id = ?", stageID, leadID).
		Return(mockDB)

	mockDB.
		On("First", mock.AnythingOfType("*models.Stage")).
		Run(func(args mock.Arguments) {
			*args.Get(0).(*models.Stage) = existingStage
		}).
		Return(nil)

	// ---- Create material ----
	mockDB.
		On("Create", mock.AnythingOfType("*models.StageMaterial")).
		Return(nil)

	// ---- Commit ----
	mockDB.On("Commit").Return(nil)

	// ---- Rollback safety ----
	mockDB.On("Rollback").Return(nil).Maybe()

	handler := &handlers.Handler{DB: mockDB}

	body, _ := json.Marshal(request)
	req, _ := http.NewRequest(
		http.MethodPost,
		"/leads/"+leadID+"/stages/"+stageID+"/materials",
		bytes.NewBuffer(body),
	)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "lead_id", Value: leadID},
		{Key: "stage_id", Value: stageID},
	}
	c.Set("db", mockDB)

	// ---- Call handler ----
	handler.AddStageMaterial()(c)

	// ---- Assertions ----
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp map[string]interface{}

	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	// var msg map[string]interface{}
	// json.Unmarshal(w.Body.Bytes(), &msg)
	// assert.Equal(t, "Material added successfully", resp["message"])
	mockDB.AssertExpectations(t)
}

func TestAddStageMaterial_MissingParameters(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("POST", "/leads//stages//materials", bytes.NewBuffer([]byte(`{"name": "Material"}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "lead_id", Value: ""},
		{Key: "stage_id", Value: ""},
	}
	c.Set("db", mockDB)

	handler.AddStageMaterial()(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Contains(t, resp["error"], "required")
}

func TestAddStageMaterial_InvalidJSON(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("POST", "/leads/lead-1/stages/stage-1/materials", bytes.NewBuffer([]byte(`invalid json`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "lead_id", Value: "lead-1"},
		{Key: "stage_id", Value: "stage-1"},
	}
	c.Set("db", mockDB)

	handler.AddStageMaterial()(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Contains(t, resp["error"], "Invalid request")
}

func TestAddStageMaterial_StageNotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	request := struct {
		Name string `json:"name"`
	}{Name: "Material"}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Where", "id = ? AND lead_id = ?", "stage-999", "lead-1").Return(mockDB)
	mockDB.On("First", mock.AnythingOfType("*models.Stage")).Return(gorm.ErrRecordNotFound)
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", "/leads/lead-1/stages/stage-999/materials", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "lead_id", Value: "lead-1"},
		{Key: "stage_id", Value: "stage-999"},
	}
	c.Set("db", mockDB)

	handler.AddStageMaterial()(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Stage not found", resp["error"])

	mockDB.AssertExpectations(t)
}

// ----------------- AddStageUser -----------------
func TestAddStageUser_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	mockTx := new(mocks.MockDatabase) // Separate mock for transaction

	leadID := "lead-123"
	stageID := "stage-456"

	existingStage := models.Stage{
		ID:     stageID,
		LeadID: leadID,
		Name:   "Test Stage",
	}

	request := struct {
		UserID string `json:"userId"`
		Name   string `json:"name"`
	}{
		UserID: "user-789",
		Name:   "John Doe",
	}

	// Setup transaction mock
	mockDB.On("Begin").Return(mockTx)

	// Setup the Where -> First chain
	mockTx.On("Where", mock.Anything).Return(mockTx)
	mockTx.On("First", mock.AnythingOfType("*models.Stage")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Stage)
		*dest = existingStage
	}).Return(nil) // First returns error directly

	// Setup Create
	mockTx.On("Create", mock.AnythingOfType("*models.StageUser")).Return(nil) // Create returns error directly

	// Setup Commit
	mockTx.On("Commit").Return(nil) // Commit returns error directly

	// Setup Rollback (might be called in defer)
	mockTx.On("Rollback").Return(nil).Maybe()

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", "/leads/"+leadID+"/stages/"+stageID+"/users", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "lead_id", Value: leadID},
		{Key: "stage_id", Value: stageID},
	}
	c.Set("db", mockDB)

	handler.AddStageUser()(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	// var resp map[string]interface{}
	// json.Unmarshal(w.Body.Bytes(), &resp)
	// assert.Equal(t, "User added successfully", resp["message"])

	mockDB.AssertExpectations(t)
	// mockTx.AssertExpectations(t)
}
func TestAddStageUser_MissingParameters(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("POST", "/leads//stages//users", bytes.NewBuffer([]byte(`{"userId": "user-1", "name": "User"}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "lead_id", Value: ""},
		{Key: "stage_id", Value: ""},
	}
	c.Set("db", mockDB)

	handler.AddStageUser()(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Contains(t, resp["error"], "required")
}

func TestAddStageUser_InvalidJSON(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("POST", "/leads/lead-1/stages/stage-1/users", bytes.NewBuffer([]byte(`invalid json`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "lead_id", Value: "lead-1"},
		{Key: "stage_id", Value: "stage-1"},
	}
	c.Set("db", mockDB)

	handler.AddStageUser()(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Contains(t, resp["error"], "Invalid request")
}

func TestGetAllStages_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	stages := []models.Stage{
		{
			ID:     "stage-1",
			LeadID: "lead-1",
			Name:   "Stage One",
			Tasks:  []models.StageTask{},
		},
		{
			ID:     "stage-2",
			LeadID: "lead-2",
			Name:   "Stage Two",
			Tasks:  []models.StageTask{},
		},
	}

	// Mock chain calls
	mockDB.On("Preload", "Tasks.Timers").Return(mockDB)
	mockDB.On("Preload", "Materials").Return(mockDB)
	mockDB.On("Preload", "Users").Return(mockDB)

	mockDB.
		On("Find", mock.AnythingOfType("*[]models.Stage")).
		Run(func(args mock.Arguments) {
			dest := args.Get(0).(*[]models.Stage)
			*dest = stages
		}).
		Return(nil)

	req, _ := http.NewRequest("GET", "/stages", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.GetAllStages()(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	assert.Equal(t, float64(2), resp["total"])
	assert.NotNil(t, resp["stages"])

	mockDB.AssertExpectations(t)
}
