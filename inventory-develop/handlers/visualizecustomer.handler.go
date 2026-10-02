package handlers

import (
	"fmt"
	"inventory/db"
	"inventory/models"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type VisualizeCustomerHandler struct {
	s3Client s3iface.S3API
	region   string
}

func NewVisualizeCustomerHandler(s3Client s3iface.S3API, region string) *VisualizeCustomerHandler {
	return &VisualizeCustomerHandler{
		s3Client: s3Client,
		region:   region,
	}
}

func (h *VisualizeCustomerHandler) generatePresignedURL(bucket, key string) (string, error) {
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

func (h *VisualizeCustomerHandler) GetCustomerByID(c *gin.Context) {
	id := c.Param("id")

	dbObj, exists := c.Get("db")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection not found"})
		return
	}

	database, ok := dbObj.(db.Database)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid database connection state"})
		return
	}

	var customer models.VisualizeCustomer
	if err := database.Where("id = ?", id).First(&customer); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"remaining_tries": customer.UsageCount,
	})
}
func (h *VisualizeCustomerHandler) getBucket() string {
	return "tulip-visualize"
}

func (h *VisualizeCustomerHandler) CreateCustomer(c *gin.Context) {
	const maxTries = 5
	var customer models.VisualizeCustomer

	// Bind incoming JSON payload to the model
	if err := c.ShouldBindJSON(&customer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	// Retrieve the tenant database from the Gin context
	dbObj, exists := c.Get("db")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection not found"})
		return
	}

	database, ok := dbObj.(db.Database)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid database connection state"})
		return
	}

	var existingCustomer models.VisualizeCustomer
	err := database.Where("email = ?", customer.Email).First(&existingCustomer)

	if err == nil {
		remaining := existingCustomer.UsageCount
		if remaining <= 0 {
			c.JSON(http.StatusForbidden, gin.H{"error": "You have used all 5 tries.", "remaining": 0})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message":   "Welcome back",
			"exists":    true,
			"data":      existingCustomer,
			"remaining": remaining,
		})
		return
	}

	// Customer does not exist, create new
	customer.UsageCount = 5
	if err := database.Create(&customer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save customer data"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Customer saved successfully",
		"data":      customer,
		"remaining": maxTries,
	})
}

func (h *VisualizeCustomerHandler) CheckEmailExists(c *gin.Context) {
	email := strings.TrimSpace(c.Query("email"))
	lastName := strings.TrimSpace(c.Query("last_name"))
	phone := strings.TrimSpace(c.Query("phone"))

	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}

	dbObj, exists := c.Get("db")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection not found"})
		return
	}

	database, ok := dbObj.(db.Database)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid database connection state"})
		return
	}

	var customer models.VisualizeCustomer
	err := database.Where("email = ?", email).First(&customer)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"exists": false, "remaining": 5})
		return
	}

	remaining := customer.UsageCount
	if remaining <= 0 {
		c.JSON(http.StatusForbidden, gin.H{"exists": true, "remaining": 0, "error": "You have used all 5 tries."})
		return
	}

	// Check optional combination of last_name and phone if provided
	if lastName != "" && !strings.EqualFold(lastName, customer.LastName) {
		c.JSON(http.StatusConflict, gin.H{"exists": true, "remaining": remaining, "error": "Email is registered with a different last name"})
		return
	}

	if phone != "" && phone != customer.Phone {
		c.JSON(http.StatusConflict, gin.H{"exists": true, "remaining": remaining, "error": "Email is registered with a different phone"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"exists":    true,
		"remaining": remaining,
		"last_name": customer.LastName,
		"phone":     customer.Phone,
		"matches":   true,
	})
}

func (h *VisualizeCustomerHandler) CheckPhoneExists(c *gin.Context) {
	phone := c.Query("phone")
	if phone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Phone is required"})
		return
	}

	dbObj, exists := c.Get("db")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection not found"})
		return
	}

	database, ok := dbObj.(db.Database)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid database connection state"})
		return
	}

	var customer models.VisualizeCustomer
	err := database.Where("phone = ?", phone).First(&customer)
	if err == nil {
		c.JSON(http.StatusOK, gin.H{"exists": true, "error": "Phone number already exists"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"exists": false})
}

func (h *VisualizeCustomerHandler) UpdateCustomerImages(c *gin.Context) {
	id := c.Param("id")

	dbObj, exists := c.Get("db")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection not found"})
		return
	}

	database, ok := dbObj.(db.Database)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid database connection state"})
		return
	}

	var customer models.VisualizeCustomer
	if err := database.Where("id = ?", id).First(&customer); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}

	if customer.UsageCount <= 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "All 5 tries have been used."})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse multipart form"})
		return
	}

	uploadedImages := form.File["uploaded_image"]
	if len(uploadedImages) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "uploaded_image file is required"})
		return
	}

	resultImages := form.File["result_image"]
	if len(resultImages) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "result_image file is required"})
		return
	}

	uploadedImageFile, err := uploadedImages[0].Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open uploaded_image"})
		return
	}
	defer uploadedImageFile.Close()

	resultImageFile, err := resultImages[0].Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open result_image"})
		return
	}
	defer resultImageFile.Close()

	bucket := h.getBucket()

	uploadedFileExt := filepath.Ext(uploadedImages[0].Filename)
	uploadedUniqueId := uuid.New().String()
	uploadedS3Key := fmt.Sprintf("files/%s%s", uploadedUniqueId, uploadedFileExt)

	_, err = h.s3Client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(uploadedS3Key),
		Body:        uploadedImageFile,
		ContentType: aws.String(uploadedImages[0].Header.Get("Content-Type")),
	})

	if err != nil {
		log.Println("s3 upload error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload uploaded_image"})
		return
	}
	uploadedImageURL := uploadedS3Key

	resultFileExt := filepath.Ext(resultImages[0].Filename)
	resultUniqueId := uuid.New().String()
	resultS3Key := fmt.Sprintf("files/%s%s", resultUniqueId, resultFileExt)

	_, err = h.s3Client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(resultS3Key),
		Body:        resultImageFile,
		ContentType: aws.String(resultImages[0].Header.Get("Content-Type")),
	})

	if err != nil {
		log.Println("s3 upload error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload result_image"})
		return
	}
	resultImageURL := resultS3Key

	// Append new image pair to the images array
	newImagePair := models.ImagePair{
		Uploaded: uploadedImageURL,
		Result:   resultImageURL,
	}
	customer.Images = append(customer.Images, newImagePair)

	customer.UsageCount = customer.UsageCount - 1

	if err := database.Save(&customer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save customer images"})
		return
	}

	presignedUploaded, err := h.generatePresignedURL(bucket, uploadedImageURL)
	if err != nil {
		presignedUploaded = uploadedImageURL
	}
	presignedResult, err := h.generatePresignedURL(bucket, resultImageURL)
	if err != nil {
		presignedResult = resultImageURL
	}
	c.JSON(http.StatusOK, gin.H{
		"message":        "Images saved successfully",
		"remaining":      customer.UsageCount,
		"uploaded_image": presignedUploaded,
		"result_image":   presignedResult,
	})

}

func (h *VisualizeCustomerHandler) GetAllCustomers(c *gin.Context) {
	dbObj, exists := c.Get("db")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection not found"})
		return
	}

	database, ok := dbObj.(db.Database)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid database connection state"})
		return
	}

	var customers []models.VisualizeCustomer
	if err := database.Find(&customers); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch customers"})
		return
	}

	bucket := h.getBucket()
	for i := range customers {
		for j := range customers[i].Images {
			if url, err := h.generatePresignedURL(bucket, customers[i].Images[j].Uploaded); err == nil {
				customers[i].Images[j].Uploaded = url
			}
			if url, err := h.generatePresignedURL(bucket, customers[i].Images[j].Result); err == nil {
				customers[i].Images[j].Result = url
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  customers,
		"total": len(customers),
	})
}
