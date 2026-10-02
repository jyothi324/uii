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
	"github.com/tuliporg/agent-management-api/handlers"
	"github.com/tuliporg/agent-management-api/mocks"
	"github.com/tuliporg/agent-management-api/models"
	"gorm.io/gorm"
)

func TestCreateTask_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	task := models.Task{
		Title:       "Test Task",
		Description: "Test Description",
		DueDate:     "2025-12-31",
		Priority:    "high",
		Status:      "pending",
		TaskType:    "general",
	}

	body, _ := json.Marshal(task)

	// Mock DB calls
	mockDB.On("Begin").Return(mockDB).Once()

	mockDB.
		On("Create", mock.AnythingOfType("*models.Task")).
		Run(func(args mock.Arguments) {
			t := args.Get(0).(*models.Task)
			t.ID = 1
		}).
		Return(nil).
		Once()

	mockDB.On("Commit").Return(nil).Once()

	// HTTP Setup
	req, _ := http.NewRequest(
		http.MethodPost,
		"/tasks",
		bytes.NewBuffer(body),
	)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	// Call Handler
	handler.CreateTask(c)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "discount creation successfully", resp["message"])
	assert.NotNil(t, resp["data"])

	mockDB.AssertExpectations(t)
}

func TestCreateTask_InvalidInput(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	// Invalid JSON
	body := []byte(`{"title": }`)

	req, _ := http.NewRequest(
		http.MethodPost,
		"/tasks",
		bytes.NewBuffer(body),
	)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateTask(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid input", resp["error"])
}

func TestCreateTask_DatabaseError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	task := models.Task{
		Title:    "Test Task",
		Priority: "high",
		Status:   "pending",
		TaskType: "general",
	}

	body, _ := json.Marshal(task)

	mockDB.On("Begin").Return(mockDB).Once()
	mockDB.
		On("Create", mock.AnythingOfType("*models.Task")).
		Return(gorm.ErrInvalidData).
		Once()
	mockDB.On("Rollback").Return(nil).Once()

	req, _ := http.NewRequest(
		http.MethodPost,
		"/tasks",
		bytes.NewBuffer(body),
	)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateTask(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	mockDB.AssertExpectations(t)
}

func TestGetTask_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	taskID := "1"

	mockDB.
		On("First", mock.AnythingOfType("*models.Task"), "id = ?", taskID).
		Run(func(args mock.Arguments) {
			task := args.Get(0).(*models.Task)
			task.ID = 1
			task.Title = "Test Task"
			task.Status = "pending"
		}).
		Return(nil).
		Once()

	req, _ := http.NewRequest(http.MethodGet, "/tasks/"+taskID, nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: taskID}}
	c.Set("db", mockDB)

	handler.GetTask(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotNil(t, resp["data"])

	mockDB.AssertExpectations(t)
}

func TestGetTask_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	taskID := "999"

	mockDB.
		On("First", mock.AnythingOfType("*models.Task"), "id = ?", taskID).
		Return(gorm.ErrRecordNotFound).
		Once()

	req, _ := http.NewRequest(http.MethodGet, "/tasks/"+taskID, nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: taskID}}
	c.Set("db", mockDB)

	handler.GetTask(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	mockDB.AssertExpectations(t)
}

func TestGetTask_MissingID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest(http.MethodGet, "/tasks/", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: ""}}
	c.Set("db", mockDB)

	handler.GetTask(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateTask_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	taskID := "1"
	now := time.Now()
	updateInput := models.Task{
		Status:      "completed",
		CompletedAt: &now,
	}

	body, _ := json.Marshal(updateInput)

	mockDB.On("Begin").Return(mockDB).Once()

	mockDB.
		On("First", mock.AnythingOfType("*models.Task"), "id = ?", taskID).
		Run(func(args mock.Arguments) {
			task := args.Get(0).(*models.Task)
			task.ID = 1
			task.Title = "Test Task"
			task.Status = "pending"
		}).
		Return(nil).
		Once()

	mockDB.
		On("Save", mock.AnythingOfType("*models.Task")).
		Return(nil).
		Once()

	mockDB.On("Commit").Return(nil).Once()

	req, _ := http.NewRequest(
		http.MethodPut,
		"/tasks/"+taskID,
		bytes.NewBuffer(body),
	)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: taskID}}
	c.Set("db", mockDB)

	handler.UpdateTask(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "Task updated successfully", resp["message"])
	assert.NotNil(t, resp["data"])

	mockDB.AssertExpectations(t)
}

func TestUpdateTask_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	taskID := "999"
	updateInput := models.Task{
		Status: "completed",
	}

	body, _ := json.Marshal(updateInput)

	mockDB.On("Begin").Return(mockDB).Once()

	mockDB.
		On("First", mock.AnythingOfType("*models.Task"), "id = ?", taskID).
		Return(gorm.ErrRecordNotFound).
		Once()

	mockDB.On("Rollback").Return(nil).Once()

	req, _ := http.NewRequest(
		http.MethodPut,
		"/tasks/"+taskID,
		bytes.NewBuffer(body),
	)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: taskID}}
	c.Set("db", mockDB)

	handler.UpdateTask(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	mockDB.AssertExpectations(t)
}

func TestUpdateTask_InvalidInput(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	taskID := "1"
	body := []byte(`{"status": }`)

	req, _ := http.NewRequest(
		http.MethodPut,
		"/tasks/"+taskID,
		bytes.NewBuffer(body),
	)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: taskID}}
	c.Set("db", mockDB)

	handler.UpdateTask(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteTask_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	taskID := "1"

	mockDB.On("Begin").Return(mockDB).Once()

	mockDB.
		On("First", mock.AnythingOfType("*models.Task"), "id = ?", taskID).
		Run(func(args mock.Arguments) {
			task := args.Get(0).(*models.Task)
			task.ID = 1
			task.Title = "Test Task"
		}).
		Return(nil).
		Once()

	mockDB.
		On("Delete", mock.AnythingOfType("*models.Task")).
		Return(nil).
		Once()

	mockDB.On("Commit").Return(nil).Once()

	req, _ := http.NewRequest(http.MethodDelete, "/tasks/"+taskID, nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: taskID}}
	c.Set("db", mockDB)

	handler.DeleteTask(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "Residential entry deleted successfully", resp["message"])

	mockDB.AssertExpectations(t)
}

func TestDeleteTask_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	taskID := "999"

	mockDB.On("Begin").Return(mockDB).Once()

	mockDB.
		On("First", mock.AnythingOfType("*models.Task"), "id = ?", taskID).
		Return(gorm.ErrRecordNotFound).
		Once()

	mockDB.On("Rollback").Return(nil).Once()

	req, _ := http.NewRequest(http.MethodDelete, "/tasks/"+taskID, nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: taskID}}
	c.Set("db", mockDB)

	handler.DeleteTask(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	mockDB.AssertExpectations(t)
}

func TestDeleteTask_MissingID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest(http.MethodDelete, "/tasks/", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: ""}}
	c.Set("db", mockDB)

	handler.DeleteTask(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetAllTasks_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	mockDB.
		On("Find", mock.AnythingOfType("*[]models.Task")).
		Run(func(args mock.Arguments) {
			tasks := args.Get(0).(*[]models.Task)
			*tasks = []models.Task{
				{ID: 1, Title: "Task 1", Status: "pending"},
				{ID: 2, Title: "Task 2", Status: "completed"},
			}
		}).
		Return(nil).
		Once()

	req, _ := http.NewRequest(http.MethodGet, "/tasks", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.GetAllTasks(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotNil(t, resp["data"])

	mockDB.AssertExpectations(t)
}

func TestGetAllTasks_DatabaseError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	mockDB.
		On("Find", mock.AnythingOfType("*[]models.Task")).
		Return(gorm.ErrInvalidDB).
		Once()

	req, _ := http.NewRequest(http.MethodGet, "/tasks", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.GetAllTasks(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	mockDB.AssertExpectations(t)
}
