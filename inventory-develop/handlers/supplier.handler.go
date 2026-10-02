package handlers

import (
	"errors"
	"inventory/middleware"
	"inventory/models"
	"inventory/utils"
	"log"
	"net/http"
	"strings"
	"inventory/db"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ─── helpers ──────────────────────────────────────────────────────────────────

func nilIfEmpty(s *string) *string {
	if s != nil && *s == "" {
		return nil
	}
	return s
}

func isAddressComplete(a models.Address) bool {
	return strings.TrimSpace(a.Street) != "" &&
		strings.TrimSpace(a.City) != "" &&
		strings.TrimSpace(a.State) != "" &&
		strings.TrimSpace(a.Zip) != "" &&
		strings.TrimSpace(a.Country) != ""
}

// ─── request body ─────────────────────────────────────────────────────────────

type supplierInput struct {
	ID       *uint   `json:"id"` 
	Name     string  `json:"name"`
	Phone    *string `json:"phone"`
	Email    *string `json:"email"`
	Category string  `json:"category"`
	Website  *string `json:"website"`

	// new fields (go into Details)
	Address              models.Address `json:"address"`
	PaymentTerms         string         `json:"paymentTerms"`
	PricingType          string         `json:"pricingType"`
	IsReferred           bool           `json:"isReferred"`
	ReferrerName         string         `json:"referrerName"`
	ReferrerPhone        string         `json:"referrerPhone"`
	ReferrerEmail        string         `json:"referrerEmail"`
	ReferrerCompanyName  string         `json:"referrerCompanyName"`
	ReferrerCompanyPhone string         `json:"referrerCompanyPhone"`
	ReferrerCompanyEmail string         `json:"referrerCompanyEmail"`
	PrimaryContact       *string        `json:"primaryContact"`
	SecondaryContact     *string        `json:"secondaryContact"`
	PrimaryPhone         *string        `json:"primaryPhone"`
	SecondaryPhone       *string        `json:"secondaryPhone"`
	Notes                []models.Note  `json:"notes"`
}

func buildDetails(in *supplierInput) models.SupplierDetails {
	d := models.SupplierDetails{
		Address:      in.Address,
		PaymentTerms: in.PaymentTerms,
		PricingType:  in.PricingType,
		IsReferred:   in.IsReferred,
		Notes:        in.Notes,
	}
	if in.IsReferred {
		d.ReferrerName = in.ReferrerName
		d.ReferrerPhone = in.ReferrerPhone
		d.ReferrerEmail = in.ReferrerEmail
		d.ReferrerCompanyName = in.ReferrerCompanyName
		d.ReferrerCompanyPhone = in.ReferrerCompanyPhone
		d.ReferrerCompanyEmail = in.ReferrerCompanyEmail
	}
	return d
}

// ─── handlers ─────────────────────────────────────────────────────────────────

func CreateSupplier() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)

		var input supplierInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		if strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.Category) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Name and Category are required"})
			return
		}

		if !isAddressComplete(input.Address) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Address is required: street, city, state, zip and country must all be provided",
			})
			return
		}

		// Best-effort: a slow/unavailable geocoder should never block
		// creating a supplier — it just won't have a map pin yet.
		if lat, lng, err := utils.GeocodeAddress(
			input.Address.Street, input.Address.City, input.Address.State, input.Address.Zip, input.Address.Country,
		); err == nil {
			input.Address.Lat = &lat
			input.Address.Lng = &lng
		} else {
			log.Println("geocode supplier address failed (continuing without coordinates):", err)
		}

		supplier := models.Supplier{
			Name:     input.Name,
			Phone:    nilIfEmpty(input.Phone),
			Email:    nilIfEmpty(input.Email),
			Category: input.Category,
			Website:  nilIfEmpty(input.Website),
			Details:  buildDetails(&input),
			PrimaryContact:       input.PrimaryContact,
			SecondaryContact:     input.SecondaryContact,
			PrimaryPhone:         input.PrimaryPhone,
			SecondaryPhone:       input.SecondaryPhone,
		}

		if err := database.Create(&supplier); err != nil {
			if strings.Contains(err.Error(), "idx_suppliers_email") {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Email already exists"})
				return
			}
			if strings.Contains(err.Error(), "idx_suppliers_phone") {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Phone already exists"})
				return
			}
			log.Println("create supplier error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create supplier"})
			return
		}

		c.JSON(http.StatusCreated, supplier)
	}
}

func GetAllSuppliers() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)
		var suppliers []models.Supplier

		if err := database.
			Order("updated_at desc").
			Find(&suppliers); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, suppliers)
	}
}

func GetSupplierByID() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing supplier id"})
			return
		}

		var supplier models.Supplier
		err := database.First(&supplier, id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Println("supplier not found:", id)
				c.JSON(http.StatusOK, gin.H{
					"message": "Supplier not found",
					"data":    []interface{}{},
				})
				return
			}
			log.Println("fetch supplier error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch supplier"})
			return
		}

		c.JSON(http.StatusOK, supplier)
	}
}

func UpdateSupplier() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)
		id := c.Param("id")

		var supplier models.Supplier
		if err := database.First(&supplier, id); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Supplier not found"})
			return
		}

		var input supplierInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		if strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.Category) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Name and Category are required"})
			return
		}

		if !isAddressComplete(input.Address) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Address is required: street, city, state, zip and country must all be provided",
			})
			return
		}

		oldID := supplier.ID
		newID := oldID
		idChanging := input.ID != nil && *input.ID != oldID

		if idChanging {
			newID = *input.ID
			var existing models.Supplier
			if err := database.First(&existing, newID); err == nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "This ID is already in use"})
				return
			}
		}

		// Re-geocode whenever the address changes; keep the old coordinates
		// if the address didn't change or the geocoder call fails.
		if lat, lng, err := utils.GeocodeAddress(
			input.Address.Street, input.Address.City, input.Address.State, input.Address.Zip, input.Address.Country,
		); err == nil {
			input.Address.Lat = &lat
			input.Address.Lng = &lng
		} else {
			log.Println("geocode supplier address failed (keeping previous coordinates):", err)
			input.Address.Lat = supplier.Details.Address.Lat
			input.Address.Lng = supplier.Details.Address.Lng
		}

		supplier.Name = input.Name
		supplier.Phone = nilIfEmpty(input.Phone)
		supplier.Email = nilIfEmpty(input.Email)
		supplier.Category = input.Category
		supplier.Website = nilIfEmpty(input.Website)
		supplier.Details = buildDetails(&input)
		supplier.PrimaryContact = input.PrimaryContact
		supplier.SecondaryContact = input.SecondaryContact
		supplier.PrimaryPhone = input.PrimaryPhone
		supplier.SecondaryPhone = input.SecondaryPhone

		err := database.Transaction(func(tx db.Database) error {
			if idChanging {
				// Re-point purchases FK before changing the primary key
				if err := tx.Exec(
					"UPDATE purchases SET supplier_id = ? WHERE supplier_id = ?",
					newID, oldID,
				); err != nil {
					return err
				}

				if err := tx.Exec(
					`UPDATE materialchunks
					SET supplier_details = jsonb_set(
						supplier_details,
						'{id}',
						to_jsonb(?::int)
					)
					WHERE supplier_details->>'id' ~ '^[0-9]+$'
					AND (supplier_details->>'id')::int = ?`,
					newID,
					oldID,
				); err != nil {
					return err
				}

				// Now change the primary key itself
				if err := tx.Exec(
					"UPDATE suppliers SET id = ? WHERE id = ?",
					newID, oldID,
				); err != nil {
					return err
				}

				supplier.ID = newID
			}

			if err := tx.Save(&supplier); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			if strings.Contains(err.Error(), "email") {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Email already exists"})
				return
			}
			if strings.Contains(err.Error(), "phone") {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Phone already exists"})
				return
			}
			log.Println("update supplier error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, supplier)
	}
}

func DeleteSupplier() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)
		id := c.Param("id")

		if err := database.Delete(&models.Supplier{}, id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Supplier deleted successfully"})
	}
}