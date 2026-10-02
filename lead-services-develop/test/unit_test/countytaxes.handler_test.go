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

func TestCreateCountyTax_CreateError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := new(mocks.MockDatabase)

	input := models.CountyTax{
		Name:          "County A",
		TaxPercentage: 5.5,
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Create", mock.Anything).Return(errors.New("create failed"))
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/county-taxes", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateCountyTax(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	assert.Equal(t, "Failed to create county taxes", resp.Error)
	assert.Equal(t, "DATABASE_ERROR", resp.Code)

	mockDB.AssertExpectations(t)
}

func TestCreateCountyTax_CommitError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := new(mocks.MockDatabase)

	input := models.CountyTax{
		Name:          "County A",
		TaxPercentage: 5.5,
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Create", mock.Anything).Return(nil)
	mockDB.On("Commit").Return(errors.New("commit failed"))
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/county-taxes", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateCountyTax(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	assert.Equal(t, "Failed to commit transaction", resp.Error)
	assert.Equal(t, "DATABASE_ERROR", resp.Code)

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

func TestGetCountyTax_MissingID(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/county-taxes/", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.GetCountyTax(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
}

func TestGetCountyTax_NotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("First", mock.Anything, "id = ?", "1").
		Return(gorm.ErrRecordNotFound)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/county-taxes/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.GetCountyTax(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "NOT_FOUND", resp.Code)
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

func TestUpdateCountyTax_MissingID(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("PUT", "/county-taxes/", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.UpdateCountyTax(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateCountyTax_InvalidJSON(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("PUT", "/county-taxes/1", bytes.NewBuffer([]byte("{invalid json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.UpdateCountyTax(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateCountyTax_NotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "1").
		Return(gorm.ErrRecordNotFound)
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	body := `{"name":"Test","tax_percentage":5}`
	req, _ := http.NewRequest("PUT", "/county-taxes/1", bytes.NewBuffer([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.UpdateCountyTax(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockDB.AssertExpectations(t)
}

func TestUpdateCountyTax_SaveError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	existing := models.CountyTax{ID: 1, Name: "Old", TaxPercentage: 5}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "1").
		Run(func(args mock.Arguments) {
			*args.Get(0).(*models.CountyTax) = existing
		}).Return(nil)

	mockDB.On("Save", mock.Anything).
		Return(errors.New("save failed"))
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	body := `{"name":"New","tax_percentage":6}`
	req, _ := http.NewRequest("PUT", "/county-taxes/1", bytes.NewBuffer([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.UpdateCountyTax(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockDB.AssertExpectations(t)
}

func TestUpdateCountyTax_CommitError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	existing := models.CountyTax{ID: 1}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "1").
		Run(func(args mock.Arguments) {
			*args.Get(0).(*models.CountyTax) = existing
		}).Return(nil)

	mockDB.On("Save", mock.Anything).Return(nil)
	mockDB.On("Commit").
		Return(errors.New("commit failed"))
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	body := `{"name":"Updated","tax_percentage":7}`
	req, _ := http.NewRequest("PUT", "/county-taxes/1", bytes.NewBuffer([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.UpdateCountyTax(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockDB.AssertExpectations(t)
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

func TestDeleteCountyTax_MissingID(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("DELETE", "/county-taxes/", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.DeleteCountyTax(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteCountyTax_NotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "1").
		Return(gorm.ErrRecordNotFound)
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("DELETE", "/county-taxes/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.DeleteCountyTax(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockDB.AssertExpectations(t)
}

func TestDeleteCountyTax_FirstDBError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "1").
		Return(errors.New("db error"))
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("DELETE", "/county-taxes/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.DeleteCountyTax(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockDB.AssertExpectations(t)
}

func TestDeleteCountyTax_DeleteError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	existing := models.CountyTax{ID: 1}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "1").
		Run(func(args mock.Arguments) {
			*args.Get(0).(*models.CountyTax) = existing
		}).Return(nil)

	mockDB.On("Delete", mock.Anything).
		Return(errors.New("delete failed"))
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("DELETE", "/county-taxes/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.DeleteCountyTax(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockDB.AssertExpectations(t)
}

func TestDeleteCountyTax_CommitError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	existing := models.CountyTax{ID: 1}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "1").
		Run(func(args mock.Arguments) {
			*args.Get(0).(*models.CountyTax) = existing
		}).Return(nil)

	mockDB.On("Delete", mock.Anything).Return(nil)
	mockDB.On("Commit").
		Return(errors.New("commit failed"))
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("DELETE", "/county-taxes/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.DeleteCountyTax(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockDB.AssertExpectations(t)
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

func TestGetAllCountyTaxes_CountError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("Model", mock.AnythingOfType("*models.CountyTax")).Return(mockDB)
	mockDB.On("Count", mock.AnythingOfType("*int64")).Return(mockDB)
	mockDB.On("Error").Return(errors.New("count failed"))

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/county-taxes", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.GetAllCountyTaxes(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockDB.AssertExpectations(t)
}

func TestGetAllCountyTaxes_AllTrue_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	countyList := []models.CountyTax{
		{ID: 1, Name: "A"},
		{ID: 2, Name: "B"},
	}

	var total int64 = 2

	mockDB.On("Model", mock.AnythingOfType("*models.CountyTax")).Return(mockDB)
	mockDB.On("Count", mock.AnythingOfType("*int64")).Run(func(args mock.Arguments) {
		*args.Get(0).(*int64) = total
	}).Return(mockDB)
	mockDB.On("Error").Return(nil)

	mockDB.On("Find", mock.AnythingOfType("*[]models.CountyTax")).
		Run(func(args mock.Arguments) {
			*args.Get(0).(*[]models.CountyTax) = countyList
		}).Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/county-taxes?all=true", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.GetAllCountyTaxes(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockDB.AssertExpectations(t)
}

func TestGetAllCountyTaxes_AllTrue_FindError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("Model", mock.AnythingOfType("*models.CountyTax")).Return(mockDB)
	mockDB.On("Count", mock.AnythingOfType("*int64")).Return(mockDB)
	mockDB.On("Error").Return(nil)

	mockDB.On("Find", mock.AnythingOfType("*[]models.CountyTax")).
		Return(errors.New("find failed"))

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/county-taxes?all=true", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.GetAllCountyTaxes(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockDB.AssertExpectations(t)
}

func TestGetAllCountyTaxes_PaginatedFindError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("Model", mock.AnythingOfType("*models.CountyTax")).Return(mockDB)
	mockDB.On("Count", mock.AnythingOfType("*int64")).Return(mockDB)
	mockDB.On("Error").Return(nil)

	mockDB.On("Offset", mock.Anything).Return(mockDB)
	mockDB.On("Limit", mock.Anything).Return(mockDB)
	mockDB.On("Find", mock.AnythingOfType("*[]models.CountyTax")).
		Return(errors.New("find failed"))

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/county-taxes?page=1&limit=10", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.GetAllCountyTaxes(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockDB.AssertExpectations(t)
}
