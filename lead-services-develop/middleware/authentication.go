package middleware

import (
	"net/http"
	"strings"
	"time"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var (
	jwtSecretOnce  sync.Once
	jwtSecretBytes []byte
)

func getJWTSecret() []byte {
	jwtSecretOnce.Do(func() {
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			panic("JWT_SECRET environment variable is required")
		}
		jwtSecretBytes = []byte(secret)
	})
	return jwtSecretBytes
}


// tenant
type CustomClaims struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

const (
	TenantContextKey = "tenant"
	DBContextKey     = "db"
)
const TokenExpiry = 12 * time.Hour

func GenerateJWT(userID, name, email, role string) (string, error) {
	claims := CustomClaims{
		UserID: userID,
		Name:   name,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "tulip-admin",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJWTSecret())
}

func AuthMiddleware(c *gin.Context) {
	if c.Request.Method == "OPTIONS" {
		c.Next()
		return
	}
	var tokenStr string
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
	} else {
		cookie, err := c.Cookie("token")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized - no token"})
			c.Abort()
			return
		}
		tokenStr = cookie
	}
	token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return getJWTSecret(), nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		c.Abort()
		return
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		c.Abort()
		return
	}

	c.Set("user_id", claims.UserID)
	c.Set("user_name", claims.Name)
	c.Set("email", claims.Email)
	c.Set("role", claims.Role)

	c.Next()
}

func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")

		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: role not found"})
			return
		}

		if userRole == "MANAGER" || userRole == "AGENT" {
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: insufficient role"})
	}
}
func TenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenant := extractTenantFromRequest(c)
		c.Set(TenantContextKey, tenant)
		c.Next()
	}
}

func extractTenantFromRequest(c *gin.Context) string {
	origin := c.GetHeader("X-Tenant")
	if origin != "" {
		// If X-Tenant is a direct identifier (no dots or protocol), use it as the tenant
		if !strings.Contains(origin, ".") && !strings.Contains(origin, "://") {
			return normalizeTenantName(sanitizeTenant(origin))
		}
		tenant := extractTenantFromURL(origin)
		if tenant != "" && (strings.Contains(origin, "localhost") || strings.Contains(origin, "tulipcrm.com")) {
			return tenant
		}
	}
	if referer := c.GetHeader("Referer"); referer != "" {
		tenant := extractTenantFromURL(referer)
		if tenant != "" {
			return tenant
		}
	}
	if referer := c.GetHeader("Origin"); referer != "" {
		tenant := extractTenantFromURL(referer)
		if tenant != "" {
			return tenant
		}
	}
	host := c.Request.Host
	return extractTenantFromHost(host)
}
func extractTenantFromURL(urlStr string) string {
	urlStr = strings.TrimPrefix(urlStr, "http://")
	urlStr = strings.TrimPrefix(urlStr, "https://")
	host := urlStr
	if idx := strings.Index(urlStr, "/"); idx != -1 {
		host = urlStr[:idx]
	}
	return extractTenantFromHost(host)
}
func extractTenantFromHost(host string) string {
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}
	parts := strings.Split(host, ".")
	if len(parts) == 1 {
		return "default"
	}
	subdomain := parts[0]
	if subdomain == "" || subdomain == "www" {
		return "default"
	}
	return normalizeTenantName(sanitizeTenant(subdomain))
}
func sanitizeTenant(tenant string) string {
	tenant = strings.ToLower(tenant)
	var result strings.Builder
	for _, ch := range tenant {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-' || ch == '_' {
			result.WriteRune(ch)
		}
	}
	sanitized := result.String()
	if sanitized == "" {
		return "default"
	}
	return sanitized
}
var envSuffixes = []string{
	"staging", "stage", "production", "prod",
	"dev", "test", "uat", "qa", "local",
}

func normalizeTenantName(sub string) string {
	sub = strings.TrimSuffix(sub, "-")
	sub = strings.TrimSuffix(sub, "_")

	for _, suf := range envSuffixes {
		if strings.HasSuffix(sub, suf) && len(sub) > len(suf) {
			trimmed := strings.TrimSuffix(sub, suf)
			trimmed = strings.TrimRight(trimmed, "-_")
			if trimmed != "" {
				return trimmed
			}
		}
	}
	return sub
}
func GetTenant(c *gin.Context) string {
	tenant, exists := c.Get(TenantContextKey)
	if !exists {
		return "default"
	}
	return tenant.(string)
}

// GetRequestUser extracts user ID and name from Gin context or request headers.
// It prioritizes context values set by AuthMiddleware, then falls back to common headers.
// Returns "system" and "System" if no user information is found.
func GetRequestUser(c *gin.Context) (string, string) {
	userID, userName := "", ""

	// Priority 1: Check Gin Context (populated by AuthMiddleware)
	if val, ok := c.Get("user_id"); ok {
		userID, _ = val.(string)
	}
	if val, ok := c.Get("user_name"); ok {
		userName, _ = val.(string)
	}
	if userName == "" {
		if val, ok := c.Get("email"); ok {
			userName, _ = val.(string)
		}
	}

	// Priority 2: Check standard and custom headers (resilience for proxies)
	if userID == "" {
		userID = c.GetHeader("X-User-Id")
		if userID == "" {
			userID = c.GetHeader("X-User-ID")
		}
		if userID == "" {
			userID = c.GetHeader("x-user-id")
		}
		if userID == "" {
			userID = c.GetHeader("X-Agent-ID")
		}
	}
	if userName == "" {
		userName = c.GetHeader("X-User-Name")
		if userName == "" {
			userName = c.GetHeader("x-user-name")
		}
		if userName == "" {
			userName = c.GetHeader("User-Email")
		}
	}

	if userID == "" {
		userID = "system"
	}
	if userName == "" {
		userName = "System"
	}

	return userID, userName
}

func GetRawSubdomain(c *gin.Context) string {
	if origin := c.GetHeader("X-Tenant"); origin != "" {
		if sub := extractRawSubdomainFromURL(origin); sub != "" {
			return sub
		}
	}
	if referer := c.GetHeader("Referer"); referer != "" {
		if sub := extractRawSubdomainFromURL(referer); sub != "" {
			return sub
		}
	}
	if origin := c.GetHeader("Origin"); origin != "" {
		if sub := extractRawSubdomainFromURL(origin); sub != "" {
			return sub
		}
	}
	return extractRawSubdomainFromHost(c.Request.Host)
}

func extractRawSubdomainFromURL(urlStr string) string {
	urlStr = strings.TrimPrefix(urlStr, "http://")
	urlStr = strings.TrimPrefix(urlStr, "https://")
	host := urlStr
	if idx := strings.Index(urlStr, "/"); idx != -1 {
		host = urlStr[:idx]
	}
	return extractRawSubdomainFromHost(host)
}

func extractRawSubdomainFromHost(host string) string {
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}
	parts := strings.Split(host, ".")
	if len(parts) == 1 || parts[0] == "" {
		return ""
	}
	return sanitizeTenant(parts[0])
}