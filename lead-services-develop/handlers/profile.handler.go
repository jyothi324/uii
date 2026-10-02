package handlers

import (
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tuliporg/agent-management-api/middleware"
	"github.com/tuliporg/agent-management-api/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type AddressPayload struct {
	State     string  `json:"State"`
	County    string  `json:"County"`
	Street    string  `json:"Street"`
	UnitSuite string  `json:"UnitSuite"`
	City      string  `json:"City"`
	Zip       string  `json:"Zip"`
	Formatted string  `json:"formatted"`
	Lat       float64 `json:"lat"`
	Lng       float64 `json:"lng"`
}

type Note struct {
	Text      string `json:"text"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type ProfileDetails struct {
	ReferrerFirstName string `json:"ReferrerFirstName,omitempty"`
	ReferrerLastName  string `json:"ReferrerLastName,omitempty"`
	ReferrerPhone     string `json:"ReferrerPhone,omitempty"`
	ReferrerEmail     string `json:"ReferrerEmail,omitempty"`
	ReferrerType      string `json:"ReferrerType,omitempty"`

	CompanyName  string `json:"CompanyName,omitempty"`
	CompanyPhone string `json:"CompanyPhone,omitempty"`
	CompanyEmail string `json:"CompanyEmail,omitempty"`

	ReffererCompanyName  string `json:"ReferrerCompanyName,omitempty"`
	ReffererCompanyPhone string `json:"ReferrerCompanyPhone,omitempty"`
	ReffererCompanyEmail string `json:"ReferrerCompanyEmail,omitempty"`
	CountyTax            string `json:"CountyTax,omitempty"`
	PaymentTerms         string `json:"PaymentTerms,omitempty"`
	PricingType          string `json:"PricingType,omitempty"`
	SalesPerson          string `json:"SalesPerson,omitempty"`
	Notes                []Note `json:"Notes,omitempty"`
}

type ProfileRequest struct {
	Name            string         `json:"Name" binding:"required"`
	Phone           string         `json:"Phone" binding:"required"`
	Secondaryphone   string         `json:"SecondaryPhone"`
	Cellphone       string         `json:"CellPhone"`
	Email           string         `json:"Email"`
	SecondaryEmail  string         `json:"SecondaryEmail"`
	SameAsShipping  bool           `json:"SameAsShipping"`
	ShippingAddress AddressPayload `json:"ShippingAddress"`
	BillingAddress  AddressPayload `json:"BillingAddress"`
	Details         ProfileDetails `json:"Details"`
	Latitude        float64        `json:"Latitude"`
	Longitude       float64        `json:"Longitude"`
}

// Create
func (h *Handler) CreateProfile(c *gin.Context) {
	db := middleware.GetDB(c)
	if db == nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Database not initialized",
			Code:  "DATABASE_ERROR",
		})
		return
	}

	var req ProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body",
			Code:  "VALIDATION_ERROR",
		})
		return
	}

	var existing models.Profile
	if err := db.Where("phone = ?", req.Phone).First(&existing); err == nil {
		c.JSON(http.StatusConflict, ErrorResponse{
			Error: "Phone already exists",
			Code:  "DUPLICATE_PROFILE",
		})
		return
	}

	tx := db.Begin()
	if tx == nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to start transaction",
			Code:  "DATABASE_ERROR",
		})
		return
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	profile := models.Profile{
		Name:           req.Name,
		Phone:          req.Phone,
		Secondaryphone:  req.Secondaryphone,
		Cellphone:      req.Cellphone,
		Email:          req.Email,
		SecondaryEmail: req.SecondaryEmail,
		SameAsShipping: req.SameAsShipping,
	}

	// Shipping address JSON
	shippingJSON, _ := json.Marshal(models.Address{
		State:     req.ShippingAddress.State,
		County:    req.ShippingAddress.County,
		Street:    req.ShippingAddress.Street,
		UnitSuite: req.ShippingAddress.UnitSuite,
		City:      req.ShippingAddress.City,
		Zip:       req.ShippingAddress.Zip,
		Formatted: req.ShippingAddress.Formatted,
		Lat:       req.Latitude,
		Lng:       req.Longitude,
	})
	profile.ShippingAddress = datatypes.JSON(shippingJSON)

	// Billing address JSON
	if req.SameAsShipping {
		profile.BillingAddress = profile.ShippingAddress
	} else {
		billingJSON, _ := json.Marshal(models.Address{
			State:     req.BillingAddress.State,
			County:    req.BillingAddress.County,
			Street:    req.BillingAddress.Street,
			UnitSuite: req.BillingAddress.UnitSuite,
			City:      req.BillingAddress.City,
			Zip:       req.BillingAddress.Zip,
			Formatted: req.BillingAddress.Formatted,
			Lat:       req.BillingAddress.Lat,
			Lng:       req.BillingAddress.Lng,
		})
		profile.BillingAddress = datatypes.JSON(billingJSON)
	}

	detailsJSON, _ := json.Marshal(req.Details)
	profile.Details = datatypes.JSON(detailsJSON)

	if err := tx.Create(&profile); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to create profile",
			Code:  "DATABASE_ERROR",
		})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to commit transaction",
			Code:  "DATABASE_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, profile)
}

func (h *Handler) GetProfile(c *gin.Context) {
	db := middleware.GetDB(c)

	identifier := c.Param("phoneoremail")
	if identifier == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Phone or email is required",
			Code:  "VALIDATION_ERROR",
		})
		return
	}

	var profile models.Profile

	if err := db.
		Where("email = ? OR phone = ? OR secondary_email = ? OR secondaryphone = ? OR cellphone = ?",identifier, identifier, identifier, identifier, identifier).
		First(&profile); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "Profile not found",
				Code:  "NOT_FOUND",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: err.Error(),
			Code:  "DATABASE_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, profile)
}

func (h *Handler) GetAllProfiles(c *gin.Context) {
	db := middleware.GetDB(c)
	if db == nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Database not initialized",
			Code:  "DATABASE_ERROR",
		})
		return
	}

	var profiles []models.Profile
	var total int64

	pageStr := c.Query("page")
	limitStr := c.Query("limit")
	search := c.Query("search")

	paginationRequested := !(pageStr == "" && limitStr == "")

	page := 1
	limit := 10

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := (page - 1) * limit

	query := db.Model(&models.Profile{})

	if search != "" {
		searchTerm := "%" + search + "%"
		query = query.Where(
			"name ILIKE ? OR email ILIKE ? OR phone ILIKE ?",
			searchTerm, searchTerm, searchTerm,
		)
	}

	countQuery := query
	if err := countQuery.Count(&total).Error(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: err.Error(),
			Code:  "DATABASE_ERROR",
		})
		return
	}

	if !paginationRequested {
		if err := query.Order("id DESC").Find(&profiles); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: err.Error(),
				Code:  "DATABASE_ERROR",
			})
			return
		}
		if total == 0 {
			limit = 0
		} else {
			limit = int(total)
		}
		page = 1
	} else {
		if err := query.
			Limit(limit).
			Offset(offset).
			Order("id DESC").
			Find(&profiles); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: err.Error(),
				Code:  "DATABASE_ERROR",
			})
			return
		}
	}

	totalPages := 0
	if limit > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}

	c.JSON(http.StatusOK, gin.H{
		"data": profiles,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": totalPages,
		},
	})
}

func (h *Handler) UpdateProfile(c *gin.Context) {
	db := middleware.GetDB(c)

	id := c.Param("id")

	var req ProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body",
			Code:  "VALIDATION_ERROR",
		})
		return
	}

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var p models.Profile

	if err := tx.First(&p, "id = ?", id); err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "Profile not found",
				Code:  "NOT_FOUND",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to fetch profile",
			Code:  "DATABASE_ERROR",
		})
		return
	}

	// Update basic fields
	p.Name = req.Name
	p.Phone = req.Phone
	p.Secondaryphone = req.Secondaryphone
	p.Cellphone = req.Cellphone
	p.Email = req.Email
	p.SecondaryEmail = req.SecondaryEmail
	p.SameAsShipping = req.SameAsShipping

	// Shipping JSON
	shippingJSON, _ := json.Marshal(models.Address{
		State:     req.ShippingAddress.State,
		County:    req.ShippingAddress.County,
		Street:    req.ShippingAddress.Street,
		UnitSuite: req.ShippingAddress.UnitSuite,
		City:      req.ShippingAddress.City,
		Zip:       req.ShippingAddress.Zip,
		Formatted: req.ShippingAddress.Formatted,
		Lat:       req.ShippingAddress.Lat,
		Lng:       req.ShippingAddress.Lng,
	})
	p.ShippingAddress = datatypes.JSON(shippingJSON)

	// Billing JSON
	if req.SameAsShipping {
		p.BillingAddress = p.ShippingAddress
	} else {
		billingJSON, _ := json.Marshal(models.Address{
			State:     req.BillingAddress.State,
			County:    req.BillingAddress.County,
			Street:    req.BillingAddress.Street,
			UnitSuite: req.BillingAddress.UnitSuite,
			City:      req.BillingAddress.City,
			Zip:       req.BillingAddress.Zip,
			Formatted: req.BillingAddress.Formatted,
			Lat:       req.BillingAddress.Lat,
			Lng:       req.BillingAddress.Lng,
		})
		p.BillingAddress = datatypes.JSON(billingJSON)
	}

	// Build updated details - use incoming notes directly
	updatedDetails := ProfileDetails{
		Notes:                req.Details.Notes,
		ReferrerFirstName:    req.Details.ReferrerFirstName,
		ReferrerLastName:     req.Details.ReferrerLastName,
		ReferrerPhone:        req.Details.ReferrerPhone,
		ReferrerEmail:        req.Details.ReferrerEmail,
		ReferrerType:         req.Details.ReferrerType,
		CompanyName:          req.Details.CompanyName,
		CompanyPhone:         req.Details.CompanyPhone,
		CompanyEmail:         req.Details.CompanyEmail,
		ReffererCompanyName:  req.Details.ReffererCompanyName,
		ReffererCompanyPhone: req.Details.ReffererCompanyPhone,
		ReffererCompanyEmail: req.Details.ReffererCompanyEmail,
		CountyTax:            req.Details.CountyTax,
		PaymentTerms:         req.Details.PaymentTerms,
		PricingType:          req.Details.PricingType,
		SalesPerson:          req.Details.SalesPerson,
	}

	detailsJSON, _ := json.Marshal(updatedDetails)
	p.Details = datatypes.JSON(detailsJSON)

	if err := tx.Save(&p); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to update profile",
			Code:  "DATABASE_ERROR",
		})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to commit",
			Code:  "DATABASE_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": p})
}

// Delete
func (h *Handler) DeleteProfile(c *gin.Context) {
	db := middleware.GetDB(c)

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Profile id is required",
			Code:  "VALIDATION_ERROR",
		})
		return
	}

	var profile models.Profile

	if err := db.First(&profile, "id = ?", id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "Profile not found",
				Code:  "NOT_FOUND",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to fetch profile",
			Code:  "DATABASE_ERROR",
		})
		return
	}

	if err := db.Delete(&profile); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to delete profile",
			Code:  "DATABASE_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile deleted successfully",
	})
}
