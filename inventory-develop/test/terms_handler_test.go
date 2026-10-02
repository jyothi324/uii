package test

import (
	"bytes"
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
	"gorm.io/gorm"
)

func TestHandleSaveTerms(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("save policies successfully", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Model", mock.Anything).Return(mockDB)
		mockDB.On("Delete", mock.Anything, "1 = 1").Return(nil)
		mockDB.On("Create", mock.Anything).Return(nil)

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/settings/terms", handlers.HandleSaveTerms())

		body := map[string]interface{}{
			"policies": []map[string]string{
				{"title": "Warranty", "content": "Warranty terms"},
			},
		}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/settings/terms", bytes.NewBuffer(jsonBody))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("replace existing policies successfully", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Model", mock.Anything).Return(mockDB)
		mockDB.On("Delete", mock.Anything, "1 = 1").Return(nil)
		mockDB.On("Create", mock.Anything).Return(nil)

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/settings/terms", handlers.HandleSaveTerms())

		body := map[string]interface{}{
			"policies": []map[string]string{
				{"title": "Warranty", "content": "Updated warranty"},
				{"title": "Delivery", "content": "Updated delivery"},
			},
		}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/settings/terms", bytes.NewBuffer(jsonBody))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestHandleGetTerms(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("fetch policies successfully", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Order", "position asc").Return(mockDB)
		mockDB.On("Find", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			out := args.Get(0).(*[]models.Policy)
			*out = []models.Policy{
				{ID: 1, Title: "Warranty", Content: "Stored warranty"},
				{ID: 2, Title: "Delivery", Content: "Stored delivery"},
			}
		})

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.GET("/settings/terms", handlers.HandleGetTerms())
		req, _ := http.NewRequest("GET", "/settings/terms", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Stored warranty")
		assert.Contains(t, w.Body.String(), "Stored delivery")
	})

	// t.Run("falls back to legacy terms successfully", func(t *testing.T) {
	// 	mockDB := &mocks.MockDatabase{}
	// 	mockDB.On("Order", "position asc").Return(mockDB)
	// 	mockDB.On("Find", mock.Anything).Return(nil)
	// 	mockDB.On("Model", mock.Anything).Return(mockDB)
	// 	mockDB.On("First", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
	// 		// out := args.Get(0).(*models.Terms)
	// 		// out.Content = "Stored content"
	// 	})

	// 	router := gin.New()
	// 	router.Use(injectDB(mockDB))
	// 	router.GET("/settings/terms", handlers.HandleGetTerms())
	// 	req, _ := http.NewRequest("GET", "/settings/terms", nil)
	// 	w := httptest.NewRecorder()
	// 	router.ServeHTTP(w, req)
	// 	assert.Equal(t, http.StatusOK, w.Code)
	// 	assert.Contains(t, w.Body.String(), "Stored content")
	// })
}

func TestHandleSaveVerbiage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("create new verbiage successfully", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Model", mock.Anything).Return(mockDB)
		mockDB.On("First", mock.Anything).Return(gorm.ErrRecordNotFound)
		mockDB.On("Create", mock.Anything).Return(nil)

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/settings/verbiage", handlers.HandleSaveVerbiage())

		body := map[string]string{
			"change_order": "New change order text",
			"cancellation": "New cancellation text",
			"disputes":     "New disputes text",
		}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/settings/verbiage", bytes.NewBuffer(jsonBody))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "saved successfully")
	})

	t.Run("update existing verbiage successfully", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Model", mock.Anything).Return(mockDB)
		mockDB.On("First", mock.Anything).Return(nil)
		mockDB.On("Save", mock.Anything).Return(nil)

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/settings/verbiage", handlers.HandleSaveVerbiage())

		body := map[string]string{
			"change_order": "Updated change order",
			"cancellation": "Updated cancellation",
			"disputes":     "Updated disputes",
		}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/settings/verbiage", bytes.NewBuffer(jsonBody))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("returns 400 on invalid JSON body", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/settings/verbiage", handlers.HandleSaveVerbiage())

		req, _ := http.NewRequest("POST", "/settings/verbiage", bytes.NewBufferString("{invalid json"))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid request")
	})

	t.Run("returns 500 on real DB error during First", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Model", mock.Anything).Return(mockDB)
		mockDB.On("First", mock.Anything).Return(errors.New("connection reset"))

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/settings/verbiage", handlers.HandleSaveVerbiage())

		body := map[string]string{"change_order": "text"}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/settings/verbiage", bytes.NewBuffer(jsonBody))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "DB error")
	})

	t.Run("returns 500 when Create fails", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Model", mock.Anything).Return(mockDB)
		mockDB.On("First", mock.Anything).Return(gorm.ErrRecordNotFound)
		mockDB.On("Create", mock.Anything).Return(errors.New("insert failed"))

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/settings/verbiage", handlers.HandleSaveVerbiage())

		body := map[string]string{"change_order": "text"}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/settings/verbiage", bytes.NewBuffer(jsonBody))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to save verbiage")
	})

	t.Run("returns 500 when Save fails", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Model", mock.Anything).Return(mockDB)
		mockDB.On("First", mock.Anything).Return(nil)
		mockDB.On("Save", mock.Anything).Return(errors.New("update failed"))

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/settings/verbiage", handlers.HandleSaveVerbiage())

		body := map[string]string{"change_order": "text"}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/settings/verbiage", bytes.NewBuffer(jsonBody))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to update verbiage")
	})
}

func TestHandleGetVerbiage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("fetch verbiage successfully", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Model", mock.Anything).Return(mockDB)
		mockDB.On("First", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			out := args.Get(0).(*models.Verbiage)
			out.ChangeOrder = "Stored change order"
			out.Cancellation = "Stored cancellation"
			out.Disputes = "Stored disputes"
		})

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.GET("/settings/verbiage", handlers.HandleGetVerbiage())

		req, _ := http.NewRequest("GET", "/settings/verbiage", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Stored change order")
		assert.Contains(t, w.Body.String(), "Stored cancellation")
		assert.Contains(t, w.Body.String(), "Stored disputes")
	})

	t.Run("returns empty strings when no record found", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Model", mock.Anything).Return(mockDB)
		mockDB.On("First", mock.Anything).Return(gorm.ErrRecordNotFound)

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.GET("/settings/verbiage", handlers.HandleGetVerbiage())

		req, _ := http.NewRequest("GET", "/settings/verbiage", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]string
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "", resp["change_order"])
		assert.Equal(t, "", resp["cancellation"])
		assert.Equal(t, "", resp["disputes"])
	})

	t.Run("returns 500 on real DB error", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Model", mock.Anything).Return(mockDB)
		mockDB.On("First", mock.Anything).Return(errors.New("connection reset"))

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.GET("/settings/verbiage", handlers.HandleGetVerbiage())

		req, _ := http.NewRequest("GET", "/settings/verbiage", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "DB error")
	})
}
