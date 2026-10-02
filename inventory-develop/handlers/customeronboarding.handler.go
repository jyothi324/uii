package handlers

import (
	"fmt"
	"inventory/middleware"
	"inventory/models"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type OnboardingHandler struct {
	s3Client s3iface.S3API
	region   string
}

func NewOnboardingHandler(s3Client s3iface.S3API, region string) *OnboardingHandler {
	return &OnboardingHandler{
		s3Client: s3Client,
		region:   region,
	}
}

func (h *OnboardingHandler) GetBucket(c *gin.Context) string {
	return "tulipfiles"
}

func (h *OnboardingHandler) generatePresignedURL(bucket, key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("empty key")
	}
	if strings.HasPrefix(key, "https://") || strings.HasPrefix(key, "http://") {
		parts := strings.SplitN(key, "/", 4)
		if len(parts) == 4 {
			key = parts[3]
		}
	}
	req, _ := h.s3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return req.Presign(1 * time.Hour)
}

func (h *OnboardingHandler) HandleCreateCustomerOnboarding() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)

		form, err := c.MultipartForm()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse multipart form"})
			return
		}

		// Extract fields matching the keys sent from the frontend state
		companyName := strings.TrimSpace(c.PostForm("companyName"))
		adminName := strings.TrimSpace(c.PostForm("adminName"))
		email := strings.TrimSpace(c.PostForm("email"))
		address := strings.TrimSpace(c.PostForm("address"))
		phone := strings.TrimSpace(c.PostForm("phone"))
		subdomain := strings.TrimSpace(c.PostForm("subdomain"))
		region := strings.TrimSpace(c.PostForm("region"))
		timezone := strings.TrimSpace(c.PostForm("timezone"))
		dbInstanceID := strings.TrimSpace(c.PostForm("dbInstanceId"))
		deploymentStatus := strings.TrimSpace(c.PostForm("deploymentStatus"))

		usersStr := c.PostForm("users")
		var users int
		if usersStr != "" {
			users, _ = strconv.Atoi(usersStr)
		}

		provisioningDateStr := c.PostForm("provisioningDate")
		var provisioningDate time.Time
		if provisioningDateStr != "" {
			// Date format from HTML5 input type="date" is YYYY-MM-DD
			provisioningDate, _ = time.Parse("2006-01-02", provisioningDateStr)
		}

		// Handle Logo Upload to S3
		var logoURL string
		logos := form.File["logo"]
		if len(logos) > 0 {
			fileHeader := logos[0]
			file, err := fileHeader.Open()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open logo file"})
				return
			}
			defer file.Close()

			fileExt := filepath.Ext(fileHeader.Filename)
			s3Key := fmt.Sprintf("onboarding/logos/%s%s", uuid.New().String(), fileExt)
			bucket := h.GetBucket(c)

			_, err = h.s3Client.PutObject(&s3.PutObjectInput{
				Bucket:      aws.String(bucket),
				Key:         aws.String(s3Key),
				Body:        file,
				ContentType: aws.String(fileHeader.Header.Get("Content-Type")),
			})
			if err != nil {
				log.Println("Logo S3 upload error:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload logo"})
				return
			}
			logoURL = s3Key
		}

		onboarding := models.CustomerOnboarding{
			CompanyName:      companyName,
			AdminName:        adminName,
			Email:            email,
			Address:          address,
			Phone:            phone,
			Subdomain:        subdomain,
			LogoURL:          logoURL,
			Region:           region,
			Timezone:         timezone,
			Users:            users,
			DBInstanceID:     dbInstanceID,
			ProvisioningDate: provisioningDate,
			DeploymentStatus: deploymentStatus,
		}

		if err := database.Create(&onboarding); err != nil {
			log.Println("Create onboarding error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create customer onboarding record"})
			return
		}

		if onboarding.LogoURL != "" {
			if url, err := h.generatePresignedURL(h.GetBucket(c), onboarding.LogoURL); err == nil {
				onboarding.LogoURL = url
			}
		}
		c.JSON(http.StatusCreated, onboarding)
	}
}
func (h *OnboardingHandler) HandleGetAllCustomerOnboardings() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)

		var onboardings []models.CustomerOnboarding
		if err := database.Find(&onboardings); err != nil {
			log.Println("Get all onboardings error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve customer onboarding records"})
			return
		}

		for i := range onboardings {
			if onboardings[i].LogoURL != "" {
				if url, err := h.generatePresignedURL(h.GetBucket(c), onboardings[i].LogoURL); err == nil {
					onboardings[i].LogoURL = url
				}
			}
		}
		c.JSON(http.StatusOK, onboardings)
	}
}
func (h *OnboardingHandler) HandleUpdateCustomerOnboarding() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)

		id := c.Param("id")

		var onboarding models.CustomerOnboarding
		if err := database.First(&onboarding, id); err != nil {
			log.Println("Update onboarding error:", err)
			c.JSON(http.StatusNotFound, gin.H{"error": "Customer onboarding not found"})
			return
		}

		form, err := c.MultipartForm()
		if err == nil && form != nil {
			// Handle Logo Update if a new file is provided
			logos := form.File["logo"]
			if len(logos) > 0 {
				fileHeader := logos[0]
				file, err := fileHeader.Open()
				if err == nil {
					defer file.Close()
					fileExt := filepath.Ext(fileHeader.Filename)
					s3Key := fmt.Sprintf("onboarding/logos/%s%s", uuid.New().String(), fileExt)
					bucket := h.GetBucket(c)

					_, err = h.s3Client.PutObject(&s3.PutObjectInput{
						Bucket:      aws.String(bucket),
						Key:         aws.String(s3Key),
						Body:        file,
						ContentType: aws.String(fileHeader.Header.Get("Content-Type")),
					})
					if err != nil {
						log.Println("Logo S3 upload error during update:", err)
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload new logo"})
						return
					}
					onboarding.LogoURL = s3Key
				}
			}
		}

		// Update fields based on form data
		if companyName := c.PostForm("companyName"); companyName != "" {
			onboarding.CompanyName = companyName
		}
		if adminName := c.PostForm("adminName"); adminName != "" {
			onboarding.AdminName = adminName
		}
		if email := c.PostForm("email"); email != "" {
			onboarding.Email = email
		}
		if address := c.PostForm("address"); address != "" {
			onboarding.Address = address
		}
		if phone := c.PostForm("phone"); phone != "" {
			onboarding.Phone = phone
		}
		if subdomain := c.PostForm("subdomain"); subdomain != "" {
			onboarding.Subdomain = subdomain
		}
		if region := c.PostForm("region"); region != "" {
			onboarding.Region = region
		}
		if timezone := c.PostForm("timezone"); timezone != "" {
			onboarding.Timezone = timezone
		}
		if usersStr := c.PostForm("users"); usersStr != "" {
			users, _ := strconv.Atoi(usersStr)
			onboarding.Users = users
		}
		if dbInstanceID := c.PostForm("dbInstanceId"); dbInstanceID != "" { // Match frontend key
			onboarding.DBInstanceID = dbInstanceID
		}
		if provisioningDateStr := c.PostForm("provisioningDate"); provisioningDateStr != "" {
			provisioningDate, _ := time.Parse("2006-01-02", provisioningDateStr)
			onboarding.ProvisioningDate = provisioningDate
		}
		if deploymentStatus := c.PostForm("deploymentStatus"); deploymentStatus != "" {
			onboarding.DeploymentStatus = deploymentStatus
		}

		if err := database.Save(&onboarding); err != nil {
			log.Println("Save onboarding error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update customer onboarding record"})
			return
		}

		if onboarding.LogoURL != "" {
			if url, err := h.generatePresignedURL(h.GetBucket(c), onboarding.LogoURL); err == nil {
				onboarding.LogoURL = url
			}
		}
		c.JSON(http.StatusOK, onboarding)
	}
}
