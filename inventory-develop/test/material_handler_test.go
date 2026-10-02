package test

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"inventory/handlers"
	"inventory/mocks"
	"inventory/models"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// mockPresignRequest wires up the mock S3 client so that generatePresignedURL
// does not panic. The returned *request.Request has a minimal HTTPRequest so
// Presign can run; it will return an error (no real credentials), causing the
// handler to fall back to returning the raw S3 key — which is fine for tests.

func injectDB(db *mocks.MockDatabase) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	}
}

// injectTenant sets the tenant key so FileHandler.GetBucket() doesn't panic.
func injectTenant(tenant string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("tenant", tenant)
		c.Next()
	}
}

func newFileHandler(s3mock *mocks.MockS3Client) *handlers.FileHandler {
	return handlers.NewFileHandler(s3mock, "us-east-1")
}

// multipartBody builds a multipart/form-data body.
// If addImage is true a tiny fake PNG is appended under the "image" key.
func multipartBody(fields map[string]string, addImage bool) (*bytes.Buffer, string) {
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	for k, v := range fields {
		_ = w.WriteField(k, v)
	}
	if addImage {
		fw, _ := w.CreateFormFile("image", "test.png")
		fw.Write([]byte{0x89}) // minimal payload
	}
	w.Close()
	return body, w.FormDataContentType()
}

func copyFields(src map[string]string) map[string]string {
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// baseCreateFields are the minimum valid fields for HandleCreateMaterial.
var baseCreateFields = map[string]string{
	"name":             "Steel Rod",
	"category":         "Metal",
	"unit":             "kg",
	"present_location": "Rack A",
}

// ── HandleCreateMaterial ──────────────────────────────────────────────────────

func TestHandleCreateMaterial(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("valid request - creates material", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)
		mockDB.On("Model", mock.Anything).Return(mockDB)
		mockDB.On("Where", mock.Anything, mock.Anything).Return(mockDB)
		mockDB.On("Count", mock.Anything).Return(mockDB)
		mockS3.On("PutObject", mock.AnythingOfType("*s3.PutObjectInput")).
			Return(&s3.PutObjectOutput{}, nil)
		mockDB.On("Create", mock.AnythingOfType("*models.Materials")).Return(nil)

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB), injectTenant("acme"))
		router.POST("/material/create", fh.HandleCreateMaterial())

		body, ct := multipartBody(baseCreateFields, true)
		req, _ := http.NewRequest("POST", "/material/create", body)
		req.Header.Set("Content-Type", ct)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), "Material created successfully")
		mockDB.AssertExpectations(t)
		mockS3.AssertExpectations(t)
	})

	t.Run("missing required field - no name", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)
		mockDB.On("Model", mock.Anything).Return(mockDB)
		mockDB.On("Where", mock.Anything, mock.Anything).Return(mockDB)
		mockDB.On("Count", mock.Anything).Return(mockDB)
		mockS3.On("PutObject", mock.AnythingOfType("*s3.PutObjectInput")).
			Return(&s3.PutObjectOutput{}, nil).Maybe()

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB), injectTenant("acme"))
		router.POST("/material/create", fh.HandleCreateMaterial())

		fields := copyFields(baseCreateFields)
		delete(fields, "name")
		body, ct := multipartBody(fields, true)
		req, _ := http.NewRequest("POST", "/material/create", body)
		req.Header.Set("Content-Type", ct)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "required")
	})

	t.Run("database error on material insert", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		mockDB.On("Model", mock.Anything).Return(mockDB)
		mockDB.On("Where", mock.Anything, mock.Anything).Return(mockDB)
		mockDB.On("Count", mock.Anything).Return(mockDB)

		mockS3.On("PutObject", mock.AnythingOfType("*s3.PutObjectInput")).
			Return(&s3.PutObjectOutput{}, nil)
		mockDB.On("Create", mock.AnythingOfType("*models.Materials")).Return(assert.AnError)

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB), injectTenant("acme"))
		router.POST("/material/create", fh.HandleCreateMaterial())

		body, ct := multipartBody(baseCreateFields, true)
		req, _ := http.NewRequest("POST", "/material/create", body)
		req.Header.Set("Content-Type", ct)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to create material")
	})
}

// ── HandleFetchMaterials ──────────────────────────────────────────────────────

func TestHandleFetchMaterials(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("fetch all materials successfully", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		mockDB.On("Preload", "Chunks", mock.Anything).Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]models.Materials")).Return(nil)

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB))
		router.GET("/material/all", fh.HandleFetchMaterials())

		req, _ := http.NewRequest("GET", "/material/all", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("fetch materials - empty database returns zero total", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		mockDB.On("Preload", "Chunks", mock.Anything).Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]models.Materials")).Return(nil)

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB))
		router.GET("/material/all", fh.HandleFetchMaterials())

		req, _ := http.NewRequest("GET", "/material/all", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["total"])
	})

	t.Run("database query error", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		mockDB.On("Preload", "Chunks", mock.Anything).Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]models.Materials")).Return(assert.AnError)

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB))
		router.GET("/material/all", fh.HandleFetchMaterials())

		req, _ := http.NewRequest("GET", "/material/all", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to fetch materials")
	})

	t.Run("fetch with date range query parameters", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		mockDB.On("Where", "created_at BETWEEN ? AND ?", mock.Anything, mock.Anything).Return(mockDB)
		mockDB.On("Preload", "Chunks", mock.Anything).Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]models.Materials")).Return(nil)

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB))
		router.GET("/material/all", fh.HandleFetchMaterials())

		req, _ := http.NewRequest("GET", "/material/all?start_date=2024-01-01&end_date=2024-12-31", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("backfills missing unit from material type", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		mockDB.On("Preload", "Chunks", mock.Anything).Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]models.Materials")).
			Return(nil).
			Run(func(args mock.Arguments) {
				m := args.Get(0).(*[]models.Materials)
				*m = []models.Materials{
					{ID: 1, Name: "Soap Stone", Type: "slab", Unit: ""},
					{ID: 2, Name: "Granite", Type: "slab", Unit: "sqft"},
				}
			}).Once()

		mockDB.On("Where", "name IN ?", mock.Anything).Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]models.MaterialType")).
			Return(nil).
			Run(func(args mock.Arguments) {
				mt := args.Get(0).(*[]models.MaterialType)
				*mt = []models.MaterialType{{Name: "slab", Unit: "sqft"}}
			}).Once()

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB))
		router.GET("/material/all", fh.HandleFetchMaterials())

		req, _ := http.NewRequest("GET", "/material/all", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		data := resp["data"].([]interface{})
		first := data[0].(map[string]interface{})
		assert.Equal(t, "sqft", first["Unit"])
		mockDB.AssertExpectations(t)
	})

	t.Run("no backfill query when all units already present", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		mockDB.On("Preload", "Chunks", mock.Anything).Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]models.Materials")).
			Return(nil).
			Run(func(args mock.Arguments) {
				m := args.Get(0).(*[]models.Materials)
				*m = []models.Materials{{ID: 1, Name: "Granite", Type: "slab", Unit: "sqft"}}
			}).Once()

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB))
		router.GET("/material/all", fh.HandleFetchMaterials())

		req, _ := http.NewRequest("GET", "/material/all", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		// "Where"/second "Find" must never be called since nothing is missing a unit.
		mockDB.AssertNotCalled(t, "Where", "name IN ?", mock.Anything)
		mockDB.AssertExpectations(t)
	})
}

// ── HandleMaterialGetByID ─────────────────────────────────────────────────────

func TestHandleMaterialGetByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("get material by id successfully", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		mockDB.On("Preload", "Chunks").Return(mockDB)
		mockDB.On("First", mock.AnythingOfType("*models.Materials"), "1").
			Return(nil).
			Run(func(args mock.Arguments) {
				m := args.Get(0).(*models.Materials)
				m.ID = 1
				m.Name = "Steel Rod"
			}).Once()

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB))
		router.GET("/material/:id", fh.HandleMaterialGetByID())

		req, _ := http.NewRequest("GET", "/material/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Steel Rod")
		mockDB.AssertExpectations(t)
	})

	t.Run("missing id - route not matched returns 404", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB))
		router.GET("/material/:id", fh.HandleMaterialGetByID())

		req, _ := http.NewRequest("GET", "/material/", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("get by id backfills missing unit from material type", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		mockDB.On("Preload", "Chunks").Return(mockDB)
		mockDB.On("First", mock.AnythingOfType("*models.Materials"), "1").
			Return(nil).
			Run(func(args mock.Arguments) {
				m := args.Get(0).(*models.Materials)
				m.ID = 1
				m.Name = "Soap Stone"
				m.Type = "slab"
				m.Unit = ""
			}).Once()

		mockDB.On("Where", "name = ?", "slab").Return(mockDB)
		mockDB.On("First", mock.AnythingOfType("*models.MaterialType")).
			Return(nil).
			Run(func(args mock.Arguments) {
				mt := args.Get(0).(*models.MaterialType)
				mt.Name = "slab"
				mt.Unit = "sqft"
			}).Once()

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB))
		router.GET("/material/:id", fh.HandleMaterialGetByID())

		req, _ := http.NewRequest("GET", "/material/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "sqft", resp["Unit"])
		mockDB.AssertExpectations(t)
	})

	t.Run("material not found", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		mockDB.On("Preload", "Chunks").Return(mockDB)
		mockDB.On("First", mock.AnythingOfType("*models.Materials"), "99999").
			Return(gorm.ErrRecordNotFound).Once()

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB))
		router.GET("/material/:id", fh.HandleMaterialGetByID())

		req, _ := http.NewRequest("GET", "/material/99999", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Material not found")
		mockDB.AssertExpectations(t)
	})

	t.Run("database error on fetch", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		mockDB.On("Preload", "Chunks").Return(mockDB)
		mockDB.On("First", mock.AnythingOfType("*models.Materials"), "1").
			Return(assert.AnError).Once()

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB))
		router.GET("/material/:id", fh.HandleMaterialGetByID())

		req, _ := http.NewRequest("GET", "/material/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to fetch material")
		mockDB.AssertExpectations(t)
	})
}

// ── HandleMaterialUpdate ──────────────────────────────────────────────────────

func TestHandleMaterialUpdate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("update text fields - no image", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		existing := models.Materials{ID: 1, Name: "Steel Rod", Category: "Metal"}

		mockDB.On("Preload", "Chunks").Return(mockDB)
		mockDB.On("First", mock.AnythingOfType("*models.Materials"), "1").
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.Materials) = existing }).Once()
		mockDB.On("Save", mock.AnythingOfType("*models.Materials")).Return(nil).Once()

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB), injectTenant("acme"))
		router.PUT("/material/:id", fh.HandleMaterialUpdate())

		body, ct := multipartBody(map[string]string{"name": "Iron Bar", "category": "Metal"}, false)
		req, _ := http.NewRequest("PUT", "/material/1", body)
		req.Header.Set("Content-Type", ct)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("material not found", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		mockDB.On("Preload", "Chunks").Return(mockDB)
		mockDB.On("First", mock.AnythingOfType("*models.Materials"), "999").
			Return(gorm.ErrRecordNotFound).Once()

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB), injectTenant("acme"))
		router.PUT("/material/:id", fh.HandleMaterialUpdate())

		body, ct := multipartBody(map[string]string{"name": "X"}, false)
		req, _ := http.NewRequest("PUT", "/material/999", body)
		req.Header.Set("Content-Type", ct)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Material not found")
		mockDB.AssertExpectations(t)
	})

	t.Run("database error on fetch", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		mockDB.On("Preload", "Chunks").Return(mockDB)
		mockDB.On("First", mock.AnythingOfType("*models.Materials"), "1").
			Return(assert.AnError).Once()

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB), injectTenant("acme"))
		router.PUT("/material/:id", fh.HandleMaterialUpdate())

		body, ct := multipartBody(map[string]string{"name": "X"}, false)
		req, _ := http.NewRequest("PUT", "/material/1", body)
		req.Header.Set("Content-Type", ct)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Database error")
		mockDB.AssertExpectations(t)
	})

	t.Run("database error on save", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		existing := models.Materials{ID: 1, Name: "Steel Rod"}

		mockDB.On("Preload", "Chunks").Return(mockDB)
		mockDB.On("First", mock.AnythingOfType("*models.Materials"), "1").
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.Materials) = existing }).Once()
		mockDB.On("Save", mock.AnythingOfType("*models.Materials")).Return(assert.AnError).Once()

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB), injectTenant("acme"))
		router.PUT("/material/:id", fh.HandleMaterialUpdate())

		body, ct := multipartBody(map[string]string{"name": "Updated"}, false)
		req, _ := http.NewRequest("PUT", "/material/1", body)
		req.Header.Set("Content-Type", ct)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to update material")
		mockDB.AssertExpectations(t)
	})

	t.Run("invalid available value", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		existing := models.Materials{ID: 1, Name: "Steel Rod"}

		mockDB.On("Preload", "Chunks").Return(mockDB)
		mockDB.On("First", mock.AnythingOfType("*models.Materials"), "1").
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.Materials) = existing }).Once()

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB), injectTenant("acme"))
		router.PUT("/material/:id", fh.HandleMaterialUpdate())

		body, ct := multipartBody(map[string]string{"available": "bad"}, false)
		req, _ := http.NewRequest("PUT", "/material/1", body)
		req.Header.Set("Content-Type", ct)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid available value")
	})

	t.Run("invalid square feet value", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		existing := models.Materials{ID: 1, Name: "Steel Rod"}

		mockDB.On("Preload", "Chunks").Return(mockDB)
		mockDB.On("First", mock.AnythingOfType("*models.Materials"), "1").
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.Materials) = existing }).Once()

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB), injectTenant("acme"))
		router.PUT("/material/:id", fh.HandleMaterialUpdate())

		body, ct := multipartBody(map[string]string{"square_feet": "bad"}, false)
		req, _ := http.NewRequest("PUT", "/material/1", body)
		req.Header.Set("Content-Type", ct)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid square_feet value")
	})

	t.Run("invalid retail cost", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		existing := models.Materials{ID: 1, Name: "Steel Rod"}

		mockDB.On("Preload", "Chunks").Return(mockDB)
		mockDB.On("First", mock.AnythingOfType("*models.Materials"), "1").
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.Materials) = existing }).Once()

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB), injectTenant("acme"))
		router.PUT("/material/:id", fh.HandleMaterialUpdate())

		body, ct := multipartBody(map[string]string{"retail_cost": "bad"}, false)
		req, _ := http.NewRequest("PUT", "/material/1", body)
		req.Header.Set("Content-Type", ct)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid retail_cost value")
		mockDB.AssertExpectations(t)
	})

	t.Run("invalid threshold value returns 400", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		existing := models.Materials{ID: 1, Name: "Steel Rod"}

		mockDB.On("Preload", "Chunks").Return(mockDB)
		mockDB.On("First", mock.AnythingOfType("*models.Materials"), "1").
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.Materials) = existing }).Once()

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB), injectTenant("acme"))
		router.PUT("/material/:id", fh.HandleMaterialUpdate())

		body, ct := multipartBody(map[string]string{"threshold": "bad"}, false)
		req, _ := http.NewRequest("PUT", "/material/1", body)
		req.Header.Set("Content-Type", ct)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid threshold value")
		mockDB.AssertExpectations(t)
	})

	t.Run("invalid procost value returns 400", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		existing := models.Materials{ID: 1, Name: "Steel Rod"}

		mockDB.On("Preload", "Chunks").Return(mockDB)
		mockDB.On("First", mock.AnythingOfType("*models.Materials"), "1").
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.Materials) = existing }).Once()

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB), injectTenant("acme"))
		router.PUT("/material/:id", fh.HandleMaterialUpdate())

		body, ct := multipartBody(map[string]string{"procost": "bad"}, false)
		req, _ := http.NewRequest("PUT", "/material/1", body)
		req.Header.Set("Content-Type", ct)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid procost value")
		mockDB.AssertExpectations(t)
	})

	t.Run("increase available - new chunks are created (non-procurement)", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		existing := models.Materials{
			ID:        1,
			Name:      "Steel Rod",
			Serialnum: "SR-001",
			Available: 2,
			Chunks: []models.Materialchunks{
				{ID: 1, Serialnum: "SR-001-1"},
				{ID: 2, Serialnum: "SR-001-2"},
			},
		}

		mockDB.On("Preload", "Chunks").Return(mockDB)
		mockDB.On("First", mock.AnythingOfType("*models.Materials"), "1").
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.Materials) = existing }).Once()
		mockDB.On("Save", mock.AnythingOfType("*models.Materials")).Return(nil).Once()
		// maxChunkSuffix calls Where("material_id = ?", materialID).Find(&chunks)
		mockDB.On("Where", "material_id = ?", uint(1)).Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]models.Materialchunks")).Return(nil).Once()
		mockDB.On("Create", mock.AnythingOfType("*[]models.Materialchunks")).Return(nil).Once()

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB), injectTenant("acme"))
		router.PUT("/material/:id", fh.HandleMaterialUpdate())

		body, ct := multipartBody(map[string]string{"available": "4"}, false)
		req, _ := http.NewRequest("PUT", "/material/1", body)
		req.Header.Set("Content-Type", ct)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("procurement flag - new chunks created with lot number", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		existing := models.Materials{
			ID:        1,
			Name:      "Steel Rod",
			Serialnum: "SR-001",
			Available: 2,
			Chunks: []models.Materialchunks{
				{ID: 1, Serialnum: "SR-001-1"},
				{ID: 2, Serialnum: "SR-001-2"},
			},
		}
		mockDB.On("Model", mock.Anything).Return(mockDB)
		mockDB.On("Where", mock.Anything, mock.Anything).Return(mockDB)
		mockDB.On("Count", mock.Anything).Return(mockDB)
		mockDB.On("Preload", "Chunks").Return(mockDB)
		mockDB.On("First", mock.AnythingOfType("*models.Materials"), "1").
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.Materials) = existing }).Once()
		mockDB.On("Save", mock.AnythingOfType("*models.Materials")).Return(nil).Once()
		mockDB.On("Select", mock.Anything).Return(mockDB).Maybe()
		mockDB.On("Order", mock.Anything).Return(mockDB).Maybe()
		mockDB.On("Last", mock.Anything).Return(nil).Maybe()
		// maxChunkSuffix: Where("material_id = ?", materialID).Find(&chunks)
		mockDB.On("Find", mock.AnythingOfType("*[]models.Materialchunks")).Return(nil).Once()
		mockDB.On("Create", mock.AnythingOfType("*[]models.Materialchunks")).Return(nil).Once()

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB), injectTenant("acme"))
		router.PUT("/material/:id", fh.HandleMaterialUpdate())

		body, ct := multipartBody(map[string]string{
			"available":   "4",
			"procurement": "true",
			"lot":         "3",
		}, false)
		req, _ := http.NewRequest("PUT", "/material/1", body)
		req.Header.Set("Content-Type", ct)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("procurement - create chunks db error returns 500", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		existing := models.Materials{
			ID:        1,
			Serialnum: "SR-001",
			Available: 0,
			Chunks:    []models.Materialchunks{},
		}

		mockDB.On("Model", mock.Anything).Return(mockDB)
		mockDB.On("Where", mock.Anything, mock.Anything).Return(mockDB)
		mockDB.On("Count", mock.Anything).Return(mockDB)
		mockDB.On("Select", mock.Anything).Return(mockDB).Maybe()
		mockDB.On("Order", mock.Anything).Return(mockDB).Maybe()
		mockDB.On("Last", mock.Anything).Return(nil).Maybe()
		mockDB.On("Preload", "Chunks").Return(mockDB)
		mockDB.On("First", mock.AnythingOfType("*models.Materials"), "1").
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.Materials) = existing }).Once()
		mockDB.On("Save", mock.AnythingOfType("*models.Materials")).Return(nil).Once()
		mockDB.On("Find", mock.AnythingOfType("*[]models.Materialchunks")).Return(nil).Once()
		mockDB.On("Create", mock.AnythingOfType("*[]models.Materialchunks")).Return(assert.AnError).Once()

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB), injectTenant("acme"))
		router.PUT("/material/:id", fh.HandleMaterialUpdate())

		body, ct := multipartBody(map[string]string{
			"available":   "2",
			"procurement": "true",
			"lot":         "L1",
		}, false)
		req, _ := http.NewRequest("PUT", "/material/1", body)
		req.Header.Set("Content-Type", ct)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to create new chunks")
		mockDB.AssertExpectations(t)
	})

	t.Run("non-procurement - create chunks db error returns 500", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		existing := models.Materials{
			ID:        1,
			Serialnum: "SR-001",
			Available: 0,
			Chunks:    []models.Materialchunks{},
		}

		mockDB.On("Preload", "Chunks").Return(mockDB)
		mockDB.On("First", mock.AnythingOfType("*models.Materials"), "1").
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.Materials) = existing }).Once()
		mockDB.On("Save", mock.AnythingOfType("*models.Materials")).Return(nil).Once()
		mockDB.On("Where", "material_id = ?", uint(1)).Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]models.Materialchunks")).Return(nil).Once()
		mockDB.On("Create", mock.AnythingOfType("*[]models.Materialchunks")).Return(assert.AnError).Once()

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB), injectTenant("acme"))
		router.PUT("/material/:id", fh.HandleMaterialUpdate())

		body, ct := multipartBody(map[string]string{"available": "3"}, false)
		req, _ := http.NewRequest("PUT", "/material/1", body)
		req.Header.Set("Content-Type", ct)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to create new chunks")
		mockDB.AssertExpectations(t)
	})
}

// ── HandleMaterialChunkUpdate ─────────────────────────────────────────────────

func TestHandleMaterialChunkUpdate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("update chunk location successfully", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		existing := models.Materialchunks{ID: 1, Serialnum: "SR-001-1", Dimensions: ""}

		mockDB.On("First", mock.AnythingOfType("*models.Materialchunks"), "1").
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.Materialchunks) = existing }).Once()
		mockDB.On("Save", mock.AnythingOfType("*models.Materialchunks")).Return(nil).Once()

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB), injectTenant("acme"))
		router.PUT("/chunk/:id", fh.HandleMaterialChunkUpdate())

		body, ct := multipartBody(map[string]string{"present_location": "Rack B"}, false)
		req, _ := http.NewRequest("PUT", "/chunk/1", body)
		req.Header.Set("Content-Type", ct)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("chunk not found", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		mockDB.On("First", mock.AnythingOfType("*models.Materialchunks"), "99").
			Return(gorm.ErrRecordNotFound).Once()

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB), injectTenant("acme"))
		router.PUT("/chunk/:id", fh.HandleMaterialChunkUpdate())

		body, ct := multipartBody(map[string]string{"present_location": "Rack B"}, false)
		req, _ := http.NewRequest("PUT", "/chunk/99", body)
		req.Header.Set("Content-Type", ct)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Chunk not found")
		mockDB.AssertExpectations(t)
	})

	t.Run("database error on fetch", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		mockDB.On("First", mock.AnythingOfType("*models.Materialchunks"), "1").
			Return(assert.AnError).Once()

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB), injectTenant("acme"))
		router.PUT("/chunk/:id", fh.HandleMaterialChunkUpdate())

		body, ct := multipartBody(map[string]string{"present_location": "X"}, false)
		req, _ := http.NewRequest("PUT", "/chunk/1", body)
		req.Header.Set("Content-Type", ct)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Database error")
		mockDB.AssertExpectations(t)
	})

	t.Run("dimension change creates revision chunk", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		existing := models.Materialchunks{
			ID:         1,
			MaterialID: 10,
			Serialnum:  "SR-001-1",
			Dimensions: "100x200",
		}

		mockDB.On("First", mock.AnythingOfType("*models.Materialchunks"), "1").
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.Materialchunks) = existing }).Once()

		// revision count: Model().Where("material_id = ? AND serialnum LIKE ?", ...).Count(...)
		mockDB.On("Model", mock.AnythingOfType("*models.Materialchunks")).Return(mockDB)
		mockDB.On("Where", mock.Anything, mock.Anything, mock.Anything).Return(mockDB)
		mockDB.On("Count", mock.AnythingOfType("*int64")).
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*int64) = 0 }).Once()

		mockDB.On("Create", mock.AnythingOfType("*models.Materialchunks")).Return(nil).Once()
		// original chunk saved (serialnum has no -r so status stays as-is)
		mockDB.On("Save", mock.AnythingOfType("*models.Materialchunks")).Return(nil).Once()

		// available count sync: Where("material_id = ?", ...).Find(...)
		mockDB.On("Where", mock.Anything, mock.Anything).Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]models.Materialchunks")).Return(nil).Once()

		// material available update: First(...).Save(...)
		mockDB.On("First", mock.AnythingOfType("*models.Materials"), mock.Anything).Return(nil).Once()
		mockDB.On("Save", mock.AnythingOfType("*models.Materials")).Return(nil).Once()

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB), injectTenant("acme"))
		router.PUT("/chunk/:id", fh.HandleMaterialChunkUpdate())

		body, ct := multipartBody(map[string]string{"dimensions": "100x150"}, false)
		req, _ := http.NewRequest("PUT", "/chunk/1", body)
		req.Header.Set("Content-Type", ct)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NotNil(t, resp["revision"])
		assert.Contains(t, resp["message"], "SR-001-1-r1")
		mockDB.AssertExpectations(t)
	})

	t.Run("same dimension value - no revision created", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		existing := models.Materialchunks{ID: 1, Serialnum: "SR-001-1", Dimensions: "100x200"}

		mockDB.On("First", mock.AnythingOfType("*models.Materialchunks"), "1").
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.Materialchunks) = existing }).Once()
		// same dimensions → isRemnantRequest is false → plain Save
		mockDB.On("Save", mock.AnythingOfType("*models.Materialchunks")).Return(nil).Once()

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB), injectTenant("acme"))
		router.PUT("/chunk/:id", fh.HandleMaterialChunkUpdate())

		body, ct := multipartBody(map[string]string{"dimensions": "100x200"}, false)
		req, _ := http.NewRequest("PUT", "/chunk/1", body)
		req.Header.Set("Content-Type", ct)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockDB.AssertExpectations(t)
	})

	// "badformat" != "100x200" and no_revision is false → isRemnantRequest=true.
	// The handler does NOT validate dimension format before creating the revision;
	// it creates the new chunk with whatever string was passed, then returns 200.
	t.Run("invalid dimension format still creates revision and returns 200", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		existing := models.Materialchunks{
			ID:         1,
			MaterialID: 10,
			Serialnum:  "SR-001-1",
			Dimensions: "100x200",
		}

		mockDB.On("First", mock.AnythingOfType("*models.Materialchunks"), "1").
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.Materialchunks) = existing }).Once()

		mockDB.On("Model", mock.AnythingOfType("*models.Materialchunks")).Return(mockDB)
		mockDB.On("Where", mock.Anything, mock.Anything, mock.Anything).Return(mockDB)
		mockDB.On("Count", mock.AnythingOfType("*int64")).
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*int64) = 0 }).Once()

		mockDB.On("Create", mock.AnythingOfType("*models.Materialchunks")).Return(nil).Once()
		mockDB.On("Save", mock.AnythingOfType("*models.Materialchunks")).Return(nil).Once()

		mockDB.On("Where", mock.Anything, mock.Anything).Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]models.Materialchunks")).Return(nil).Once()

		mockDB.On("First", mock.AnythingOfType("*models.Materials"), mock.Anything).Return(nil).Once()
		mockDB.On("Save", mock.AnythingOfType("*models.Materials")).Return(nil).Once()

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB), injectTenant("acme"))
		router.PUT("/chunk/:id", fh.HandleMaterialChunkUpdate())

		body, ct := multipartBody(map[string]string{"dimensions": "badformat"}, false)
		req, _ := http.NewRequest("PUT", "/chunk/1", body)
		req.Header.Set("Content-Type", ct)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NotNil(t, resp["revision"])
		mockDB.AssertExpectations(t)
	})

	t.Run("database error on save", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		existing := models.Materialchunks{ID: 1, Serialnum: "SR-001-1", Dimensions: ""}

		mockDB.On("First", mock.AnythingOfType("*models.Materialchunks"), "1").
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.Materialchunks) = existing }).Once()
		mockDB.On("Save", mock.AnythingOfType("*models.Materialchunks")).Return(assert.AnError).Once()

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB), injectTenant("acme"))
		router.PUT("/chunk/:id", fh.HandleMaterialChunkUpdate())

		body, ct := multipartBody(map[string]string{"present_location": "Rack C"}, false)
		req, _ := http.NewRequest("PUT", "/chunk/1", body)
		req.Header.Set("Content-Type", ct)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to update chunk")
	})

	t.Run("missing id - route not matched returns 404", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB), injectTenant("acme"))
		router.PUT("/chunk/:id", fh.HandleMaterialChunkUpdate())

		body, ct := multipartBody(map[string]string{}, false)
		req, _ := http.NewRequest("PUT", "/chunk/", body)
		req.Header.Set("Content-Type", ct)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("no_revision flag skips remnant flow and saves directly", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		existing := models.Materialchunks{ID: 1, Serialnum: "SR-001-1", Dimensions: "100x200"}

		mockDB.On("First", mock.AnythingOfType("*models.Materialchunks"), "1").
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.Materialchunks) = existing }).Once()
		mockDB.On("Save", mock.AnythingOfType("*models.Materialchunks")).Return(nil).Once()

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB), injectTenant("acme"))
		router.PUT("/chunk/:id", fh.HandleMaterialChunkUpdate())

		body, ct := multipartBody(map[string]string{
			"dimensions":  "50x50",
			"no_revision": "true",
		}, false)
		req, _ := http.NewRequest("PUT", "/chunk/1", body)
		req.Header.Set("Content-Type", ct)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("assigned chunk remnant preserves original status and creates available remnant", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		existing := models.Materialchunks{
			ID:         2,
			MaterialID: 10,
			Serialnum:  "SR-001-1",
			Dimensions: "100x200",
			Status:     "assigned",
		}

		mockDB.On("First", mock.AnythingOfType("*models.Materialchunks"), "2").
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.Materialchunks) = existing }).Once()

		mockDB.On("Model", mock.AnythingOfType("*models.Materialchunks")).Return(mockDB)
		mockDB.On("Where", "material_id = ? AND serialnum LIKE ?", mock.Anything, mock.Anything).Return(mockDB)
		mockDB.On("Order", "serialnum DESC").Return(mockDB)
		mockDB.On("Limit", 1).Return(mockDB)
		mockDB.On("Pluck", "serialnum", mock.Anything).Return(nil).Once()

		mockDB.On("Create", mock.AnythingOfType("*models.Materialchunks")).Return(nil).Once()

		// This is for the material.available update
		mockDB.On("Exec", "UPDATE materials SET available = available + 1, updated_at = ? WHERE id = ?", mock.Anything, mock.Anything).Return(nil).Once()

		// Mock the picture enrichment at the end
		mockDB.On("Where", "id = ?", mock.Anything).Return(mockDB).Maybe()
		mockDB.On("Find", mock.Anything).Return(nil).Maybe()
		mockDB.On("First", mock.Anything, mock.Anything).Return(nil).Maybe()

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB), injectTenant("acme"))
		router.PUT("/chunk/:id", fh.HandleMaterialChunkUpdate())

		body, ct := multipartBody(map[string]string{"dimensions": "80x100"}, false)
		req, _ := http.NewRequest("PUT", "/chunk/2", body)
		req.Header.Set("Content-Type", ct)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NotNil(t, resp["remnant_chunk"])
		originalChunk := resp["original_chunk"].(map[string]interface{})
		remnantChunk := resp["remnant_chunk"].(map[string]interface{})

		assert.Equal(t, "assigned", originalChunk["status"])
		assert.Equal(t, "available", remnantChunk["status"])
		assert.Contains(t, remnantChunk["serialnum"], "-r1")
		mockDB.AssertExpectations(t)
	})

	t.Run("create revision chunk db error returns 500", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		existing := models.Materialchunks{
			ID:         1,
			MaterialID: 10,
			Serialnum:  "SR-001-1",
			Dimensions: "100x200",
		}

		mockDB.On("First", mock.AnythingOfType("*models.Materialchunks"), "1").
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.Materialchunks) = existing }).Once()

		mockDB.On("Model", mock.AnythingOfType("*models.Materialchunks")).Return(mockDB)
		mockDB.On("Where", mock.Anything, mock.Anything, mock.Anything).Return(mockDB)
		mockDB.On("Count", mock.AnythingOfType("*int64")).
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*int64) = 0 }).Once()

		mockDB.On("Create", mock.AnythingOfType("*models.Materialchunks")).Return(assert.AnError).Once()

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB), injectTenant("acme"))
		router.PUT("/chunk/:id", fh.HandleMaterialChunkUpdate())

		body, ct := multipartBody(map[string]string{"dimensions": "100x150"}, false)
		req, _ := http.NewRequest("PUT", "/chunk/1", body)
		req.Header.Set("Content-Type", ct)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to create revision chunk")
		mockDB.AssertExpectations(t)
	})
}

// ── HandleChunkStatusUpdate ───────────────────────────────────────────────────
//
// The handler uses database.Exec(...) for all writes — never Save.
// Valid statuses are "available" and "assigned" only.

func TestHandleChunkStatusUpdate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jsonBody := func(status string) *bytes.Buffer {
		b, _ := json.Marshal(map[string]string{"status": status})
		return bytes.NewBuffer(b)
	}

	t.Run("update chunk status to assigned successfully", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		existing := models.Materialchunks{ID: 1, MaterialID: 10, Status: "available"}

		mockDB.On("First", mock.AnythingOfType("*models.Materialchunks"), "1").
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.Materialchunks) = existing }).Once()
		// UPDATE materialchunks SET status=?, job_id=?, updated_at=? WHERE id=?
		mockDB.On("Exec",
			"UPDATE materialchunks SET status = ?, job_id = ?, updated_at = ? WHERE id = ?",
			mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		).Return(nil).Once()
		// Fetch all chunks for this material to recalculate available count
		// First query used for remnant/revision count
		mockDB.On(
			"Where",
			"material_id = ? AND (serialnum = ? OR serialnum LIKE ?)",
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(mockDB).Maybe()

		// Second query used for available count recalculation
		mockDB.On(
			"Where",
			"material_id = ?",
			uint(10),
		).Return(mockDB)

		// Fetch chunks
		mockDB.On(
			"Find",
			mock.AnythingOfType("*[]models.Materialchunks"),
		).Return(nil)
		// UPDATE materials SET available=?, updated_at=? WHERE id=?
		mockDB.On("Exec",
			"UPDATE materials SET available = ?, updated_at = ? WHERE id = ?",
			mock.Anything, mock.Anything, mock.Anything,
		).Return(nil).Once()
		// Final fetch of updated chunk
		mockDB.On("First", mock.AnythingOfType("*models.Materialchunks"), "1").
			Return(nil).Once()

		router := gin.New()
		fh := handlers.NewFileHandler(mockS3, "us-east-1")
		router.Use(injectDB(mockDB))
		router.PATCH("/chunk/:id/status", fh.HandleChunkStatusUpdate())

		req, _ := http.NewRequest("PATCH", "/chunk/1/status", jsonBody("assigned"))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("update chunk status to available successfully", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		existing := models.Materialchunks{ID: 2, MaterialID: 10, Status: "assigned"}

		mockDB.On("First", mock.AnythingOfType("*models.Materialchunks"), "2").
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.Materialchunks) = existing }).Once()
		mockDB.On("Exec",
			"UPDATE materialchunks SET status = ?, job_id = ?, updated_at = ? WHERE id = ?",
			mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		).Return(nil).Once()
		// First query used for remnant/revision count
		mockDB.On(
			"Where",
			"material_id = ? AND (serialnum = ? OR serialnum LIKE ?)",
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(mockDB).Maybe()

		// Second query used for available count recalculation
		mockDB.On(
			"Where",
			"material_id = ?",
			uint(10),
		).Return(mockDB)

		// Fetch chunks
		mockDB.On(
			"Find",
			mock.AnythingOfType("*[]models.Materialchunks"),
		).Return(nil)
		mockDB.On("Exec",
			"UPDATE materials SET available = ?, updated_at = ? WHERE id = ?",
			mock.Anything, mock.Anything, mock.Anything,
		).Return(nil).Once()
		mockDB.On("First", mock.AnythingOfType("*models.Materialchunks"), "2").
			Return(nil).Once()

		router := gin.New()
		fh := handlers.NewFileHandler(mockS3, "us-east-1")
		router.Use(injectDB(mockDB))
		router.PATCH("/chunk/:id/status", fh.HandleChunkStatusUpdate())

		req, _ := http.NewRequest("PATCH", "/chunk/2/status", jsonBody("available"))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("missing status returns 400", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		router := gin.New()
		fh := handlers.NewFileHandler(mockS3, "us-east-1")
		router.Use(injectDB(mockDB))
		router.PATCH("/chunk/:id/status", fh.HandleChunkStatusUpdate())

		req, _ := http.NewRequest("PATCH", "/chunk/1/status", bytes.NewBufferString(`{}`))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "status is required")
	})

	// Handler only accepts "available" or "assigned" — anything else is 400.
	t.Run("invalid status value returns 400", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		router := gin.New()
		fh := handlers.NewFileHandler(mockS3, "us-east-1")
		router.Use(injectDB(mockDB))
		router.PATCH("/chunk/:id/status", fh.HandleChunkStatusUpdate())

		req, _ := http.NewRequest("PATCH", "/chunk/1/status", jsonBody("sold"))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "status must be")
	})

	t.Run("chunk not found returns 404", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		mockDB.On("First", mock.AnythingOfType("*models.Materialchunks"), "99").
			Return(gorm.ErrRecordNotFound).Once()

		router := gin.New()
		fh := handlers.NewFileHandler(mockS3, "us-east-1")
		router.Use(injectDB(mockDB))
		router.PATCH("/chunk/:id/status", fh.HandleChunkStatusUpdate())

		req, _ := http.NewRequest("PATCH", "/chunk/99/status", jsonBody("assigned"))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Chunk not found")
		mockDB.AssertExpectations(t)
	})

	t.Run("database error on fetch returns 500", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		mockDB.On("First", mock.AnythingOfType("*models.Materialchunks"), "1").
			Return(assert.AnError).Once()

		router := gin.New()
		fh := handlers.NewFileHandler(mockS3, "us-east-1")
		router.Use(injectDB(mockDB))
		router.PATCH("/chunk/:id/status", fh.HandleChunkStatusUpdate())

		req, _ := http.NewRequest("PATCH", "/chunk/1/status", jsonBody("assigned"))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Database error")
		mockDB.AssertExpectations(t)
	})

	t.Run("exec error on status update returns 500", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		existing := models.Materialchunks{ID: 1, MaterialID: 10, Status: "available"}

		mockDB.On("First", mock.AnythingOfType("*models.Materialchunks"), "1").
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.Materialchunks) = existing }).Once()
		mockDB.On("Exec",
			"UPDATE materialchunks SET status = ?, job_id = ?, updated_at = ? WHERE id = ?",
			mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		).Return(assert.AnError).Once()

		router := gin.New()
		fh := handlers.NewFileHandler(mockS3, "us-east-1")
		router.Use(injectDB(mockDB))
		router.PATCH("/chunk/:id/status", fh.HandleChunkStatusUpdate())

		req, _ := http.NewRequest("PATCH", "/chunk/1/status", jsonBody("assigned"))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to update chunk status")
		mockDB.AssertExpectations(t)
	})
}

// ── DeleteMaterialByID ────────────────────────────────────────────────────────

func TestDeleteMaterialByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("delete material and its chunks successfully", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		existing := models.Materials{
			ID:   1,
			Name: "Steel Rod",
			Chunks: []models.Materialchunks{
				{ID: 1, Serialnum: "SR-001-1"},
			},
		}

		mockDB.On("Preload", "Chunks").Return(mockDB)
		mockDB.On("First", mock.AnythingOfType("*models.Materials"), "1").
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.Materials) = existing }).Once()

		mockDB.On("Where", "material_id = ?", mock.Anything).Return(mockDB)
		mockDB.On("Delete", mock.AnythingOfType("*models.Materialchunks")).Return(nil).Once()
		mockDB.On("Delete", mock.AnythingOfType("*models.Materials")).Return(nil).Once()

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB))
		router.DELETE("/material/:id", fh.DeleteMaterialByID())

		req, _ := http.NewRequest("DELETE", "/material/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Material deleted successfully")
		mockDB.AssertExpectations(t)
	})

	t.Run("material not found", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		mockDB.On("Preload", "Chunks").Return(mockDB)
		mockDB.On("First", mock.AnythingOfType("*models.Materials"), "999").
			Return(gorm.ErrRecordNotFound).Once()

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB))
		router.DELETE("/material/:id", fh.DeleteMaterialByID())

		req, _ := http.NewRequest("DELETE", "/material/999", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Material not found")
		mockDB.AssertExpectations(t)
	})

	t.Run("database error on fetch", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		mockDB.On("Preload", "Chunks").Return(mockDB)
		mockDB.On("First", mock.AnythingOfType("*models.Materials"), "1").
			Return(assert.AnError).Once()

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB))
		router.DELETE("/material/:id", fh.DeleteMaterialByID())

		req, _ := http.NewRequest("DELETE", "/material/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to fetch material")
		mockDB.AssertExpectations(t)
	})

	t.Run("database error on delete material", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		existing := models.Materials{ID: 1}

		mockDB.On("Preload", "Chunks").Return(mockDB)
		mockDB.On("First", mock.AnythingOfType("*models.Materials"), "1").
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.Materials) = existing }).Once()
		mockDB.On("Delete", mock.AnythingOfType("*models.Materials")).Return(assert.AnError).Once()

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB))
		router.DELETE("/material/:id", fh.DeleteMaterialByID())

		req, _ := http.NewRequest("DELETE", "/material/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to delete material")
		mockDB.AssertExpectations(t)
	})

	t.Run("missing id - route not matched returns 404", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB))
		router.DELETE("/material/:id", fh.DeleteMaterialByID())

		req, _ := http.NewRequest("DELETE", "/material/", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

// ── HandleDeleteMaterialImage ─────────────────────────────────────────────────

func TestHandleDeleteMaterialImage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jsonBody := func(imageURL string) *bytes.Buffer {
		b, _ := json.Marshal(map[string]string{"image_url": imageURL})
		return bytes.NewBuffer(b)
	}

	t.Run("deletes image successfully", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		existing := models.Materials{
			ID:       1,
			Pictures: pq.StringArray{"https://bucket.s3.amazonaws.com/img1.jpg", "https://bucket.s3.amazonaws.com/img2.jpg"},
		}

		mockDB.On("First", mock.AnythingOfType("*models.Materials"), "1").
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.Materials) = existing }).Once()
		mockDB.On("Save", mock.AnythingOfType("*models.Materials")).Return(nil).Once()

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB))
		router.DELETE("/material/:id/image", fh.HandleDeleteMaterialImage())

		req, _ := http.NewRequest("DELETE", "/material/1/image", jsonBody("https://bucket.s3.amazonaws.com/img1.jpg"))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "Image deleted successfully", resp["message"])
		pics := resp["pictures"].([]interface{})
		assert.Len(t, pics, 1)
		mockDB.AssertExpectations(t)
	})

	t.Run("missing image_url in body returns 400", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB))
		router.DELETE("/material/:id/image", fh.HandleDeleteMaterialImage())

		req, _ := http.NewRequest("DELETE", "/material/1/image", bytes.NewBufferString(`{}`))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "image_url is required")
	})

	t.Run("invalid JSON body returns 400", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB))
		router.DELETE("/material/:id/image", fh.HandleDeleteMaterialImage())

		req, _ := http.NewRequest("DELETE", "/material/1/image", bytes.NewBufferString(`not-json`))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "image_url is required")
	})

	t.Run("material not found returns 404", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		mockDB.On("First", mock.AnythingOfType("*models.Materials"), "999").
			Return(gorm.ErrRecordNotFound).Once()

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB))
		router.DELETE("/material/:id/image", fh.HandleDeleteMaterialImage())

		req, _ := http.NewRequest("DELETE", "/material/999/image", jsonBody("https://bucket.s3.amazonaws.com/img1.jpg"))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Material not found")
		mockDB.AssertExpectations(t)
	})

	t.Run("database error on fetch returns 500", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		mockDB.On("First", mock.AnythingOfType("*models.Materials"), "1").
			Return(assert.AnError).Once()

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB))
		router.DELETE("/material/:id/image", fh.HandleDeleteMaterialImage())

		req, _ := http.NewRequest("DELETE", "/material/1/image", jsonBody("https://bucket.s3.amazonaws.com/img1.jpg"))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Database error")
		mockDB.AssertExpectations(t)
	})

	t.Run("image URL not found in pictures returns 404", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		existing := models.Materials{
			ID:       1,
			Pictures: pq.StringArray{"https://bucket.s3.amazonaws.com/img1.jpg"},
		}

		mockDB.On("First", mock.AnythingOfType("*models.Materials"), "1").
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.Materials) = existing }).Once()

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB))
		router.DELETE("/material/:id/image", fh.HandleDeleteMaterialImage())

		req, _ := http.NewRequest("DELETE", "/material/1/image", jsonBody("https://bucket.s3.amazonaws.com/does-not-exist.jpg"))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Image URL not found")
		mockDB.AssertExpectations(t)
	})

	t.Run("database error on save returns 500", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		existing := models.Materials{
			ID:       1,
			Pictures: pq.StringArray{"https://bucket.s3.amazonaws.com/img1.jpg"},
		}

		mockDB.On("First", mock.AnythingOfType("*models.Materials"), "1").
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.Materials) = existing }).Once()
		mockDB.On("Save", mock.AnythingOfType("*models.Materials")).Return(assert.AnError).Once()

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB))
		router.DELETE("/material/:id/image", fh.HandleDeleteMaterialImage())

		req, _ := http.NewRequest("DELETE", "/material/1/image", jsonBody("https://bucket.s3.amazonaws.com/img1.jpg"))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to update material")
		mockDB.AssertExpectations(t)
	})

	t.Run("missing id - route not matched returns 404", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB))
		router.DELETE("/material/:id/image", fh.HandleDeleteMaterialImage())

		req, _ := http.NewRequest("DELETE", "/material/image", jsonBody("https://bucket.s3.amazonaws.com/img1.jpg"))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
	t.Run("delete chunk successfully and recalculate material available count", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}

		chunk := models.Materialchunks{
			ID:         1,
			MaterialID: 10,
			Serialnum:  "SR-001-1",
			Status:     "available",
		}

		chunks := []models.Materialchunks{
			{ID: 2, MaterialID: 10, Status: "available"},
			{ID: 3, MaterialID: 10, Status: "used"},
			{ID: 4, MaterialID: 10, Status: "available"},
		}

		material := models.Materials{
			ID:        10,
			Name:      "Steel Rod",
			Available: 5,
		}

		// Fetch chunk
		mockDB.On("First",
			mock.AnythingOfType("*models.Materialchunks"),
			"1",
		).Return(nil).Run(func(args mock.Arguments) {
			*args.Get(0).(*models.Materialchunks) = chunk
		}).Once()

		// Delete chunk
		mockDB.On("Delete",
			mock.AnythingOfType("*models.Materialchunks"),
		).Return(nil).Once()

		// Fetch remaining chunks
		mockDB.On("Where",
			"material_id = ?",
			chunk.MaterialID,
		).Return(mockDB).Once()

		mockDB.On("Find",
			mock.AnythingOfType("*[]models.Materialchunks"),
		).Return(nil).Run(func(args mock.Arguments) {
			*args.Get(0).(*[]models.Materialchunks) = chunks
		}).Once()

		// Fetch material
		mockDB.On("First",
			mock.AnythingOfType("*models.Materials"),
			chunk.MaterialID,
		).Return(nil).Run(func(args mock.Arguments) {
			*args.Get(0).(*models.Materials) = material
		}).Once()

		// Save material with recalculated available count = 2
		mockDB.On("Save",
			mock.MatchedBy(func(m *models.Materials) bool {
				return m.ID == 10 && m.Available == 2
			}),
		).Return(nil).Once()

		router := gin.New()
		fh := newFileHandler(mockS3)

		router.Use(injectDB(mockDB), injectTenant("acme"))
		router.DELETE("/chunk/:id", fh.HandleDeleteMaterialChunk())

		req, _ := http.NewRequest("DELETE", "/chunk/1", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Chunk deleted successfully")

		mockDB.AssertExpectations(t)
	})
	t.Run("database error fetching chunk", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}

		mockDB.On("First",
			mock.AnythingOfType("*models.Materialchunks"),
			"1",
		).Return(assert.AnError).Once()

		router := gin.New()
		fh := newFileHandler(mockS3)

		router.Use(injectDB(mockDB), injectTenant("acme"))
		router.DELETE("/chunk/:id", fh.HandleDeleteMaterialChunk())

		req, _ := http.NewRequest("DELETE", "/chunk/1", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Database error")

		mockDB.AssertExpectations(t)
	})

	t.Run("delete chunk fails", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}

		chunk := models.Materialchunks{
			ID:         1,
			MaterialID: 10,
		}

		mockDB.On("First",
			mock.AnythingOfType("*models.Materialchunks"),
			"1",
		).Return(nil).Run(func(args mock.Arguments) {
			*args.Get(0).(*models.Materialchunks) = chunk
		}).Once()

		mockDB.On("Delete",
			mock.AnythingOfType("*models.Materialchunks"),
		).Return(assert.AnError).Once()

		router := gin.New()
		fh := newFileHandler(mockS3)

		router.Use(injectDB(mockDB), injectTenant("acme"))
		router.DELETE("/chunk/:id", fh.HandleDeleteMaterialChunk())

		req, _ := http.NewRequest("DELETE", "/chunk/1", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to delete chunk")

		mockDB.AssertExpectations(t)
	})
}

// ── HandleReplaceMaterialImage ────────────────────────────────────────────────

func TestHandleReplaceMaterialImage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const existingURL = "https://bucket.s3.amazonaws.com/img1.jpg"

	replaceBody := func(replaceURL string, addImage bool) (*bytes.Buffer, string) {
		fields := map[string]string{}
		if replaceURL != "" {
			fields["replace_url"] = replaceURL
		}
		return multipartBody(fields, addImage)
	}

	t.Run("replaces image successfully", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		existing := models.Materials{
			ID:       1,
			Pictures: pq.StringArray{existingURL, "https://bucket.s3.amazonaws.com/img2.jpg"},
		}

		mockDB.On("First", mock.AnythingOfType("*models.Materials"), "1").
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.Materials) = existing }).Once()
		mockS3.On("PutObject", mock.AnythingOfType("*s3.PutObjectInput")).
			Return(&s3.PutObjectOutput{}, nil).Once()
		mockDB.On("Save", mock.AnythingOfType("*models.Materials")).Return(nil).Once()

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB), injectTenant("acme"))
		router.POST("/material/:id/image/replace", fh.HandleReplaceMaterialImage())

		body, ct := replaceBody(existingURL, true)
		req, _ := http.NewRequest("POST", "/material/1/image/replace", body)
		req.Header.Set("Content-Type", ct)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "Image replaced successfully", resp["message"])
		assert.NotEmpty(t, resp["new_url"])
		pics := resp["pictures"].([]interface{})
		assert.Len(t, pics, 2)
		for _, p := range pics {
			assert.NotEqual(t, existingURL, p)
		}
		mockDB.AssertExpectations(t)
		mockS3.AssertExpectations(t)
	})

	t.Run("missing replace_url returns 400", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB), injectTenant("acme"))
		router.POST("/material/:id/image/replace", fh.HandleReplaceMaterialImage())

		body, ct := replaceBody("", true)
		req, _ := http.NewRequest("POST", "/material/1/image/replace", body)
		req.Header.Set("Content-Type", ct)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "replace_url is required")
	})

	t.Run("material not found returns 404", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		mockDB.On("First", mock.AnythingOfType("*models.Materials"), "999").
			Return(gorm.ErrRecordNotFound).Once()

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB), injectTenant("acme"))
		router.POST("/material/:id/image/replace", fh.HandleReplaceMaterialImage())

		body, ct := replaceBody(existingURL, true)
		req, _ := http.NewRequest("POST", "/material/999/image/replace", body)
		req.Header.Set("Content-Type", ct)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Material not found")
		mockDB.AssertExpectations(t)
	})

	t.Run("database error on fetch returns 500", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		mockDB.On("First", mock.AnythingOfType("*models.Materials"), "1").
			Return(assert.AnError).Once()

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB), injectTenant("acme"))
		router.POST("/material/:id/image/replace", fh.HandleReplaceMaterialImage())

		body, ct := replaceBody(existingURL, true)
		req, _ := http.NewRequest("POST", "/material/1/image/replace", body)
		req.Header.Set("Content-Type", ct)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Database error")
		mockDB.AssertExpectations(t)
	})

	t.Run("replace_url not found in pictures returns 404", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		existing := models.Materials{
			ID:       1,
			Pictures: pq.StringArray{"https://bucket.s3.amazonaws.com/other.jpg"},
		}

		mockDB.On("First", mock.AnythingOfType("*models.Materials"), "1").
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.Materials) = existing }).Once()

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB), injectTenant("acme"))
		router.POST("/material/:id/image/replace", fh.HandleReplaceMaterialImage())

		body, ct := replaceBody(existingURL, true)
		req, _ := http.NewRequest("POST", "/material/1/image/replace", body)
		req.Header.Set("Content-Type", ct)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "replace_url not found")
		mockDB.AssertExpectations(t)
	})

	t.Run("missing image file returns 400", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		existing := models.Materials{
			ID:       1,
			Pictures: pq.StringArray{existingURL},
		}

		mockDB.On("First", mock.AnythingOfType("*models.Materials"), "1").
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.Materials) = existing }).Once()

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB), injectTenant("acme"))
		router.POST("/material/:id/image/replace", fh.HandleReplaceMaterialImage())

		body, ct := replaceBody(existingURL, false)
		req, _ := http.NewRequest("POST", "/material/1/image/replace", body)
		req.Header.Set("Content-Type", ct)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "image file is required")
		mockDB.AssertExpectations(t)
	})

	t.Run("S3 upload error returns 500", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		existing := models.Materials{
			ID:       1,
			Pictures: pq.StringArray{existingURL},
		}

		mockDB.On("First", mock.AnythingOfType("*models.Materials"), "1").
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.Materials) = existing }).Once()
		mockS3.On("PutObject", mock.AnythingOfType("*s3.PutObjectInput")).
			Return(nil, assert.AnError).Once()

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB), injectTenant("acme"))
		router.POST("/material/:id/image/replace", fh.HandleReplaceMaterialImage())

		body, ct := replaceBody(existingURL, true)
		req, _ := http.NewRequest("POST", "/material/1/image/replace", body)
		req.Header.Set("Content-Type", ct)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to upload image")
		mockDB.AssertExpectations(t)
		mockS3.AssertExpectations(t)
	})

	t.Run("database error on save returns 500", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		existing := models.Materials{
			ID:       1,
			Pictures: pq.StringArray{existingURL},
		}

		mockDB.On("First", mock.AnythingOfType("*models.Materials"), "1").
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.Materials) = existing }).Once()
		mockS3.On("PutObject", mock.AnythingOfType("*s3.PutObjectInput")).
			Return(&s3.PutObjectOutput{}, nil).Once()
		mockDB.On("Save", mock.AnythingOfType("*models.Materials")).Return(assert.AnError).Once()

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB), injectTenant("acme"))
		router.POST("/material/:id/image/replace", fh.HandleReplaceMaterialImage())

		body, ct := replaceBody(existingURL, true)
		req, _ := http.NewRequest("POST", "/material/1/image/replace", body)
		req.Header.Set("Content-Type", ct)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to update image")
		mockDB.AssertExpectations(t)
		mockS3.AssertExpectations(t)
	})

	t.Run("missing id - route not matched returns 400", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB), injectTenant("acme"))
		router.POST("/material/:id/image/replace", fh.HandleReplaceMaterialImage())

		body, ct := replaceBody(existingURL, true)
		req, _ := http.NewRequest("POST", "/material//image/replace", body)
		req.Header.Set("Content-Type", ct)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// ── HandleBulkChunkStatus ─────────────────────────────────────────────────────
//
// Valid statuses: "available" and "assigned" only (handler rejects anything else).
// All writes use database.Exec(...), never Save.
// Release block only frees chunks whose Status == "assigned".

func TestHandleBulkChunkStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	bulkBody := func(v interface{}) *bytes.Buffer {
		b, _ := json.Marshal(v)
		return bytes.NewBuffer(b)
	}

	t.Run("invalid status returns 400", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		router := gin.New()
		fh := handlers.NewFileHandler(mockS3, "us-east-1")
		router.Use(injectDB(mockDB))
		router.POST("/chunks/status", fh.HandleBulkChunkStatus())

		req, _ := http.NewRequest("POST", "/chunks/status", bulkBody(map[string]interface{}{
			"chunk_ids": []string{"1"},
			"status":    "sold", // invalid
		}))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "status must be")
	})

	t.Run("invalid request body returns 400", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		router := gin.New()
		fh := handlers.NewFileHandler(mockS3, "us-east-1")
		router.Use(injectDB(mockDB))
		router.POST("/chunks/status", fh.HandleBulkChunkStatus())

		req, _ := http.NewRequest("POST", "/chunks/status", bytes.NewBufferString("not-json"))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid request body")
	})

	// Assign a free chunk to a job.
	t.Run("assign chunks successfully", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		chunk := models.Materialchunks{ID: 1, MaterialID: 10, Status: "available"}

		mockDB.On("Where", "id IN ?", mock.Anything).Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]models.Materialchunks")).
			Return(nil).
			Run(func(args mock.Arguments) {
				*args.Get(0).(*[]models.Materialchunks) = []models.Materialchunks{chunk}
			}).Once()
		// Exec to update chunk status
		mockDB.On("Exec",
			"UPDATE materialchunks SET status = ?, job_id = ?, updated_at = ? WHERE id = ?",
			mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		).Return(nil).Once()
		// Recalculate available count
		mockDB.On("Where", "material_id = ?", mock.Anything).Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]models.Materialchunks")).Return(nil).Once()
		mockDB.On("First", mock.AnythingOfType("*models.Materials"), mock.Anything).Return(nil).Once()
		mockDB.On("Exec",
			"UPDATE materials SET available = ?, updated_at = ? WHERE id = ?",
			mock.Anything, mock.Anything, mock.Anything,
		).Return(nil).Once()

		router := gin.New()
		fh := handlers.NewFileHandler(mockS3, "us-east-1")
		router.Use(injectDB(mockDB))
		router.POST("/chunks/status", fh.HandleBulkChunkStatus())

		req, _ := http.NewRequest("POST", "/chunks/status", bulkBody(map[string]interface{}{
			"chunk_ids":   []string{"1"},
			"status":      "assigned",
			"job_id":      "JOB-001",
			"release_ids": []string{},
		}))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(1), resp["updated_count"])
		mockDB.AssertExpectations(t)
	})

	// Release previously assigned chunks back to available.
	t.Run("release chunks back to available", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		// Chunk must have Status == "assigned" for the release block to act on it.
		chunk := models.Materialchunks{ID: 2, MaterialID: 10, Status: "assigned", JobID: "JOB-001"}

		mockDB.On("Where", "id IN ?", mock.Anything).Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]models.Materialchunks")).
			Return(nil).
			Run(func(args mock.Arguments) {
				*args.Get(0).(*[]models.Materialchunks) = []models.Materialchunks{chunk}
			}).Once()
		// Exec to release
		mockDB.On("Exec",
			"UPDATE materialchunks SET status = ?, job_id = ?, updated_at = ? WHERE id = ?",
			mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		).Return(nil).Once()
		// Recalculate available count
		mockDB.On("Where", "material_id = ?", mock.Anything).Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]models.Materialchunks")).Return(nil).Once()
		mockDB.On("First", mock.AnythingOfType("*models.Materials"), mock.Anything).Return(nil).Once()
		mockDB.On("Exec",
			"UPDATE materials SET available = ?, updated_at = ? WHERE id = ?",
			mock.Anything, mock.Anything, mock.Anything,
		).Return(nil).Once()

		router := gin.New()
		fh := handlers.NewFileHandler(mockS3, "us-east-1")
		router.Use(injectDB(mockDB))
		router.POST("/chunks/status", fh.HandleBulkChunkStatus())

		req, _ := http.NewRequest("POST", "/chunks/status", bulkBody(map[string]interface{}{
			"chunk_ids":   []string{},
			"status":      "available",
			"job_id":      "",
			"release_ids": []string{"2"},
		}))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockDB.AssertExpectations(t)
	})

	// A chunk that is not "assigned" is silently skipped in the release block.
	t.Run("non-assigned chunk skipped in release block", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		chunk := models.Materialchunks{ID: 1, MaterialID: 10, Status: "available"}

		mockDB.On("Where", "id IN ?", mock.Anything).Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]models.Materialchunks")).
			Return(nil).
			Run(func(args mock.Arguments) {
				*args.Get(0).(*[]models.Materialchunks) = []models.Materialchunks{chunk}
			}).Once()
		// No Exec call expected because chunk.Status != "assigned"

		router := gin.New()
		fh := handlers.NewFileHandler(mockS3, "us-east-1")
		router.Use(injectDB(mockDB))
		router.POST("/chunks/status", fh.HandleBulkChunkStatus())

		req, _ := http.NewRequest("POST", "/chunks/status", bulkBody(map[string]interface{}{
			"chunk_ids":   []string{},
			"status":      "available",
			"job_id":      "",
			"release_ids": []string{"1"},
		}))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockDB.AssertExpectations(t)
	})

	// Chunk already assigned to a different job: handler looks for an alternative.
	// When no alternative exists the chunk is skipped.
	t.Run("chunk taken by another job is skipped when no alternative found", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		chunk := models.Materialchunks{ID: 1, MaterialID: 10, Status: "assigned", JobID: "OTHER-JOB"}

		mockDB.On("Where", "id IN ?", mock.Anything).Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]models.Materialchunks")).
			Return(nil).
			Run(func(args mock.Arguments) {
				*args.Get(0).(*[]models.Materialchunks) = []models.Materialchunks{chunk}
			}).Once()
		// Alternative lookup: Where("material_id = ? AND status = ? AND id NOT IN ?", ...).First(...)
		mockDB.On("Where", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockDB)
		mockDB.On("First", mock.AnythingOfType("*models.Materialchunks"), mock.Anything).
			Return(gorm.ErrRecordNotFound).Once()

		router := gin.New()
		fh := handlers.NewFileHandler(mockS3, "us-east-1")
		router.Use(injectDB(mockDB))
		router.POST("/chunks/status", fh.HandleBulkChunkStatus())

		req, _ := http.NewRequest("POST", "/chunks/status", bulkBody(map[string]interface{}{
			"chunk_ids":   []string{"1"},
			"status":      "assigned",
			"job_id":      "MY-JOB",
			"release_ids": []string{},
		}))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["updated_count"])
		skipped := resp["skipped_ids"].([]interface{})
		assert.Len(t, skipped, 1)
		mockDB.AssertExpectations(t)
	})

	t.Run("db error fetching release chunks returns 500", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		mockDB.On("Where", "id IN ?", mock.Anything).Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]models.Materialchunks")).Return(assert.AnError).Once()

		router := gin.New()
		fh := handlers.NewFileHandler(mockS3, "us-east-1")
		router.Use(injectDB(mockDB))
		router.POST("/chunks/status", fh.HandleBulkChunkStatus())

		req, _ := http.NewRequest("POST", "/chunks/status", bulkBody(map[string]interface{}{
			"chunk_ids":   []string{},
			"status":      "available",
			"job_id":      "",
			"release_ids": []string{"1"},
		}))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "failed to fetch chunks to release")
		mockDB.AssertExpectations(t)
	})

	t.Run("db error fetching chunk_ids returns 500", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		mockDB.On("Where", "id IN ?", mock.Anything).Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]models.Materialchunks")).Return(assert.AnError).Once()

		router := gin.New()
		fh := handlers.NewFileHandler(mockS3, "us-east-1")
		router.Use(injectDB(mockDB))
		router.POST("/chunks/status", fh.HandleBulkChunkStatus())

		req, _ := http.NewRequest("POST", "/chunks/status", bulkBody(map[string]interface{}{
			"chunk_ids":   []string{"1"},
			"status":      "assigned",
			"job_id":      "JOB-001",
			"release_ids": []string{},
		}))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "failed to fetch chunks")
		mockDB.AssertExpectations(t)
	})
}

// ── Additional HandleMaterialUpdate coverage ──────────────────────────────────

func TestHandleMaterialUpdateAdditional(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("available count already met - no new chunks created (non-procurement)", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		existing := models.Materials{
			ID:        1,
			Name:      "Steel Rod",
			Serialnum: "SR-001",
			Available: 2,
			Chunks: []models.Materialchunks{
				{ID: 1, Serialnum: "SR-001-1"},
				{ID: 2, Serialnum: "SR-001-2"},
				{ID: 3, Serialnum: "SR-001-3"},
			},
		}

		mockDB.On("Preload", "Chunks").Return(mockDB)
		mockDB.On("First", mock.AnythingOfType("*models.Materials"), "1").
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.Materials) = existing }).Once()
		mockDB.On("Save", mock.AnythingOfType("*models.Materials")).Return(nil).Once()
		// newCount (2) <= currentCount (3): maxChunkSuffix and Create NOT called

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB), injectTenant("acme"))
		router.PUT("/material/:id", fh.HandleMaterialUpdate())

		body, ct := multipartBody(map[string]string{"available": "2"}, false)
		req, _ := http.NewRequest("PUT", "/material/1", body)
		req.Header.Set("Content-Type", ct)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("procurement - available already met so no chunks created", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		existing := models.Materials{
			ID:        1,
			Name:      "Steel Rod",
			Serialnum: "SR-001",
			Available: 3,
			Chunks: []models.Materialchunks{
				{ID: 1, Serialnum: "SR-001-1", Status: "available"},
				{ID: 2, Serialnum: "SR-001-2", Status: "available"},
				{ID: 3, Serialnum: "SR-001-3", Status: "available"},
			},
		}

		mockDB.On("Preload", "Chunks").Return(mockDB)
		mockDB.On("First", mock.AnythingOfType("*models.Materials"), "1").
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.Materials) = existing }).Once()
		mockDB.On("Save", mock.AnythingOfType("*models.Materials")).Return(nil).Once()
		// availableChunkCount (3) == newCount (3) → chunksToCreate=0, no Create called

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB), injectTenant("acme"))
		router.PUT("/material/:id", fh.HandleMaterialUpdate())

		body, ct := multipartBody(map[string]string{
			"available":   "3",
			"procurement": "true",
			"lot":         "L1",
		}, false)
		req, _ := http.NewRequest("PUT", "/material/1", body)
		req.Header.Set("Content-Type", ct)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("chunk revision save error returns 500", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		mockS3 := &mocks.MockS3Client{}
		mockPresignRequest(mockS3)

		existing := models.Materialchunks{
			ID:         1,
			MaterialID: 10,
			Serialnum:  "SR-001-1",
			Dimensions: "100x200",
		}

		mockDB.On("First", mock.AnythingOfType("*models.Materialchunks"), "1").
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*models.Materialchunks) = existing }).Once()

		mockDB.On("Model", mock.AnythingOfType("*models.Materialchunks")).Return(mockDB)
		mockDB.On("Where", mock.Anything, mock.Anything, mock.Anything).Return(mockDB)
		mockDB.On("Count", mock.AnythingOfType("*int64")).
			Return(nil).
			Run(func(args mock.Arguments) { *args.Get(0).(*int64) = 0 }).Once()

		mockDB.On("Create", mock.AnythingOfType("*models.Materialchunks")).Return(nil).Once()
		// Save on the original chunk fails
		mockDB.On("Save", mock.AnythingOfType("*models.Materialchunks")).Return(assert.AnError).Once()

		router := gin.New()
		fh := newFileHandler(mockS3)
		router.Use(injectDB(mockDB), injectTenant("acme"))
		router.PUT("/chunk/:id", fh.HandleMaterialChunkUpdate())

		body, ct := multipartBody(map[string]string{"dimensions": "100x150"}, false)
		req, _ := http.NewRequest("PUT", "/chunk/1", body)
		req.Header.Set("Content-Type", ct)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to update original chunk")
		mockDB.AssertExpectations(t)
	})
}

// ── Picture-enrichment smoke tests ────────────────────────────────────────────

func TestHandleFetchMaterialsWithPictures(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := &mocks.MockDatabase{}
	mockS3 := &mocks.MockS3Client{}
	mockPresignRequest(mockS3)

	materials := []models.Materials{
		{ID: 1, Name: "Steel", Pictures: pq.StringArray{"test-key"}},
	}

	mockDB.On("Preload", "Chunks", mock.Anything).Return(mockDB)
	mockDB.On("Find", mock.AnythingOfType("*[]models.Materials")).
		Return(nil).
		Run(func(args mock.Arguments) {
			*args.Get(0).(*[]models.Materials) = materials
		}).Once()

	router := gin.New()
	fh := handlers.NewFileHandler(mockS3, "us-east-1")
	router.Use(injectDB(mockDB), injectTenant("acme"))
	router.GET("/materials", fh.HandleFetchMaterials())

	req, _ := http.NewRequest("GET", "/materials", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Steel")
	mockDB.AssertExpectations(t)
	mockS3.AssertExpectations(t)
}

func TestHandleMaterialGetByIDWithPictures(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := &mocks.MockDatabase{}
	mockS3 := &mocks.MockS3Client{}
	mockPresignRequest(mockS3)

	material := models.Materials{
		ID:       1,
		Name:     "Steel Rod",
		Pictures: pq.StringArray{"test-key"},
	}

	mockDB.On("Preload", "Chunks").Return(mockDB)
	mockDB.On("First", mock.AnythingOfType("*models.Materials"), "1").
		Return(nil).
		Run(func(args mock.Arguments) { *args.Get(0).(*models.Materials) = material }).Once()

	router := gin.New()
	fh := handlers.NewFileHandler(mockS3, "us-east-1")
	router.Use(injectDB(mockDB), injectTenant("acme"))
	router.GET("/material/:id", fh.HandleMaterialGetByID())

	req, _ := http.NewRequest("GET", "/material/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Steel Rod")
	mockDB.AssertExpectations(t)
	mockS3.AssertExpectations(t)
}
