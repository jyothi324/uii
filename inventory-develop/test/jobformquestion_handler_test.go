package test

import (
	"fmt"
	"bytes"
	"encoding/json"
	"inventory/db"
	"inventory/handlers"
	"inventory/middleware"
	"inventory/models"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ── Test helpers ────────────────────────────────────────────────

// setupTestDB spins up a fresh in-memory SQLite DB, migrates the models
// this handler file touches, and returns both:
//   - a db.Database wrapper (what gets injected into the gin context,
//     since middleware.GetDB asserts to this interface)
//   - the raw *gorm.DB (for direct seeding/assertions in test bodies)
func setupTestDB(t *testing.T) (db.Database, *gorm.DB) {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	gormDB, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	err = gormDB.AutoMigrate(&models.JobFormQuestion{}, &models.JobFormSectionTitles{})
	require.NoError(t, err)

	return &db.GormDB{DB: gormDB}, gormDB
}

// withDB injects the given db.Database into every request's gin context
// under the key middleware.GetDB expects.
func withDB(database db.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(middleware.DBContextKey, database)
		c.Next()
	}
}

func setupRouter(t *testing.T, database db.Database) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(withDB(database))

	r.GET("/settings/job-form-questions", handlers.HandleListJobFormQuestions())
	r.POST("/settings/job-form-questions", handlers.HandleCreateJobFormQuestion())
	r.PUT("/settings/job-form-questions/:id", handlers.HandleUpdateJobFormQuestion())
	r.DELETE("/settings/job-form-questions/:id", handlers.HandleDeleteJobFormQuestion())

	r.GET("/settings/job-form-titles", handlers.HandleGetJobFormSectionTitles())
	r.POST("/settings/job-form-titles", handlers.HandleUpsertJobFormSectionTitles())

	return r
}

func doRequest(r *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		b, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(b)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func decodeBody(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var out map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &out)
	require.NoError(t, err)
	return out
}

func seedQuestion(t *testing.T, gormDB *gorm.DB, stage, category, text string, sortOrder int) models.JobFormQuestion {
	t.Helper()
	q := models.JobFormQuestion{
		Stage:     stage,
		Category:  category,
		Text:      text,
		SortOrder: sortOrder,
	}
	require.NoError(t, gormDB.Create(&q).Error)
	return q
}

// ── HandleListJobFormQuestions ──────────────────────────────────

func TestHandleListJobFormQuestions(t *testing.T) {
	t.Run("returns empty list when no questions exist", func(t *testing.T) {
		database, _ := setupTestDB(t)
		r := setupRouter(t, database)

		w := doRequest(r, http.MethodGet, "/settings/job-form-questions", nil)

		assert.Equal(t, http.StatusOK, w.Code)
		body := decodeBody(t, w)
		data, ok := body["data"].([]interface{})
		require.True(t, ok)
		assert.Len(t, data, 0)
	})

	t.Run("filters by stage and category and orders by sort_order", func(t *testing.T) {
		database, gormDB := setupTestDB(t)
		r := setupRouter(t, database)

		seedQuestion(t, gormDB, "Cutting", "fabricator_checklist", "Second question", 2)
		seedQuestion(t, gormDB, "Cutting", "fabricator_checklist", "First question", 1)
		seedQuestion(t, gormDB, "Template", "fabricator_checklist", "Other stage question", 1)
		seedQuestion(t, gormDB, "Cutting", "other_category", "Different category", 1)

		w := doRequest(r, http.MethodGet, "/settings/job-form-questions?stage=Cutting&category=fabricator_checklist", nil)

		assert.Equal(t, http.StatusOK, w.Code)
		body := decodeBody(t, w)
		data, ok := body["data"].([]interface{})
		require.True(t, ok)
		require.Len(t, data, 2)

		first := data[0].(map[string]interface{})
		second := data[1].(map[string]interface{})
		assert.Equal(t, "First question", first["text"])
		assert.Equal(t, "Second question", second["text"])
	})

	t.Run("returns all questions when no filters provided", func(t *testing.T) {
		database, gormDB := setupTestDB(t)
		r := setupRouter(t, database)

		seedQuestion(t, gormDB, "Cutting", "fabricator_checklist", "Q1", 1)
		seedQuestion(t, gormDB, "Template", "fabricator_checklist", "Q2", 1)

		w := doRequest(r, http.MethodGet, "/settings/job-form-questions", nil)

		assert.Equal(t, http.StatusOK, w.Code)
		body := decodeBody(t, w)
		data, ok := body["data"].([]interface{})
		require.True(t, ok)
		assert.Len(t, data, 2)
	})
}

// ── HandleCreateJobFormQuestion ─────────────────────────────────

func TestHandleCreateJobFormQuestion(t *testing.T) {
	t.Run("creates a question with valid input", func(t *testing.T) {
		database, gormDB := setupTestDB(t)
		r := setupRouter(t, database)

		reqBody := map[string]interface{}{
			"stage":      "Cutting",
			"category":   "fabricator_checklist",
			"text":       "Parking Issues?",
			"sort_order": 3,
		}

		w := doRequest(r, http.MethodPost, "/settings/job-form-questions", reqBody)

		assert.Equal(t, http.StatusCreated, w.Code)
		body := decodeBody(t, w)
		assert.Equal(t, "Question created successfully", body["message"])

		data, ok := body["data"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "Parking Issues?", data["text"])
		assert.Equal(t, "Cutting", data["stage"])
		assert.Equal(t, "fabricator_checklist", data["category"])

		var count int64
		gormDB.Model(&models.JobFormQuestion{}).Count(&count)
		assert.Equal(t, int64(1), count)
	})

	t.Run("trims whitespace from stage, category and text", func(t *testing.T) {
		database, _ := setupTestDB(t)
		r := setupRouter(t, database)

		reqBody := map[string]interface{}{
			"stage":      "  Cutting  ",
			"category":   "  fabricator_checklist  ",
			"text":       "  Ceiling Height issues?  ",
			"sort_order": 0,
		}

		w := doRequest(r, http.MethodPost, "/settings/job-form-questions", reqBody)

		assert.Equal(t, http.StatusCreated, w.Code)
		body := decodeBody(t, w)
		data := body["data"].(map[string]interface{})
		assert.Equal(t, "Cutting", data["stage"])
		assert.Equal(t, "fabricator_checklist", data["category"])
		assert.Equal(t, "Ceiling Height issues?", data["text"])
	})

	t.Run("rejects missing stage", func(t *testing.T) {
		database, _ := setupTestDB(t)
		r := setupRouter(t, database)

		reqBody := map[string]interface{}{
			"category": "fabricator_checklist",
			"text":     "Some question",
		}

		w := doRequest(r, http.MethodPost, "/settings/job-form-questions", reqBody)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		body := decodeBody(t, w)
		assert.Equal(t, "stage, category and text are required", body["error"])
	})

	t.Run("rejects missing category", func(t *testing.T) {
		database, _ := setupTestDB(t)
		r := setupRouter(t, database)

		reqBody := map[string]interface{}{
			"stage": "Cutting",
			"text":  "Some question",
		}

		w := doRequest(r, http.MethodPost, "/settings/job-form-questions", reqBody)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("rejects missing text", func(t *testing.T) {
		database, _ := setupTestDB(t)
		r := setupRouter(t, database)

		reqBody := map[string]interface{}{
			"stage":    "Cutting",
			"category": "fabricator_checklist",
			"text":     "   ",
		}

		w := doRequest(r, http.MethodPost, "/settings/job-form-questions", reqBody)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("rejects malformed JSON body", func(t *testing.T) {
		database, _ := setupTestDB(t)
		r := setupRouter(t, database)

		req := httptest.NewRequest(http.MethodPost, "/settings/job-form-questions", bytes.NewBufferString("{invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// ── HandleUpdateJobFormQuestion ─────────────────────────────────

func TestHandleUpdateJobFormQuestion(t *testing.T) {
	t.Run("updates text and sort_order of existing question", func(t *testing.T) {
		database, gormDB := setupTestDB(t)
		r := setupRouter(t, database)

		q := seedQuestion(t, gormDB, "Cutting", "fabricator_checklist", "Old text", 1)

		reqBody := map[string]interface{}{
			"text":       "Updated text",
			"sort_order": 5,
		}

		w := doRequest(r, http.MethodPut, "/settings/job-form-questions/"+strconv.Itoa(int(q.ID)), reqBody)

		assert.Equal(t, http.StatusOK, w.Code)
		body := decodeBody(t, w)
		data := body["data"].(map[string]interface{})
		assert.Equal(t, "Updated text", data["text"])
		assert.Equal(t, float64(5), data["sort_order"])

		var updated models.JobFormQuestion
		require.NoError(t, gormDB.First(&updated, q.ID).Error)
		assert.Equal(t, "Updated text", updated.Text)
	})

	t.Run("rejects malformed JSON body", func(t *testing.T) {
	database, gormDB := setupTestDB(t)
	r := setupRouter(t, database)

	q := seedQuestion(t, gormDB, "Cutting", "fabricator_checklist", "Old text", 1)

	req := httptest.NewRequest(http.MethodPut, "/settings/job-form-questions/"+strconv.Itoa(int(q.ID)), bytes.NewBufferString("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
})

	t.Run("returns 404 for non-existent id", func(t *testing.T) {
		database, _ := setupTestDB(t)
		r := setupRouter(t, database)

		reqBody := map[string]interface{}{"text": "Doesn't matter", "sort_order": 0}
		w := doRequest(r, http.MethodPut, "/settings/job-form-questions/9999", reqBody)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("returns 400 for non-numeric id", func(t *testing.T) {
		database, _ := setupTestDB(t)
		r := setupRouter(t, database)

		reqBody := map[string]interface{}{"text": "Doesn't matter", "sort_order": 0}
		w := doRequest(r, http.MethodPut, "/settings/job-form-questions/abc", reqBody)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("rejects empty text on update", func(t *testing.T) {
		database, gormDB := setupTestDB(t)
		r := setupRouter(t, database)

		q := seedQuestion(t, gormDB, "Cutting", "fabricator_checklist", "Old text", 1)

		reqBody := map[string]interface{}{"text": "   ", "sort_order": 1}
		w := doRequest(r, http.MethodPut, "/settings/job-form-questions/"+strconv.Itoa(int(q.ID)), reqBody)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// ── HandleDeleteJobFormQuestion ─────────────────────────────────

func TestHandleDeleteJobFormQuestion(t *testing.T) {
	t.Run("deletes an existing question", func(t *testing.T) {
		database, gormDB := setupTestDB(t)
		r := setupRouter(t, database)

		q := seedQuestion(t, gormDB, "Cutting", "fabricator_checklist", "To be deleted", 1)

		w := doRequest(r, http.MethodDelete, "/settings/job-form-questions/"+strconv.Itoa(int(q.ID)), nil)

		assert.Equal(t, http.StatusOK, w.Code)
		body := decodeBody(t, w)
		assert.Equal(t, "Question deleted successfully", body["message"])

		var count int64
		gormDB.Model(&models.JobFormQuestion{}).Where("id = ?", q.ID).Count(&count)
		assert.Equal(t, int64(0), count)
	})

	t.Run("returns 404 for non-existent id", func(t *testing.T) {
		database, _ := setupTestDB(t)
		r := setupRouter(t, database)

		w := doRequest(r, http.MethodDelete, "/settings/job-form-questions/9999", nil)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("returns 400 for non-numeric id", func(t *testing.T) {
		database, _ := setupTestDB(t)
		r := setupRouter(t, database)

		w := doRequest(r, http.MethodDelete, "/settings/job-form-questions/abc", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("does not affect other questions", func(t *testing.T) {
		database, gormDB := setupTestDB(t)
		r := setupRouter(t, database)

		q1 := seedQuestion(t, gormDB, "Cutting", "fabricator_checklist", "Keep me", 1)
		q2 := seedQuestion(t, gormDB, "Cutting", "fabricator_checklist", "Delete me", 2)

		w := doRequest(r, http.MethodDelete, "/settings/job-form-questions/"+strconv.Itoa(int(q2.ID)), nil)
		assert.Equal(t, http.StatusOK, w.Code)

		var remaining models.JobFormQuestion
		require.NoError(t, gormDB.First(&remaining, q1.ID).Error)
		assert.Equal(t, "Keep me", remaining.Text)
	})
}

// ── HandleGetJobFormSectionTitles ────────────────────────────────

func TestHandleGetJobFormSectionTitles(t *testing.T) {
	t.Run("returns defaults when no titles saved for stage", func(t *testing.T) {
		database, _ := setupTestDB(t)
		r := setupRouter(t, database)

		w := doRequest(r, http.MethodGet, "/settings/job-form-titles?stage=Cutting", nil)

		assert.Equal(t, http.StatusOK, w.Code)
		body := decodeBody(t, w)
		data := body["data"].(map[string]interface{})
		assert.Equal(t, "Cutting", data["stage"])
		assert.Equal(t, "MATERIALS", data["materials_title"])
		assert.Equal(t, "SERVICES", data["services_title"])
		assert.Equal(t, "Fabricator's checklist After Fabrication Complete (circle one)", data["fabricator_checklist_title"])
	})

	t.Run("returns saved titles for stage", func(t *testing.T) {
		database, gormDB := setupTestDB(t)
		r := setupRouter(t, database)

		titles := models.JobFormSectionTitles{
			Stage:                    "Cutting",
			MaterialsTitle:           "Custom Materials",
			ServicesTitle:            "Custom Services",
			FabricatorChecklistTitle: "Custom Checklist",
		}
		require.NoError(t, gormDB.Create(&titles).Error)

		w := doRequest(r, http.MethodGet, "/settings/job-form-titles?stage=Cutting", nil)

		assert.Equal(t, http.StatusOK, w.Code)
		body := decodeBody(t, w)
		data := body["data"].(map[string]interface{})
		assert.Equal(t, "Custom Materials", data["materials_title"])
		assert.Equal(t, "Custom Services", data["services_title"])
		assert.Equal(t, "Custom Checklist", data["fabricator_checklist_title"])
	})

	t.Run("rejects missing stage query param", func(t *testing.T) {
		database, _ := setupTestDB(t)
		r := setupRouter(t, database)

		w := doRequest(r, http.MethodGet, "/settings/job-form-titles", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// ── HandleUpsertJobFormSectionTitles ─────────────────────────────

func TestHandleUpsertJobFormSectionTitles(t *testing.T) {
	t.Run("creates new titles when none exist for stage", func(t *testing.T) {
		database, gormDB := setupTestDB(t)
		r := setupRouter(t, database)

		reqBody := map[string]interface{}{
			"stage":                      "Cutting",
			"materials_title":            "MATERIALS X",
			"services_title":             "SERVICES X",
			"fabricator_checklist_title": "Checklist X",
		}

		w := doRequest(r, http.MethodPost, "/settings/job-form-titles", reqBody)

		assert.Equal(t, http.StatusCreated, w.Code)
		body := decodeBody(t, w)
		assert.Equal(t, "Titles created successfully", body["message"])

		var saved models.JobFormSectionTitles
		require.NoError(t, gormDB.Where("stage = ?", "Cutting").First(&saved).Error)
		assert.Equal(t, "MATERIALS X", saved.MaterialsTitle)
	})

	t.Run("rejects malformed JSON body", func(t *testing.T) {
	database, _ := setupTestDB(t)
	r := setupRouter(t, database)

	req := httptest.NewRequest(http.MethodPost, "/settings/job-form-titles", bytes.NewBufferString("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
})
	t.Run("updates existing titles for stage", func(t *testing.T) {
		database, gormDB := setupTestDB(t)
		r := setupRouter(t, database)

		existing := models.JobFormSectionTitles{
			Stage:                    "Cutting",
			MaterialsTitle:           "Old Materials",
			ServicesTitle:            "Old Services",
			FabricatorChecklistTitle: "Old Checklist",
		}
		require.NoError(t, gormDB.Create(&existing).Error)

		reqBody := map[string]interface{}{
			"stage":                      "Cutting",
			"materials_title":            "New Materials",
			"services_title":             "New Services",
			"fabricator_checklist_title": "New Checklist",
		}

		w := doRequest(r, http.MethodPost, "/settings/job-form-titles", reqBody)

		assert.Equal(t, http.StatusOK, w.Code)
		body := decodeBody(t, w)
		assert.Equal(t, "Titles updated successfully", body["message"])

		var count int64
		gormDB.Model(&models.JobFormSectionTitles{}).Where("stage = ?", "Cutting").Count(&count)
		assert.Equal(t, int64(1), count) // no duplicate row created

		var updated models.JobFormSectionTitles
		require.NoError(t, gormDB.Where("stage = ?", "Cutting").First(&updated).Error)
		assert.Equal(t, "New Materials", updated.MaterialsTitle)
		assert.Equal(t, "New Services", updated.ServicesTitle)
		assert.Equal(t, "New Checklist", updated.FabricatorChecklistTitle)
	})

	t.Run("rejects missing stage", func(t *testing.T) {
		database, _ := setupTestDB(t)
		r := setupRouter(t, database)

		reqBody := map[string]interface{}{
			"materials_title": "MATERIALS",
		}

		w := doRequest(r, http.MethodPost, "/settings/job-form-titles", reqBody)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("keeps titles for different stages independent", func(t *testing.T) {
		database, gormDB := setupTestDB(t)
		r := setupRouter(t, database)

		doRequest(r, http.MethodPost, "/settings/job-form-titles", map[string]interface{}{
			"stage": "Cutting", "materials_title": "Cutting Materials",
			"services_title": "S", "fabricator_checklist_title": "C",
		})
		doRequest(r, http.MethodPost, "/settings/job-form-titles", map[string]interface{}{
			"stage": "Template", "materials_title": "Template Materials",
			"services_title": "S", "fabricator_checklist_title": "C",
		})

		var count int64
		gormDB.Model(&models.JobFormSectionTitles{}).Count(&count)
		assert.Equal(t, int64(2), count)
	})
}