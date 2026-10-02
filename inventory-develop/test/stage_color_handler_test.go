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
	"gorm.io/gorm"
)

// ── HandleSaveStageColors ─────────────────────────────────────────────────────

func TestHandleSaveStageColors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successfully creates a new stage color", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		// Mock to return record not found, indicating it's a new stage color
		mockDB.On("Where", "stage = ?", "Template").Return(mockDB)
		mockDB.On("First", mock.AnythingOfType("*models.StageColor")).Return(gorm.ErrRecordNotFound).Once()
		mockDB.On("Create", mock.AnythingOfType("*models.StageColor")).Return(nil).Once()

		router := gin.New()
		router.Use(func(c *gin.Context) { c.Set("db", mockDB); c.Next() })
		router.POST("/settings/stage-colors", handlers.HandleSaveStageColors())

		reqBody := map[string]interface{}{
			"colors": []map[string]string{
				{"stage_name": "Template", "color_hex": "#4B5EAA"},
			},
		}
		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/settings/stage-colors", bytes.NewBuffer(jsonBody))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Stage colors saved successfully")
		mockDB.AssertExpectations(t)
	})

	t.Run("successfully updates an existing stage color", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		existingColor := models.StageColor{ID: 1, Stage: "Template", Color: "#000000"}

		// Mock to return an existing record, indicating it should be updated
		mockDB.On("Where", "stage = ?", "Template").Return(mockDB)
		mockDB.On("First", mock.AnythingOfType("*models.StageColor")).
			Return(nil).
			Run(func(args mock.Arguments) {
				*args.Get(0).(*models.StageColor) = existingColor
			}).Once()
		mockDB.On("Save", mock.AnythingOfType("*models.StageColor")).Return(nil).Once()

		router := gin.New()
		router.Use(func(c *gin.Context) { c.Set("db", mockDB); c.Next() })
		router.POST("/settings/stage-colors", handlers.HandleSaveStageColors())

		reqBody := map[string]interface{}{
			"colors": []map[string]string{
				{"stage_name": "Template", "color_hex": "#FFFFFF"},
			},
		}
		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/settings/stage-colors", bytes.NewBuffer(jsonBody))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Stage colors saved successfully")
		mockDB.AssertExpectations(t)
	})

	t.Run("returns 400 for invalid JSON body", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		router := gin.New()
		router.Use(func(c *gin.Context) { c.Set("db", mockDB); c.Next() })
		router.POST("/settings/stage-colors", handlers.HandleSaveStageColors())

		req, _ := http.NewRequest("POST", "/settings/stage-colors", bytes.NewBuffer([]byte(`{invalid-json}`)))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request body")
	})

	t.Run("returns 400 when colors array is empty", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		router := gin.New()
		router.Use(func(c *gin.Context) { c.Set("db", mockDB); c.Next() })
		router.POST("/settings/stage-colors", handlers.HandleSaveStageColors())

		req, _ := http.NewRequest("POST", "/settings/stage-colors", bytes.NewBuffer([]byte(`{"colors": []}`)))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "colors array is required")
	})

	t.Run("returns 400 when stage_name or color_hex is missing", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		router := gin.New()
		router.Use(func(c *gin.Context) { c.Set("db", mockDB); c.Next() })
		router.POST("/settings/stage-colors", handlers.HandleSaveStageColors())

		reqBody := map[string]interface{}{
			"colors": []map[string]string{
				{"stage_name": "Template"}, // Missing color_hex
			},
		}
		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/settings/stage-colors", bytes.NewBuffer(jsonBody))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "stage_name and color_hex are required")
	})

	t.Run("returns 500 when querying database fails", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Where", "stage = ?", "Template").Return(mockDB)
		mockDB.On("First", mock.AnythingOfType("*models.StageColor")).Return(errors.New("db error")).Once()

		router := gin.New()
		router.Use(func(c *gin.Context) { c.Set("db", mockDB); c.Next() })
		router.POST("/settings/stage-colors", handlers.HandleSaveStageColors())

		reqBody := map[string]interface{}{
			"colors": []map[string]string{
				{"stage_name": "Template", "color_hex": "#4B5EAA"},
			},
		}
		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/settings/stage-colors", bytes.NewBuffer(jsonBody))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to query stage colors")
	})

	t.Run("returns 500 when saving a new color fails", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Where", "stage = ?", "Template").Return(mockDB)
		mockDB.On("First", mock.AnythingOfType("*models.StageColor")).Return(gorm.ErrRecordNotFound).Once()
		mockDB.On("Create", mock.AnythingOfType("*models.StageColor")).Return(errors.New("insert error")).Once()

		router := gin.New()
		router.Use(func(c *gin.Context) { c.Set("db", mockDB); c.Next() })
		router.POST("/settings/stage-colors", handlers.HandleSaveStageColors())

		reqBody := map[string]interface{}{
			"colors": []map[string]string{
				{"stage_name": "Template", "color_hex": "#4B5EAA"},
			},
		}
		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/settings/stage-colors", bytes.NewBuffer(jsonBody))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to save stage color")
	})
}

// ── GetStageColors ────────────────────────────────────────────────────────────

func TestGetStageColors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successfully fetches stage colors", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Find", mock.AnythingOfType("*[]models.StageColor")).Return(nil).Run(func(args mock.Arguments) {
			colors := args.Get(0).(*[]models.StageColor)
			*colors = []models.StageColor{
				{ID: 1, Stage: "Template", Color: "#4B5EAA"},
			}
		})

		router := gin.New()
		router.Use(func(c *gin.Context) { c.Set("db", mockDB); c.Next() })
		router.GET("/settings/stage-colors", handlers.GetStageColors())

		req, _ := http.NewRequest("GET", "/settings/stage-colors", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Template")
		assert.Contains(t, w.Body.String(), "#4B5EAA")
		mockDB.AssertExpectations(t)
	})

	t.Run("returns 500 when querying stage colors fails", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Find", mock.AnythingOfType("*[]models.StageColor")).Return(errors.New("db error"))

		router := gin.New()
		router.Use(func(c *gin.Context) { c.Set("db", mockDB); c.Next() })
		router.GET("/settings/stage-colors", handlers.GetStageColors())

		req, _ := http.NewRequest("GET", "/settings/stage-colors", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to fetch stage colors")
		mockDB.AssertExpectations(t)
	})
}
