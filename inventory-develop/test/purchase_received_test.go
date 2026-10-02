package test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"inventory/handlers"
	"inventory/mocks"
	"inventory/models"

	"mime/multipart"
    "net/textproto"
    "strings"

    "github.com/aws/aws-sdk-go/aws/request"
    "github.com/aws/aws-sdk-go/service/s3/s3iface"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/aws"
"github.com/aws/aws-sdk-go/aws/client/metadata"
)

type MockS3Client struct {
    s3iface.S3API
}

func (m *MockS3Client) GetObjectRequest(input *s3.GetObjectInput) (*request.Request, *s3.GetObjectOutput) {
	req := request.New(
		aws.Config{
			Region: aws.String("us-east-2"),
		},
		metadata.ClientInfo{},
		request.Handlers{},
		nil,
		&request.Operation{
			Name:       "GetObject",
			HTTPMethod: "GET",
			HTTPPath:   "/{Bucket}/{Key+}",
		},
		&s3.GetObjectInput{},
		&s3.GetObjectOutput{},
	)

	req.HTTPRequest, _ = http.NewRequest(
		"GET",
		"https://example.com/test",
		nil,
	)

	return req, &s3.GetObjectOutput{}
}

func (m *MockS3Client) PutObject(input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
    return &s3.PutObjectOutput{}, nil
}

func TestEnrichPRImageURL(t *testing.T) {
    t.Run("returns empty when image key empty", func(t *testing.T) {
        result := handlers.EnrichPRImageURL("", "bucket",
            func(bucket, key string) (string, error) {
                return "signed-url", nil
            })

        assert.Equal(t, "", result)
    })

    t.Run("returns signed url", func(t *testing.T) {
        result := handlers.EnrichPRImageURL("image.png", "bucket",
            func(bucket, key string) (string, error) {
                return "signed-url", nil
            })

        assert.Equal(t, "signed-url", result)
    })

    t.Run("returns original key on error", func(t *testing.T) {
        result := handlers.EnrichPRImageURL("image.png", "bucket",
            func(bucket, key string) (string, error) {
                return "", errors.New("presign failed")
            })

        assert.Equal(t, "image.png", result)
    })
}

func TestUploadImage(t *testing.T) {
    gin.SetMode(gin.TestMode)

    mockS3 := &MockS3Client{}
    handler := handlers.NewPurchaseReceivedHandler(mockS3, "us-east-2")

    router := gin.New()

    router.POST("/upload", func(c *gin.Context) {
        result, err := handler.UploadImage(c, "image")

        assert.Nil(t, err)
        assert.True(t, strings.Contains(result, "purchase-received/"))

        c.JSON(http.StatusOK, gin.H{"image": result})
    })

    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)

    h := make(textproto.MIMEHeader)
    h.Set("Content-Disposition", `form-data; name="image"; filename="test.png"`)
    h.Set("Content-Type", "image/png")

    part, _ := writer.CreatePart(h)
    part.Write([]byte("fake-image"))

    writer.Close()

    req, _ := http.NewRequest("POST", "/upload", body)
    req.Header.Set("Content-Type", writer.FormDataContentType())

    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)
}

// ── HandleCreatePurchaseReceived ──────────────────────────────────────────────

func TestHandleCreatePurchaseReceived(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("valid request", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		// Handler verifies the purchase exists first, then creates the record.
		mockDB.On("First", mock.AnythingOfType("*models.Purchase"), mock.Anything).
			Return(nil).Once()
		mockDB.On("Create", mock.AnythingOfType("*models.PurchaseReceived")).Return(nil).Once()
		mockDB.On("Create", mock.AnythingOfType("*handlers.ActivityEvent")).Return(nil).Once()
		mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(nil).Once()

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/purchase-received/create", handlers.HandleCreatePurchaseReceived())

		reqBody := models.PurchaseReceived{PurchaseID: 1, InvoiceID: "INV-001", Price: 500.0}
		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/purchase-received/create", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), "Purchase received created successfully")

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "INV-001", response["invoice_id"])
		mockDB.AssertExpectations(t)
	})

	t.Run("invalid request body - malformed JSON", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/purchase-received/create", handlers.HandleCreatePurchaseReceived())

		req, _ := http.NewRequest("POST", "/purchase-received/create",
			bytes.NewBuffer([]byte(`{"price": "bad"}`)))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request body")
	})

	t.Run("referenced purchase not found", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("First", mock.AnythingOfType("*models.Purchase"), mock.Anything).
			Return(gorm.ErrRecordNotFound).Once()

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/purchase-received/create", handlers.HandleCreatePurchaseReceived())

		reqBody := models.PurchaseReceived{PurchaseID: 999, InvoiceID: "INV-001", Price: 500.0}
		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/purchase-received/create", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Purchase not found")
		mockDB.AssertExpectations(t)
	})

	t.Run("database error verifying purchase", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("First", mock.AnythingOfType("*models.Purchase"), mock.Anything).
			Return(errors.New("db error")).Once()

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/purchase-received/create", handlers.HandleCreatePurchaseReceived())

		reqBody := models.PurchaseReceived{PurchaseID: 1, InvoiceID: "INV-001", Price: 500.0}
		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/purchase-received/create", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to verify purchase")
		mockDB.AssertExpectations(t)
	})

	t.Run("database error on insert", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("First", mock.AnythingOfType("*models.Purchase"), mock.Anything).
			Return(nil).Once()
		mockDB.On("Create", mock.AnythingOfType("*models.PurchaseReceived")).
			Return(errors.New("db insert error"))

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/purchase-received/create", handlers.HandleCreatePurchaseReceived())

		reqBody := models.PurchaseReceived{PurchaseID: 1, InvoiceID: "INV-001", Price: 500.0}
		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/purchase-received/create", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to create purchase received")
		mockDB.AssertExpectations(t)
	})
}


// ── HandleFetchPurchaseReceiveds ──────────────────────────────────────────────

func TestHandleFetchPurchaseReceiveds(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("fetch all purchase receiveds successfully", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		// Handler chains: database.Preload("Purchase").Preload("Purchase.Materials").Find(...)
		mockDB.On("Preload", "Purchase").Return(mockDB)
		mockDB.On("Preload", "Purchase.Materials").Return(mockDB)
		mockDB.On("Preload", "Purchase.Supplier").Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]models.PurchaseReceived")).Return(nil)

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.GET("/purchase-received/all", handlers.NewPurchaseReceivedHandler(&s3.S3{}, "us-east-2").HandleFetchPurchaseReceiveds())

		req, _ := http.NewRequest("GET", "/purchase-received/all", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("fetch purchase receiveds - empty database", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Preload", "Purchase").Return(mockDB)
		mockDB.On("Preload", "Purchase.Materials").Return(mockDB)
		mockDB.On("Preload", "Purchase.Supplier").Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]models.PurchaseReceived")).Return(nil)

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.GET("/purchase-received/all", handlers.NewPurchaseReceivedHandler(&s3.S3{}, "us-east-2").HandleFetchPurchaseReceiveds())

		req, _ := http.NewRequest("GET", "/purchase-received/all", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var items []models.PurchaseReceived
		json.Unmarshal(w.Body.Bytes(), &items)
		assert.Equal(t, 0, len(items))
	})

	t.Run("database query error", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Preload", "Purchase").Return(mockDB)
		mockDB.On("Preload", "Purchase.Materials").Return(mockDB)
		mockDB.On("Preload", "Purchase.Supplier").Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]models.PurchaseReceived")).
			Return(errors.New("db error"))

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.GET("/purchase-received/all", handlers.NewPurchaseReceivedHandler(&s3.S3{}, "us-east-2").HandleFetchPurchaseReceiveds())

		req, _ := http.NewRequest("GET", "/purchase-received/all", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to fetch purchase received")
	})
}

// ── HandlePurchaseReceivedGetByID ─────────────────────────────────────────────

func TestHandlePurchaseReceivedGetByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("get purchase received by id successfully", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		// Handler chains: database.Preload("Purchase").Preload("Purchase.Materials").First(...)
		mockDB.On("Preload", "Purchase").Return(mockDB)
		mockDB.On("Preload", "Purchase.Materials").Return(mockDB)
		mockDB.On("Preload", "Purchase.Supplier").Return(mockDB)
		mockDB.On("First", mock.AnythingOfType("*models.PurchaseReceived"), mock.Anything).
			Return(nil).
			Run(func(args mock.Arguments) {
				pr := args.Get(0).(*models.PurchaseReceived)
				pr.ID = 1
				pr.InvoiceID = "INV-001"
			}).Once()

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.GET("/purchase-received/:id", handlers.NewPurchaseReceivedHandler(&s3.S3{}, "us-east-2").HandlePurchaseReceivedGetByID())

		req, _ := http.NewRequest("GET", "/purchase-received/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "INV-001")
		mockDB.AssertExpectations(t)
	})

	t.Run("missing id - route not matched returns 404", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.GET("/purchase-received/:id", handlers.NewPurchaseReceivedHandler(&s3.S3{}, "us-east-2").HandlePurchaseReceivedGetByID())

		req, _ := http.NewRequest("GET", "/purchase-received/", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("purchase received not found", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Preload", "Purchase").Return(mockDB)
		mockDB.On("Preload", "Purchase.Materials").Return(mockDB)
		mockDB.On("Preload", "Purchase.Supplier").Return(mockDB)
		// Handler returns 200 with a message when the record is not found (by design).
		mockDB.On("First", mock.AnythingOfType("*models.PurchaseReceived"), mock.Anything).
			Return(gorm.ErrRecordNotFound)

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.GET("/purchase-received/:id", handlers.NewPurchaseReceivedHandler(&s3.S3{}, "us-east-2").HandlePurchaseReceivedGetByID())

		req, _ := http.NewRequest("GET", "/purchase-received/99999", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Purchase received not found")
	})

	t.Run("database error on fetch", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("Preload", "Purchase").Return(mockDB)
		mockDB.On("Preload", "Purchase.Materials").Return(mockDB)
		mockDB.On("Preload", "Purchase.Supplier").Return(mockDB)
		mockDB.On("First", mock.AnythingOfType("*models.PurchaseReceived"), mock.Anything).
			Return(errors.New("db error"))

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.GET("/purchase-received/:id", handlers.NewPurchaseReceivedHandler(&s3.S3{}, "us-east-2").HandlePurchaseReceivedGetByID())

		req, _ := http.NewRequest("GET", "/purchase-received/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to fetch purchase received")
	})
}

// ── HandlePurchaseReceivedUpdate ──────────────────────────────────────────────

func TestHandlePurchaseReceivedUpdate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("update purchase received successfully", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		existingPR := models.PurchaseReceived{PurchaseID: 1, InvoiceID: "INV-001", Price: 500.0}

		mockDB.On("First", mock.AnythingOfType("*models.PurchaseReceived"), "1").
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.PurchaseReceived) = existingPR }).Once()
		mockDB.On("Save", mock.AnythingOfType("*models.PurchaseReceived")).Return(nil).Once()
		mockDB.On("Create", mock.AnythingOfType("*handlers.ActivityEvent")).Return(nil).Once()
		mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(nil).Once()

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.PUT("/purchase-received/:id", handlers.HandlePurchaseReceivedUpdate())

		body := models.PurchaseReceived{PurchaseID: 1, InvoiceID: "INV-001-UPDATED", Price: 600.0}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("PUT", "/purchase-received/1", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Purchase Received updated successfully")
		mockDB.AssertExpectations(t)
	})

	t.Run("invalid request body - malformed JSON", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.PUT("/purchase-received/:id", handlers.HandlePurchaseReceivedUpdate())

		req, _ := http.NewRequest("PUT", "/purchase-received/1",
			bytes.NewBuffer([]byte(`{"price": "bad"}`)))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request body")
	})

	t.Run("purchase received not found", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		// Handler returns 200 with a message when the record is not found (by design).
		mockDB.On("First", mock.AnythingOfType("*models.PurchaseReceived"), "999").
			Return(gorm.ErrRecordNotFound).Once()

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.PUT("/purchase-received/:id", handlers.HandlePurchaseReceivedUpdate())

		body := models.PurchaseReceived{PurchaseID: 1, InvoiceID: "INV-001", Price: 500.0}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("PUT", "/purchase-received/999", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Purchase received not found")
		mockDB.AssertExpectations(t)
	})

	t.Run("database error on fetch", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("First", mock.AnythingOfType("*models.PurchaseReceived"), "1").
			Return(errors.New("db error")).Once()

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.PUT("/purchase-received/:id", handlers.HandlePurchaseReceivedUpdate())

		body := models.PurchaseReceived{PurchaseID: 1, InvoiceID: "INV-001", Price: 500.0}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("PUT", "/purchase-received/1", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to fetch purchase received")
		mockDB.AssertExpectations(t)
	})

	t.Run("database error on save", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		existingPR := models.PurchaseReceived{PurchaseID: 1, InvoiceID: "INV-001", Price: 500.0}

		mockDB.On("First", mock.AnythingOfType("*models.PurchaseReceived"), "1").
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.PurchaseReceived) = existingPR }).Once()
		mockDB.On("Save", mock.AnythingOfType("*models.PurchaseReceived")).
			Return(errors.New("save error")).Once()

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.PUT("/purchase-received/:id", handlers.HandlePurchaseReceivedUpdate())

		body := models.PurchaseReceived{PurchaseID: 1, InvoiceID: "INV-001", Price: 500.0}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("PUT", "/purchase-received/1", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to update purchase received")
		mockDB.AssertExpectations(t)
	})
}

// ── DeletePurchaseReceivedByID ────────────────────────────────────────────────

func TestDeletePurchaseReceivedByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("delete purchase received successfully", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		existing := models.PurchaseReceived{ID: 1}

		mockDB.On("First", mock.AnythingOfType("*models.PurchaseReceived"), "1").
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.PurchaseReceived) = existing }).Once()
		mockDB.On("Delete", mock.AnythingOfType("*models.PurchaseReceived")).Return(nil).Once()

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.DELETE("/purchase-received/:id", handlers.DeletePurchaseReceivedByID())

		req, _ := http.NewRequest("DELETE", "/purchase-received/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Purchase received deleted successfully")
		mockDB.AssertExpectations(t)
	})

	t.Run("purchase received not found", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		// Handler returns 200 with a message when the record is not found (by design).
		mockDB.On("First", mock.AnythingOfType("*models.PurchaseReceived"), "999").
			Return(gorm.ErrRecordNotFound).Once()

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.DELETE("/purchase-received/:id", handlers.DeletePurchaseReceivedByID())

		req, _ := http.NewRequest("DELETE", "/purchase-received/999", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Purchase received not found")
		mockDB.AssertExpectations(t)
	})

	t.Run("database error on fetch", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockDB.On("First", mock.AnythingOfType("*models.PurchaseReceived"), "1").
			Return(errors.New("db error")).Once()

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.DELETE("/purchase-received/:id", handlers.DeletePurchaseReceivedByID())

		req, _ := http.NewRequest("DELETE", "/purchase-received/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to fetch purchase received")
		mockDB.AssertExpectations(t)
	})

	t.Run("database error on delete", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		existing := models.PurchaseReceived{ID: 1}

		mockDB.On("First", mock.AnythingOfType("*models.PurchaseReceived"), "1").
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.PurchaseReceived) = existing }).Once()
		mockDB.On("Delete", mock.AnythingOfType("*models.PurchaseReceived")).
			Return(errors.New("delete error")).Once()

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.DELETE("/purchase-received/:id", handlers.DeletePurchaseReceivedByID())

		req, _ := http.NewRequest("DELETE", "/purchase-received/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to delete purchase received")
		mockDB.AssertExpectations(t)
	})

	t.Run("missing id - route not matched returns 404", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.DELETE("/purchase-received/:id", handlers.DeletePurchaseReceivedByID())

		req, _ := http.NewRequest("DELETE", "/purchase-received/", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestHandleCreatePurchaseReceivedWithImage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("create with image success", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &MockS3Client{}

		mockDB.On("First", mock.AnythingOfType("*models.Purchase"), mock.Anything).
			Return(nil).Once()

		mockDB.On("Create", mock.AnythingOfType("*models.PurchaseReceived")).
			Return(nil).Once()
		mockDB.On("Create", mock.AnythingOfType("*handlers.ActivityEvent")).
			Return(nil).Once()
		mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(nil).Once()

		handler := handlers.NewPurchaseReceivedHandler(mockS3, "us-east-2")

		router := gin.New()
		router.Use(injectDB(mockDB))

		router.POST("/purchase-received/create-with-image",
			handler.HandleCreatePurchaseReceivedWithImage())

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		writer.WriteField("purchase_id", "1")
		writer.WriteField("invoice_id", "INV-001")
		writer.WriteField("Price", "500")
		writer.WriteField("received_quantity", "10")

		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", `form-data; name="image"; filename="test.png"`)
		h.Set("Content-Type", "image/png")

		part, _ := writer.CreatePart(h)
		part.Write([]byte("fake-image"))

		writer.Close()

		req, _ := http.NewRequest(
			"POST",
			"/purchase-received/create-with-image",
			body,
		)

		req.Header.Set("Content-Type", writer.FormDataContentType())

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), "Purchase received created successfully")

		mockDB.AssertExpectations(t)
	})
}

func TestHandlePurchaseReceivedUpdateWithImage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("update with image success", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &MockS3Client{}

		existing := models.PurchaseReceived{
			ID:         1,
			InvoiceID:  "OLD",
			Image:      "old.png",
			PurchaseID: 1,
		}

		mockDB.On("First",
			mock.AnythingOfType("*models.PurchaseReceived"),
			"1",
		).Return(nil).
			Run(func(args mock.Arguments) {
				*args.Get(0).(*models.PurchaseReceived) = existing
			}).Once()

		mockDB.On("Save",
			mock.AnythingOfType("*models.PurchaseReceived"),
		).Return(nil).Once()
		mockDB.On("Create", mock.AnythingOfType("*handlers.ActivityEvent")).Return(nil).Once()
		mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(nil).Once()

		handler := handlers.NewPurchaseReceivedHandler(mockS3, "us-east-2")

		router := gin.New()
		router.Use(injectDB(mockDB))

		router.PUT("/purchase-received/:id/with-image",
			handler.HandlePurchaseReceivedUpdateWithImage())

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		writer.WriteField("invoice_id", "UPDATED")

		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", `form-data; name="image"; filename="test.png"`)
		h.Set("Content-Type", "image/png")

		part, _ := writer.CreatePart(h)
		part.Write([]byte("fake-image"))

		writer.Close()

		req, _ := http.NewRequest(
			"PUT",
			"/purchase-received/1/with-image",
			body,
		)

		req.Header.Set("Content-Type", writer.FormDataContentType())

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Purchase Received updated successfully")

		mockDB.AssertExpectations(t)
	})
}

func TestUploadImage_NoFile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockS3 := &MockS3Client{}
	handler := handlers.NewPurchaseReceivedHandler(mockS3, "us-east-2")

	router := gin.New()

	router.POST("/upload", func(c *gin.Context) {
		result, err := handler.UploadImage(c, "image")

		assert.Nil(t, err)
		assert.Equal(t, "", result)

		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("POST", "/upload", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleCreatePurchaseReceivedWithImage_PurchaseNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := &mocks.MockDatabase{}
	mockS3 := &MockS3Client{}

	mockDB.On("First",
		mock.AnythingOfType("*models.Purchase"),
		mock.Anything,
	).Return(gorm.ErrRecordNotFound)

	handler := handlers.NewPurchaseReceivedHandler(mockS3, "us-east-2")

	router := gin.New()
	router.Use(injectDB(mockDB))

	router.POST("/purchase-received/create-with-image",
		handler.HandleCreatePurchaseReceivedWithImage())

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	writer.WriteField("purchase_id", "999")

	writer.Close()

	req, _ := http.NewRequest(
		"POST",
		"/purchase-received/create-with-image",
		body,
	)

	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandlePurchaseReceivedUpdateWithImage_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := &mocks.MockDatabase{}
	mockS3 := &MockS3Client{}

	mockDB.On("First",
		mock.AnythingOfType("*models.PurchaseReceived"),
		"999",
	).Return(gorm.ErrRecordNotFound)

	handler := handlers.NewPurchaseReceivedHandler(mockS3, "us-east-2")

	router := gin.New()
	router.Use(injectDB(mockDB))

	router.PUT("/purchase-received/:id/with-image",
		handler.HandlePurchaseReceivedUpdateWithImage())

	reqBody := &bytes.Buffer{}
	writer := multipart.NewWriter(reqBody)
	writer.Close()

	req, _ := http.NewRequest(
		"PUT",
		"/purchase-received/999/with-image",
		reqBody,
	)

	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandlePurchaseReceivedUpdateWithImage_SaveError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := &mocks.MockDatabase{}
	mockS3 := &MockS3Client{}

	existing := models.PurchaseReceived{
		ID: 1,
	}

	mockDB.On("First",
		mock.AnythingOfType("*models.PurchaseReceived"),
		"1",
	).Return(nil).
		Run(func(args mock.Arguments) {
			*args.Get(0).(*models.PurchaseReceived) = existing
		}).Once()

	mockDB.On("Save",
		mock.AnythingOfType("*models.PurchaseReceived"),
	).Return(errors.New("save failed")).Once()

	handler := handlers.NewPurchaseReceivedHandler(mockS3, "us-east-2")

	router := gin.New()
	router.Use(injectDB(mockDB))

	router.PUT("/purchase-received/:id/with-image",
		handler.HandlePurchaseReceivedUpdateWithImage())

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.Close()

	req, _ := http.NewRequest(
		"PUT",
		"/purchase-received/1/with-image",
		body,
	)

	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleCreatePurchaseReceivedWithImage_DBError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := &mocks.MockDatabase{}
	mockS3 := &MockS3Client{}

	mockDB.On(
		"First",
		mock.AnythingOfType("*models.Purchase"),
		mock.Anything,
	).Return(nil).Once()

	mockDB.On(
		"Create",
		mock.AnythingOfType("*models.PurchaseReceived"),
	).Return(errors.New("db create failed")).Once()

	handler := handlers.NewPurchaseReceivedHandler(mockS3, "us-east-2")

	router := gin.New()
	router.Use(injectDB(mockDB))

	router.POST(
		"/purchase-received/create-with-image",
		handler.HandleCreatePurchaseReceivedWithImage(),
	)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	writer.WriteField("purchase_id", "1")
	writer.WriteField("invoice_id", "INV-001")
	writer.WriteField("Price", "500")
	writer.WriteField("received_quantity", "10")

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="image"; filename="test.png"`)
	h.Set("Content-Type", "image/png")

	part, _ := writer.CreatePart(h)
	part.Write([]byte("fake-image"))

	writer.Close()

	req, _ := http.NewRequest(
		"POST",
		"/purchase-received/create-with-image",
		body,
	)

	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to create purchase received")

	mockDB.AssertExpectations(t)
}
