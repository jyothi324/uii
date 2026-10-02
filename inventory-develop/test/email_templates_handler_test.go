package test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"inventory/handlers"
	"inventory/mocks"
	"inventory/models"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var testDB *gorm.DB

func setupEmailTemplateRouter(mockDB *mocks.MockDatabase) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Setup in-memory SQLite database for testing
	var err error
	testDB, err = gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to test database: %v", err))
	}

	// Auto-migrate schema
	testDB.AutoMigrate(&models.EmailTemplate{})

	// Middleware to inject DB into context
	router.Use(func(c *gin.Context) {
		c.Set("db", testDB)
		c.Set("db", mockDB)
		c.Next()
	})

	// Define routes
	router.GET("/settings/email-templates", handlers.HandleListEmailTemplates())
	router.POST("/settings/email-templates", handlers.HandleCreateEmailTemplate())
	router.PUT("/settings/email-templates/:id", handlers.HandleUpdateEmailTemplate())
	router.DELETE("/settings/email-templates/:id", handlers.HandleDeleteEmailTemplate())

	return router
}

func cleanupTestDB() {
	// Clean up the table after each test
	testDB.Exec("DELETE FROM email_templates")
}

func TestEmailTemplateHandlers(t *testing.T) {
	t.Run("Create_Success", func(t *testing.T) {
		defer cleanupTestDB()
		mockDB := &mocks.MockDatabase{}
		router := setupEmailTemplateRouter(mockDB)

		mockDB.On("Model", mock.Anything).Return(mockDB)
		mockDB.On("Where", "lower(name) = lower(?)", "Welcome Email").Return(mockDB)
		mockDB.On("First", mock.Anything).Return(gorm.ErrRecordNotFound)
		mockDB.On("Create", mock.AnythingOfType("*models.EmailTemplate")).Return(nil).Run(func(args mock.Arguments) {
			arg := args.Get(0).(*models.EmailTemplate)
			arg.ID = 1
		})

		payload := `{"name": "Welcome Email", "subject": "Welcome!", "content": "Hello there."}`
		req, _ := http.NewRequest("POST", "/settings/email-templates", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var createdTemplate models.EmailTemplate
		json.Unmarshal(w.Body.Bytes(), &createdTemplate)

		assert.Equal(t, "Welcome Email", createdTemplate.Name)
		assert.Equal(t, "Welcome!", createdTemplate.Subject)
		assert.NotZero(t, createdTemplate.ID)
	})

	t.Run("Create_DuplicateName", func(t *testing.T) {
		defer cleanupTestDB()
		mockDB := &mocks.MockDatabase{}
		router := setupEmailTemplateRouter(mockDB)

		// Pre-populate a template
		testDB.Create(&models.EmailTemplate{Name: "Duplicate Name"})
		mockDB.On("Model", mock.Anything).Return(mockDB)
		mockDB.On("Where", "lower(name) = lower(?)", "Duplicate Name").Return(mockDB)
		// Simulate finding an existing record
		mockDB.On("First", mock.Anything).Return(nil)

		payload := `{"name": "Duplicate Name"}`
		req, _ := http.NewRequest("POST", "/settings/email-templates", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "template name already exists")
	})

	t.Run("Create_MissingName", func(t *testing.T) {
		defer cleanupTestDB()
		mockDB := &mocks.MockDatabase{}
		router := setupEmailTemplateRouter(mockDB)

		payload := `{"subject": "No name subject"}`
		req, _ := http.NewRequest("POST", "/settings/email-templates", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "name is required")
	})

	t.Run("Create_InvalidJSON", func(t *testing.T) {
		defer cleanupTestDB()
		mockDB := &mocks.MockDatabase{}
		router := setupEmailTemplateRouter(mockDB)

		payload := `{name: Welcome Email}`
		req, _ := http.NewRequest("POST", "/settings/email-templates", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid request")
	})

	t.Run("List_Success", func(t *testing.T) {
		defer cleanupTestDB()
		mockDB := &mocks.MockDatabase{}
		router := setupEmailTemplateRouter(mockDB)

		// Pre-populate templates
		testDB.Create(&models.EmailTemplate{Name: "Template A"})
		testDB.Create(&models.EmailTemplate{Name: "Template B"})
		templatesMock := []models.EmailTemplate{
			{ID: 2, Name: "Template B"},
			{ID: 1, Name: "Template A"},
		}

		mockDB.On("Model", mock.Anything).Return(mockDB)
		mockDB.On("Order", "created_at desc").Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]models.EmailTemplate")).Return(nil).Run(func(args mock.Arguments) {
			arg := args.Get(0).(*[]models.EmailTemplate)
			*arg = templatesMock
		})

		req, _ := http.NewRequest("GET", "/settings/email-templates", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var templates []models.EmailTemplate
		json.Unmarshal(w.Body.Bytes(), &templates)

		assert.Len(t, templates, 2)
		assert.Equal(t, "Template B", templates[0].Name) // Ordered by created_at desc
		assert.Equal(t, "Template A", templates[1].Name)
	})

	t.Run("List_Empty", func(t *testing.T) {
		defer cleanupTestDB()
		mockDB := &mocks.MockDatabase{}
		router := setupEmailTemplateRouter(mockDB)

		mockDB.On("Model", mock.Anything).Return(mockDB)
		mockDB.On("Order", "created_at desc").Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]models.EmailTemplate")).Return(nil).Run(func(args mock.Arguments) {
			arg := args.Get(0).(*[]models.EmailTemplate)
			*arg = []models.EmailTemplate{}
		})

		req, _ := http.NewRequest("GET", "/settings/email-templates", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "[]", w.Body.String())
	})

	t.Run("Update_Success", func(t *testing.T) {
		defer cleanupTestDB()
		mockDB := &mocks.MockDatabase{}
		router := setupEmailTemplateRouter(mockDB)

		tmpl := models.EmailTemplate{Name: "Original Name", Subject: "Old Subject"}
		testDB.Create(&tmpl)
		tmpl = models.EmailTemplate{ID: 1, Name: "Original Name", Subject: "Old Subject"}

		mockDB.On("Model", mock.Anything).Return(mockDB)
		mockDB.On("Where", "id = ?", uint64(1)).Return(mockDB)
		mockDB.On("First", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			arg := args.Get(0).(*models.EmailTemplate)
			*arg = tmpl
		}).Once()

		// Uniqueness check for new name
		mockDB.On("Where", "lower(name) = lower(?) AND id != ?", "Updated Name", uint64(1)).Return(mockDB)
		mockDB.On("First", mock.Anything).Return(gorm.ErrRecordNotFound).Once()

		mockDB.On("Save", mock.AnythingOfType("*models.EmailTemplate")).Return(nil)

		payload := `{"name": "Updated Name", "subject": "New Subject"}`
		url := fmt.Sprintf("/settings/email-templates/%d", tmpl.ID)
		req, _ := http.NewRequest("PUT", url, bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var updatedTemplate models.EmailTemplate
		json.Unmarshal(w.Body.Bytes(), &updatedTemplate)

		assert.Equal(t, "Updated Name", updatedTemplate.Name)
		assert.Equal(t, "New Subject", updatedTemplate.Subject)
	})

	t.Run("Update_NotFound", func(t *testing.T) {
		defer cleanupTestDB()
		mockDB := &mocks.MockDatabase{}
		router := setupEmailTemplateRouter(mockDB)

		mockDB.On("Model", mock.Anything).Return(mockDB)
		mockDB.On("Where", "id = ?", uint64(999)).Return(mockDB)
		mockDB.On("First", mock.Anything).Return(gorm.ErrRecordNotFound)

		payload := `{"name": "Does not matter"}`
		req, _ := http.NewRequest("PUT", "/settings/email-templates/999", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "template not found")
	})

	t.Run("Update_DuplicateName", func(t *testing.T) {
		defer cleanupTestDB()
		mockDB := &mocks.MockDatabase{}
		router := setupEmailTemplateRouter(mockDB)

		// Pre-populate templates
		tmplToUpdate := models.EmailTemplate{Name: "Original Name"}
		testDB.Create(&tmplToUpdate)
		existingTmpl := models.EmailTemplate{Name: "Existing Name"}
		testDB.Create(&existingTmpl)
		tmplToUpdate = models.EmailTemplate{ID: 1, Name: "Original Name"}

		// Try to update tmplToUpdate to have the same name as existingTmpl
		mockDB.On("Model", mock.Anything).Return(mockDB)
		mockDB.On("Where", "id = ?", uint64(1)).Return(mockDB)
		mockDB.On("First", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			arg := args.Get(0).(*models.EmailTemplate)
			*arg = tmplToUpdate
		}).Once()

		// Uniqueness check fails
		mockDB.On("Where", "lower(name) = lower(?) AND id != ?", "Existing Name", uint64(1)).Return(mockDB)
		mockDB.On("First", mock.Anything).Return(nil).Once() // Simulate finding a record

		payload := `{"name": "Existing Name"}`
		url := fmt.Sprintf("/settings/email-templates/%d", tmplToUpdate.ID)
		req, _ := http.NewRequest("PUT", url, bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "template name already exists")
	})

	t.Run("Delete_Success", func(t *testing.T) {
		defer cleanupTestDB()
		mockDB := &mocks.MockDatabase{}
		router := setupEmailTemplateRouter(mockDB)

		tmpl := models.EmailTemplate{Name: "To Be Deleted"}
		testDB.Create(&tmpl)
		mockDB.On("Model", mock.Anything).Return(mockDB)
		mockDB.On("Where", "id = ?", uint64(1)).Return(mockDB)
		mockDB.On("Delete", mock.AnythingOfType("*models.EmailTemplate")).Return(nil)

		url := fmt.Sprintf("/settings/email-templates/%d", tmpl.ID)
		req, _ := http.NewRequest("DELETE", url, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "deleted")
		mockDB.AssertExpectations(t)
	})

	t.Run("Delete_NotFound", func(t *testing.T) {
		defer cleanupTestDB()
		mockDB := &mocks.MockDatabase{}
		router := setupEmailTemplateRouter(mockDB)

		mockDB.On("Model", mock.Anything).Return(mockDB)
		mockDB.On("Where", "id = ?", uint64(999)).Return(mockDB)
		// Simulate a DB error that isn't gorm.ErrRecordNotFound to hit the error path
		mockDB.On("Delete", mock.AnythingOfType("*models.EmailTemplate")).Return(errors.New("some db error"))

		req, _ := http.NewRequest("DELETE", "/settings/email-templates/999", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// This handler currently returns 500 when the Delete call returns an error.
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
