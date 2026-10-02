package test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"inventory/handlers"
	"inventory/mocks"
	"inventory/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func validCreateSessionInput() map[string]interface{} {
	return map[string]interface{}{
		"title":           "Corporate Inventory",
		"start_date":      time.Now().Format(time.RFC3339Nano),
		"location":        "Corporate Warehouse",
		"session_type":    "slab",
		"internal_notes":  "notes",
		"production_team": []string{"Alice", "Bob"},
	}
}

func toJSON(fields map[string]interface{}) *bytes.Buffer {
	b, _ := json.Marshal(fields)
	return bytes.NewBuffer(b)
}

func newRouter(mockDB *mocks.MockDatabase) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(injectDB(mockDB))
	router.GET("/audit/sessions", handlers.HandleAuditSessionsCollection())
	router.POST("/audit/sessions", handlers.HandleAuditSessionsCollection())
	router.DELETE("/audit/sessions/:id", handlers.HandleAuditSessionItem())
	router.GET("/audit/meta", handlers.HandleAuditMeta())
	return router
}

// ── Pure helper functions (no DB / HTTP needed) ─────────────────────────────
// These exercise joinTeam, parseLockedMaterialIDs, joinMaterialIDs, and
// computeAuditCounts indirectly through handler behavior since they are
// unexported; direct coverage comes from the scenarios below that rely on
// their output (e.g. production_team round-tripping, SystemCount locking,
// found/missing/unidentified counts in list and detail responses).

// ── HandleAuditMeta ──────────────────────────────────────────────────────────

func TestHandleAuditMeta(t *testing.T) {
	t.Run("default filter options", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Model", mock.AnythingOfType("*models.Materials")).Return(mockDB)
		mockDB.On("Model", mock.AnythingOfType("*models.Materialchunks")).Return(mockDB)
		mockDB.On("Where", mock.Anything, mock.Anything).Return(mockDB)
		mockDB.On("Distinct", mock.Anything).Return(mockDB)
		mockDB.On("Pluck", "category", mock.AnythingOfType("*[]string")).
			Return(nil).Run(func(a mock.Arguments) {
			*a.Get(1).(*[]string) = []string{"Stone"}
		})
		mockDB.On("Pluck", "type", mock.AnythingOfType("*[]string")).
			Return(nil).Run(func(a mock.Arguments) {
			*a.Get(1).(*[]string) = []string{"Slab"}
		})
		mockDB.On("Pluck", "sub_category", mock.AnythingOfType("*[]string")).
			Return(nil).Run(func(a mock.Arguments) {
			*a.Get(1).(*[]string) = []string{"Granite"}
		})
		mockDB.On("Pluck", "present_location", mock.AnythingOfType("*[]string")).
			Return(nil).Run(func(a mock.Arguments) {
			*a.Get(1).(*[]string) = []string{"Warehouse A"}
		})

		router := newRouter(mockDB)
		req, _ := http.NewRequest("GET", "/audit/meta", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Warehouse A")
	})

	t.Run("material search - invalid session_type", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		router := newRouter(mockDB)
		req, _ := http.NewRequest("GET", "/audit/meta?search=marble", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "session_type must be")
	})

	t.Run("material search - valid", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Model", mock.AnythingOfType("*models.Materials")).Return(mockDB)
		mockDB.On("Joins", mock.Anything, mock.Anything).Return(mockDB)
		mockDB.On("Where", mock.Anything, mock.Anything).Return(mockDB)
		mockDB.On("Distinct", mock.Anything).Return(mockDB)
		mockDB.On("Order", mock.Anything).Return(mockDB)
		mockDB.On("Limit", 20).Return(mockDB)
		mockDB.On("Pluck", "materials.name", mock.AnythingOfType("*[]string")).
			Return(nil).Run(func(a mock.Arguments) {
			*a.Get(1).(*[]string) = []string{"Carrara Marble"}
		})

		router := newRouter(mockDB)
		req, _ := http.NewRequest("GET", "/audit/meta?search=marble&session_type=slab", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Carrara Marble")
	})
}

// ── GET /audit/sessions (list) ──────────────────────────────────────────────

func TestListAuditSessions(t *testing.T) {
	t.Run("fetch active sessions successfully", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Where", "status = ?", "active").Return(mockDB)
		mockDB.On("Order", "start_date DESC").Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]models.AuditSession"), mock.Anything).
			Return(nil).
			Run(func(a mock.Arguments) {
				sl := a.Get(0).(*[]models.AuditSession)
				*sl = []models.AuditSession{{ID: 1, Title: "Corporate Inventory", Status: "active", SystemCount: 10}}
			}).Once()
		// attachCounts fans out one Where().Find() per session
		mockDB.On("Where", "audit_session_id = ?", uint(1)).Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]models.AuditLine"), mock.Anything).
			Return(nil).
			Run(func(a mock.Arguments) {
				ll := a.Get(0).(*[]models.AuditLine)
				*ll = []models.AuditLine{{ScanStatus: "reconciled"}, {ScanStatus: "missing"}}
			}).Once()

		router := newRouter(mockDB)
		req, _ := http.NewRequest("GET", "/audit/sessions?status=active", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Corporate Inventory")
		mockDB.AssertExpectations(t)
	})

	t.Run("invalid status", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		router := newRouter(mockDB)
		req, _ := http.NewRequest("GET", "/audit/sessions?status=bogus", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "status must be")
	})

	t.Run("database error", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Where", "status = ?", "active").Return(mockDB)
		mockDB.On("Order", "start_date DESC").Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]models.AuditSession"), mock.Anything).
			Return(errors.New("db error"))

		router := newRouter(mockDB)
		req, _ := http.NewRequest("GET", "/audit/sessions", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to fetch sessions")
	})

	t.Run("completed sessions - orders by completed_at", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Where", "status = ?", "completed").Return(mockDB)
		mockDB.On("Order", "completed_at DESC").Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]models.AuditSession"), mock.Anything).Return(nil)

		router := newRouter(mockDB)
		req, _ := http.NewRequest("GET", "/audit/sessions?status=completed", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// ── GET /audit/sessions?id=<id> (detail) ────────────────────────────────────

func TestGetAuditSession(t *testing.T) {
	t.Run("get session successfully", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Preload", "Lines", mock.Anything).Return(mockDB)
		mockDB.On("First", mock.AnythingOfType("*models.AuditSession"), "5").
			Return(nil).
			Run(func(a mock.Arguments) {
				s := a.Get(0).(*models.AuditSession)
				s.ID = 5
				s.Title = "Corporate Inventory"
				s.SystemCount = 4
				s.Lines = []models.AuditLine{{ScanStatus: "reconciled"}, {ScanStatus: "unidentified"}}
			}).Once()

		router := newRouter(mockDB)
		req, _ := http.NewRequest("GET", "/audit/sessions?id=5", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Corporate Inventory")
		mockDB.AssertExpectations(t)
	})

	t.Run("session not found", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Preload", "Lines", mock.Anything).Return(mockDB)
		mockDB.On("First", mock.AnythingOfType("*models.AuditSession"), "999").
			Return(gorm.ErrRecordNotFound)

		router := newRouter(mockDB)
		req, _ := http.NewRequest("GET", "/audit/sessions?id=999", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Audit session not found")
	})

	t.Run("database error", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Preload", "Lines", mock.Anything).Return(mockDB)
		mockDB.On("First", mock.AnythingOfType("*models.AuditSession"), "5").
			Return(errors.New("db error"))

		router := newRouter(mockDB)
		req, _ := http.NewRequest("GET", "/audit/sessions?id=5", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// ── POST /audit/sessions (create, id absent or -1) ──────────────────────────

func TestCreateAuditSession(t *testing.T) {
	t.Run("valid request - creates session", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Create", mock.AnythingOfType("*models.AuditSession")).Return(nil)

		router := newRouter(mockDB)
		req, _ := http.NewRequest("POST", "/audit/sessions", toJSON(validCreateSessionInput()))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var resp models.AuditSession
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "Corporate Inventory", resp.Title)
		assert.Equal(t, "Alice, Bob", resp.ProductionTeam)
		mockDB.AssertExpectations(t)
	})

	t.Run("explicit id=-1 still creates", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Create", mock.AnythingOfType("*models.AuditSession")).Return(nil)

		router := newRouter(mockDB)
		req, _ := http.NewRequest("POST", "/audit/sessions?id=-1", toJSON(validCreateSessionInput()))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("invalid JSON body", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		router := newRouter(mockDB)
		req, _ := http.NewRequest("POST", "/audit/sessions", bytes.NewBuffer([]byte(`{bad}`)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request body")
	})

	t.Run("missing required fields", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		router := newRouter(mockDB)
		input := validCreateSessionInput()
		delete(input, "location")
		req, _ := http.NewRequest("POST", "/audit/sessions", toJSON(input))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "required")
	})

	t.Run("invalid session_type", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		router := newRouter(mockDB)
		input := validCreateSessionInput()
		input["session_type"] = "bogus"
		req, _ := http.NewRequest("POST", "/audit/sessions", toJSON(input))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "session_type must be")
	})

	t.Run("invalid start_date format", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		router := newRouter(mockDB)
		input := validCreateSessionInput()
		input["start_date"] = "not-a-date"
		req, _ := http.NewRequest("POST", "/audit/sessions", toJSON(input))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid start_date format")
	})

	t.Run("database error on create", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Create", mock.AnythingOfType("*models.AuditSession")).
			Return(errors.New("connection refused"))

		router := newRouter(mockDB)
		req, _ := http.NewRequest("POST", "/audit/sessions", toJSON(validCreateSessionInput()))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to create audit session")
	})
}

// ── POST /audit/sessions?id=<id> (edit / complete) ──────────────────────────

func TestUpdateAuditSession(t *testing.T) {
	t.Run("edit fields successfully", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("First", mock.AnythingOfType("*models.AuditSession"), "7").
			Return(nil).
			Run(func(a mock.Arguments) {
				s := a.Get(0).(*models.AuditSession)
				*s = models.AuditSession{ID: 7, Status: "active", SessionType: "slab"}
			}).Once()
		mockDB.On("Save", mock.AnythingOfType("*models.AuditSession")).Return(nil).Once()

		router := newRouter(mockDB)
		req, _ := http.NewRequest("POST", "/audit/sessions?id=7", toJSON(validCreateSessionInput()))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("complete session successfully", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("First", mock.AnythingOfType("*models.AuditSession"), "7").
			Return(nil).
			Run(func(a mock.Arguments) {
				s := a.Get(0).(*models.AuditSession)
				*s = models.AuditSession{ID: 7, Status: "active"}
			}).Once()
		mockDB.On("Save", mock.AnythingOfType("*models.AuditSession")).Return(nil).Once()

		router := newRouter(mockDB)
		body := map[string]interface{}{"action": "complete", "completed_by": "Vishal"}
		req, _ := http.NewRequest("POST", "/audit/sessions?id=7", toJSON(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp models.AuditSession
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "completed", resp.Status)
		assert.Equal(t, "Vishal", resp.CompletedBy)
		mockDB.AssertExpectations(t)
	})

	t.Run("complete - session already inactive", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("First", mock.AnythingOfType("*models.AuditSession"), "7").
			Return(nil).
			Run(func(a mock.Arguments) {
				s := a.Get(0).(*models.AuditSession)
				*s = models.AuditSession{ID: 7, Status: "completed"}
			}).Once()

		router := newRouter(mockDB)
		body := map[string]interface{}{"action": "complete"}
		req, _ := http.NewRequest("POST", "/audit/sessions?id=7", toJSON(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Session is not active")
	})

	t.Run("edit - session not active", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("First", mock.AnythingOfType("*models.AuditSession"), "7").
			Return(nil).
			Run(func(a mock.Arguments) {
				s := a.Get(0).(*models.AuditSession)
				*s = models.AuditSession{ID: 7, Status: "completed"}
			}).Once()

		router := newRouter(mockDB)
		req, _ := http.NewRequest("POST", "/audit/sessions?id=7", toJSON(validCreateSessionInput()))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Only active sessions can be edited")
	})

	t.Run("session not found", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("First", mock.AnythingOfType("*models.AuditSession"), "999").
			Return(gorm.ErrRecordNotFound)

		router := newRouter(mockDB)
		req, _ := http.NewRequest("POST", "/audit/sessions?id=999", toJSON(validCreateSessionInput()))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("invalid JSON body", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("First", mock.AnythingOfType("*models.AuditSession"), "7").
			Return(nil).
			Run(func(a mock.Arguments) {
				s := a.Get(0).(*models.AuditSession)
				*s = models.AuditSession{ID: 7, Status: "active"}
			}).Once()

		router := newRouter(mockDB)
		req, _ := http.NewRequest("POST", "/audit/sessions?id=7", bytes.NewBuffer([]byte(`{bad}`)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid session_type on edit", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("First", mock.AnythingOfType("*models.AuditSession"), "7").
			Return(nil).
			Run(func(a mock.Arguments) {
				s := a.Get(0).(*models.AuditSession)
				*s = models.AuditSession{ID: 7, Status: "active", SessionType: "slab"}
			}).Once()

		router := newRouter(mockDB)
		input := validCreateSessionInput()
		input["session_type"] = "bogus"
		req, _ := http.NewRequest("POST", "/audit/sessions?id=7", toJSON(input))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid start_date on edit", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("First", mock.AnythingOfType("*models.AuditSession"), "7").
			Return(nil).
			Run(func(a mock.Arguments) {
				s := a.Get(0).(*models.AuditSession)
				*s = models.AuditSession{ID: 7, Status: "active", SessionType: "slab"}
			}).Once()

		router := newRouter(mockDB)
		input := validCreateSessionInput()
		input["start_date"] = "bad-date"
		req, _ := http.NewRequest("POST", "/audit/sessions?id=7", toJSON(input))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("cannot change session_type after materials locked", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("First", mock.AnythingOfType("*models.AuditSession"), "7").
			Return(nil).
			Run(func(a mock.Arguments) {
				s := a.Get(0).(*models.AuditSession)
				*s = models.AuditSession{ID: 7, Status: "active", SessionType: "slab", LockedMaterialIDs: "1,2"}
			}).Once()

		router := newRouter(mockDB)
		input := validCreateSessionInput()
		input["session_type"] = "non_slab"
		req, _ := http.NewRequest("POST", "/audit/sessions?id=7", toJSON(input))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Cannot change session type")
	})

	t.Run("database error on save", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("First", mock.AnythingOfType("*models.AuditSession"), "7").
			Return(nil).
			Run(func(a mock.Arguments) {
				s := a.Get(0).(*models.AuditSession)
				*s = models.AuditSession{ID: 7, Status: "active", SessionType: "slab"}
			}).Once()
		mockDB.On("Save", mock.AnythingOfType("*models.AuditSession")).
			Return(errors.New("db error")).Once()

		router := newRouter(mockDB)
		req, _ := http.NewRequest("POST", "/audit/sessions?id=7", toJSON(validCreateSessionInput()))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// ── DELETE /audit/sessions/:id ───────────────────────────────────────────────

func TestDeleteAuditSession(t *testing.T) {
	t.Run("delete successfully", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("First", mock.AnythingOfType("*models.AuditSession"), "3").
			Return(nil).
			Run(func(a mock.Arguments) {
				s := a.Get(0).(*models.AuditSession)
				s.ID = 3
			}).Once()
		mockDB.On("Where", "audit_session_id = ?", uint(3)).Return(mockDB)
		mockDB.On("Delete", mock.AnythingOfType("*models.AuditLine"), mock.Anything).Return(nil).Once()
		mockDB.On("Delete", mock.AnythingOfType("*models.AuditSession"), mock.Anything).Return(nil).Once()

		router := newRouter(mockDB)
		req, _ := http.NewRequest("DELETE", "/audit/sessions/3", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Audit session deleted")
		mockDB.AssertExpectations(t)
	})

	t.Run("session not found", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("First", mock.AnythingOfType("*models.AuditSession"), "999").
			Return(gorm.ErrRecordNotFound)

		router := newRouter(mockDB)
		req, _ := http.NewRequest("DELETE", "/audit/sessions/999", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("database error on fetch", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("First", mock.AnythingOfType("*models.AuditSession"), "3").
			Return(errors.New("db error"))

		router := newRouter(mockDB)
		req, _ := http.NewRequest("DELETE", "/audit/sessions/3", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("failure deleting lines", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("First", mock.AnythingOfType("*models.AuditSession"), "3").
			Return(nil).
			Run(func(a mock.Arguments) {
				s := a.Get(0).(*models.AuditSession)
				s.ID = 3
			}).Once()
		mockDB.On("Where", "audit_session_id = ?", uint(3)).Return(mockDB)
		mockDB.On("Delete", mock.AnythingOfType("*models.AuditLine"), mock.Anything).
			Return(errors.New("db error")).Once()

		router := newRouter(mockDB)
		req, _ := http.NewRequest("DELETE", "/audit/sessions/3", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to delete audit lines")
	})

	t.Run("failure deleting session", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("First", mock.AnythingOfType("*models.AuditSession"), "3").
			Return(nil).
			Run(func(a mock.Arguments) {
				s := a.Get(0).(*models.AuditSession)
				s.ID = 3
			}).Once()
		mockDB.On("Where", "audit_session_id = ?", uint(3)).Return(mockDB)
		mockDB.On("Delete", mock.AnythingOfType("*models.AuditLine"), mock.Anything).Return(nil).Once()
		mockDB.On("Delete", mock.AnythingOfType("*models.AuditSession"), mock.Anything).
			Return(errors.New("db error")).Once()

		router := newRouter(mockDB)
		req, _ := http.NewRequest("DELETE", "/audit/sessions/3", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to delete audit session")
	})
}

// ── GET /audit/sessions?id=<id>&lines=true (searchAuditCandidates) ──────────
//
// NOTE: buildCandidateQuery chains Model -> Joins -> a variable number of
// Where clauses depending on which filters are set. These tests use permissive
// mock.Anything stubs for the chain methods (Model/Joins/Where/Order/Select/
// Offset/Limit/Distinct all just return mockDB) so the exact filter-branch
// path taken doesn't need to be hand-enumerated. Terminal calls (Count,
// Pluck, Scan, Save) are asserted precisely.
func stubChain(mockDB *mocks.MockDatabase) {
	mockDB.On("Model", mock.Anything).Return(mockDB).Maybe()
	mockDB.On("Joins", mock.Anything, mock.Anything).Return(mockDB).Maybe()
	mockDB.On("Where", mock.Anything, mock.Anything).Return(mockDB).Maybe()
	mockDB.On("Where", mock.Anything, mock.Anything, mock.Anything).Return(mockDB).Maybe()
	mockDB.On("Order", mock.Anything).Return(mockDB).Maybe()
	mockDB.On("Select", mock.Anything, mock.Anything).Return(mockDB).Maybe()
	mockDB.On("Offset", mock.Anything).Return(mockDB).Maybe()
	mockDB.On("Limit", mock.Anything).Return(mockDB).Maybe()
	mockDB.On("Distinct", mock.Anything).Return(mockDB).Maybe()
}

func TestSearchAuditCandidates(t *testing.T) {
	t.Run("search with results, no new materials to lock", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("First", mock.AnythingOfType("*models.AuditSession"), "9").
			Return(nil).
			Run(func(a mock.Arguments) {
				s := a.Get(0).(*models.AuditSession)
				*s = models.AuditSession{ID: 9, SessionType: "slab", LockedMaterialIDs: "1"}
			}).Once()
		stubChain(mockDB)
		mockDB.On("Pluck", "serialnum", mock.AnythingOfType("*[]string")).Return(nil).Maybe()
		mockDB.On("Count", mock.AnythingOfType("*int64")).
			Return(nil).
			Run(func(a mock.Arguments) { *a.Get(0).(*int64) = 2 }).Maybe()
		mockDB.On("Scan", mock.Anything).Return(nil).Maybe()
		// lockNewMaterialsIntoSystemCount's Pluck(materialchunks.material_id) returns
		// only the already-locked material (id 1), so no Save call should occur.
		mockDB.On("Pluck", "materialchunks.material_id", mock.AnythingOfType("*[]uint")).
			Return(nil).
			Run(func(a mock.Arguments) {
				*a.Get(1).(*[]uint) = []uint{1}
			}).Maybe()

		router := newRouter(mockDB)
		req, _ := http.NewRequest("GET", "/audit/sessions?id=9&lines=true&search=marble&page=1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("session not found", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("First", mock.AnythingOfType("*models.AuditSession"), "999").
			Return(gorm.ErrRecordNotFound)

		router := newRouter(mockDB)
		req, _ := http.NewRequest("GET", "/audit/sessions?id=999&lines=true", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

// ── POST /audit/sessions?id=<id>&lines=true (addAuditLines) ─────────────────

func TestAddAuditLines(t *testing.T) {
	t.Run("spare item add - recognized in system", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("First", mock.AnythingOfType("*models.AuditSession"), "9").
			Return(nil).
			Run(func(a mock.Arguments) {
				s := a.Get(0).(*models.AuditSession)
				*s = models.AuditSession{ID: 9, SessionType: "slab", Status: "active"}
			}).Once()
		// duplicate-serial check
		mockDB.On("Where", "audit_session_id = ? AND serialnum = ?", uint(9), "SN-100").
			Return(mockDB).Once()
		mockDB.On("First", mock.AnythingOfType("*models.AuditLine"), mock.Anything).
			Return(gorm.ErrRecordNotFound).Once()
		// chunk enrichment lookup
		mockDB.On("Where", "serialnum = ?", "SN-100").Return(mockDB).Once()
		mockDB.On("First", mock.AnythingOfType("*models.Materialchunks"), mock.Anything).
			Return(gorm.ErrRecordNotFound).Once()
		mockDB.On("Create", mock.AnythingOfType("*models.AuditLine")).Return(nil).Once()

		router := newRouter(mockDB)
		body := map[string]interface{}{"serial": "SN-100", "note": "spare found on floor"}
		req, _ := http.NewRequest("POST", "/audit/sessions?id=9&lines=true", toJSON(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("spare item - duplicate serial rejected", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("First", mock.AnythingOfType("*models.AuditSession"), "9").
			Return(nil).
			Run(func(a mock.Arguments) {
				s := a.Get(0).(*models.AuditSession)
				*s = models.AuditSession{ID: 9, SessionType: "slab", Status: "active"}
			}).Once()
		mockDB.On("Where", "audit_session_id = ? AND serialnum = ?", uint(9), "SN-100").
			Return(mockDB).Once()
		mockDB.On("First", mock.AnythingOfType("*models.AuditLine"), mock.Anything).
			Return(nil).Once()

		router := newRouter(mockDB)
		body := map[string]interface{}{"serial": "SN-100"}
		req, _ := http.NewRequest("POST", "/audit/sessions?id=9&lines=true", toJSON(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), "already been added")
	})

	t.Run("session not active", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("First", mock.AnythingOfType("*models.AuditSession"), "9").
			Return(nil).
			Run(func(a mock.Arguments) {
				s := a.Get(0).(*models.AuditSession)
				*s = models.AuditSession{ID: 9, Status: "completed"}
			}).Once()

		router := newRouter(mockDB)
		body := map[string]interface{}{"serial": "SN-100"}
		req, _ := http.NewRequest("POST", "/audit/sessions?id=9&lines=true", toJSON(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Session is not active")
	})

	t.Run("bulk filtered add - no new matches", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("First", mock.AnythingOfType("*models.AuditSession"), "9").
			Return(nil).
			Run(func(a mock.Arguments) {
				s := a.Get(0).(*models.AuditSession)
				*s = models.AuditSession{ID: 9, SessionType: "slab", Status: "active"}
			}).Once()
		stubChain(mockDB)
		mockDB.On("Pluck", "serialnum", mock.AnythingOfType("*[]string")).Return(nil).Maybe()
		mockDB.On("Scan", mock.Anything).Return(nil).Maybe()

		router := newRouter(mockDB)
		body := map[string]interface{}{"category": []string{"Granite"}}
		req, _ := http.NewRequest("POST", "/audit/sessions?id=9&lines=true", toJSON(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "\"added\":0")
	})

	t.Run("session not found", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("First", mock.AnythingOfType("*models.AuditSession"), "999").
			Return(gorm.ErrRecordNotFound)

		router := newRouter(mockDB)
		body := map[string]interface{}{"serial": "SN-1"}
		req, _ := http.NewRequest("POST", "/audit/sessions?id=999&lines=true", toJSON(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("invalid JSON body", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("First", mock.AnythingOfType("*models.AuditSession"), "9").
			Return(nil).
			Run(func(a mock.Arguments) {
				s := a.Get(0).(*models.AuditSession)
				*s = models.AuditSession{ID: 9, Status: "active"}
			}).Once()

		router := newRouter(mockDB)
		req, _ := http.NewRequest("POST", "/audit/sessions?id=9&lines=true", bytes.NewBuffer([]byte(`{bad}`)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// ── Method dispatch / not-allowed branches ──────────────────────────────────

func TestHandleAuditSessionsCollection_MethodNotAllowed(t *testing.T) {
	mockDB := &mocks.MockDatabase{}
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(injectDB(mockDB))
	router.PATCH("/audit/sessions", handlers.HandleAuditSessionsCollection())

	req, _ := http.NewRequest("PATCH", "/audit/sessions", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHandleAuditSessionItem_MethodNotAllowed(t *testing.T) {
	mockDB := &mocks.MockDatabase{}
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(injectDB(mockDB))
	router.PATCH("/audit/sessions/:id", handlers.HandleAuditSessionItem())

	req, _ := http.NewRequest("PATCH", "/audit/sessions/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}