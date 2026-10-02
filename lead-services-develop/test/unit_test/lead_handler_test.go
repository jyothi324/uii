package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
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

// TestCreateLead_Success tests successful lead creation
func TestCreateLead_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("Model", mock.Anything).Return(mockDB)
	mockDB.On("Where", mock.Anything, mock.Anything).Return(mockDB)
	mockDB.On("Count", mock.Anything).Return(mockDB)
	mockDB.On("Begin").Return(mockDB).Once()
	mockDB.On("Create", mock.AnythingOfType("*models.Lead")).Return(nil).Once()
	mockDB.On("Commit").Return(nil).Once()
	// mockDB.On("Rollback").Return(nil)

	handler := &handlers.LeadHandler{}

	// ----- FIX: Create multipart body -----
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add sample form fields (your handler expects these)
	writer.WriteField("first_name", "John")
	writer.WriteField("last_name", "Doe")
	writer.WriteField("email", "john@example.com")
	writer.WriteField("phone_number", "+1234567890")
	writer.WriteField("JobName", "Sample Job")

	// If handler expects files, add example file:
	part, _ := writer.CreateFormFile("document", "test.pdf")
	part.Write([]byte("dummy pdf content"))

	writer.Close()

	// ----- Create request -----
	req := httptest.NewRequest("POST", "/leads", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateLead(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Lead created successfully", resp["message"])

	mockDB.AssertExpectations(t)

}

func TestCreateLead_MissingJobName(t *testing.T) {
	// Mock database
	mockDB := new(mocks.MockDatabase)

	// Create a new lead handler
	leadHandler := handlers.NewLeadHandler(nil, "dummy-bucket", "dummy-region")

	// Set up expectations for the mock database
	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Rollback").Return(nil)
	// mockDB.On("First", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	// Create a new Gin context for testing
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Create a multipart form request body without JobName
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.WriteField("DeliveryType", "DELIVERY")
	writer.WriteField("Status", "NEW")
	writer.Close()

	// Create a new HTTP request with the multipart form body
	req, _ := http.NewRequest("POST", "/leads", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Set the mock database in the Gin context
	c.Request = req
	c.Set("db", mockDB)

	// Call the CreateLead handler
	leadHandler.CreateLead(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Assert the response body
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "INTERNAL_SERVER_ERROR", resp.Code)
	assert.Equal(t, "Unexpected error occurred", resp.Error)

	// Assert that the mock database methods were called as expected
	mockDB.AssertExpectations(t)
}

// // ----------------- GetLead -----------------
func TestGetLead_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	expectedLead := models.Lead{
		ID:      "LEAD001",
		JobName: "Test Job",
		Status:  models.NEW,
	}

	mockDB.On("Model", mock.AnythingOfType("*models.Lead")).Return(mockDB)
	mockDB.On("Where", "id = ?", "LEAD001").Return(mockDB)
	mockDB.On("Preload", "Associates").Return(mockDB)
	mockDB.On("Preload", "ResidentialEntries.ManagerResidentialEntry").Return(mockDB)
	mockDB.On("Preload", "ResidentialEntries.ManagerResidentialType").Return(mockDB)
	mockDB.On("Preload", "Estimates.MaterialRows").Return(mockDB)
	mockDB.On("Preload", "Estimates.ServiceRows").Return(mockDB)
	mockDB.On("Preload", "Files").Return(mockDB)
	mockDB.On("Find", mock.AnythingOfType("*models.Lead")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Lead)
		*dest = expectedLead
	}).Return(nil)

	// Mock Association calls
	// mockDB.On("Model", mock.AnythingOfType("*models.Lead")).Return(mockDB)
	// mockDB.On("Association", "Associates").Return(&mocks.MockAssociation{})

	handler := &handlers.LeadHandler{}

	req, _ := http.NewRequest("GET", "/leads/LEAD001", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "LEAD001"}}
	c.Set("db", mockDB)

	handler.GetLead(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Lead retrieved successfully", resp["message"])

	mockDB.AssertExpectations(t)
}

func TestGetLead_InvalidID(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.LeadHandler{}

	req, _ := http.NewRequest("GET", "/leads/", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: ""}}
	c.Set("db", mockDB)

	handler.GetLead(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "BAD_REQUEST", resp.Code)
}

func TestGetLead_NotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("Model", mock.AnythingOfType("*models.Lead")).Return(mockDB)
	mockDB.On("Where", "id = ?", "INVALID").Return(mockDB)
	mockDB.On("Preload", mock.Anything).Return(mockDB).Times(6)
	mockDB.On("Find", mock.AnythingOfType("*models.Lead")).Return(errors.New("record not found"))

	handler := &handlers.LeadHandler{}

	req, _ := http.NewRequest("GET", "/leads/INVALID", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "INVALID"}}
	c.Set("db", mockDB)

	handler.GetLead(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "DATABASE_ERROR", resp.Code)

	mockDB.AssertExpectations(t)
}

// ----------------- GetAllLeads -----------------
func TestGetAllLeads_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	leadsList := []models.Lead{
		{ID: "LEAD001", JobName: "Job 1", Status: models.NEW},
		{ID: "LEAD002", JobName: "Job 2", Status: models.IN_PROGRESS},
	}

	var total int64 = 2

	// Mock Count
	mockDB.On("Model", mock.AnythingOfType("*models.Lead")).Return(mockDB)
	mockDB.On("Count", mock.AnythingOfType("*int64")).Run(func(args mock.Arguments) {
		countPtr := args.Get(0).(*int64)
		*countPtr = total
	}).Return(mockDB)
	mockDB.On("Error").Return(nil).Once()

	// Mock Find with preloads
	mockDB.On("Model", mock.AnythingOfType("*models.Lead")).Return(mockDB)
	mockDB.On("Preload", "Associates").Return(mockDB)
	mockDB.On("Preload", "ResidentialEntries.ManagerResidentialEntry").Return(mockDB)
	mockDB.On("Preload", "ResidentialEntries.ManagerResidentialType").Return(mockDB)
	mockDB.On("Preload", "Estimates.MaterialRows").Return(mockDB)
	mockDB.On("Preload", "Estimates.ServiceRows").Return(mockDB)
	mockDB.On("Preload", "Files").Return(mockDB)
	mockDB.On("Order", "created_at desc").Return(mockDB)
	mockDB.On("Find", mock.AnythingOfType("*[]models.Lead")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*[]models.Lead)
		*dest = leadsList
	}).Return(nil)

	handler := &handlers.LeadHandler{}

	req, _ := http.NewRequest("GET", "/leads", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.GetAllLeads(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Leads retrieved successfully", resp["message"])

	data, ok := resp["data"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(data))

	mockDB.AssertExpectations(t)
}

// ----------------- DeleteLead -----------------
func TestDeleteLead_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	existingLead := models.Lead{
		ID:      "LEAD001",
		JobName: "Test Job",
		Status:  models.NEW,
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Preload", "Files").Return(mockDB)
	mockDB.On("First", mock.AnythingOfType("*models.Lead"), "id = ?", "LEAD001").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Lead)
		*dest = existingLead
	}).Return(nil)

	// Mock deletion operations
	mockDB.On("Exec", "DELETE FROM lead_associates WHERE lead_id = ?", "LEAD001").Return(nil)
	mockDB.On("Where", "lead_id = ?", "LEAD001").Return(mockDB).Times(4)
	mockDB.On("Delete", mock.AnythingOfType("*models.File")).Return(nil)
	mockDB.On("Delete", mock.AnythingOfType("*models.ResidentialEntry")).Return(nil)
	mockDB.On("Find", mock.AnythingOfType("*[]models.Estimate")).Return(nil)
	mockDB.On("Delete", mock.AnythingOfType("*models.Estimate")).Return(nil)
	mockDB.On("Delete", mock.AnythingOfType("*models.Lead")).Return(nil)
	mockDB.On("Commit").Return(nil)

	handler := &handlers.LeadHandler{}

	req, _ := http.NewRequest("DELETE", "/leads/LEAD001", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "LEAD001"}}
	c.Set("db", mockDB)

	handler.DeleteLead(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Lead deleted successfully", resp["message"])

	mockDB.AssertExpectations(t)
}

func TestDeleteLead_InvalidID(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.LeadHandler{}

	req, _ := http.NewRequest("DELETE", "/leads/", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: ""}}
	c.Set("db", mockDB)

	handler.DeleteLead(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "BAD_REQUEST", resp.Code)
}

func TestDeleteLead_NotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Preload", "Files").Return(mockDB)
	mockDB.On("First", mock.AnythingOfType("*models.Lead"), "id = ?", "INVALID").Return(gorm.ErrRecordNotFound)
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.LeadHandler{}

	req, _ := http.NewRequest("DELETE", "/leads/INVALID", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "INVALID"}}
	c.Set("db", mockDB)

	handler.DeleteLead(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "NOT_FOUND", resp.Code)

	mockDB.AssertExpectations(t)
}

// ----------------- ToggleFavorite -----------------
func TestToggleFavorite_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	existingLead := models.Lead{
		ID:       "LEAD001",
		JobName:  "Test Job",
		Status:   models.NEW,
		Favorite: "false",
	}

	updateData := map[string]string{
		"favorite": "true",
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.AnythingOfType("*models.Lead"), "id = ?", "LEAD001").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Lead)
		*dest = existingLead
	}).Return(nil)
	mockDB.On("Save", mock.AnythingOfType("*models.Lead")).Return(nil)
	mockDB.On("Commit").Return(nil)

	handler := &handlers.LeadHandler{}

	data, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", "/leads/LEAD001/favorite", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "LEAD001"}}
	c.Set("db", mockDB)

	handler.ToggleFavorite(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Lead favorite status updated successfully", resp["message"])

	mockDB.AssertExpectations(t)
}

func TestToggleFavorite_InvalidID(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.LeadHandler{}

	req, _ := http.NewRequest("PUT", "/leads//favorite", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: ""}}
	c.Set("db", mockDB)

	handler.ToggleFavorite(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "BAD_REQUEST", resp.Code)
}

// ----------------- UpdateStatus -----------------
func TestUpdateStatus_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	existingLead := models.Lead{
		ID:      "LEAD001",
		JobName: "Test Job",
		Status:  models.NEW,
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.AnythingOfType("*models.Lead"), "id = ?", "LEAD001").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Lead)
		*dest = existingLead
	}).Return(nil)
	mockDB.On("Save", mock.AnythingOfType("*models.Lead")).Return(nil)
	mockDB.On("Commit").Return(nil)

	handler := &handlers.LeadHandler{}

	req, _ := http.NewRequest("PUT", "/leads/LEAD001/status?status=IN_PROGRESS", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "LEAD001"}}
	c.Set("db", mockDB)

	handler.UpdateStatus(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Lead favorite status updated successfully", resp["message"])

	mockDB.AssertExpectations(t)
}

func TestUpdateStatus_RejectedWithReason(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	existingLead := models.Lead{
		ID:      "LEAD001",
		JobName: "Test Job",
		Status:  models.NEW,
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.AnythingOfType("*models.Lead"), "id = ?", "LEAD001").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Lead)
		*dest = existingLead
	}).Return(nil)
	mockDB.On("Save", mock.AnythingOfType("*models.Lead")).Return(nil)
	mockDB.On("Commit").Return(nil)

	handler := &handlers.LeadHandler{}

	req, _ := http.NewRequest("PUT", "/leads/LEAD001/status?status=REJECTED&rejection_reason=Not+qualified", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "LEAD001"}}
	c.Set("db", mockDB)

	handler.UpdateStatus(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"Status":"REJECTED"`)
	assert.Contains(t, w.Body.String(), `"RejectionReason":"Not qualified"`)
	mockDB.AssertExpectations(t)
}

func TestUpdateStatus_RejectedWithoutReason(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	existingLead := models.Lead{
		ID:      "LEAD001",
		JobName: "Test Job",
		Status:  models.NEW,
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.AnythingOfType("*models.Lead"), "id = ?", "LEAD001").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Lead)
		*dest = existingLead
	}).Return(nil)

	handler := &handlers.LeadHandler{}

	req, _ := http.NewRequest("PUT", "/leads/LEAD001/status?status=REJECTED", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "LEAD001"}}
	c.Set("db", mockDB)

	handler.UpdateStatus(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Rejection reason is required for REJECTED status", resp["error"])
}

// ----------------- GetLeadByPhoneOrEmail -----------------
func TestGetLeadByPhoneOrEmail_ByEmail_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	expectedLead := models.Lead{
		ID:      "LEAD001",
		JobName: "Test Job",
		Status:  models.NEW,
	}

	mockDB.On("Model", mock.AnythingOfType("*models.Lead")).Return(mockDB)
	mockDB.On("Preload", mock.Anything).Return(mockDB).Times(6)
	mockDB.On("Where", "details->>'ShipToEmail' = ?", "test@example.com").Return(mockDB)
	mockDB.On("First", mock.AnythingOfType("*models.Lead")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Lead)
		*dest = expectedLead
	}).Return(nil)

	// // Mock Association
	// mockDB.On("Model", mock.AnythingOfType("*models.Lead")).Return(mockDB)
	// mockDB.On("Association", "Associates").Return(&mocks.MockAssociation{})

	handler := &handlers.LeadHandler{}

	req, _ := http.NewRequest("GET", "/leads/search?email=test@example.com", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.GetLeadByPhoneOrEmail(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Lead retrieved successfully", resp["message"])
	fmt.Println(resp["data"], "DATA")

	mockDB.AssertExpectations(t)
}

func TestGetLeadByPhoneOrEmail_NoParameters(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.LeadHandler{}

	req, _ := http.NewRequest("GET", "/leads/search", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.GetLeadByPhoneOrEmail(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "BAD_REQUEST", resp.Code)
}

// ----------------- LeadDetails -----------------
func TestLeadDetails_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	existingLead := models.Lead{
		ID:      "LEAD001",
		JobName: "Test Job",
		Status:  models.NEW,
	}

	updateData := map[string]string{
		"BillToCustomer": "John Doe",
		"PaymentTerm":    "CASH",
		"PricingType":    "RETAIL",
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.AnythingOfType("*models.Lead"), "id = ?", "LEAD001").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Lead)
		*dest = existingLead
	}).Return(nil)
	mockDB.On("Save", mock.AnythingOfType("*models.Lead")).Return(nil)
	mockDB.On("Commit").Return(nil)

	handler := &handlers.LeadHandler{}

	data, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", "/leads/LEAD001/details", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "LEAD001"}}
	c.Set("db", mockDB)

	handler.LeadDetails(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Lead favorite status updated successfully", resp["message"])

	mockDB.AssertExpectations(t)
}

// -------------------UpdateLead---------------

// 1. Test for empty/missing ID parameter
func TestUpdateLead_EmptyID(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.LeadHandler{}

	req, _ := http.NewRequest("PUT", "/lead/", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: ""}}
	c.Set("db", mockDB)

	handler.UpdateLead(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "BAD_REQUEST", response.Code)
	assert.Equal(t, "Invalid lead ID", response.Error)
}

// 2. Test for failed multipart form parsing
func TestUpdateLead_ParseMultipartFormError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.LeadHandler{}

	req, _ := http.NewRequest("PUT", "/lead/LEAD123", bytes.NewBuffer([]byte("invalid data")))
	req.Header.Set("Content-Type", "multipart/form-data")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "LEAD123"}}
	c.Set("db", mockDB)

	handler.UpdateLead(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "VALIDATION_ERROR", response.Code)
	assert.Equal(t, "Failed to parse form", response.Error)
}

// 3. Test for lead not found
func TestUpdateLead_LeadNotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.LeadHandler{}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Rollback").Return(nil)
	mockDB.On("Preload", mock.Anything).Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "NOTFOUND").Return(gorm.ErrRecordNotFound)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.Close()

	req, _ := http.NewRequest("PUT", "/lead/NOTFOUND", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "NOTFOUND"}}
	c.Set("db", mockDB)

	handler.UpdateLead(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var response handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "NOT_FOUND", response.Code)
}

// 4. Test for database error when fetching lead
func TestUpdateLead_DatabaseErrorOnFetch(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.LeadHandler{}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Rollback").Return(nil)
	mockDB.On("Preload", mock.Anything).Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "LEAD123").Return(errors.New("database error"))

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.Close()

	req, _ := http.NewRequest("PUT", "/lead/LEAD123", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "LEAD123"}}
	c.Set("db", mockDB)

	handler.UpdateLead(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "DATABASE_ERROR", response.Code)
}

// 5. Test for missing required fields (JobName and Status)
func TestUpdateLead_MissingRequiredFields(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.LeadHandler{}

	existingLead := &models.Lead{
		ID:      "LEAD123",
		JobName: "",
		Status:  "",
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Rollback").Return(nil)
	mockDB.On("Preload", mock.Anything).Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "LEAD123").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Lead)
		*dest = *existingLead
	}).Return(nil)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("JobName", "")
	writer.WriteField("Status", "")
	writer.Close()

	req, _ := http.NewRequest("PUT", "/lead/LEAD123", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "LEAD123"}}
	c.Set("db", mockDB)

	handler.UpdateLead(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "VALIDATION_ERROR", response.Code)
	assert.Equal(t, "Missing required fields", response.Error)
}

// 6. Test for database error when fetching CountyTax
func TestUpdateLead_DatabaseErrorOnCountyTaxFetch(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.LeadHandler{}

	existingLead := &models.Lead{
		ID:      "LEAD123",
		JobName: "Test Job",
		Status:  models.NEW,
		Details: models.LeadDetails{
			CountyTax: &models.CountyTax{ID: 1},
		},
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Rollback").Return(nil)
	mockDB.On("Preload", mock.Anything).Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "LEAD123").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Lead)
		*dest = *existingLead
	}).Return(nil)
	mockDB.On("First", mock.Anything, "id = ?", uint(1)).Return(errors.New("database error"))

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("JobName", "Test Job")
	writer.WriteField("Status", string(models.NEW))
	writer.Close()

	req, _ := http.NewRequest("PUT", "/lead/LEAD123", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "LEAD123"}}
	c.Set("db", mockDB)

	handler.UpdateLead(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "DATABASE_ERROR", response.Code)
}

// 7. Test for database error when finding shipping address
func TestUpdateLead_DatabaseErrorOnShippingAddressFetch(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.LeadHandler{}

	shippingAddressID := uint(1)
	existingLead := &models.Lead{
		ID:      "LEAD123",
		JobName: "Test Job",
		Status:  models.NEW,
		Details: models.LeadDetails{
			ShippingAddressID: &shippingAddressID,
		},
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Rollback").Return(nil)
	mockDB.On("Preload", mock.Anything).Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "LEAD123").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Lead)
		*dest = *existingLead
	}).Return(nil)
	mockDB.On("First", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	mockDB.On("First", mock.Anything, uint(1)).Return(errors.New("database error"))

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("JobName", "Test Job")
	writer.WriteField("Status", string(models.NEW))
	writer.WriteField("Details[ShippingAddress][Street]", "123 Main St")
	writer.Close()

	req, _ := http.NewRequest("PUT", "/lead/LEAD123", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "LEAD123"}}
	c.Set("db", mockDB)

	handler.UpdateLead(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "INTERNAL_SERVER_ERROR", response.Code)
}

// 8. Test for database error when saving shipping address
func TestUpdateLead_DatabaseErrorOnShippingAddressSave(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.LeadHandler{}

	existingLead := &models.Lead{
		ID:      "LEAD123",
		JobName: "Test Job",
		Status:  models.NEW,
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Rollback").Return(nil)
	mockDB.On("Preload", mock.Anything).Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "LEAD123").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Lead)
		*dest = *existingLead
	}).Return(nil)
	mockDB.On("First", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockDB.On("Save", mock.MatchedBy(func(v interface{}) bool {
		_, ok := v.(*models.Address)
		return ok
	})).Return(errors.New("database error"))

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("JobName", "Test Job")
	writer.WriteField("Status", string(models.NEW))
	writer.WriteField("Details[ShippingAddress][Street]", "123 Main St")
	writer.Close()

	req, _ := http.NewRequest("PUT", "/lead/LEAD123", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "LEAD123"}}
	c.Set("db", mockDB)

	handler.UpdateLead(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "INTERNAL_SERVER_ERROR", response.Code)
	assert.Equal(t, "Unexpected error occurred", response.Error)
}

// 9. Test for database error when finding billing address
func TestUpdateLead_DatabaseErrorOnBillingAddressFetch(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.LeadHandler{}

	billingAddressID := uint(1)
	existingLead := &models.Lead{
		ID:      "LEAD123",
		JobName: "Test Job",
		Status:  models.NEW,
		Details: models.LeadDetails{
			BillingAddressID: &billingAddressID,
		},
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Rollback").Return(nil)
	mockDB.On("Preload", mock.Anything).Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "LEAD123").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Lead)
		*dest = *existingLead
	}).Return(nil)
	mockDB.On("First", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	mockDB.On("First", mock.Anything, uint(1)).Return(errors.New("database error"))

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("JobName", "Test Job")
	writer.WriteField("Status", string(models.NEW))
	writer.WriteField("Details[BillingAddress][Street]", "456 Billing St")
	writer.Close()

	req, _ := http.NewRequest("PUT", "/lead/LEAD123", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "LEAD123"}}
	c.Set("db", mockDB)

	handler.UpdateLead(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "INTERNAL_SERVER_ERROR", response.Code)
}

// 10. Test for database error when saving billing address
func TestUpdateLead_DatabaseErrorOnBillingAddressSave(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.LeadHandler{}

	existingLead := &models.Lead{
		ID:      "LEAD123",
		JobName: "Test Job",
		Status:  models.NEW,
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Rollback").Return(nil)
	mockDB.On("Preload", mock.Anything).Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "LEAD123").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Lead)
		*dest = *existingLead
	}).Return(nil)
	mockDB.On("First", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockDB.On("Save", mock.MatchedBy(func(v interface{}) bool {
		_, ok := v.(*models.Address)
		return ok
	})).Return(errors.New("database error"))

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("JobName", "Test Job")
	writer.WriteField("Status", string(models.NEW))
	writer.WriteField("Details[BillingAddress][Street]", "456 Billing St")
	writer.Close()

	req, _ := http.NewRequest("PUT", "/lead/LEAD123", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "LEAD123"}}
	c.Set("db", mockDB)

	handler.UpdateLead(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "INTERNAL_SERVER_ERROR", response.Code)
	assert.Equal(t, "Unexpected error occurred", response.Error)
}

// 11. Test for invalid pickup location
func TestUpdateLead_InvalidPickupLocation(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.LeadHandler{}

	existingLead := &models.Lead{
		ID:      "LEAD123",
		JobName: "Test Job",
		Status:  models.NEW,
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Rollback").Return(nil)
	mockDB.On("Preload", mock.Anything).Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "LEAD123").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Lead)
		*dest = *existingLead
	}).Return(nil)
	mockDB.On("First", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockDB.On("First", mock.Anything, "INVALID_PICKUP").Return(gorm.ErrRecordNotFound)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("JobName", "Test Job")
	writer.WriteField("Status", string(models.NEW))
	writer.WriteField("Details[IsPickUpDelivery]", "true")
	writer.WriteField("Details[PickupDeliveryAddressID]", "INVALID_PICKUP")
	writer.Close()

	req, _ := http.NewRequest("PUT", "/lead/LEAD123", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "LEAD123"}}
	c.Set("db", mockDB)

	handler.UpdateLead(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "INTERNAL_SERVER_ERROR", response.Code)
	assert.Equal(t, "Unexpected error occurred", response.Error)
}

// 12. Test for failed parsing of residential entries JSON
func TestUpdateLead_InvalidResidentialEntriesJSON(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.LeadHandler{}

	existingLead := &models.Lead{
		ID:      "LEAD123",
		JobName: "Test Job",
		Status:  models.NEW,
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Rollback").Return(nil)
	mockDB.On("Preload", mock.Anything).Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "LEAD123").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Lead)
		*dest = *existingLead
	}).Return(nil)
	mockDB.On("First", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("JobName", "Test Job")
	writer.WriteField("Status", string(models.NEW))
	writer.WriteField("ResidentialEntries", "invalid json")
	writer.Close()

	req, _ := http.NewRequest("PUT", "/lead/LEAD123", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "LEAD123"}}
	c.Set("db", mockDB)

	handler.UpdateLead(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "INTERNAL_SERVER_ERROR", response.Code)
	assert.Equal(t, "Unexpected error occurred", response.Error)
}

// 13. Test for database error when fetching existing residential entries
func TestUpdateLead_DatabaseErrorOnFetchResidentialEntries(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.LeadHandler{}

	existingLead := &models.Lead{
		ID:      "LEAD123",
		JobName: "Test Job",
		Status:  models.NEW,
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Rollback").Return(nil)
	mockDB.On("Preload", mock.Anything).Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "LEAD123").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Lead)
		*dest = *existingLead
	}).Return(nil)
	mockDB.On("First", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockDB.On("Where", "lead_id = ?", "LEAD123").Return(mockDB)
	mockDB.On("Find", mock.Anything).Return(errors.New("database error"))

	residentialEntries := []models.ResidentialEntry{{ID: 1}}
	entriesJSON, _ := json.Marshal(residentialEntries)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("JobName", "Test Job")
	writer.WriteField("Status", string(models.NEW))
	writer.WriteField("ResidentialEntries", string(entriesJSON))
	writer.Close()

	req, _ := http.NewRequest("PUT", "/lead/LEAD123", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "LEAD123"}}
	c.Set("db", mockDB)

	handler.UpdateLead(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "INTERNAL_SERVER_ERROR", response.Code)
}

// 14. Test for database error when deleting stages for removed residential entry
func TestUpdateLead_DatabaseErrorOnDeleteStages(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.LeadHandler{}

	existingLead := &models.Lead{
		ID:      "LEAD123",
		JobName: "Test Job",
		Status:  models.NEW,
	}

	existingEntry := models.ResidentialEntry{ID: 1}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Rollback").Return(nil)
	mockDB.On("Preload", mock.Anything).Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "LEAD123").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Lead)
		*dest = *existingLead
	}).Return(nil)
	mockDB.On("First", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockDB.On("Where", "lead_id = ?", "LEAD123").Return(mockDB)
	mockDB.On("Find", mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*[]models.ResidentialEntry)
		*dest = []models.ResidentialEntry{existingEntry}
	}).Return(nil)
	mockDB.On("Where", "residential_entry_id = ?", uint(1)).Return(mockDB)
	mockDB.On("Delete", mock.Anything).Return(errors.New("database error"))

	residentialEntries := []models.ResidentialEntry{} // Empty to trigger deletion
	entriesJSON, _ := json.Marshal(residentialEntries)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("JobName", "Test Job")
	writer.WriteField("Status", string(models.NEW))
	writer.WriteField("ResidentialEntries", string(entriesJSON))
	writer.Close()

	req, _ := http.NewRequest("PUT", "/lead/LEAD123", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "LEAD123"}}
	c.Set("db", mockDB)

	handler.UpdateLead(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "INTERNAL_SERVER_ERROR", response.Code)
}

// 15. Test for database error when deleting residential entry
func TestUpdateLead_DatabaseErrorOnDeleteResidentialEntry(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.LeadHandler{}

	existingLead := &models.Lead{
		ID:      "LEAD123",
		JobName: "Test Job",
		Status:  models.NEW,
	}

	existingEntry := models.ResidentialEntry{ID: 1}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Rollback").Return(nil)
	mockDB.On("Preload", mock.Anything).Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "LEAD123").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Lead)
		*dest = *existingLead
	}).Return(nil)
	mockDB.On("First", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockDB.On("Where", "lead_id = ?", "LEAD123").Return(mockDB)
	mockDB.On("Find", mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*[]models.ResidentialEntry)
		*dest = []models.ResidentialEntry{existingEntry}
	}).Return(nil)
	mockDB.On("Where", "residential_entry_id = ?", uint(1)).Return(mockDB)
	mockDB.On("Delete", mock.MatchedBy(func(v interface{}) bool {
		_, ok := v.(*models.Stage)
		return ok
	})).Return(nil)
	mockDB.On("Delete", mock.MatchedBy(func(v interface{}) bool {
		_, ok := v.(*models.ResidentialEntry)
		return ok
	})).Return(errors.New("database error"))

	residentialEntries := []models.ResidentialEntry{}
	entriesJSON, _ := json.Marshal(residentialEntries)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("JobName", "Test Job")
	writer.WriteField("Status", string(models.NEW))
	writer.WriteField("ResidentialEntries", string(entriesJSON))
	writer.Close()

	req, _ := http.NewRequest("PUT", "/lead/LEAD123", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "LEAD123"}}
	c.Set("db", mockDB)

	handler.UpdateLead(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "INTERNAL_SERVER_ERROR", response.Code)
	assert.Equal(t, "Unexpected error occurred", response.Error)
}
