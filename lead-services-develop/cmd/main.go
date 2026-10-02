package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/tuliporg/agent-management-api/billing"
	"github.com/tuliporg/agent-management-api/db"
	"github.com/tuliporg/agent-management-api/middleware"
	"github.com/tuliporg/agent-management-api/routes"
	"github.com/tuliporg/agent-management-api/services"
)

func CSPMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' https://cdnjs.cloudflare.com; " +
			"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; " +
			"font-src 'self' https://fonts.gstatic.com; " +
			"img-src 'self' data: https:; " +
			"connect-src 'self'; " +
			"frame-ancestors 'none'; " +
			"base-uri 'self'"
		c.Header("Content-Security-Policy", csp)
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		// Allow all origins dynamically
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
		} else {
			c.Header("Access-Control-Allow-Origin", "*")
		}
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-User-Timezone, User-Email, x-trace-id, X-Tenant, X-User-Id, X-User-Name, X-Request-User-Id, X-Request-User-Name, withCredentials")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
	}
	dbConfig := db.GetDBConfig()
	tenantManager := db.InitTenantManager(dbConfig)
	db.InitAITenantManager(db.GetAIDBConfig())
	if err := db.InitControlPlaneDB(); err != nil {
		log.Printf("[billing] control-plane init: %v", err)
	} else if gdb, err := db.GetControlPlaneDB(); err == nil {
		if err := billing.MigrateControlPlane(gdb); err != nil {
			log.Printf("[billing] migrate: %v", err)
		}
	}
	if len(os.Args) > 1 && os.Args[1] == "migrate-only" {
		defaultDB, err := db.InitDB()
		if err != nil {
		}
		db.Migrate(defaultDB)
		return
	}

	if len(os.Args) > 1 && os.Args[1] == "geocode-customers" {
		tenants := os.Args[2:]
		if len(tenants) == 0 {
			tenants = []string{"default"}
		}
		if err := services.GeocodeExistingCustomers(tenants); err != nil {
			log.Fatal("Customer geocoding failed:", err)
		}
		return
	}
	r := gin.Default()
	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		traceId := param.Keys["traceId"]

		return fmt.Sprintf(
			"[GIN] traceId=%v %s %s %d %s\n",
			traceId,
			param.Method,
			param.Path,
			param.StatusCode,
			param.Latency,
		)
	}))

	// r.Use(CSPMiddleware())
	r.Use(CORSMiddleware())
	r.Use(middleware.TraceMiddleware())
	r.Use(middleware.TenantMiddleware())
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
		})
	})
	r.Use(middleware.DatabaseMiddleware())
	routes.SetupRoutes(r)
	srv := &http.Server{
		Addr:    "0.0.0.0:8085",
		Handler: r,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
	}
	tenantManager.CloseAll()
}
