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

// ----------------- CreateProfile -----------------
func TestCreateProfile_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := new(mocks.MockDatabase)

	// ---------- INPUT ----------
	input := handlers.ProfileRequest{
		Name:           "John Doe",
		Phone:          "+1234567890",
		Email:          "john@example.com",
		SameAsShipping: false,
		ShippingAddress: handlers.AddressPayload{
			State:     "California",
			County:    "Los Angeles",
			Street:    "123 Main St",
			UnitSuite: "Apt 4B",
			City:      "Los Angeles",
			Zip:       "90001",
			Formatted: "123 Main St, Apt 4B, Los Angeles, CA 90001",
			Lat:       34.0522,
			Lng:       -118.2437,
		},
		BillingAddress: handlers.AddressPayload{
			State:     "California",
			County:    "Los Angeles",
			Street:    "123 Main St",
			UnitSuite: "Apt 4B",
			City:      "Los Angeles",
			Zip:       "90001",
			Formatted: "123 Main St, Apt 4B, Los Angeles, CA 90001",
			Lat:       34.0522,
			Lng:       -118.2437,
		},
	}

	// ---------- MOCKS ----------
	mockDB.On("Begin").Return(mockDB)

	// Rollback is usually deferred — allow it
	mockDB.On("Rollback").Return(nil).Maybe()

	// Create shipping & billing addresses
	mockDB.
		On("Create", mock.AnythingOfType("*models.Address")).
		Run(func(args mock.Arguments) {
			addr := args.Get(0).(*models.Address)
			addr.ID = 1
		}).
		Return(nil).
		Twice()

	// Create profile
	mockDB.
		On("Create", mock.AnythingOfType("*models.Profile")).
		Run(func(args mock.Arguments) {
			profile := args.Get(0).(*models.Profile)
			profile.ID = 1
		}).
		Return(nil).
		Once()

	mockDB.On("Commit").Return(nil).Once()

	// Reload profile with relations
	mockDB.On("Preload", "ShippingAddress").Return(mockDB)
	mockDB.On("Preload", "BillingAddress").Return(mockDB)
	mockDB.
		On("First", mock.AnythingOfType("*models.Profile"), uint(1)).
		Run(func(args mock.Arguments) {
			p := args.Get(0).(*models.Profile)
			p.ID = 1
			p.Name = input.Name
			p.Phone = input.Phone
			p.Email = input.Email
			p.ShippingAddress = models.Address{ID: 1, State: "California"}
			p.BillingAddress = models.Address{ID: 1, State: "California"}
		}).
		Return(nil).
		Once()

	// ---------- HANDLER ----------
	handler := &handlers.Handler{DB: mockDB}

	// ---------- REQUEST ----------
	body, _ := json.Marshal(input)
	req, _ := http.NewRequest(http.MethodPost, "/profiles", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	// ---------- EXECUTE ----------
	handler.CreateProfile(c)

	// ---------- ASSERT ----------
	assert.Equal(t, http.StatusCreated, w.Code, w.Body.String())

	var resp models.Profile
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	assert.Equal(t, input.Name, resp.Name)
	assert.Equal(t, input.Phone, resp.Phone)
	assert.Equal(t, input.Email, resp.Email)
	assert.Equal(t, uint(1), resp.ID)

	mockDB.AssertExpectations(t)
}

func TestCreateProfile_Success_DifferentBillingAddress(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockDB := new(mocks.MockDatabase)

	input := handlers.ProfileRequest{
		Name:  "Jane Doe",
		Phone: "+1234567890",
		Email: "jane@example.com",
		ShippingAddress: handlers.AddressPayload{
			State:     "California",
			County:    "Los Angeles",
			Street:    "123 Main St",
			City:      "Los Angeles",
			Zip:       "90001",
			Formatted: "123 Main St, Los Angeles, CA 90001",
			Lat:       34.0522,
			Lng:       -118.2437,
		},
		SameAsShipping: false,
		BillingAddress: handlers.AddressPayload{
			State:     "New York",
			County:    "Manhattan",
			Street:    "456 Park Ave",
			City:      "New York",
			Zip:       "10001",
			Formatted: "456 Park Ave, New York, NY 10001",
			Lat:       40.7128,
			Lng:       -74.0060,
		},
	}

	// Mock transaction
	mockDB.On("Begin").Return(mockDB)

	// Mock creating shipping address
	mockDB.On("Create", mock.AnythingOfType("*models.Address")).Run(func(args mock.Arguments) {
		addr := args.Get(0).(*models.Address)
		addr.ID = 1 // First address (shipping)
	}).Return(nil).Once()

	// Mock creating billing address (different from shipping)
	mockDB.On("Create", mock.AnythingOfType("*models.Address")).Run(func(args mock.Arguments) {
		addr := args.Get(0).(*models.Address)
		addr.ID = 2 // Second address (billing)
	}).Return(nil).Once()

	// Mock creating profile
	mockDB.On("Create", mock.AnythingOfType("*models.Profile")).Run(func(args mock.Arguments) {
		profile := args.Get(0).(*models.Profile)
		profile.ID = 1
	}).Return(nil).Once()

	mockDB.On("Commit").Return(nil)

	// Mock loading relations
	mockDB.On("Preload", "ShippingAddress").Return(mockDB)
	mockDB.On("Preload", "BillingAddress").Return(mockDB)
	mockDB.On("First", mock.AnythingOfType("*models.Profile"), uint(1)).Run(func(args mock.Arguments) {
		profile := args.Get(0).(*models.Profile)
		profile.ID = 1
		profile.Name = input.Name
		profile.ShippingAddress = models.Address{ID: 1, State: "California"}
		profile.BillingAddress = models.Address{ID: 2, State: "New York"}
	}).Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/profiles", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateProfile(c)

	assert.Equal(t, http.StatusCreated, w.Code)

	mockDB.AssertExpectations(t)
}

func TestCreateProfile_InvalidInput(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("POST", "/profiles", bytes.NewBuffer([]byte(`invalid json`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateProfile(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
}

func TestCreateProfile_MissingRequiredFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockDB := new(mocks.MockDatabase)
	handler := &handlers.Handler{DB: mockDB}

	input := handlers.ProfileRequest{
		Name: "John Doe",
		// Missing Phone, Email, and ShippingAddress
	}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/profiles", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateProfile(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
}

func TestCreateProfile_ShippingAddressCreationError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := new(mocks.MockDatabase)

	input := handlers.ProfileRequest{
		Name:           "John Doe",
		Phone:          "+1234567890",
		Email:          "john@example.com",
		SameAsShipping: true,
		ShippingAddress: handlers.AddressPayload{
			State:  "California",
			County: "Los Angeles",
			Street: "123 Main St",
			City:   "Los Angeles",
			Zip:    "90001",
		},
		BillingAddress: handlers.AddressPayload{
			State:  "California",
			County: "Los Angeles",
			Street: "123 Main St",
			City:   "Los Angeles",
			Zip:    "90001",
		},
	}

	// ---- MOCKS ----
	mockDB.On("Begin").Return(mockDB).Once()

	mockDB.
		On("Create", mock.AnythingOfType("*models.Address")).
		Return(gorm.ErrInvalidData).
		Once() // shipping address fails

	mockDB.On("Rollback").Return(nil).Maybe()

	// ---- HANDLER ----
	handler := &handlers.Handler{DB: mockDB}

	body, _ := json.Marshal(input)
	req, _ := http.NewRequest(http.MethodPost, "/profiles", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	// ---- EXECUTE ----
	handler.CreateProfile(c)

	// ---- ASSERT ----
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp handlers.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	assert.Equal(t, "DATABASE_ERROR", resp.Code)
	assert.Contains(t, resp.Error, "Failed to create shipping address")

	mockDB.AssertExpectations(t)
}

func TestCreateProfile_BillingAddressCreationError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockDB := new(mocks.MockDatabase)

	input := handlers.ProfileRequest{
		Name:  "Jane Doe",
		Phone: "+1234567890",
		Email: "jane@example.com",
		ShippingAddress: handlers.AddressPayload{
			State:  "California",
			County: "Los Angeles",
			Street: "123 Main St",
			City:   "Los Angeles",
			Zip:    "90001",
		},
		SameAsShipping: false,
		BillingAddress: handlers.AddressPayload{
			State:  "New York",
			County: "Manhattan",
			Street: "456 Park Ave",
			City:   "New York",
			Zip:    "10001",
		},
	}

	mockDB.On("Begin").Return(mockDB)
	// Shipping address succeeds
	mockDB.On("Create", mock.AnythingOfType("*models.Address")).Run(func(args mock.Arguments) {
		addr := args.Get(0).(*models.Address)
		addr.ID = 1
	}).Return(nil).Once()
	// Billing address fails
	mockDB.On("Create", mock.AnythingOfType("*models.Address")).Return(gorm.ErrInvalidData).Once()
	mockDB.On("Rollback").Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/profiles", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.CreateProfile(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "DATABASE_ERROR", resp.Code)
	assert.Contains(t, resp.Error, "Failed to create billing address")

	mockDB.AssertExpectations(t)
}

func TestCreateProfile_ProfileCreationError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := new(mocks.MockDatabase)

	input := handlers.ProfileRequest{
		Name:           "John Doe",
		Phone:          "+1234567890",
		Email:          "john@example.com",
		SameAsShipping: true,
		ShippingAddress: handlers.AddressPayload{
			State:  "California",
			County: "Los Angeles",
			Street: "123 Main St",
			City:   "Los Angeles",
			Zip:    "90001",
		},
		BillingAddress: handlers.AddressPayload{
			State:  "California",
			County: "Los Angeles",
			Street: "123 Main St",
			City:   "Los Angeles",
			Zip:    "90001",
		},
	}

	// ---- MOCKS ----
	mockDB.On("Begin").Return(mockDB).Once()

	mockDB.
		On("Create", mock.AnythingOfType("*models.Address")).
		Run(func(args mock.Arguments) {
			addr := args.Get(0).(*models.Address)
			addr.ID = 1
		}).
		Return(nil).
		Once()

	mockDB.
		On("Create", mock.AnythingOfType("*models.Profile")).
		Return(gorm.ErrInvalidData).
		Once()

	mockDB.On("Rollback").Return(nil).Once()

	// ---- HANDLER ----
	handler := &handlers.Handler{DB: mockDB}

	body, _ := json.Marshal(input)
	req, _ := http.NewRequest(http.MethodPost, "/profiles", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	// ---- EXECUTE ----
	handler.CreateProfile(c)

	// ---- ASSERT ----
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp handlers.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	assert.Equal(t, "DATABASE_ERROR", resp.Code)
	assert.Contains(t, resp.Error, "Failed to create profile")

	mockDB.AssertExpectations(t)
}

func TestCreateProfile_CommitError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := new(mocks.MockDatabase)

	input := handlers.ProfileRequest{
		Name:           "John Doe",
		Phone:          "+1234567890",
		Email:          "john@example.com",
		SameAsShipping: true,
		ShippingAddress: handlers.AddressPayload{
			State:  "California",
			County: "Los Angeles",
			Street: "123 Main St",
			City:   "Los Angeles",
			Zip:    "90001",
		},
		BillingAddress: handlers.AddressPayload{
			State:  "California",
			County: "Los Angeles",
			Street: "123 Main St",
			City:   "Los Angeles",
			Zip:    "90001",
		},
	}

	// ---- MOCKS ----
	mockDB.On("Begin").Return(mockDB).Once()

	mockDB.
		On("Create", mock.AnythingOfType("*models.Address")).
		Run(func(args mock.Arguments) {
			addr := args.Get(0).(*models.Address)
			addr.ID = 1
		}).
		Return(nil).
		Once()

	mockDB.
		On("Create", mock.AnythingOfType("*models.Profile")).
		Run(func(args mock.Arguments) {
			p := args.Get(0).(*models.Profile)
			p.ID = 1
		}).
		Return(nil).
		Once()

	mockDB.
		On("Commit").
		Return(gorm.ErrInvalidTransaction).
		Once()

	mockDB.On("Rollback").Return(nil).Maybe()

	// ---- HANDLER ----
	handler := &handlers.Handler{DB: mockDB}

	body, _ := json.Marshal(input)
	req, _ := http.NewRequest(http.MethodPost, "/profiles", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	// ---- EXECUTE ----
	handler.CreateProfile(c)

	// ---- ASSERT ----
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp handlers.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	assert.Equal(t, "DATABASE_ERROR", resp.Code)
	assert.Contains(t, resp.Error, "Failed to commit")

	mockDB.AssertExpectations(t)
}

func TestCreateProfile_PreloadError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := new(mocks.MockDatabase)

	input := handlers.ProfileRequest{
		Name:           "John Doe",
		Phone:          "+1234567890",
		Email:          "john@example.com",
		SameAsShipping: true,
		ShippingAddress: handlers.AddressPayload{
			State:  "California",
			County: "Los Angeles",
			Street: "123 Main St",
			City:   "Los Angeles",
			Zip:    "90001",
		},
		BillingAddress: handlers.AddressPayload{
			State:  "California",
			County: "Los Angeles",
			Street: "123 Main St",
			City:   "Los Angeles",
			Zip:    "90001",
		},
	}

	// ---- MOCKS ----
	mockDB.On("Begin").Return(mockDB).Once()

	mockDB.
		On("Create", mock.AnythingOfType("*models.Address")).
		Run(func(args mock.Arguments) {
			addr := args.Get(0).(*models.Address)
			addr.ID = 1
		}).
		Return(nil).
		Once() // ✅ only once when SameAsShipping=true

	mockDB.
		On("Create", mock.AnythingOfType("*models.Profile")).
		Run(func(args mock.Arguments) {
			p := args.Get(0).(*models.Profile)
			p.ID = 1
		}).
		Return(nil).
		Once()

	mockDB.On("Commit").Return(nil).Once()
	mockDB.On("Rollback").Return(nil).Maybe()

	// ---- PRELOAD FAILURE ----
	mockDB.On("Preload", "ShippingAddress").Return(mockDB).Once()
	mockDB.On("Preload", "BillingAddress").Return(mockDB).Once()
	mockDB.
		On("First", mock.AnythingOfType("*models.Profile"), uint(1)).
		Return(gorm.ErrRecordNotFound).
		Once()

	// ---- HANDLER ----
	handler := &handlers.Handler{DB: mockDB}

	body, _ := json.Marshal(input)
	req, _ := http.NewRequest(http.MethodPost, "/profiles", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	// ---- EXECUTE ----
	handler.CreateProfile(c)

	// ---- ASSERT ----
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp handlers.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	assert.Equal(t, "DATABASE_ERROR", resp.Code)
	assert.Contains(t, resp.Error, "Failed to load profile")

	mockDB.AssertExpectations(t)
}

// ----------------- GetAllProfiles -----------------
func TestGetAllProfiles_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockDB := new(mocks.MockDatabase)

	profileList := []models.Profile{
		{
			ID:    1,
			Name:  "John Doe",
			Phone: "+1234567890",
			Email: "john@example.com",
			ShippingAddress: models.Address{
				ID:    1,
				State: "California",
				City:  "Los Angeles",
			},
			BillingAddress: models.Address{
				ID:    1,
				State: "California",
				City:  "Los Angeles",
			},
		},
		{
			ID:    2,
			Name:  "Jane Smith",
			Phone: "+0987654321",
			Email: "jane@example.com",
			ShippingAddress: models.Address{
				ID:    2,
				State: "New York",
				City:  "New York",
			},
			BillingAddress: models.Address{
				ID:    3,
				State: "New York",
				City:  "New York",
			},
		},
	}

	mockDB.On("Model", mock.AnythingOfType("*models.Profile")).Return(mockDB)
	mockDB.On("Preload", "ShippingAddress").Return(mockDB)
	mockDB.On("Preload", "BillingAddress").Return(mockDB)
	mockDB.On("Find", mock.AnythingOfType("*[]models.Profile")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*[]models.Profile)
		*dest = profileList
	}).Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/profiles", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.GetAllProfiles(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	data, ok := resp["data"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(data))

	mockDB.AssertExpectations(t)
}

func TestGetAllProfiles_EmptyResult(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockDB := new(mocks.MockDatabase)

	mockDB.On("Model", mock.AnythingOfType("*models.Profile")).Return(mockDB)
	mockDB.On("Preload", "ShippingAddress").Return(mockDB)
	mockDB.On("Preload", "BillingAddress").Return(mockDB)
	mockDB.On("Find", mock.AnythingOfType("*[]models.Profile")).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*[]models.Profile)
		*dest = []models.Profile{}
	}).Return(nil)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/profiles", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.GetAllProfiles(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	data, ok := resp["data"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 0, len(data))

	mockDB.AssertExpectations(t)
}

func TestGetAllProfiles_DatabaseError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockDB := new(mocks.MockDatabase)

	mockDB.On("Model", mock.AnythingOfType("*models.Profile")).Return(mockDB)
	mockDB.On("Preload", "ShippingAddress").Return(mockDB)
	mockDB.On("Preload", "BillingAddress").Return(mockDB)
	mockDB.On("Find", mock.AnythingOfType("*[]models.Profile")).Return(gorm.ErrInvalidDB)

	handler := &handlers.Handler{DB: mockDB}

	req, _ := http.NewRequest("GET", "/profiles", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	handler.GetAllProfiles(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp handlers.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "DATABASE_ERROR", resp.Code)
	assert.Contains(t, resp.Error, "Failed to retrieve profiles")

	mockDB.AssertExpectations(t)
}

// func TestGetAllProfiles_NilDatabase(t *testing.T) {
//     gin.SetMode(gin.TestMode)
//     handler := &handlers.Handler{DB: nil}
// 	mockDB := new(mocks.MockDatabase)

// 	mockDB.On("Model", mock.AnythingOfType("*models.Profile")).Return(mockDB)
// 	mockDB.On("Find", mock.Anything).Return(nil)

//     req, _ := http.NewRequest("GET", "/profiles", nil)
//     w := httptest.NewRecorder()
//     c, _ := gin.CreateTestContext(w)
//     c.Request = req
//     // Do NOT set db in context
//     // c.Set("db", mockDB)

//     handler.GetAllProfiles(c)

//     assert.Equal(t, http.StatusInternalServerError, w.Code)
//     var resp handlers.ErrorResponse
//     json.Unmarshal(w.Body.Bytes(), &resp)
//     assert.Equal(t, "DATABASE_ERROR", resp.Code)
//     assert.Contains(t, resp.Error, "Database not initialized")
// }
