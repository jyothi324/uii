package test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"inventory/handlers"
	"inventory/mocks"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/datatypes"
)

func TestGetRecentActivity(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Helper to create activity events for testing
	createActivityEvent := func(id int64, entityType, entityID, action, userID string, payload []byte) handlers.ActivityEvent {
		userName := "Test User"
		return handlers.ActivityEvent{
			ID:         id,
			EntityType: entityType,
			EntityID:   entityID,
			Action:     action,
			UserID:     userID,
			UserName:   &userName,
			Payload:    datatypes.JSON(payload),
			CreatedAt:  time.Now(),
		}
	}

	newRouter := func(mockDB *mocks.MockDatabase) *gin.Engine {
		r := gin.New()
		r.Use(injectDB(mockDB))
		r.GET("/activity/recent", handlers.GetRecentActivity)
		return r
	}

	doRequest := func(router *gin.Engine, queryParams string) *httptest.ResponseRecorder {
		url := "/activity/recent"
		if queryParams != "" {
			url += "?" + queryParams
		}
		req, _ := http.NewRequest("GET", url, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w
	}

	t.Run("successful retrieval with default limit", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		events := []handlers.ActivityEvent{
			createActivityEvent(1, "lead", "lead-123", "created", "user-1", []byte(`{"name":"John"}`)),
			createActivityEvent(2, "lead", "lead-124", "updated", "user-2", []byte(`{"status":"active"}`)),
		}

		mockDB.On("Order", "created_at desc").Return(mockDB)
		mockDB.On("Limit", 100).Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]handlers.ActivityEvent")).
			Return(nil).
			Run(func(args mock.Arguments) {
				dest := args.Get(0).(*[]handlers.ActivityEvent)
				*dest = events
			})

		w := doRequest(newRouter(mockDB), "")

		assert.Equal(t, http.StatusOK, w.Code)
		var result []handlers.ActivityEvent
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Equal(t, 2, len(result))
		assert.Equal(t, int64(1), result[0].ID)
		assert.Equal(t, "lead", result[0].EntityType)
		mockDB.AssertExpectations(t)
	})

	t.Run("successful retrieval with custom limit", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		events := []handlers.ActivityEvent{
			createActivityEvent(1, "lead", "lead-123", "created", "user-1", []byte(`{"name":"John"}`)),
		}

		mockDB.On("Order", "created_at desc").Return(mockDB)
		mockDB.On("Limit", 50).Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]handlers.ActivityEvent")).
			Return(nil).
			Run(func(args mock.Arguments) {
				dest := args.Get(0).(*[]handlers.ActivityEvent)
				*dest = events
			})

		w := doRequest(newRouter(mockDB), "limit=50")

		assert.Equal(t, http.StatusOK, w.Code)
		var result []handlers.ActivityEvent
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Equal(t, 1, len(result))
		mockDB.AssertCalled(t, "Limit", 50)
		mockDB.AssertExpectations(t)
	})

	t.Run("limit capped at max 200", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		mockDB.On("Order", "created_at desc").Return(mockDB)
		mockDB.On("Limit", 200).Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]handlers.ActivityEvent")).Return(nil)

		w := doRequest(newRouter(mockDB), "limit=500")

		assert.Equal(t, http.StatusOK, w.Code)
		mockDB.AssertCalled(t, "Limit", 200)
	})

	t.Run("invalid limit parameter - non-numeric", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		mockDB.On("Order", "created_at desc").Return(mockDB)
		mockDB.On("Limit", 100).Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]handlers.ActivityEvent")).Return(nil)

		w := doRequest(newRouter(mockDB), "limit=invalid")

		assert.Equal(t, http.StatusOK, w.Code)
		mockDB.AssertCalled(t, "Limit", 100)
	})

	t.Run("invalid limit parameter - zero", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		mockDB.On("Order", "created_at desc").Return(mockDB)
		mockDB.On("Limit", 100).Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]handlers.ActivityEvent")).Return(nil)

		w := doRequest(newRouter(mockDB), "limit=0")

		assert.Equal(t, http.StatusOK, w.Code)
		mockDB.AssertCalled(t, "Limit", 100)
	})

	t.Run("invalid limit parameter - negative", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		mockDB.On("Order", "created_at desc").Return(mockDB)
		mockDB.On("Limit", 100).Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]handlers.ActivityEvent")).Return(nil)

		w := doRequest(newRouter(mockDB), "limit=-10")

		assert.Equal(t, http.StatusOK, w.Code)
		mockDB.AssertCalled(t, "Limit", 100)
	})

	t.Run("database error during find", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		mockDB.On("Order", "created_at desc").Return(mockDB)
		mockDB.On("Limit", 100).Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]handlers.ActivityEvent")).
			Return(errors.New("database connection lost"))

		w := doRequest(newRouter(mockDB), "")

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "failed to load activity events")
		mockDB.AssertExpectations(t)
	})

	t.Run("empty activity events", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		mockDB.On("Order", "created_at desc").Return(mockDB)
		mockDB.On("Limit", 100).Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]handlers.ActivityEvent")).
			Return(nil).
			Run(func(args mock.Arguments) {
				dest := args.Get(0).(*[]handlers.ActivityEvent)
				*dest = []handlers.ActivityEvent{}
			})

		w := doRequest(newRouter(mockDB), "")

		assert.Equal(t, http.StatusOK, w.Code)
		var result []handlers.ActivityEvent
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Equal(t, 0, len(result))
		mockDB.AssertExpectations(t)
	})

	t.Run("multiple activity events in order", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		now := time.Now()
		events := []handlers.ActivityEvent{
			createActivityEvent(3, "lead", "lead-126", "deleted", "user-3", []byte(`{"reason":"inactive"}`)),
			createActivityEvent(2, "lead", "lead-125", "updated", "user-2", []byte(`{"status":"closed"}`)),
			createActivityEvent(1, "lead", "lead-123", "created", "user-1", []byte(`{"name":"John"}`)),
		}
		events[0].CreatedAt = now
		events[1].CreatedAt = now.Add(-1 * time.Minute)
		events[2].CreatedAt = now.Add(-2 * time.Minute)

		mockDB.On("Order", "created_at desc").Return(mockDB)
		mockDB.On("Limit", 100).Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]handlers.ActivityEvent")).
			Return(nil).
			Run(func(args mock.Arguments) {
				dest := args.Get(0).(*[]handlers.ActivityEvent)
				*dest = events
			})

		w := doRequest(newRouter(mockDB), "")

		assert.Equal(t, http.StatusOK, w.Code)
		var result []handlers.ActivityEvent
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Equal(t, 3, len(result))
		// Most recent first
		assert.Equal(t, int64(3), result[0].ID)
		assert.Equal(t, int64(2), result[1].ID)
		assert.Equal(t, int64(1), result[2].ID)
		mockDB.AssertExpectations(t)
	})

	t.Run("activity with null user_name", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		events := []handlers.ActivityEvent{
			{
				ID:         1,
				EntityType: "lead",
				EntityID:   "lead-123",
				Action:     "created",
				UserID:     "system",
				UserName:   nil,
				Payload:    datatypes.JSON(`{}`),
				CreatedAt:  time.Now(),
			},
		}

		mockDB.On("Order", "created_at desc").Return(mockDB)
		mockDB.On("Limit", 100).Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]handlers.ActivityEvent")).
			Return(nil).
			Run(func(args mock.Arguments) {
				dest := args.Get(0).(*[]handlers.ActivityEvent)
				*dest = events
			})

		w := doRequest(newRouter(mockDB), "")

		assert.Equal(t, http.StatusOK, w.Code)
		var result []handlers.ActivityEvent
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Equal(t, 1, len(result))
		assert.Nil(t, result[0].UserName)
		mockDB.AssertExpectations(t)
	})
}
