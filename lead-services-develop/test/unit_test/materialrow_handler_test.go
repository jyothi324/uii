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

// ----------------- CreateMaterialRow -----------------
func TestCreateMaterialRow_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	input := models.MaterialRow{
		Details: models.MaterialRowDetails{
			Description:        "Test Material",
			Material:           "Wood",
			Category:           "Building Materials",
			SubCategory:        "Lumber",
			Quantity:           10,
			AdditionalDiscount: 5,
			UnitPrice:          100,
			Total:              1000,
			IsCustom:           false,
		},
		EstimateID:         "EST001",
		ResidentialEntryID: 1,
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Create", mock.Anything).Return(nil)
	mockDB.On("Commit").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/material-rows", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateMaterialRow(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "MaterialRow creation successfully", resp["message"])
	assert.NotNil(t, resp["data"])

	mockDB.AssertExpectations(t)
}

func TestCreateMaterialRow_InvalidInput(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("POST", "/material-rows", bytes.NewBuffer([]byte(`invalid json`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateMaterialRow(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
	assert.Equal(t, "Invalid input", resp.Error)
}

func TestCreateMaterialRow_DatabaseError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	input := models.MaterialRow{
		Details: models.MaterialRowDetails{
			Description: "Test Material",
			Material:    "Wood",
			Category:    "Building Materials",
			SubCategory: "Lumber",
		},
		EstimateID:         "EST001",
		ResidentialEntryID: 1,
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Create", mock.Anything).Return(gorm.ErrInvalidTransaction)
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/material-rows", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateMaterialRow(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "DATABASE_ERROR", resp.Code)
	assert.Equal(t, "Failed to create material", resp.Error)

	mockDB.AssertExpectations(t)
}

func TestCreateMaterialRow_CommitError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	input := models.MaterialRow{
		Details: models.MaterialRowDetails{
			Description: "Test Material",
			Material:    "Wood",
		},
		EstimateID: "EST001",
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Create", mock.Anything).Return(nil)
	mockDB.On("Commit").Return(gorm.ErrInvalidTransaction)
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/material-rows", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateMaterialRow(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "DATABASE_ERROR", resp.Code)
	assert.Equal(t, "Failed to commit transaction", resp.Error)

	mockDB.AssertExpectations(t)
}

// ----------------- GetMaterialRow -----------------
func TestGetMaterialRow_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	materialRows := []models.MaterialRow{
		{
			ID: 1,
			Details: models.MaterialRowDetails{
				Description: "Test Material 1",
				Material:    "Wood",
				Category:    "Building Materials",
				SubCategory: "Lumber",
				Quantity:    10,
				UnitPrice:   100,
				Total:       1000,
			},
			EstimateID:         "EST001",
			ResidentialEntryID: 1,
		},
		{
			ID: 2,
			Details: models.MaterialRowDetails{
				Description: "Test Material 2",
				Material:    "Steel",
				Category:    "Building Materials",
				SubCategory: "Metal",
				Quantity:    5,
				UnitPrice:   200,
				Total:       1000,
			},
			EstimateID:         "EST001",
			ResidentialEntryID: 2,
		},
	}

	residentialEntry := models.ResidentialEntry{
		ID: 1,
		// Add other ResidentialEntry fields as needed
	}

	mockDB.On("Find", mock.AnythingOfType("*[]models.MaterialRow"), "estimate_id = ?", "EST001").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*[]models.MaterialRow)
		*dest = materialRows
	}).Return(nil)

	// Mock Preload chain for residential entry lookup
	mockDB.On("Preload", "ManagerResidentialEntry").Return(mockDB).Times(2)
	mockDB.On("Preload", "ManagerResidentialType").Return(mockDB).Times(2)
	mockDB.On("Where", "id = ?", mock.Anything).Return(mockDB).Times(2)
	mockDB.On("Find", mock.AnythingOfType("*models.ResidentialEntry")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.ResidentialEntry)
		*dest = residentialEntry
	}).Return(nil).Times(2)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/material-rows/EST001", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "EST001"}}
	c.Set("db", mockDB)

	handler.GetMaterialRow(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string][]models.MaterialRow
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, 2, len(resp["data"]))
	assert.Equal(t, "Test Material 1", resp["data"][0].Details.Description)
	assert.Equal(t, uint(10), resp["data"][0].Details.Quantity)

	mockDB.AssertExpectations(t)
}

func TestGetMaterialRow_MissingID(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/material-rows/", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: ""}}
	c.Set("db", mockDB)

	handler.GetMaterialRow(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
	assert.Equal(t, "Missing ID", resp.Error)
}

func TestGetMaterialRow_NotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("Find", mock.AnythingOfType("*[]models.MaterialRow"), "estimate_id = ?", "EST999").Return(gorm.ErrRecordNotFound)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/material-rows/EST999", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "EST999"}}
	c.Set("db", mockDB)

	handler.GetMaterialRow(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "NOT_FOUND", resp.Code)

	mockDB.AssertExpectations(t)
}

func TestGetMaterialRow_DatabaseError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("Find", mock.AnythingOfType("*[]models.MaterialRow"), "estimate_id = ?", "EST001").Return(gorm.ErrInvalidDB)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/material-rows/EST001", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "EST001"}}
	c.Set("db", mockDB)

	handler.GetMaterialRow(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "DATABASE_ERROR", resp.Code)

	mockDB.AssertExpectations(t)
}

func TestGetMaterialRow_ResidentialEntryNotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	materialRows := []models.MaterialRow{
		{
			ID: 1,
			Details: models.MaterialRowDetails{
				Description: "Test Material",
				Material:    "Wood",
			},
			EstimateID:         "EST001",
			ResidentialEntryID: 999,
		},
	}

	mockDB.On("Find", mock.AnythingOfType("*[]models.MaterialRow"), "estimate_id = ?", "EST001").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*[]models.MaterialRow)
		*dest = materialRows
	}).Return(nil)

	// Mock Preload chain for residential entry lookup that returns not found
	mockDB.On("Preload", "ManagerResidentialEntry").Return(mockDB)
	mockDB.On("Preload", "ManagerResidentialType").Return(mockDB)
	mockDB.On("Where", "id = ?", uint(999)).Return(mockDB)
	mockDB.On("Find", mock.AnythingOfType("*models.ResidentialEntry")).Return(gorm.ErrRecordNotFound)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/material-rows/EST001", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "EST001"}}
	c.Set("db", mockDB)

	handler.GetMaterialRow(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "NOT_FOUND", resp.Code)
	assert.Contains(t, resp.Error, "Residential entry not found")

	mockDB.AssertExpectations(t)
}

func TestGetMaterialRow_ResidentialEntryDatabaseError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	materialRows := []models.MaterialRow{
		{
			ID: 1,
			Details: models.MaterialRowDetails{
				Description: "Test Material",
			},
			EstimateID:         "EST001",
			ResidentialEntryID: 1,
		},
	}

	mockDB.On("Find", mock.AnythingOfType("*[]models.MaterialRow"), "estimate_id = ?", "EST001").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*[]models.MaterialRow)
		*dest = materialRows
	}).Return(nil)

	// Mock Preload chain for residential entry lookup that returns database error
	mockDB.On("Preload", "ManagerResidentialEntry").Return(mockDB)
	mockDB.On("Preload", "ManagerResidentialType").Return(mockDB)
	mockDB.On("Where", "id = ?", uint(1)).Return(mockDB)
	mockDB.On("Find", mock.AnythingOfType("*models.ResidentialEntry")).Return(gorm.ErrInvalidDB)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/material-rows/EST001", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "EST001"}}
	c.Set("db", mockDB)

	handler.GetMaterialRow(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "DATABASE_ERROR", resp.Code)

	mockDB.AssertExpectations(t)
}

func TestGetMaterialRow_EmptyList(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	materialRows := []models.MaterialRow{}

	mockDB.On("Find", mock.AnythingOfType("*[]models.MaterialRow"), "estimate_id = ?", "EST001").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*[]models.MaterialRow)
		*dest = materialRows
	}).Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/material-rows/EST001", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "EST001"}}
	c.Set("db", mockDB)

	handler.GetMaterialRow(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string][]models.MaterialRow
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, 0, len(resp["data"]))

	mockDB.AssertExpectations(t)
}

// ----------------- UpdateMaterialRow -----------------
func TestUpdateMaterialRow_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	existing := models.MaterialRow{
		ID: 1,
		Details: models.MaterialRowDetails{
			Description: "Old Material",
			Material:    "Wood",
			Quantity:    10,
			UnitPrice:   100,
			Total:       1000,
		},
		EstimateID: "EST001",
	}

	updated := models.MaterialRow{
		Details: models.MaterialRowDetails{
			Description: "Updated Material",
			Material:    "Steel",
			Quantity:    20,
			UnitPrice:   150,
			Total:       3000,
		},
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "1").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.MaterialRow)
		*dest = existing
	}).Return(nil)
	mockDB.On("Save", mock.Anything).Return(nil)
	mockDB.On("Commit").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(updated)
	req, _ := http.NewRequest("PUT", "/material-rows/1", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.UpdateMaterialRow(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "MaterialRow updated successfully", resp["message"])
	assert.NotNil(t, resp["data"])

	mockDB.AssertExpectations(t)
}

func TestUpdateMaterialRow_NotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	input := models.MaterialRow{
		Details: models.MaterialRowDetails{
			Description: "Test Material",
		},
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "999").Return(gorm.ErrRecordNotFound)
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("PUT", "/material-rows/999", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "999"}}
	c.Set("db", mockDB)

	handler.UpdateMaterialRow(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "NOT_FOUND", resp.Code)

	mockDB.AssertExpectations(t)
}

func TestUpdateMaterialRow_InvalidInput(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("PUT", "/material-rows/1", bytes.NewBuffer([]byte(`invalid json`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.UpdateMaterialRow(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
}

func TestUpdateMaterialRow_MissingID(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	input := models.MaterialRow{
		Details: models.MaterialRowDetails{
			Description: "Test",
		},
	}
	data, _ := json.Marshal(input)

	req, _ := http.NewRequest("PUT", "/material-rows/", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: ""}}
	c.Set("db", mockDB)

	handler.UpdateMaterialRow(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
}

func TestUpdateMaterialRow_DatabaseError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	input := models.MaterialRow{
		Details: models.MaterialRowDetails{
			Description: "Test",
		},
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "1").Return(gorm.ErrInvalidDB)
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("PUT", "/material-rows/1", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.UpdateMaterialRow(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "DATABASE_ERROR", resp.Code)

	mockDB.AssertExpectations(t)
}

func TestUpdateMaterialRow_SaveError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	existing := models.MaterialRow{
		ID: 1,
		Details: models.MaterialRowDetails{
			Description: "Test",
		},
	}
	updated := models.MaterialRow{
		Details: models.MaterialRowDetails{
			Description: "Updated",
		},
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "1").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.MaterialRow)
		*dest = existing
	}).Return(nil)
	mockDB.On("Save", mock.Anything).Return(gorm.ErrInvalidDB)
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(updated)
	req, _ := http.NewRequest("PUT", "/material-rows/1", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.UpdateMaterialRow(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "DATABASE_ERROR", resp.Code)

	mockDB.AssertExpectations(t)
}

func TestUpdateMaterialRow_CommitError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	existing := models.MaterialRow{
		ID: 1,
		Details: models.MaterialRowDetails{
			Description: "Test",
		},
	}
	updated := models.MaterialRow{
		Details: models.MaterialRowDetails{
			Description: "Updated",
		},
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "1").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.MaterialRow)
		*dest = existing
	}).Return(nil)
	mockDB.On("Save", mock.Anything).Return(nil)
	mockDB.On("Commit").Return(gorm.ErrInvalidTransaction)
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(updated)
	req, _ := http.NewRequest("PUT", "/material-rows/1", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.UpdateMaterialRow(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "DATABASE_ERROR", resp.Code)

	mockDB.AssertExpectations(t)
}

// ----------------- DeleteMaterialRow -----------------
func TestDeleteMaterialRow_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	existing := models.MaterialRow{
		ID: 1,
		Details: models.MaterialRowDetails{
			Description: "Test Material",
			Material:    "Wood",
		},
		EstimateID: "EST001",
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "1").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.MaterialRow)
		*dest = existing
	}).Return(nil)
	mockDB.On("Delete", mock.Anything).Return(nil)
	mockDB.On("Commit").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("DELETE", "/material-rows/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.DeleteMaterialRow(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Residential entry deleted successfully", resp["message"])

	mockDB.AssertExpectations(t)
}

func TestDeleteMaterialRow_NotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "999").Return(gorm.ErrRecordNotFound)
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("DELETE", "/material-rows/999", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "999"}}
	c.Set("db", mockDB)

	handler.DeleteMaterialRow(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "NOT_FOUND", resp.Code)

	mockDB.AssertExpectations(t)
}

func TestDeleteMaterialRow_MissingID(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("DELETE", "/material-rows/", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: ""}}
	c.Set("db", mockDB)

	handler.DeleteMaterialRow(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
}

func TestDeleteMaterialRow_DatabaseError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "1").Return(gorm.ErrInvalidDB)
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("DELETE", "/material-rows/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.DeleteMaterialRow(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "DATABASE_ERROR", resp.Code)

	mockDB.AssertExpectations(t)
}
