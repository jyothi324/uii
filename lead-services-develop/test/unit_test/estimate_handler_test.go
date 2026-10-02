package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"bytes"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/tuliporg/agent-management-api/handlers"
	"github.com/tuliporg/agent-management-api/mocks"
	"github.com/tuliporg/agent-management-api/models"
	"gorm.io/gorm"
	// "fmt"
)

func TestDuplicateEstimateJob_Success_Minimal(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	estimate := models.Estimate{
		ID:   "EST-1",
		Lead: models.Lead{ID: "LEAD-1"},
	}

	// Preload chain
	mockDB.On("Preload", mock.Anything).Return(mockDB)

	// Original fetch
	mockDB.On("Where", "id = ?", "EST-1").Return(mockDB)
	mockDB.On("First", mock.Anything).
		Run(func(args mock.Arguments) {
			*(args.Get(0).(*models.Estimate)) = estimate
		}).
		Return(nil).
		Once()

	mockDB.On("Model", mock.Anything).Return(mockDB)
	mockDB.On("Where", "lead_id = ?", "LEAD-1").Return(mockDB)
	mockDB.On("Update", "lead_id", nil).Return(nil)

	mockDB.On("Where", "id LIKE ?", "EST-1-v%").Return(mockDB)
	mockDB.On("Find", mock.Anything).Return(nil)

	mockDB.On("Create", mock.Anything).Return(nil)

	mockDB.On("Where", "id = ?", mock.Anything).Return(mockDB)
	mockDB.On("First", mock.Anything).Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("POST", "/estimates/EST-1/duplicate-job", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "EST-1"}}
	c.Set("db", mockDB)

	handler.DuplicateEstimateJob(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockDB.AssertExpectations(t)
}

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

func TestDuplicateEstimate_FindEstimateDBError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("Preload", mock.Anything).Return(mockDB)
	mockDB.On("Where", "id = ?", "123").Return(mockDB)
	mockDB.On("First", mock.AnythingOfType("*models.Estimate")).
		Return(errors.New("db failure")).Once()

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("POST", "/estimates/123/duplicate", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "123"}}
	c.Set("db", mockDB)

	handler.DuplicateEstimate(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp ErrorResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "DATABASE_ERROR", resp.Code)

	mockDB.AssertExpectations(t)
}

func TestDuplicateEstimate_FindVersionsError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	leadID := "10"
	estimate := &models.Estimate{
		ID:     "1",
		LeadID: &leadID,
	}

	mockDB.On("Preload", mock.Anything).Return(mockDB)
	mockDB.On("Where", "id = ?", "123").Return(mockDB)

	mockDB.On("First", mock.AnythingOfType("*models.Estimate")).
		Run(func(args mock.Arguments) {
			*args.Get(0).(*models.Estimate) = *estimate
		}).Return(nil)

	// ❗ Version fetch fails
	mockDB.On("Where", "id LIKE ?", "123-v%").Return(mockDB)
	mockDB.On("Find", mock.AnythingOfType("*[]models.Estimate")).
		Return(errors.New("query failed")).Once()

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("POST", "/estimates/123/duplicate", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "123"}}
	c.Set("db", mockDB)

	handler.DuplicateEstimate(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp ErrorResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "DATABASE_ERROR", resp.Code)

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

func TestCreateStagesForEstimate_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	estimateID := "EST123"
	leadID := "LEAD123"

	estimate := models.Estimate{
		ID:     estimateID,
		LeadID: &leadID,
		ResidentialEntries: []models.ResidentialEntry{
			{ID: 1},
		},
	}

	// ---- Estimate fetch ----
	mockDB.On("Preload", "ResidentialEntries").Return(mockDB)
	mockDB.On("Preload", "Lead").Return(mockDB)

	// IMPORTANT: allow ANY id (GenerateEmployeeID also uses Where)
	mockDB.On("Where", "id = ?", mock.Anything).
		Return(mockDB)

	mockDB.On("First", mock.Anything).
		Run(func(args mock.Arguments) {
			arg := args.Get(0).(*models.Estimate)
			*arg = estimate
		}).
		Return(nil)

	// ---- Transaction ----
	mockDB.On("Begin").Return(mockDB)

	// residential_entry_id lookup
	mockDB.On("Where", "residential_entry_id = ?", uint(1)).
		Return(mockDB)

	mockDB.On("Find", mock.Anything).
		Run(func(args mock.Arguments) {
			arg := args.Get(0).(*[]models.Stage)
			*arg = []models.Stage{} // no existing stages
		}).
		Return(nil)

	mockDB.On("Model", mock.Anything).
		Return(mockDB)

	mockDB.On("Count", mock.Anything).
		Run(func(args mock.Arguments) {
			arg := args.Get(0).(*int64)
			*arg = (0)
		}).
		Return(mockDB)

	mockDB.On("Create", mock.Anything).
		Return(nil)

	mockDB.On("Commit").
		Return(nil)

	// CRITICAL: always mock Rollback for deferred recovery
	mockDB.On("Rollback").
		Return(nil).
		Maybe()

	// ---- HTTP ----
	req, _ := http.NewRequest("POST", "/estimates/EST123/stages", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: estimateID}}
	c.Set("db", mockDB)

	handler.CreateStagesForEstimate(c)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &response)
	fmt.Println("Respose JSON:", response)

	assert.Equal(t, "Stages created successfully", response["message"])
	// assert.Equal(t, float64(6), response["stages_created_count"])

	mockDB.AssertExpectations(t)
}

func TestCreateStagesForEstimate_NoEstimateID(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("POST", "/estimates//stages", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: ""}}
	c.Set("db", mockDB)

	handler.CreateStagesForEstimate(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "VALIDATION_ERROR", response.Code)
}

func TestCreateStagesForEstimate_EstimateNotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	estimateID := "NOTFOUND"

	mockDB.On("Preload", "ResidentialEntries").Return(mockDB)
	mockDB.On("Preload", "Lead").Return(mockDB)
	mockDB.On("Where", "id = ?", estimateID).Return(mockDB)
	mockDB.On("First", mock.Anything).Return(gorm.ErrRecordNotFound)

	req, _ := http.NewRequest("POST", "/estimates/NOTFOUND/stages", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: estimateID}}
	c.Set("db", mockDB)

	handler.CreateStagesForEstimate(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "NOT_FOUND", response.Code)
}

// func TestCreateStagesForEstimate_StagesAlreadyExist(t *testing.T) {
// 	mockDB := new(mocks.MockDatabase)
// 	handler := &handlers.Handler{DB: mockDB}

// 	estimateID := "EST123"
// 	leadID := "LEAD123"

// 	estimate := models.Estimate{
// 		ID:     estimateID,
// 		LeadID: &leadID,
// 		ResidentialEntries: []models.ResidentialEntry{
// 			{ID: 1},
// 		},
// 	}

// 	existingStages := []models.Stage{{ID: "stage1"}}

// 	mockDB.On("Preload", "ResidentialEntries").Return(mockDB)
// 	mockDB.On("Preload", "Lead").Return(mockDB)
// 	mockDB.On("Where", "id = ?", estimateID).Return(mockDB)
// 	mockDB.On("First", mock.Anything).Run(func(args mock.Arguments) {
// 		arg := args.Get(0).(*models.Estimate)
// 		*arg = estimate
// 	}).Return(nil)

// 	mockDB.On("Begin").Return(mockDB)
// 	mockDB.On("Where", "residential_entry_id = ?", uint(1)).Return(mockDB)
// 	mockDB.On("Find", mock.Anything).Run(func(args mock.Arguments) {
// 		arg := args.Get(0).(*[]models.Stage)
// 		*arg = existingStages
// 	}).Return(nil)
// 	mockDB.On("Commit").Return(nil)
// 	mockDB.On("Rollback").Return(nil).Maybe()

// 	req, _ := http.NewRequest("POST", "/estimates/EST123/stages", nil)
// 	w := httptest.NewRecorder()
// 	c, _ := gin.CreateTestContext(w)
// 	c.Request = req
// 	c.Params = gin.Params{{Key: "id", Value: estimateID}}
// 	c.Set("db", mockDB)

// 	handler.CreateStagesForEstimate(c)

// 	handler.CreateStagesForEstimate(c)

// 	assert.Equal(t, http.StatusInternalServerError, w.Code)
// 	var response ErrorResponse
// 	err := json.Unmarshal(w.Body.Bytes(), &response)
// 	assert.NoError(t, err)
// 	assert.Equal(t, "DATABASE_ERROR", response.Code)
// 	assert.Equal(t, "Failed to check existing stages", response.Error)

// }

func TestCreateStagesForEstimate_FetchEstimateError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	estimateID := "EST123"

	mockDB.On("Preload", "ResidentialEntries").Return(mockDB)
	mockDB.On("Preload", "Lead").Return(mockDB)
	mockDB.On("Where", "id = ?", estimateID).Return(mockDB)
	mockDB.On("First", mock.Anything).Return(errors.New("db error"))

	req, _ := http.NewRequest("POST", "/estimates/EST123/stages", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: estimateID}}
	c.Set("db", mockDB)

	handler.CreateStagesForEstimate(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "DATABASE_ERROR", response.Code)
}

// func TestCreateStagesForEstimate_CreateStageError(t *testing.T) {

// 	mockDB := new(mocks.MockDatabase)

// 	handler := &handlers.Handler{DB: mockDB}

// 	estimateID := "EST123"

// 	leadID := "LEAD123"

// 	estimate := models.Estimate{

// 		ID:     estimateID,

// 		LeadID: &leadID,

// 		ResidentialEntries: []models.ResidentialEntry{

// 			{ID: 1},

// 		},

// 	}

// 	mockDB.On("Preload", "ResidentialEntries").Return(mockDB)

// 	mockDB.On("Preload", "Lead").Return(mockDB)

// 	mockDB.On("Where", "id = ?", estimateID).Return(mockDB)

// 	mockDB.On("First", mock.Anything).Run(func(args mock.Arguments) {

// 		arg := args.Get(0).(*models.Estimate)

// 		*arg = estimate

// 	}).Return(nil)

// 	mockDB.On("Begin").Return(mockDB)

// 	mockDB.On("Where", "residential_entry_id = ?", uint(1)).Return(mockDB)

// 	mockDB.On("Find", mock.Anything).Return(gorm.ErrRecordNotFound)

// 	mockDB.On("Model", mock.Anything).Return(mockDB)

// 	mockDB.On("Count", mock.Anything).Return(mockDB).Run(func(args mock.Arguments) {

// 		arg := args.Get(0).(*int64)

// 		*arg = 0

// 	})

// 	mockDB.On("Create", mock.Anything).Return(errors.New("db error"))

// 	mockDB.On("Rollback").Return(nil)

// 	req, _ := http.NewRequest("POST", "/estimates/EST123/stages", nil)

// 	w := httptest.NewRecorder()

// 	c, _ := gin.CreateTestContext(w)

// 	c.Request = req

// 	c.Params = gin.Params{{Key: "id", Value: estimateID}}

// 	c.Set("db", mockDB)

// 	handler.CreateStagesForEstimate(c)

// 	assert.Equal(t, http.StatusInternalServerError, w.Code)

// 	var response ErrorResponse

// 	json.Unmarshal(w.Body.Bytes(), &response)

// 	assert.Equal(t, "DATABASE_ERROR", response.Code)

// }

// func TestCreateStagesForEstimate_CommitError(t *testing.T) {

// 	mockDB := new(mocks.MockDatabase)

// 	handler := &handlers.Handler{DB: mockDB}

// 	estimateID := "EST123"

// 	leadID := "LEAD123"

// 	estimate := models.Estimate{

// 		ID:     estimateID,

// 		LeadID: &leadID,

// 		ResidentialEntries: []models.ResidentialEntry{

// 			{ID: 1},

// 		},

// 	}

// 	mockDB.On("Preload", "ResidentialEntries").Return(mockDB)

// 	mockDB.On("Preload", "Lead").Return(mockDB)

// 	mockDB.On("Where", "id = ?", estimateID).Return(mockDB)

// 	mockDB.On("First", mock.Anything).Run(func(args mock.Arguments) {

// 		arg := args.Get(0).(*models.Estimate)

// 		*arg = estimate

// 	}).Return(nil)

// 	mockDB.On("Begin").Return(mockDB)

// 	mockDB.On("Where", "residential_entry_id = ?", uint(1)).Return(mockDB)

// 	mockDB.On("Find", mock.Anything).Return(gorm.ErrRecordNotFound)

// 	mockDB.On("Model", mock.Anything).Return(mockDB)

// 	mockDB.On("Count", mock.Anything).Return(mockDB).Run(func(args mock.Arguments) {

// 		arg := args.Get(0).(*int64)

// 		*arg = 0

// 	})

// 	mockDB.On("Create", mock.Anything).Return(nil)

// 	mockDB.On("Commit").Return(errors.New("db error"))

// 	mockDB.On("Rollback").Return(nil)

// 	req, _ := http.NewRequest("POST", "/estimates/EST123/stages", nil)

// 	w := httptest.NewRecorder()

// 	c, _ := gin.CreateTestContext(w)

// 	c.Request = req

// 	c.Params = gin.Params{{Key: "id", Value: estimateID}}

// 	c.Set("db", mockDB)

// 	handler.CreateStagesForEstimate(c)

// 	assert.Equal(t, http.StatusInternalServerError, w.Code)

// 	var response ErrorResponse

// 	json.Unmarshal(w.Body.Bytes(), &response)

// 	assert.Equal(t, "DATABASE_ERROR", response.Code)

// }

func TestCreateStagesForEstimate_CheckExistingStagesError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	estimateID := "EST123"
	leadID := "LEAD123"

	estimate := models.Estimate{
		ID:     estimateID,
		LeadID: &leadID,
		ResidentialEntries: []models.ResidentialEntry{
			{ID: 1},
		},
	}

	mockDB.On("Preload", "ResidentialEntries").Return(mockDB)
	mockDB.On("Preload", "Lead").Return(mockDB)
	mockDB.On("Where", "id = ?", estimateID).Return(mockDB)
	mockDB.On("First", mock.Anything).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Estimate)
		*arg = estimate
	}).Return(nil)

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Where", "residential_entry_id = ?", uint(1)).Return(mockDB)
	mockDB.On("Find", mock.Anything).Return(errors.New("db error"))
	mockDB.On("Rollback").Return(nil)

	req, _ := http.NewRequest("POST", "/estimates/EST123/stages", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: estimateID}}
	c.Set("db", mockDB)

	handler.CreateStagesForEstimate(c)

	assert.Equal(t, http.StatusOK, w.Code)

}

func TestPromoteVersionToBase_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	baseID := "EST123"
	versionID := "EST123-v1"

	versionEstimate := models.Estimate{ID: versionID}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Where", "id = ?", versionID).Return(mockDB)
	mockDB.On("First", mock.Anything).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Estimate)
		*arg = versionEstimate
	}).Return(nil)

	mockDB.On("Where", "estimate_id = ?", baseID).Return(mockDB)
	mockDB.On("Delete", &models.MaterialRow{}).Return(nil)
	mockDB.On("Delete", &models.ServiceRow{}).Return(nil)
	mockDB.On("Delete", &models.ResidentialEntry{}).Return(nil)

	mockDB.On("Unscoped").Return(mockDB)
	mockDB.On("Where", "id = ?", baseID).Return(mockDB)
	mockDB.On("Delete", &models.Estimate{}).Return(nil)

	mockDB.On("Model", &models.Estimate{}).Return(mockDB)
	mockDB.On("Where", "id = ?", versionID).Return(mockDB)
	mockDB.On("Update", "id", baseID).Return(nil)

	mockDB.On("Model", &models.MaterialRow{}).Return(mockDB)
	mockDB.On("Where", "estimate_id = ?", versionID).Return(mockDB)
	mockDB.On("Update", "estimate_id", baseID).Return(nil)

	mockDB.On("Model", &models.ServiceRow{}).Return(mockDB)
	mockDB.On("Where", "estimate_id = ?", versionID).Return(mockDB)
	mockDB.On("Update", "estimate_id", baseID).Return(nil)

	mockDB.On("Where", "estimate_id = ?", versionID).Return(mockDB)
	mockDB.On("Find", mock.Anything).Return(nil)
	mockDB.On("Model", &models.ResidentialEntry{}).Return(mockDB)

	versionPattern := baseID + "-v%"
	mockDB.On("Where", "estimate_id LIKE ? AND estimate_id != ?", versionPattern, baseID).Return(mockDB)
	mockDB.On("Delete", &models.MaterialRow{}).Return(nil)
	mockDB.On("Delete", &models.ServiceRow{}).Return(nil)
	mockDB.On("Delete", &models.ResidentialEntry{}).Return(nil)

	mockDB.On("Unscoped").Return(mockDB)
	mockDB.On("Where", "id LIKE ? AND id != ?", versionPattern, baseID).Return(mockDB)
	mockDB.On("Delete", &models.Estimate{}).Return(nil)

	mockDB.On("Commit").Return(nil)
	mockDB.On("Rollback").Return(nil).Maybe()

	req, _ := http.NewRequest("POST", fmt.Sprintf("/estimates/%s/promote", versionID), nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: versionID}}
	c.Set("db", mockDB)

	handler.PromoteVersionToBase(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Version promoted to base and all other versions deleted", response["message"])
}

func TestPromoteVersionToBase_DeleteVersionsForBase(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	baseID := "EST123"

	mockDB.On("Begin").Return(mockDB)

	versionPattern := baseID + "-v%"
	mockDB.On("Where", "estimate_id LIKE ? AND estimate_id != ?", versionPattern, baseID).Return(mockDB)
	mockDB.On("Delete", &models.MaterialRow{}).Return(nil)
	mockDB.On("Delete", &models.ServiceRow{}).Return(nil)
	mockDB.On("Delete", &models.ResidentialEntry{}).Return(nil)

	mockDB.On("Unscoped").Return(mockDB)
	mockDB.On("Where", "id LIKE ? AND id != ?", versionPattern, baseID).Return(mockDB)
	mockDB.On("Delete", &models.Estimate{}).Return(nil)

	mockDB.On("Commit").Return(nil)
	mockDB.On("Rollback").Return(nil).Maybe()

	req, _ := http.NewRequest("POST", fmt.Sprintf("/estimates/%s/promote", baseID), nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: baseID}}
	c.Set("db", mockDB)

	handler.PromoteVersionToBase(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "All versioned estimates deleted, base unchanged", response["message"])
}

func TestPromoteVersionToBase_InvalidID(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("POST", "/estimates//promote", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: ""}}
	c.Set("db", mockDB)

	handler.PromoteVersionToBase(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "VALIDATION_ERROR", response.Code)
}

func TestPromoteVersionToBase_VersionNotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	versionID := "EST123-v99"

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Where", "id = ?", versionID).Return(mockDB)
	mockDB.On("First", mock.Anything).Return(gorm.ErrRecordNotFound)
	mockDB.On("Rollback").Return(nil)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/estimates/%s/promote", versionID), nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: versionID}}
	c.Set("db", mockDB)

	handler.PromoteVersionToBase(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var response ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "NOT_FOUND", response.Code)
}

func TestPromoteVersionToBase_DBError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	baseID := "EST123"
	versionID := "EST123-v1"

	versionEstimate := models.Estimate{
		ID: versionID,
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Where", "id = ?", versionID).Return(mockDB)
	mockDB.On("First", mock.Anything).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Estimate)
		*arg = versionEstimate
	}).Return(nil)

	mockDB.On("Where", "estimate_id = ?", baseID).Return(mockDB)
	mockDB.On("Delete", &models.MaterialRow{}).Return(nil)
	mockDB.On("Delete", &models.ServiceRow{}).Return(nil)
	mockDB.On("Delete", &models.ResidentialEntry{}).Return(nil)

	mockDB.On("Unscoped").Return(mockDB)
	mockDB.On("Where", "id = ?", baseID).Return(mockDB)
	mockDB.On("Delete", &models.Estimate{}).Return(errors.New("db error"))

	mockDB.On("Rollback").Return(nil)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/estimates/%s/promote", versionID), nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: versionID}}
	c.Set("db", mockDB)

	handler.PromoteVersionToBase(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "DATABASE_ERROR", response.Code)
}

// ------------------CreateEstimate----------------
func TestCreateEstimate_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	leadID := "LEAD123"
	taxID := uint(1)
	discountID := uint(1)
	input := models.Estimate{
		Name:               "Test Estimate",
		LeadID:             &leadID,
		Details:            models.EstimateDetails{TaxID: &taxID, DiscountID: &discountID},
		ResidentialEntries: []models.ResidentialEntry{{ID: 1}},
		MaterialRows: []models.MaterialRow{
			{Details: models.MaterialRowDetails{Material: "Material 1"}},
		},
		ServiceRows: []models.ServiceRow{
			{Details: models.ServiceRowDetails{Name: "Service 1"}},
		},
	}
	body, _ := json.Marshal(input)

	// Mock DB calls
	mockDB.On("Model", &models.Agent{}).Return(mockDB)
	mockDB.On("Count", mock.Anything).Return(mockDB).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*int64)
		*arg = 0
	})
	mockDB.On("First", mock.AnythingOfType("*models.CountyTax"), "id = ?", &taxID).Return(nil)
	mockDB.On("First", mock.AnythingOfType("*models.Discount"), "id = ?", &discountID).Return(nil)
	mockDB.On("First", mock.AnythingOfType("*models.Lead"), "id = ?", &leadID).Return(nil)
	mockDB.On("Save", mock.AnythingOfType("*models.Lead")).Return(nil)
	mockDB.On("Where", "id = ?", mock.Anything).Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", uint(1)).Return(gorm.ErrRecordNotFound)
	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Create", mock.Anything).Return(nil)
	mockDB.On("First", &models.ResidentialEntry{}).Return(gorm.ErrRecordNotFound)
	mockDB.On("Exec", `SELECT setval('residential_entries_id_seq', (SELECT COALESCE(MAX(id), 1) FROM residential_entries))`).Return(nil)
	mockDB.On("Save", mock.AnythingOfType("*models.MaterialRow")).Return(nil)
	mockDB.On("Save", mock.AnythingOfType("*models.ServiceRow")).Return(nil)
	mockDB.On("Commit").Return(nil)
	mockDB.On("Rollback").Return(nil).Maybe()

	req, _ := http.NewRequest("POST", "/estimates", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateEstimate(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "estimate created successfully", resp["message"])
}

func TestCreateEstimate_SaveLeadError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	leadID := "LEAD123"
	taxID := uint(1)
	discountID := uint(1)

	input := models.Estimate{
		Name:   "Test Estimate",
		LeadID: &leadID,
		Details: models.EstimateDetails{
			TaxID:      &taxID,
			DiscountID: &discountID,
		},
	}
	body, _ := json.Marshal(input)

	mockDB.On("Model", &models.Agent{}).Return(mockDB)
	mockDB.On("Count", mock.Anything).Return(mockDB).
		Run(func(args mock.Arguments) {
			*args.Get(0).(*int64) = 0
		})
	mockDB.On("Where", "id = ?", mock.Anything).Return(mockDB)

	mockDB.On(
		"First",
		mock.AnythingOfType("*models.CountyTax"),
		"id = ?",
		&taxID,
	).Return(nil)

	mockDB.On(
		"First",
		mock.AnythingOfType("*models.Discount"),
		"id = ?",
		&discountID,
	).Return(nil)

	mockDB.On(
		"First",
		mock.AnythingOfType("*models.Lead"),
		"id = ?",
		&leadID,
	).Return(nil)

	mockDB.On("Save", mock.Anything).Return(errors.New("db error"))

	req, _ := http.NewRequest("POST", "/estimates", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateEstimate(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCreateEstimate_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	// HTTP Request
	req, _ := http.NewRequest("POST", "/estimates", bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateEstimate(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
}

func TestCreateEstimate_CountyTaxNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	leadID := "LEAD123"
	taxID := uint(99)
	discountID := uint(1)

	input := models.Estimate{
		Name:    "Test Estimate",
		LeadID:  &leadID,
		Details: models.EstimateDetails{TaxID: &taxID, DiscountID: &discountID},
	}
	body, _ := json.Marshal(input)

	mockDB.On("Model", &models.Agent{}).Return(mockDB)

	mockDB.On("Count", mock.Anything).Return(mockDB).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*int64)
		*arg = 0
	})

	// REQUIRED: GenerateEmployeeID
	mockDB.On("Where", "id = ?", mock.Anything).Return(mockDB)

	// CountyTax NOT found
	mockDB.On(
		"First",
		mock.AnythingOfType("*models.CountyTax"),
		"id = ?",
		&taxID,
	).Return(gorm.ErrRecordNotFound)

	req, _ := http.NewRequest("POST", "/estimates", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateEstimate(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp ErrorResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "NOT_FOUND", resp.Code)
}

func TestCreateEstimate_DiscountNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	leadID := "LEAD123"
	taxID := uint(1)
	discountID := uint(99)

	input := models.Estimate{
		Name:    "Test Estimate",
		LeadID:  &leadID,
		Details: models.EstimateDetails{TaxID: &taxID, DiscountID: &discountID},
	}
	body, _ := json.Marshal(input)

	mockDB.On("Model", &models.Agent{}).Return(mockDB)

	mockDB.On("Count", mock.Anything).Return(mockDB).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*int64)
		*arg = 0
	})

	// Required for GenerateEmployeeID
	mockDB.On("Where", "id = ?", mock.Anything).Return(mockDB)

	// County tax exists
	mockDB.On(
		"First",
		mock.AnythingOfType("*models.CountyTax"),
		"id = ?",
		&taxID,
	).Return(nil)

	// Discount NOT found
	mockDB.On(
		"First",
		mock.AnythingOfType("*models.Discount"),
		"id = ?",
		&discountID,
	).Return(gorm.ErrRecordNotFound)

	req, _ := http.NewRequest("POST", "/estimates", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateEstimate(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "NOT_FOUND", resp.Code)
}

func TestCreateEstimate_LeadNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	leadID := "LEAD999"
	taxID := uint(1)
	discountID := uint(1)

	input := models.Estimate{
		Name:   "Test Estimate",
		LeadID: &leadID,
		Details: models.EstimateDetails{
			TaxID:      &taxID,
			DiscountID: &discountID,
		},
	}
	body, _ := json.Marshal(input)

	// Agent count (GenerateEmployeeID path)
	mockDB.On("Model", &models.Agent{}).Return(mockDB)
	mockDB.On("Count", mock.Anything).Return(mockDB).
		Run(func(args mock.Arguments) {
			*args.Get(0).(*int64) = 0
		})

	// REQUIRED for GenerateEmployeeID
	mockDB.On("Where", "id = ?", mock.Anything).Return(mockDB)

	// CountyTax FOUND
	mockDB.On(
		"First",
		mock.AnythingOfType("*models.CountyTax"),
		"id = ?",
		&taxID,
	).Return(nil)

	// Discount FOUND
	mockDB.On(
		"First",
		mock.AnythingOfType("*models.Discount"),
		"id = ?",
		&discountID,
	).Return(nil)

	// Lead NOT FOUND
	mockDB.On(
		"First",
		mock.AnythingOfType("*models.Lead"),
		"id = ?",
		&leadID,
	).Return(gorm.ErrRecordNotFound)

	req, _ := http.NewRequest("POST", "/estimates", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateEstimate(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp ErrorResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
}

func TestCreateEstimate_CreateEstimateError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	leadID := "LEAD123"
	taxID := uint(1)
	discountID := uint(1)
	input := models.Estimate{
		Name:               "Test Estimate",
		LeadID:             &leadID,
		Details:            models.EstimateDetails{TaxID: &taxID, DiscountID: &discountID},
		ResidentialEntries: []models.ResidentialEntry{{ID: 1}},
	}
	body, _ := json.Marshal(input)

	// Mock DB calls
	mockDB.On("Model", &models.Agent{}).Return(mockDB)
	mockDB.On("Count", mock.Anything).Return(mockDB).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*int64)
		*arg = 0
	})
	mockDB.On("First", mock.AnythingOfType("*models.CountyTax"), "id = ?", &taxID).Return(nil)
	mockDB.On("First", mock.AnythingOfType("*models.Discount"), "id = ?", &discountID).Return(nil)
	mockDB.On("First", mock.AnythingOfType("*models.Lead"), "id = ?", &leadID).Return(nil)
	mockDB.On("Save", mock.AnythingOfType("*models.Lead")).Return(nil)
	mockDB.On("Where", "id = ?", mock.Anything).Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", uint(1)).Return(gorm.ErrRecordNotFound)
	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Create", mock.Anything).Return(errors.New("db error"))
	mockDB.On("Rollback").Return(nil)
	mockDB.On("First", &models.ResidentialEntry{}).Return(gorm.ErrRecordNotFound)
	mockDB.On("Exec", `SELECT setval('residential_entries_id_seq', (SELECT COALESCE(MAX(id), 1) FROM residential_entries))`).Return(nil)

	req, _ := http.NewRequest("POST", "/estimates", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateEstimate(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "DATABASE_ERROR", resp.Code)
}

func TestCreateEstimate_CommitError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	leadID := "LEAD123"
	taxID := uint(1)
	discountID := uint(1)
	input := models.Estimate{
		Name:               "Test Estimate",
		LeadID:             &leadID,
		Details:            models.EstimateDetails{TaxID: &taxID, DiscountID: &discountID},
		ResidentialEntries: []models.ResidentialEntry{{ID: 1}},
	}
	body, _ := json.Marshal(input)

	// Mock DB calls
	mockDB.On("Model", &models.Agent{}).Return(mockDB)
	mockDB.On("Count", mock.Anything).Return(mockDB).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*int64)
		*arg = 0
	})
	mockDB.On("First", mock.AnythingOfType("*models.CountyTax"), "id = ?", &taxID).Return(nil)
	mockDB.On("First", mock.AnythingOfType("*models.Discount"), "id = ?", &discountID).Return(nil)
	mockDB.On("First", mock.AnythingOfType("*models.Lead"), "id = ?", &leadID).Return(nil)
	mockDB.On("Save", mock.AnythingOfType("*models.Lead")).Return(nil)
	mockDB.On("Where", "id = ?", mock.Anything).Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", uint(1)).Return(gorm.ErrRecordNotFound)
	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Create", mock.Anything).Return(nil)
	mockDB.On("Commit").Return(errors.New("db error"))
	mockDB.On("Rollback").Return(nil)
	mockDB.On("First", &models.ResidentialEntry{}).Return(gorm.ErrRecordNotFound)
	mockDB.On("Exec", `SELECT setval('residential_entries_id_seq', (SELECT COALESCE(MAX(id), 1) FROM residential_entries))`).Return(nil)

	req, _ := http.NewRequest("POST", "/estimates", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateEstimate(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "DATABASE_ERROR", resp.Code)
}

func TestCreateEstimate_InvalidManagerResidentialEntryID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	leadID := "LEAD123"
	taxID := uint(1)
	discountID := uint(1)
	invalidManagerResidentialEntryID := uint(999)

	input := models.Estimate{
		Name:   "Test Estimate",
		LeadID: &leadID,
		Details: models.EstimateDetails{
			TaxID:      &taxID,
			DiscountID: &discountID,
		},
		ResidentialEntries: []models.ResidentialEntry{
			{
				ManagerResidentialEntryID: invalidManagerResidentialEntryID,
			},
		},
	}
	body, _ := json.Marshal(input)

	// Mocks for validations before residential entries
	mockDB.On("Model", &models.Agent{}).Return(mockDB).Maybe()
	mockDB.On("Count", mock.Anything).Return(mockDB).Run(func(args mock.Arguments) {
		*args.Get(0).(*int64) = 0
	}).Maybe()
	mockDB.On("Where", "id = ?", mock.Anything).Return(mockDB).Maybe()
	mockDB.On("First", mock.AnythingOfType("*models.CountyTax"), "id = ?", &taxID).Return(nil)
	mockDB.On("First", mock.AnythingOfType("*models.Discount"), "id = ?", &discountID).Return(nil)
	mockDB.On("First", mock.AnythingOfType("*models.Lead"), "id = ?", &leadID).Return(nil)
	mockDB.On("Save", mock.AnythingOfType("*models.Lead")).Return(nil)

	// Mock for invalid ManagerResidentialEntryID
	mockDB.On("Where", "id = ?", invalidManagerResidentialEntryID).
		Return(mockDB)

	mockDB.On(
		"First",
		mock.AnythingOfType("*models.ManagerResidentialEntry"),
	).Return(gorm.ErrRecordNotFound)

	mockDB.On(
		"First",
		mock.AnythingOfType("*models.ResidentialEntry"),
	).Return(gorm.ErrRecordNotFound)

	req, _ := http.NewRequest("POST", "/estimates", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateEstimate(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
	assert.Equal(t, "Invalid ManagerResidentialEntryID", resp.Error)
	mockDB.AssertExpectations(t)
}

func TestCreateEstimate_CreateEstimateDBError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	leadID := "LEAD123"
	taxID := uint(1)
	discountID := uint(1)
	input := models.Estimate{
		Name:   "Test Estimate",
		LeadID: &leadID,
		Details: models.EstimateDetails{
			TaxID:      &taxID,
			DiscountID: &discountID,
		},
	}
	body, _ := json.Marshal(input)

	// Mock DB calls
	mockDB.On("Model", &models.Agent{}).Return(mockDB)
	mockDB.On("Count", mock.Anything).Return(mockDB).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*int64)
		*arg = 0
	})
	mockDB.On("First", mock.AnythingOfType("*models.CountyTax"), "id = ?", &taxID).Return(nil)
	mockDB.On("First", mock.AnythingOfType("*models.Discount"), "id = ?", &discountID).Return(nil)
	mockDB.On("First", mock.AnythingOfType("*models.Lead"), "id = ?", &leadID).Return(nil)
	mockDB.On("Save", mock.AnythingOfType("*models.Lead")).Return(nil)
	mockDB.On("Where", "id = ?", mock.Anything).Return(mockDB)
	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Create", mock.Anything).Return(errors.New("db error"))
	mockDB.On("Rollback").Return(nil)

	req, _ := http.NewRequest("POST", "/estimates", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateEstimate(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "DATABASE_ERROR", resp.Code)
}

func TestCreateEstimate_FindCountyTaxError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	leadID := "LEAD123"
	taxID := uint(1)
	discountID := uint(1)
	input := models.Estimate{
		Name:   "Test Estimate",
		LeadID: &leadID,
		Details: models.EstimateDetails{
			TaxID:      &taxID,
			DiscountID: &discountID,
		},
	}
	body, _ := json.Marshal(input)

	// Mock DB calls
	mockDB.On("Model", &models.Agent{}).Return(mockDB)
	mockDB.On("Count", mock.Anything).Return(mockDB).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*int64)
		*arg = 0
	})
	mockDB.On("First", mock.AnythingOfType("*models.CountyTax"), "id = ?", &taxID).Return(errors.New("db error"))
	mockDB.On("Where", "id = ?", mock.Anything).Return(mockDB)

	req, _ := http.NewRequest("POST", "/estimates", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateEstimate(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "DATABASE_ERROR", resp.Code)
}

func TestCreateEstimate_FindDiscountError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	leadID := "LEAD123"
	taxID := uint(1)
	discountID := uint(1)
	input := models.Estimate{
		Name:   "Test Estimate",
		LeadID: &leadID,
		Details: models.EstimateDetails{
			TaxID:      &taxID,
			DiscountID: &discountID,
		},
	}
	body, _ := json.Marshal(input)

	// Mock DB calls
	mockDB.On("Model", &models.Agent{}).Return(mockDB)
	mockDB.On("Count", mock.Anything).Return(mockDB).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*int64)
		*arg = 0
	})
	mockDB.On("First", mock.AnythingOfType("*models.CountyTax"), "id = ?", &taxID).Return(nil)
	mockDB.On("First", mock.AnythingOfType("*models.Discount"), "id = ?", &discountID).Return(errors.New("db error"))
	mockDB.On("Where", "id = ?", mock.Anything).Return(mockDB)

	req, _ := http.NewRequest("POST", "/estimates", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateEstimate(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "DATABASE_ERROR", resp.Code)
}

func TestCreateEstimate_InvalidManagerResidentialTypeID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	leadID := "LEAD123"
	taxID := uint(1)
	discountID := uint(1)
	validManagerResidentialEntryID := uint(111)
	invalidManagerResidentialTypeID := uint(888)

	input := models.Estimate{
		Name:   "Test Estimate",
		LeadID: &leadID,
		Details: models.EstimateDetails{
			TaxID:      &taxID,
			DiscountID: &discountID,
		},
		ResidentialEntries: []models.ResidentialEntry{
			{
				ManagerResidentialEntryID: validManagerResidentialEntryID,
				ManagerResidentialTypeID:  invalidManagerResidentialTypeID,
			},
		},
	}
	body, _ := json.Marshal(input)

	// Mocks for validations before residential entries
	mockDB.On("Model", &models.Agent{}).Return(mockDB).Maybe()
	mockDB.On("Count", mock.Anything).Return(mockDB).Run(func(args mock.Arguments) {
		*args.Get(0).(*int64) = 0
	}).Maybe()
	mockDB.On("Where", "id = ?", mock.Anything).Return(mockDB).Maybe()

	mockDB.On("First", mock.AnythingOfType("*models.CountyTax"), "id = ?", &taxID).
		Return(nil)
	mockDB.On("First", mock.AnythingOfType("*models.Discount"), "id = ?", &discountID).
		Return(nil)
	mockDB.On("First", mock.AnythingOfType("*models.Lead"), "id = ?", &leadID).
		Return(nil)

		// Lead save (already fixed earlier)
	mockDB.On("Save", mock.AnythingOfType("*models.Lead")).
		Return(nil)

	// Valid ManagerResidentialEntry
	mockDB.On("Where", "id = ?", validManagerResidentialEntryID).
		Return(mockDB)
	mockDB.On("First", mock.AnythingOfType("*models.ManagerResidentialEntry")).
		Run(func(args mock.Arguments) {
			*args.Get(0).(*models.ManagerResidentialEntry) =
				models.ManagerResidentialEntry{ID: validManagerResidentialEntryID}
		}).
		Return(nil)

	// Invalid ManagerResidentialType
	mockDB.On("Where", "id = ?", invalidManagerResidentialTypeID).
		Return(mockDB)
	mockDB.On("First", mock.AnythingOfType("*models.ManagerResidentialType")).
		Return(gorm.ErrRecordNotFound)

	// ✅ MISSING MOCK (this fixes the panic)
	mockDB.On(
		"First",
		mock.AnythingOfType("*models.ResidentialEntry"),
		mock.Anything,
	).Return(gorm.ErrRecordNotFound)

	req, _ := http.NewRequest("POST", "/estimates", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateEstimate(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
	assert.Equal(t, "Invalid ManagerResidentialTypeID", resp.Error)
	mockDB.AssertExpectations(t)
}
