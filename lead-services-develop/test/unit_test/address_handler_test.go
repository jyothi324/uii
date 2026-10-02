package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/tuliporg/agent-management-api/handlers"
	"github.com/tuliporg/agent-management-api/mocks"
	"github.com/tuliporg/agent-management-api/models"
	"gorm.io/gorm"
)

// ----------------- CreateAddress -----------------
func TestCreateAddress_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	input := models.Address{
		State:     "California",
		County:    "Los Angeles",
		Street:    "123 Main St",
		UnitSuite: "Apt 4B",
		City:      "Los Angeles",
		Zip:       "90001",
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Create", mock.Anything).Return(nil)
	mockDB.On("Commit").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/addresses", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateAddress(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp models.Address
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, input.State, resp.State)
	assert.Equal(t, input.City, resp.City)
	assert.Equal(t, input.Zip, resp.Zip)

	mockDB.AssertExpectations(t)
}

func TestCreateAddress_InvalidJSON(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("POST", "/addresses", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateAddress(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
	assert.Equal(t, "Invalid request body", resp.Error)
}

func TestCreateAddress_DatabaseError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	input := models.Address{
		State:  "California",
		City:   "Los Angeles",
		Zip:    "90001",
		Street: "123 Main St",
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Create", mock.Anything).Return(gorm.ErrInvalidData)
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/addresses", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateAddress(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "DATABASE_ERROR", resp.Code)
	assert.Equal(t, "Failed to create address", resp.Error)

	mockDB.AssertExpectations(t)
}

func TestCreateAddress_CommitError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	input := models.Address{
		State:  "California",
		City:   "Los Angeles",
		Zip:    "90001",
		Street: "123 Main St",
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Create", mock.Anything).Return(nil)
	mockDB.On("Commit").Return(gorm.ErrInvalidTransaction)
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/addresses", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateAddress(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "DATABASE_ERROR", resp.Code)
	assert.Equal(t, "Failed to commit transaction", resp.Error)

	mockDB.AssertExpectations(t)
}

// ----------------- GetAddress -----------------
func TestGetAddress_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	expected := models.Address{
		ID:        1,
		State:     "California",
		County:    "Los Angeles",
		Street:    "123 Main St",
		UnitSuite: "Apt 4B",
		City:      "Los Angeles",
		Zip:       "90001",
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "1").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Address)
		*dest = expected
	}).Return(nil)
	mockDB.On("Commit").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/addresses/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.GetAddress(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.Address
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, expected.State, resp.State)
	assert.Equal(t, expected.City, resp.City)
	assert.Equal(t, expected.Zip, resp.Zip)

	mockDB.AssertExpectations(t)
}

func TestGetAddress_MissingID(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/addresses/", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: ""}}
	c.Set("db", mockDB)

	handler.GetAddress(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
	assert.Equal(t, "Missing ID", resp.Error)
}

func TestGetAddress_NotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "999").Return(gorm.ErrRecordNotFound)
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/addresses/999", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "999"}}
	c.Set("db", mockDB)

	handler.GetAddress(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "NOT_FOUND", resp.Code)
	assert.Equal(t, "Address not found", resp.Error)

	mockDB.AssertExpectations(t)
}

func TestGetAddress_DatabaseError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "1").Return(gorm.ErrInvalidDB)
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/addresses/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.GetAddress(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "DATABASE_ERROR", resp.Code)

	mockDB.AssertExpectations(t)
}

// ----------------- GetAllAddresses -----------------
func TestGetAllAddresses_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	addressList := []models.Address{
		{ID: 1, State: "California", City: "Los Angeles", Zip: "90001"},
		{ID: 2, State: "New York", City: "New York", Zip: "10001"},
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Find", mock.AnythingOfType("*[]models.Address")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*[]models.Address)
		*dest = addressList
	}).Return(nil)
	mockDB.On("Commit").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/addresses", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.GetAllAddresses(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp []models.Address
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, 2, len(resp))
	assert.Equal(t, "California", resp[0].State)
	assert.Equal(t, "New York", resp[1].State)

	mockDB.AssertExpectations(t)
}

func TestGetAllAddresses_EmptyList(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	addressList := []models.Address{}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Find", mock.AnythingOfType("*[]models.Address")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*[]models.Address)
		*dest = addressList
	}).Return(nil)
	mockDB.On("Commit").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/addresses", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.GetAllAddresses(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp []models.Address
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, 0, len(resp))

	mockDB.AssertExpectations(t)
}

func TestGetAllAddresses_DatabaseError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Find", mock.AnythingOfType("*[]models.Address")).Return(gorm.ErrInvalidDB)
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/addresses", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.GetAllAddresses(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "DATABASE_ERROR", resp.Code)
	assert.Equal(t, "Failed to retrieve addresses", resp.Error)

	mockDB.AssertExpectations(t)
}

// ----------------- UpdateAddress -----------------
func TestUpdateAddress_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	existing := models.Address{
		ID:     1,
		State:  "California",
		City:   "Los Angeles",
		Zip:    "90001",
		Street: "123 Main St",
	}

	updated := models.Address{
		State:  "New York",
		City:   "New York",
		Zip:    "10001",
		Street: "456 Park Ave",
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "1").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Address)
		*dest = existing
	}).Return(nil)
	mockDB.On("Save", mock.Anything).Return(nil)
	mockDB.On("Commit").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(updated)
	req, _ := http.NewRequest("PUT", "/addresses/1", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.UpdateAddress(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.Address
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, updated.State, resp.State)
	assert.Equal(t, updated.City, resp.City)

	mockDB.AssertExpectations(t)
}

func TestUpdateAddress_MissingID(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	input := models.Address{State: "California", City: "Los Angeles"}
	data, _ := json.Marshal(input)

	req, _ := http.NewRequest("PUT", "/addresses/", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: ""}}
	c.Set("db", mockDB)

	handler.UpdateAddress(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
	assert.Equal(t, "Missing ID", resp.Error)
}

func TestUpdateAddress_NotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	input := models.Address{State: "California", City: "Los Angeles"}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "999").Return(gorm.ErrRecordNotFound)
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("PUT", "/addresses/999", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "999"}}
	c.Set("db", mockDB)

	handler.UpdateAddress(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "NOT_FOUND", resp.Code)
	assert.Equal(t, "Address not found", resp.Error)

	mockDB.AssertExpectations(t)
}

func TestUpdateAddress_InvalidJSON(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	existing := models.Address{ID: 1, State: "California"}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "1").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Address)
		*dest = existing
	}).Return(nil)
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("PUT", "/addresses/1", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.UpdateAddress(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
	assert.Equal(t, "Invalid request body", resp.Error)

	mockDB.AssertExpectations(t)
}

func TestUpdateAddress_SaveError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	existing := models.Address{ID: 1, State: "California"}
	updated := models.Address{State: "New York", City: "New York"}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "1").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Address)
		*dest = existing
	}).Return(nil)
	mockDB.On("Save", mock.Anything).Return(gorm.ErrInvalidData)
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(updated)
	req, _ := http.NewRequest("PUT", "/addresses/1", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.UpdateAddress(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "DATABASE_ERROR", resp.Code)
	assert.Equal(t, "Failed to update address", resp.Error)

	mockDB.AssertExpectations(t)
}

// ----------------- DeleteAddress -----------------
func TestDeleteAddress_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	existing := models.Address{
		ID:    1,
		State: "California",
		City:  "Los Angeles",
		Zip:   "90001",
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "1").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Address)
		*dest = existing
	}).Return(nil)
	mockDB.On("Delete", mock.Anything).Return(nil)
	mockDB.On("Commit").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("DELETE", "/addresses/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.DeleteAddress(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "address deleted successfully", resp["message"])

	mockDB.AssertExpectations(t)
}

func TestDeleteAddress_MissingID(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("DELETE", "/addresses/", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: ""}}
	c.Set("db", mockDB)

	handler.DeleteAddress(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
	assert.Equal(t, "Missing ID", resp.Error)
}

func TestDeleteAddress_NotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "999").Return(gorm.ErrRecordNotFound)
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("DELETE", "/addresses/999", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "999"}}
	c.Set("db", mockDB)

	handler.DeleteAddress(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "NOT_FOUND", resp.Code)
	assert.Equal(t, "Address not found", resp.Error)

	mockDB.AssertExpectations(t)
}

func TestDeleteAddress_DeleteError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	existing := models.Address{ID: 1, State: "California"}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "1").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Address)
		*dest = existing
	}).Return(nil)
	mockDB.On("Delete", mock.Anything).Return(gorm.ErrInvalidData)
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("DELETE", "/addresses/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.DeleteAddress(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "DATABASE_ERROR", resp.Code)
	assert.Equal(t, "Failed to delete address", resp.Error)

	mockDB.AssertExpectations(t)
}

func TestDeleteAddress_CommitError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	existing := models.Address{ID: 1, State: "California"}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "1").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Address)
		*dest = existing
	}).Return(nil)
	mockDB.On("Delete", mock.Anything).Return(nil)
	mockDB.On("Commit").Return(gorm.ErrInvalidTransaction)
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("DELETE", "/addresses/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.DeleteAddress(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "DATABASE_ERROR", resp.Code)
	assert.Equal(t, "Failed to commit transaction", resp.Error)

	mockDB.AssertExpectations(t)
}
