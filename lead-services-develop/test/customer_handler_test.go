package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/tuliporg/agent-management-api/db"
	"github.com/tuliporg/agent-management-api/handlers"
	"github.com/tuliporg/agent-management-api/mocks"
	"github.com/tuliporg/agent-management-api/models"
	"gorm.io/gorm"
)

func TestCreateCustomer_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	input := models.Customer{
		Name:  "Test Customer",
		Email: "test@example.com",
		Phone: "1234567890",
	}

	mockDB.On("Model", mock.Anything).Return(mockDB)
	mockDB.On("Where", "id = ?", mock.Anything).Return(mockDB)
	mockDB.On("Count", mock.Anything).Return(mockDB).Run(func(args mock.Arguments) {
		countPtr := args.Get(0).(*int64)
		*countPtr = 0
	})
	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Create", mock.Anything).Return(nil)
	mockDB.On("Commit").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/customers", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateCustomer(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Customer created successfully", resp["message"])

	mockDB.AssertExpectations(t)
}

func TestCreateCustomer_EmptyName_ReuseExisting(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	input := models.Customer{
		Name: "   ",
	}

	existingUnknown := models.Customer{
		ID:   "UNK123",
		Name: "Unknown company",
	}

	mockDB.On("Where", "LOWER(name) = LOWER(?)", "Unknown company").Return(mockDB)
	mockDB.On("First", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		customer := args.Get(0).(*models.Customer)
		*customer = existingUnknown
	})

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/customers", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateCustomer(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Reused existing Unknown company", resp["message"])

	mockDB.AssertExpectations(t)
}

func TestCreateCustomer_EmptyName_CreateNew(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	input := models.Customer{
		Name: "",
	}

	mockDB.On("Where", "LOWER(name) = LOWER(?)", "Unknown company").Return(mockDB)
	mockDB.On("First", mock.Anything).Return(gorm.ErrRecordNotFound)
	mockDB.On("Model", mock.Anything).Return(mockDB)
	mockDB.On("Where", "id = ?", mock.Anything).Return(mockDB)
	mockDB.On("Count", mock.Anything).Return(mockDB).Run(func(args mock.Arguments) {
		countPtr := args.Get(0).(*int64)
		*countPtr = 0
	})
	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Create", mock.Anything).Return(nil)
	mockDB.On("Commit").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/customers", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateCustomer(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Customer created successfully", resp["message"])

	mockDB.AssertExpectations(t)
}

func TestCreateCustomer_InvalidJSON(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("POST", "/customers", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateCustomer(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Invalid input", resp.Error)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
}

func TestCreateCustomer_DatabaseError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	input := models.Customer{
		Name: "Test Customer",
	}

	mockDB.On("Model", mock.Anything).Return(mockDB)
	mockDB.On("Where", "id = ?", mock.Anything).Return(mockDB)
	mockDB.On("Count", mock.Anything).Return(mockDB).Run(func(args mock.Arguments) {
		countPtr := args.Get(0).(*int64)
		*countPtr = 0
	})
	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Create", mock.Anything).Return(errors.New("database error"))
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/customers", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateCustomer(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Failed to create customer", resp.Error)

	mockDB.AssertExpectations(t)
}

func TestCreateCustomer_CommitError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	input := models.Customer{
		Name: "Test Customer",
	}

	mockDB.On("Model", mock.Anything).Return(mockDB)
	mockDB.On("Where", "id = ?", mock.Anything).Return(mockDB)
	mockDB.On("Count", mock.Anything).Return(mockDB).Run(func(args mock.Arguments) {
		countPtr := args.Get(0).(*int64)
		*countPtr = 0
	})
	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Create", mock.Anything).Return(nil)
	mockDB.On("Commit").Return(errors.New("commit failed"))
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/customers", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateCustomer(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Failed to commit transaction", resp.Error)

	mockDB.AssertExpectations(t)
}

func TestGetAllCustomer_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	customers := []models.Customer{
		{ID: "C1", Name: "Customer 1"},
		{ID: "C2", Name: "Customer 2"},
	}

	mockDB.On("Find", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*[]models.Customer)
		*dest = customers
	})

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/customers", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.GetAllCustomer(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NotNil(t, resp["data"])

	mockDB.AssertExpectations(t)
}

func TestGetAllCustomer_NotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("Find", mock.Anything).Return(gorm.ErrRecordNotFound)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/customers", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.GetAllCustomer(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "No customers found", resp.Error)

	mockDB.AssertExpectations(t)
}

func TestGetAllCustomer_DatabaseError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("Find", mock.Anything).Return(errors.New("database error"))

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/customers", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.GetAllCustomer(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Failed to retrieve customers", resp.Error)

	mockDB.AssertExpectations(t)
}

func TestGetCustomer_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	customer := models.Customer{
		ID:   "TEST123",
		Name: "Test Customer",
	}

	mockDB.On("First", mock.Anything, "id = ?", "TEST123").Return(nil).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Customer)
		*dest = customer
	})

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/customers/TEST123", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "TEST123"}}
	c.Set("db", mockDB)

	handler.GetCustomer(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NotNil(t, resp["data"])

	mockDB.AssertExpectations(t)
}

func TestGetCustomer_NotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("First", mock.Anything, "id = ?", "NOTFOUND").Return(gorm.ErrRecordNotFound)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/customers/NOTFOUND", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "NOTFOUND"}}
	c.Set("db", mockDB)

	handler.GetCustomer(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Customer not found", resp.Error)

	mockDB.AssertExpectations(t)
}

func TestGetCustomer_MissingID(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/customers/", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: ""}}
	c.Set("db", mockDB)

	handler.GetCustomer(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Missing ID", resp.Error)
}

func TestGetCustomer_UnknownSpecialCase_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	unknownCustomer := models.Customer{
		ID:   "UNK123",
		Name: "Unknown company",
	}

	mockDB.On("Where", "LOWER(name) = LOWER(?)", "Unknown company").Return(mockDB)
	mockDB.On("First", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Customer)
		*dest = unknownCustomer
	}).Once()
	mockDB.On("First", mock.Anything, "id = ?", "UNK123").Return(nil).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Customer)
		*dest = unknownCustomer
	}).Once()

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/customers/unknown", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "unknown"}}
	c.Set("db", mockDB)

	handler.GetCustomer(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NotNil(t, resp["data"])

	mockDB.AssertExpectations(t)
}

func TestGetCustomer_UnknownSpecialCase_NotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("Where", "LOWER(name) = LOWER(?)", "Unknown company").Return(mockDB)
	mockDB.On("First", mock.Anything).Return(gorm.ErrRecordNotFound)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/customers/unknown", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "unknown"}}
	c.Set("db", mockDB)

	handler.GetCustomer(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Unknown company not found", resp.Error)

	mockDB.AssertExpectations(t)
}

func TestGetCustomer_DatabaseError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("First", mock.Anything, "id = ?", "TEST123").Return(errors.New("database error"))

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/customers/TEST123", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "TEST123"}}
	c.Set("db", mockDB)

	handler.GetCustomer(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Failed to retrieve customer", resp.Error)

	mockDB.AssertExpectations(t)
}

func TestUpdateCustomer_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	existingCustomer := models.Customer{
		ID:   "TEST123",
		Name: "Old Name",
	}

	updatedCustomer := models.Customer{
		ID:   "TEST123",
		Name: "Updated Customer",
	}

	updateData := map[string]string{
		"name": "Updated Customer",
	}

	mockDB.On("Transaction", mock.MatchedBy(func(arg interface{}) bool {
		_, ok := arg.(func(db.Database) error)
		return ok
	})).Return(nil).Run(func(args mock.Arguments) {
		fn := args.Get(0).(func(db.Database) error)
		mockDB.On("First", mock.Anything, "id = ?", "TEST123").Return(nil).Run(func(args mock.Arguments) {
			dest := args.Get(0).(*models.Customer)
			*dest = existingCustomer
		}).Once()
		mockDB.On("Model", mock.Anything).Return(mockDB)
		mockDB.On("Update", "name", "Updated Customer").Return(nil)
		mockDB.On("First", mock.Anything, "id = ?", "TEST123").Return(nil).Run(func(args mock.Arguments) {
			dest := args.Get(0).(*models.Customer)
			*dest = updatedCustomer
		}).Once()
		fn(mockDB)
	})

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", "/customers/TEST123", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "TEST123"}}
	c.Set("db", mockDB)

	handler.UpdateCustomer(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Customer updated successfully", resp["message"])

	mockDB.AssertExpectations(t)
}

func TestUpdateCustomer_NotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	updateData := map[string]string{
		"name": "Updated Customer",
	}

	mockDB.On("Transaction", mock.MatchedBy(func(arg interface{}) bool {
		_, ok := arg.(func(db.Database) error)
		return ok
	})).Return(gorm.ErrRecordNotFound).Run(func(args mock.Arguments) {
		fn := args.Get(0).(func(db.Database) error)
		mockDB.On("First", mock.Anything, "id = ?", "NOTFOUND").Return(gorm.ErrRecordNotFound)
		fn(mockDB)
	})

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", "/customers/NOTFOUND", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "NOTFOUND"}}
	c.Set("db", mockDB)

	handler.UpdateCustomer(c)

	// assert.Equal(t, http.StatusNotFound, w.Code)
	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	// assert.Equal(t, "Customer not found", resp.Error)

	mockDB.AssertExpectations(t)
}

func TestUpdateCustomer_MissingID(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	updateData := map[string]string{
		"name": "Updated Customer",
	}

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", "/customers/", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: ""}}
	c.Set("db", mockDB)

	handler.UpdateCustomer(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Missing ID", resp.Error)
}

func TestUpdateCustomer_InvalidJSON(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("PUT", "/customers/TEST123", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "TEST123"}}
	c.Set("db", mockDB)

	handler.UpdateCustomer(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Invalid input", resp.Error)
}

// func TestUpdateCustomer_WhitespaceOnlyName(t *testing.T) {
// 	mockDB := new(mocks.MockDatabase)

// 	updateData := map[string]string{
// 		"name": "   ",
// 	}

// 	handler := &handlers.Handler{DB: mockDB}

// 	data, _ := json.Marshal(updateData)
// 	req, _ := http.NewRequest("PUT", "/customers/TEST123", bytes.NewBuffer(data))
// 	req.Header.Set("Content-Type", "application/json")
// 	w := httptest.NewRecorder()
// 	c, _ := gin.CreateTestContext(w)
// 	c.Request = req
// 	c.Params = gin.Params{{Key: "id", Value: "TEST123"}}
// 	c.Set("db", mockDB)

// 	handler.UpdateCustomer(c)

// 	assert.Equal(t, http.StatusBadRequest, w.Code)
// 	var resp ErrorResponse
// 	json.Unmarshal(w.Body.Bytes(), &resp)
// 	assert.Equal(t, "Invalid name", resp.Error)
// }

func TestDeleteCustomer_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	customer := models.Customer{
		ID:   "TEST123",
		Name: "Test Customer",
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "TEST123").Return(nil).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Customer)
		*dest = customer
	})
	mockDB.On("Delete", mock.Anything).Return(nil)
	mockDB.On("Commit").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("DELETE", "/customers/TEST123", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "TEST123"}}
	c.Set("db", mockDB)

	handler.DeleteCustomer(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Customer deleted successfully", resp["message"])

	mockDB.AssertExpectations(t)
}

func TestDeleteCustomer_NotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "NOTFOUND").Return(gorm.ErrRecordNotFound)
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("DELETE", "/customers/NOTFOUND", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "NOTFOUND"}}
	c.Set("db", mockDB)

	handler.DeleteCustomer(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Customer not found", resp.Error)

	mockDB.AssertExpectations(t)
}

func TestDeleteCustomer_MissingID(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("DELETE", "/customers/", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: ""}}
	c.Set("db", mockDB)

	handler.DeleteCustomer(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Missing ID", resp.Error)
}

func TestDeleteCustomer_DatabaseError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	customer := models.Customer{
		ID:   "TEST123",
		Name: "Test Customer",
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "TEST123").Return(nil).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Customer)
		*dest = customer
	})
	mockDB.On("Delete", mock.Anything).Return(errors.New("database error"))
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("DELETE", "/customers/TEST123", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "TEST123"}}
	c.Set("db", mockDB)

	handler.DeleteCustomer(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Failed to delete customer", resp.Error)

	mockDB.AssertExpectations(t)
}

func TestDeleteCustomer_CommitError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	customer := models.Customer{
		ID:   "TEST123",
		Name: "Test Customer",
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "TEST123").Return(nil).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Customer)
		*dest = customer
	})
	mockDB.On("Delete", mock.Anything).Return(nil)
	mockDB.On("Commit").Return(errors.New("commit failed"))
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("DELETE", "/customers/TEST123", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "TEST123"}}
	c.Set("db", mockDB)

	handler.DeleteCustomer(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Failed to commit transaction", resp.Error)

	mockDB.AssertExpectations(t)
}
