package test

import (
	"bytes"
	"encoding/json"
	"errors"
	"inventory/handlers"
	"inventory/mocks"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAdminLoginHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Helper to create a request
	newRequest := func(body interface{}) *http.Request {
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/admin/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		return req
	}

	t.Run("successful login", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		email := "admin@example.com"
		password := "secret123"

		// Mock Raw call. req.Email is passed as an argument.
		mockDB.On("Raw", "SELECT password FROM admin WHERE email = ?", email).Return(mockDB)
		
		// Mock Scan call to return success and set password
		mockDB.On("Scan", mock.AnythingOfType("*string")).Return(nil).Run(func(args mock.Arguments) {
			p := args.Get(0).(*string)
			*p = password
		})

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/admin/login", handlers.AdminLoginHandler)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, newRequest(handlers.LoginRequest{Email: email, Password: password}))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "login successful")
		mockDB.AssertExpectations(t)
	})

	t.Run("invalid request body - missing fields", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/admin/login", handlers.AdminLoginHandler)

		w := httptest.NewRecorder()
		// Email is required by LoginRequest struct binding
		router.ServeHTTP(w, newRequest(map[string]string{"password": "test"}))

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid request body")
	})

	t.Run("database connection not found in context", func(t *testing.T) {
		router := gin.New()
		// No injectDB middleware provided
		router.POST("/admin/login", handlers.AdminLoginHandler)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, newRequest(handlers.LoginRequest{Email: "a@b.com", Password: "p"}))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "database connection not found")
	})

	t.Run("internal server error during database query", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		email := "admin@test.com"

		mockDB.On("Raw", mock.Anything, email).Return(mockDB)
		mockDB.On("Scan", mock.Anything).Return(errors.New("db query error"))

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/admin/login", handlers.AdminLoginHandler)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, newRequest(handlers.LoginRequest{Email: email, Password: "any"}))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "internal server error")
		mockDB.AssertExpectations(t)
	})

	t.Run("unauthorized - user not found", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		email := "unknown@test.com"

		mockDB.On("Raw", mock.Anything, email).Return(mockDB)
		mockDB.On("Scan", mock.AnythingOfType("*string")).Return(nil).Run(func(args mock.Arguments) {
			p := args.Get(0).(*string)
			*p = "" // Password remains empty if user not found
		})

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/admin/login", handlers.AdminLoginHandler)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, newRequest(handlers.LoginRequest{Email: email, Password: "any"}))

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "invalid email or password")
		mockDB.AssertExpectations(t)
	})

	t.Run("unauthorized - wrong password", func(t *testing.T) {
		mockDB := &mocks.MockDatabase{}
		email := "admin@test.com"
		correctPassword := "real_password"

		mockDB.On("Raw", mock.Anything, email).Return(mockDB)
		mockDB.On("Scan", mock.AnythingOfType("*string")).Return(nil).Run(func(args mock.Arguments) {
			p := args.Get(0).(*string)
			*p = correctPassword
		})

		router := gin.New()
		router.Use(injectDB(mockDB))
		router.POST("/admin/login", handlers.AdminLoginHandler)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, newRequest(handlers.LoginRequest{Email: email, Password: "wrong_password"}))

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "invalid email or password")
		mockDB.AssertExpectations(t)
	})
}
