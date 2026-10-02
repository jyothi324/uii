package test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"inventory/db"
	"inventory/handlers"
	"inventory/mocks"
	"inventory/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func validSupplierInput() map[string]interface{} {
	return map[string]interface{}{
		"name":     "Acme Corp",
		"category": "Raw Materials",
		"phone":    "+1-555-0100",
		"email":    "acme@example.com",
		"website":  "https://acme.example.com",
		"address": map[string]interface{}{
			"street":  "123 Main St",
			"city":    "Springfield",
			"state":   "IL",
			"zip":     "62701",
			"country": "USA",
		},
		"paymentTerms": "Net30",
		"pricingType":  "Fixed",
		"isReferred":   false,
		"notes":        []interface{}{},
	}
}

func supplierJSON(fields map[string]interface{}) *bytes.Buffer {
	b, _ := json.Marshal(fields)
	return bytes.NewBuffer(b)
}

// ── CreateSupplier ────────────────────────────────────────────────────────────

func TestCreateSupplier(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("valid request - creates supplier", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Create", mock.AnythingOfType("*models.Supplier")).Return(nil)

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/supplier", handlers.CreateSupplier())

		req, _ := http.NewRequest("POST", "/supplier", supplierJSON(validSupplierInput()))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var resp models.Supplier
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "Acme Corp", resp.Name)
		mockDB.AssertExpectations(t)
	})

	t.Run("invalid JSON body", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/supplier", handlers.CreateSupplier())

		req, _ := http.NewRequest("POST", "/supplier", bytes.NewBuffer([]byte(`{bad}`)))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request body")
	})

	t.Run("missing name field", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/supplier", handlers.CreateSupplier())

		input := validSupplierInput()
		delete(input, "name")
		req, _ := http.NewRequest("POST", "/supplier", supplierJSON(input))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Name and Category are required")
	})

	t.Run("missing category field", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/supplier", handlers.CreateSupplier())

		input := validSupplierInput()
		delete(input, "category")
		req, _ := http.NewRequest("POST", "/supplier", supplierJSON(input))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Name and Category are required")
	})

	t.Run("incomplete address - missing city", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/supplier", handlers.CreateSupplier())

		input := validSupplierInput()
		input["address"] = map[string]interface{}{
			"street":  "123 Main St",
			"state":   "IL",
			"zip":     "62701",
			"country": "USA",
			// city omitted intentionally
		}
		req, _ := http.NewRequest("POST", "/supplier", supplierJSON(input))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Address is required")
	})

	t.Run("duplicate email - unique constraint error", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Create", mock.AnythingOfType("*models.Supplier")).
			Return(errors.New("ERROR: duplicate key value violates unique constraint idx_suppliers_email"))

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/supplier", handlers.CreateSupplier())

		req, _ := http.NewRequest("POST", "/supplier", supplierJSON(validSupplierInput()))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Email already exists")
		mockDB.AssertExpectations(t)
	})

	t.Run("duplicate phone - unique constraint error", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Create", mock.AnythingOfType("*models.Supplier")).
			Return(errors.New("ERROR: duplicate key value violates unique constraint idx_suppliers_phone"))

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/supplier", handlers.CreateSupplier())

		req, _ := http.NewRequest("POST", "/supplier", supplierJSON(validSupplierInput()))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Phone already exists")
		mockDB.AssertExpectations(t)
	})

	t.Run("general database error", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Create", mock.AnythingOfType("*models.Supplier")).
			Return(errors.New("connection refused"))

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/supplier", handlers.CreateSupplier())

		req, _ := http.NewRequest("POST", "/supplier", supplierJSON(validSupplierInput()))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to create supplier")
		mockDB.AssertExpectations(t)
	})

	t.Run("referred supplier stores referrer fields in details", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Create", mock.AnythingOfType("*models.Supplier")).Return(nil)

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/supplier", handlers.CreateSupplier())

		input := validSupplierInput()
		input["isReferred"] = true
		input["referrerName"] = "Bob"
		input["referrerPhone"] = "+1-555-9999"
		input["referrerEmail"] = "bob@example.com"
		input["referrerCompanyName"] = "BobCo"
		input["referrerCompanyPhone"] = "+1-555-8888"
		input["referrerCompanyEmail"] = "info@bobco.com"

		req, _ := http.NewRequest("POST", "/supplier", supplierJSON(input))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var resp models.Supplier
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.True(t, resp.Details.IsReferred)
		assert.Equal(t, "Bob", resp.Details.ReferrerName)
		mockDB.AssertExpectations(t)
	})
}

// ── GetAllSuppliers ───────────────────────────────────────────────────────────

func TestGetAllSuppliers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("fetch all suppliers successfully", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		// Handler chains: database.Order("updated_at desc").Find(...)
		mockDB.On("Order", "updated_at desc").Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]models.Supplier")).
			Return(nil).
			Run(func(args mock.Arguments) {
				sl := args.Get(0).(*[]models.Supplier)
				*sl = []models.Supplier{
					{ID: 1, Name: "Acme Corp"},
					{ID: 2, Name: "Globex"},
				}
			}).Once()

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.GET("/supplier", handlers.GetAllSuppliers())

		req, _ := http.NewRequest("GET", "/supplier", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var suppliers []models.Supplier
		json.Unmarshal(w.Body.Bytes(), &suppliers)
		assert.Equal(t, 2, len(suppliers))
		mockDB.AssertExpectations(t)
	})

	t.Run("fetch suppliers - empty database", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Order", "updated_at desc").Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]models.Supplier")).Return(nil)

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.GET("/supplier", handlers.GetAllSuppliers())

		req, _ := http.NewRequest("GET", "/supplier", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var suppliers []models.Supplier
		json.Unmarshal(w.Body.Bytes(), &suppliers)
		assert.Equal(t, 0, len(suppliers))
	})

	t.Run("database error", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Order", "updated_at desc").Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]models.Supplier")).
			Return(errors.New("db error"))

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.GET("/supplier", handlers.GetAllSuppliers())

		req, _ := http.NewRequest("GET", "/supplier", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// ── GetSupplierByID ───────────────────────────────────────────────────────────

func TestGetSupplierByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("get supplier by id successfully", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("First", mock.AnythingOfType("*models.Supplier"), mock.Anything).
			Return(nil).
			Run(func(args mock.Arguments) {
				s := args.Get(0).(*models.Supplier)
				s.ID = 1
				s.Name = "Acme Corp"
			}).Once()

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.GET("/supplier/:id", handlers.GetSupplierByID())

		req, _ := http.NewRequest("GET", "/supplier/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Acme Corp")
		mockDB.AssertExpectations(t)
	})

	t.Run("missing id - route not matched returns 404", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.GET("/supplier/:id", handlers.GetSupplierByID())

		req, _ := http.NewRequest("GET", "/supplier/", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("supplier not found", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		// Handler returns 200 with a message when the record is not found (by design).
		mockDB.On("First", mock.AnythingOfType("*models.Supplier"), mock.Anything).
			Return(gorm.ErrRecordNotFound)

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.GET("/supplier/:id", handlers.GetSupplierByID())

		req, _ := http.NewRequest("GET", "/supplier/99999", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Supplier not found")
	})

	t.Run("database error on fetch", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("First", mock.AnythingOfType("*models.Supplier"), mock.Anything).
			Return(errors.New("db error"))

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.GET("/supplier/:id", handlers.GetSupplierByID())

		req, _ := http.NewRequest("GET", "/supplier/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to fetch supplier")
	})
}

// ── UpdateSupplier ────────────────────────────────────────────────────────────

func TestUpdateSupplier(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("update supplier successfully", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		existing := models.Supplier{ID: 1, Name: "Acme Corp", Category: "Raw Materials"}

		mockDB.On("First", mock.AnythingOfType("*models.Supplier"), mock.Anything).
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.Supplier) = existing }).Once()

		mockDB.On("Transaction", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				fc := args.Get(0).(func(tx db.Database) error)
				fc(mockDB)
			}).
			Return(nil).Once()

		mockDB.On("Save", mock.AnythingOfType("*models.Supplier")).Return(nil).Once()

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.PUT("/supplier/:id", handlers.UpdateSupplier())

		req, _ := http.NewRequest("PUT", "/supplier/1", supplierJSON(validSupplierInput()))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp models.Supplier
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "Acme Corp", resp.Name)
		mockDB.AssertExpectations(t)
	})

	t.Run("supplier not found", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		// UpdateSupplier returns 404 (unlike the Get/Delete handlers) when not found.
		mockDB.On("First", mock.AnythingOfType("*models.Supplier"), mock.Anything).
			Return(gorm.ErrRecordNotFound).Once()

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.PUT("/supplier/:id", handlers.UpdateSupplier())

		req, _ := http.NewRequest("PUT", "/supplier/999", supplierJSON(validSupplierInput()))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Supplier not found")
		mockDB.AssertExpectations(t)
	})

	t.Run("invalid JSON body", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		// UpdateSupplier fetches first, then binds JSON.
		mockDB.On("First", mock.AnythingOfType("*models.Supplier"), mock.Anything).
			Return(nil).Once()

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.PUT("/supplier/:id", handlers.UpdateSupplier())

		req, _ := http.NewRequest("PUT", "/supplier/1", bytes.NewBuffer([]byte(`{bad}`)))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request body")
	})

	t.Run("missing required fields in update body", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("First", mock.AnythingOfType("*models.Supplier"), mock.Anything).
			Return(nil).Once()

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.PUT("/supplier/:id", handlers.UpdateSupplier())

		input := validSupplierInput()
		delete(input, "name")
		req, _ := http.NewRequest("PUT", "/supplier/1", supplierJSON(input))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Name and Category are required")
	})

	t.Run("incomplete address in update body", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("First", mock.AnythingOfType("*models.Supplier"), mock.Anything).
			Return(nil).Once()

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.PUT("/supplier/:id", handlers.UpdateSupplier())

		input := validSupplierInput()
		input["address"] = map[string]interface{}{"street": "123 Main"} // incomplete
		req, _ := http.NewRequest("PUT", "/supplier/1", supplierJSON(input))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Address is required")
	})

	t.Run("duplicate email on save", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		existing := models.Supplier{ID: 1, Name: "Acme Corp", Category: "Raw Materials"}

		mockDB.On("First", mock.AnythingOfType("*models.Supplier"), mock.Anything).
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.Supplier) = existing }).Once()

		mockDB.On("Transaction", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				fc := args.Get(0).(func(tx db.Database) error)
				fc(mockDB)
			}).
			Return(errors.New("email constraint violation")).Once()

		// The error message must contain "email" for the handler's strings.Contains check.
		mockDB.On("Save", mock.AnythingOfType("*models.Supplier")).
			Return(errors.New("email constraint violation")).Once()

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.PUT("/supplier/:id", handlers.UpdateSupplier())

		req, _ := http.NewRequest("PUT", "/supplier/1", supplierJSON(validSupplierInput()))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Email already exists")
		mockDB.AssertExpectations(t)
	})

	t.Run("general database error on save", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		existing := models.Supplier{ID: 1, Name: "Acme Corp", Category: "Raw Materials"}

		mockDB.On("First", mock.AnythingOfType("*models.Supplier"), mock.Anything).
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.Supplier) = existing }).Once()

		mockDB.On("Transaction", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				fc := args.Get(0).(func(tx db.Database) error)
				fc(mockDB)
			}).
			Return(errors.New("connection refused")).Once()

		// Error message has no "email" or "phone" substring → handler falls through to 500.
		mockDB.On("Save", mock.AnythingOfType("*models.Supplier")).
			Return(errors.New("connection refused")).Once()

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.PUT("/supplier/:id", handlers.UpdateSupplier())

		req, _ := http.NewRequest("PUT", "/supplier/1", supplierJSON(validSupplierInput()))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("id change to an ID already in use", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		existing := models.Supplier{ID: 1, Name: "Acme Corp", Category: "Raw Materials"}
		conflicting := models.Supplier{ID: 2, Name: "Other Co", Category: "Metals"}

		// First call: fetch the supplier being edited (ID 1).
		mockDB.On("First", mock.AnythingOfType("*models.Supplier"), "1").
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.Supplier) = existing }).Once()

		// Second call: check whether the requested new ID (2) already exists.
		mockDB.On("First", mock.AnythingOfType("*models.Supplier"), uint(2)).
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.Supplier) = conflicting }).Once()

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.PUT("/supplier/:id", handlers.UpdateSupplier())

		input := validSupplierInput()
		input["id"] = 2
		req, _ := http.NewRequest("PUT", "/supplier/1", supplierJSON(input))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "This ID is already in use")
		mockDB.AssertExpectations(t)
	})
}

// ── DeleteSupplier ────────────────────────────────────────────────────────────

func TestDeleteSupplier(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("delete supplier successfully", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		// DeleteSupplier calls db.Delete(&models.Supplier{}, id) directly — no prior First.
		mockDB.On("Delete", mock.AnythingOfType("*models.Supplier"), mock.Anything).
			Return(nil).Once()

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.DELETE("/supplier/:id", handlers.DeleteSupplier())

		req, _ := http.NewRequest("DELETE", "/supplier/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Supplier deleted successfully")
		mockDB.AssertExpectations(t)
	})

	t.Run("database error on delete", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Delete", mock.AnythingOfType("*models.Supplier"), mock.Anything).
			Return(errors.New("db error")).Once()

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.DELETE("/supplier/:id", handlers.DeleteSupplier())

		req, _ := http.NewRequest("DELETE", "/supplier/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("missing id - route not matched returns 404", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.DELETE("/supplier/:id", handlers.DeleteSupplier())

		req, _ := http.NewRequest("DELETE", "/supplier/", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
