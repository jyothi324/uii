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

// ----------------- CreateService -----------------
func TestCreateService_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	input := models.Service{
		Details: models.ServiceDetails{
			Name:                "HVAC Installation",
			BasicUOM:            "unit",
			SKU:                 "HVAC-001",
			ServiceCategory:     models.ServiceCategory{ID: 1, Name: "Installation"},
			ServiceType:         models.ServiceType{ID: 1, Name: "Residential"},
			ServiceGroup:        models.ServiceGroup{ID: 1, Name: "HVAC"},
			SalesAccount:        "4000",
			AvgUnitCost:         5000,
			ProAcc:              100,
			RetailAcc:           200,
			Notes:               "Standard installation",
			SpecialInstructions: "Call before arrival",
			IsTaxable:           true,
		},
	}

	mockCategory := models.ServiceCategory{ID: 1, Name: "Installation"}
	mockGroup := models.ServiceGroup{ID: 1, Name: "HVAC"}
	mockType := models.ServiceType{ID: 1, Name: "Residential"}

	mockDB.On("First", mock.AnythingOfType("*models.ServiceCategory"), "id = ?", uint(1)).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.ServiceCategory)
		*dest = mockCategory
	}).Return(nil)

	mockDB.On("First", mock.AnythingOfType("*models.ServiceGroup"), "id = ?", uint(1)).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.ServiceGroup)
		*dest = mockGroup
	}).Return(nil)

	mockDB.On("First", mock.AnythingOfType("*models.ServiceType"), "id = ?", uint(1)).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.ServiceType)
		*dest = mockType
	}).Return(nil)

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Create", mock.Anything).Return(nil)
	mockDB.On("Commit").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/services", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateService(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp models.Service
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, input.Details.Name, resp.Details.Name)

	mockDB.AssertExpectations(t)
}

func TestCreateService_InvalidInput(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("POST", "/services", bytes.NewBuffer([]byte(`invalid json`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateService(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
}

func TestCreateService_InvalidServiceCategoryID(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	input := models.Service{
		Details: models.ServiceDetails{
			Name:            "HVAC Installation",
			ServiceCategory: models.ServiceCategory{ID: 999},
		},
	}

	mockDB.On("First", mock.AnythingOfType("*models.ServiceCategory"), "id = ?", uint(999)).Return(gorm.ErrRecordNotFound)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/services", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateService(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)

	mockDB.AssertExpectations(t)
}

func TestCreateService_InvalidServiceGroupID(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	input := models.Service{
		Details: models.ServiceDetails{
			Name:         "HVAC Installation",
			ServiceGroup: models.ServiceGroup{ID: 999},
		},
	}

	mockDB.On("First", mock.AnythingOfType("*models.ServiceGroup"), "id = ?", uint(999)).Return(gorm.ErrRecordNotFound)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/services", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateService(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)

	mockDB.AssertExpectations(t)
}

func TestCreateService_InvalidServiceTypeID(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	input := models.Service{
		Details: models.ServiceDetails{
			Name:        "HVAC Installation",
			ServiceType: models.ServiceType{ID: 999},
		},
	}

	mockDB.On("First", mock.AnythingOfType("*models.ServiceType"), "id = ?", uint(999)).Return(gorm.ErrRecordNotFound)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/services", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateService(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)

	mockDB.AssertExpectations(t)
}

func TestCreateService_DatabaseError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	input := models.Service{
		Details: models.ServiceDetails{
			Name: "HVAC Installation",
		},
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("Create", mock.Anything).Return(gorm.ErrInvalidData)
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/services", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateService(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "DATABASE_ERROR", resp.Code)

	mockDB.AssertExpectations(t)
}

// ----------------- GetService -----------------
func TestGetService_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	expected := models.Service{
		ID: 1,
		Details: models.ServiceDetails{
			Name:     "HVAC Installation",
			BasicUOM: "unit",
			SKU:      "HVAC-001",
		},
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "1").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Service)
		*dest = expected
	}).Return(nil)
	mockDB.On("Commit").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/services/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.GetService(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.Service
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, expected.Details.Name, resp.Details.Name)

	mockDB.AssertExpectations(t)
}

func TestGetService_MissingID(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/services/", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: ""}}
	c.Set("db", mockDB)

	handler.GetService(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
}

func TestGetService_NotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.Anything, "id = ?", "999").Return(gorm.ErrRecordNotFound)
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/services/999", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "999"}}
	c.Set("db", mockDB)

	handler.GetService(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "NOT_FOUND", resp.Code)

	mockDB.AssertExpectations(t)
}

// ----------------- GetAllServices -----------------
func TestGetAllServices_Success_Paginated(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	serviceList := []models.Service{
		{ID: 1, Details: models.ServiceDetails{Name: "Service A"}},
		{ID: 2, Details: models.ServiceDetails{Name: "Service B"}},
	}

	var total int64 = 2

	mockDB.On("Model", mock.AnythingOfType("*models.Service")).Return(mockDB)
	mockDB.On("Count", mock.AnythingOfType("*int64")).Run(func(args mock.Arguments) {
		countPtr := args.Get(0).(*int64)
		*countPtr = total
	}).Return(mockDB)
	mockDB.On("Error").Return(nil).Once()
	mockDB.On("Offset", mock.AnythingOfType("int")).Return(mockDB)
	mockDB.On("Limit", mock.AnythingOfType("int")).Return(mockDB)
	mockDB.On("Find", mock.AnythingOfType("*[]models.Service")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*[]models.Service)
		*dest = serviceList
	}).Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/services?page=1&limit=10", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.GetAllServices(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(total), resp["total"])
	assert.Equal(t, float64(1), resp["page"])

	data, ok := resp["data"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(data))

	mockDB.AssertExpectations(t)
}

func TestGetAllServices_Success_All(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	serviceList := []models.Service{
		{ID: 1, Details: models.ServiceDetails{Name: "Service A"}},
		{ID: 2, Details: models.ServiceDetails{Name: "Service B"}},
		{ID: 3, Details: models.ServiceDetails{Name: "Service C"}},
	}

	var total int64 = 3

	mockDB.On("Model", mock.AnythingOfType("*models.Service")).Return(mockDB)
	mockDB.On("Count", mock.AnythingOfType("*int64")).Run(func(args mock.Arguments) {
		countPtr := args.Get(0).(*int64)
		*countPtr = total
	}).Return(mockDB)
	mockDB.On("Error").Return(nil).Once()
	mockDB.On("Find", mock.AnythingOfType("*[]models.Service")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*[]models.Service)
		*dest = serviceList
	}).Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/services?all=true", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.GetAllServices(c)

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

func TestGetAllServices_WithSearch(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	serviceList := []models.Service{
		{ID: 1, Details: models.ServiceDetails{Name: "HVAC Installation"}},
	}

	var total int64 = 1

	mockDB.On("Model", mock.AnythingOfType("*models.Service")).Return(mockDB)
	mockDB.On("Where", "LOWER(details->>'Name') LIKE ?", "%hvac%").Return(mockDB)
	mockDB.On("Count", mock.AnythingOfType("*int64")).Run(func(args mock.Arguments) {
		countPtr := args.Get(0).(*int64)
		*countPtr = total
	}).Return(mockDB)
	mockDB.On("Error").Return(nil).Once()
	mockDB.On("Offset", mock.AnythingOfType("int")).Return(mockDB)
	mockDB.On("Limit", mock.AnythingOfType("int")).Return(mockDB)
	mockDB.On("Find", mock.AnythingOfType("*[]models.Service")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*[]models.Service)
		*dest = serviceList
	}).Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/services?search=HVAC&page=1&limit=10", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.GetAllServices(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(total), resp["total"])

	mockDB.AssertExpectations(t)
}

func TestGetAllServices_CountError(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	mockDB.On("Model", mock.AnythingOfType("*models.Service")).Return(mockDB)
	mockDB.On("Count", mock.AnythingOfType("*int64")).Return(mockDB)
	mockDB.On("Error").Return(gorm.ErrInvalidDB).Once()

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/services?page=1&limit=10", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.GetAllServices(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "DATABASE_ERROR", resp.Code)

	mockDB.AssertExpectations(t)
}

// ----------------- UpdateService -----------------
func TestUpdateService_Success(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	existing := models.Service{
		ID: 1,
		Details: models.ServiceDetails{
			Name:     "Old Service",
			BasicUOM: "unit",
		},
	}

	updated := models.Service{
		Details: models.ServiceDetails{
			Name:         "Updated Service",
			BasicUOM:     "hour",
			ServiceGroup: models.ServiceGroup{ID: 1},
			ServiceType:  models.ServiceType{ID: 1},
		},
	}

	mockGroup := models.ServiceGroup{ID: 1, Name: "HVAC"}
	mockType := models.ServiceType{ID: 1, Name: "Residential"}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.AnythingOfType("*models.Service"), "id = ?", "1").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Service)
		*dest = existing
	}).Return(nil)

	mockDB.On("First", mock.AnythingOfType("*models.ServiceGroup"), "id = ?", uint(1)).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.ServiceGroup)
		*dest = mockGroup
	}).Return(nil)

	mockDB.On("First", mock.AnythingOfType("*models.ServiceType"), "id = ?", uint(1)).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.ServiceType)
		*dest = mockType
	}).Return(nil)

	mockDB.On("Save", mock.Anything).Return(nil)
	mockDB.On("Commit").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(updated)
	req, _ := http.NewRequest("PUT", "/services/1", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.UpdateService(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.Service
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, updated.Details.Name, resp.Details.Name)

	mockDB.AssertExpectations(t)
}

func TestUpdateService_MissingID(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	input := models.Service{Details: models.ServiceDetails{Name: "Updated"}}
	data, _ := json.Marshal(input)

	req, _ := http.NewRequest("PUT", "/services/", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: ""}}
	c.Set("db", mockDB)

	handler.UpdateService(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
}

func TestUpdateService_InvalidInput(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("PUT", "/services/1", bytes.NewBuffer([]byte(`invalid json`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.UpdateService(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
}

func TestUpdateService_NotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	input := models.Service{Details: models.ServiceDetails{Name: "Updated"}}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.AnythingOfType("*models.Service"), "id = ?", "999").Return(gorm.ErrRecordNotFound)
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("PUT", "/services/999", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "999"}}
	c.Set("db", mockDB)

	handler.UpdateService(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "NOT_FOUND", resp.Code)

	mockDB.AssertExpectations(t)
}

func TestUpdateService_InvalidServiceGroupID(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	existing := models.Service{
		ID: 1,
		Details: models.ServiceDetails{
			Name: "Service",
		},
	}

	input := models.Service{
		Details: models.ServiceDetails{
			Name:         "Updated Service",
			ServiceGroup: models.ServiceGroup{ID: 999},
		},
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.AnythingOfType("*models.Service"), "id = ?", "1").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Service)
		*dest = existing
	}).Return(nil)
	mockDB.On("First", mock.AnythingOfType("*models.ServiceGroup"), "id = ?", uint(999)).Return(gorm.ErrRecordNotFound)
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("PUT", "/services/1", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.UpdateService(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)

	mockDB.AssertExpectations(t)
}

func TestUpdateService_InvalidServiceTypeID(t *testing.T) {
	mockDB := new(mocks.MockDatabase)

	existing := models.Service{
		ID: 1,
		Details: models.ServiceDetails{
			Name: "Service",
		},
	}

	input := models.Service{
		Details: models.ServiceDetails{
			Name:        "Updated Service",
			ServiceType: models.ServiceType{ID: 999},
		},
	}

	mockDB.On("Begin").Return(mockDB)
	mockDB.On("First", mock.AnythingOfType("*models.Service"), "id = ?", "1").Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.Service)
		*dest = existing
	}).Return(nil)
	mockDB.On("First", mock.AnythingOfType("*models.ServiceType"), "id = ?", uint(999)).Return(gorm.ErrRecordNotFound)
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("PUT", "/services/1", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	handler.UpdateService(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)

	mockDB.AssertExpectations(t)
}

// ----------------- DeleteService -----------------
func TestDeleteService(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		mockSetup      func(db *mocks.MockDatabase)
		expectedStatus int
		expectedCode   string
		expectedMsg    string
	}{
		{
			name: "Missing ID",
			id:   "",
			mockSetup: func(db *mocks.MockDatabase) {
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "VALIDATION_ERROR",
		},
		{
			name: "Service Not Found",
			id:   "999",
			mockSetup: func(db *mocks.MockDatabase) {
				db.On("Begin").Return(db)
				db.On("First", mock.Anything, "id = ?", "999").
					Return(gorm.ErrRecordNotFound)
				db.On("Rollback").Return(nil)
			},
			expectedStatus: http.StatusNotFound,
			expectedCode:   "NOT_FOUND",
		},
		{
			name: "Success",
			id:   "1",
			mockSetup: func(db *mocks.MockDatabase) {
				existing := models.Service{
					ID: 1,
					Details: models.ServiceDetails{
						Name: "HVAC Installation",
					},
				}

				db.On("Begin").Return(db)
				db.On("First", mock.Anything, "id = ?", "1").
					Run(func(args mock.Arguments) {
						dest := args.Get(0).(*models.Service)
						*dest = existing
					}).Return(nil)
				db.On("Delete", mock.Anything).Return(nil)
				db.On("Commit").Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "service deleted successfully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(mocks.MockDatabase)
			handler := &handlers.Handler{DB: mockDB}

			tt.mockSetup(mockDB)

			req, _ := http.NewRequest("DELETE", "/services/"+tt.id, nil)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{{Key: "id", Value: tt.id}}
			c.Set("db", mockDB)

			handler.DeleteService(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedCode != "" {
				var resp handlers.ErrorResponse
				_ = json.Unmarshal(w.Body.Bytes(), &resp)
				assert.Equal(t, tt.expectedCode, resp.Code)
			}

			if tt.expectedMsg != "" {
				var resp map[string]interface{}
				_ = json.Unmarshal(w.Body.Bytes(), &resp)
				assert.Equal(t, tt.expectedMsg, resp["message"])
			}

			mockDB.AssertExpectations(t)
		})
	}
}

// I'm using these two test func in sincle func
// func TestDeleteService_MissingID(t *testing.T) {
// 	mockDB := new(mocks.MockDatabase)
// 	handler := &handlers.Handler{DB: mockDB}

// 	req, _ := http.NewRequest("DELETE", "/services/", nil)
// 	w := httptest.NewRecorder()
// 	c, _ := gin.CreateTestContext(w)
// 	c.Request = req
// 	c.Params = gin.Params{{Key: "id", Value: ""}}
// 	c.Set("db", mockDB)

// 	handler.DeleteService(c)

// 	assert.Equal(t, http.StatusBadRequest, w.Code)
// 	var resp handlers.ErrorResponse
// 	json.Unmarshal(w.Body.Bytes(), &resp)
// 	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
// }

// func TestDeleteService_NotFound(t *testing.T) {
// 	mockDB := new(mocks.MockDatabase)

// 	mockDB.On("Begin").Return(mockDB)
// 	mockDB.On("First", mock.Anything, "id = ?", "999").Return(gorm.ErrRecordNotFound)
// 	mockDB.On("Rollback").Return(nil)

// 	handler := &handlers.Handler{DB: mockDB}

// 	req, _ := http.NewRequest("DELETE", "/services/999", nil)
// 	w := httptest.NewRecorder()
// 	c, _ := gin.CreateTestContext(w)
// 	c.Request = req
// 	c.Params = gin.Params{{Key: "id", Value: "999"}}
// 	c.Set("db", mockDB)

// 	handler.DeleteService(c)

// 	assert.Equal(t, http.StatusNotFound, w.Code)
// 	var resp handlers.ErrorResponse
// 	json.Unmarshal(w.Body.Bytes(), &resp)
// 	assert.Equal(t, "NOT_FOUND", resp.Code)

// 	mockDB.AssertExpectations(t)
// }
