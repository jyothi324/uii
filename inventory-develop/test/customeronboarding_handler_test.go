package test

import (
	"bytes"
	"encoding/json"
	"errors"
	"inventory/handlers"
	"inventory/mocks"
	"inventory/models"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestCustomerOnboardingHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("HandleCreateCustomerOnboarding - Success with Logo", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		handler := handlers.NewOnboardingHandler(mockS3, "us-east-1")

	mockS3.On("PutObject", mock.Anything).Return(&s3.PutObjectOutput{}, nil)
		mockS3.On("GetObjectRequest", mock.Anything).Return(fakeS3Request(), &s3.GetObjectOutput{})
		mockDB.On("Create", mock.AnythingOfType("*models.CustomerOnboarding")).Return(nil)

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/onboarding", handler.HandleCreateCustomerOnboarding())

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("companyName", "Acme Corp")
		_ = writer.WriteField("adminName", "Jane Doe")
		_ = writer.WriteField("email", "jane@acme.com")
		_ = writer.WriteField("users", "10")
		_ = writer.WriteField("provisioningDate", "2023-12-25")

		part, _ := writer.CreateFormFile("logo", "logo.png")
		_, _ = part.Write([]byte("fake image content"))
		writer.Close()

		req, _ := http.NewRequest("POST", "/onboarding", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockDB.AssertExpectations(t)
		mockS3.AssertExpectations(t)
	})

	t.Run("HandleCreateCustomerOnboarding - DB Error", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		handler := handlers.NewOnboardingHandler(mockS3, "us-east-1")

		mockDB.On("Create", mock.Anything).Return(errors.New("db save error"))

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/onboarding", handler.HandleCreateCustomerOnboarding())

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("companyName", "Error Ltd")
		writer.Close()

		req, _ := http.NewRequest("POST", "/onboarding", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("HandleGetAllCustomerOnboardings - Success", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		handler := handlers.NewOnboardingHandler(nil, "us-east-1")

		mockDB.On("Find", mock.AnythingOfType("*[]models.CustomerOnboarding")).Return(nil)

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.GET("/onboarding", handler.HandleGetAllCustomerOnboardings())

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/onboarding", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("HandleUpdateCustomerOnboarding - Success with Text Fields", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		handler := handlers.NewOnboardingHandler(nil, "us-east-1")

		existing := models.CustomerOnboarding{CompanyName: "Old Company"}
		mockDB.On("First", mock.AnythingOfType("*models.CustomerOnboarding"), "1").
			Return(nil).
			Run(func(args mock.Arguments) {
				out := args.Get(0).(*models.CustomerOnboarding)
				*out = existing
			})
		mockDB.On("Save", mock.AnythingOfType("*models.CustomerOnboarding")).Return(nil)

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.PUT("/onboarding/:id", handler.HandleUpdateCustomerOnboarding())

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("companyName", "Updated Company")
		_ = writer.WriteField("users", "20")
		writer.Close()

		req, _ := http.NewRequest("PUT", "/onboarding/1", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp models.CustomerOnboarding
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "Updated Company", resp.CompanyName)
		assert.Equal(t, 20, resp.Users)
		mockDB.AssertExpectations(t)
	})

	t.Run("HandleUpdateCustomerOnboarding - Record Not Found", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		handler := handlers.NewOnboardingHandler(nil, "us-east-1")

		mockDB.On("First", mock.Anything, "99").Return(gorm.ErrRecordNotFound)

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.PUT("/onboarding/:id", handler.HandleUpdateCustomerOnboarding())

		req, _ := http.NewRequest("PUT", "/onboarding/99", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
