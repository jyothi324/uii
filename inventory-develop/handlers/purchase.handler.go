package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"inventory/middleware"
	"inventory/models"
	"io"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type PurchaseRequest struct {
	models.Purchase
	// We use the model directly, but ensure PurchaseItems is handled
}

// calculateSqFt helper function to parse dimensions (e.g., "120x60") and return square feet
func calculateSqFt(val string) float64 {
	val = strings.ToLower(strings.TrimSpace(val))
	parts := strings.Split(val, "x")
	if len(parts) != 2 {
		return 0
	}
	l, err1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	w, err2 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err1 != nil || err2 != nil {
		return 0
	}
	sqft := (l / 12.0) * (w / 12.0)
	return math.Round(sqft*100) / 100
}

// parsePurchaseItems helper function to construct the JSON array from comma-separated strings
func parsePurchaseItems(p models.Purchase, materialID uint) models.PurchaseItems {
	if p.Quantity == "" || p.CurrentPrice == "" {
		return nil
	}

	qParts := strings.Split(p.Quantity, ",")
	pParts := strings.Split(p.CurrentPrice, ",")

	dParts := []string{}
	if p.Dimensions != "" {
		dParts = strings.Split(p.Dimensions, ",")
	}

	if len(qParts) != len(pParts) {
		return nil
	}

	items := make(models.PurchaseItems, 0, len(qParts))
	for i := 0; i < len(qParts); i++ {
		dim := ""
		if i < len(dParts) {
			dim = strings.TrimSpace(dParts[i])
		}
		items = append(items, models.PurchaseItem{
			MaterialID:   materialID,
			Quantity:     strings.TrimSpace(qParts[i]),
			CurrentPrice: strings.TrimSpace(pParts[i]),
			Dimensions:   dim,
			SquareFeet:   calculateSqFt(dim),
		})
	}
	return items
}

func HandleCreatePurchase() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)

		// Read body once to check format
		bodyBytes, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		var req models.Purchase
		// Try to bind as a full Purchase object
		if err := json.Unmarshal(bodyBytes, &req); err != nil {
			// If that fails, check if the body is just the items array
			var items []models.PurchaseItem
			if errArr := json.Unmarshal(bodyBytes, &items); errArr == nil {
				req.PurchaseItems = items
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
				return
			}
		}
		// Normalize invalid ID
		if req.CountyTaxID != nil && *req.CountyTaxID == 0 {
			req.CountyTaxID = nil
		}

		// Log for debugging
		if req.CountyTaxID != nil {
			log.Printf("Purchase created with County Tax ID: %d | Name: %s | Percentage: %.2f%%",
				*req.CountyTaxID, req.CountyName, req.TaxPercentage)
		}

		// Auto-populate purchase_items from comma-separated strings if items array is empty
		if len(req.PurchaseItems) == 0 {
			req.PurchaseItems = parsePurchaseItems(req, req.MaterialID)
		}

		now := time.Now()
		req.CreatedAt = now
		req.UpdatedAt = now

		if err := database.Create(&req); err != nil {
			log.Println("insert purchase error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create purchase",
			})
			return
		}

		var supplierName string
		if req.SupplierID != 0 {
			var s models.Supplier
			if err := database.First(&s, req.SupplierID); err == nil {
				supplierName = s.Name
			}
		}

		payload, _ := json.Marshal(map[string]interface{}{
			"material_name":      req.Name,
			"status":             req.Status,
			"current_price":      req.CurrentPrice,
			"quantity":           req.Quantity,
			"purchase_items":     req.PurchaseItems,
			"supplier":           supplierName,
			"inventory_type":     req.InventoryType,
			"note":               req.Note,
			"county_tax_id":      req.CountyTaxID,
			"county_name":        req.CountyName,
			"tax_percentage":     req.TaxPercentage,
			"deliveryCharge":     req.DeliveryCharges,
			"additional_charges": req.AdditionalCharges,
		})

		activity := ActivityEvent{
			EntityType: "purchase",
			EntityID:   fmt.Sprintf("%d", req.ID),
			Action:     "created",
			Payload:    datatypes.JSON(payload),
			CreatedAt:  now,
		}
		database.Create(&activity)
		eventJSON, _ := json.Marshal(activity)
		database.Exec("SELECT pg_notify('activity_events_channel', ?)", string(eventJSON))

		c.JSON(http.StatusCreated, gin.H{
			"id":                 req.ID,
			"name":               req.Name,
			"quantity":           req.Quantity,
			"current_price":      req.CurrentPrice,
			"status":             req.Status,
			"dimensions":         req.Dimensions,
			"square_feet":        req.SquareFeet,
			"purchase_items":     req.PurchaseItems,
			"inventory_type":     req.InventoryType,
			"note":               req.Note,
			"county_tax_id":      req.CountyTaxID,
			"county_name":        req.CountyName,
			"tax_percentage":     req.TaxPercentage,
			"delivery_charges":   req.DeliveryCharges,
			"additional_charges": req.AdditionalCharges,
			"message":            "Purchase created successfully",
		})

		log.Printf("Purchase created | ID=%d | Name=%s | CountyTaxID=%v | Tax=%.2f%%",
			req.ID, req.Name, req.CountyTaxID, req.TaxPercentage)
	}
}

func HandleFetchPurchases() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)
		var purchases []models.Purchase
		if err := database.Preload("Materials").Preload("Supplier").Find(&purchases); err != nil {
			log.Println("fetch purchases error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch purchases",
			})
			return
		}
		c.JSON(http.StatusOK, purchases)
		log.Printf("Fetched %d purchases", len(purchases))
	}
}

func HandlePurchaseGetByID() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing id User"})
			return
		}

		var purchase models.Purchase

		err := database.Preload("Materials").Preload("Supplier").First(&purchase, id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Println("purchase not found:", id)
				c.JSON(http.StatusOK, gin.H{
					"message": "Purchase not found",
					"data":    []interface{}{},
				})
				return
			}

			log.Println("fetch purchase error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch purchase"})
			return
		}

		c.JSON(http.StatusOK, purchase)
	}
}

func HandlePurchaseUpdate() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := middleware.GetDB(c)
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing id parameter"})
			return
		}

		// Read body once to check format
		bodyBytes, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		var req models.Purchase
		// Try to bind as a full Purchase object
		if err := json.Unmarshal(bodyBytes, &req); err != nil {
			// If that fails, check if the body is just the items array
			var items []models.PurchaseItem
			if errArr := json.Unmarshal(bodyBytes, &items); errArr == nil {
				req.PurchaseItems = items
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
				return
			}
		}

		// Check if purchase exists
		var existingPurchase models.Purchase
		if err := db.First(&existingPurchase, id); err != nil {
			if err == gorm.ErrRecordNotFound {
				log.Println("purchase not found:", id)
				c.JSON(http.StatusOK, gin.H{
					"message": "Purchase not found",
					"data":    []interface{}{},
				})
				return
			}
			log.Println("fetch purchase error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch purchase"})
			return
		}

		// Capture old state for change tracking
		oldPurchase := existingPurchase
		var oldSupplierName string
		if oldPurchase.SupplierID != 0 {
			var s models.Supplier
			if err := db.First(&s, oldPurchase.SupplierID); err == nil {
				oldSupplierName = s.Name
			}
		}

		now := time.Now()
		existingPurchase.UpdatedAt = now
		stringsUpdated := false
		if req.Name != "" {
			existingPurchase.Name = req.Name
		}
		if req.Quantity != "" {
			existingPurchase.Quantity = req.Quantity
			stringsUpdated = true
		}
		if req.CurrentPrice != "" {
			existingPurchase.CurrentPrice = req.CurrentPrice
			stringsUpdated = true
		}
		if req.MaterialID != 0 {
			existingPurchase.MaterialID = req.MaterialID
		}
		if req.SupplierID != 0 {
			existingPurchase.SupplierID = req.SupplierID
		}
		if req.Dimensions != "" {
			existingPurchase.Dimensions = req.Dimensions
			stringsUpdated = true
		}
		if req.Status != "" {
			existingPurchase.Status = req.Status
		}
		if req.RejectionReason != "" {
			existingPurchase.RejectionReason = req.RejectionReason
		}
		if req.SquareFeet != 0 {
			existingPurchase.SquareFeet = req.SquareFeet
		}

		itemsProvided := false
		if req.PurchaseItems != nil {
			existingPurchase.PurchaseItems = req.PurchaseItems
			itemsProvided = true
		}
		if req.BillId != "" {
			existingPurchase.BillId = req.BillId
		}

		// If items weren't explicitly provided, but component strings were updated, re-parse.
		// Or if existing items are empty, try to parse.
		if !itemsProvided && (stringsUpdated || len(existingPurchase.PurchaseItems) == 0) {
			existingPurchase.PurchaseItems = parsePurchaseItems(existingPurchase, existingPurchase.MaterialID)
		}
		if req.InventoryType != "" {
			existingPurchase.InventoryType = req.InventoryType
		}
		existingPurchase.CountyTaxID = req.CountyTaxID
		existingPurchase.CountyName = req.CountyName
		existingPurchase.TaxPercentage = req.TaxPercentage
		existingPurchase.AdditionalCharges = req.AdditionalCharges
		existingPurchase.DeliveryCharges = req.DeliveryCharges
		existingPurchase.Note = req.Note
		
		var supplierName string
		if existingPurchase.SupplierID != 0 {
			var s models.Supplier
			if err := db.First(&s, existingPurchase.SupplierID); err == nil {
				supplierName = s.Name
			}
		}

		if err := db.Save(&existingPurchase); err != nil {
			log.Println("update purchase error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update purchase"})
			return
		}

		changedFields := getPurchaseChanges(
			oldPurchase, existingPurchase,
			oldSupplierName, supplierName,
		)

		payload, _ := json.Marshal(map[string]interface{}{
			"material_name":     existingPurchase.Name,
			"status":            existingPurchase.Status,
			"purchase_items":    existingPurchase.PurchaseItems,
			"supplier":          supplierName,
			"inventory_type":    existingPurchase.InventoryType,
			"changed_fields":    changedFields,
			"note":              existingPurchase.Note,
			"county_tax_id":     existingPurchase.CountyTaxID,
			"county_name":       existingPurchase.CountyName,
			"tax_percentage":    existingPurchase.TaxPercentage,
			"delivery_charges":  existingPurchase.DeliveryCharges,
			"additional_charge": existingPurchase.AdditionalCharges,
		})
		activity := ActivityEvent{
			EntityType: "purchase",
			EntityID:   fmt.Sprintf("%d", existingPurchase.ID),
			Action:     "updated",
			Payload:    datatypes.JSON(payload),
			CreatedAt:  now,
		}
		db.Create(&activity)
		eventJSON, _ := json.Marshal(activity)
		db.Exec("SELECT pg_notify('activity_events_channel', ?)", string(eventJSON))

		c.JSON(http.StatusOK, gin.H{
			"message": "Purchase updated successfully",
		})
		log.Printf("Purchase updated | ID=%s | Name=%s",
			id, existingPurchase.Name)
	}
}

func getPurchaseChanges(old, new models.Purchase, oldSupplier, newSupplier string) map[string]map[string]string {
	changes := make(map[string]map[string]string)
	if old.Name != new.Name {
		changes["material_name"] = map[string]string{"old_value": old.Name, "new_value": new.Name}
	}
	if old.Status != new.Status {
		changes["status"] = map[string]string{"old_value": old.Status, "new_value": new.Status}
	}
	if old.Status != "Returned" && new.Status == "Returned" {
		changes["status"] = map[string]string{"old_value": old.Status, "new_value": "Returned"}
	}
	if old.Quantity != new.Quantity {
		changes["quantity"] = map[string]string{"old_value": old.Quantity, "new_value": new.Quantity}
	}
	if old.CurrentPrice != new.CurrentPrice {
		changes["unit_price"] = map[string]string{"old_value": old.CurrentPrice, "new_value": new.CurrentPrice}
	}
	if old.InventoryType != new.InventoryType {
		changes["inventory_type"] = map[string]string{"old_value": old.InventoryType, "new_value": new.InventoryType}
	}
	if old.Dimensions != new.Dimensions {
		changes["dimensions"] = map[string]string{"old_value": old.Dimensions, "new_value": new.Dimensions}
	}
	if oldSupplier != newSupplier {
		changes["supplier"] = map[string]string{"old_value": oldSupplier, "new_value": newSupplier}
	}
	return changes
}
func DeletePurchaseByID() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing id parameter"})
			return
		}

		// Check if purchase exists
		var purchase models.Purchase
		if err := database.First(&purchase, id); err != nil {
			if err == gorm.ErrRecordNotFound {
				log.Println("purchase not found:", id)
				c.JSON(http.StatusOK, gin.H{
					"message": "Purchase not found",
					"data":    []interface{}{},
				})
				return
			}
			log.Println("fetch purchase error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch purchase"})
			return
		}

		// Delete the purchase
		if err := database.Delete(&purchase); err != nil {
			log.Println("delete purchase error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete purchase"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Purchase deleted successfully"})
		log.Printf("Purchase deleted | ID=%s", id)
	}
}
