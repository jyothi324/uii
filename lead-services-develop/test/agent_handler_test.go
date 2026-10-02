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
)

// Import ErrorResponse from handlers package
type ErrorResponse = handlers.ErrorResponse

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		email    string
		expected bool
	}{
		{"test@example.com", true},
		{"user.name@domain.co.uk", true},
		{"invalid.email@", false},
		{"@invalid.com", false},
		{"nodomain@", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			result := handlers.IsValidEmail(tt.email)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCheckEmailExists_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	agent := &models.Agent{
		ID:    "1",
		Email: "test@example.com",
		Name:  "Test Agent",
	}

	mockDB.On("Where", mock.Anything, mock.Anything).Return(mockDB)
	mockDB.On("First", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Agent)
		*dest = *agent
	}).Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	// Create context with query parameter
	req, _ := http.NewRequest("GET", "/checkemailexists?email=test@example.com", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{}
	c.Set("db", mockDB)

	handler.CheckEmailExists(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, true, response["exists"])
}

func TestCheckEmailExists_NotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("Where", mock.Anything, mock.Anything).Return(mockDB)
	mockDB.On("First", mock.Anything, mock.Anything).Return(errors.New("record not found"))

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/checkemailexists?email=notfound@example.com", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CheckEmailExists(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "User not found", response["message"])
}

func TestCheckEmailExists_MissingEmail(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/checkemailexists", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CheckEmailExists(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Email is required", response["error"])
}

func TestGetAgent_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	agent := &models.Agent{
		ID:       "123",
		Name:     "Test Agent",
		Email:    "test@example.com",
		Password: "secret",
	}

	mockDB.On("Where", mock.Anything, mock.Anything).Return(mockDB)
	mockDB.On("First", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Agent)
		*dest = *agent
	}).Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/agent/123", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "123"}}
	c.Set("db", mockDB)

	handler.GetAgent(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var response models.Agent
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "123", response.ID)
	assert.Equal(t, "Test Agent", response.Name)
	assert.Empty(t, response.Password)
}

func TestUpdateAgent_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	existingAgent := &models.Agent{
		ID:    "123",
		Name:  "Old Name",
		Email: "old@example.com",
		Role:  "AGENT",
	}

	updatedAgent := models.Agent{
		ID:    "123",
		Name:  "New Name",
		Email: "new@example.com",
		Role:  "MANAGER",
		Details: models.AgentDetails{
			Branch:      "Main",
			Phone:       "1234567890",
			TOTPSecret:  "secret",
			TOTPEnabled: true,
		},
	}

	// Mock 1: Check if agent exists by ID (returns the agent)
	mockDB.On("Where", "id = ?", "123").Return(mockDB).Once()
	mockDB.On("First", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Agent)
		*dest = *existingAgent
	}).Return(nil).Once()

	// Mock 2: Check if email exists for another agent (should return "not found" - email is available)
	mockDB.On("Where", "email = ? AND id != ?", "new@example.com", "123").Return(mockDB).Once()
	mockDB.On("First", mock.Anything, mock.Anything).Return(errors.New("record not found")).Once()

	// Mock 3: Save the updated agent
	mockDB.On("Save", mock.Anything).Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	updateData, _ := json.Marshal(updatedAgent)
	req, _ := http.NewRequest("PUT", "/agent/123", bytes.NewBuffer(updateData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "123"}}
	c.Set("db", mockDB)

	handler.UpdateAgent(c)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify all expectations were met
	mockDB.AssertExpectations(t)
}

func TestUpdateAgent_NotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("Where", mock.Anything, mock.Anything).Return(mockDB)
	mockDB.On("First", mock.Anything, mock.Anything).Return(errors.New("record not found"))

	handler := &handlers.Handler{DB: mockDB}

	updateData := models.Agent{Name: "New Name"}
	data, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", "/agent/999", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "999"}}
	c.Set("db", mockDB)

	handler.UpdateAgent(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCreateAgent_ValidationError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	invalidRequest := map[string]interface{}{
		"name": "Test",
	}

	data, _ := json.Marshal(invalidRequest)
	req, _ := http.NewRequest("POST", "/agent", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateAgent(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "VALIDATION_ERROR", response.Code)
}

func TestCheckMFA_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	agent := &models.Agent{
		ID:    "1",
		Email: "test@example.com",
		Details: models.AgentDetails{
			TOTPEnabled: true,
		},
	}

	mockDB.On("Where", mock.Anything, mock.Anything).Return(mockDB)
	mockDB.On("First", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Agent)
		*dest = *agent
	}).Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/check-mfa", nil)
	req.Header.Set("user-email", "test@example.com")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CheckMFA(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, true, response["totp_enabled"])
}

func TestCheckMFA_MissingEmail(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/check-mfa", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CheckMFA(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Email is required in the header", response["error"])
}
