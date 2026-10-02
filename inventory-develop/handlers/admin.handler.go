package handlers

import (
	"inventory/db"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func AdminLoginHandler(c *gin.Context) {
	dbVal, exists := c.Get("db")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database connection not found"})
		return
	}
	tenantDB := dbVal.(db.Database)

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	var storedPassword string
	err := tenantDB.Raw("SELECT password FROM admin WHERE email = ?", req.Email).Scan(&storedPassword)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if storedPassword == "" || req.Password != storedPassword {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "login successful"})
}
