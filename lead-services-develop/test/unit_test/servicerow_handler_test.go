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

// ----------------- CreateServiceRow -----------------
func TestCreateServiceRow_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	input := models.ServiceRow{
		Details: models.ServiceRowDetails{
			Name:               "Basic Service",
			PackageName:        "Standard Package",
			ServiceGroupName:   "Group A",
			AdditionalDiscount: 5,
			Description:        "Test service description",
			UnitPrice:          1000,
			Total:              950,
			IsCustom:           false,
		},
		EstimateID:         "EST123",
		ResidentialEntryID: 1,
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Create", mock.Anything).Return(nil)
	mockDB.On("Commit").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/service-rows", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateServiceRow(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "ServiceRow creation successfully", resp["message"])

	mockDB.AssertExpectations(t)
}

func TestCreateServiceRow_InvalidInput(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("POST", "/service-rows", bytes.NewBuffer([]byte(`invalid json`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateServiceRow(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
	assert.Equal(t, "Invalid input", resp.Error)
}

func TestCreateServiceRow_DatabaseError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	input := models.ServiceRow{
		Details: models.ServiceRowDetails{
			Name:      "Basic Service",
			UnitPrice: 1000,
			Total:     1000,
		},
		EstimateID:         "EST123",
		ResidentialEntryID: 1,
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Create", mock.Anything).Return(gorm.ErrInvalidData)
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/service-rows", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateServiceRow(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "DATABASE_ERROR", resp.Code)
	assert.Equal(t, "Failed to create service", resp.Error)

	mockDB.AssertExpectations(t)
}

func TestCreateServiceRow_CommitError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	input := models.ServiceRow{
		Details: models.ServiceRowDetails{
			Name:      "Basic Service",
			UnitPrice: 1000,
			Total:     1000,
		},
		EstimateID:         "EST123",
		ResidentialEntryID: 1,
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Create", mock.Anything).Return(nil)
	mockDB.On("Commit").Return(gorm.ErrInvalidTransaction)
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/service-rows", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateServiceRow(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "DATABASE_ERROR", resp.Code)
	assert.Equal(t, "Failed to commit transaction", resp.Error)

	mockDB.AssertExpectations(t)
}

// ----------------- GetServiceRow -----------------
func TestGetServiceRow_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	estimateID := "EST123"

	serviceRows := []models.ServiceRow{
		{
			ID: 1,
			Details: models.ServiceRowDetails{
				Name:      "Service 1",
				UnitPrice: 1000,
				Total:     1000,
			},
			EstimateID:         estimateID,
			ResidentialEntryID: 1,
		},
		{
			ID: 2,
			Details: models.ServiceRowDetails{
				Name:      "Service 2",
				UnitPrice: 2000,
				Total:     2000,
			},
			EstimateID:         estimateID,
			ResidentialEntryID: 2,
		},
	}

	residentialEntry1 := models.ResidentialEntry{
		ID:                        1,
		ManagerResidentialEntryID: 10,
		ManagerResidentialTypeID:  20,
		Details:                   "Residential details 1",
	}

	residentialEntry2 := models.ResidentialEntry{
		ID:                        2,
		ManagerResidentialEntryID: 11,
		ManagerResidentialTypeID:  21,
		Details:                   "Residential details 2",
	}

	mockDB.On("Where", "estimate_id = ?", estimateID).Return(mockDB)
	mockDB.On("Find", mock.AnythingOfType("*[]models.ServiceRow")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*[]models.ServiceRow)
		*dest = serviceRows
	}).Return(nil)

	// Mock the Preload chains for each residential entry
	mockDB.On("Preload", "ManagerResidentialEntry").Return(mockDB).Times(2)
	mockDB.On("Preload", "ManagerResidentialType").Return(mockDB).Times(2)
	mockDB.On("Where", "id = ?", uint(1)).Return(mockDB).Once()
	mockDB.On("Find", mock.AnythingOfType("*models.ResidentialEntry")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.ResidentialEntry)
		*dest = residentialEntry1
	}).Return(nil).Once()

	mockDB.On("Where", "id = ?", uint(2)).Return(mockDB).Once()
	mockDB.On("Find", mock.AnythingOfType("*models.ResidentialEntry")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.ResidentialEntry)
		*dest = residentialEntry2
	}).Return(nil).Once()

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/service-rows/"+estimateID, nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: estimateID}}
	c.Set("db", mockDB)

	handler.GetServiceRow(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	data, ok := resp["data"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(data))

	mockDB.AssertExpectations(t)
}

func TestGetServiceRow_MissingID(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/service-rows/", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: ""}}
	c.Set("db", mockDB)

	handler.GetServiceRow(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
	assert.Equal(t, "Missing ID", resp.Error)
}

func TestGetServiceRow_ResidentialEntryNotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	estimateID := "EST123"

	serviceRows := []models.ServiceRow{
		{
			ID:                 1,
			EstimateID:         estimateID,
			ResidentialEntryID: 999,
		},
	}

	mockDB.On("Where", "estimate_id = ?", estimateID).Return(mockDB)
	mockDB.On("Find", mock.AnythingOfType("*[]models.ServiceRow")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*[]models.ServiceRow)
		*dest = serviceRows
	}).Return(nil)

	mockDB.On("Preload", "ManagerResidentialEntry").Return(mockDB)
	mockDB.On("Preload", "ManagerResidentialType").Return(mockDB)
	mockDB.On("Where", "id = ?", uint(999)).Return(mockDB)
	mockDB.On("Find", mock.AnythingOfType("*models.ResidentialEntry")).Return(gorm.ErrRecordNotFound)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/service-rows/"+estimateID, nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: estimateID}}
	c.Set("db", mockDB)

	handler.GetServiceRow(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "NOT_FOUND", resp.Code)

	mockDB.AssertExpectations(t)
}

// ----------------- UpdateServiceRow -----------------
func TestUpdateServiceRow_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	serviceRowID := "1"
	existing := models.ServiceRow{
		ID: 1,
		Details: models.ServiceRowDetails{
			Name:      "Old Service",
			UnitPrice: 1000,
			Total:     1000,
		},
		EstimateID:         "EST123",
		ResidentialEntryID: 1,
	}

	input := models.ServiceRow{
		Details: models.ServiceRowDetails{
			Name:      "Updated Service",
			UnitPrice: 1500,
			Total:     1500,
		},
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", serviceRowID).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.ServiceRow)
		*dest = existing
	}).Return(nil)
	mockDB.On("Save", mock.Anything).Return(nil)
	mockDB.On("Commit").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("PUT", "/service-rows/"+serviceRowID, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: serviceRowID}}
	c.Set("db", mockDB)

	handler.UpdateServiceRow(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "ServiceRow updated successfully", resp["message"])

	mockDB.AssertExpectations(t)
}

func TestUpdateServiceRow_MissingID(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	input := models.ServiceRow{
		Details: models.ServiceRowDetails{
			Name: "Updated",
		},
	}
	data, _ := json.Marshal(input)

	req, _ := http.NewRequest("PUT", "/service-rows/", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: ""}}
	c.Set("db", mockDB)

	handler.UpdateServiceRow(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
	assert.Equal(t, "Missing ID", resp.Error)
}

func TestUpdateServiceRow_InvalidInput(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("PUT", "/service-rows/1", bytes.NewBuffer([]byte(`invalid json`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.UpdateServiceRow(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
}

func TestUpdateServiceRow_NotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	input := models.ServiceRow{
		Details: models.ServiceRowDetails{
			Name: "Updated",
		},
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "999").Return(gorm.ErrRecordNotFound)
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("PUT", "/service-rows/999", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "999"}}
	c.Set("db", mockDB)

	handler.UpdateServiceRow(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "NOT_FOUND", resp.Code)

	mockDB.AssertExpectations(t)
}

func TestUpdateServiceRow_SaveError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	serviceRowID := "1"
	existing := models.ServiceRow{ID: 1}
	input := models.ServiceRow{
		Details: models.ServiceRowDetails{
			Name: "Updated",
		},
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", serviceRowID).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.ServiceRow)
		*dest = existing
	}).Return(nil)
	mockDB.On("Save", mock.Anything).Return(gorm.ErrInvalidData)
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("PUT", "/service-rows/"+serviceRowID, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: serviceRowID}}
	c.Set("db", mockDB)

	handler.UpdateServiceRow(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "DATABASE_ERROR", resp.Code)

	mockDB.AssertExpectations(t)
}

// ----------------- DeleteServiceRow -----------------
func TestDeleteServiceRow_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	serviceRowID := "1"
	existing := models.ServiceRow{
		ID: 1,
		Details: models.ServiceRowDetails{
			Name: "Service to Delete",
		},
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", serviceRowID).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.ServiceRow)
		*dest = existing
	}).Return(nil)
	mockDB.On("Delete", mock.Anything).Return(nil)
	mockDB.On("Commit").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("DELETE", "/service-rows/"+serviceRowID, nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: serviceRowID}}
	c.Set("db", mockDB)

	handler.DeleteServiceRow(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Residential entry deleted successfully", resp["message"])

	mockDB.AssertExpectations(t)
}

func TestDeleteServiceRow_MissingID(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("DELETE", "/service-rows/", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: ""}}
	c.Set("db", mockDB)

	handler.DeleteServiceRow(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
	assert.Equal(t, "Missing ID", resp.Error)
}

func TestDeleteServiceRow_NotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "999").Return(gorm.ErrRecordNotFound)
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("DELETE", "/service-rows/999", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "999"}}
	c.Set("db", mockDB)

	handler.DeleteServiceRow(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "NOT_FOUND", resp.Code)

	mockDB.AssertExpectations(t)
}

func TestDeleteServiceRow_DeleteError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	serviceRowID := "1"
	existing := models.ServiceRow{ID: 1}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", serviceRowID).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.ServiceRow)
		*dest = existing
	}).Return(nil)
	mockDB.On("Delete", mock.Anything).Return(gorm.ErrInvalidData)
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("DELETE", "/service-rows/"+serviceRowID, nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: serviceRowID}}
	c.Set("db", mockDB)

	handler.DeleteServiceRow(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "DATABASE_ERROR", resp.Code)
	assert.Equal(t, "Failed to delete residential entry", resp.Error)

	mockDB.AssertExpectations(t)
}
