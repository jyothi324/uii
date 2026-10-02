package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"inventory/middleware"
	"inventory/models"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type PurchaseReceivedHandler struct {
	s3Client s3iface.S3API
	region   string
}

func NewPurchaseReceivedHandler(s3Client s3iface.S3API, region string) *PurchaseReceivedHandler {
	return &PurchaseReceivedHandler{s3Client: s3Client, region: region}
}

func (h *PurchaseReceivedHandler) getBucket(c *gin.Context) string {
	subdomain := middleware.GetRawSubdomain(c)
	if subdomain == "" {
		return "tulip-app-files"
	}
	return fmt.Sprintf("tulip-%s-files", subdomain)
}

func (h *PurchaseReceivedHandler) generatePresignedURL(bucket, key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("empty key")
	}
	req, _ := h.s3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return req.Presign(1 * time.Hour)
}

func (h *PurchaseReceivedHandler) UploadImage(c *gin.Context, fileKey string) (string, error) {
	file, header, err := c.Request.FormFile(fileKey)
	if err != nil {
		return "", nil
	}
	defer file.Close()
	bucket := h.getBucket(c)
	s3Key := fmt.Sprintf("purchase-received/%s%s", uuid.New().String(), filepath.Ext(header.Filename))
	_, err = h.s3Client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(s3Key),
		Body:        file,
		ContentType: aws.String(header.Header.Get("Content-Type")),
	})
	if err != nil {
		return "", fmt.Errorf("s3 upload failed: %w", err)
	}
	return s3Key, nil
}

func enrichPurchaseReceivedImage(h *PurchaseReceivedHandler, c *gin.Context, pr *models.PurchaseReceived) {
	if pr.Image != "" {
		bucket := h.getBucket(c)
		if url, err := h.generatePresignedURL(bucket, pr.Image); err == nil {
			pr.Image = url
		}
	}
}

// enrichPRImageStandalone generates a presigned URL for a PurchaseReceived image
func EnrichPRImageURL(imageKey, bucket string, presignFn func(string, string) (string, error)) string {
	if imageKey == "" {
		return ""
	}
	if url, err := presignFn(bucket, imageKey); err == nil {
		return url
	}
	return imageKey
}

func HandleCreatePurchaseReceived() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)
		var req models.PurchaseReceived
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		var purchase models.Purchase
		if err := database.First(&purchase, req.PurchaseID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Purchase not found"})
				return
			}
			log.Println("fetch purchase error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify purchase"})
			return
		}
		now := time.Now()
		req.CreatedAt = now
		req.UpdatedAt = now
		if err := database.Create(&req); err != nil {
			log.Println("insert purchase received error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create purchase received"})
			return
		}

		payload, _ := json.Marshal(map[string]interface{}{
			"purchase_id": req.PurchaseID,
			"invoice_id":  req.InvoiceID,
			"price":       req.Price,
			"quantity":    req.ReceivedQuantity,
		})
		activity := ActivityEvent{
			EntityType: "purchase-received",
			EntityID:   fmt.Sprintf("%d", req.ID),
			Action:     "created",
			Payload:    datatypes.JSON(payload),
			CreatedAt:  now,
		}
		database.Create(&activity)
		eventJSON, _ := json.Marshal(activity)
		database.Exec("SELECT pg_notify('activity_events_channel', ?)", string(eventJSON))

		c.JSON(http.StatusCreated, gin.H{
			"id":           req.ID,
			"purchase_id":  req.PurchaseID,
			"invoice_id":   req.InvoiceID,
			"price":        req.Price,
			"message":      "Purchase received created successfully",
		})
	}
}

// HandleCreatePurchaseReceivedWithImage
func (h *PurchaseReceivedHandler) HandleCreatePurchaseReceivedWithImage() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)

		purchaseIDStr := c.PostForm("purchase_id")
		invoiceID := c.PostForm("invoice_id")
		priceStr := c.PostForm("Price")
		receivedQtyStr := c.PostForm("received_quantity")

		var purchaseID uint
		fmt.Sscanf(purchaseIDStr, "%d", &purchaseID)
		var price float64
		fmt.Sscanf(priceStr, "%f", &price)
		var receivedQty float64
		fmt.Sscanf(receivedQtyStr, "%f", &receivedQty)

		var purchase models.Purchase
		if err := database.First(&purchase, purchaseID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Purchase not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify purchase"})
			return
		}

		// Upload image if provided
		imageKey, err := h.UploadImage(c, "image")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
			return
		}

		now := time.Now()
		pr := models.PurchaseReceived{
			PurchaseID:       purchaseID,
			InvoiceID:        invoiceID,
			Price:            price,
			ReceivedQuantity: receivedQty,
			Image:            imageKey,
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		if err := database.Create(&pr); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create purchase received"})
			return
		}

		payload, _ := json.Marshal(map[string]interface{}{
			"purchase_id":       purchaseID,
			"invoice_id":        invoiceID,
			"price":             price,
			"received_quantity": receivedQty,
		})
		activity := ActivityEvent{
			EntityType: "purchase-received",
			EntityID:   fmt.Sprintf("%d", pr.ID),
			Action:     "created",
			Payload:    datatypes.JSON(payload),
			CreatedAt:  now,
		}
		database.Create(&activity)
		eventJSON, _ := json.Marshal(activity)
		database.Exec("SELECT pg_notify('activity_events_channel', ?)", string(eventJSON))

		enrichPurchaseReceivedImage(h, c, &pr)
		c.JSON(http.StatusCreated, gin.H{
			"id":                pr.ID,
			"purchase_id":       pr.PurchaseID,
			"invoice_id":        pr.InvoiceID,
			"price":             pr.Price,
			"received_quantity": pr.ReceivedQuantity,
			"image":             pr.Image,
			"message":           "Purchase received created successfully",
		})
	}
}

// handlers/purchase_received.go
func (h *PurchaseReceivedHandler) HandleFetchPurchaseReceiveds() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)
		var purchaseReceiveds []models.PurchaseReceived
		if err := database.Preload("Purchase").Preload("Purchase.Materials").Preload("Purchase.Supplier").Find(&purchaseReceiveds); err != nil {
			log.Println("fetch purchase received error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch purchase received"})
			return
		}

		bucket := h.getBucket(c)
		for i := range purchaseReceiveds {
			// Presign the top-level received image (existing behavior)
			if purchaseReceiveds[i].Image != "" {
				if url, err := h.generatePresignedURL(bucket, purchaseReceiveds[i].Image); err == nil {
					purchaseReceiveds[i].Image = url
				}
			}
			// ── NEW: presign the nested material's pictures too ──
			h.presignMaterialPictures(c, &purchaseReceiveds[i].Purchase.Materials)
		}

		c.JSON(http.StatusOK, purchaseReceiveds)
		log.Printf("Fetched %d purchase received", len(purchaseReceiveds))
	}
}

func (h *PurchaseReceivedHandler) HandlePurchaseReceivedGetByID() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing id parameter"})
			return
		}
		var purchaseReceived models.PurchaseReceived
		err := database.Preload("Purchase").Preload("Purchase.Materials").Preload("Purchase.Supplier").First(&purchaseReceived, id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusOK, gin.H{"message": "Purchase received not found", "data": []interface{}{}})
				return
			}
			log.Println("fetch purchase received error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch purchase received"})
			return
		}

		// Presign the top-level received image
		if purchaseReceived.Image != "" {
			bucket := h.getBucket(c)
			if url, err := h.generatePresignedURL(bucket, purchaseReceived.Image); err == nil {
				purchaseReceived.Image = url
			}
		}

		// ── NEW: presign the nested material's pictures too ──
		h.presignMaterialPictures(c, &purchaseReceived.Purchase.Materials)

		c.JSON(http.StatusOK, purchaseReceived)
	}
}

func HandlePurchaseReceivedUpdate() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing id parameter"})
			return
		}
		var req models.PurchaseReceived
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		var existing models.PurchaseReceived
		if err := database.First(&existing, id); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusOK, gin.H{"message": "Purchase received not found", "data": []interface{}{}})
				return
			}
			log.Println("fetch purchase received error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch purchase received"})
			return
		}
		req.ID = existing.ID
		req.Image = existing.Image // preserve existing image
		req.CreatedAt = existing.CreatedAt
		req.UpdatedAt = time.Now()
		if err := database.Save(&req); err != nil {
			log.Println("update purchase received error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update purchase received"})
			return
		}

		payload, _ := json.Marshal(map[string]interface{}{
			"purchase_id": req.PurchaseID,
			"invoice_id":  req.InvoiceID,
		})
		activity := ActivityEvent{
			EntityType: "purchase-received",
			EntityID:   fmt.Sprintf("%d", req.ID),
			Action:     "updated",
			Payload:    datatypes.JSON(payload),
			CreatedAt:  time.Now(),
		}
		if err := database.Create(&activity); err != nil {
			log.Printf("failed to log purchase received update: %v", err)
		} else {
			eventJSON, _ := json.Marshal(activity)
			database.Exec("SELECT pg_notify('activity_events_channel', ?)", string(eventJSON))
		}

		c.JSON(http.StatusOK, gin.H{"message": "Purchase Received updated successfully"})
	}
}

// presignMaterialPictures replaces raw S3 keys in a Material's Pictures
// slice with fresh presigned URLs, in place.
func (h *PurchaseReceivedHandler) presignMaterialPictures(c *gin.Context, mat *models.Materials) {
	if mat == nil || len(mat.Pictures) == 0 {
		return
	}
	bucket := h.getBucket(c)
	presigned := make([]string, 0, len(mat.Pictures))
	for _, key := range mat.Pictures {
		if key == "" {
			continue
		}
		if url, err := h.generatePresignedURL(bucket, key); err == nil {
			presigned = append(presigned, url)
		} else {
			log.Printf("[presignMaterialPictures] failed for key=%s: %v", key, err)
			presigned = append(presigned, key)
		}
	}
	mat.Pictures = presigned
}
// HandlePurchaseReceivedUpdateWithImage handles multipart update including optional image replacement.
func (h *PurchaseReceivedHandler) HandlePurchaseReceivedUpdateWithImage() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing id parameter"})
			return
		}

		var existing models.PurchaseReceived
		if err := database.First(&existing, id); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusOK, gin.H{"message": "Purchase received not found", "data": []interface{}{}})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch purchase received"})
			return
		}

		// Update scalar fields from form
		if v := c.PostForm("invoice_id"); v != "" {
			existing.InvoiceID = v
		}
		if v := c.PostForm("Price"); v != "" {
			var f float64
			fmt.Sscanf(v, "%f", &f)
			existing.Price = f
		}
		if v := c.PostForm("received_quantity"); v != "" {
			var f float64
			fmt.Sscanf(v, "%f", &f)
			existing.ReceivedQuantity = f
		}
		if v := c.PostForm("purchase_id"); v != "" {
			var pid uint
			fmt.Sscanf(v, "%d", &pid)
			existing.PurchaseID = pid
		}

		// Upload new image if provided — replaces the existing one
		imageKey, err := h.UploadImage(c, "image")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
			return
		}
		if imageKey != "" {
			existing.Image = imageKey // replace
		}

		existing.UpdatedAt = time.Now()
		if err := database.Save(&existing); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update purchase received"})
			return
		}

		payload, _ := json.Marshal(map[string]interface{}{
			"purchase_id":       existing.PurchaseID,
			"invoice_id":        existing.InvoiceID,
			"received_quantity": existing.ReceivedQuantity,
		})
		activity := ActivityEvent{
			EntityType: "purchase-received",
			EntityID:   fmt.Sprintf("%d", existing.ID),
			Action:     "updated",
			Payload:    datatypes.JSON(payload),
			CreatedAt:  time.Now(),
		}
		database.Create(&activity)
		eventJSON, _ := json.Marshal(activity)
		database.Exec("SELECT pg_notify('activity_events_channel', ?)", string(eventJSON))

		enrichPurchaseReceivedImage(h, c, &existing)
		c.JSON(http.StatusOK, gin.H{
			"message":           "Purchase Received updated successfully",
			"id":                existing.ID,
			"image":             existing.Image,
			"received_quantity": existing.ReceivedQuantity,
		})
	}
}

func DeletePurchaseReceivedByID() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing id parameter"})
			return
		}

		var purchaseReceived models.PurchaseReceived
		if err := database.First(&purchaseReceived, id); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusOK, gin.H{
					"message": "Purchase received not found",
					 "data":    []interface{}{},
					})
				return
			}
			log.Println("fetch purchase received error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch purchase received"})
			return
		}

		if err := database.Delete(&purchaseReceived); err != nil {
			log.Println("delete purchase received error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete purchase received"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Purchase received deleted successfully"})
		log.Printf("Purchase received deleted | ID=%s", id)
	}
}
