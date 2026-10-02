package handlers

import (
	"inventory/middleware"
	"inventory/models"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type feedbackInput struct {
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
	Email     string  `json:"email"`
	Phone     *string `json:"phone"`
	Message   string  `json:"message"`
}

func CreateFeedback() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)

		var input feedbackInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// Required field validation
		if strings.TrimSpace(input.FirstName) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "First name is required"})
			return
		}
		if strings.TrimSpace(input.LastName) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Last name is required"})
			return
		}
		if strings.TrimSpace(input.Email) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
			return
		}
		if strings.TrimSpace(input.Message) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Feedback message is required"})
			return
		}
		if len(input.Message) > 1000 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Message must be 1000 characters or fewer"})
			return
		}

		feedback := models.Feedback{
			FirstName: strings.TrimSpace(input.FirstName),
			LastName:  strings.TrimSpace(input.LastName),
			Email:     strings.TrimSpace(input.Email),
			Phone:     nilIfEmpty(input.Phone),
			Message:   strings.TrimSpace(input.Message),
		}

		if err := database.Create(&feedback); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save feedback"})
			return
		}

		c.JSON(http.StatusCreated, feedback)
	}
}