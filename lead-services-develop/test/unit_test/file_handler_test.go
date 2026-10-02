package handlers_test

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/tuliporg/agent-management-api/handlers"
	"github.com/tuliporg/agent-management-api/mocks"
	"gorm.io/gorm"
)

func TestCreateFile_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a temporary file to upload
	tmpfile, err := os.CreateTemp("", "testfile-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.WriteString("this is a test file")
	tmpfile.Close()

	// Prepare the multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("files", filepath.Base(tmpfile.Name()))
	if err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	io.Copy(part, file)
	writer.WriteField("category", "test_category")
	writer.WriteField("type", "test_type")
	writer.WriteField("lead_id", "123")
	writer.Close()

	// Mocks
	mockDB := new(mocks.MockDatabase)
	mockS3 := new(mocks.MockS3Client)

	mockDB.On("Create", mock.Anything).Return(nil)
	mockS3.On("PutObject", mock.Anything).Return(&s3.PutObjectOutput{}, nil)

	// Handler
	fileHandler := handlers.NewFileHandler(mockS3, "test-bucket", "us-east-1")

	// Request
	req, _ := http.NewRequest(http.MethodPost, "/files", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	// Execute
	fileHandler.CreateFile(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, strings.Contains(w.Body.String(), "uploadedFiles"))
	mockDB.AssertExpectations(t)
	mockS3.AssertExpectations(t)
}

func TestCreateFile_NoFiles(t *testing.T) {
	gin.SetMode(gin.TestMode)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.Close()

	mockDB := new(mocks.MockDatabase)
	mockS3 := new(mocks.MockS3Client)

	fileHandler := handlers.NewFileHandler(mockS3, "test-bucket", "us-east-1")

	req, _ := http.NewRequest(http.MethodPost, "/files", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	fileHandler.CreateFile(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "No files uploaded")
}

func TestCreateFile_S3UploadError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tmpfile, err := os.CreateTemp("", "testfile-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.WriteString("this is a test file")
	tmpfile.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("files", filepath.Base(tmpfile.Name()))
	file, _ := os.Open(tmpfile.Name())
	defer file.Close()
	io.Copy(part, file)
	writer.Close()

	mockDB := new(mocks.MockDatabase)
	mockS3 := new(mocks.MockS3Client)

	mockS3.On("PutObject", mock.Anything).Return(nil, assert.AnError)

	fileHandler := handlers.NewFileHandler(mockS3, "test-bucket", "us-east-1")

	req, _ := http.NewRequest(http.MethodPost, "/files", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	fileHandler.CreateFile(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to upload file")
	mockS3.AssertExpectations(t)
}

func TestCreateFile_DBError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tmpfile, err := os.CreateTemp("", "testfile-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.WriteString("this is a test file")
	tmpfile.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("files", filepath.Base(tmpfile.Name()))
	file, _ := os.Open(tmpfile.Name())
	defer file.Close()
	io.Copy(part, file)
	writer.Close()

	mockDB := new(mocks.MockDatabase)
	mockS3 := new(mocks.MockS3Client)

	mockS3.On("PutObject", mock.Anything).Return(&s3.PutObjectOutput{}, nil)
	mockDB.On("Create", mock.Anything).Return(assert.AnError)

	fileHandler := handlers.NewFileHandler(mockS3, "test-bucket", "us-east-1")

	req, _ := http.NewRequest(http.MethodPost, "/files", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("db", mockDB)

	fileHandler.CreateFile(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "DB save failed")
	mockS3.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

func TestGetFile_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockDB := new(mocks.MockDatabase)
	fileHandler := handlers.NewFileHandler(nil, "", "")

	mockDB.On("Preload", "Lead").Return(mockDB)
	mockDB.On("Where", "lead_id = ?", "1").Return(mockDB)
	mockDB.On("Find", mock.Anything).Return(nil)

	req, _ := http.NewRequest(http.MethodGet, "/files/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	fileHandler.GetFile(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockDB.AssertExpectations(t)
}

func TestGetFile_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockDB := new(mocks.MockDatabase)
	fileHandler := handlers.NewFileHandler(nil, "", "")

	mockDB.On("Preload", "Lead").Return(mockDB)
	mockDB.On("Where", "lead_id = ?", "1").Return(mockDB)
	mockDB.On("Find", mock.Anything).Run(func(args mock.Arguments) {
		// Do nothing to simulate not found
	}).Return(nil)

	req, _ := http.NewRequest(http.MethodGet, "/files/1", nil)
	w := httptest.NewRecorder()
	c, router := gin.CreateTestContext(w)
	router.GET("/files/:id", fileHandler.GetFile)
	c.Request = req
	c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	fileHandler.GetFile(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockDB.AssertExpectations(t)
}

func TestUpdateFile_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockDB := new(mocks.MockDatabase)
	fileHandler := handlers.NewFileHandler(nil, "", "")

	mockDB.On("First", mock.Anything, uint64(1)).Return(nil)
	mockDB.On("Save", mock.Anything).Return(nil)

	body := strings.NewReader(`{"Name": "new.jpg", "Type": "image/jpeg", "Size": "1234", "StoredFileName": "new.jpg", "FilePath": "/new.jpg"}`)
	req, _ := http.NewRequest(http.MethodPut, "/files/1", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	fileHandler.UpdateFile(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockDB.AssertExpectations(t)
}

func TestUpdateFile_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fileHandler := handlers.NewFileHandler(nil, "", "")

	// Create a mock DB to satisfy GetDB
	mockDB := new(mocks.MockDatabase)

	req, _ := http.NewRequest(http.MethodPut, "/files/invalid", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{gin.Param{Key: "id", Value: "invalid"}}
	c.Set("db", mockDB) // <-- must set the DB

	fileHandler.UpdateFile(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateFile_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockDB := new(mocks.MockDatabase)
	fileHandler := handlers.NewFileHandler(nil, "", "")

	mockDB.On("First", mock.Anything, uint64(1)).Return(gorm.ErrRecordNotFound)

	body := strings.NewReader(`{"Name": "new.jpg", "Type": "image/jpeg", "Size": "1234", "StoredFileName": "new.jpg", "FilePath": "/new.jpg"}`)
	req, _ := http.NewRequest(http.MethodPut, "/files/1", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	fileHandler.UpdateFile(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockDB.AssertExpectations(t)
}

func TestDeleteFile_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fileHandler := handlers.NewFileHandler(nil, "", "")

	// Mock DB just to avoid panic
	mockDB := new(mocks.MockDatabase)

	req, _ := http.NewRequest(http.MethodDelete, "/files/invalid", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{gin.Param{Key: "id", Value: "invalid"}}
	c.Set("db", mockDB) // <-- must set db in context

	fileHandler.DeleteFile(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateFileCategory_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockDB := new(mocks.MockDatabase)
	fileHandler := handlers.NewFileHandler(nil, "", "")

	mockDB.On("First", mock.Anything, "1").Return(nil)
	mockDB.On("Save", mock.Anything).Return(nil)

	body := strings.NewReader(`{"ProjectType": "new_type", "Category": "new_category"}`)
	req, _ := http.NewRequest(http.MethodPatch, "/files/1", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	fileHandler.UpdateFileCategory(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockDB.AssertExpectations(t)
}

func TestDeleteFile_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockDB := new(mocks.MockDatabase)
	fileHandler := handlers.NewFileHandler(nil, "", "")

	mockDB.On("Delete", mock.Anything, uint64(1)).Return(nil)

	req, _ := http.NewRequest(http.MethodDelete, "/files/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	fileHandler.DeleteFile(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockDB.AssertExpectations(t)
}

func TestGetFile_DBError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockDB := new(mocks.MockDatabase)
	fileHandler := handlers.NewFileHandler(nil, "", "")

	mockDB.On("Preload", "Lead").Return(mockDB)
	mockDB.On("Where", "lead_id = ?", "1").Return(mockDB)
	mockDB.On("Find", mock.Anything).Return(gorm.ErrInvalidDB)

	req, _ := http.NewRequest(http.MethodGet, "/files/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	fileHandler.GetFile(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockDB.AssertExpectations(t)
}

func TestUpdateFileCategory_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockDB := new(mocks.MockDatabase)
	fileHandler := handlers.NewFileHandler(nil, "", "")

	mockDB.On("First", mock.Anything, "1").Return(gorm.ErrRecordNotFound)

	body := strings.NewReader(`{"ProjectType": "new_type"}`)
	req, _ := http.NewRequest(http.MethodPatch, "/files/1", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	fileHandler.UpdateFileCategory(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockDB.AssertExpectations(t)
}

func TestUpdateFileCategory_InvalidInput(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fileHandler := handlers.NewFileHandler(nil, "", "")
	mockDB := new(mocks.MockDatabase)
	mockDB.On("First", mock.Anything, "1").Return(nil)

	body := strings.NewReader(`invalid json`)
	req, _ := http.NewRequest(http.MethodPatch, "/files/1", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	fileHandler.UpdateFileCategory(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateFileCategory_DBError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockDB := new(mocks.MockDatabase)
	fileHandler := handlers.NewFileHandler(nil, "", "")

	mockDB.On("First", mock.Anything, "1").Return(nil)
	mockDB.On("Save", mock.Anything).Return(gorm.ErrInvalidDB)

	body := strings.NewReader(`{"ProjectType": "new_type"}`)
	req, _ := http.NewRequest(http.MethodPatch, "/files/1", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	fileHandler.UpdateFileCategory(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockDB.AssertExpectations(t)
}

func TestUpdateFile_InvalidInput(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockDB := new(mocks.MockDatabase)
	fileHandler := handlers.NewFileHandler(nil, "", "")

	mockDB.On("First", mock.Anything, uint64(1)).Return(nil)

	body := strings.NewReader(`invalid json`)
	req, _ := http.NewRequest(http.MethodPut, "/files/1", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	fileHandler.UpdateFile(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateFile_MissingFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockDB := new(mocks.MockDatabase)
	fileHandler := handlers.NewFileHandler(nil, "", "")

	mockDB.On("First", mock.Anything, uint64(1)).Return(nil)

	body := strings.NewReader(`{"Name": "new.jpg"}`)
	req, _ := http.NewRequest(http.MethodPut, "/files/1", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	fileHandler.UpdateFile(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteFile_DBError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockDB := new(mocks.MockDatabase)
	fileHandler := handlers.NewFileHandler(nil, "", "")

	mockDB.On("Delete", mock.Anything, uint64(1)).Return(gorm.ErrInvalidDB)

	req, _ := http.NewRequest(http.MethodDelete, "/files/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	fileHandler.DeleteFile(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockDB.AssertExpectations(t)
}

func TestUpdateFile_InvalidLeadID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockDB := new(mocks.MockDatabase)
	fileHandler := handlers.NewFileHandler(nil, "", "")

	leadID := "999"

	mockDB.On("First", mock.Anything, uint64(1)).Return(nil)
	mockDB.On("First", mock.Anything, &leadID).Return(gorm.ErrRecordNotFound)

	body := strings.NewReader(`{"Name": "new.jpg", "Type": "image/jpeg", "Size": "1234", "StoredFileName": "new.jpg", "FilePath": "/new.jpg", "LeadID": "999"}`)
	req, _ := http.NewRequest(http.MethodPut, "/files/1", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	fileHandler.UpdateFile(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateFile_DBSaveError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockDB := new(mocks.MockDatabase)
	fileHandler := handlers.NewFileHandler(nil, "", "")

	mockDB.On("First", mock.Anything, uint64(1)).Return(nil)
	mockDB.On("Save", mock.Anything).Return(gorm.ErrInvalidDB)

	body := strings.NewReader(`{"Name": "new.jpg", "Type": "image/jpeg", "Size": "1234", "StoredFileName": "new.jpg", "FilePath": "/new.jpg"}`)
	req, _ := http.NewRequest(http.MethodPut, "/files/1", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
	c.Set("db", mockDB)

	fileHandler.UpdateFile(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockDB.AssertExpectations(t)
}
