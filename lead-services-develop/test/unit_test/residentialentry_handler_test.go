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

// ----------------- CreateResidentialEntry -----------------
func TestCreateResidentialEntry_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mre := models.ManagerResidentialEntry{ID: 1}
	mrt := models.ManagerResidentialType{ID: 1}
	leadID := "1"
	estimateID := "1"

	input := models.ResidentialEntry{
		ManagerResidentialEntryID: 1,
		ManagerResidentialTypeID:  1,
		LeadID:                    &leadID,
		EstimateID:                &estimateID,
		Details:                   "Test details",
	}

	// Mock ManagerResidentialEntry lookup
	mockDB.On("First", mock.AnythingOfType("*models.ManagerResidentialEntry"), uint(1)).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.ManagerResidentialEntry)
		*dest = mre
	}).Return(nil)

	// Mock ManagerResidentialType lookup
	mockDB.On("First", mock.AnythingOfType("*models.ManagerResidentialType"), uint(1)).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.ManagerResidentialType)
		*dest = mrt
	}).Return(nil)

	// Mock Lead lookup
	mockDB.On("First", mock.AnythingOfType("*models.Lead"), &leadID).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Lead)
		*dest = models.Lead{ID: "1"}
	}).Return(nil)

	// Mock Estimate lookup
	mockDB.On("First", mock.AnythingOfType("*models.Estimate"), &estimateID).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Estimate)
		*dest = models.Estimate{ID: "1"}
	}).Return(nil)

	// Mock Create
	mockDB.On("Create", mock.Anything).Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/residential-entries", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateResidentialEntry(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Residential entry created successfully", resp["message"])

	mockDB.AssertExpectations(t)
}

func TestCreateResidentialEntry_InvalidInput(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("POST", "/residential-entries", bytes.NewBuffer([]byte(`invalid json`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateResidentialEntry(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
}

func TestCreateResidentialEntry_MissingRequiredFields(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	input := models.ResidentialEntry{
		ManagerResidentialEntryID: 1,
		// Missing ManagerResidentialTypeID
	}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/residential-entries", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateResidentialEntry(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
	assert.Equal(t, "Missing required fields", resp.Error)
}

func TestCreateResidentialEntry_InvalidManagerResidentialEntryID(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	input := models.ResidentialEntry{
		ManagerResidentialEntryID: 999,
		ManagerResidentialTypeID:  1,
	}

	mockDB.On("First", mock.AnythingOfType("*models.ManagerResidentialEntry"), uint(999)).Return(gorm.ErrRecordNotFound)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/residential-entries", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateResidentialEntry(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
	assert.Equal(t, "Invalid ManagerResidentialEntryID", resp.Error)

	mockDB.AssertExpectations(t)
}

func TestCreateResidentialEntry_InvalidManagerResidentialTypeID(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mre := models.ManagerResidentialEntry{ID: 1}
	input := models.ResidentialEntry{
		ManagerResidentialEntryID: 1,
		ManagerResidentialTypeID:  999,
	}

	mockDB.On("First", mock.AnythingOfType("*models.ManagerResidentialEntry"), uint(1)).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.ManagerResidentialEntry)
		*dest = mre
	}).Return(nil)

	mockDB.On("First", mock.AnythingOfType("*models.ManagerResidentialType"), uint(999)).Return(gorm.ErrRecordNotFound)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/residential-entries", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateResidentialEntry(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
	assert.Equal(t, "Invalid ManagerResidentialTypeID", resp.Error)

	mockDB.AssertExpectations(t)
}

func TestCreateResidentialEntry_InvalidLeadID(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mre := models.ManagerResidentialEntry{ID: 1}
	mrt := models.ManagerResidentialType{ID: 1}
	leadID := "999"

	input := models.ResidentialEntry{
		ManagerResidentialEntryID: 1,
		ManagerResidentialTypeID:  1,
		LeadID:                    &leadID,
	}

	mockDB.On("First", mock.AnythingOfType("*models.ManagerResidentialEntry"), uint(1)).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.ManagerResidentialEntry)
		*dest = mre
	}).Return(nil)

	mockDB.On("First", mock.AnythingOfType("*models.ManagerResidentialType"), uint(1)).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.ManagerResidentialType)
		*dest = mrt
	}).Return(nil)

	mockDB.On("First", mock.AnythingOfType("*models.Lead"), &leadID).Return(gorm.ErrRecordNotFound)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/residential-entries", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateResidentialEntry(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
	assert.Equal(t, "Invalid LeadID", resp.Error)

	mockDB.AssertExpectations(t)
}

func TestCreateResidentialEntry_InvalidEstimateID(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mre := models.ManagerResidentialEntry{ID: 1}
	mrt := models.ManagerResidentialType{ID: 1}
	estimateID := "999"

	input := models.ResidentialEntry{
		ManagerResidentialEntryID: 1,
		ManagerResidentialTypeID:  1,
		EstimateID:                &estimateID,
	}

	mockDB.On("First", mock.AnythingOfType("*models.ManagerResidentialEntry"), uint(1)).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.ManagerResidentialEntry)
		*dest = mre
	}).Return(nil)

	mockDB.On("First", mock.AnythingOfType("*models.ManagerResidentialType"), uint(1)).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.ManagerResidentialType)
		*dest = mrt
	}).Return(nil)

	mockDB.On("First", mock.AnythingOfType("*models.Estimate"), &estimateID).Return(gorm.ErrRecordNotFound)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/residential-entries", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateResidentialEntry(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
	assert.Equal(t, "Invalid EstimateID", resp.Error)

	mockDB.AssertExpectations(t)
}

// ----------------- GetResidentialEntry -----------------
func TestGetResidentialEntry_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	expected := models.ResidentialEntry{
		ID:                        1,
		ManagerResidentialEntryID: 1,
		ManagerResidentialTypeID:  1,
		Details:                   "Test details",
	}

	mockDB.On("Preload", "ManagerResidentialEntry").Return(mockDB)
	mockDB.On("Preload", "ManagerResidentialType").Return(mockDB)
	mockDB.On("Preload", "Lead").Return(mockDB)
	mockDB.On("Preload", "Estimate").Return(mockDB)
	mockDB.On("First", mock.Anything, uint64(1)).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.ResidentialEntry)
		*dest = expected
	}).Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/residential-entries/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.GetResidentialEntry(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]models.ResidentialEntry
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, expected.Details, resp["data"].Details)

	mockDB.AssertExpectations(t)
}

func TestGetResidentialEntry_InvalidID(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/residential-entries/invalid", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "invalid"}}
	c.Set("db", mockDB)

	handler.GetResidentialEntry(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
	assert.Equal(t, "Invalid ID", resp.Error)
}

func TestGetResidentialEntry_NotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("Preload", "ManagerResidentialEntry").Return(mockDB)
	mockDB.On("Preload", "ManagerResidentialType").Return(mockDB)
	mockDB.On("Preload", "Lead").Return(mockDB)
	mockDB.On("Preload", "Estimate").Return(mockDB)
	mockDB.On("First", mock.Anything, uint64(999)).Return(gorm.ErrRecordNotFound)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/residential-entries/999", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "999"}}
	c.Set("db", mockDB)

	handler.GetResidentialEntry(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "NOT_FOUND", resp.Code)

	mockDB.AssertExpectations(t)
}

// ----------------- UpdateResidentialEntry -----------------
func TestUpdateResidentialEntry_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	existing := models.ResidentialEntry{
		ID:                        1,
		ManagerResidentialEntryID: 1,
		ManagerResidentialTypeID:  1,
		Details:                   "Old details",
	}

	updated := models.ResidentialEntry{
		ManagerResidentialEntryID: 2,
		ManagerResidentialTypeID:  2,
		Details:                   "New details",
	}

	mre := models.ManagerResidentialEntry{ID: 2}
	mrt := models.ManagerResidentialType{ID: 2}

	mockDB.On("First", mock.AnythingOfType("*models.ResidentialEntry"), uint64(1)).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.ResidentialEntry)
		*dest = existing
	}).Return(nil)

	mockDB.On("First", mock.AnythingOfType("*models.ManagerResidentialEntry"), uint(2)).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.ManagerResidentialEntry)
		*dest = mre
	}).Return(nil)

	mockDB.On("First", mock.AnythingOfType("*models.ManagerResidentialType"), uint(2)).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.ManagerResidentialType)
		*dest = mrt
	}).Return(nil)

	mockDB.On("Save", mock.Anything).Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(updated)
	req, _ := http.NewRequest("PUT", "/residential-entries/1", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.UpdateResidentialEntry(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Residential entry updated successfully", resp["message"])

	mockDB.AssertExpectations(t)
}

func TestUpdateResidentialEntry_NotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("First", mock.AnythingOfType("*models.ResidentialEntry"), uint64(999)).Return(gorm.ErrRecordNotFound)

	handler := &handlers.Handler{DB: mockDB}

	updated := models.ResidentialEntry{
		ManagerResidentialEntryID: 1,
		ManagerResidentialTypeID:  1,
	}

	data, _ := json.Marshal(updated)
	req, _ := http.NewRequest("PUT", "/residential-entries/999", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "999"}}
	c.Set("db", mockDB)

	handler.UpdateResidentialEntry(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "NOT_FOUND", resp.Code)

	mockDB.AssertExpectations(t)
}

func TestUpdateResidentialEntry_InvalidID(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("PUT", "/residential-entries/invalid", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "invalid"}}
	c.Set("db", mockDB)

	handler.UpdateResidentialEntry(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
}

// ----------------- DeleteResidentialEntry -----------------
func TestDeleteResidentialEntry_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	// Change this line - expect string "1" not uint64(1)
	mockDB.On("Delete", mock.AnythingOfType("*models.ResidentialEntry"), "1").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("DELETE", "/residential-entries/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.DeleteResidentialEntry(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Residential entry deleted successfully", resp["message"])

	mockDB.AssertExpectations(t)
}

// ----------------- GetAllResidentialEntry -----------------
func TestGetAllResidentialEntry_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	entryList := []models.ResidentialEntry{
		{ID: 1, ManagerResidentialEntryID: 1, ManagerResidentialTypeID: 1},
		{ID: 2, ManagerResidentialEntryID: 2, ManagerResidentialTypeID: 2},
	}

	mockDB.On("Find", mock.AnythingOfType("*[]models.ResidentialEntry")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*[]models.ResidentialEntry)
		*dest = entryList
	}).Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/residential-entries", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.GetAllResidentialEntry(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	data, ok := resp["data"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(data))

	mockDB.AssertExpectations(t)
}

// ----------------- GetAllByLead -----------------
func TestGetAllByLead_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	estimateID := "1"
	entryList := []models.ResidentialEntry{
		{ID: 1, EstimateID: &estimateID},
		{ID: 2, EstimateID: &estimateID},
	}

	mockDB.On("Preload", "ManagerResidentialEntry").Return(mockDB)
	mockDB.On("Preload", "ManagerResidentialType").Return(mockDB)
	mockDB.On("Preload", "Lead").Return(mockDB)
	mockDB.On("Preload", "Estimate").Return(mockDB)
	mockDB.On("Where", "estimate_id = ?", "1").Return(mockDB)
	mockDB.On("Find", mock.AnythingOfType("*[]models.ResidentialEntry")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*[]models.ResidentialEntry)
		*dest = entryList
	}).Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/residential-entries/estimate/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.GetAllByLead(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	data, ok := resp["data"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(data))

	mockDB.AssertExpectations(t)
}

// ----------------- GetOnlyByLead -----------------
func TestGetOnlyByLead_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	leadID := "1"
	entryList := []models.ResidentialEntry{
		{ID: 1, LeadID: &leadID},
		{ID: 2, LeadID: &leadID},
	}

	mockDB.On("Preload", "ManagerResidentialEntry").Return(mockDB)
	mockDB.On("Preload", "ManagerResidentialType").Return(mockDB)
	mockDB.On("Preload", "Lead").Return(mockDB)
	mockDB.On("Preload", "Estimate").Return(mockDB)
	mockDB.On("Where", "lead_id = ?", "1").Return(mockDB)
	mockDB.On("Find", mock.AnythingOfType("*[]models.ResidentialEntry")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*[]models.ResidentialEntry)
		*dest = entryList
	}).Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/residential-entries/lead/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.GetOnlyByLead(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	data, ok := resp["data"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(data))

	mockDB.AssertExpectations(t)
}
