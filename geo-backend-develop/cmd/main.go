package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func shutdownTimeout() time.Duration {
	const defaultSeconds = 30
	raw := strings.TrimSpace(os.Getenv("SHUTDOWN_TIMEOUT_SECONDS"))
	if raw == "" {
		return defaultSeconds * time.Second
	}
	seconds, err := strconv.Atoi(raw)
	if err != nil || seconds <= 0 {
		return defaultSeconds * time.Second
	}
	return time.Duration(seconds) * time.Second
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Failed to load .env:", err)
	}
	initJWTSecret()
	initCORS()
	if authDisabled() {
		log.Println("⚠️ AUTH_DISABLED=true — JWT auth is off")
	}

	hub = &Hub{
		deliveryPersons: make(map[string]*DeliveryPerson),
		deliveryTasks:   make(map[string]*DeliveryTask),
		trackingClients: make(map[string]*TrackingClient),
		adminClients:    make(map[string]*AdminClient),
		register:        make(chan interface{}),
		unregister:      make(chan interface{}),
		stop:            make(chan struct{}),
	}
	go hub.run()

	port := "3000"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	baseURL = os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:" + port
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(corsMiddleware())

	// Public routes — no auth (customer tracking, health)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy", "timestamp": time.Now().Unix()})
	})
	r.GET("/ws/track/:taskId", handleTrackingWS)
	r.GET("/track/:taskId", handleTrackingPage)
	r.GET("/r/:taskId", handleTrackingRedirect)

	protected := r.Group("/")
	protected.Use(AuthMiddleware())
	{
		protected.GET("/ws/delivery", handleDeliveryWS)
		protected.GET("/ws/admin", handleAdminWS)

		protected.POST("/api/tasks/:id", handleCreateTask)
		protected.GET("/api/tasks/:taskId", handleGetTask)
		protected.GET("/api/tasks", handleGetAllTasks)
		protected.GET("/api/stats", handleGetStats)
		protected.POST("/api/notify-customer", handleNotifyCustomer)
		protected.POST("/api/message-notify", handleMessageNotification)
		protected.GET("/mapbox/search/suggest", handleMapboxSuggest)
		protected.GET("/mapbox/search/retrieve", handleMapboxRetrieve)
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		log.Printf("🚀 Server starting on port %s...", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutdown signal received, draining connections...")
	hub.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout())
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("⚠️ HTTP server shutdown error: %v", err)
	}

	log.Println("✅ Server stopped")
}
