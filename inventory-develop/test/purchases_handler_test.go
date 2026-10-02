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

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandleCreatePurchase(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("valid request - no material id", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Create", mock.AnythingOfType("*models.Purchase")).Return(nil)
		mockDB.On("Create", mock.AnythingOfType("*handlers.ActivityEvent")).Return(nil)
		mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(nil)

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/purchase/create", handlers.HandleCreatePurchase())

		reqBody := models.Purchase{Name: "Test Item", Quantity: "10", CurrentPrice: "99.99"}
		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/purchase/create", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), "Purchase created successfully")

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Test Item", response["name"])
		mockDB.AssertExpectations(t)
	})

	t.Run("valid request - with material id", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Create", mock.AnythingOfType("*models.Purchase")).Return(nil)
		mockDB.On("Create", mock.AnythingOfType("*handlers.ActivityEvent")).Return(nil)
		mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(nil)

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/purchase/create", handlers.HandleCreatePurchase())

		reqBody := models.Purchase{Name: "Test Item", MaterialID: 1, Quantity: "10", CurrentPrice: "99.99"}
		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/purchase/create", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), "Purchase created successfully")
		mockDB.AssertExpectations(t)
	})

	t.Run("valid request - array of items body", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Create", mock.AnythingOfType("*models.Purchase")).Return(nil)
		mockDB.On("Create", mock.AnythingOfType("*handlers.ActivityEvent")).Return(nil)
		mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(nil)

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/purchase/create", handlers.HandleCreatePurchase())

		items := []models.PurchaseItem{{MaterialID: 1, Quantity: "5", CurrentPrice: "10", Dimensions: "10x10"}}
		jsonBody, _ := json.Marshal(items)
		req, _ := http.NewRequest("POST", "/purchase/create", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("valid request - auto-populate items from strings", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Create", mock.AnythingOfType("*models.Purchase")).Return(nil)
		mockDB.On("Create", mock.AnythingOfType("*handlers.ActivityEvent")).Return(nil)
		mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(nil)

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/purchase/create", handlers.HandleCreatePurchase())

		reqBody := models.Purchase{Name: "Bulk", Quantity: "5,10", CurrentPrice: "10,20", Dimensions: "10x10,20x20", MaterialID: 1}
		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/purchase/create", bytes.NewBuffer(jsonBody))
		router.ServeHTTP(httptest.NewRecorder(), req)
	})

	t.Run("invalid request body - malformed JSON", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/purchase/create", handlers.HandleCreatePurchase())

		req, _ := http.NewRequest("POST", "/purchase/create",
			bytes.NewBuffer([]byte(`{"name": "Test", "quantity": "invalid"`))) // Truly malformed JSON
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request body")
	})

	t.Run("database error on insert", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Create", mock.AnythingOfType("*models.Purchase")).
			Return(errors.New("db insert error"))

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/purchase/create", handlers.HandleCreatePurchase())

		reqBody := models.Purchase{Name: "Test Item", Quantity: "10", CurrentPrice: "99.99"}
		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/purchase/create", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to create purchase")
		mockDB.AssertExpectations(t)
	})
}

// ── HandleFetchPurchases ──────────────────────────────────────────────────────

func TestHandleFetchPurchases(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("fetch all purchases successfully", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		// Handler chains: database.Preload("Materials").Find(...)
		mockDB.On("Preload", "Materials").Return(mockDB)
		mockDB.On("Preload", "Supplier").Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]models.Purchase")).Return(nil)

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.GET("/purchase/all", handlers.HandleFetchPurchases())

		req, _ := http.NewRequest("GET", "/purchase/all", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("fetch purchases - empty database", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Preload", "Materials").Return(mockDB)
		mockDB.On("Preload", "Supplier").Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]models.Purchase")).Return(nil)

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.GET("/purchase/all", handlers.HandleFetchPurchases())

		req, _ := http.NewRequest("GET", "/purchase/all", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var purchases []models.Purchase
		json.Unmarshal(w.Body.Bytes(), &purchases)
		assert.Equal(t, 0, len(purchases))
	})

	t.Run("database query error", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Preload", "Materials").Return(mockDB)
		mockDB.On("Preload", "Supplier").Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]models.Purchase")).
			Return(errors.New("db error"))

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.GET("/purchase/all", handlers.HandleFetchPurchases())

		req, _ := http.NewRequest("GET", "/purchase/all", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to fetch purchases")
	})
}

// ── HandlePurchaseGetByID ─────────────────────────────────────────────────────

func TestHandlePurchaseGetByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("get purchase by id successfully", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Preload", "Materials").Return(mockDB)
		mockDB.On("Preload", "Supplier").Return(mockDB)
		mockDB.On("First", mock.AnythingOfType("*models.Purchase"), mock.Anything).
			Return(nil).
			Run(func(args mock.Arguments) {
				p := args.Get(0).(*models.Purchase)
				p.ID = 1
				p.Name = "Test"
				p.Quantity = "10"
			}).Once()

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.GET("/purchase/:id", handlers.HandlePurchaseGetByID())

		req, _ := http.NewRequest("GET", "/purchase/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Test")
		mockDB.AssertExpectations(t)
	})

	t.Run("missing id - route not matched returns 404", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Preload", "Materials").Return(mockDB)
		mockDB.On("Preload", "Supplier").Return(mockDB)

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.GET("/purchase/:id", handlers.HandlePurchaseGetByID())

		req, _ := http.NewRequest("GET", "/purchase/", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("purchase not found", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Preload", "Materials").Return(mockDB)
		mockDB.On("Preload", "Supplier").Return(mockDB)
		// Handler returns 200 with a message when the record is not found (by design).
		mockDB.On("First", mock.AnythingOfType("*models.Purchase"), mock.Anything).
			Return(gorm.ErrRecordNotFound)

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.GET("/purchase/:id", handlers.HandlePurchaseGetByID())

		req, _ := http.NewRequest("GET", "/purchase/99999", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Purchase not found")
	})

	t.Run("database error on fetch", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Preload", "Materials").Return(mockDB)
		mockDB.On("Preload", "Supplier").Return(mockDB)
		mockDB.On("First", mock.AnythingOfType("*models.Purchase"), mock.Anything).
			Return(errors.New("db error"))

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.GET("/purchase/:id", handlers.HandlePurchaseGetByID())

		req, _ := http.NewRequest("GET", "/purchase/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to fetch purchase")
	})
}

// ── HandlePurchaseUpdate ──────────────────────────────────────────────────────

func TestHandlePurchaseUpdate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("update purchase status successfully", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		existing := models.Purchase{Name: "Test Item", CurrentPrice: "99.99"}

		mockDB.On("First", mock.AnythingOfType("*models.Purchase"), "1").
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.Purchase) = existing }).Once()
		mockDB.On("Save", mock.AnythingOfType("*models.Purchase")).Return(nil).Once()
		mockDB.On("Create", mock.AnythingOfType("*handlers.ActivityEvent")).Return(nil).Once()
		mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(nil).Once()

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.PUT("/purchase/:id", handlers.HandlePurchaseUpdate())

		body := models.Purchase{Name: "Updated Item", Status: "Approved", CurrentPrice: "120.00"}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("PUT", "/purchase/1", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Purchase updated successfully")
		mockDB.AssertExpectations(t)
	})

	t.Run("update triggers re-parsing items", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		existing := models.Purchase{ID: 1, Name: "Test", Quantity: "10", CurrentPrice: "50", Dimensions: "10x10", MaterialID: 1}

		mockDB.On("First", mock.AnythingOfType("*models.Purchase"), "1").
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.Purchase) = existing }).Once()

		mockDB.On("Save", mock.MatchedBy(func(p *models.Purchase) bool {
			return p.Quantity == "20" && len(p.PurchaseItems) > 0
		})).Return(nil).Once()
		mockDB.On("Create", mock.AnythingOfType("*handlers.ActivityEvent")).Return(nil).Once()
		mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(nil).Once()

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.PUT("/purchase/:id", handlers.HandlePurchaseUpdate())

		body := map[string]interface{}{"quantity": "20"}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("PUT", "/purchase/1", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(httptest.NewRecorder(), req)
	})

	t.Run("invalid request body - malformed JSON", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.PUT("/purchase/:id", handlers.HandlePurchaseUpdate())

		req, _ := http.NewRequest("PUT", "/purchase/1",
			bytes.NewBuffer([]byte(`{"quantity":"bad"`))) // Truly malformed JSON
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request body")
	})

	t.Run("purchase not found", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		// Handler returns 200 with a message when the record is not found (by design).
		mockDB.On("First", mock.AnythingOfType("*models.Purchase"), "999").
			Return(gorm.ErrRecordNotFound).Once()

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.PUT("/purchase/:id", handlers.HandlePurchaseUpdate())

		body := models.Purchase{Name: "Test Item", CurrentPrice: "99.99"}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("PUT", "/purchase/999", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Purchase not found")
		mockDB.AssertExpectations(t)
	})

	t.Run("database error on fetch", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("First", mock.AnythingOfType("*models.Purchase"), "1").
			Return(errors.New("db error")).Once()

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.PUT("/purchase/:id", handlers.HandlePurchaseUpdate())

		body := models.Purchase{Name: "Test Item", CurrentPrice: "99.99"}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("PUT", "/purchase/1", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to fetch purchase")
		mockDB.AssertExpectations(t)
	})

	t.Run("database error on save", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		existing := models.Purchase{Name: "Test Item", CurrentPrice: "99.99"}

		mockDB.On("First", mock.AnythingOfType("*models.Purchase"), "1").
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.Purchase) = existing }).Once()
		mockDB.On("Save", mock.AnythingOfType("*models.Purchase")).
			Return(errors.New("save error")).Once()

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.PUT("/purchase/:id", handlers.HandlePurchaseUpdate())

		body := models.Purchase{Name: "Test Item", CurrentPrice: "99.99"}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("PUT", "/purchase/1", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to update purchase")
		mockDB.AssertExpectations(t)
	})

	t.Run("update sets rejection reason", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		existing := models.Purchase{Name: "Test Item", CurrentPrice: "99.99", Status: "Pending"}

		mockDB.On("First", mock.AnythingOfType("*models.Purchase"), "1").
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.Purchase) = existing }).Once()
		mockDB.On("Save", mock.AnythingOfType("*models.Purchase")).Return(nil).Once()
		mockDB.On("Create", mock.AnythingOfType("*handlers.ActivityEvent")).Return(nil).Once()
		mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(nil).Once()

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.PUT("/purchase/:id", handlers.HandlePurchaseUpdate())

		body := models.Purchase{Status: "Rejected", RejectionReason: "Out of budget"}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("PUT", "/purchase/1", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Purchase updated successfully")
		mockDB.AssertExpectations(t)
	})
}

// ── DeletePurchaseByID ────────────────────────────────────────────────────────

func TestDeletePurchaseByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("delete purchase successfully", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		existing := models.Purchase{ID: 1, Name: "Item"}

		mockDB.On("First", mock.AnythingOfType("*models.Purchase"), "1").
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.Purchase) = existing }).Once()
		mockDB.On("Delete", mock.AnythingOfType("*models.Purchase")).Return(nil).Once()

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.DELETE("/purchase/:id", handlers.DeletePurchaseByID())

		req, _ := http.NewRequest("DELETE", "/purchase/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Purchase deleted successfully")
		mockDB.AssertExpectations(t)
	})

	t.Run("purchase not found", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		// Handler returns 200 with a message when the record is not found (by design).
		mockDB.On("First", mock.AnythingOfType("*models.Purchase"), "999").
			Return(gorm.ErrRecordNotFound).Once()

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.DELETE("/purchase/:id", handlers.DeletePurchaseByID())

		req, _ := http.NewRequest("DELETE", "/purchase/999", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Purchase not found")
		mockDB.AssertExpectations(t)
	})

	t.Run("database error on fetch", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("First", mock.AnythingOfType("*models.Purchase"), "1").
			Return(errors.New("db error")).Once()

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.DELETE("/purchase/:id", handlers.DeletePurchaseByID())

		req, _ := http.NewRequest("DELETE", "/purchase/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to fetch purchase")
		mockDB.AssertExpectations(t)
	})

	t.Run("database error on delete", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		existing := models.Purchase{ID: 1}

		mockDB.On("First", mock.AnythingOfType("*models.Purchase"), "1").
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.Purchase) = existing }).Once()
		mockDB.On("Delete", mock.AnythingOfType("*models.Purchase")).
			Return(errors.New("delete error")).Once()

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.DELETE("/purchase/:id", handlers.DeletePurchaseByID())

		req, _ := http.NewRequest("DELETE", "/purchase/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to delete purchase")
		mockDB.AssertExpectations(t)
	})

	t.Run("missing id - route not matched returns 404", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.DELETE("/purchase/:id", handlers.DeletePurchaseByID())

		req, _ := http.NewRequest("DELETE", "/purchase/", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
