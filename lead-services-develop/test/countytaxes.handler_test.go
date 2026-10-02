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
)

// ----------------- CreateCountyTax -----------------
func TestCreateCountyTax_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	input := models.CountyTax{
		Name:          "County A",
		TaxPercentage: 5.5,
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Create", mock.Anything).Return(nil)
	mockDB.On("Commit").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/county-taxes", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateCountyTax(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "CountyTax creation successfully", resp["message"])

	mockDB.AssertExpectations(t)
}

func TestCreateCountyTax_InvalidInput(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("POST", "/county-taxes", bytes.NewBuffer([]byte(`invalid json`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateCountyTax(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
}

// ----------------- GetCountyTax -----------------
func TestGetCountyTax_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	expected := models.CountyTax{ID: 1, Name: "County A", TaxPercentage: 5.5}

	mockDB.On("First", mock.Anything, "id = ?", "1").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.CountyTax)
		*dest = expected
	}).Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/county-taxes/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.GetCountyTax(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]models.CountyTax
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, expected.Name, resp["data"].Name)
}

// ----------------- UpdateCountyTax -----------------
func TestUpdateCountyTax_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	existing := models.CountyTax{ID: 1, Name: "County A", TaxPercentage: 5.0}
	updated := models.CountyTax{ID: 1, Name: "County A Updated", TaxPercentage: 6.0}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "1").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.CountyTax)
		*dest = existing
	}).Return(nil)
	mockDB.On("Save", mock.Anything).Return(nil)
	mockDB.On("Commit").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(updated)
	req, _ := http.NewRequest("PUT", "/county-taxes/1", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.UpdateCountyTax(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]models.CountyTax
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, updated.Name, resp["data"].Name)
}

// ----------------- DeleteCountyTax -----------------
func TestDeleteCountyTax_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	existing := models.CountyTax{ID: 1, Name: "County A", TaxPercentage: 5.0}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "1").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.CountyTax)
		*dest = existing
	}).Return(nil)
	mockDB.On("Delete", mock.Anything).Return(nil)
	mockDB.On("Commit").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("DELETE", "/county-taxes/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.DeleteCountyTax(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ----------------- GetAllCountyTaxes -----------------
func TestGetAllCountyTaxes_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	countyList := []models.CountyTax{
		{ID: 1, Name: "County A", TaxPercentage: 5.0},
		{ID: 2, Name: "County B", TaxPercentage: 6.0},
	}

	var total int64 = 2

	// Mock the Model call
	mockDB.On("Model", mock.AnythingOfType("*models.CountyTax")).Return(mockDB)

	// Mock the Count call
	mockDB.On("Count", mock.AnythingOfType("*int64")).Run(func(args mock.Arguments) {
		countPtr := args.Get(0).(*int64)
		*countPtr = total
	}).Return(mockDB) // Return mockDB, not nil
	mockDB.On("Error").Return(nil).Once()

	// Mock the Offset call
	mockDB.On("Offset", mock.AnythingOfType("int")).Return(mockDB)

	// Mock the Limit call
	mockDB.On("Limit", mock.AnythingOfType("int")).Return(mockDB)

	// Mock the Find call
	mockDB.On("Find", mock.AnythingOfType("*[]models.CountyTax")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*[]models.CountyTax)
		*dest = countyList
	}).Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	// Add query parameters for pagination
	req, _ := http.NewRequest("GET", "/county-taxes?page=1&limit=10", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.GetAllCountyTaxes(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(total), resp["total"])

	// Verify the county taxes are in the response
	data, ok := resp["data"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(data))

	// Verify all expectations were met
	mockDB.AssertExpectations(t)
}
