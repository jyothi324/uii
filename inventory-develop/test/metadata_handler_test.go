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

func TestHandleCreateMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)

	newRouter := func(mockDB *mocks.MockDatabase) *gin.Engine {
		r := gin.New()
		r.Use(injectDB(mockDB))
		r.POST("/metadata", handlers.HandleCreateMetadata())
		return r
	}

	doRequest := func(router *gin.Engine, body map[string]interface{}, typeParam string) *httptest.ResponseRecorder {
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/metadata?type="+typeParam, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w
	}

	t.Run("create type - success", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Create", mock.AnythingOfType("*models.MaterialType")).Return(nil)

		body := map[string]interface{}{"name": "Sheet Metal"}
		w := doRequest(newRouter(mockDB), body, "type")

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), "created successfully")
		mockDB.AssertExpectations(t)
	})

	t.Run("create category - success", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Create", mock.AnythingOfType("*models.MaterialCategory")).Return(nil)

		body := map[string]interface{}{"name": "Raw Materials"}
		w := doRequest(newRouter(mockDB), body, "category")

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), "created successfully")
		mockDB.AssertExpectations(t)
	})

	t.Run("create subcategory - success", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Create", mock.AnythingOfType("*models.MaterialSubCategory")).Return(nil)

		body := map[string]interface{}{"name": "Ferrous"}
		w := doRequest(newRouter(mockDB), body, "subcategory")

		assert.Equal(t, http.StatusCreated, w.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("create origin - success", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Create", mock.AnythingOfType("*models.MaterialOrigin")).Return(nil)

		body := map[string]interface{}{"name": "Domestic"}
		w := doRequest(newRouter(mockDB), body, "origin")

		assert.Equal(t, http.StatusCreated, w.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("create finish - success", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Create", mock.AnythingOfType("*models.MaterialFinish")).Return(nil)

		body := map[string]interface{}{"name": "Matte"}
		w := doRequest(newRouter(mockDB), body, "finish")

		assert.Equal(t, http.StatusCreated, w.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("create basecolor - success", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Create", mock.AnythingOfType("*models.MaterialBaseColor")).Return(nil)

		body := map[string]interface{}{"name": "Black"}
		w := doRequest(newRouter(mockDB), body, "basecolor")

		assert.Equal(t, http.StatusCreated, w.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("create level - success", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Create", mock.AnythingOfType("*models.MaterialLevel")).Return(nil)

		body := map[string]interface{}{"name": "Grade A"}
		w := doRequest(newRouter(mockDB), body, "level")
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("update type - success", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("First", mock.AnythingOfType("*models.MaterialType"), uint(1)).Return(nil)
		mockDB.On("Save", mock.AnythingOfType("*models.MaterialType")).Return(nil)

		body := map[string]interface{}{"id": 1, "name": "Updated type"}
		w := doRequest(newRouter(mockDB), body, "type")

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "updated successfully")
		mockDB.AssertExpectations(t)
	})

	t.Run("update - name gets set before save", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("First", mock.AnythingOfType("*models.MaterialType"), uint(5)).Return(nil)
		mockDB.On("Save", mock.AnythingOfType("*models.MaterialType")).
			Return(nil).
			Run(func(args mock.Arguments) {
				f := args.Get(0).(*models.MaterialType)
				assert.Equal(t, "New Name", f.Name)
			})

		body := map[string]interface{}{"id": 5, "name": "New Name"}
		w := doRequest(newRouter(mockDB), body, "type")

		assert.Equal(t, http.StatusOK, w.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("database connection unavailable", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		// Simulate a database error on creation
		mockDB.On("Create", mock.AnythingOfType("*models.MaterialType")).Return(errors.New("db connection error"))

		r := newRouter(mockDB)                                           // Use the helper that injects the mockDB
		w := doRequest(r, map[string]interface{}{"name": "foo"}, "type") // This will now call the handler with the mockDB
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to create") // Handler returns "Failed to create" on DB error
	})

	t.Run("invalid type param", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		body := map[string]interface{}{"name": "Foo"}
		w := doRequest(newRouter(mockDB), body, "invalid")

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid or missing 'type' query param")
		mockDB.AssertNotCalled(t, "Create")
	})

	t.Run("missing type param", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		jsonBody, _ := json.Marshal(map[string]interface{}{"name": "Foo"})
		req, _ := http.NewRequest("POST", "/metadata", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		newRouter(mockDB).ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid or missing 'type' query param")
		mockDB.AssertNotCalled(t, "Create")
	})

	t.Run("malformed JSON body", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		req, _ := http.NewRequest("POST", "/metadata?type=type", bytes.NewBufferString(`{"name": 123}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		newRouter(mockDB).ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request body")
		mockDB.AssertNotCalled(t, "Create")
	})

	t.Run("update type - record not found", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("First", mock.AnythingOfType("*models.MaterialType"), uint(9999)).
			Return(gorm.ErrRecordNotFound)

		body := map[string]interface{}{"id": 9999, "name": "Ghost"}
		w := doRequest(newRouter(mockDB), body, "type")

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockDB.AssertExpectations(t)
		mockDB.AssertNotCalled(t, "Save")
	})

	t.Run("database error on create", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Create", mock.AnythingOfType("*models.MaterialType")).
			Return(errors.New("db insert error"))

		body := map[string]interface{}{"name": "Error type"}
		w := doRequest(newRouter(mockDB), body, "type")

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to create")
		mockDB.AssertExpectations(t)
	})

	t.Run("database error on fetch during update", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("First", mock.AnythingOfType("*models.MaterialType"), uint(2)).
			Return(errors.New("db connection error"))

		body := map[string]interface{}{"id": 2, "name": "Updated"}
		w := doRequest(newRouter(mockDB), body, "type")

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to fetch")
		mockDB.AssertExpectations(t)
		mockDB.AssertNotCalled(t, "Save")
	})

	t.Run("database error on save during update", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("First", mock.AnythingOfType("*models.MaterialType"), uint(3)).Return(nil)
		mockDB.On("Save", mock.AnythingOfType("*models.MaterialType")).
			Return(errors.New("db save error"))

		body := map[string]interface{}{"id": 3, "name": "Updated"}
		w := doRequest(newRouter(mockDB), body, "type")

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to update")
		mockDB.AssertExpectations(t)
	})
}

func TestGetAllMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)

	newRouter := func(mockDB *mocks.MockDatabase) *gin.Engine {
		r := gin.New()
		r.Use(injectDB(mockDB))
		r.GET("/metadata", handlers.GetAllMetadata())
		return r
	}

	doRequest := func(router *gin.Engine, typeParam string) *httptest.ResponseRecorder {
		req, _ := http.NewRequest("GET", "/metadata?type="+typeParam, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w
	}

	t.Run("fetch all types - success", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Order", "updated_at desc").Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]models.MaterialType")).Return(nil)

		w := doRequest(newRouter(mockDB), "type")

		assert.Equal(t, http.StatusOK, w.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("fetch all categories - success", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Order", "updated_at desc").Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]models.MaterialCategory")).Return(nil)

		w := doRequest(newRouter(mockDB), "category")

		assert.Equal(t, http.StatusOK, w.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("fetch all subcategories - success", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Order", "updated_at desc").Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]models.MaterialSubCategory")).Return(nil)

		w := doRequest(newRouter(mockDB), "subcategory")

		assert.Equal(t, http.StatusOK, w.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("fetch all origins - success", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Order", "updated_at desc").Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]models.MaterialOrigin")).Return(nil)

		w := doRequest(newRouter(mockDB), "origin")

		assert.Equal(t, http.StatusOK, w.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("fetch all finishes - success", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Order", "updated_at desc").Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]models.MaterialFinish")).Return(nil)

		w := doRequest(newRouter(mockDB), "finish")

		assert.Equal(t, http.StatusOK, w.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("fetch all basecolors - success", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Order", "updated_at desc").Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]models.MaterialBaseColor")).Return(nil)

		w := doRequest(newRouter(mockDB), "basecolor")

		assert.Equal(t, http.StatusOK, w.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("invalid type param", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		w := doRequest(newRouter(mockDB), "bogus")

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid or missing 'type' query param")
		mockDB.AssertNotCalled(t, "Find")
	})

	t.Run("missing type param", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		req, _ := http.NewRequest("GET", "/metadata", nil)
		w := httptest.NewRecorder()
		newRouter(mockDB).ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid or missing 'type' query param")
		mockDB.AssertNotCalled(t, "Find")
	})

	t.Run("database error on find", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Order", "updated_at desc").Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]models.MaterialType")).
			Return(errors.New("db connection error"))

		w := doRequest(newRouter(mockDB), "type")

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to fetch")
		mockDB.AssertExpectations(t)
	})
}

func TestDeleteMetadataByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	newRouter := func(mockDB *mocks.MockDatabase) *gin.Engine {
		r := gin.New()
		r.Use(injectDB(mockDB))
		r.DELETE("/metadata/:id", handlers.DeleteMetadataByID())
		return r
	}

	doRequest := func(router *gin.Engine, id, typeParam string) *httptest.ResponseRecorder {
		req, _ := http.NewRequest("DELETE", "/metadata/"+id+"?type="+typeParam, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w
	}

	t.Run("delete type - success", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("First", mock.AnythingOfType("*models.MaterialType"), "1").Return(nil)
		mockDB.On("Delete", mock.AnythingOfType("*models.MaterialType")).Return(nil)

		w := doRequest(newRouter(mockDB), "1", "type")

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "deleted successfully")
		mockDB.AssertExpectations(t)
	})

	t.Run("delete category - success", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("First", mock.AnythingOfType("*models.MaterialCategory"), "2").Return(nil)
		mockDB.On("Delete", mock.AnythingOfType("*models.MaterialCategory")).Return(nil)

		w := doRequest(newRouter(mockDB), "2", "category")

		assert.Equal(t, http.StatusOK, w.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("record not found - returns 200 with not found message", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("First", mock.AnythingOfType("*models.MaterialType"), "9999").
			Return(gorm.ErrRecordNotFound)

		w := doRequest(newRouter(mockDB), "9999", "type")

		// Handler intentionally returns 200 with "not found" for missing records.
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "not found")
		mockDB.AssertNotCalled(t, "Delete")
	})

	t.Run("invalid type param", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		w := doRequest(newRouter(mockDB), "1", "nope")

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid or missing 'type' query param")
		mockDB.AssertNotCalled(t, "First")
		mockDB.AssertNotCalled(t, "Delete")
	})

	t.Run("missing type param", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		req, _ := http.NewRequest("DELETE", "/metadata/1", nil)
		w := httptest.NewRecorder()
		newRouter(mockDB).ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid or missing 'type' query param")
		mockDB.AssertNotCalled(t, "Delete")
	})

	t.Run("database error on first", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("First", mock.AnythingOfType("*models.MaterialType"), "1").
			Return(errors.New("db connection error"))

		w := doRequest(newRouter(mockDB), "1", "type")

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to fetch")
		mockDB.AssertExpectations(t)
		mockDB.AssertNotCalled(t, "Delete")
	})

	t.Run("database error on delete", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("First", mock.AnythingOfType("*models.MaterialType"), "1").Return(nil)
		mockDB.On("Delete", mock.AnythingOfType("*models.MaterialType")).
			Return(errors.New("db delete error"))

		w := doRequest(newRouter(mockDB), "1", "type")

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to delete")
		mockDB.AssertExpectations(t)
	})
}
