package test

import (
	"bytes"
	"encoding/json"
	"inventory/handlers"
	"inventory/mocks"
	"inventory/models"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestGetTenantBrand(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns brand info successfully", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)
		handler := handlers.NewOnboardingHandler(mockS3, "us-east-1")

		brand := models.TenantBrand{
			CompanyName: "Tulip",
			LogoURL:     "logo.png",
			Address: models.TenantAddress{
				Street:  "18440 Paulson Dr",
				City:    "Port Charlotte",
				State:   "FL",
				Zip:     "33954",
				Country: "USA",
			},
		}
		mockDB.On("First", mock.AnythingOfType("*models.TenantBrand")).Return(nil).Run(func(args mock.Arguments) {
			out := args.Get(0).(*models.TenantBrand)
			*out = brand
		})

		router := gin.New()
		router.Use(injectDB(mockDB), injectTenant("test"))
		router.GET("/brand", handler.GetTenantBrand())
		req, _ := http.NewRequest("GET", "/brand", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Tulip")
		assert.Contains(t, w.Body.String(), "18440 Paulson Dr")
	})

	t.Run("returns empty response when not found", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		handler := handlers.NewOnboardingHandler(nil, "us-east-1")
		mockDB.On("First", mock.Anything).Return(gorm.ErrRecordNotFound)
		router := gin.New()
		router.Use(injectDB(mockDB))
		router.GET("/brand", handler.GetTenantBrand())
		req, _ := http.NewRequest("GET", "/brand", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("returns address info successfully", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		handler := handlers.NewOnboardingHandler(nil, "us-east-1")

		brand := models.TenantBrand{
			CompanyName: "Tulip",
			Address: models.TenantAddress{
				Street:  "18440 Paulson Dr",
				City:    "Port Charlotte",
				State:   "FL",
				Zip:     "33954",
				Country: "USA",
				Website: "https://www.acergranitefla.com/",
				Email:   "contact@acergranitecorp.com",
				Contact: "+19416245958",
			},
		}
		mockDB.On("First", mock.AnythingOfType("*models.TenantBrand")).Return(nil).Run(func(args mock.Arguments) {
			out := args.Get(0).(*models.TenantBrand)
			*out = brand
		})

		router := gin.New()
		router.Use(injectDB(mockDB), injectTenant("test"))
		router.GET("/address", handler.GetTenantAddress())
		req, _ := http.NewRequest("GET", "/address", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "18440 Paulson Dr")
		assert.Contains(t, w.Body.String(), "https://www.acergranitefla.com/")
		assert.Contains(t, w.Body.String(), "contact@acergranitecorp.com")
	})
}

func TestSaveTenantAddress(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("updates existing tenant address", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		handler := handlers.NewOnboardingHandler(nil, "us-east-1")

		brand := models.TenantBrand{CompanyName: "Tulip", Subdomain: "test"}
		mockDB.On("First", mock.AnythingOfType("*models.TenantBrand")).Return(nil).Run(func(args mock.Arguments) {
			out := args.Get(0).(*models.TenantBrand)
			*out = brand
		})
		mockDB.On("Save", mock.AnythingOfType("*models.TenantBrand")).Return(nil).Run(func(args mock.Arguments) {
			saved := args.Get(0).(*models.TenantBrand)
			assert.Equal(t, "500 Market St", saved.Address.Street)
			assert.Equal(t, "San Francisco", saved.Address.City)
			assert.Equal(t, "https://www.acergranitefla.com/", saved.Address.Website)
			assert.Equal(t, "contact@acergranitecorp.com", saved.Address.Email)
			assert.Equal(t, "+19416245958", saved.Address.Contact)
		})

		body, _ := json.Marshal(map[string]interface{}{
			"address": map[string]string{
				"street":  " 500 Market St ",
				"city":    " San Francisco ",
				"state":   "CA",
				"zip":     "94105",
				"country": "USA",
				"website": " https://www.acergranitefla.com/ ",
				"email":   " contact@acergranitecorp.com ",
				"contact": " +19416245958 ",
			},
		})

		router := gin.New()
		router.Use(injectDB(mockDB), injectTenant("test"))
		router.POST("/address", handler.SaveTenantAddress())
		req, _ := http.NewRequest("POST", "/address", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "500 Market St")
		mockDB.AssertExpectations(t)
	})
}
