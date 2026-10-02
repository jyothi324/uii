package test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"inventory/handlers"
	"inventory/mocks"
	"inventory/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateFeedback(t *testing.T) {
	gin.SetMode(gin.TestMode)

	phoneStr := "+91-9876543210"

	// ── Happy paths ────────────────────────────────────────────────────────────

	t.Run("valid request - all fields including phone", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Create", mock.AnythingOfType("*models.Feedback")).Return(nil)

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/feedback", handlers.CreateFeedback())

		body := map[string]interface{}{
			"firstName": "John",
			"lastName":  "Doe",
			"email":     "john@example.com",
			"phone":     phoneStr,
			"message":   "Great service!",
		}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/feedback", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("valid request - phone omitted (nil)", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Create", mock.AnythingOfType("*models.Feedback")).Return(nil)

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/feedback", handlers.CreateFeedback())

		body := map[string]interface{}{
			"firstName": "Jane",
			"lastName":  "Smith",
			"email":     "jane@example.com",
			"message":   "Loved it!",
		}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/feedback", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("valid request - fields get trimmed before save", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Create", mock.AnythingOfType("*models.Feedback")).
			Return(nil).
			Run(func(args mock.Arguments) {
				// Whitespace must have been stripped before reaching the DB
				f := args.Get(0).(*models.Feedback)
				assert.Equal(t, "Alice", f.FirstName)
				assert.Equal(t, "Wonder", f.LastName)
				assert.Equal(t, "alice@example.com", f.Email)
				assert.Equal(t, "Good stuff", f.Message)
			})

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/feedback", handlers.CreateFeedback())

		body := map[string]interface{}{
			"firstName": "  Alice  ",
			"lastName":  "  Wonder  ",
			"email":     "  alice@example.com  ",
			"message":   "  Good stuff  ",
		}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/feedback", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("valid request - message exactly 1000 characters", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Create", mock.AnythingOfType("*models.Feedback")).Return(nil)

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/feedback", handlers.CreateFeedback())

		body := map[string]interface{}{
			"firstName": "Max",
			"lastName":  "Len",
			"email":     "max@example.com",
			"message":   string(make([]byte, 1000)), // 1000 zero-bytes → valid
		}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/feedback", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockDB.AssertExpectations(t)
	})

	// ── Malformed / wrong-type JSON ────────────────────────────────────────────

	t.Run("invalid request body - malformed JSON", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/feedback", handlers.CreateFeedback())

		req, _ := http.NewRequest("POST", "/feedback",
			bytes.NewBuffer([]byte(`{"firstName": 123}`))) // firstName must be string
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request body")
	})

	// ── Required-field validation ──────────────────────────────────────────────

	t.Run("missing firstName", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/feedback", handlers.CreateFeedback())

		body := map[string]interface{}{
			"lastName": "Doe",
			"email":    "john@example.com",
			"message":  "Hello",
		}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/feedback", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "First name is required")
	})

	t.Run("blank firstName (whitespace only)", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/feedback", handlers.CreateFeedback())

		body := map[string]interface{}{
			"firstName": "   ",
			"lastName":  "Doe",
			"email":     "john@example.com",
			"message":   "Hello",
		}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/feedback", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "First name is required")
	})

	t.Run("missing lastName", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/feedback", handlers.CreateFeedback())

		body := map[string]interface{}{
			"firstName": "John",
			"email":     "john@example.com",
			"message":   "Hello",
		}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/feedback", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Last name is required")
	})

	t.Run("blank lastName (whitespace only)", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/feedback", handlers.CreateFeedback())

		body := map[string]interface{}{
			"firstName": "John",
			"lastName":  "   ",
			"email":     "john@example.com",
			"message":   "Hello",
		}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/feedback", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Last name is required")
	})

	t.Run("missing email", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/feedback", handlers.CreateFeedback())

		body := map[string]interface{}{
			"firstName": "John",
			"lastName":  "Doe",
			"message":   "Hello",
		}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/feedback", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Email is required")
	})

	t.Run("blank email (whitespace only)", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/feedback", handlers.CreateFeedback())

		body := map[string]interface{}{
			"firstName": "John",
			"lastName":  "Doe",
			"email":     "   ",
			"message":   "Hello",
		}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/feedback", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Email is required")
	})

	t.Run("missing message", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/feedback", handlers.CreateFeedback())

		body := map[string]interface{}{
			"firstName": "John",
			"lastName":  "Doe",
			"email":     "john@example.com",
		}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/feedback", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Feedback message is required")
	})

	t.Run("blank message (whitespace only)", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/feedback", handlers.CreateFeedback())

		body := map[string]interface{}{
			"firstName": "John",
			"lastName":  "Doe",
			"email":     "john@example.com",
			"message":   "   ",
		}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/feedback", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Feedback message is required")
	})

	// ── Message length boundary ────────────────────────────────────────────────

	t.Run("message exceeds 1000 characters", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/feedback", handlers.CreateFeedback())

		body := map[string]interface{}{
			"firstName": "John",
			"lastName":  "Doe",
			"email":     "john@example.com",
			"message":   string(make([]byte, 1001)), // 1001 bytes → invalid
		}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/feedback", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Message must be 1000 characters or fewer")
	})

	// ── Database error ─────────────────────────────────────────────────────────

	t.Run("database error on insert", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Create", mock.AnythingOfType("*models.Feedback")).
			Return(errors.New("db insert error"))

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/feedback", handlers.CreateFeedback())

		body := map[string]interface{}{
			"firstName": "John",
			"lastName":  "Doe",
			"email":     "john@example.com",
			"message":   "Hello!",
		}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/feedback", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to save feedback")
		mockDB.AssertExpectations(t)
	})
}