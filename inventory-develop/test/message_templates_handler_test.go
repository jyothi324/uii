package test

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"inventory/handlers"
	"inventory/mocks"
	"inventory/models"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// AnyTime matches any time.Time value.
type AnyTime struct{}

// Match satisfies sqlmock.Argument interface
func (a AnyTime) Match(v driver.Value) bool {
	// This is a placeholder now that we're not using sqlmock directly for this test.
	// The logic can be simplified or adapted if needed for other mock arguments.
	return true
}

func setupMessageTemplateTestRouter() (*gin.Engine, *mocks.MockDatabase) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	mockDB := new(mocks.MockDatabase)

	r.Use(func(c *gin.Context) {
		c.Set("db", mockDB)
		c.Next()
	})

	return r, mockDB
}

func TestHandleListMessageTemplates(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		r, mockDB := setupMessageTemplateTestRouter()
		r.GET("/templates", handlers.HandleListMessageTemplates())

		mockDB.On("Model", mock.Anything).Return(mockDB)
		mockDB.On("Order", "created_at desc").Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]models.MessageTemplate")).Return(nil).Run(func(args mock.Arguments) {
			arg := args.Get(0).(*[]models.MessageTemplate)
			*arg = []models.MessageTemplate{{ID: 1, Name: "Test Template"}}
		})

		req, _ := http.NewRequest(http.MethodGet, "/templates", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var templates []models.MessageTemplate
		err := json.Unmarshal(w.Body.Bytes(), &templates)
		assert.NoError(t, err)
		assert.Len(t, templates, 1)
		assert.Equal(t, "Test Template", templates[0].Name)
		mockDB.AssertExpectations(t)
	})

	t.Run("db error", func(t *testing.T) {
		r, mockDB := setupMessageTemplateTestRouter()
		r.GET("/templates", handlers.HandleListMessageTemplates())

		mockDB.On("Model", mock.Anything).Return(mockDB)
		mockDB.On("Order", "created_at desc").Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]models.MessageTemplate")).Return(errors.New("db error"))

		req, _ := http.NewRequest(http.MethodGet, "/templates", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to retrieve templates")
		mockDB.AssertExpectations(t)
	})

	t.Run("no templates found", func(t *testing.T) {
		r, mockDB := setupMessageTemplateTestRouter()
		r.GET("/templates", handlers.HandleListMessageTemplates())

		mockDB.On("Model", mock.Anything).Return(mockDB)
		mockDB.On("Order", "created_at desc").Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]models.MessageTemplate")).Return(nil)

		req, _ := http.NewRequest(http.MethodGet, "/templates", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "[]", w.Body.String())
		mockDB.AssertExpectations(t)
	})
}

func TestHandleCreateMessageTemplate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		r, mockDB := setupMessageTemplateTestRouter()
		r.POST("/templates", handlers.HandleCreateMessageTemplate())

		input := handlers.MessageTemplateInput{
			Name:     "New Template",
			Subject:  "New Subject",
			Content:  "New Content",
			Category: "New Category",
		}
		body, _ := json.Marshal(input)

		mockDB.On("Model", mock.Anything).Return(mockDB)
		mockDB.On("Where", "lower(name) = lower(?)", input.Name).Return(mockDB)
		mockDB.On("First", mock.Anything).Return(errors.New("record not found")) // gorm.ErrRecordNotFound
		mockDB.On("Create", mock.AnythingOfType("*models.MessageTemplate")).Return(nil)

		req, _ := http.NewRequest(http.MethodPost, "/templates", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var template models.MessageTemplate
		err := json.Unmarshal(w.Body.Bytes(), &template)
		assert.NoError(t, err)
		assert.Equal(t, input.Name, template.Name)
		mockDB.AssertExpectations(t)
	})

	t.Run("invalid request", func(t *testing.T) {
		r, _ := setupMessageTemplateTestRouter()
		r.POST("/templates", handlers.HandleCreateMessageTemplate())

		req, _ := http.NewRequest(http.MethodPost, "/templates", bytes.NewBufferString(`{"name":"test"}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request")
	})

	t.Run("duplicate name", func(t *testing.T) {
		r, mockDB := setupMessageTemplateTestRouter()
		r.POST("/templates", handlers.HandleCreateMessageTemplate())

		input := handlers.MessageTemplateInput{
			Name:     "Existing Template",
			Subject:  "Subject",
			Content:  "Content",
			Category: "Category",
		}
		body, _ := json.Marshal(input)

		mockDB.On("Model", mock.Anything).Return(mockDB)
		mockDB.On("Where", "lower(name) = lower(?)", input.Name).Return(mockDB)
		mockDB.On("First", mock.Anything).Return(nil) // Found a record

		req, _ := http.NewRequest(http.MethodPost, "/templates", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Template name already exists")
		mockDB.AssertExpectations(t)
	})

	t.Run("create error", func(t *testing.T) {
		r, mockDB := setupMessageTemplateTestRouter()
		r.POST("/templates", handlers.HandleCreateMessageTemplate())

		input := handlers.MessageTemplateInput{
			Name:     "New Template",
			Subject:  "New Subject",
			Content:  "New Content",
			Category: "New Category",
		}
		body, _ := json.Marshal(input)

		mockDB.On("Model", mock.Anything).Return(mockDB)
		mockDB.On("Where", "lower(name) = lower(?)", input.Name).Return(mockDB)
		mockDB.On("First", mock.Anything).Return(errors.New("record not found"))
		mockDB.On("Create", mock.AnythingOfType("*models.MessageTemplate")).Return(errors.New("db create error"))

		req, _ := http.NewRequest(http.MethodPost, "/templates", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to create template")
		mockDB.AssertExpectations(t)
	})
}

func TestHandleUpdateMessageTemplate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		r, mockDB := setupMessageTemplateTestRouter()
		r.PUT("/templates/:id", handlers.HandleUpdateMessageTemplate())

		input := handlers.MessageTemplateInput{
			Name:     "Updated Name",
			Subject:  "Updated Subject",
			Content:  "Updated Content",
			Category: "Updated Category",
		}
		body, _ := json.Marshal(input)

		mockDB.On("Model", mock.Anything).Return(mockDB).Times(2)
		// Check for duplicate name - not found
		mockDB.On("Where", "lower(name) = lower(?) AND id != ?", input.Name, uint64(1)).Return(mockDB)
		mockDB.On("First", mock.Anything).Return(errors.New("record not found")).Once()
		// Find existing template - found
		mockDB.On("Where", "id = ?", uint64(1)).Return(mockDB)
		mockDB.On("First", mock.Anything).Return(nil).Once()
		mockDB.On("Save", mock.AnythingOfType("*models.MessageTemplate")).Return(nil)

		req, _ := http.NewRequest(http.MethodPut, "/templates/1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var template models.MessageTemplate
		err := json.Unmarshal(w.Body.Bytes(), &template)
		assert.NoError(t, err)
		assert.Equal(t, input.Name, template.Name)
		mockDB.AssertExpectations(t)
	})

	t.Run("invalid id", func(t *testing.T) {
		r, _ := setupMessageTemplateTestRouter()
		r.PUT("/templates/:id", handlers.HandleUpdateMessageTemplate())

		req, _ := http.NewRequest(http.MethodPut, "/templates/abc", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid template ID")
	})

	t.Run("template not found", func(t *testing.T) {
		r, mockDB := setupMessageTemplateTestRouter()
		r.PUT("/templates/:id", handlers.HandleUpdateMessageTemplate())

		input := handlers.MessageTemplateInput{
			Name: "Updated Name", Subject: "S", Content: "C", Category: "C",
		}
		body, _ := json.Marshal(input)

		mockDB.On("Model", mock.Anything).Return(mockDB).Times(2)
		// Check for duplicate name - not found
		mockDB.On("Where", "lower(name) = lower(?) AND id != ?", input.Name, uint64(99)).Return(mockDB)
		mockDB.On("First", mock.Anything).Return(errors.New("record not found")).Once()
		// Find existing template - not found
		mockDB.On("Where", "id = ?", uint64(99)).Return(mockDB)
		mockDB.On("First", mock.Anything).Return(errors.New("record not found")).Once()

		req, _ := http.NewRequest(http.MethodPut, "/templates/99", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Template not found")
		mockDB.AssertExpectations(t)
	})

	t.Run("duplicate name on update", func(t *testing.T) {
		r, mockDB := setupMessageTemplateTestRouter()
		r.PUT("/templates/:id", handlers.HandleUpdateMessageTemplate())

		input := handlers.MessageTemplateInput{
			Name: "Existing Name", Subject: "S", Content: "C", Category: "C",
		}
		body, _ := json.Marshal(input)

		mockDB.On("Model", mock.Anything).Return(mockDB)
		mockDB.On("Where", "lower(name) = lower(?) AND id != ?", input.Name, uint64(1)).Return(mockDB)
		mockDB.On("First", mock.Anything).Return(nil) // Found a duplicate

		req, _ := http.NewRequest(http.MethodPut, "/templates/1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Template name already exists")
		mockDB.AssertExpectations(t)
	})

	t.Run("save error", func(t *testing.T) {
		r, mockDB := setupMessageTemplateTestRouter()
		r.PUT("/templates/:id", handlers.HandleUpdateMessageTemplate())

		input := handlers.MessageTemplateInput{
			Name: "Updated Name", Subject: "S", Content: "C", Category: "C",
		}
		body, _ := json.Marshal(input)

		mockDB.On("Model", mock.Anything).Return(mockDB).Times(2)
		// Check for duplicate name - not found
		mockDB.On("Where", "lower(name) = lower(?) AND id != ?", input.Name, uint64(1)).Return(mockDB)
		mockDB.On("First", mock.Anything).Return(errors.New("record not found")).Once()
		// Find existing template - found
		mockDB.On("Where", "id = ?", uint64(1)).Return(mockDB)
		mockDB.On("First", mock.Anything).Return(nil).Once()
		// Save fails
		mockDB.On("Save", mock.AnythingOfType("*models.MessageTemplate")).Return(errors.New("db save error"))

		req, _ := http.NewRequest(http.MethodPut, "/templates/1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to update template")
		mockDB.AssertExpectations(t)
	})
}

func TestHandleDeleteMessageTemplate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		r, mockDB := setupMessageTemplateTestRouter()
		r.DELETE("/templates/:id", handlers.HandleDeleteMessageTemplate())

		mockDB.On("Model", mock.Anything).Return(mockDB)
		mockDB.On("Where", "id = ?", uint64(1)).Return(mockDB)
		mockDB.On("Take", mock.Anything).Return(nil)
		mockDB.On("Delete", mock.Anything).Return(nil)

		req, _ := http.NewRequest(http.MethodDelete, "/templates/1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Template deleted successfully")
		mockDB.AssertExpectations(t)
	})

	t.Run("template not found", func(t *testing.T) {
		r, mockDB := setupMessageTemplateTestRouter()
		r.DELETE("/templates/:id", handlers.HandleDeleteMessageTemplate())

		mockDB.On("Model", mock.Anything).Return(mockDB)
		mockDB.On("Where", "id = ?", uint64(99)).Return(mockDB)
		mockDB.On("Take", mock.Anything).Return(errors.New("record not found"))

		req, _ := http.NewRequest(http.MethodDelete, "/templates/99", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Template not found")
		mockDB.AssertExpectations(t)
	})

	t.Run("delete error", func(t *testing.T) {
		r, mockDB := setupMessageTemplateTestRouter()
		r.DELETE("/templates/:id", handlers.HandleDeleteMessageTemplate())

		mockDB.On("Model", mock.Anything).Return(mockDB)
		mockDB.On("Where", "id = ?", uint64(1)).Return(mockDB)
		mockDB.On("Take", mock.Anything).Return(nil)
		mockDB.On("Delete", mock.Anything).Return(errors.New("db delete error"))

		req, _ := http.NewRequest(http.MethodDelete, "/templates/1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to delete template")
		mockDB.AssertExpectations(t)
	})
}
