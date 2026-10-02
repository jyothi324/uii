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
	"github.com/tuliporg/agent-management-api/handlers"
	"github.com/tuliporg/agent-management-api/mocks"
	"github.com/tuliporg/agent-management-api/models"
	"gorm.io/gorm"
)

// ----------------- CreatePickup -----------------
func TestCreatePickup_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	input := models.PickupLocations{
		Name:    "Main Warehouse",
		Street:  "123 Main St",
		City:    "New York",
		State:   "NY",
		Zip:     "10001",
		Country: "USA",
	}

	// Mock DB flow
	mockDB.On("Begin").Return(mockDB).Once()
	mockDB.On("Create", mock.AnythingOfType("*models.PickupLocations")).Return(nil).Once()
	mockDB.On("Commit").Return(nil).Once()

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/pickup-locations", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreatePickup(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "location creation successfully", resp["message"])

	mockDB.AssertExpectations(t)
}

func TestCreatePickup_InvalidInput(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("POST", "/pickup-locations", bytes.NewBuffer([]byte(`invalid json`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreatePickup(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
}

func TestCreatePickup_DatabaseError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	input := models.PickupLocations{
		Name: "Test Location",
		City: "Test City",
	}

	mockDB.On("Begin").Return(mockDB).Once()
	mockDB.On("Create", mock.AnythingOfType("*models.PickupLocations")).Return(errors.New("database error")).Once()
	mockDB.On("Rollback").Return(nil).Once()

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/pickup-locations", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreatePickup(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "DATABASE_ERROR", resp.Code)

	mockDB.AssertExpectations(t)
}

// ----------------- GetPickup -----------------
func TestGetPickup_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	expected := models.PickupLocations{
		ID:      1,
		Name:    "Main Warehouse",
		Street:  "123 Main St",
		City:    "New York",
		State:   "NY",
		Zip:     "10001",
		Country: "USA",
	}

	mockDB.On("First", mock.AnythingOfType("*models.PickupLocations"), "id = ?", "1").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.PickupLocations)
		*dest = expected
	}).Return(nil).Once()

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/pickup-locations/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.GetPickup(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]models.PickupLocations
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, expected.Name, resp["data"].Name)

	mockDB.AssertExpectations(t)
}

func TestGetPickup_NotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("First", mock.AnythingOfType("*models.PickupLocations"), "id = ?", "999").Return(gorm.ErrRecordNotFound).Once()

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/pickup-locations/999", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "999"}}
	c.Set("db", mockDB)

	handler.GetPickup(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "NOT_FOUND", resp.Code)

	mockDB.AssertExpectations(t)
}

func TestGetPickup_MissingID(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/pickup-locations/", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: ""}}
	c.Set("db", mockDB)

	handler.GetPickup(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
}

// ----------------- UpdatePickup -----------------
func TestUpdatePickup_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	existing := models.PickupLocations{
		ID:      1,
		Name:    "Old Name",
		City:    "Old City",
		State:   "NY",
		Zip:     "10001",
		Country: "USA",
	}

	updated := models.PickupLocations{
		Name:    "New Name",
		City:    "New City",
		State:   "CA",
		Zip:     "90001",
		Country: "USA",
	}

	mockDB.On("Begin").Return(mockDB).Once()
	mockDB.On("Error").Return(nil).Once()

	mockDB.On("Where", "id = ?", "1").Return(mockDB).Once()
	mockDB.On("First", mock.AnythingOfType("*models.PickupLocations")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.PickupLocations)
		*dest = existing
	}).Return(nil).Once()

	mockDB.On("Save", mock.AnythingOfType("*models.PickupLocations")).Return(nil).Once()

	mockDB.On("Commit").Return(nil).Once()

	// Mock the final fetch
	mockDB.On("Where", "id = ?", "1").Return(mockDB).Once()
	mockDB.On("First", mock.AnythingOfType("*models.PickupLocations")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.PickupLocations)
		dest.ID = 1
		dest.Name = updated.Name
		dest.City = updated.City
	}).Return(nil).Once()

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(updated)
	req, _ := http.NewRequest("PUT", "/pickup-locations/1", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.UpdatePickup(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Pickup location updated successfully", resp["message"])

	mockDB.AssertExpectations(t)
}

func TestUpdatePickup_NotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	input := models.PickupLocations{
		Name: "New Name",
	}

	mockDB.On("Begin").Return(mockDB).Once()
	mockDB.On("Error").Return(nil).Once()

	mockDB.On("Where", "id = ?", "999").Return(mockDB).Once()
	mockDB.On("First", mock.AnythingOfType("*models.PickupLocations")).Return(gorm.ErrRecordNotFound).Once()
	mockDB.On("Rollback").Return(nil).Once()

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("PUT", "/pickup-locations/999", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "999"}}
	c.Set("db", mockDB)

	handler.UpdatePickup(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "NOT_FOUND", resp.Code)

	mockDB.AssertExpectations(t)
}

func TestUpdatePickup_NoFieldsProvided(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	input := models.PickupLocations{} // Empty input

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("PUT", "/pickup-locations/1", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.UpdatePickup(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
}

// ----------------- DeletePickup -----------------
func TestDeletePickup_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	existing := models.PickupLocations{
		ID:   1,
		Name: "Test Location",
		City: "Test City",
	}

	mockDB.On("Begin").Return(mockDB).Once()
	mockDB.On("First", mock.AnythingOfType("*models.PickupLocations"), "id = ?", "1").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.PickupLocations)
		*dest = existing
	}).Return(nil).Once()

	mockDB.On("Delete", mock.AnythingOfType("*models.PickupLocations")).Return(nil).Once()

	mockDB.On("Commit").Return(nil).Once()

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("DELETE", "/pickup-locations/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.DeletePickup(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "location deleted successfully", resp["message"])

	mockDB.AssertExpectations(t)
}

func TestDeletePickup_NotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("Begin").Return(mockDB).Once()
	mockDB.On("First", mock.AnythingOfType("*models.PickupLocations"), "id = ?", "999").Return(gorm.ErrRecordNotFound).Once()
	mockDB.On("Rollback").Return(nil).Once()

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("DELETE", "/pickup-locations/999", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "999"}}
	c.Set("db", mockDB)

	handler.DeletePickup(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "NOT_FOUND", resp.Code)

	mockDB.AssertExpectations(t)
}

// ----------------- ListPickupLocations -----------------
func TestListPickupLocations_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	pickupList := []models.PickupLocations{
		{ID: 1, Name: "Location A", City: "New York", State: "NY"},
		{ID: 2, Name: "Location B", City: "Los Angeles", State: "CA"},
	}

	var total int64 = 2

	mockDB.On("Model", mock.AnythingOfType("*models.PickupLocations")).Return(mockDB).Once()
	mockDB.On("Count", mock.AnythingOfType("*int64")).Run(func(args mock.Arguments) {
		countPtr := args.Get(0).(*int64)
		*countPtr = total
	}).Return(mockDB).Once()
	mockDB.On("Error").Return(nil).Once()

	mockDB.On("Offset", mock.AnythingOfType("int")).Return(mockDB).Once()
	mockDB.On("Limit", mock.AnythingOfType("int")).Return(mockDB).Once()
	mockDB.On("Find", mock.AnythingOfType("*[]models.PickupLocations")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*[]models.PickupLocations)
		*dest = pickupList
	}).Return(nil).Once()

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/pickup-locations?page=1&limit=10", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.ListPickupLocations(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(total), resp["total"])

	data, ok := resp["data"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(data))

	mockDB.AssertExpectations(t)
}

func TestListPickupLocations_AllRecords(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	pickupList := []models.PickupLocations{
		{ID: 1, Name: "Location A", City: "New York"},
		{ID: 2, Name: "Location B", City: "Los Angeles"},
		{ID: 3, Name: "Location C", City: "Chicago"},
	}

	var total int64 = 3

	mockDB.On("Model", mock.AnythingOfType("*models.PickupLocations")).Return(mockDB).Once()
	mockDB.On("Count", mock.AnythingOfType("*int64")).Run(func(args mock.Arguments) {
		countPtr := args.Get(0).(*int64)
		*countPtr = total
	}).Return(mockDB).Once()
	mockDB.On("Error").Return(nil).Once()

	mockDB.On("Find", mock.AnythingOfType("*[]models.PickupLocations")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*[]models.PickupLocations)
		*dest = pickupList
	}).Return(nil).Once()

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/pickup-locations?all=true", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.ListPickupLocations(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(total), resp["total"])
	assert.Equal(t, true, resp["all"])

	data, ok := resp["data"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 3, len(data))

	mockDB.AssertExpectations(t)
}

func TestListPickupLocations_CountError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("Model", mock.AnythingOfType("*models.PickupLocations")).Return(mockDB).Once()
	mockDB.On("Count", mock.AnythingOfType("*int64")).Return(mockDB).Once()
	mockDB.On("Error").Return(errors.New("database error")).Once()

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/pickup-locations?page=1&limit=10", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.ListPickupLocations(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "DATABASE_ERROR", resp.Code)

	mockDB.AssertExpectations(t)
}
