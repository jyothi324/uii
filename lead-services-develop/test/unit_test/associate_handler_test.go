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

// type ErrorResponse = handlers.ErrorResponse

func TestCreateAssociateCustomer_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	// Assign mock DB to handler
	handler := &handlers.Handler{
		DB: mockDB,
	}

	input := models.AssociateCustomer{
		Name: "John Architect",
		Role: models.ARCHITECT,
		Details: models.AssociateDetails{
			Email: "john@example.com",
			Phone: "1234567890",
		},
	}

	// Mock GORM calls for ID generation
	mockDB.On("Model", mock.MatchedBy(func(v interface{}) bool {
		_, ok := v.(*models.Agent)
		return ok
	})).Return(mockDB)
	mockDB.On("Where", mock.Anything, mock.Anything).Return(mockDB)
	mockDB.On("Count", mock.Anything).Run(func(args mock.Arguments) {
		countPtr := args.Get(0).(*int64)
		*countPtr = 0
	}).Return(mockDB)

	// Mock for create
	mockDB.On("Create", mock.MatchedBy(func(v interface{}) bool {
		_, ok := v.(*models.AssociateCustomer)
		return ok
	})).Return(nil)

	// Prepare HTTP request
	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/associate", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	c.Set("db", mockDB)
	// Set tenant if your handler uses it
	c.Set("tenant", "default")

	// Call the handler
	handler.CreateAssociateCustomer(c)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Associate customer created successfully", response["message"])
}
func TestCreateAssociateCustomer_BadRequest_MissingFields(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	// Mock ID generation
	mockDB.On("Model", mock.MatchedBy(func(v interface{}) bool {
		_, ok := v.(*models.Agent)
		return ok
	})).Return(mockDB)
	mockDB.On("Where", mock.Anything, mock.Anything).Return(mockDB)
	mockDB.On("Count", mock.Anything).Run(func(args mock.Arguments) {
		countPtr := args.Get(0).(*int64)
		*countPtr = 0
	}).Return(mockDB)

	input := map[string]interface{}{
		"name": "John",
	}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/associate", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateAssociateCustomer(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "VALIDATION_ERROR", response.Code)
}

func TestCreateAssociateCustomer_BadRequest_InvalidRole(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	// Mock ID generation
	mockDB.On("Model", mock.MatchedBy(func(v interface{}) bool {
		_, ok := v.(*models.Agent)
		return ok
	})).Return(mockDB)
	mockDB.On("Where", mock.Anything, mock.Anything).Return(mockDB)
	mockDB.On("Count", mock.Anything).Run(func(args mock.Arguments) {
		countPtr := args.Get(0).(*int64)
		*countPtr = 0
	}).Return(mockDB)

	input := models.AssociateCustomer{
		Name: "John",
		Role: "INVALID_ROLE",
		Details: models.AssociateDetails{
			Email: "john@example.com",
			Phone: "1234567890",
		},
	}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/associate", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateAssociateCustomer(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "VALIDATION_ERROR", response.Code)
	assert.Contains(t, response.Details, "ARCHITECT, DESIGNER, or CONTRACTOR")
}

func TestCreateAssociateCustomer_InternalServerError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	// Mock ID generation
	mockDB.On("Model", mock.MatchedBy(func(v interface{}) bool {
		_, ok := v.(*models.Agent)
		return ok
	})).Return(mockDB)
	mockDB.On("Where", mock.Anything, mock.Anything).Return(mockDB)
	mockDB.On("Count", mock.Anything).Run(func(args mock.Arguments) {
		countPtr := args.Get(0).(*int64)
		*countPtr = 0
	}).Return(mockDB)

	input := models.AssociateCustomer{
		Name: "John Architect",
		Role: models.ARCHITECT,
		Details: models.AssociateDetails{
			Email: "john@example.com",
			Phone: "1234567890",
		},
	}

	mockDB.On("Create", mock.MatchedBy(func(v interface{}) bool {
		return true
	})).Return(errors.New("database error"))

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/associate", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateAssociateCustomer(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "DATABASE_ERROR", response.Code)
}

// func TestCreateAssociateCustomer_InvalidJSON(t *testing.T) {
// 	mockDB := new(mocks.MockDatabase)
// 	handler := &handlers.Handler{DB: mockDB}

// 	req, _ := http.NewRequest(
// 		"POST",
// 		"/associate",
// 		bytes.NewBufferString("{invalid-json"),
// 	)
// 	req.Header.Set("Content-Type", "application/json")

// 	w := httptest.NewRecorder()
// 	c, _ := gin.CreateTestContext(w)
// 	c.Request = req

// 	handler.CreateAssociateCustomer(c)

// 	assert.Equal(t, http.StatusBadRequest, w.Code)

// 	var resp ErrorResponse
// 	_ = json.Unmarshal(w.Body.Bytes(), &resp)
// 	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
// }

// GET Associate Customer
func TestGetAssociateCustomer_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{}

	associate := &models.AssociateCustomer{
		ID:   "ASSOC123",
		Name: "John Architect",
		Role: models.ARCHITECT,
		Details: models.AssociateDetails{
			Email:     "john@example.com",
			Phone:     "1234567890",
			CompanyID: "COMP123",
		},
	}

	company := &models.Customer{
		ID:   "COMP123",
		Name: "ABC Company",
	}

	mockDB.On("Where", "id = ?", "ASSOC123").Return(mockDB)
	mockDB.On("First", mock.MatchedBy(func(v interface{}) bool {
		return true
	})).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.AssociateCustomer)
		*dest = *associate
	}).Return(nil)

	mockDB.On("First", mock.MatchedBy(func(v interface{}) bool {
		return true
	}), "id = ?", "COMP123").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Customer)
		*dest = *company
	}).Return(nil)

	req, _ := http.NewRequest("GET", "/associate/ASSOC123", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "ASSOC123"}}
	c.Set("db", mockDB)

	handler.GetAssociateCustomer(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotNil(t, response["data"])
}

func TestGetAssociateCustomer_BadRequest_NoID(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{}

	req, _ := http.NewRequest("GET", "/associate/", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{}
	c.Set("db", mockDB)

	handler.GetAssociateCustomer(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "VALIDATION_ERROR", response.Code)
}

func TestGetAssociateCustomer_NotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{}

	mockDB.On("Where", "id = ?", "NOTFOUND").Return(mockDB)
	mockDB.On("First", mock.MatchedBy(func(v interface{}) bool {
		return true
	}), mock.Anything, mock.Anything).Return(gorm.ErrRecordNotFound)

	req, _ := http.NewRequest("GET", "/associate/NOTFOUND", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "NOTFOUND"}}
	c.Set("db", mockDB)

	handler.GetAssociateCustomer(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var response ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "NOT_FOUND", response.Code)
}

// ------------------update_Associte--------------------
func TestUpdateAssociateCustomer_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{}

	existingAssociate := &models.AssociateCustomer{
		ID:   "ASSOC123",
		Name: "Old Name",
		Role: models.ARCHITECT,
		Details: models.AssociateDetails{
			Email: "old@example.com",
			Phone: "1234567890",
		},
	}

	updatedAssociate := models.AssociateCustomer{
		ID:   "ASSOC123",
		Name: "New Name",
		Role: models.DESIGNER,
		Details: models.AssociateDetails{
			Email: "new@example.com",
			Phone: "9876543210",
		},
	}

	mockDB.On("Where", "id = ?", "ASSOC123").Return(mockDB)
	mockDB.On("First", mock.MatchedBy(func(v interface{}) bool {
		return true
	}), mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.AssociateCustomer)
		*dest = *existingAssociate
	}).Return(nil)

	mockDB.On("Where", "id != ?", "ASSOC123").Return(mockDB)
	mockDB.On("Find", mock.MatchedBy(func(v interface{}) bool {
		return true
	}), mock.Anything).Return(nil)

	mockDB.On("Save", mock.MatchedBy(func(v interface{}) bool {
		return true
	})).Return(nil)

	data, _ := json.Marshal(updatedAssociate)
	req, _ := http.NewRequest("PUT", "/associate/ASSOC123", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "ASSOC123"}}
	c.Set("db", mockDB)

	handler.UpdateAssociateCustomer(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Associate customer updated successfully", response["message"])
}

func TestUpdateAssociateCustomer_NotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{}

	mockDB.On("Where", "id = ?", "NOTFOUND").Return(mockDB)
	mockDB.On("First", mock.MatchedBy(func(v interface{}) bool {
		return true
	}), mock.Anything, mock.Anything).Return(gorm.ErrRecordNotFound)

	updateData := models.AssociateCustomer{
		Name: "New Name",
		Role: models.DESIGNER,
		Details: models.AssociateDetails{
			Email: "new@example.com",
			Phone: "9876543210",
		},
	}

	data, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", "/associate/NOTFOUND", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "NOTFOUND"}}
	c.Set("db", mockDB)

	handler.UpdateAssociateCustomer(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var response ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "NOT_FOUND", response.Code)
}

// 1. Test for empty/missing ID parameter
func TestUpdateAssociateCustomer_EmptyID(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{}

	req, _ := http.NewRequest("PUT", "/associate/", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: ""}}
	c.Set("db", mockDB)

	handler.UpdateAssociateCustomer(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "VALIDATION_ERROR", response.Code)
	assert.Equal(t, "Invalid ID", response.Error)
}

// 2. Test for database error when fetching associate
func TestUpdateAssociateCustomer_DatabaseErrorOnFetch(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{}

	mockDB.On("First", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("database connection failed"))

	updateData := models.AssociateCustomer{
		Name: "Test",
		Role: models.DESIGNER,
		Details: models.AssociateDetails{
			Email: "test@example.com",
			Phone: "1234567890",
		},
	}

	data, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", "/associate/ASSOC123", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "ASSOC123"}}
	c.Set("db", mockDB)

	handler.UpdateAssociateCustomer(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "DATABASE_ERROR", response.Code)
}

// 3. Test for invalid JSON input
func TestUpdateAssociateCustomer_InvalidJSON(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{}

	existingAssociate := &models.AssociateCustomer{
		ID:   "ASSOC123",
		Name: "Old Name",
		Role: models.ARCHITECT,
	}

	mockDB.On("First", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.AssociateCustomer)
		*dest = *existingAssociate
	}).Return(nil)

	req, _ := http.NewRequest("PUT", "/associate/ASSOC123", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "ASSOC123"}}
	c.Set("db", mockDB)

	handler.UpdateAssociateCustomer(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "VALIDATION_ERROR", response.Code)
}

// 4. Test for missing required fields
func TestUpdateAssociateCustomer_MissingRequiredFields(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{}

	existingAssociate := &models.AssociateCustomer{
		ID:   "ASSOC123",
		Name: "Old Name",
		Role: models.ARCHITECT,
	}

	mockDB.On("First", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.AssociateCustomer)
		*dest = *existingAssociate
	}).Return(nil)

	// Missing email and phone
	updateData := models.AssociateCustomer{
		Name: "New Name",
		Role: models.DESIGNER,
		Details: models.AssociateDetails{
			Email: "",
			Phone: "",
		},
	}

	data, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", "/associate/ASSOC123", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "ASSOC123"}}
	c.Set("db", mockDB)

	handler.UpdateAssociateCustomer(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "VALIDATION_ERROR", response.Code)
	assert.Equal(t, "Missing required fields", response.Error)
}

// 5. Test for invalid role
func TestUpdateAssociateCustomer_InvalidRole(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{}

	existingAssociate := &models.AssociateCustomer{
		ID:   "ASSOC123",
		Name: "Old Name",
		Role: models.ARCHITECT,
	}

	mockDB.On("First", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.AssociateCustomer)
		*dest = *existingAssociate
	}).Return(nil)

	updateData := models.AssociateCustomer{
		Name: "New Name",
		Role: "INVALID_ROLE",
		Details: models.AssociateDetails{
			Email: "test@example.com",
			Phone: "1234567890",
		},
	}

	data, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", "/associate/ASSOC123", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "ASSOC123"}}
	c.Set("db", mockDB)

	handler.UpdateAssociateCustomer(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "VALIDATION_ERROR", response.Code)
	assert.Equal(t, "Invalid Role", response.Error)
}

// 6. Test for invalid CompanyID - not found
func TestUpdateAssociateCustomer_InvalidCompanyID(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{}

	existingAssociate := &models.AssociateCustomer{
		ID:   "ASSOC123",
		Name: "Old Name",
		Role: models.ARCHITECT,
	}

	mockDB.On("First", mock.Anything, "id = ?", "ASSOC123").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.AssociateCustomer)
		*dest = *existingAssociate
	}).Return(nil)

	mockDB.On("First", mock.Anything, "id = ?", "INVALID_COMPANY").Return(gorm.ErrRecordNotFound)

	updateData := models.AssociateCustomer{
		Name: "New Name",
		Role: models.DESIGNER,
		Details: models.AssociateDetails{
			Email:     "test@example.com",
			Phone:     "1234567890",
			CompanyID: "INVALID_COMPANY",
		},
	}

	data, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", "/associate/ASSOC123", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "ASSOC123"}}
	c.Set("db", mockDB)

	handler.UpdateAssociateCustomer(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "VALIDATION_ERROR", response.Code)
	assert.Equal(t, "Invalid CompanyID", response.Error)
}

// 7. Test for database error when checking CompanyID
func TestUpdateAssociateCustomer_DatabaseErrorOnCompanyCheck(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{}

	existingAssociate := &models.AssociateCustomer{
		ID:   "ASSOC123",
		Name: "Old Name",
		Role: models.ARCHITECT,
	}

	mockDB.On("First", mock.Anything, "id = ?", "ASSOC123").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.AssociateCustomer)
		*dest = *existingAssociate
	}).Return(nil)

	mockDB.On("First", mock.Anything, "id = ?", "COMPANY123").Return(errors.New("database error"))

	updateData := models.AssociateCustomer{
		Name: "New Name",
		Role: models.DESIGNER,
		Details: models.AssociateDetails{
			Email:     "test@example.com",
			Phone:     "1234567890",
			CompanyID: "COMPANY123",
		},
	}

	data, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", "/associate/ASSOC123", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "ASSOC123"}}
	c.Set("db", mockDB)

	handler.UpdateAssociateCustomer(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "DATABASE_ERROR", response.Code)
}

// 8. Test for duplicate email
func TestUpdateAssociateCustomer_DuplicateEmail(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{}

	existingAssociate := &models.AssociateCustomer{
		ID:   "ASSOC123",
		Name: "Old Name",
		Role: models.ARCHITECT,
		Details: models.AssociateDetails{
			Email: "old@example.com",
			Phone: "1234567890",
		},
	}

	otherAssociate := models.AssociateCustomer{
		ID:   "ASSOC456",
		Name: "Other",
		Role: models.DESIGNER,
		Details: models.AssociateDetails{
			Email: "duplicate@example.com",
			Phone: "9999999999",
		},
	}

	mockDB.On("First", mock.Anything, "id = ?", "ASSOC123").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.AssociateCustomer)
		*dest = *existingAssociate
	}).Return(nil)

	mockDB.On("Where", "id != ?", "ASSOC123").Return(mockDB)
	mockDB.On("Find", mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*[]models.AssociateCustomer)
		*dest = []models.AssociateCustomer{otherAssociate}
	}).Return(nil)

	updateData := models.AssociateCustomer{
		Name: "New Name",
		Role: models.DESIGNER,
		Details: models.AssociateDetails{
			Email: "duplicate@example.com",
			Phone: "1111111111",
		},
	}

	data, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", "/associate/ASSOC123", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "ASSOC123"}}
	c.Set("db", mockDB)

	handler.UpdateAssociateCustomer(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "VALIDATION_ERROR", response.Code)
	assert.Equal(t, "Email already exists", response.Error)
}

// 9. Test for database error when checking email uniqueness
func TestUpdateAssociateCustomer_DatabaseErrorOnEmailCheck(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{}

	existingAssociate := &models.AssociateCustomer{
		ID:   "ASSOC123",
		Name: "Old Name",
		Role: models.ARCHITECT,
	}

	mockDB.On("First", mock.Anything, "id = ?", "ASSOC123").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.AssociateCustomer)
		*dest = *existingAssociate
	}).Return(nil)

	mockDB.On("Where", "id != ?", "ASSOC123").Return(mockDB)
	mockDB.On("Find", mock.Anything).Return(errors.New("database error"))

	updateData := models.AssociateCustomer{
		Name: "New Name",
		Role: models.DESIGNER,
		Details: models.AssociateDetails{
			Email: "test@example.com",
			Phone: "1234567890",
		},
	}

	data, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", "/associate/ASSOC123", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "ASSOC123"}}
	c.Set("db", mockDB)

	handler.UpdateAssociateCustomer(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "DATABASE_ERROR", response.Code)
}

// 10. Test for duplicate phone
func TestUpdateAssociateCustomer_DuplicatePhone(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{}

	existingAssociate := &models.AssociateCustomer{
		ID:   "ASSOC123",
		Name: "Old Name",
		Role: models.ARCHITECT,
		Details: models.AssociateDetails{
			Email: "old@example.com",
			Phone: "1234567890",
		},
	}

	otherAssociate := models.AssociateCustomer{
		ID:   "ASSOC456",
		Name: "Other",
		Role: models.DESIGNER,
		Details: models.AssociateDetails{
			Email: "other@example.com",
			Phone: "9999999999",
		},
	}

	mockDB.On("First", mock.Anything, "id = ?", "ASSOC123").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.AssociateCustomer)
		*dest = *existingAssociate
	}).Return(nil)

	mockDB.On("Where", "id != ?", "ASSOC123").Return(mockDB)
	mockDB.On("Find", mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*[]models.AssociateCustomer)
		*dest = []models.AssociateCustomer{otherAssociate}
	}).Return(nil).Times(2)

	updateData := models.AssociateCustomer{
		Name: "New Name",
		Role: models.DESIGNER,
		Details: models.AssociateDetails{
			Email: "unique@example.com",
			Phone: "9999999999",
		},
	}

	data, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", "/associate/ASSOC123", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "ASSOC123"}}
	c.Set("db", mockDB)

	handler.UpdateAssociateCustomer(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "VALIDATION_ERROR", response.Code)
	assert.Equal(t, "Phone already exists", response.Error)
}

// 11. Test for database error when checking phone uniqueness
func TestUpdateAssociateCustomer_DatabaseErrorOnPhoneCheck(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{}

	existingAssociate := &models.AssociateCustomer{
		ID:   "ASSOC123",
		Name: "Old Name",
		Role: models.ARCHITECT,
	}

	mockDB.On("First", mock.Anything, "id = ?", "ASSOC123").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.AssociateCustomer)
		*dest = *existingAssociate
	}).Return(nil)

	mockDB.On("Where", "id != ?", "ASSOC123").Return(mockDB)
	mockDB.On("Find", mock.Anything).Return(nil).Once()
	mockDB.On("Find", mock.Anything).Return(errors.New("database error")).Once()

	updateData := models.AssociateCustomer{
		Name: "New Name",
		Role: models.DESIGNER,
		Details: models.AssociateDetails{
			Email: "test@example.com",
			Phone: "1234567890",
		},
	}

	data, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", "/associate/ASSOC123", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "ASSOC123"}}
	c.Set("db", mockDB)

	handler.UpdateAssociateCustomer(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "DATABASE_ERROR", response.Code)
}

// 12. Test for database error when saving
func TestUpdateAssociateCustomer_DatabaseErrorOnSave(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{}

	existingAssociate := &models.AssociateCustomer{
		ID:   "ASSOC123",
		Name: "Old Name",
		Role: models.ARCHITECT,
	}

	mockDB.On("First", mock.Anything, "id = ?", "ASSOC123").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.AssociateCustomer)
		*dest = *existingAssociate
	}).Return(nil)

	mockDB.On("Where", "id != ?", "ASSOC123").Return(mockDB)
	mockDB.On("Find", mock.Anything).Return(nil)
	mockDB.On("Save", mock.Anything).Return(errors.New("database save error"))

	updateData := models.AssociateCustomer{
		Name: "New Name",
		Role: models.DESIGNER,
		Details: models.AssociateDetails{
			Email: "test@example.com",
			Phone: "1234567890",
		},
	}

	data, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", "/associate/ASSOC123", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "ASSOC123"}}
	c.Set("db", mockDB)

	handler.UpdateAssociateCustomer(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "DATABASE_ERROR", response.Code)
	assert.Equal(t, "Failed to update associate customer", response.Error)
}

func TestDeleteAssociateCustomer_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{}

	mockDB.On("Delete", mock.Anything, "id = ?", "ASSOC123").Return(nil)

	req, _ := http.NewRequest("DELETE", "/associate/ASSOC123", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "ASSOC123"}}
	c.Set("db", mockDB)

	handler.DeleteAssociateCustomer(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Associate customer deleted successfully", response["message"])
}

func TestDeleteAssociateCustomer_BadRequest_NoID(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{}

	req, _ := http.NewRequest("DELETE", "/associate/", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{}
	c.Set("db", mockDB)

	handler.DeleteAssociateCustomer(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "VALIDATION_ERROR", response.Code)
}

func TestGetAllAssociateCustomers_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{}

	associates := []models.AssociateCustomer{
		{
			ID:   "ASSOC1",
			Name: "John Architect",
			Role: models.ARCHITECT,
		},
		{
			ID:   "ASSOC2",
			Name: "Jane Designer",
			Role: models.DESIGNER,
		},
	}

	mockDB.On("Find", mock.MatchedBy(func(v interface{}) bool {
		return true
	}), mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*[]models.AssociateCustomer)
		*dest = associates
	}).Return(nil)

	req, _ := http.NewRequest("GET", "/associates", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.GetAllAssociateCustomers(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotNil(t, response["data"])
}

func TestCreateAssociateCustomer_WithCompanyNotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	// Mock ID generation
	mockDB.On("Model", mock.MatchedBy(func(v interface{}) bool {
		_, ok := v.(*models.Agent)
		return ok
	})).Return(mockDB)
	mockDB.On("Where", mock.Anything, mock.Anything).Return(mockDB)
	mockDB.On("Count", mock.Anything).Run(func(args mock.Arguments) {
		countPtr := args.Get(0).(*int64)
		*countPtr = 0
	}).Return(mockDB)

	input := models.AssociateCustomer{
		Name: "John Architect",
		Role: models.ARCHITECT,
		Details: models.AssociateDetails{
			Email:       "john@example.com",
			Phone:       "1234567890",
			CompanyID:   "COMP123",
			CompanyName: "ABC Company",
		},
	}

	mockDB.On("First", mock.MatchedBy(func(v interface{}) bool {
		_, ok := v.(*models.Customer)
		return ok
	}), "id = ?", "COMP123").Return(gorm.ErrRecordNotFound)

	mockDB.On("Create", mock.MatchedBy(func(v interface{}) bool {
		return true
	})).Return(nil)

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/associate", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateAssociateCustomer(c)

	assert.Equal(t, http.StatusCreated, w.Code)
}
