package handlers

import (
	"errors"
	"inventory/middleware"
	"inventory/models"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *OnboardingHandler) GetTenantBrand() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)

		var brand models.TenantBrand
		err := database.First(&brand)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusOK, gin.H{
					"found":        false,
					"company_name": "",
					"logo_url":     "",
					"subdomain":    "",
					"address":      models.TenantAddress{},
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch brand"})
			return
		}

		bucket := h.GetBucket(c)
		if url, err := h.generatePresignedURL(bucket, brand.LogoURL); err == nil {
			// Make sure the URL has proper & characters
			brand.LogoURL = strings.ReplaceAll(url, "\\u0026", "&")
		}
		c.JSON(http.StatusOK, gin.H{
			"found":        true,
			"company_name": brand.CompanyName,
			"logo_url":     brand.LogoURL,
			"subdomain":    brand.Subdomain,
			"address":      brand.Address,
		})
	}
}

func (h *OnboardingHandler) SaveTenantAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)
		if database == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection unavailable"})
			return
		}

		var req struct {
			Address     *models.TenantAddress `json:"address"`
			CompanyName *string                `json:"companyName"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		companyNameOnly := req.CompanyName != nil && req.Address == nil
		if !companyNameOnly && req.Address == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "address is required"})
			return
		}

		if companyNameOnly {
			trimmedName := strings.TrimSpace(*req.CompanyName)
			req.CompanyName = &trimmedName
		} else {
			req.Address.CompanyName = strings.TrimSpace(req.Address.CompanyName)
			req.Address.Street = strings.TrimSpace(req.Address.Street)
			req.Address.UnitSuite = strings.TrimSpace(req.Address.UnitSuite)
			req.Address.City = strings.TrimSpace(req.Address.City)
			req.Address.State = strings.TrimSpace(req.Address.State)
			req.Address.Zip = strings.TrimSpace(req.Address.Zip)
			req.Address.Country = strings.TrimSpace(req.Address.Country)
			req.Address.Website = strings.TrimSpace(req.Address.Website)
			req.Address.Email = strings.TrimSpace(req.Address.Email)
			req.Address.Contact = strings.TrimSpace(req.Address.Contact)
			req.Address.Latitude = strings.TrimSpace(req.Address.Latitude)
			req.Address.Longitude = strings.TrimSpace(req.Address.Longitude)
		}

		var brand models.TenantBrand
		err := database.First(&brand)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch brand"})
				return
			}

			subdomain := strings.TrimSpace(c.GetHeader("X-Tenant"))
			if subdomain == "" {
				subdomain = "app"
			}
			newAddress := models.TenantAddress{}
			if companyNameOnly {
				newAddress.CompanyName = *req.CompanyName
			} else {
				newAddress = *req.Address
			}
			brand = models.TenantBrand{
				CompanyName: subdomain,
				Subdomain:   subdomain,
				Address:     newAddress,
			}
			if err := database.Create(&brand); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save address"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "address saved successfully", "address": brand.Address})
			return
		}

		if companyNameOnly {
			brand.Address.CompanyName = *req.CompanyName
		} else {
			brand.Address = *req.Address
		}

		if err := database.Save(&brand); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update address"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "address saved successfully", "address": brand.Address})
	}
}

func (h *OnboardingHandler) GetTenantAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)

		var brand models.TenantBrand
		err := database.First(&brand)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusOK, gin.H{"found": false, "address": models.TenantAddress{}})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch address"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"found": true, "address": brand.Address})
	}
}

// ProxyLogo fetches an S3 image server-side and streams it back,
// avoiding browser CORS restrictions on direct S3 fetch.
func (h *OnboardingHandler) ProxyLogo() gin.HandlerFunc {
	return func(c *gin.Context) {
		logoURL := c.Query("url")
		if logoURL == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing url"})
			return
		}

		resp, err := http.Get(logoURL)
		if err != nil || resp.StatusCode != http.StatusOK {
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to fetch logo"})
			return
		}
		defer resp.Body.Close()

		c.Header("Content-Type", resp.Header.Get("Content-Type"))
		c.Header("Cache-Control", "public, max-age=3600")
		io.Copy(c.Writer, resp.Body)
	}
}
