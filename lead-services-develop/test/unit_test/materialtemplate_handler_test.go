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

// ----------------- CreateMaterialTemplate -----------------
func TestCreateMaterialTemplate_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	input := models.MaterialTemplate{
		Name:     "Template A",
		Category: "Category A",
		SubCategory: []models.SubCategoryDetails{
			{Cost: "100", Name: "SubCat1", PricingType: "Fixed"},
		},
	}

	mockDB.On("Create", mock.Anything).Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/material-templates", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateMaterialTemplate(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Material template created successfully", resp["message"])

	mockDB.AssertExpectations(t)
}

func TestCreateMaterialTemplate_InvalidInput(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("POST", "/material-templates", bytes.NewBuffer([]byte(`invalid json`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateMaterialTemplate(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
}

func TestCreateMaterialTemplate_MissingRequiredFields(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	input := models.MaterialTemplate{
		Name: "Template A",
		// Missing Category
	}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/material-templates", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateMaterialTemplate(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
	assert.Equal(t, "Missing required fields", resp.Error)
}

func TestCreateMaterialTemplate_WithManagerResidentialEntryID_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mre := models.ManagerResidentialEntry{ID: 1}
	input := models.MaterialTemplate{
		Name:                      "Template A",
		Category:                  "Category A",
		ManagerResidentialEntryID: 1,
	}

	mockDB.On("First", mock.AnythingOfType("*models.ManagerResidentialEntry"), uint(1)).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.ManagerResidentialEntry)
		*dest = mre
	}).Return(nil)
	mockDB.On("Create", mock.Anything).Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/material-templates", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateMaterialTemplate(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockDB.AssertExpectations(t)
}

func TestCreateMaterialTemplate_WithInvalidManagerResidentialEntryID(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	input := models.MaterialTemplate{
		Name:                      "Template A",
		Category:                  "Category A",
		ManagerResidentialEntryID: 999,
	}

	mockDB.On("First", mock.AnythingOfType("*models.ManagerResidentialEntry"), uint(999)).Return(gorm.ErrRecordNotFound)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/material-templates", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateMaterialTemplate(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
	assert.Equal(t, "Invalid ManagerResidentialEntryID", resp.Error)
}

// ----------------- GetMaterialTemplate -----------------
func TestGetMaterialTemplate_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	expected := models.MaterialTemplate{
		ID:       1,
		Name:     "Template A",
		Category: "Category A",
	}

	mockDB.On("Preload", "ManagerResidentialEntry").Return(mockDB)
	mockDB.On("Preload", "ManagerResidentialType").Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "1").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.MaterialTemplate)
		*dest = expected
	}).Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/material-templates/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.GetMaterialTemplate(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]models.MaterialTemplate
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, expected.Name, resp["data"].Name)
	mockDB.AssertExpectations(t)
}

func TestGetMaterialTemplate_NotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("Preload", "ManagerResidentialEntry").Return(mockDB)
	mockDB.On("Preload", "ManagerResidentialType").Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "999").Return(gorm.ErrRecordNotFound)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/material-templates/999", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "999"}}
	c.Set("db", mockDB)

	handler.GetMaterialTemplate(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "NOT_FOUND", resp.Code)
}

func TestGetMaterialTemplate_MissingID(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/material-templates/", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: ""}}
	c.Set("db", mockDB)

	handler.GetMaterialTemplate(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
}

// ----------------- ListMaterialTemplates -----------------
func TestListMaterialTemplates_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	templateList := []models.MaterialTemplate{
		{ID: 1, Name: "Template A", Category: "Category A"},
		{ID: 2, Name: "Template B", Category: "Category B"},
	}

	var total int64 = 2

	mockDB.On("Model", mock.AnythingOfType("*models.MaterialTemplate")).Return(mockDB)
	mockDB.On("Count", mock.AnythingOfType("*int64")).Run(func(args mock.Arguments) {
		countPtr := args.Get(0).(*int64)
		*countPtr = total
	}).Return(mockDB)
	mockDB.On("Error").Return(nil).Once()
	mockDB.On("Offset", mock.AnythingOfType("int")).Return(mockDB)
	mockDB.On("Limit", mock.AnythingOfType("int")).Return(mockDB)
	mockDB.On("Find", mock.AnythingOfType("*[]models.MaterialTemplate")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*[]models.MaterialTemplate)
		*dest = templateList
	}).Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/material-templates?page=1&limit=10", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.ListMaterialTemplates(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(total), resp["total"])

	data, ok := resp["data"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(data))

	mockDB.AssertExpectations(t)
}

func TestListMaterialTemplates_WithAllParameter(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	templateList := []models.MaterialTemplate{
		{ID: 1, Name: "Template A", Category: "Category A"},
		{ID: 2, Name: "Template B", Category: "Category B"},
		{ID: 3, Name: "Template C", Category: "Category C"},
	}

	var total int64 = 3

	mockDB.On("Model", mock.AnythingOfType("*models.MaterialTemplate")).Return(mockDB)
	mockDB.On("Count", mock.AnythingOfType("*int64")).Run(func(args mock.Arguments) {
		countPtr := args.Get(0).(*int64)
		*countPtr = total
	}).Return(mockDB)
	mockDB.On("Error").Return(nil).Once()
	mockDB.On("Find", mock.AnythingOfType("*[]models.MaterialTemplate")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*[]models.MaterialTemplate)
		*dest = templateList
	}).Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/material-templates?all=true", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.ListMaterialTemplates(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(total), resp["total"])
	assert.Equal(t, true, resp["all"])

	data, ok := resp["data"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 3, len(data))

	mockDB.AssertExpectations(t)
}

// ----------------- UpdateMaterialTemplate -----------------
func TestUpdateMaterialTemplate_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	existing := models.MaterialTemplate{
		ID:       1,
		Name:     "Template A",
		Category: "Category A",
	}
	updated := models.MaterialTemplate{
		ID:       1,
		Name:     "Template A Updated",
		Category: "Category A Updated",
	}

	mockDB.On("First", mock.AnythingOfType("*models.ManagerResidentialEntry"), uint(0)).Return(nil).Maybe()
	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "1").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.MaterialTemplate)
		*dest = existing
	}).Return(nil)
	mockDB.On("Save", mock.Anything).Return(nil)
	mockDB.On("Commit").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(updated)
	req, _ := http.NewRequest("PUT", "/material-templates/1", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.UpdateMaterialTemplate(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Material template updated successfully", resp["message"])

	mockDB.AssertExpectations(t)
}

func TestUpdateMaterialTemplate_NotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	updated := models.MaterialTemplate{
		Name:     "Template A Updated",
		Category: "Category A Updated",
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "999").Return(gorm.ErrRecordNotFound)
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(updated)
	req, _ := http.NewRequest("PUT", "/material-templates/999", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "999"}}
	c.Set("db", mockDB)

	handler.UpdateMaterialTemplate(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "NOT_FOUND", resp.Code)

	mockDB.AssertExpectations(t)
}

func TestUpdateMaterialTemplate_MissingRequiredFields(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	input := models.MaterialTemplate{
		Name: "Template A",
		// Missing Category
	}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("PUT", "/material-templates/1", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.UpdateMaterialTemplate(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
}

// ----------------- DeleteMaterialTemplate -----------------
func TestDeleteMaterialTemplate_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	existing := models.MaterialTemplate{
		ID:       1,
		Name:     "Template A",
		Category: "Category A",
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "1").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.MaterialTemplate)
		*dest = existing
	}).Return(nil)
	mockDB.On("Delete", mock.Anything).Return(nil)
	mockDB.On("Commit").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("DELETE", "/material-templates/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.DeleteMaterialTemplate(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Material template deleted successfully", resp["message"])

	mockDB.AssertExpectations(t)
}

func TestDeleteMaterialTemplate_NotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "999").Return(gorm.ErrRecordNotFound)
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("DELETE", "/material-templates/999", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "999"}}
	c.Set("db", mockDB)

	handler.DeleteMaterialTemplate(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "NOT_FOUND", resp.Code)

	mockDB.AssertExpectations(t)
}

func TestDeleteMaterialTemplate_MissingID(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("DELETE", "/material-templates/", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: ""}}
	c.Set("db", mockDB)

	handler.DeleteMaterialTemplate(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
}
