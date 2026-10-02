package handlers

import (
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

// ----------------- GetEstimate -----------------
func TestGetEstimate_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	leadID := "lead-123"
	expected := models.Estimate{
		ID:     "EST-001",
		Name:   "Test Estimate",
		Status: "draft",
		LeadID: &leadID,
	}

	// Mock Preload chain
	mockDB.On("Preload", "Lead").Return(mockDB).Once()
	mockDB.On("Preload", "MaterialRows").Return(mockDB).Once()
	mockDB.On("Preload", "ServiceRows").Return(mockDB).Once()
	mockDB.On("Preload", "ResidentialEntries").Return(mockDB).Once()
	mockDB.On("Preload", "ResidentialEntries.ManagerResidentialEntry").Return(mockDB).Once()
	mockDB.On("Preload", "ResidentialEntries.ManagerResidentialType").Return(mockDB).Once()

	mockDB.On("First", mock.AnythingOfType("*models.Estimate"), "id = ?", "EST-001").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Estimate)
		*dest = expected
	}).Return(nil).Once()

	// Mock Lead fetch
	mockDB.On("Model", mock.AnythingOfType("*models.Lead")).Return(mockDB).Once()
	mockDB.On("Where", "id = ?", &leadID).Return(mockDB).Once()
	mockDB.On("Preload", "Files").Return(mockDB).Once()
	mockDB.On("First", mock.AnythingOfType("*models.Lead")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Lead)
		dest.ID = leadID
	}).Return(nil).Once()

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/estimates/EST-001", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "EST-001"}}
	c.Set("db", mockDB)

	handler.GetEstimate(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "estimate retrieved successfully", resp["message"])

	mockDB.AssertExpectations(t)
}

func TestGetEstimate_NotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("Preload", "Lead").Return(mockDB).Once()
	mockDB.On("Preload", "MaterialRows").Return(mockDB).Once()
	mockDB.On("Preload", "ServiceRows").Return(mockDB).Once()
	mockDB.On("Preload", "ResidentialEntries").Return(mockDB).Once()
	mockDB.On("Preload", "ResidentialEntries.ManagerResidentialEntry").Return(mockDB).Once()
	mockDB.On("Preload", "ResidentialEntries.ManagerResidentialType").Return(mockDB).Once()

	mockDB.On("First", mock.AnythingOfType("*models.Estimate"), "id = ?", "INVALID").Return(gorm.ErrRecordNotFound).Once()

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/estimates/INVALID", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "INVALID"}}
	c.Set("db", mockDB)

	handler.GetEstimate(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "NOT_FOUND", resp.Code)

	mockDB.AssertExpectations(t)
}

// ----------------- GetEstimates -----------------
func TestGetEstimates_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	estimates := []models.Estimate{
		{ID: "EST-001", Name: "Estimate 1"},
		{ID: "EST-002", Name: "Estimate 2"},
	}

	var total int64 = 2

	// Mock Preload chain for Find
	mockDB.On("Preload", "Lead").Return(mockDB).Once()
	mockDB.On("Preload", "MaterialRows").Return(mockDB).Once()
	mockDB.On("Preload", "ServiceRows").Return(mockDB).Once()
	mockDB.On("Preload", "ResidentialEntries").Return(mockDB).Once()
	mockDB.On("Offset", 0).Return(mockDB).Once()
	mockDB.On("Limit", 10).Return(mockDB).Once()
	mockDB.On("Find", mock.AnythingOfType("*[]models.Estimate")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*[]models.Estimate)
		*dest = estimates
	}).Return(nil).Once()

	// Mock Count
	mockDB.On("Model", mock.AnythingOfType("*models.Estimate")).Return(mockDB).Once()
	mockDB.On("Count", mock.AnythingOfType("*int64")).Run(func(args mock.Arguments) {
		countPtr := args.Get(0).(*int64)
		*countPtr = total
	}).Return(mockDB).Once()
	mockDB.On("Error").Return(nil).Once()

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/estimates?page=1&limit=10", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.GetEstimates(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "estimates retrieved successfully", resp["message"])

	mockDB.AssertExpectations(t)
}

// ----------------- DeleteEstimate -----------------
func TestDeleteEstimate_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	existing := models.Estimate{
		ID:   "EST-001",
		Name: "Test Estimate",
	}

	mockDB.On("First", mock.AnythingOfType("*models.Estimate"), "id = ?", "EST-001").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Estimate)
		*dest = existing
	}).Return(nil).Once()

	mockDB.On("Begin").Return(mockDB).Once()
	mockDB.On("Delete", mock.AnythingOfType("*models.Estimate")).Return(nil).Once()
	mockDB.On("Commit").Return(nil).Once()

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("DELETE", "/estimates/EST-001", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "EST-001"}}
	c.Set("db", mockDB)

	handler.DeleteEstimate(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "estimate deleted successfully", resp["message"])

	mockDB.AssertExpectations(t)
}

func TestDeleteEstimate_NotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("First", mock.AnythingOfType("*models.Estimate"), "id = ?", "INVALID").Return(gorm.ErrRecordNotFound).Once()

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("DELETE", "/estimates/INVALID", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "INVALID"}}
	c.Set("db", mockDB)

	handler.DeleteEstimate(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "NOT_FOUND", resp.Code)

	mockDB.AssertExpectations(t)
}

// ----------------- GetEstimatesByLead -----------------
func TestGetEstimatesByLead_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	estimates := []models.Estimate{
		{ID: "EST-001", Name: "Estimate 1"},
		{ID: "EST-002", Name: "Estimate 2"},
	}

	mockDB.On("Preload", "Lead").Return(mockDB).Once()
	mockDB.On("Preload", "MaterialRows").Return(mockDB).Once()
	mockDB.On("Preload", "ServiceRows").Return(mockDB).Once()
	mockDB.On("Preload", "ResidentialEntries").Return(mockDB).Once()
	mockDB.On("Where", "lead_id = ?", "lead-123").Return(mockDB).Once()
	mockDB.On("Find", mock.AnythingOfType("*[]models.Estimate")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*[]models.Estimate)
		*dest = estimates
	}).Return(nil).Once()

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/leads/lead-123/estimates", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "leadId", Value: "lead-123"}}
	c.Set("db", mockDB)

	handler.GetEstimatesByLead(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "estimates retrieved successfully", resp["message"])

	mockDB.AssertExpectations(t)
}

// ----------------- GetEstimatesByStatus -----------------
func TestGetEstimatesByStatus_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	estimates := []models.Estimate{
		{ID: "EST-001", Name: "Estimate 1", Status: "approved"},
		{ID: "EST-002", Name: "Estimate 2", Status: "approved"},
	}

	mockDB.On("Preload", "Lead").Return(mockDB).Once()
	mockDB.On("Preload", "MaterialRows").Return(mockDB).Once()
	mockDB.On("Preload", "ServiceRows").Return(mockDB).Once()
	mockDB.On("Preload", "ResidentialEntries").Return(mockDB).Once()
	mockDB.On("Where", "status = ?", "approved").Return(mockDB).Once()
	mockDB.On("Find", mock.AnythingOfType("*[]models.Estimate")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*[]models.Estimate)
		*dest = estimates
	}).Return(nil).Once()

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/estimates/status/approved", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "status", Value: "approved"}}
	c.Set("db", mockDB)

	handler.GetEstimatesByStatus(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "estimates retrieved successfully", resp["message"])

	mockDB.AssertExpectations(t)
}

// ----------------- GetAllEstimates -----------------
func TestGetAllEstimates_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	estimates := []models.Estimate{
		{ID: "EST-001", Name: "Estimate 1"},
		{ID: "EST-002", Name: "Estimate 2"},
	}

	var total int64 = 2

	// Mock Count
	mockDB.On("Model", mock.AnythingOfType("*models.Estimate")).Return(mockDB).Once()
	mockDB.On("Count", mock.AnythingOfType("*int64")).Run(func(args mock.Arguments) {
		countPtr := args.Get(0).(*int64)
		*countPtr = total
	}).Return(mockDB).Once()
	mockDB.On("Error").Return(nil).Once()

	// Mock query chain
	mockDB.On("Model", mock.AnythingOfType("*models.Estimate")).Return(mockDB).Once()
	mockDB.On("Preload", "ResidentialEntries").Return(mockDB).Once()
	mockDB.On("Preload", "MaterialRows").Return(mockDB).Once()
	mockDB.On("Preload", "ServiceRows").Return(mockDB).Once()
	mockDB.On("Preload", "Lead").Return(mockDB).Once()
	mockDB.On("Order", "created_at desc").Return(mockDB).Once()
	mockDB.On("Find", mock.AnythingOfType("*[]models.Estimate")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*[]models.Estimate)
		*dest = estimates
	}).Return(nil).Once()

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/estimates/all", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.GetAllEstimates(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "estimates retrieved successfully", resp["message"])

	mockDB.AssertExpectations(t)
}

// ----------------- GetEstimateVersions -----------------
func TestGetEstimateVersions_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	versions := []models.Estimate{
		{ID: "EST-001-v1"},
		{ID: "EST-001-v2"},
		{ID: "EST-001-v3"},
	}

	mockDB.On("Where", "id LIKE ?", "EST-001-v%").Return(mockDB).Once()
	mockDB.On("Select", "id").Return(mockDB).Once()
	mockDB.On("Find", mock.AnythingOfType("*[]models.Estimate")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*[]models.Estimate)
		*dest = versions
	}).Return(nil).Once()

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/estimates/EST-001/versions", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "EST-001"}}
	c.Set("db", mockDB)

	handler.GetEstimateVersions(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "EST-001", resp["original_id"])
	assert.Equal(t, float64(3), resp["count"])

	mockDB.AssertExpectations(t)
}

func TestGetEstimateVersions_InvalidID(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/estimates//versions", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: ""}}
	c.Set("db", mockDB)

	handler.GetEstimateVersions(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
}

// ----------------- DuplicateEstimate -----------------
func TestDuplicateEstimate_InvalidID(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("POST", "/estimates//duplicate", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: ""}}
	c.Set("db", mockDB)

	handler.DuplicateEstimate(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
}

func TestDuplicateEstimate_NotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("Preload", "Lead").Return(mockDB).Once()
	mockDB.On("Preload", "MaterialRows").Return(mockDB).Once()
	mockDB.On("Preload", "ServiceRows").Return(mockDB).Once()
	mockDB.On("Preload", "ResidentialEntries").Return(mockDB).Once()
	mockDB.On("Preload", "ResidentialEntries.ManagerResidentialEntry").Return(mockDB).Once()
	mockDB.On("Preload", "ResidentialEntries.ManagerResidentialType").Return(mockDB).Once()
	mockDB.On("Where", "id = ?", "INVALID").Return(mockDB).Once()
	mockDB.On("First", mock.AnythingOfType("*models.Estimate")).Return(gorm.ErrRecordNotFound).Once()

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("POST", "/estimates/INVALID/duplicate", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "INVALID"}}
	c.Set("db", mockDB)

	handler.DuplicateEstimate(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "NOT_FOUND", resp.Code)

	mockDB.AssertExpectations(t)
}

// ----------------- Error Cases -----------------
func TestGetEstimates_CountError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("Preload", "Lead").Return(mockDB).Once()
	mockDB.On("Preload", "MaterialRows").Return(mockDB).Once()
	mockDB.On("Preload", "ServiceRows").Return(mockDB).Once()
	mockDB.On("Preload", "ResidentialEntries").Return(mockDB).Once()
	mockDB.On("Offset", 0).Return(mockDB).Once()
	mockDB.On("Limit", 10).Return(mockDB).Once()
	mockDB.On("Find", mock.AnythingOfType("*[]models.Estimate")).Return(nil).Once()

	mockDB.On("Model", mock.AnythingOfType("*models.Estimate")).Return(mockDB).Once()
	mockDB.On("Count", mock.AnythingOfType("*int64")).Return(mockDB).Once()
	mockDB.On("Error").Return(errors.New("database error")).Once()

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/estimates?page=1&limit=10", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.GetEstimates(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "DATABASE_ERROR", resp.Code)

	mockDB.AssertExpectations(t)
}
