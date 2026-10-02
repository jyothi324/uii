package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
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
func TestCreateAgent_InvalidJSON(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{}

	req, _ := http.NewRequest(
		"POST",
		"/agent",
		bytes.NewBuffer([]byte("{invalid-json")),
	)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// 🔑 THIS IS NON-OPTIONAL
	c.Set("db", mockDB)

	handler.CreateAgent(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp ErrorResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
}

func TestCreateAgent_MissingRequiredFields(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{}

	reqBody := map[string]interface{}{
		"name": "Test", // missing email + role
	}
	data, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/agent", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// 🔑 REQUIRED or handler will panic
	c.Set("db", mockDB)

	handler.CreateAgent(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp ErrorResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
}

func TestCreateAgent_TOTPEnabled_GenerateSecret(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{}

	// Email uniqueness check
	mockDB.
		On("Where", mock.Anything, mock.Anything).
		Return(mockDB)
	mockDB.
		On("First", mock.Anything).
		Return(errors.New("not found"))

	// 🔑 Required for GenerateEmployeeID
	mockDB.
		On("Model", mock.Anything).
		Return(mockDB)
	mockDB.
		On("Count", mock.Anything).
		Return(nil)

	// Save agent
	mockDB.
		On("Create", mock.Anything).
		Return(nil)

	reqBody := map[string]interface{}{
		"name":  "Test",
		"email": "test@test.com",
		"role":  "agent",
		"details": map[string]interface{}{
			"totp_enabled": true,
		},
	}
	data, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/agent", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateAgent(c)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCreateAgent_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{}

	mockDB.
		On("Where", mock.Anything, mock.Anything).
		Return(mockDB)
	mockDB.
		On("First", mock.Anything).
		Return(errors.New("not found"))

	mockDB.
		On("Model", mock.Anything).
		Return(mockDB)
	mockDB.
		On("Count", mock.Anything).
		Return(nil)

	mockDB.
		On("Create", mock.Anything).
		Return(nil)

	reqBody := map[string]interface{}{
		"name":     "Test",
		"email":    "test@test.com",
		"role":     "agent",
		"password": "password123",
	}
	data, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/agent", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	c.Set("db", mockDB)

	handler.CreateAgent(c)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCreateAgent_DuplicateEmail(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{}

	mockDB.
		On("Where", "email = ?", "test@test.com").
		Return(mockDB)
	mockDB.
		On("First", mock.Anything).
		Return(nil)

	reqBody := map[string]interface{}{
		"name":  "Test",
		"email": "test@test.com",
		"role":  "agent",
	}
	data, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/agent", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	c.Set("db", mockDB)

	handler.CreateAgent(c)

	assert.Equal(t, http.StatusConflict, w.Code)
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

func TestListAgents_AllTrue(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	// Mock chaining methods
	mockDB.On("Model", mock.Anything).Return(mockDB)
	mockDB.On("Where", mock.Anything, mock.Anything).Return(mockDB)
	mockDB.On("Where", mock.Anything).Return(mockDB)
	mockDB.On("Count", mock.Anything).Return(mockDB)
	mockDB.On("Select", mock.Anything).Return(mockDB)
	mockDB.On("Offset", mock.Anything).Return(mockDB)
	mockDB.On("Limit", mock.Anything).Return(mockDB)
	mockDB.On("Find", mock.Anything).Return(nil)
	mockDB.On("Error").Return(nil)

	// Create test context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/agents?all=true", nil)
	c.Set("db", mockDB)

	// Call handler
	handler := &handlers.Handler{}
	handler.ListAgents(c)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Contains(t, resp, "data")
	assert.Contains(t, resp["data"].(map[string]interface{}), "data")
	assert.Contains(t, resp["data"].(map[string]interface{}), "total")
}

func TestListAgents_Pagination(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("Model", mock.Anything).Return(mockDB)
	mockDB.On("Where", mock.Anything, mock.Anything).Return(mockDB)
	mockDB.On("Count", mock.Anything).Return(mockDB)
	mockDB.On("Select", mock.Anything).Return(mockDB)
	mockDB.On("Offset", mock.Anything).Return(mockDB)
	mockDB.On("Limit", mock.Anything).Return(mockDB)
	mockDB.On("Find", mock.Anything).Return(nil)
	mockDB.On("Error").Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/agents?page=1&page_size=10", nil)
	c.Set("db", mockDB)

	handler := &handlers.Handler{}
	handler.ListAgents(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

// func TestListAgents_WithFilters(t *testing.T) {
// 	mockDB := new(mocks.MockDatabase)

// 	mockDB.On("Model", mock.Anything).Return(mockDB)
// 	mockDB.On("Where", mock.Anything, mock.Anything).Return(mockDB)
// 	mockDB.On("Count", mock.Anything).Return(mockDB)
// 	mockDB.On("Select", mock.Anything).Return(mockDB)
// 	mockDB.On("Offset", mock.Anything).Return(mockDB)
// 	mockDB.On("Limit", mock.Anything).Return(mockDB)
// 	mockDB.On("Find", mock.Anything).Return(nil)
// 	mockDB.On("Error").Return(nil)

// 	req := httptest.NewRequest(http.MethodGet, "/agents?search=john&role=admin&assignee_type=user&phone=123", nil)
// 	w := httptest.NewRecorder()
// 	c, _ := gin.CreateTestContext(w)
// 	c.Request = req
// 	c.Set("db", mockDB)

// 	handler := &handlers.Handler{}
// 	handler.ListAgents(c)

// 	assert.Equal(t, http.StatusOK, w.Code)
// }

func TestListAgents_CountError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("Model", mock.Anything).Return(mockDB)
	mockDB.On("Where", mock.Anything, mock.Anything).Return(mockDB)
	mockDB.On("Count", mock.Anything).Return(mockDB)
	mockDB.On("Error").Return(errors.New("count error"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/agents", nil)
	c.Set("db", mockDB)

	handler := &handlers.Handler{}
	handler.ListAgents(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestListAgents_FindError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("Model", mock.Anything).Return(mockDB)
	mockDB.On("Where", mock.Anything, mock.Anything).Return(mockDB)
	mockDB.On("Count", mock.Anything).Return(mockDB)
	mockDB.On("Find", mock.Anything).Return(errors.New("find error"))
	mockDB.On("Error").Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/agents?all=true", nil)
	c.Set("db", mockDB)

	handler := &handlers.Handler{}
	handler.ListAgents(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestDeleteAgent(t *testing.T) {
	agentID := "123"

	t.Run("Invalid ID", func(t *testing.T) {
		mockDB := new(mocks.MockDatabase) // fresh mock
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: ""}}
		c.Set("db", mockDB)

		handler := &handlers.Handler{}
		handler.DeleteAgent(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("DB connection failed", func(t *testing.T) {
		mockDB := new(mocks.MockDatabase)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: agentID}}
		c.Set("db", mockDB)

		// Configure mock to fail on First
		mockDB.On("First", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("DB connection failed"))

		handler := &handlers.Handler{}
		handler.DeleteAgent(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Agent not found", func(t *testing.T) {
		mockDB := new(mocks.MockDatabase)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: agentID}}
		c.Set("db", mockDB)

		mockDB.On("First", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("record not found"))

		handler := &handlers.Handler{}
		handler.DeleteAgent(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Delete error", func(t *testing.T) {
		mockDB := new(mocks.MockDatabase)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: agentID}}
		c.Set("db", mockDB)

		agent := &models.Agent{ID: agentID}
		mockDB.On("First", mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			arg := args.Get(0).(*models.Agent)
			*arg = *agent
		})
		mockDB.On("Delete", mock.Anything).Return(errors.New("delete failed"))

		handler := &handlers.Handler{}
		handler.DeleteAgent(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Successful delete", func(t *testing.T) {
		mockDB := new(mocks.MockDatabase)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: agentID}}
		c.Set("db", mockDB)

		agent := &models.Agent{ID: agentID}
		mockDB.On("First", mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			arg := args.Get(0).(*models.Agent)
			*arg = *agent
		})
		mockDB.On("Delete", mock.Anything).Return(nil)

		handler := &handlers.Handler{}
		handler.DeleteAgent(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestUpdateSuperuserStatus(t *testing.T) {
	agentID := "123"

	t.Run("Invalid ID", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: ""}}

		handler := &handlers.Handler{}
		handler.UpdateSuperuserStatus(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Agent not found", func(t *testing.T) {
		mockDB := new(mocks.MockDatabase)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: agentID}}
		c.Set("db", mockDB)

		// First returns error simulating "not found"
		mockDB.On("Where", mock.Anything, mock.Anything).Return(mockDB)
		mockDB.On("First", mock.Anything).Return(errors.New("record not found"))

		handler := &handlers.Handler{}
		handler.UpdateSuperuserStatus(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Invalid request body", func(t *testing.T) {
		mockDB := new(mocks.MockDatabase)
		mockDB.On("Where", mock.Anything, mock.Anything).Return(mockDB)

		// Mock First so the panic doesn't happen
		mockDB.On("First", mock.Anything).Return(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: agentID}}
		c.Set("db", mockDB)

		// Send invalid JSON
		c.Request = httptest.NewRequest(http.MethodPatch, "/", strings.NewReader("{invalid json"))

		handler := &handlers.Handler{}
		handler.UpdateSuperuserStatus(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("DB save error", func(t *testing.T) {
		mockDB := new(mocks.MockDatabase)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: agentID}}
		c.Set("db", mockDB)

		agent := &models.Agent{ID: agentID}
		mockDB.On("Where", mock.Anything, mock.Anything).Return(mockDB)
		mockDB.On("First", mock.Anything).Run(func(args mock.Arguments) {
			arg := args.Get(0).(*models.Agent)
			*arg = *agent
		}).Return(nil)
		mockDB.On("Save", mock.Anything).Return(errors.New("update failed"))

		handler := &handlers.Handler{}
		body := strings.NewReader(`{"superuser": true}`)
		c.Request = httptest.NewRequest(http.MethodPatch, "/", body)

		handler.UpdateSuperuserStatus(c)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Successful update", func(t *testing.T) {
		mockDB := new(mocks.MockDatabase)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: agentID}}
		c.Set("db", mockDB)

		agent := &models.Agent{ID: agentID}
		mockDB.On("Where", mock.Anything, mock.Anything).Return(mockDB)
		mockDB.On("First", mock.Anything).Run(func(args mock.Arguments) {
			arg := args.Get(0).(*models.Agent)
			*arg = *agent
		}).Return(nil)
		mockDB.On("Save", mock.Anything).Return(nil)

		handler := &handlers.Handler{}
		body := strings.NewReader(`{"superuser": true}`)
		c.Request = httptest.NewRequest(http.MethodPatch, "/", body)

		handler.UpdateSuperuserStatus(c)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestCheckCompanyExists_MissingCompany(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &handlers.Handler{}

	req, _ := http.NewRequest("GET", "/checkcompanyexists", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.CheckCompanyExists(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	assert.Equal(t, false, resp["success"])
	assert.Equal(t, "Company name required", resp["message"])
}
