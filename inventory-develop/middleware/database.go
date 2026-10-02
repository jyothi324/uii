package middleware

import (
	"log"
	"net/http"
	"strings"

	"inventory/db"

	"github.com/gin-gonic/gin"
)

const (
	TenantContextKey = "tenant"
	DBContextKey     = "db"
)

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
		tenant := extractTenantFromURL(origin)
		if tenant != "" && strings.Contains(origin, "localhost") || strings.Contains(origin, "tulipcrm.com") {
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
	if subdomain == "" {
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
	if len(parts) == 1 {
		return ""
	}
	if parts[0] == "" {
		return ""
	}
	return sanitizeTenant(parts[0]) 
}
func GetTenant(c *gin.Context) string {
	tenant, exists := c.Get(TenantContextKey)
	if !exists {
		return "default"
	}
	return tenant.(string)
}

func DatabaseMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
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
			log.Printf("Database connection failed for tenant %s: %v", tenant, err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Database connection failed",
			})
			c.Abort()
			return
		}

		c.Set(DBContextKey, tenantDB)
		c.Next()
	}
}

func GetDB(c *gin.Context) db.Database {
	return c.MustGet(DBContextKey).(db.Database)
}