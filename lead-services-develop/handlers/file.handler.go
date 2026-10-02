package handlers

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/tuliporg/agent-management-api/middleware"
	"github.com/tuliporg/agent-management-api/models"
	"gorm.io/gorm"
)

type FileHandler struct {
	s3Client s3iface.S3API
	bucket   string
	region   string
}

func NewFileHandler(s3Client s3iface.S3API, bucket, region string) *FileHandler {

	return &FileHandler{
		s3Client: s3Client,
		bucket:   bucket,
		region:   region,
	}
}

func (h *FileHandler) generatePresignedURL(c *gin.Context, s3Key string) (string, error) {
	if s3Key == "" {
		return "", fmt.Errorf("empty s3Key")
	}
	// Handle legacy full URLs stored before the migration
	if strings.HasPrefix(s3Key, "https://") || strings.HasPrefix(s3Key, "http://") {
		parts := strings.Split(s3Key, "/")
		if len(parts) >= 4 {
			s3Key = strings.Join(parts[3:], "/")
		}
	}
	req, _ := h.s3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket:                     aws.String(h.getBucketName(c)),
		Key:                        aws.String(s3Key),
		ResponseContentDisposition: attachmentDispositionForKey(s3Key),
	})
	return req.Presign(1 * time.Hour)
}

// attachmentDispositionForKey forces the browser to DOWNLOAD (never inline-render)
// file types that can execute script when served inline from S3 — chiefly SVG and
// HTML/XML. Returning nil keeps the default inline behavior for safe types like
// images and PDFs, so their preview/open UX is unchanged. This neutralizes stored
// XSS from uploaded attachments that are later opened via window.open / navigation.
func attachmentDispositionForKey(s3Key string) *string {
	lower := strings.ToLower(s3Key)
	for _, ext := range []string{".svg", ".svgz", ".html", ".htm", ".xhtml", ".xml"} {
		if strings.HasSuffix(lower, ext) {
			return aws.String("attachment")
		}
	}
	return nil
}

func (h *FileHandler) getBucketName(c *gin.Context) string {
	subdomain := middleware.GetRawSubdomain(c)
	if subdomain == "" || subdomain == "default" {
		return h.bucket
	}
	return fmt.Sprintf("tulip-%s-files", subdomain)
}

func (h *FileHandler) CreateFile(c *gin.Context) {
	db := middleware.GetDB(c)
	bucket := h.getBucketName(c)
	category := c.PostForm("category")
	projecttype := c.PostForm("type")
	leadID := c.PostForm("lead_id")
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse multipart form"})
		return
	}
	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No files uploaded"})
		return
	}
	var uploadedFiles []models.File
	for _, fileHeader := range files {
		file1, err := fileHeader.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to open file",
				Code:    "FILE_ERROR",
				Details: err.Error(),
			})
			return
		}
		// Read file content into memory first
		// fileContent, err := io.ReadAll(file)
		if err != nil {
			file1.Close()
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to read file content",
				Code:    "FILE_ERROR",
				Details: err.Error(),
			})
			return
		}
		fileExt := filepath.Ext(fileHeader.Filename)
		uniqueID := uuid.New().String()
		s3Key := fmt.Sprintf("files/%s%s", uniqueID, fileExt)
		storedFileName := fmt.Sprintf("%s%s", uniqueID, fileExt)
		// Upload to S3
		_, err = h.s3Client.PutObject(&s3.PutObjectInput{
			Bucket:             aws.String(bucket),
			Key:                aws.String(s3Key),
			Body:               file1,
			ContentDisposition: aws.String("attachment"),
			ContentType:        aws.String(fileHeader.Header.Get("Content-Type")),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file"})
			return
		}
		// url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucket, h.region, s3Key)
		file := models.File{
			ID:             uint(rand.Uint32()),
			Name:           fileHeader.Filename,
			Type:           fileHeader.Header.Get("Content-Type"),
			Size:           fmt.Sprintf("%d", fileHeader.Size),
			UploadedAt:     time.Now(),
			StoredFileName: storedFileName,
			Category:       category,
			ProjectType:    projecttype,
			FilePath:       s3Key,
		}
		if leadID != "" {
			file.LeadID = &leadID
		}
		if err := db.Create(&file); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("DB save failed: %v", err)})
			return
		}
	}
	// Return uploaded file metadata
	c.JSON(http.StatusOK, gin.H{
		"uploadedFiles": uploadedFiles,
	})
}

// GetFile retrieves a File by ID.
func (h *FileHandler) GetFile(c *gin.Context) {
	db := middleware.GetDB(c)
 
	leadID := c.Param("id")
	var files []models.File
	result := db.Preload("Lead").Where("lead_id = ?", leadID).Find(&files)
	if result != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Database error",
			Code:    "DATABASE_ERROR",
			Details: result.Error(),
		})
		return
	}
	if len(files) == 0 {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "No files exist for this lead",
			Code:    "NO_FILES_FOUND",
			Details: "No file is associated with the given lead ID",
		})
		return
	}
 
	// Enrich each file with a fresh presigned URL
	for i := range files {
		if files[i].FilePath == "" {
			continue
		}
		url, err := h.generatePresignedURL(c, files[i].FilePath)
		if err == nil {
			files[i].PresignedURL = url
		}
	}
 
	c.JSON(http.StatusOK, gin.H{
		"data": files,
	})
}
func (h *FileHandler) UpdateFileCategory(c *gin.Context) {
	db := middleware.GetDB(c)

	id := c.Param("id")
	var file models.File
	if err := db.First(&file, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "File not found",
				Code:    "NOT_FOUND",
				Details: "No file found with the provided ID",
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Database error",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
		}
		return
	}
	// Accept only ProjectType and Category fields
	var input struct {
		ProjectType string
		Category    string
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid input",
			Code:    "VALIDATION_ERROR",
			Details: err.Error(),
		})
		return
	}
	// Update only allowed fields
	if input.ProjectType != "" {
		file.ProjectType = input.ProjectType
	}
	if input.Category != "" {
		file.Category = input.Category
	}
	if err := db.Save(&file); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to update file",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "File updated successfully",
		"data":    file,
	})
}

// UpdateFile updates an existing File.
func (h *FileHandler) UpdateFile(c *gin.Context) {
	db := middleware.GetDB(c)

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid ID",
			Code:    "VALIDATION_ERROR",
			Details: "ID must be a valid number",
		})
		return
	}

	var file models.File
	if err := db.First(&file, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "File not found",
				Code:    "NOT_FOUND",
				Details: "No file found with the provided ID",
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Database error",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
		}
		return
	}

	// Only accept the fields a rename operation needs
	var input struct {
		LeadID      *string   `json:"LeadID,omitempty"`
		Name        string `json:"Name"`
		Category    string `json:"Category"`
		ProjectType string `json:"ProjectType"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid input",
			Code:    "VALIDATION_ERROR",
			Details: err.Error(),
		})
		return
	}

	// Only Name is required
	if input.Name == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Missing required fields",
			Code:    "VALIDATION_ERROR",
			Details: "Name is required",
		})
		return
	}
		// Validate LeadID if provided
	if input.LeadID != nil {
		var lead models.Lead
		if err := db.First(&lead, input.LeadID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusBadRequest, ErrorResponse{
					Error:   "Invalid LeadID",
					Code:    "VALIDATION_ERROR",
					Details: "Lead not found",
				})
			} else {
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "Database error",
					Code:    "DATABASE_ERROR",
					Details: "An error occurred while querying the database",
				})
			}
			return
		}
	}

	// Partial update — StoredFileName, FilePath, Type, Size, LeadID stay untouched
	file.Name = input.Name
	if input.Category != "" {
		file.Category = input.Category
	}
	if input.ProjectType != "" {
		file.ProjectType = input.ProjectType
	}

	if err := db.Save(&file); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to update file",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "File updated successfully",
		"data":    file,
	})
}
func (h *FileHandler) DeleteFile(c *gin.Context) {
	db := middleware.GetDB(c)

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid ID",
			Code:    "VALIDATION_ERROR",
			Details: "ID must be a valid number",
		})
		return
	}

	if err := db.Delete(&models.File{}, id); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to delete file",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "File deleted successfully",
	})
}
