package websocket

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	gorillawebsocket "github.com/gorilla/websocket"
)

var upgrader = gorillawebsocket.Upgrader{
	ReadBufferSize:   1024,
	WriteBufferSize:  1024,
	HandshakeTimeout: 10 * time.Second,
	CheckOrigin: func(r *http.Request) bool {
		return validateOrigin(r)
	},
}

func validateOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}

	allowedOrigins := os.Getenv("ACTIVITY_WS_ALLOWED_ORIGINS")
	if strings.TrimSpace(allowedOrigins) == "*" {
		return true
	}

	if allowedOrigins != "" {
		for _, candidate := range strings.Split(allowedOrigins, ",") {
			if strings.EqualFold(strings.TrimSpace(candidate), origin) {
				return true
			}
		}
	}

	parsed, err := url.Parse(origin)
	if err != nil {
		log.Printf("websocket origin parse failed: %q, err=%v", origin, err)
		return false
	}

	host := strings.ToLower(parsed.Hostname())
	if host == "localhost" || host == "127.0.0.1" || strings.HasSuffix(host, ".localhost") {
		return true
	}

	if host == "tulipcrm.com" || strings.HasSuffix(host, ".tulipcrm.com") {
		return true
	}

	log.Printf("websocket origin rejected: %q, allowed=%q", origin, allowedOrigins)
	return false
}

func ServeActivityWebSocket(c *gin.Context, manager *Manager) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("websocket upgrade failed remote=%s origin=%q error=%v", c.ClientIP(), c.Request.Header.Get("Origin"), err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	tenant := c.Query("tenant")
	if tenant == "" {
		// Fallback to subdomain extraction if not provided in the query string,
		// matching the logic used in the Chat Server's extractTenantFromHost.
		host := c.Request.Host
		if colon := strings.Index(host, ":"); colon != -1 {
			host = host[:colon]
		}
		parts := strings.Split(host, ".")
		// Matches tenant.tulipcrm.com (parts >= 3) or tenant.localhost (parts >= 2)
		if len(parts) >= 3 || (len(parts) >= 2 && parts[len(parts)-1] == "localhost") {
			tenant = parts[0]
		} else {
			tenant = "default"
		}
	}

	log.Printf("websocket client connected remote=%s origin=%q tenant=%q path=%s", conn.RemoteAddr(), c.Request.Header.Get("Origin"), tenant, c.Request.URL.Path)
	client := newClient(conn, manager, tenant)
	client.start()
}
