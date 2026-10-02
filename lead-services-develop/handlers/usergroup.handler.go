package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tuliporg/agent-management-api/middleware"
	"github.com/tuliporg/agent-management-api/models"
	"gorm.io/gorm"
)

func (h *Handler) CreateUserGroup(c *gin.Context) {
	db := middleware.GetDB(c)

	var input models.UserGroups
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid input",
			Code:    "VALIDATION_ERROR",
			Details: err.Error(),
		})
		return
	}

	tx := db.Begin()

	// Extract user IDs from input
	var userIDs []string
	for _, u := range input.Users {
		userIDs = append(userIDs, u.ID)
	}

	// Fetch existing agents from DB
	var agents []models.Agent
	if len(userIDs) > 0 {
		if err := tx.Where("id IN ?", userIDs).Find(&agents); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to fetch users",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
			return
		}
	}

	input.Users = agents

	// Create group first
	if err := tx.Create(&input); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to create User Group",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	// Attach existing agents

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to commit transaction",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User Group Created successfully",
		"data":    input,
	})
}
func (h *Handler) RemoveUserFromGroup(c *gin.Context) {
	db := middleware.GetDB(c)

	groupID := c.Param("id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Missing Group ID",
			Code:    "VALIDATION_ERROR",
			Details: "Group ID is required in URL path",
		})
		return
	}

	input := c.Param("user_id")

	// Fetch group
	var group models.UserGroups
	if err := db.First(&group, "id = ?", groupID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "UserGroup not found",
				Code:    "NOT_FOUND",
				Details: err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Database error",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
		}
		return
	}

	// Fetch user
	var agent models.Agent
	if err := db.First(&agent, "id = ?", input); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "User not found",
				Code:    "NOT_FOUND",
				Details: err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Database error",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
		}
		return
	}

	// Remove association (🔥 the key line)
	if err := db.Model(&group).Association("Users").Delete(&agent); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to remove user from group",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User removed from group successfully",
	})
}
func (h *Handler) UpdateUserGroup(c *gin.Context) {
	db := middleware.GetDB(c)

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Missing ID",
			Code:    "VALIDATION_ERROR",
			Details: "ID is required in URL path",
		})
		return
	}

	var input models.UserGroups
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid input",
			Code:    "VALIDATION_ERROR",
			Details: err.Error(),
		})
		return
	}

	var existing models.UserGroups
	if err := db.First(&existing, "id = ?", id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "UserGroup not found",
				Code:    "NOT_FOUND",
				Details: err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Database error",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
		}
		return
	}

	existing.GroupName = input.GroupName
	if err := db.Save(&existing); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to update group",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	var userIDs []string
	for _, u := range input.Users {
		userIDs = append(userIDs, u.ID)
	}

	var agents []models.Agent
	if len(userIDs) > 0 {
		if err := db.Where("id IN ?", userIDs).Find(&agents); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to fetch users",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
			return
		}
	}
	if err := db.Model(&existing).Association("Users").Replace(agents); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to update users",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "UserGroup updated successfully",
		"data":    existing,
	})
}
func (h *Handler) DeleteUserGroup(c *gin.Context) {
	db := middleware.GetDB(c)

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Missing ID",
			Code:    "VALIDATION_ERROR",
			Details: "ID is required in URL path",
		})
		return
	}

	tx := db.Begin()
	var userGroup models.UserGroups
	if err := tx.First(&userGroup, "id = ?", id); err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "User group not found",
				Code:    "NOT_FOUND",
				Details: err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to find User Group entry",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
		}
		return
	}
	if err := tx.Model(&userGroup).Association("Users").Clear(); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to clear associations",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}
	if err := tx.Delete(&userGroup); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to delete UserGroup",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to commit transaction",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User group deletion successfully",
	})
}
func (h *Handler) GetAllUserGroups(c *gin.Context) {
	db := middleware.GetDB(c)

	all := c.Query("all")

	var userGroups []models.UserGroups
	var total int64

	// Count total

	query := db.Preload("Users")

	if all == "true" {
		if err := query.Find(&userGroups); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to retrieve all user groups",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":  userGroups,
			"total": total,
			"all":   true,
		})
		return
	}

	if err := query.Find(&userGroups); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to retrieve user groups",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": userGroups,
	})
}
