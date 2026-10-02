package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"inventory/handlers"
	"inventory/mocks"
	"inventory/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestHandleUpsertJobTaskForm(t *testing.T) {
	gin.SetMode(gin.TestMode)

t.Run("creates a new task form", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		mockDB.On("Model", mock.Anything).Return(mockDB)
		mockDB.On("Where", "job_id = ? AND task_id = ? AND stage = ?", "JOB-1", "TASK-1", "Template").Return(mockDB)
		mockDB.On("First", mock.AnythingOfType("*models.JobTaskForm")).Return(gorm.ErrRecordNotFound).Once()
		mockDB.On("Create", mock.AnythingOfType("*models.JobTaskForm")).Return(nil).Once()

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/forms/tasks", handlers.HandleUpsertJobTaskForm())

		body := map[string]any{
			"job_id":      "JOB-1",
			"task_id":     "TASK-1",
			"task_name":   "Measure Site",
			"stage":       "Template",
			"salesperson": "Alex",
		}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/forms/tasks", bytes.NewBuffer(jsonBody))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), "Task form created successfully")
		mockDB.AssertExpectations(t)
	})

	t.Run("updates an existing task form", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		existing := models.JobTaskForm{ID: 3, JobID: "JOB-9", TaskID: "TASK-4", Stage: "Install"}

		mockDB.On("Model", mock.Anything).Return(mockDB)
		mockDB.On("Where", "job_id = ? AND task_id = ? AND stage = ?", "JOB-9", "TASK-4", "Install").Return(mockDB)
		mockDB.On("First", mock.AnythingOfType("*models.JobTaskForm")).
			Return(nil).
			Run(func(args mock.Arguments) {
				*args.Get(0).(*models.JobTaskForm) = existing
			}).Once()
		mockDB.On("Save", mock.AnythingOfType("*models.JobTaskForm")).Return(nil).Once()

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/forms/tasks", handlers.HandleUpsertJobTaskForm())

		body := map[string]any{
			"job_id":      "JOB-9",
			"task_id":     "TASK-4",
			"task_name":   "Install Countertop",
			"stage":       "Install",
			"salesperson": "Robin",
			"form_data":   map[string]any{"title": "Install Form"},
		}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/forms/tasks", bytes.NewBuffer(jsonBody))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Task form updated successfully")
		mockDB.AssertExpectations(t)
	})
}
