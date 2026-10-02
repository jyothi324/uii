package middleware

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	// _ "github.com/lib/pq"
	"github.com/tuliporg/agent-management-api/db"
)

func checkDatabaseExists(dbName string) bool {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
	)

	sqlDB, err := sql.Open("postgres", connStr)
	if err != nil {
		return false
	}
	defer sqlDB.Close()

	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)"
	err = sqlDB.QueryRow(query, dbName).Scan(&exists)

	if err != nil {
		return false
	}
	return exists
}

func DatabaseMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// SaaS billing uses the control-plane DB and must work without a tenant_* DB
		// (e.g. Stripe webhooks, pre-provision signup). Existing routes are unchanged.
		if strings.HasPrefix(c.Request.URL.Path, "/api/billing") {
			c.Next()
			return
		}

		tenant := GetTenant(c)
		manager := db.GetTenantManager()
		if manager == nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Database configuration error",
			})
			c.Abort()
			return
		}

		tenantDB, err := manager.GetTenantDB(tenant)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":  "Database connection failed",
				"tenant": tenant,
				"detail": err.Error(),
			})
			c.Abort()
			return
		}

		c.Set(DBContextKey, tenantDB)
		c.Next()
	}
}

// Keep your existing GetDB function as is
func GetDB(c *gin.Context) db.Database {
	dbInterface, exists := c.Get(DBContextKey)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "user not found",
			"message": "no database with that company profile",
		})
	}
	return dbInterface.(db.Database)
}
