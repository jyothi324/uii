package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"inventory/db"
	"inventory/handlers"
	"inventory/mocks"
	"inventory/models"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// Helper to create a new VisualizeCustomerHandler
func newVisualizeCustomerHandler(s3mock s3iface.S3API) *handlers.VisualizeCustomerHandler {
	return handlers.NewVisualizeCustomerHandler(s3mock, "us-east-2")
}

// Helper to inject database into Gin context
func injectDBForCustomer(database db.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("db", database)
		c.Next()
	}
}

// ── TestNewVisualizeCustomerHandler

func TestNewVisualizeCustomerHandler(t *testing.T) {
	t.Run("creates handler with correct s3 client and region", func(t *testing.T) {
		mockS3 := &mocks.MockS3Client{}
		region := "us-east-2"

		handler := handlers.NewVisualizeCustomerHandler(mockS3, region)

		assert.NotNil(t, handler)
	})
}

// ── TestGetCustomerByID

func TestGetCustomerByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns customer usage count when customer found", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}

		customer := models.VisualizeCustomer{
			ID:         1,
			Email:      "test@example.com",
			FirstName:  "John",
			LastName:   "Doe",
			UsageCount: 3,
		}

		mockDB.On("Where", "id = ?", "1").Return(mockDB)
		mockDB.On("First", mock.MatchedBy(func(v interface{}) bool {
			c := v.(*models.VisualizeCustomer)
			*c = customer
			return true
		})).Return(nil)

		router := gin.New()
		handler := newVisualizeCustomerHandler(mockS3)
		router.Use(injectDBForCustomer(mockDB))
		router.GET("/visualizer/customers/:id", handler.GetCustomerByID)

		req := httptest.NewRequest("GET", "/visualizer/customers/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, float64(3), response["remaining_tries"])
	})

	t.Run("returns error when database connection not found", func(t *testing.T) {
		mockS3 := &mocks.MockS3Client{}

		router := gin.New()
		handler := newVisualizeCustomerHandler(mockS3)
		router.GET("/visualizer/customers/:id", handler.GetCustomerByID)

		req := httptest.NewRequest("GET", "/visualizer/customers/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Database connection not found")
	})

	t.Run("returns error when customer not found", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}

		mockDB.On("Where", "id = ?", "999").Return(mockDB)
		mockDB.On("First", mock.Anything).Return(gorm.ErrRecordNotFound)

		router := gin.New()
		handler := newVisualizeCustomerHandler(mockS3)
		router.Use(injectDBForCustomer(mockDB))
		router.GET("/visualizer/customers/:id", handler.GetCustomerByID)

		req := httptest.NewRequest("GET", "/visualizer/customers/999", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Customer not found")
	})
}

// ── TestCreateCustomer

func TestCreateCustomer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("creates new customer successfully", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}

		newCustomer := models.VisualizeCustomer{
			FirstName: "Jane",
			LastName:  "Smith",
			Email:     "jane@example.com",
			Phone:     "1234567890",
		}

		mockDB.On("Where", "email = ?", "jane@example.com").Return(mockDB)
		mockDB.On("First", mock.Anything).Return(gorm.ErrRecordNotFound)
		mockDB.On("Create", mock.Anything).Return(nil)

		router := gin.New()
		handler := newVisualizeCustomerHandler(mockS3)
		router.Use(injectDBForCustomer(mockDB))
		router.POST("/visualizer/customers", handler.CreateCustomer)

		payload := newCustomer
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/visualizer/customers", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Customer saved successfully", response["message"])
		assert.Equal(t, float64(5), response["remaining"])
	})

	t.Run("returns existing customer with remaining tries", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}

		existingCustomer := models.VisualizeCustomer{
			ID:         1,
			FirstName:  "John",
			LastName:   "Doe",
			Email:      "john@example.com",
			UsageCount: 2,
		}

		mockDB.On("Where", "email = ?", "john@example.com").Return(mockDB)
		mockDB.On("First", mock.MatchedBy(func(v interface{}) bool {
			c := v.(*models.VisualizeCustomer)
			*c = existingCustomer
			return true
		})).Return(nil)

		router := gin.New()
		handler := newVisualizeCustomerHandler(mockS3)
		router.Use(injectDBForCustomer(mockDB))
		router.POST("/visualizer/customers", handler.CreateCustomer)

		payload := existingCustomer
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/visualizer/customers", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Welcome back", response["message"])
		assert.Equal(t, float64(2), response["remaining"])
		assert.Equal(t, true, response["exists"])
	})

	t.Run("returns error when customer has no remaining tries", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}

		exhaustedCustomer := models.VisualizeCustomer{
			ID:         1,
			Email:      "exhausted@example.com",
			UsageCount: 0,
		}

		mockDB.On("Where", "email = ?", "exhausted@example.com").Return(mockDB)
		mockDB.On("First", mock.MatchedBy(func(v interface{}) bool {
			c := v.(*models.VisualizeCustomer)
			*c = exhaustedCustomer
			return true
		})).Return(nil)

		router := gin.New()
		handler := newVisualizeCustomerHandler(mockS3)
		router.Use(injectDBForCustomer(mockDB))
		router.POST("/visualizer/customers", handler.CreateCustomer)

		payload := exhaustedCustomer
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/visualizer/customers", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Contains(t, w.Body.String(), "You have used all 5 tries")
	})

	t.Run("returns error on invalid JSON payload", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}

		router := gin.New()
		handler := newVisualizeCustomerHandler(mockS3)
		router.Use(injectDBForCustomer(mockDB))
		router.POST("/visualizer/customers", handler.CreateCustomer)

		req := httptest.NewRequest("POST", "/visualizer/customers", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request payload")
	})

	t.Run("returns error when database connection not found", func(t *testing.T) {
		mockS3 := &mocks.MockS3Client{}

		router := gin.New()
		handler := newVisualizeCustomerHandler(mockS3)
		router.POST("/visualizer/customers", handler.CreateCustomer)

		payload := models.VisualizeCustomer{
			Email:     "test@example.com",
			FirstName: "Test",
			LastName:  "User",
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/visualizer/customers", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Database connection not found")
	})
}

// ── TestCheckEmailExists

func TestCheckEmailExists(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns remaining tries for new email", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}

		mockDB.On("Where", "email = ?", "newemail@example.com").Return(mockDB)
		mockDB.On("First", mock.Anything).Return(gorm.ErrRecordNotFound)

		router := gin.New()
		handler := newVisualizeCustomerHandler(mockS3)
		router.Use(injectDBForCustomer(mockDB))
		router.GET("/visualizer/check-email", handler.CheckEmailExists)

		req := httptest.NewRequest("GET", "/visualizer/check-email?email=newemail@example.com", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, false, response["exists"])
		assert.Equal(t, float64(5), response["remaining"])
	})

	t.Run("returns customer info when email exists and matches", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}

		customer := models.VisualizeCustomer{
			ID:         1,
			Email:      "john@example.com",
			FirstName:  "John",
			LastName:   "Doe",
			Phone:      "1234567890",
			UsageCount: 3,
		}

		mockDB.On("Where", "email = ?", "john@example.com").Return(mockDB)
		mockDB.On("First", mock.MatchedBy(func(v interface{}) bool {
			c := v.(*models.VisualizeCustomer)
			*c = customer
			return true
		})).Return(nil)

		router := gin.New()
		handler := newVisualizeCustomerHandler(mockS3)
		router.Use(injectDBForCustomer(mockDB))
		router.GET("/visualizer/check-email", handler.CheckEmailExists)

		req := httptest.NewRequest("GET", "/visualizer/check-email?email=john@example.com&last_name=Doe&phone=1234567890", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, true, response["exists"])
		assert.Equal(t, true, response["matches"])
		assert.Equal(t, float64(3), response["remaining"])
	})

	t.Run("returns error when email is missing", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}

		router := gin.New()
		handler := newVisualizeCustomerHandler(mockS3)
		router.Use(injectDBForCustomer(mockDB))
		router.GET("/visualizer/check-email", handler.CheckEmailExists)

		req := httptest.NewRequest("GET", "/visualizer/check-email", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Email is required")
	})
}

// ── TestCheckPhoneExists

func TestCheckPhoneExists(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns false when phone does not exist", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}

		mockDB.On("Where", "phone = ?", "9999999999").Return(mockDB)
		mockDB.On("First", mock.Anything).Return(gorm.ErrRecordNotFound)

		router := gin.New()
		handler := newVisualizeCustomerHandler(mockS3)
		router.Use(injectDBForCustomer(mockDB))
		router.GET("/visualizer/check-phone", handler.CheckPhoneExists)

		req := httptest.NewRequest("GET", "/visualizer/check-phone?phone=9999999999", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, false, response["exists"])
	})

	t.Run("returns true when phone already exists", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}

		customer := models.VisualizeCustomer{
			ID:    1,
			Phone: "1234567890",
		}

		mockDB.On("Where", "phone = ?", "1234567890").Return(mockDB)
		mockDB.On("First", mock.MatchedBy(func(v interface{}) bool {
			c := v.(*models.VisualizeCustomer)
			*c = customer
			return true
		})).Return(nil)

		router := gin.New()
		handler := newVisualizeCustomerHandler(mockS3)
		router.Use(injectDBForCustomer(mockDB))
		router.GET("/visualizer/check-phone", handler.CheckPhoneExists)

		req := httptest.NewRequest("GET", "/visualizer/check-phone?phone=1234567890", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, true, response["exists"])
		assert.Contains(t, w.Body.String(), "Phone number already exists")
	})

	t.Run("returns error when phone is missing", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}

		router := gin.New()
		handler := newVisualizeCustomerHandler(mockS3)
		router.Use(injectDBForCustomer(mockDB))
		router.GET("/visualizer/check-phone", handler.CheckPhoneExists)

		req := httptest.NewRequest("GET", "/visualizer/check-phone", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Phone is required")
	})
}

// ── TestUpdateCustomerImages

func TestUpdateCustomerImages(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Helper to create multipart form with two files
	createMultipartForm := func(uploadedFileName, resultFileName string) (*bytes.Buffer, string) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		uploadedPart, _ := writer.CreateFormFile("uploaded_image", uploadedFileName)
		uploadedPart.Write([]byte("fake png data"))

		resultPart, _ := writer.CreateFormFile("result_image", resultFileName)
		resultPart.Write([]byte("fake jpg data"))

		writer.Close()
		return body, writer.FormDataContentType()
	}

	t.Run("uploads images successfully and decrements usage count", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		customer := models.VisualizeCustomer{
			ID:         1,
			Email:      "test@example.com",
			UsageCount: 3,
			Images:     []models.ImagePair{},
		}

		mockDB.On("Where", "id = ?", "1").Return(mockDB)
		mockDB.On("First", mock.MatchedBy(func(v interface{}) bool {
			c := v.(*models.VisualizeCustomer)
			*c = customer
			return true
		})).Return(nil)

		mockS3.On("PutObject", mock.MatchedBy(func(input *s3.PutObjectInput) bool {
			return strings.Contains(*input.Key, "files/")
		})).Return(&s3.PutObjectOutput{}, nil)

		mockDB.On("Save", mock.Anything).Return(nil)

		router := gin.New()
		handler := newVisualizeCustomerHandler(mockS3)
		router.Use(injectDBForCustomer(mockDB))
		router.PUT("/visualizer/customers/:id/images", handler.UpdateCustomerImages)

		body, contentType := createMultipartForm("uploaded.png", "result.jpg")

		req := httptest.NewRequest("PUT", "/visualizer/customers/1/images", body)
		req.Header.Set("Content-Type", contentType)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Images saved successfully", response["message"])
		assert.Equal(t, float64(2), response["remaining"])
	})

	t.Run("returns error when customer not found", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		mockDB.On("Where", "id = ?", "999").Return(mockDB)
		mockDB.On("First", mock.Anything).Return(gorm.ErrRecordNotFound)

		router := gin.New()
		handler := newVisualizeCustomerHandler(mockS3)
		router.Use(injectDBForCustomer(mockDB))
		router.PUT("/visualizer/customers/:id/images", handler.UpdateCustomerImages)

		body, contentType := createMultipartForm("uploaded.png", "result.jpg")

		req := httptest.NewRequest("PUT", "/visualizer/customers/999/images", body)
		req.Header.Set("Content-Type", contentType)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Customer not found")
	})

	t.Run("returns error when customer has no remaining tries", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		exhaustedCustomer := models.VisualizeCustomer{
			ID:         1,
			Email:      "exhausted@example.com",
			UsageCount: 0,
		}

		mockDB.On("Where", "id = ?", "1").Return(mockDB)
		mockDB.On("First", mock.MatchedBy(func(v interface{}) bool {
			c := v.(*models.VisualizeCustomer)
			*c = exhaustedCustomer
			return true
		})).Return(nil)

		router := gin.New()
		handler := newVisualizeCustomerHandler(mockS3)
		router.Use(injectDBForCustomer(mockDB))
		router.PUT("/visualizer/customers/:id/images", handler.UpdateCustomerImages)

		body, contentType := createMultipartForm("uploaded.png", "result.jpg")

		req := httptest.NewRequest("PUT", "/visualizer/customers/1/images", body)
		req.Header.Set("Content-Type", contentType)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Contains(t, w.Body.String(), "All 5 tries have been used")
	})

	t.Run("returns error when uploaded_image is missing", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		customer := models.VisualizeCustomer{
			ID:         1,
			UsageCount: 3,
		}

		mockDB.On("Where", "id = ?", "1").Return(mockDB)
		mockDB.On("First", mock.MatchedBy(func(v interface{}) bool {
			c := v.(*models.VisualizeCustomer)
			*c = customer
			return true
		})).Return(nil)

		router := gin.New()
		handler := newVisualizeCustomerHandler(mockS3)
		router.Use(injectDBForCustomer(mockDB))
		router.PUT("/visualizer/customers/:id/images", handler.UpdateCustomerImages)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		resultPart, _ := writer.CreateFormFile("result_image", "result.jpg")
		resultPart.Write([]byte("fake jpg data"))
		writer.Close()

		req := httptest.NewRequest("PUT", "/visualizer/customers/1/images", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "uploaded_image file is required")
	})

	t.Run("returns error when result_image is missing", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		customer := models.VisualizeCustomer{
			ID:         1,
			UsageCount: 3,
		}

		mockDB.On("Where", "id = ?", "1").Return(mockDB)
		mockDB.On("First", mock.MatchedBy(func(v interface{}) bool {
			c := v.(*models.VisualizeCustomer)
			*c = customer
			return true
		})).Return(nil)

		router := gin.New()
		handler := newVisualizeCustomerHandler(mockS3)
		router.Use(injectDBForCustomer(mockDB))
		router.PUT("/visualizer/customers/:id/images", handler.UpdateCustomerImages)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		uploadedPart, _ := writer.CreateFormFile("uploaded_image", "uploaded.png")
		uploadedPart.Write([]byte("fake png data"))
		writer.Close()

		req := httptest.NewRequest("PUT", "/visualizer/customers/1/images", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "result_image file is required")
	})

	t.Run("returns error when S3 upload fails", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		customer := models.VisualizeCustomer{
			ID:         1,
			UsageCount: 3,
		}

		mockDB.On("Where", "id = ?", "1").Return(mockDB)
		mockDB.On("First", mock.MatchedBy(func(v interface{}) bool {
			c := v.(*models.VisualizeCustomer)
			*c = customer
			return true
		})).Return(nil)

		mockS3.On("PutObject", mock.Anything).Return((*s3.PutObjectOutput)(nil), fmt.Errorf("S3 upload failed"))

		router := gin.New()
		handler := newVisualizeCustomerHandler(mockS3)
		router.Use(injectDBForCustomer(mockDB))
		router.PUT("/visualizer/customers/:id/images", handler.UpdateCustomerImages)

		body, contentType := createMultipartForm("uploaded.png", "result.jpg")

		req := httptest.NewRequest("PUT", "/visualizer/customers/1/images", body)
		req.Header.Set("Content-Type", contentType)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to upload")
	})

	t.Run("returns error when database find fails in GetAllCustomers", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Find", mock.Anything).Return(fmt.Errorf("db error"))

		router := gin.New()
		handler := newVisualizeCustomerHandler(nil)
		router.Use(injectDBForCustomer(mockDB))
		router.GET("/visualizer/customers", handler.GetAllCustomers)

		req := httptest.NewRequest("GET", "/visualizer/customers", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
