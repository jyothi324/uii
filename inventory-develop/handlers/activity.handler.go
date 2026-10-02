package handlers

import (
	"inventory/middleware"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

type ActivityEvent struct {
	ID         int64          `json:"id"`
	EntityType string         `json:"entity_type"`
	EntityID   string         `json:"entity_id"`
	Action     string         `json:"action"`
	UserID     string         `json:"user_id"`
	UserName   *string        `json:"user_name"`
	Payload    datatypes.JSON `json:"payload"`
	CreatedAt  time.Time      `json:"created_at"`
}

func GetRecentActivity(c *gin.Context) {
	// Get the tenant-specific database connection from middleware
	tenantDB := middleware.GetDB(c)

	limitParam := c.DefaultQuery("limit", "100")
	limit, err := strconv.Atoi(limitParam)
	if err != nil || limit <= 0 {
		limit = 100
	}
	if limit > 200 {
		limit = 200
	}

	var events []ActivityEvent
	err = tenantDB.Order("created_at desc").Limit(limit).Find(&events)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load activity events"})
		return
	}
	c.JSON(http.StatusOK, events)
}
