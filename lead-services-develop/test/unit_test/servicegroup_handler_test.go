package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"errors"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/tuliporg/agent-management-api/handlers"
	"github.com/tuliporg/agent-management-api/mocks"
	"github.com/tuliporg/agent-management-api/models"
	"gorm.io/gorm"
)

// ----------------- CreateServiceGroup -----------------
func TestCreateServiceGroup_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	input := handlers.CreateServiceGroupInput{
		Name:       "Test Group",
		ServiceIDs: []uint{1, 2, 3},
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Create", mock.Anything).Return(nil)
	mockDB.On("Commit").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/service-groups", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateServiceGroup(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "service group created", resp["message"])

	// Verify data structure
	data_resp, ok := resp["data"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "Test Group", data_resp["name"])

	mockDB.AssertExpectations(t)
}

func TestCreateServiceGroup_WithoutServiceIDs(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	input := handlers.CreateServiceGroupInput{
		Name:       "Empty Group",
		ServiceIDs: []uint{},
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Create", mock.Anything).Return(nil)
	mockDB.On("Commit").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/service-groups", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateServiceGroup(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockDB.AssertExpectations(t)
}

func TestCreateServiceGroup_DuplicateServiceIDs(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	input := handlers.CreateServiceGroupInput{
		Name:       "Duplicate Group",
		ServiceIDs: []uint{1, 2, 2, 3, 1}, // Duplicates should be removed
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Create", mock.Anything).Return(nil)
	mockDB.On("Commit").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/service-groups", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateServiceGroup(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockDB.AssertExpectations(t)
}

func TestCreateServiceGroup_InvalidInput(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("POST", "/service-groups", bytes.NewBuffer([]byte(`invalid json`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateServiceGroup(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateServiceGroup_MissingName(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	input := map[string]interface{}{
		"service_ids": []uint{1, 2, 3},
	}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/service-groups", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateServiceGroup(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ----------------- GetServiceGroup -----------------
func TestGetServiceGroup_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	expectedGroup := handlers.ServiceGroupLite{
		ID:         1,
		Name:       "Test Group",
		ServiceIDs: models.Int64Array{1, 2, 3},
	}

	mockServices := []models.Service{
		{
			ID: 1,
			Details: models.ServiceDetails{
				Name:     "Service 1",
				BasicUOM: "Hour",
			},
		},
		{
			ID: 2,
			Details: models.ServiceDetails{
				Name:     "Service 2",
				BasicUOM: "Day",
			},
		},
	}

	mockDB.On("Model", mock.AnythingOfType("*models.ServiceGroup")).Return(mockDB)
	mockDB.On("Select", "id, name, service_ids").Return(mockDB)
	mockDB.On("First", mock.Anything, "1").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*handlers.ServiceGroupLite)
		*dest = expectedGroup
	}).Return(nil)

	mockDB.On("Model", mock.AnythingOfType("*models.Service")).Return(mockDB)
	mockDB.On("Select", "id, details").Return(mockDB)
	mockDB.On("Where", "id IN ?", mock.Anything).Return(mockDB)
	mockDB.On("Find", mock.AnythingOfType("*[]models.Service")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*[]models.Service)
		*dest = mockServices
	}).Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/service-groups/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.GetServiceGroup(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp handlers.ServiceGroupWithServices
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, expectedGroup.Name, resp.Name)
	assert.Equal(t, 2, len(resp.Services))

	mockDB.AssertExpectations(t)
}

func TestGetServiceGroup_NotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("Model", mock.AnythingOfType("*models.ServiceGroup")).Return(mockDB)
	mockDB.On("Select", "id, name, service_ids").Return(mockDB)
	mockDB.On("First", mock.Anything, "999").Return(gorm.ErrRecordNotFound)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/service-groups/999", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "999"}}
	c.Set("db", mockDB)

	handler.GetServiceGroup(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockDB.AssertExpectations(t)
}

func TestGetServiceGroup_FirstDBError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("Model", mock.Anything).Return(mockDB)
	mockDB.On("Select", "id, name, service_ids").Return(mockDB)
	mockDB.On("First", mock.Anything, "1").
		Return(errors.New("db failure"))

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/service-groups/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.GetServiceGroup(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockDB.AssertExpectations(t)
}

func TestGetServiceGroup_EmptyServiceIDs(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	group := handlers.ServiceGroupLite{
		ID:         1,
		Name:       "Empty Services",
		ServiceIDs: models.Int64Array{},
	}

	mockDB.On("Model", mock.Anything).Return(mockDB)
	mockDB.On("Select", "id, name, service_ids").Return(mockDB)
	mockDB.On("First", mock.Anything, "1").
		Run(func(args mock.Arguments) {
			*args.Get(0).(*handlers.ServiceGroupLite) = group
		}).Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/service-groups/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.GetServiceGroup(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockDB.AssertExpectations(t)
}

func TestGetServiceGroup_DuplicateAndZeroServiceIDs(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	group := handlers.ServiceGroupLite{
		ID:         1,
		Name:       "Dup Test",
		ServiceIDs: models.Int64Array{0, 1, 1, 2, 2},
	}

	services := []models.Service{
		{ID: 1, Details: models.ServiceDetails{Name: "S1"}},
		{ID: 2, Details: models.ServiceDetails{Name: "S2"}},
	}

	mockDB.On("Model", mock.AnythingOfType("*models.ServiceGroup")).Return(mockDB)
	mockDB.On("Select", "id, name, service_ids").Return(mockDB)
	mockDB.On("First", mock.Anything, "1").
		Run(func(args mock.Arguments) {
			*args.Get(0).(*handlers.ServiceGroupLite) = group
		}).Return(nil)

	mockDB.On("Model", mock.AnythingOfType("*models.Service")).Return(mockDB)
	mockDB.On("Select", "id, details").Return(mockDB)
	mockDB.On("Where", "id IN ?", mock.Anything).Return(mockDB)
	mockDB.On("Find", mock.Anything).
		Run(func(args mock.Arguments) {
			*args.Get(0).(*[]models.Service) = services
		}).Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/service-groups/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.GetServiceGroup(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockDB.AssertExpectations(t)
}

func TestGetServiceGroup_ServiceFetchError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	group := handlers.ServiceGroupLite{
		ID:         1,
		Name:       "Fail",
		ServiceIDs: models.Int64Array{1},
	}

	mockDB.On("Model", mock.Anything).Return(mockDB)
	mockDB.On("Select", "id, name, service_ids").Return(mockDB)
	mockDB.On("First", mock.Anything, "1").
		Run(func(args mock.Arguments) {
			*args.Get(0).(*handlers.ServiceGroupLite) = group
		}).Return(nil)

	mockDB.On("Model", mock.AnythingOfType("*models.Service")).Return(mockDB)
	mockDB.On("Select", "id, details").Return(mockDB)
	mockDB.On("Where", "id IN ?", mock.Anything).Return(mockDB)
	mockDB.On("Find", mock.Anything).
		Return(errors.New("service fetch error"))

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/service-groups/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.GetServiceGroup(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockDB.AssertExpectations(t)
}

func TestGetServiceGroup_PartialServicesFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	group := handlers.ServiceGroupLite{
		ID:         1,
		Name:       "Partial",
		ServiceIDs: models.Int64Array{1, 2, 3},
	}

	services := []models.Service{
		{ID: 1, Details: models.ServiceDetails{Name: "S1"}},
		// ID 2 & 3 missing
	}

	mockDB.On("Model", mock.Anything).Return(mockDB)
	mockDB.On("Select", "id, name, service_ids").Return(mockDB)
	mockDB.On("First", mock.Anything, "1").
		Run(func(args mock.Arguments) {
			*args.Get(0).(*handlers.ServiceGroupLite) = group
		}).Return(nil)

	mockDB.On("Model", mock.AnythingOfType("*models.Service")).Return(mockDB)
	mockDB.On("Select", "id, details").Return(mockDB)
	mockDB.On("Where", "id IN ?", mock.Anything).Return(mockDB)
	mockDB.On("Find", mock.Anything).
		Run(func(args mock.Arguments) {
			*args.Get(0).(*[]models.Service) = services
		}).Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/service-groups/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.GetServiceGroup(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockDB.AssertExpectations(t)
}

// ----------------- GetAllServiceGroups -----------------

func TestGetAllServiceGroups_AllRecords_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	groupList := []handlers.ServiceGroupLite{
		{ID: 1, Name: "Group A", ServiceIDs: models.Int64Array{1}},
		{ID: 2, Name: "Group B", ServiceIDs: models.Int64Array{2}},
	}

	mockServices := []models.Service{
		{ID: 1, Details: models.ServiceDetails{Name: "Service 1"}},
		{ID: 2, Details: models.ServiceDetails{Name: "Service 2"}},
	}

	// Mock query for all records
	// --- Service groups query ---
	mockDB.On("Model", mock.AnythingOfType("*models.ServiceGroup")).Return(mockDB)
	mockDB.On(
		"Select",
		"id, name, service_ids, description",
	).Return(mockDB)
	mockDB.On("Find", mock.AnythingOfType("*[]handlers.ServiceGroupLite")).
		Run(func(args mock.Arguments) {
			dest := args.Get(0).(*[]handlers.ServiceGroupLite)
			*dest = groupList
		}).
		Return(nil)

	// --- Services query ---
	mockDB.On("Model", mock.AnythingOfType("*models.Service")).Return(mockDB)
	mockDB.On(
		"Select",
		"id, details",
	).Return(mockDB)
	mockDB.On("Where", "id IN ?", mock.Anything).Return(mockDB)
	mockDB.On("Find", mock.AnythingOfType("*[]models.Service")).
		Run(func(args mock.Arguments) {
			dest := args.Get(0).(*[]models.Service)
			*dest = mockServices
		}).
		Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/service-groups?all=true", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.GetAllServiceGroups(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	data, ok := resp["data"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(data))

	mockDB.AssertExpectations(t)
}

func TestGetAllServiceGroups_All_FindError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("Model", mock.Anything).Return(mockDB)
	mockDB.On("Select", mock.Anything).Return(mockDB)
	mockDB.On("Find", mock.Anything).
		Return(errors.New("db error"))

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/service-groups?all=true", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.GetAllServiceGroups(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockDB.AssertExpectations(t)
}

// ----------------- UpdateServiceGroup -----------------
func TestUpdateServiceGroup_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	existing := models.ServiceGroup{
		ID:         1,
		Name:       "Old Name",
		ServiceIDs: models.Int64Array{1, 2},
	}

	input := handlers.UpdateServiceGroupInput{
		Name:       "Updated Name",
		ServiceIDs: []uint{2, 3, 4},
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "1").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.ServiceGroup)
		*dest = existing
	}).Return(nil)
	mockDB.On("Save", mock.Anything).Return(nil)
	mockDB.On("Commit").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("PUT", "/service-groups/1", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.UpdateServiceGroup(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "service group updated", resp["message"])

	mockDB.AssertExpectations(t)
}

func TestUpdateServiceGroup_ClearServiceIDs(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	existing := models.ServiceGroup{
		ID:         1,
		Name:       "Test Group",
		ServiceIDs: models.Int64Array{1, 2, 3},
	}

	input := handlers.UpdateServiceGroupInput{
		Name:       "Test Group",
		ServiceIDs: []uint{}, // Clear all service IDs
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "1").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.ServiceGroup)
		*dest = existing
	}).Return(nil)
	mockDB.On("Save", mock.Anything).Return(nil)
	mockDB.On("Commit").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("PUT", "/service-groups/1", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.UpdateServiceGroup(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockDB.AssertExpectations(t)
}

func TestUpdateServiceGroup_NotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	input := handlers.UpdateServiceGroupInput{
		Name:       "Updated Name",
		ServiceIDs: []uint{1, 2},
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "999").Return(gorm.ErrRecordNotFound)
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("PUT", "/service-groups/999", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "999"}}
	c.Set("db", mockDB)

	handler.UpdateServiceGroup(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockDB.AssertExpectations(t)
}

func TestUpdateServiceGroup_InvalidJSON(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("PUT", "/service-groups/1",
		bytes.NewBuffer([]byte(`{invalid json}`)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.UpdateServiceGroup(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateServiceGroup_FirstDBError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	input := handlers.UpdateServiceGroupInput{
		Name: "Test",
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "1").
		Return(errors.New("db failure"))
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("PUT", "/service-groups/1", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.UpdateServiceGroup(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockDB.AssertExpectations(t)
}

func TestUpdateServiceGroup_SaveError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	existing := models.ServiceGroup{ID: 1}

	input := handlers.UpdateServiceGroupInput{
		Name:       "New",
		ServiceIDs: []uint{1, 2},
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "1").
		Run(func(args mock.Arguments) {
			*args.Get(0).(*models.ServiceGroup) = existing
		}).Return(nil)

	mockDB.On("Save", mock.Anything).
		Return(errors.New("save failed"))
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("PUT", "/service-groups/1", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.UpdateServiceGroup(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockDB.AssertExpectations(t)
}

func TestUpdateServiceGroup_CommitError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	existing := models.ServiceGroup{ID: 1}

	input := handlers.UpdateServiceGroupInput{
		Name:       "New",
		ServiceIDs: []uint{1},
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "1").
		Run(func(args mock.Arguments) {
			*args.Get(0).(*models.ServiceGroup) = existing
		}).Return(nil)

	mockDB.On("Save", mock.Anything).Return(nil)
	mockDB.On("Commit").
		Return(errors.New("commit failed"))

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("PUT", "/service-groups/1", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.UpdateServiceGroup(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockDB.AssertExpectations(t)
}

func TestUpdateServiceGroup_PanicRollback(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "1").
		Run(func(args mock.Arguments) {
			panic("panic inside handler")
		})
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	input := handlers.UpdateServiceGroupInput{Name: "Test"}
	data, _ := json.Marshal(input)

	req, _ := http.NewRequest("PUT", "/service-groups/1", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	assert.NotPanics(t, func() {
		handler.UpdateServiceGroup(c)
	})

	mockDB.AssertExpectations(t)
}

func TestUpdateServiceGroup_DuplicateServiceIDs(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	existing := models.ServiceGroup{ID: 1}

	input := handlers.UpdateServiceGroupInput{
		Name:       "Dup Test",
		ServiceIDs: []uint{1, 1, 2, 2, 3},
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "1").
		Run(func(args mock.Arguments) {
			*args.Get(0).(*models.ServiceGroup) = existing
		}).Return(nil)

	mockDB.On("Save", mock.Anything).Return(nil)
	mockDB.On("Commit").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("PUT", "/service-groups/1", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.UpdateServiceGroup(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockDB.AssertExpectations(t)
}

// ----------------- DeleteServiceGroup -----------------
func TestDeleteServiceGroup_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	existing := models.ServiceGroup{
		ID:         1,
		Name:       "Test Group",
		ServiceIDs: models.Int64Array{1, 2},
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "1").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.ServiceGroup)
		*dest = existing
	}).Return(nil)
	mockDB.On("Delete", mock.Anything).Return(nil)
	mockDB.On("Commit").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("DELETE", "/service-groups/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.DeleteServiceGroup(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "service group deleted", resp["message"])

	mockDB.AssertExpectations(t)
}

func TestDeleteServiceGroup_NotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "999").Return(gorm.ErrRecordNotFound)
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("DELETE", "/service-groups/999", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "999"}}
	c.Set("db", mockDB)

	handler.DeleteServiceGroup(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockDB.AssertExpectations(t)
}

func TestDeleteServiceGroup_FirstDBError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "1").
		Return(errors.New("db error"))
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("DELETE", "/service-groups/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.DeleteServiceGroup(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockDB.AssertExpectations(t)
}

func TestDeleteServiceGroup_DeleteError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	existing := models.ServiceGroup{ID: 1}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "1").
		Run(func(args mock.Arguments) {
			*args.Get(0).(*models.ServiceGroup) = existing
		}).Return(nil)

	mockDB.On("Delete", mock.Anything).
		Return(errors.New("delete failed"))
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("DELETE", "/service-groups/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.DeleteServiceGroup(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockDB.AssertExpectations(t)
}

func TestDeleteServiceGroup_CommitError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	existing := models.ServiceGroup{ID: 1}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "1").
		Run(func(args mock.Arguments) {
			*args.Get(0).(*models.ServiceGroup) = existing
		}).Return(nil)

	mockDB.On("Delete", mock.Anything).Return(nil)
	mockDB.On("Commit").
		Return(errors.New("commit failed"))

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("DELETE", "/service-groups/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.DeleteServiceGroup(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockDB.AssertExpectations(t)
}

func TestDeleteServiceGroup_PanicRollback(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "1").
		Run(func(args mock.Arguments) {
			panic("panic inside handler")
		})
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("DELETE", "/service-groups/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	assert.NotPanics(t, func() {
		handler.DeleteServiceGroup(c)
	})

	mockDB.AssertExpectations(t)
}
