package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"inventory/middleware"
	"inventory/models"
	"log"
	"net/http"
	"time"

	dbpkg "inventory/db" // adjust import path to wherever Database/GormDB live

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type ReturnRequest struct {
	Reason string `json:"reason"`
}

func HandleCreatePurchaseReturn() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := middleware.GetDB(c)
		if db == nil {
			log.Println("Database connection not found in context")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection unavailable"})
			return
		}
		purchaseID := c.Param("id")

		var req ReturnRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: reason is required"})
			return
		}

		// Use a transaction to ensure all database operations are atomic
		err := db.Transaction(func(tx dbpkg.Database) error {
			var purchase models.Purchase
			if err := tx.First(&purchase, purchaseID); err != nil {
				return err // Rollback
			}

			// Find the corresponding purchase_received record to get its ID
			var purchaseReceived models.PurchaseReceived
			tx.Where("purchase_id = ?", purchase.ID).First(&purchaseReceived)

			// Find chunks to return. Prioritize finding by purchase_received_id,
			// but fall back to purchase_id if the received record wasn't found or
			// if chunks are linked directly to the purchase.
			var chunksToReturn []models.Materialchunks
			if purchaseReceived.ID != 0 {
				tx.Where("purchase_received_id = ?", purchaseReceived.ID).Find(&chunksToReturn)
			}
			// If no chunks were found via purchase_received_id, try with purchase_id
			if len(chunksToReturn) == 0 {
				tx.Where("purchase_id = ?", purchase.ID).Find(&chunksToReturn)
			}

			if len(chunksToReturn) > 0 {
				chunkIDs := make([]uint, len(chunksToReturn))
				for i, chunk := range chunksToReturn {
					chunkIDs[i] = chunk.ID
				}

				// Count how many of the chunks to be returned are currently 'available'
				var availableChunksToReturn int64
				if err := tx.Model(&models.Materialchunks{}).Where("id IN ? AND status = ?", chunkIDs, "available").Count(&availableChunksToReturn).Error(); err != nil {
					return err // Rollback
				}

				// Mark chunks as 'returned'
				if err := tx.Model(&models.Materialchunks{}).Where("id IN ?", chunkIDs).Update("status", "returned"); err != nil {
					return err // Rollback
				}

				// Decrement the available count on the parent material only by the number of chunks
				// that were actually 'available'.
				if purchase.MaterialID != 0 && availableChunksToReturn > 0 {
					if err := tx.Model(&models.Materials{}).Where("id = ?", purchase.MaterialID).Update("available", gorm.Expr("available - ?", availableChunksToReturn)); err != nil {
						return err // Rollback
					}
				}
			}
			// Update the purchase status
			purchase.Status = "Returned"       // Set status to "Returned"
			purchase.ReturnReason = req.Reason // Save the return reason in the correct field
			if err := tx.Save(&purchase); err != nil {
				return err // Rollback
			}

			// Log activity
			now := time.Now()
			payload, _ := json.Marshal(map[string]interface{}{
				"purchase_id": purchase.ID,
				"reason":      req.Reason,
			})
			activity := ActivityEvent{
				EntityType: "purchase-return",
				EntityID:   fmt.Sprintf("%d", purchase.ID),
				Action:     "created",
				Payload:    datatypes.JSON(payload),
				CreatedAt:  now,
			}
			if err := tx.Create(&activity); err != nil {
				return err // Rollback
			}

			return nil // Commit transaction
		})

		if err != nil {
			log.Printf("Error processing purchase return: %v", err)
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Purchase or related record not found."})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process purchase return."})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Purchase returned successfully"})
	}
}
