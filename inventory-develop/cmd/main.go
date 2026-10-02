package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"inventory/db"
	"inventory/internal/postgreslistener"
	"inventory/internal/websocket"
	"inventory/middleware"
	"inventory/routes"
	"inventory/services"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			origin = "*"
		}
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Tenant, Upgrade, Connection")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func buildPostgresDSN() string {
	sslMode := os.Getenv("DB_SSL_MODE")
	if sslMode == "" {
		sslMode = "disable"
	}

	return "host=" + os.Getenv("DB_HOST") +
		" user=" + os.Getenv("DB_USER") +
		" password=" + os.Getenv("DB_PASSWORD") +
		" dbname=" + os.Getenv("DB_NAME") +
		" port=" + os.Getenv("DB_PORT") +
		" sslmode=" + sslMode
}

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Initialize database manager
	db.GetTenantManager()

	if len(os.Args) > 1 && os.Args[1] == "migrate-only" {
		defaultDB, err := db.InitDB()
		if err != nil {
			log.Fatal("Failed to initialize default database for migration:", err)
		}
		db.Migrate(defaultDB)
		return
	}

	if len(os.Args) > 1 && os.Args[1] == "geocode-suppliers" {
		tenants := os.Args[2:]
		if len(tenants) == 0 {
			tenants = []string{"default"}
		}
		if err := services.GeocodeExistingSuppliers(tenants); err != nil {
			log.Fatal("Supplier geocoding failed:", err)
		}
		return
	}

	activityManager := websocket.NewManager(log.Default())
	pgDSN := buildPostgresDSN()
	channel := os.Getenv("ACTIVITY_LISTEN_CHANNEL")
	if channel == "" {
		channel = "activity_events_channel"
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	listener := postgreslistener.NewListener(pgDSN, channel, activityManager.Broadcast, log.Default())
	go func() {
		if err := listener.Start(ctx); err != nil {
			log.Fatalf("activity listener exited: %v", err)
		}
	}()

	// Gin engine is created only when actually serving — not during migrations
	r := gin.Default()
	r.Use(CORSMiddleware())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
		})
	})

	// Register websocket endpoint before tenant/database middleware.
	// This keeps the upgrade path lightweight and avoids handshake failures
	// when the normal HTTP middleware stack cannot complete during websocket setup.
	r.GET("/ws/activity", func(c *gin.Context) {
		websocket.ServeActivityWebSocket(c, activityManager)
	})

	// Apply database middleware to the regular HTTP API routes.
	r.Use(middleware.TenantMiddleware())
	r.Use(middleware.DatabaseMiddleware())

	// Setup routes
	routes.SetupRoutes(r)

	server := &http.Server{
		Addr:              getServerAddress(),
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		log.Printf("Server started on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown signal received, stopping websocket service and listener")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	cancel()
	if err := listener.Close(); err != nil {
		log.Printf("Error closing Postgres listener: %v", err)
	}
	activityManager.Shutdown()
}

func getServerAddress() string {
	port := os.Getenv("INVENTORY_PORT")
	if port == "" {
		port = "8000"
	}
	return "0.0.0.0:" + port
}
