package handlers

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tuliporg/agent-management-api/db"
	"github.com/tuliporg/agent-management-api/middleware"
	"github.com/tuliporg/agent-management-api/models"
	"github.com/tuliporg/agent-management-api/utils"
	"gorm.io/gorm"
)

func (h *Handler) CreateCustomer(c *gin.Context) {
	d := middleware.GetDB(c)

	var input models.Customer
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid input",
			Code:    "VALIDATION_ERROR",
			Details: err.Error(),
		})
		return
	}

	rawName := strings.TrimSpace(input.Name)
	if rawName == "" {
		var existingUnknown models.Customer
		if err := d.Where("LOWER(name) = LOWER(?)", "Unknown company").First(&existingUnknown); err == nil {
			c.JSON(http.StatusOK, gin.H{
				"message": "Reused existing Unknown company",
				"data":    existingUnknown,
			})
			return
		}
		input.Name = "Unknown company"
	}

	// Geocode now so the record shows up on the Field map immediately — a
	// failure here (bad/ambiguous address, Nominatim down, or address left
	// blank) must NOT block creating the customer; it just saves without
	// coordinates, same as the manual backfill already tolerates.
	geoAddr := input.Address
	if geoAddr.Street == "" && geoAddr.City == "" && geoAddr.State == "" {
		geoAddr = input.BillingAddress
	}
	if geoAddr.Street != "" || geoAddr.City != "" || geoAddr.State != "" {
		if lat, lng, err := utils.GeocodeAddress(geoAddr.Street, geoAddr.City, geoAddr.State, geoAddr.Zip, geoAddr.Country); err == nil {
			input.Address.Lat = &lat
			input.Address.Lng = &lng
			log.Println("geocoded new customer address:", lat, lng)
		} else {
			log.Println("geocode customer address failed (continuing without coordinates):", err)
		}
	}

	tx := d.Begin()
	input.ID = utils.GenerateEmployeeID(d)
	if err := tx.Create(&input); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to create customer",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to commit transaction",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Customer created successfully",
		"data":    input,
	})
}
func (h *Handler) GetAllCustomer(c *gin.Context) {
	d := middleware.GetDB(c)

	var customers []models.Customer
	if err := d.Find(&customers); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "No customers found",
				Code:    "NOT_FOUND",
				Details: err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to retrieve customers",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": customers,
	})
}
func (h *Handler) GetCustomer(c *gin.Context) {
	d := middleware.GetDB(c)

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Missing ID",
			Code:    "VALIDATION_ERROR",
			Details: "ID is required in URL path",
		})
		return
	}

	if id == "unknown" {
		var unknown models.Customer
		if err := d.Where("LOWER(name) = LOWER(?)", "Unknown company").First(&unknown); err == nil {
			id = unknown.ID
		} else {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Unknown company not found",
				Code:    "NOT_FOUND",
				Details: "Sentinel id 'unknown' could not be resolved",
			})
			return
		}
	}

	var customer models.Customer
	if err := d.First(&customer, "id = ?", id); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Customer not found",
				Code:    "NOT_FOUND",
				Details: err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to retrieve customer",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": customer,
	})
}
func (h *Handler) UpdateCustomer(c *gin.Context) {
	dbConn := middleware.GetDB(c)

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Missing ID",
			Code:    "VALIDATION_ERROR",
			Details: "ID is required in URL path",
		})
		return
	}

	var payload struct {
		Name                        string                           `json:"name"`
		ID                          string                           `json:"id"`
		NameAlt                     string                           `json:"Name"`
		Phone                       string                           `json:"phone"`
		Email                       string                           `json:"email"`
		SecondaryEmail              string                           `json:"SecondaryEmail"`
		AccountingEmail             string                           `json:"AccountingEmail"`
		SecondaryPhoneNumber        string                           `json:"SecondaryPhoneNumber"`
		CellNumber                  string                           `json:"CellNumber"`
		Address                     models.CustomerAddress           `json:"Address"`
		BillingAddress              models.CustomerAddress           `json:"BillingAddress"`
		Details                     models.CustomerDetails           `json:"details"`
		ContactName                 string                           `json:"ContactName"`
		AdditionalPersonContacts    []models.AdditionalPersonContact `json:"AdditionalPersonContacts"`
		SalesPerson1                string                           `json:"SalesPerson1"`
		SalesPerson2                string                           `json:"SalesPerson2"`
		ProjectManager              string                           `json:"ProjectManager"`
		CustomerSince               string                           `json:"CustomerSince"`
		SalesTax                    string                           `json:"SalesTax"`
		EINNumber                   string                           `json:"EINNumber"`
		TaxExempt                   bool                             `json:"TaxExempt"`
		TaxExemptReason             string                           `json:"TaxExemptReason"`
		ExemptCertificateNumber     string                           `json:"ExemptCertificateNumber"`
		ExemptExpiryDate            string                           `json:"ExemptExpiryDate"`
		SpecialDeliveryInstructions string                           `json:"SpecialDeliveryInstructions"`
		CollectionNotes             string                           `json:"CollectionNotes"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid input",
			Code:    "VALIDATION_ERROR",
			Details: err.Error(),
		})
		return
	}

	name := strings.TrimSpace(payload.Name)
	if name == "" {
		name = strings.TrimSpace(payload.NameAlt)
	}
	newID := strings.TrimSpace(payload.ID)

	if name != "" && len(strings.TrimSpace(name)) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid name",
			Code:    "VALIDATION_ERROR",
			Details: "Name cannot be only whitespace",
		})
		return
	}

	var updated models.Customer
	err := dbConn.Transaction(func(tx db.Database) error {
		var existing models.Customer
		if err := tx.First(&existing, "id = ?", id); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return &typedError{
					status: http.StatusNotFound,
					resp: ErrorResponse{
						Error:   "Customer not found",
						Code:    "NOT_FOUND",
						Details: err.Error(),
					},
				}
			}
			return &typedError{
				status: http.StatusInternalServerError,
				resp: ErrorResponse{
					Error:   "Failed to load customer",
					Code:    "DATABASE_ERROR",
					Details: err.Error(),
				},
			}
		}

		if name != "" {
			existing.Name = name
		}
		existing.Phone = payload.Phone
		existing.Email = payload.Email
		existing.SecondaryEmail = payload.SecondaryEmail
		existing.AccountingEmail = payload.AccountingEmail
		existing.SecondaryPhoneNumber = payload.SecondaryPhoneNumber
		existing.CellNumber = payload.CellNumber
		existing.Address = payload.Address
		existing.BillingAddress = payload.BillingAddress
		existing.Details = payload.Details

		// Re-geocode whenever the address is edited. Falls back to
		// BillingAddress if Address is empty, same convention the Field
		// page's frontend already uses when reading these back. A failure
		// here must NOT block the update — save without new coordinates.
		geoAddr := existing.Address
		if geoAddr.Street == "" && geoAddr.City == "" && geoAddr.State == "" {
			geoAddr = existing.BillingAddress
		}
		if geoAddr.Street != "" || geoAddr.City != "" || geoAddr.State != "" {
			if lat, lng, err := utils.GeocodeAddress(geoAddr.Street, geoAddr.City, geoAddr.State, geoAddr.Zip, geoAddr.Country); err == nil {
				existing.Address.Lat = &lat
				existing.Address.Lng = &lng
				log.Println("geocoded updated customer address:", lat, lng)
			} else {
				log.Println("geocode customer address failed (keeping previous coordinates):", err)
			}
		}

		existing.ContactName = payload.ContactName
		existing.AdditionalPersonContacts = payload.AdditionalPersonContacts
		existing.SalesPerson1 = payload.SalesPerson1
		existing.SalesPerson2 = payload.SalesPerson2
		existing.ProjectManager = payload.ProjectManager
		existing.CustomerSince = payload.CustomerSince
		existing.SalesTax = payload.SalesTax
		existing.EINNumber = payload.EINNumber
		existing.TaxExempt = payload.TaxExempt
		existing.TaxExemptReason = payload.TaxExemptReason
		existing.ExemptCertificateNumber = payload.ExemptCertificateNumber
		existing.ExemptExpiryDate = payload.ExemptExpiryDate
		existing.SpecialDeliveryInstructions = payload.SpecialDeliveryInstructions
		existing.CollectionNotes = payload.CollectionNotes

		// Save all non-primary-key fields
		if err := tx.Save(&existing); err != nil {
			return &typedError{
				status: http.StatusInternalServerError,
				resp: ErrorResponse{
					Error:   "Failed to update customer",
					Code:    "DATABASE_ERROR",
					Details: err.Error(),
				},
			}
		}

		// If the ID has changed, we must perform a separate, targeted UPDATE on the primary key.
		// Simply changing the ID on the model and calling Save() will create a new record.
		if newID != "" && newID != id {
			if err := tx.Model(&models.Customer{}).Where("id = ?", id).Update("id", newID); err != nil {
				return &typedError{
					status: http.StatusInternalServerError,
					resp: ErrorResponse{
						Error:   "Failed to update customer ID",
						Code:    "DATABASE_ERROR",
						Details: err.Error(),
					},
				}
			}
			if err := updateAssociateCompanyID(tx, id, newID); err != nil {
				return &typedError{
					status: http.StatusInternalServerError,
					resp: ErrorResponse{
						Error:   "Failed to update associated lead company IDs",
						Code:    "DATABASE_ERROR",
						Details: err.Error(),
					},
				}
			}
			existing.ID = newID // Update the ID on the in-memory model for the response
		}

		updated = existing
		return nil
	})

	if err != nil {
		if te, ok := err.(*typedError); ok {
			c.JSON(te.status, te.resp)
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Unexpected transaction error",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Customer updated successfully",
		"data":    updated,
	})
}

func updateAssociateCompanyID(tx db.Database, oldID, newID string) error {
	if oldID == "" || newID == "" || oldID == newID {
		return nil
	}

	return tx.Exec(`
UPDATE associate_customers
SET details = jsonb_set(
	jsonb_set(
		CASE
			WHEN jsonb_typeof(COALESCE(details, '{}'::jsonb)->'company') = 'object'
			THEN COALESCE(details, '{}'::jsonb)
			ELSE jsonb_set(COALESCE(details, '{}'::jsonb), '{company}', '{}'::jsonb, true)
		END,
		'{company_id}',
		to_jsonb(CAST(? AS text)),
		true
	),
	'{company,id}',
	to_jsonb(CAST(? AS text)),
	true
)
WHERE details->>'company_id' = ?
	OR details->'company'->>'id' = ?`, newID, newID, oldID, oldID)
}

type typedError struct {
	status int
	resp   ErrorResponse
}

func (e *typedError) Error() string { return e.resp.Error }
func (h *Handler) DeleteCustomer(c *gin.Context) {
	d := middleware.GetDB(c)

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Missing ID",
			Code:    "VALIDATION_ERROR",
			Details: "ID is required in URL path",
		})
		return
	}

	tx := d.Begin()
	var customer models.Customer
	if err := tx.First(&customer, "id = ?", id); err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Customer not found",
				Code:    "NOT_FOUND",
				Details: err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to find customer",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
		}
		return
	}

	if err := tx.Delete(&customer); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to delete customer",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to commit transaction",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Customer deleted successfully",
	})
}
