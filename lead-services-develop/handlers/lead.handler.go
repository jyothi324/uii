package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"

	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tuliporg/agent-management-api/internal/activity"
	"github.com/tuliporg/agent-management-api/middleware"
	"github.com/tuliporg/agent-management-api/models"
	"github.com/tuliporg/agent-management-api/services"
	"github.com/tuliporg/agent-management-api/utils"
	"gorm.io/gorm"
	// "gorm.io/gorm/clause"
)

type LeadHandler struct {
	s3Client *s3.S3
	bucket   string
	region   string
}

func (h *LeadHandler) enrichFilesWithPresignedURLs(c *gin.Context, files []models.File) {
	for i := range files {
		key := files[i].FilePath
		if key == "" {
			continue
		}
		url, err := h.generatePresignedURL(c, key)
		if err == nil {
			files[i].PresignedURL = url
		}
	}
}

func NewLeadHandler(s3Client *s3.S3, bucket, region string) *LeadHandler {
	return &LeadHandler{
		s3Client: s3Client,
		bucket:   bucket,
		region:   region,
	}
}

func (h *LeadHandler) getBucketName(c *gin.Context) string {
	subdomain := middleware.GetRawSubdomain(c)
	if subdomain == "" || subdomain == "default" {
		return h.bucket
	}
	return fmt.Sprintf("tulip-%s-files", subdomain)
}

func getRequestUser(c *gin.Context) (string, string) {
	userID, userName := "", ""

	// Priority 1: Check Gin Context (populated by AuthMiddleware)
	if val, ok := c.Get("user_id"); ok {
		userID, _ = val.(string)
	}
	if val, ok := c.Get("user_name"); ok {
		userName, _ = val.(string)
	}
	if userName == "" {
		if val, ok := c.Get("email"); ok {
			userName, _ = val.(string)
		}
	}

	// Priority 2: Check standard and custom headers (resilience for proxies)
	if userID == "" {
		userID = c.GetHeader("X-User-Id")
		if userID == "" {
			userID = c.GetHeader("X-User-ID")
		}
		if userID == "" {
			userID = c.GetHeader("x-user-id")
		}
		if userID == "" {
			userID = c.GetHeader("X-Agent-ID")
		}
	}
	if userName == "" {
		userName = c.GetHeader("X-User-Name")
		if userName == "" {
			userName = c.GetHeader("x-user-name")
		}
		if userName == "" {
			userName = c.GetHeader("User-Email")
		}
	}

	if userID == "" {
		userID = c.GetHeader("X-Request-User-Id")
		if userID == "" {
			userID = c.GetHeader("X-Request-User-ID")
		}
	}
	if userName == "" {
		userName = c.GetHeader("X-Request-User-Name")
	}

	if userID == "" {
		userID = "system"
	}
	if userName == "" {
		userName = "System"
	}

	return userID, userName
}

func residentialEntryHasEstimate(entry models.ResidentialEntry) bool {
	return entry.EstimateID != nil && strings.TrimSpace(*entry.EstimateID) != ""
}

func leadManagedResidentialEntries(entries []models.ResidentialEntry) []models.ResidentialEntry {
	leadEntries := make([]models.ResidentialEntry, 0, len(entries))
	for _, entry := range entries {
		if !residentialEntryHasEstimate(entry) {
			leadEntries = append(leadEntries, entry)
		}
	}
	return leadEntries
}

func (h *LeadHandler) CreateLead(c *gin.Context) {
	db := middleware.GetDB(c)
	// Parse multipart form (10 MB limit)
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		log.Printf("DEBUG: Failed to parse multipart form: %v", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Failed to parse form",
			Code:    "VALIDATION_ERROR",
			Details: err.Error(),
		})
		return
	}
	var residentialEntries []models.ResidentialEntry
	if entriesJSON := c.PostForm("ResidentialEntries"); entriesJSON != "" {
		if err := json.Unmarshal([]byte(entriesJSON), &residentialEntries); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Failed to parse residential entries",
				Code:    "VALIDATION_ERROR",
				Details: err.Error(),
			})
			return
		}
	}
	var CountyTax models.CountyTax
	id := c.Request.FormValue("Details[TaxID]")
	id1 := json.Number(id)

	if id1 != "" {
		if err := db.First(&CountyTax, "id = ?", id1); err != nil {
			if err == gorm.ErrRecordNotFound {
			} else {
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "Database error while fetching tax",
					Code:    "DATABASE_ERROR",
					Details: err.Error(),
				})
				return
			}
		} else {
		}
	} else {
	}
	// var Discount models.Discount
	// discountid := c.PostForm("Details[DiscountID]")

	// if discountid != "" {

	// 	if err := db.First(&Discount, "id = ?", discountid); err != nil {
	// 		// If the record isn't found, just skip (since Discount is optional)
	// 		if err == gorm.ErrRecordNotFound {
	// 		} else {
	// 			c.JSON(http.StatusInternalServerError, ErrorResponse{
	// 				Error:   "Database error while fetching discount",
	// 				Code:    "DATABASE_ERROR",
	// 				Details: err.Error(),
	// 			})
	// 			return
	// 		}
	// 	} else {
	// 	}
	// }
	// Parse Associates
	var associates []models.AssociateCustomer
	if assocJSON := c.PostForm("associates"); assocJSON != "" {
		if err := json.Unmarshal([]byte(assocJSON), &associates); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Failed to parse associates",
				Code:    "VALIDATION_ERROR",
				Details: err.Error(),
			})
			return
		}
	} else {
	}
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Unexpected error occurred",
				Code:    "INTERNAL_SERVER_ERROR",
				Details: fmt.Sprintf("%v", r),
			})
		}
	}()
	newLead := models.Lead{
		ID:                 utils.GenerateEmployeeID(db),
		JobName:            c.PostForm("JobName"),
		DeliveryType:       models.DeliveryType(c.PostForm("DeliveryType")),
		Status:             models.LeadStatus("NEW"),
		DetailedLeadStatus: models.DetailedLeadStatus("ESTIMATE_NOT_SENT"),
		LeadType:           models.LeadType(c.PostForm("LeadType")),
		PaymentTerm:        models.PaymentTerm(c.PostForm("PaymentTerm")),
		PricingType:        models.PricingType(c.PostForm("PricingType")),
		Favorite:           c.PostForm("Favorite"),
		Sales1PersonID:     c.PostForm("Sales1PersonID"),
		Sales2PersonID:     c.PostForm("Sales2PersonID"),
	}
	if newLead.Sales1PersonID != "" {
		var s1 models.Agent
		if err := db.First(&s1, "id = ?", newLead.Sales1PersonID); err == nil {
			newLead.Sales1PersonName = s1.Name
		}
	}
	if newLead.Sales2PersonID != "" {
		var s2 models.Agent
		if err := db.First(&s2, "id = ?", newLead.Sales2PersonID); err == nil {
			newLead.Sales2PersonName = s2.Name
		}
	}
	// Initialize Details if nil
	newLead.Details = models.LeadDetails{}
	// Always set tax & discount, even if other details are missing
	newLead.Details.CountyTax = &CountyTax
	// newLead.Details.Discount = &Discount
	// Then set other optional details
	if phone := c.PostForm("Details[Phone]"); phone != "" {
		newLead.Details.Phone = phone
	}
	if secondaryPhone := c.PostForm("Details[SecondaryPhone]"); secondaryPhone != "" {
		newLead.Details.Secondaryphone = secondaryPhone
	}
	if cellPhone := c.PostForm("Details[CellPhone]"); cellPhone != "" {
		newLead.Details.Cellphone = cellPhone
	}
	if email := c.PostForm("Details[ShipToEmail]"); email != "" {
		newLead.Details.ShipToEmail = email
	}
	if secondaryEmail := c.PostForm("Details[SecondaryEmail]"); secondaryEmail != "" {
		newLead.Details.SecondaryEmail = secondaryEmail
	}
	customerPhone := c.PostForm("Details[CustomerPhone]")
	if customerPhone == "" {
		customerPhone = c.PostForm("CustomerPhone")
	}
	if customerPhone != "" {
		newLead.Details.CustomerPhone = customerPhone
	}
	customerEmail := c.PostForm("Details[CustomerEmail]")
	if customerEmail == "" {
		customerEmail = c.PostForm("CustomerEmail")
	}
	if customerEmail != "" {
		newLead.Details.CustomerEmail = customerEmail
	}
	if isCompanyContact := c.PostForm("IsCompanyContact"); isCompanyContact == "true" {
		newLead.Details.IsCompanyContact = true
	}
	if billTo := c.PostForm("Details[BillToCustomer]"); billTo != "" {
		newLead.Details.BillToCustomer = billTo
	}
	if billTo := c.PostForm("Details[MobileNumber]"); billTo != "" {
		newLead.Details.MobileNumber = billTo
	}
	if billTo := c.PostForm("Details[FaxNumber]"); billTo != "" {
		newLead.Details.FaxNumber = billTo
	}
	var notesHistory []models.NoteEntry
	if rawHistory := c.PostForm("Details[NotesHistory]"); rawHistory != "" {
		if err := json.Unmarshal([]byte(rawHistory), &notesHistory); err != nil {
			notesHistory = []models.NoteEntry{}
		}
	}
	newLead.Details.NotesHistory = notesHistory
	// Validate required fields
	if newLead.JobName == "" || newLead.Status == "" {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Missing required fields",
			Code:    "VALIDATION_ERROR",
			Details: "JobName, and Status are required",
		})
		return
	}
	shippingStreet := c.PostForm("Details[ShippingAddress][Street]")
	if shippingStreet != "" {
		shippingAddress := models.Address{
			Street:    c.PostForm("Details[ShippingAddress][Street]"),
			UnitSuite: c.PostForm("Details[ShippingAddress][UnitSuite]"),
			City:      c.PostForm("Details[ShippingAddress][City]"),
			State:     c.PostForm("Details[ShippingAddress][State]"),
			Zip:       c.PostForm("Details[ShippingAddress][Zip]"),
			County:    c.PostForm("Details[ShippingAddress][Country]"),
		}

		rawAddress := fmt.Sprintf(
			"%s %s %s %s %s",
			shippingAddress.Street,
			shippingAddress.UnitSuite,
			shippingAddress.City,
			shippingAddress.State,
			shippingAddress.Zip,
		)

		f, err := strconv.ParseFloat(c.PostForm("Details[ShippingAddress][Latitude]"), 64)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Invalid latitude value",
				Code:    "VALIDATION_ERROR",
				Details: err.Error(),
			})
			return
		}
		shippingAddress.Lat = f
		g, err := strconv.ParseFloat(c.PostForm("Details[ShippingAddress][Longitude]"), 64)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Invalid longitude value",
				Code:    "VALIDATION_ERROR",
				Details: err.Error(),
			})
			return
		}
		shippingAddress.Lng = g
		shippingAddress.Formatted = rawAddress

		// Save address to DB

		if err := tx.Create(&shippingAddress); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to save shipping address",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
			return
		}

		newLead.Details.ShippingAddressID = &shippingAddress.ID

	} else {
	}

	billingStreet := c.PostForm("Details[BillingAddress][Street]")

	if billingStreet != "" {
		billingAddress := models.Address{
			Street:    c.PostForm("Details[BillingAddress][Street]"),
			UnitSuite: c.PostForm("Details[BillingAddress][UnitSuite]"),
			City:      c.PostForm("Details[BillingAddress][City]"),
			State:     c.PostForm("Details[BillingAddress][State]"),
			Zip:       c.PostForm("Details[BillingAddress][Zip]"),
			County:    c.PostForm("Details[BillingAddress][Country]"),
		}

		if err := tx.Create(&billingAddress); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to save billing address",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
			return
		}
		newLead.Details.BillingAddressID = &billingAddress.ID
	}

	isSameAddress := c.PostForm("isSameAddress")
	if isSameAddress == "true" && newLead.Details.ShippingAddressID != nil {
		if newLead.Details.BillingAddressID != nil {
			var billingAddress models.Address
			if err := tx.First(&billingAddress, *newLead.Details.BillingAddressID); err == nil {
				if err := tx.Delete(&billingAddress); err != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, ErrorResponse{
						Error:   "Failed to update billing address",
						Code:    "DATABASE_ERROR",
						Details: err.Error(),
					})
					return
				}
			}

		}

		newLead.Details.BillingAddressID = newLead.Details.ShippingAddressID
	}

	// Create Pickup Delivery Address
	isPickupDelivery := c.PostForm("Details[IsPickUpDelivery]")

	if isPickupDelivery == "true" {
		newLead.Details.IsPickUpDelivery = true
		pickupLocationID := c.PostForm("Details[PickupDeliveryAddressID]")

		if pickupLocationID != "" {
			var pickup models.PickupLocations
			if err := tx.First(&pickup, pickupLocationID); err != nil {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, ErrorResponse{
					Error:   "Invalid pickup location",
					Code:    "VALIDATION_ERROR",
					Details: err.Error(),
				})
				return
			}
			newLead.Details.PickupDeliveryAddressID = &pickup.ID
		}
	}

	var validAssociates []*models.AssociateCustomer
	for _, associate := range associates {

		if associate.Details.FirstName == "" && associate.Details.LastName == "" && associate.Details.CompanyName == "" {
			continue
		}

		associate.Details.Email = strings.TrimSpace(associate.Details.Email)
		associate.Details.Phone = strings.TrimSpace(associate.Details.Phone)

		var existingAssociate models.AssociateCustomer
		found := false
		if associate.Details.Email != "" || associate.Details.Phone != "" {
			query := tx.Model(&models.AssociateCustomer{})
			conditions := []string{}
			args := []interface{}{}
			if associate.Details.Email != "" {
				conditions = append(conditions, "details->>'email' = ?")
				args = append(args, associate.Details.Email)
			}
			if associate.Details.Phone != "" {
				conditions = append(conditions, "details->>'phone' = ?")
				args = append(args, associate.Details.Phone)
			}
			if len(conditions) > 0 {
				query = query.Where(strings.Join(conditions, " OR "), args...)
				err := query.First(&existingAssociate)

				if err == nil {
					validAssociates = append(validAssociates, &existingAssociate)
					found = true
				} else if errors.Is(err, gorm.ErrRecordNotFound) {
					found = false
				} else {
					found = false
				}
			}
		}

		if !found {
			var company models.Customer
			if associate.Details.CompanyName != "" {
				if err := tx.Where("name = ?", associate.Details.CompanyName).First(&company); err != nil {
					if err == gorm.ErrRecordNotFound {
						company = models.Customer{
							ID:                          utils.GenerateEmployeeID(db),
							Name:                        associate.Details.CompanyName,
							Phone:                       associate.Details.CompanyPhone,
							Email:                       associate.Details.CompanyEmail,
							SecondaryEmail:              associate.Details.CompanySecondaryEmail,
							Address:                     associate.Details.CompanyAddress,
							BillingAddress:              associate.Details.CompanyBillingAddress,
							Details:                     associate.Details.CompanyDetails,
							ContactName:                 associate.Details.CompanyContactName,
							SalesPerson1:                associate.Details.CompanySalesPerson1,
							SalesPerson2:                associate.Details.CompanySalesPerson2,
							ProjectManager:              associate.Details.CompanyProjectManager,
							CustomerSince:               associate.Details.CompanyCustomerSince,
							SalesTax:                    associate.Details.CompanySalesTax,
							EINNumber:                   associate.Details.CompanyEINNumber,
							TaxExempt:                   associate.Details.CompanyTaxExempt,
							TaxExemptReason:             associate.Details.CompanyTaxExemptReason,
							ExemptCertificateNumber:     associate.Details.CompanyExemptCertificateNumber,
							ExemptExpiryDate:            associate.Details.CompanyExemptExpiryDate,
							SpecialDeliveryInstructions: associate.Details.CompanySpecialDeliveryInstructions,
							CollectionNotes:             associate.Details.CompanyCollectionNotes,
						}
						if err := tx.Create(&company); err != nil {
							tx.Rollback()
							c.JSON(http.StatusInternalServerError, ErrorResponse{
								Error:   "Failed to create company",
								Code:    "DATABASE_ERROR",
								Details: err.Error(),
							})
							return
						}
					} else {
						tx.Rollback()
						c.JSON(http.StatusInternalServerError, ErrorResponse{
							Error:   "Database error",
							Code:    "DATABASE_ERROR",
							Details: err.Error(),
						})
						return
					}
				} else {
					// Update existing company with phone and email if provided
					needsUpdate := false
					if associate.Details.CompanyPhone != "" && company.Phone != associate.Details.CompanyPhone {
						company.Phone = associate.Details.CompanyPhone
						needsUpdate = true
					}
					if associate.Details.CompanyEmail != "" && company.Email != associate.Details.CompanyEmail {
						company.Email = associate.Details.CompanyEmail
						needsUpdate = true
					}
					if associate.Details.CompanySecondaryEmail != "" && company.SecondaryEmail != associate.Details.CompanySecondaryEmail {
						company.SecondaryEmail = associate.Details.CompanySecondaryEmail
						needsUpdate = true
					}
					if needsUpdate {
						if err := tx.Save(&company); err != nil {
							fmt.Printf("DEBUG: Failed to update company: %v\n", err)
						} else {
							fmt.Printf("DEBUG: Company updated with new contact info\n")
						}
					}
				}
			}

			newAssociate := models.AssociateCustomer{
				ID:   utils.GenerateEmployeeID(db),
				Role: models.AssociateRole(associate.Role),
				Details: models.AssociateDetails{
					FirstName:                          associate.Details.FirstName,
					LastName:                           associate.Details.LastName,
					Email:                              associate.Details.Email,
					Phone:                              associate.Details.Phone,
					CompanyName:                        associate.Details.CompanyName,
					CompanyPhone:                       associate.Details.CompanyPhone,
					CompanyEmail:                       associate.Details.CompanyEmail,
					CompanySecondaryEmail:              associate.Details.CompanySecondaryEmail,
					CompanyContactName:                 associate.Details.CompanyContactName,
					CompanyAddress:                     associate.Details.CompanyAddress,
					CompanyBillingAddress:              associate.Details.CompanyBillingAddress,
					CompanyDetails:                     associate.Details.CompanyDetails,
					CompanySalesPerson1:                associate.Details.CompanySalesPerson1,
					CompanySalesPerson2:                associate.Details.CompanySalesPerson2,
					CompanyProjectManager:              associate.Details.CompanyProjectManager,
					CompanyCustomerSince:               associate.Details.CompanyCustomerSince,
					CompanySalesTax:                    associate.Details.CompanySalesTax,
					CompanyEINNumber:                   associate.Details.CompanyEINNumber,
					CompanyTaxExempt:                   associate.Details.CompanyTaxExempt,
					CompanyTaxExemptReason:             associate.Details.CompanyTaxExemptReason,
					CompanyExemptCertificateNumber:     associate.Details.CompanyExemptCertificateNumber,
					CompanyExemptExpiryDate:            associate.Details.CompanyExemptExpiryDate,
					CompanySpecialDeliveryInstructions: associate.Details.CompanySpecialDeliveryInstructions,
					CompanyCollectionNotes:             associate.Details.CompanyCollectionNotes,
					CompanyID:                          company.ID,
					Company:                            company,
				},
			}
			if err := tx.Create(&newAssociate); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "Failed to save associate",
					Code:    "DATABASE_ERROR",
					Details: err.Error(),
				})
				return
			}
			validAssociates = append(validAssociates, &newAssociate)
		}
	}

	// Handle Profile Picture Upload
	form, err := c.MultipartForm()

	if err == nil {
		files := form.File["Profile"]
		for _, fileHeader := range files {
			category := "Profile"
			file, err := fileHeader.Open()
			if err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "Failed to open file",
					Code:    "FILE_ERROR",
					Details: err.Error(),
				})
				return
			}
			fileExt := filepath.Ext(fileHeader.Filename)
			s3Key := fmt.Sprintf("files/%s%s", uuid.New().String(), fileExt)
			newFilename := s3Key

			bucket := h.getBucketName(c)
			// Upload to S3
			_, err = h.s3Client.PutObject(&s3.PutObjectInput{
				Bucket: aws.String(bucket),
				Key:    aws.String(s3Key),
				Body:   file,
			})

			if err != nil {
				fmt.Printf("DEBUG: Failed to upload file to S3: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file"})
				return
			}

			// url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucket, h.region, s3Key)
			newLead.ProfilePic = s3Key

			// Get file size for database record
			size, err := file.Seek(0, io.SeekEnd)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get file size"})
				return
			}

			// Reset file pointer to beginning for upload
			file.Seek(0, io.SeekStart)
			fileSize := int64(size)

			newFile := models.File{
				LeadID:         &newLead.ID,
				Name:           fileHeader.Filename,
				Type:           fileExt[1:],
				Size:           fmt.Sprintf("%d", fileSize),
				UploadedAt:     time.Now(),
				Category:       category,
				StoredFileName: newFilename,
				FilePath:       s3Key,
			}
			if err := tx.Create(&newFile); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "Failed to save file to database",
					Code:    "DATABASE_ERROR",
					Details: err.Error(),
				})
				return
			}
		}
	}

	if err == nil {
		form := c.Request.MultipartForm

		files := form.File["Files"]

		for i, fileHeader := range files {

			categoryKey := fmt.Sprintf("Files[%d][Category]", i)
			projectTypeKey := fmt.Sprintf("Files[%d][ProjectType]", i)

			category := ""
			projecttype := ""

			if val, ok := form.Value[categoryKey]; ok && len(val) > 0 {
				category = val[0]
			}
			if val, ok := form.Value[projectTypeKey]; ok && len(val) > 0 {
				projecttype = val[0]
			}
			file, err := fileHeader.Open()
			if err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "Failed to open file",
					Code:    "FILE_ERROR",
					Details: err.Error(),
				})
				return
			}

			// Read file content into memory first
			// fileContent, err := io.ReadAll(file)
			// Read file content into memory first
			// fileContent, err := io.ReadAll(file)
			file.Close() // Close the original file reader after successful open
			fileExt := filepath.Ext(fileHeader.Filename)
			uniqueID := uuid.New().String()
			s3Key := fmt.Sprintf("files/%s%s", uniqueID, fileExt)
			newFilename := fmt.Sprintf("%s%s", uniqueID, fileExt)
			// newLead.ProfilePic = fmt.Sprintf("%s%s", uniqueID, fileExt) // Store the S3 key instead of just filename

			bucket := h.getBucketName(c)
			// Upload to S3
			_, err = h.s3Client.PutObject(&s3.PutObjectInput{
				Bucket:             aws.String(bucket),
				Key:                aws.String(s3Key),
				Body:               file,
				ContentDisposition: aws.String("attachment"),
				ContentType:        aws.String(fileHeader.Header.Get("Content-Type")),
			})
			if err != nil {
				fmt.Printf("DEBUG: Failed to upload file to S3: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file"})
				return
			}
			// url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucket, h.region, s3Key)

			// Get file size for database record
			fileSize := fileHeader.Size

			newFile := models.File{
				ID:             uint(rand.Uint32()),
				LeadID:         &newLead.ID,
				Name:           fileHeader.Filename,
				Type:           fileHeader.Header.Get("Content-Type"), // safer than fileExt[1:]
				Size:           fmt.Sprintf("%d", fileSize),
				UploadedAt:     time.Now(),
				Category:       category,
				ProjectType:    projecttype,
				StoredFileName: newFilename,
				FilePath:       s3Key,
			}

			if err := tx.Create(&newFile); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "Failed to save file to database",
					Code:    "DATABASE_ERROR",
					Details: err.Error(),
				})
				return
			}
		}
	}
	if err := tx.Create(&newLead); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to save lead",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}
	for _, associate := range validAssociates {
		if err := tx.Exec("INSERT INTO lead_associates (lead_id, associate_customer_id) VALUES (?, ?)", newLead.ID, associate.ID); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to link associate",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
			return
		}
	}

	// Handle File Uploads
	fmt.Println("DEBUG: Processing additional file uploads")
	// pretty.Print(c.PostForm("files"), "files done")
	// leadID := newLead.ID
	// pretty.Print(leadID, "lead") // Debug: Print leadID

	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse multipart form"})
	// 	return
	// }

	// files := form.File["files"]

	// // Check if any files were uploaded
	// if len(files) == 0 {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "No files uploaded"})
	// 	return
	// }
	// pretty.Print(files[0].Filename, "fil") // Debug: Print first file name
	// uploadDir := "filestorage"
	// if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
	// 	return
	// }

	// for _, fileHeader := range files {
	// 	// Generate unique filename
	// 	storedFileName := uuid.New().String() + filepath.Ext(fileHeader.Filename)
	// 	storedPath := filepath.Join(uploadDir, storedFileName)

	// 	// Save file to disk
	// 	if err := c.SaveUploadedFile(fileHeader, storedPath); err != nil {
	// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file to disk"})
	// 		return
	// 	}

	// 	// Build file model
	// 	file := models.File{
	// 		Name:           fileHeader.Filename,
	// 		Type:           fileHeader.Header.Get("Content-Type"),
	// 		Size:           fmt.Sprintf("%d", fileHeader.Size),
	// 		UploadedAt:     time.Now(),
	// 		StoredFileName: storedFileName,
	// 		Category:       "Other",
	// 		ProjectType:    "Other",
	// 		FilePath:       storedPath,
	// 	}

	// 	// Attach leadID if provided
	// 	if leadID != "" {
	// 		file.LeadID = &leadID
	// 	}

	// 	// Save to DB
	// 	if err := db.Create(&file); err != nil {
	// 		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("DB save failed: %v", err)})
	// 		return
	// 	}
	// }
	// Handle Residential Entries
	// var jobid = utils.GenerateEmployeeID(db)
	fmt.Printf("DEBUG: Processing %d residential entries\n", len(residentialEntries))
	for i := range residentialEntries {
		residentialEntries[i].LeadID = &newLead.ID

		if err := normalizeResidentialEntryLookups(tx, &residentialEntries[i]); err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, residentialEntryLookupErrorResponse(err))
			return
		}

		if err := tx.Create(&residentialEntries[i]); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to save residential entry",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
			return
		}

		fmt.Printf("DEBUG: Residential entry %d created successfully\n", i)
		// defaultStages := []models.Stage{
		// 	{
		// 		ID:                 uuid.NewString(),
		// 		LeadID:             newLead.ID,
		// 		JobID:              jobid,
		// 		Name:               "Template",
		// 		State:              "In Progress", // First stage starts as In Progress
		// 		CreatedAt:          time.Now(),
		// 		Date:               func() *time.Time { t := time.Now(); return &t }(),
		// 		Description:        "Prepare and finalize the template for the project",
		// 		ResidentialEntryID: residentialEntries[i].ID,
		// 	},
		// 	{
		// 		ID:                 uuid.NewString(),
		// 		JobID:              jobid,
		// 		LeadID:             newLead.ID,
		// 		Name:               "CAD",
		// 		State:              "Not Started",
		// 		CreatedAt:          time.Now(),
		// 		Description:        "Create CAD designs for the project",
		// 		ResidentialEntryID: residentialEntries[i].ID,
		// 	},
		// 	{
		// 		ID:                 uuid.NewString(),
		// 		JobID:              jobid,
		// 		LeadID:             newLead.ID,
		// 		Name:               "Cutting",
		// 		State:              "Not Started",
		// 		CreatedAt:          time.Now(),
		// 		Description:        "Cut materials according to specifications",
		// 		ResidentialEntryID: residentialEntries[i].ID,
		// 	},
		// 	{
		// 		ID:                 uuid.NewString(),
		// 		JobID:              jobid,
		// 		LeadID:             newLead.ID,
		// 		Name:               "Fabrication",
		// 		State:              "Not Started",
		// 		CreatedAt:          time.Now(),
		// 		Description:        "Fabricate components for the project",
		// 		ResidentialEntryID: residentialEntries[i].ID,
		// 	},
		// 	{
		// 		ID:                 uuid.NewString(),
		// 		JobID:              jobid,
		// 		LeadID:             newLead.ID,
		// 		Name:               "Installation",
		// 		State:              "Not Started",
		// 		CreatedAt:          time.Now(),
		// 		Description:        "Install the project components on-site",
		// 		ResidentialEntryID: residentialEntries[i].ID,
		// 	},
		// 	{
		// 		ID:                 uuid.NewString(),
		// 		JobID:              jobid,
		// 		LeadID:             newLead.ID,
		// 		Name:               "Service Call",
		// 		State:              "Not Started",
		// 		CreatedAt:          time.Now(),
		// 		Description:        "Handle any service calls or maintenance",
		// 		ResidentialEntryID: residentialEntries[i].ID,
		// 	},
		// }

		// // Create default stages
		// fmt.Println("Creating default stages for lead:", newLead.ID)
		// for _, stage := range defaultStages {
		// 	if err := tx.Create(&stage); err != nil {
		// 		tx.Rollback()
		// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create default stages: " + err.Error()})
		// 		return
		// 	}
		// 	fmt.Println("Stage created:", stage.Name)
		// }
	}
	if err := incrementManagerResidentialLengths(tx, leadManagedResidentialEntries(residentialEntries)); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to update residential entry usage",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}
	newLead.ResidentialEntries = residentialEntries

	userID, userName := getRequestUser(c)
	creatorLabel := userName
	if creatorLabel == "" {
		creatorLabel = userID
	}

	// Attribution fallback to assigned salesperson if logged-in user is unavailable
	if (userID == "system" || userID == "") && newLead.Sales1PersonID != "" {
		userID = newLead.Sales1PersonID
		if newLead.Sales1PersonName != "" {
			userName = newLead.Sales1PersonName
			creatorLabel = userName
		}
	}

	log.Printf("lead created, publishing activity event for lead=%s user_id=%s user_name=%s", newLead.ID, userID, userName)

	event := activity.ActivityEvent{
		EntityType: "lead",
		EntityID:   newLead.ID,
		Action:     "created",
		UserID:     userID,
		UserName:   userName,
		Payload: map[string]any{
			"lead_name":  newLead.JobName,
			"status":     newLead.Status,
			"created_by": creatorLabel,
		},
	}
	if err := activity.InsertActivityEvent(c.Request.Context(), tx.GetUnderlying(), event); err != nil { // This line was already correct from previous fix
		log.Printf("activity event publish failed: %v", err)
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
		"message": "Lead created successfully",
		"id":      newLead.ID,
	})
}

// GetLead retrieves a lead by ID
func (h *LeadHandler) GetLead(c *gin.Context) {
	db := middleware.GetDB(c)

	// Extract and validate ID from URL parameter
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid lead ID",
			Code:  "BAD_REQUEST",
		})
		return
	}

	var lead models.Lead

	query := db.Model(&models.Lead{}).
		Where("id = ?", id).
		Preload("Associates").
		Preload("ResidentialEntries.ManagerResidentialEntry").
		Preload("ResidentialEntries.ManagerResidentialType").
		Preload("Estimates.MaterialRows").
		Preload("Estimates.ServiceRows").
		Preload("Files").
		Preload("Schedule")

	if err := query.Find(&lead); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to retrieve leads",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}
	if lead.Details.ShippingAddressID != nil {
		var shippingAddress models.Address
		if err := db.First(&shippingAddress, *lead.Details.ShippingAddressID); err != nil {
			// Log the error but don't fail; leave ShippingAddress as nil
			// e.g., log.Printf("Failed to fetch ShippingAddress for lead %s: %v", leads[i].ID, err)
		} else {
			lead.Details.ShippingAddress = &shippingAddress
		}
	}
	// var associates []*models.AssociateCustomer
	// if err := db.Model(&lead).Association("Associates").Find(&associates); err != nil {
	// 	return
	// }
	// Preload BillingAddress
	if lead.Details.BillingAddressID != nil {
		var billingAddress models.Address
		if err := db.First(&billingAddress, *lead.Details.BillingAddressID); err != nil {
			// Log the error but don't fail
		} else {
			lead.Details.BillingAddress = &billingAddress
		}
	}
	if lead.Sales1PersonID != "" {
		var Sales1Person models.Agent
		if err := db.First(&Sales1Person, &lead.Sales1PersonID); err != nil {
			// Log the error but don't fail
		} else {
			lead.Sales1Person = &Sales1Person
		}
	}
	if lead.Sales2PersonID != "" {
		var Sales2Person models.Agent
		if err := db.First(&Sales2Person, &lead.Sales2PersonID); err != nil {
		} else {
			lead.Sales2Person = &Sales2Person
		}
	}
	if lead.Details.PickupDeliveryAddressID != nil {
		var pickupAddress models.PickupLocations
		if err := db.First(&pickupAddress, *lead.Details.PickupDeliveryAddressID); err != nil {
			// Log the error but don't fail
		} else {
			lead.Details.PickupDeliveryAddress = &pickupAddress
		}
	}
	// Return success response
	h.enrichFilesWithPresignedURLs(c, lead.Files)
	c.JSON(http.StatusOK, gin.H{
		"message": "Lead retrieved successfully",
		"data":    lead,
	})
}
func (h *LeadHandler) GetLeadByPhoneOrEmail(c *gin.Context) {
	db := middleware.GetDB(c)

	// Extract query parameters
	id := c.Query("id")
	email := c.Query("email")
	phone := c.Query("phone")

	// Ensure at least one identifier is provided
	if id == "" && email == "" && phone == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Must provide at least one of: id, email, or phone",
			Code:  "BAD_REQUEST",
		})
		return
	}

	var lead models.Lead

	// Build the query based on which parameter is provided
	query := db.Model(&models.Lead{}).
		Preload("Associates").
		Preload("ResidentialEntries.ManagerResidentialEntry").
		Preload("ResidentialEntries.ManagerResidentialType").
		Preload("Estimates.MaterialRows").
		Preload("Estimates.ServiceRows").
		Preload("Files")

	if id != "" {
		query = query.Where("id = ?", id)
	} else if email != "" {
		query = query.Where("details->>'ShipToEmail' = ?", email)
	} else if phone != "" {
		query = query.Where("details->>'Phone' = ?", phone)
	}

	if err := query.First(&lead); err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Lead not found",
			Code:    "NOT_FOUND",
			Details: err.Error(),
		})
		return
	}

	// Load ShippingAddress if present
	if lead.Details.ShippingAddressID != nil {
		var shippingAddress models.Address
		if err := db.First(&shippingAddress, *lead.Details.ShippingAddressID); err == nil {
			lead.Details.ShippingAddress = &shippingAddress
		}
	}

	// Load Associates
	// var associates []*models.AssociateCustomer
	// if err := db.Model(&lead).Association("Associates").Find(&associates); err == nil {
	// 	lead.Associates = associates
	// }

	// Load BillingAddress
	if lead.Details.BillingAddressID != nil {
		var billingAddress models.Address
		if err := db.First(&billingAddress, *lead.Details.BillingAddressID); err == nil {
			lead.Details.BillingAddress = &billingAddress
		}
	}

	// Load Sales1Person
	if lead.Sales1PersonID != "" {
		var sales1 models.Agent
		if err := db.First(&sales1, "id = ?", lead.Sales1PersonID); err == nil {
			lead.Sales1Person = &sales1
		}
	}

	// Load Sales2Person
	if lead.Sales2PersonID != "" {
		var sales2 models.Agent
		if err := db.First(&sales2, "id = ?", lead.Sales2PersonID); err == nil {
			lead.Sales2Person = &sales2
		}
	}

	// Load PickupDeliveryAddress
	if lead.Details.PickupDeliveryAddressID != nil {
		var pickupAddress models.PickupLocations
		if err := db.First(&pickupAddress, *lead.Details.PickupDeliveryAddressID); err == nil {
			lead.Details.PickupDeliveryAddress = &pickupAddress
		}
	}

	// Return success response
	h.enrichFilesWithPresignedURLs(c, lead.Files)
	c.JSON(http.StatusOK, gin.H{
		"message": "Lead retrieved successfully",
		"data":    lead,
	})
}
func (h *LeadHandler) extractS3Key(url string) string {

	// Extract key from URL format: https://bucket.s3.region.amazonaws.com/key
	parts := strings.Split(url, "/")
	if len(parts) >= 4 {
		// Join all parts after the domain (in case the key contains slashes)
		return strings.Join(parts[3:], "/")
	}
	return ""
}
func (h *LeadHandler) generatePresignedURL(c *gin.Context, s3Key string) (string, error) {
	if s3Key == "" {
		return "", fmt.Errorf("empty s3Key")
	}
	// Support both legacy full URLs and new bare keys
	if strings.HasPrefix(s3Key, "https://") || strings.HasPrefix(s3Key, "http://") {
		s3Key = h.extractS3Key(s3Key)
	}
	req, _ := h.s3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket:                     aws.String(h.getBucketName(c)),
		Key:                        aws.String(s3Key),
		ResponseContentDisposition: attachmentDispositionForKey(s3Key),
	})
	return req.Presign(1 * time.Hour)
}
func (h *LeadHandler) ServeFile(c *gin.Context) {
	db := middleware.GetDB(c)

	filename := c.Param("filename")

	var file models.File
	if err := db.Where("stored_file_name = ?", filename).First(&file); err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"error":    "File not found in database",
			"filename": filename,
		})
		return
	}

	// Use FilePath directly if it's already a bare S3 key
	s3Key := file.FilePath
	if strings.HasPrefix(s3Key, "https://") || strings.HasPrefix(s3Key, "http://") {
		s3Key = h.extractS3Key(s3Key)
	}

	if s3Key == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid file URL"})
		return
	}

	url, err := h.generatePresignedURL(c, file.FilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate download URL"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"download_url": url,
		"filename":     file.Name,
		"expires_in":   "1 hour",
	})
}

func (h *LeadHandler) LeadDetails(c *gin.Context) {
	db := middleware.GetDB(c)

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid lead ID",
			Code:  "BAD_REQUEST",
		})
		return
	}
	// Begin transaction
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Unexpected error occurred",
				Code:    "INTERNAL_SERVER_ERROR",
				Details: fmt.Sprintf("%v", r),
			})
		}
	}()

	// Check if lead exists
	var existingLead models.Lead
	if err := tx.First(&existingLead, "id = ?", id); err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Lead not found",
				Code:    "NOT_FOUND",
				Details: "No lead found with the specified ID",
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

	var oldLead models.Lead
	if existingLeadJSON, err := json.Marshal(existingLead); err == nil {
		_ = json.Unmarshal(existingLeadJSON, &oldLead)
	}

	var updateData struct {
		BillToCustomer string
		CountyTax      json.RawMessage `json:"CountyTax"`
		Discount       string
		PaymentTerm    string
		PricingType    string
		Sales1PersonID string
		Sales2PersonID string
	}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Code:    "VALIDATION_ERROR",
			Details: err.Error(),
		})
		return
	}

	var CountyTax models.CountyTax
	if len(updateData.CountyTax) > 0 {
		var taxObj struct {
			ID            uint    `json:"ID"`
			Name          string  `json:"Name"`
			TaxPercentage float64 `json:"TaxPercentage"`
		}
		if err := json.Unmarshal(updateData.CountyTax, &taxObj); err == nil {
			if taxObj.Name == "other" {
				CountyTax.ID = 0
				CountyTax.Name = "other"
				CountyTax.TaxPercentage = float32(taxObj.TaxPercentage)
			} else if taxObj.ID > 0 {
				// Look up by ID
				if err := db.First(&CountyTax, "id = ?", taxObj.ID); err != nil {
					if err != gorm.ErrRecordNotFound {
						c.JSON(http.StatusInternalServerError, ErrorResponse{
							Error:   "Database error while fetching tax",
							Code:    "DATABASE_ERROR",
							Details: err.Error(),
						})
						return
					}
				}
			}
		} else {
			var taxStr string
			if err := json.Unmarshal(updateData.CountyTax, &taxStr); err == nil && taxStr != "" {
				if err := db.First(&CountyTax, "id = ?", taxStr); err != nil {
					if err != gorm.ErrRecordNotFound {
						c.JSON(http.StatusInternalServerError, ErrorResponse{
							Error:   "Database error while fetching tax",
							Code:    "DATABASE_ERROR",
							Details: err.Error(),
						})
						return
					}
				}
			}
		}
	}
	// var Discount models.Discount
	// if updateData.Discount != "" {

	// 	if err := db.First(&Discount, "id = ?", updateData.Discount); err != nil {
	// 		// If the record isn't found, just skip (since Discount is optional)
	// 		if err == gorm.ErrRecordNotFound {
	// 			// do NOT return; just continue
	// 		} else {
	// 			c.JSON(http.StatusInternalServerError, ErrorResponse{
	// 				Error:   "Database error while fetching discount",
	// 				Code:    "DATABASE_ERROR",
	// 				Details: err.Error(),
	// 			})
	// 			return
	// 		}
	// 	}
	// }
	if updateData.PaymentTerm != "" {
		existingLead.PaymentTerm = models.PaymentTerm(updateData.PaymentTerm)
	}
	if updateData.PricingType != "" {
		existingLead.PricingType = models.PricingType(updateData.PricingType)
	}
	if updateData.Sales1PersonID != "" {
		existingLead.Sales1PersonID = updateData.Sales1PersonID
		var s1 models.Agent
		if err := db.First(&s1, "id = ?", updateData.Sales1PersonID); err == nil {
			existingLead.Sales1PersonName = s1.Name
		}
	}
	if updateData.Sales2PersonID != "" {
		existingLead.Sales2PersonID = updateData.Sales2PersonID
		var s2 models.Agent
		if err := db.First(&s2, "id = ?", updateData.Sales2PersonID); err == nil {
			existingLead.Sales2PersonName = s2.Name
		}
	}
	// Update the favorite field
	existingLead.Details.BillToCustomer = updateData.BillToCustomer
	existingLead.Details.CountyTax = &CountyTax
	// existingLead.Details.Discount = &Discount

	leadChanges := getLeadChanges(oldLead, existingLead)

	userID, userName := getRequestUser(c)
	event := activity.ActivityEvent{
		EntityType: "lead",
		EntityID:   existingLead.ID,
		Action:     "details_updated",
		UserID:     userID,
		UserName:   userName,
		Payload: map[string]any{
			"updated_by":     userName,
			"changed_fields": leadChanges,
		},
	}
	if err := activity.InsertActivityEvent(c.Request.Context(), tx.GetUnderlying(), event); err != nil {
		log.Printf("activity event publish failed: %v", err)
	}

	if err := tx.Save(&existingLead); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to update lead",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to commit transaction",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Lead favorite status updated successfully",
		"lead":    existingLead,
	})

}

func (h *LeadHandler) AddLeadNote(c *gin.Context) {
	db := middleware.GetDB(c)
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid lead ID",
			Code:  "BAD_REQUEST",
		})
		return
	}

	var payload struct {
		Text             string `json:"text"`
		CompleteFollowUp bool   `json:"completeFollowUp"`
		CompletedDate    string `json:"completedDate"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Code:    "VALIDATION_ERROR",
			Details: err.Error(),
		})
		return
	}

	noteText := strings.TrimSpace(payload.Text)
	if noteText == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Note text is required",
			Code:    "VALIDATION_ERROR",
			Details: "text cannot be empty",
		})
		return
	}

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Unexpected error occurred",
				Code:    "INTERNAL_SERVER_ERROR",
				Details: fmt.Sprintf("%v", r),
			})
		}
	}()

	var lead models.Lead
	if err := tx.Preload("Schedule").First(&lead, "id = ?", id); err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Lead not found",
				Code:    "NOT_FOUND",
				Details: "No lead found with the specified ID",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Database error",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	note := models.NoteEntry{
		Text:      noteText,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	lead.Details.NotesHistory = append(lead.Details.NotesHistory, note)
	if payload.CompleteFollowUp {
		completedDate := strings.TrimSpace(payload.CompletedDate)
		if completedDate == "" && lead.Schedule != nil && lead.Schedule.ScheduledDate != nil {
			completedDate = lead.Schedule.ScheduledDate.Format("2006-01-02")
		}
		if completedDate != "" {
			if _, err := time.Parse("2006-01-02", completedDate); err != nil {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, ErrorResponse{
					Error:   "Invalid completed date format",
					Code:    "VALIDATION_ERROR",
					Details: err.Error(),
				})
				return
			}
			lead.Details.FollowUpCompletedDate = completedDate
		}
	}

	userID, userName := middleware.GetRequestUser(c)
	event := activity.ActivityEvent{
		EntityType: "lead",
		EntityID:   lead.ID,
		Action:     "note_added",
		UserID:     userID,
		UserName:   userName,
		Payload: map[string]any{
			"lead_name":       lead.JobName,
			"note_text":       note.Text,
			"note_created_at": note.CreatedAt,
			"updated_by":      userName,
		},
	}
	if err := activity.InsertActivityEvent(c.Request.Context(), tx.GetUnderlying(), event); err != nil {
		log.Printf("activity event publish failed: %v", err)
	}

	if err := tx.Save(&lead); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to save lead note",
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
		"message":               "Lead note added successfully",
		"note":                  note,
		"notesHistory":          lead.Details.NotesHistory,
		"followUpCompletedDate": lead.Details.FollowUpCompletedDate,
	})
}

func (h *LeadHandler) UpdateLead(c *gin.Context) {
	db := middleware.GetDB(c)

	// Extract and validate ID
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid lead ID",
			Code:  "BAD_REQUEST",
		})
		return
	}

	// Parse multipart form
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Failed to parse form",
			Code:    "VALIDATION_ERROR",
			Details: err.Error(),
		})
		return
	}
	// Begin transaction
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Unexpected error occurred",
				Code:    "INTERNAL_SERVER_ERROR",
				Details: fmt.Sprintf("%v", r),
			})
		}
	}()

	// Fetch existing lead with associations
	var existingLead models.Lead
	if err := tx.
		Preload("ResidentialEntries.ManagerResidentialEntry").
		Preload("Estimates").
		Preload("Associates").
		First(&existingLead, "id = ?", id); err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Lead not found",
				Code:    "NOT_FOUND",
				Details: "No lead found with the specified ID",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Database error",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}
	oldLeadManagedResidentialEntries := leadManagedResidentialEntries(existingLead.ResidentialEntries)

	var oldLead models.Lead
	existingLeadJSON, err := json.Marshal(existingLead)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to marshal existing lead for change tracking",
			Code:    "INTERNAL_SERVER_ERROR",
			Details: err.Error(),
		})
		return
	}
	if err := json.Unmarshal(existingLeadJSON, &oldLead); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to unmarshal existing lead for change tracking",
			Code:    "INTERNAL_SERVER_ERROR",
			Details: err.Error(),
		})
		return
	}

	oldShippingFormatted := ""
	if oldLead.Details.ShippingAddressID != nil {
		var oldShipping models.Address
		if err := db.First(&oldShipping, *oldLead.Details.ShippingAddressID); err == nil {
			oldShippingFormatted = formatAddress(&oldShipping)
		}
	}

	oldBillingFormatted := ""
	if oldLead.Details.BillingAddressID != nil {
		var oldBilling models.Address
		if err := db.First(&oldBilling, *oldLead.Details.BillingAddressID); err == nil {
			oldBillingFormatted = formatAddress(&oldBilling)
		}
	}

	updatedLead := models.Lead{
		ID:                 existingLead.ID,
		JobName:            c.PostForm("JobName"),
		DeliveryType:       models.DeliveryType(c.PostForm("DeliveryType")),
		Status:             models.LeadStatus(c.PostForm("Status")),
		DetailedLeadStatus: models.DetailedLeadStatus(c.PostForm("DetailedLeadStatus")),
		LeadType:           models.LeadType(c.PostForm("LeadType")),
		PaymentTerm:        models.PaymentTerm(c.PostForm("PaymentTerm")),
		PricingType:        models.PricingType(c.PostForm("PricingType")),
		// Favorite: c.PostForm("Favorite"),

		CreatedAt: existingLead.CreatedAt,
	}

	if _, ok := c.Request.Form["Sales1PersonID"]; ok {
		updatedLead.Sales1PersonID = c.PostForm("Sales1PersonID")
	} else {
		updatedLead.Sales1PersonID = existingLead.Sales1PersonID
	}
	if _, ok := c.Request.Form["Sales2PersonID"]; ok {
		updatedLead.Sales2PersonID = c.PostForm("Sales2PersonID")
	} else {
		updatedLead.Sales2PersonID = existingLead.Sales2PersonID
	}

	if updatedLead.Sales1PersonID != "" {
		var s1 models.Agent
		if err := tx.First(&s1, "id = ?", updatedLead.Sales1PersonID); err == nil {
			updatedLead.Sales1PersonName = s1.Name
		}
	} else {
		updatedLead.Sales1PersonName = ""
	}
	if updatedLead.Sales2PersonID != "" {
		var s2 models.Agent
		if err := tx.First(&s2, "id = ?", updatedLead.Sales2PersonID); err == nil {
			updatedLead.Sales2PersonName = s2.Name
		}
	} else {
		updatedLead.Sales2PersonName = ""
	}

	var notesHistory []models.NoteEntry
	if rawHistory := c.PostForm("Details[NotesHistory]"); rawHistory != "" {
		if err := json.Unmarshal([]byte(rawHistory), &notesHistory); err != nil {
			notesHistory = existingLead.Details.NotesHistory
		}
	} else {
		notesHistory = existingLead.Details.NotesHistory
	}

	// Update LeadDetails
	details := models.LeadDetails{
		Phone:          c.PostForm("Details[Phone]"),
		Secondaryphone: c.PostForm("Details[SecondaryPhone]"),
		Cellphone:      c.PostForm("Details[CellPhone]"),
		ShipToEmail:    c.PostForm("Details[ShipToEmail]"),
		SecondaryEmail: c.PostForm("Details[SecondaryEmail]"),
		BillToCustomer: c.PostForm("Details[BillToCustomer]"),
		FaxNumber:      c.PostForm("Details[FaxNumber]"),
		MobileNumber:   c.PostForm("Details[MobileNumber]"),
		CustomerPhone:  c.PostForm("Details[CustomerPhone]"),
		CustomerEmail:  c.PostForm("Details[CustomerEmail]"),
		// Notes:          c.PostForm("Details[Notes]"),
		NotesHistory: notesHistory,
		TaxID:        existingLead.Details.TaxID,
		// DiscountID:   existingLead.Details.DiscountID,
	}
	updatedLead.Details = details

	// Preserve existing values if not provided
	if updatedLead.JobName == "" {
		updatedLead.JobName = existingLead.JobName
	}
	if updatedLead.Status == "" {
		updatedLead.Status = existingLead.Status
	}
	if updatedLead.LeadType == "" {
		updatedLead.LeadType = existingLead.LeadType
	}
	if updatedLead.DeliveryType == "" {
		updatedLead.DeliveryType = existingLead.DeliveryType
	}
	if updatedLead.PaymentTerm == "" {
		updatedLead.PaymentTerm = existingLead.PaymentTerm
	}
	if updatedLead.PricingType == "" {
		updatedLead.PricingType = existingLead.PricingType
	}
	if updatedLead.Favorite == "" {
		updatedLead.Favorite = existingLead.Favorite
	}
	if details.Phone == "" {
		details.Phone = existingLead.Details.Phone
	}
	if details.Secondaryphone == "" {
		details.Secondaryphone = existingLead.Details.Secondaryphone
	}
	if details.Cellphone == "" {
		details.Cellphone = existingLead.Details.Cellphone
	}
	if isCompanyContact := c.PostForm("IsCompanyContact"); isCompanyContact == "true" {
		details.IsCompanyContact = true
	} else if isCompanyContact == "false" {
		details.IsCompanyContact = false
	}
	if details.ShipToEmail == "" {
		details.ShipToEmail = existingLead.Details.ShipToEmail
	}
	if details.BillToCustomer == "" {
		details.BillToCustomer = existingLead.Details.BillToCustomer
	}

	if details.CustomerPhone == "" {
		details.CustomerPhone = existingLead.Details.CustomerPhone
	}
	if details.CustomerEmail == "" {
		details.CustomerEmail = existingLead.Details.CustomerEmail
	}

	// Validate required fields
	if updatedLead.JobName == "" || updatedLead.Status == "" {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Missing required fields",
			Code:    "VALIDATION_ERROR",
			Details: "JobName and Status are required",
		})
		return
	}
	var countyTax models.CountyTax
	taxIDStr := c.PostForm("Details[CountyTax][ID]")
	taxName := c.PostForm("Details[CountyTax][Name]")
	taxPctStr := c.PostForm("Details[CountyTax][TaxPercentage]")

	if taxName == "other" {
		// Custom tax — trust the submitted percentage only for this explicit edit path
		var pct float64
		if taxPctStr != "" {
			pct, _ = strconv.ParseFloat(taxPctStr, 64)
		}
		countyTax.ID = 0
		countyTax.Name = "other"
		countyTax.TaxPercentage = float32(pct)
	} else if taxIDStr != "" && taxIDStr != "0" {
		// Standard tax — always look up from DB, never trust client percentage
		if err := tx.First(&countyTax, "id = ?", taxIDStr); err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "Database error while fetching tax record",
					Code:    "DATABASE_ERROR",
					Details: err.Error(),
				})
				return
			}
		}
	} else if existingLead.Details.CountyTax != nil {
		countyTax = *existingLead.Details.CountyTax
	}
	existingLead.Details.CountyTax = &countyTax

	newShippingFormatted := oldShippingFormatted
	if shippingStreet := c.PostForm("Details[ShippingAddress][Street]"); shippingStreet != "" {
		var shippingAddress models.Address
		if existingLead.Details.ShippingAddressID != nil {
			if err := tx.First(&shippingAddress, *existingLead.Details.ShippingAddressID); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "Failed to find shipping address",
					Code:    "DATABASE_ERROR",
					Details: err.Error(),
				})
				return
			}
		}
		shippingAddress.Street = c.PostForm("Details[ShippingAddress][Street]")
		shippingAddress.UnitSuite = c.PostForm("Details[ShippingAddress][UnitSuite]")
		shippingAddress.City = c.PostForm("Details[ShippingAddress][City]")
		shippingAddress.State = c.PostForm("Details[ShippingAddress][State]")
		shippingAddress.Zip = c.PostForm("Details[ShippingAddress][Zip]")
		shippingAddress.County = c.PostForm("Details[ShippingAddress][County]")
		rawAddress := fmt.Sprintf(
			"%s %s %s %s %s",
			shippingAddress.Street,
			shippingAddress.UnitSuite,
			shippingAddress.City,
			shippingAddress.State,
			shippingAddress.Zip,
		)

		// Geocode via OpenCage (if API key available)
		f, err := strconv.ParseFloat(c.PostForm("Details[ShippingAddress][Latitude]"), 64)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Invalid latitude value",
				Code:    "VALIDATION_ERROR",
				Details: err.Error(),
			})
			return
		}
		shippingAddress.Lat = f
		g, err := strconv.ParseFloat(c.PostForm("Details[ShippingAddress][Longitude]"), 64)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Invalid longitude value",
				Code:    "VALIDATION_ERROR",
				Details: err.Error(),
			})
			return
		}
		shippingAddress.Lng = g
		shippingAddress.Formatted = rawAddress

		if err := tx.Save(&shippingAddress); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to save shipping address",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
			return
		}
		updatedLead.Details.ShippingAddressID = &shippingAddress.ID
		updatedLead.Details.ShippingAddress = &shippingAddress
		newShippingFormatted = formatAddress(&shippingAddress)
	}

	newBillingFormatted := oldBillingFormatted
	if billingStreet := c.PostForm("Details[BillingAddress][Street]"); billingStreet != "" {
		var billingAddress models.Address
		if existingLead.Details.BillingAddressID != nil {
			if err := tx.First(&billingAddress, *existingLead.Details.BillingAddressID); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "Failed to find billing address",
					Code:    "DATABASE_ERROR",
					Details: err.Error(),
				})
				return
			}
		}
		billingAddress.Street = c.PostForm("Details[BillingAddress][Street]")
		billingAddress.UnitSuite = c.PostForm("Details[BillingAddress][UnitSuite]")
		billingAddress.City = c.PostForm("Details[BillingAddress][City]")
		billingAddress.State = c.PostForm("Details[BillingAddress][State]")
		billingAddress.Zip = c.PostForm("Details[BillingAddress][Zip]")
		billingAddress.County = c.PostForm("Details[BillingAddress][County]")

		if err := tx.Save(&billingAddress); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to save billing address",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
			return
		}
		updatedLead.Details.BillingAddressID = &billingAddress.ID
		updatedLead.Details.BillingAddress = &billingAddress
		newBillingFormatted = formatAddress(&billingAddress)
	}

	// Handle Pickup Delivery
	if c.PostForm("Details[IsPickUpDelivery]") == "true" {
		updatedLead.Details.IsPickUpDelivery = true
		if pickupLocationID := c.PostForm("Details[PickupDeliveryAddressID]"); pickupLocationID != "" {
			var pickup models.PickupLocations
			if err := tx.First(&pickup, pickupLocationID); err != nil {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, ErrorResponse{
					Error:   "Invalid pickup location",
					Code:    "VALIDATION_ERROR",
					Details: err.Error(),
				})
				return
			}
			existingLead.Details.PickupDeliveryAddressID = &pickup.ID
			updatedLead.Details.PickupDeliveryAddressID = &pickup.ID
		}
	} else {
		existingLead.Details.IsPickUpDelivery = false
		existingLead.Details.PickupDeliveryAddressID = nil
		updatedLead.Details.PickupDeliveryAddressID = nil
	}

	// var jobid = utils.GenerateEmployeeID(db)

	var residentialEntries []models.ResidentialEntry
	if entriesJSON := c.PostForm("ResidentialEntries"); entriesJSON != "" {
		if err := json.Unmarshal([]byte(entriesJSON), &residentialEntries); err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Failed to parse residential entries",
				Code:    "VALIDATION_ERROR",
				Details: err.Error(),
			})
			return
		}

		// Get all existing residential entry IDs from the database
		var existingEntries []models.ResidentialEntry
		if err := tx.Where("lead_id = ?", updatedLead.ID).Find(&existingEntries); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to fetch existing residential entries",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
			return
		}

		// Create a map of incoming entry IDs for quick lookup
		incomingEntryIDs := make(map[uint]bool)
		for _, entry := range residentialEntries {
			if entry.ID != 0 {
				incomingEntryIDs[entry.ID] = true
			}
		}

		preservedEstimateEntries := make([]models.ResidentialEntry, 0)

		// Delete lead-owned entries that are not in the incoming list. Estimate-owned
		// rows are owned by estimate update flows, so keep them untouched here.
		for _, existingEntry := range existingEntries {
			if !incomingEntryIDs[existingEntry.ID] {
				if residentialEntryHasEstimate(existingEntry) {
					preservedEstimateEntries = append(preservedEstimateEntries, existingEntry)
					continue
				}

				// Delete associated stages first
				if err := tx.Where("residential_entry_id = ?", existingEntry.ID).Delete(&models.Stage{}); err != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, ErrorResponse{
						Error:   "Failed to delete stages for removed residential entry",
						Code:    "DATABASE_ERROR",
						Details: err.Error(),
					})
					return
				}

				// Delete the residential entry
				if err := tx.Delete(&existingEntry); err != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, ErrorResponse{
						Error:   "Failed to delete residential entry",
						Code:    "DATABASE_ERROR",
						Details: err.Error(),
					})
					return
				}
			}
		}
		for i := range residentialEntries {

			residentialEntries[i].LeadID = &updatedLead.ID

			if err := normalizeResidentialEntryLookups(tx, &residentialEntries[i]); err != nil {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, residentialEntryLookupErrorResponse(err))
				return
			}

			var existingEntry models.ResidentialEntry
			err := tx.Where("id = ? AND lead_id = ?", residentialEntries[i].ID, updatedLead.ID).
				First(&existingEntry)

			if err == nil {
				if residentialEntryHasEstimate(existingEntry) {
					residentialEntries[i].EstimateID = existingEntry.EstimateID
				}
				existingEntry.ManagerResidentialEntryID = residentialEntries[i].ManagerResidentialEntryID
				existingEntry.ManagerResidentialTypeID = residentialEntries[i].ManagerResidentialTypeID
				existingEntry.Details = residentialEntries[i].Details
				existingEntry.LeadID = residentialEntries[i].LeadID
				existingEntry.EstimateID = residentialEntries[i].EstimateID

				if err := tx.Save(&existingEntry); err != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, ErrorResponse{
						Error:   "Failed to update residential entry",
						Code:    "DATABASE_ERROR",
						Details: err.Error(),
					})
					return
				}
				residentialEntries[i] = existingEntry
				continue
			}

			if err != gorm.ErrRecordNotFound {
				// Query error
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "Failed to check existing residential entry",
					Code:    "DATABASE_ERROR",
					Details: err.Error(),
				})
				return
			}

			// Not found - Create new entry
			if len(existingLead.Estimates) > 0 {
				residentialEntries[i].EstimateID = &existingLead.Estimates[0].ID
			}
			if err := tx.Create(&residentialEntries[i]); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "Failed to save residential entry",
					Code:    "DATABASE_ERROR",
					Details: err.Error(),
				})
				return
			}
			fmt.Printf("DEBUG: Residential entry %d created successfully\n", i)

			// Create default stages for new entry
			// defaultStages := []models.Stage{
			// 	{
			// 		ID:                 uuid.NewString(),
			// 		LeadID:             updatedLead.ID,
			// 		Name:               "Template",
			// 		State:              "In Progress",
			// 		JobID:              jobid,
			// 		CreatedAt:          time.Now(),
			// 		Date:               func() *time.Time { t := time.Now(); return &t }(),
			// 		Description:        "Prepare and finalize the template for the project",
			// 		ResidentialEntryID: residentialEntries[i].ID,
			// 	},
			// 	{
			// 		ID:                 uuid.NewString(),
			// 		LeadID:             updatedLead.ID,
			// 		JobID:              jobid,
			// 		Name:               "CAD",
			// 		State:              "Not Started",
			// 		CreatedAt:          time.Now(),
			// 		Description:        "Create CAD designs for the project",
			// 		ResidentialEntryID: residentialEntries[i].ID,
			// 	},
			// 	{
			// 		ID:                 uuid.NewString(),
			// 		LeadID:             updatedLead.ID,
			// 		JobID:              jobid,
			// 		Name:               "Cutting",
			// 		State:              "Not Started",
			// 		CreatedAt:          time.Now(),
			// 		Description:        "Cut materials according to specifications",
			// 		ResidentialEntryID: residentialEntries[i].ID,
			// 	},
			// 	{
			// 		ID:                 uuid.NewString(),
			// 		JobID:              jobid,
			// 		LeadID:             updatedLead.ID,
			// 		Name:               "Fabrication",
			// 		State:              "Not Started",
			// 		CreatedAt:          time.Now(),
			// 		Description:        "Fabricate components for the project",
			// 		ResidentialEntryID: residentialEntries[i].ID,
			// 	},
			// 	{
			// 		ID:                 uuid.NewString(),
			// 		LeadID:             updatedLead.ID,
			// 		JobID:              jobid,
			// 		Name:               "Installation",
			// 		State:              "Not Started",
			// 		CreatedAt:          time.Now(),
			// 		Description:        "Install the project components on-site",
			// 		ResidentialEntryID: residentialEntries[i].ID,
			// 	},
			// 	{
			// 		ID:                 uuid.NewString(),
			// 		LeadID:             updatedLead.ID,
			// 		JobID:              jobid,
			// 		Name:               "Service Call",
			// 		State:              "Not Started",
			// 		CreatedAt:          time.Now(),
			// 		Description:        "Handle any service calls or maintenance",
			// 		ResidentialEntryID: residentialEntries[i].ID,
			// 	},
			// }

			// fmt.Println("Creating default stages for new residential entry:", residentialEntries[i].ID)
			// for _, stage := range defaultStages {
			// 	if err := tx.Create(&stage); err != nil {
			// 		tx.Rollback()
			// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create default stages: " + err.Error()})
			// 		return
			// 	}
			// 	fmt.Println("Stage created:", stage.Name)
				// }
			}

			residentialEntries = append(residentialEntries, preservedEstimateEntries...)
			updatedLead.ResidentialEntries = residentialEntries
		} else {

			// If ResidentialEntries is empty or not provided, delete lead-owned entries.
			// Estimate-owned rows stay in place and remain managed by estimate handlers.
			var existingEntries []models.ResidentialEntry
			preservedEstimateEntries := make([]models.ResidentialEntry, 0)
			if err := tx.Where("lead_id = ?", updatedLead.ID).Find(&existingEntries); err == nil {
				for _, existingEntry := range existingEntries {
					if residentialEntryHasEstimate(existingEntry) {
						preservedEstimateEntries = append(preservedEstimateEntries, existingEntry)
						continue
					}

					// Delete associated stages first
					if err := tx.Where("residential_entry_id = ?", existingEntry.ID).Delete(&models.Stage{}); err != nil {
						tx.Rollback()
					c.JSON(http.StatusInternalServerError, ErrorResponse{
						Error:   "Failed to delete stages",
						Code:    "DATABASE_ERROR",
						Details: err.Error(),
					})
					return
				}

				// Delete the residential entry
				if err := tx.Delete(&existingEntry); err != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, ErrorResponse{
						Error:   "Failed to delete residential entry",
						Code:    "DATABASE_ERROR",
						Details: err.Error(),
					})
					return
					}
				}
			}
			updatedLead.ResidentialEntries = preservedEstimateEntries
		}

		if err := syncManagerResidentialLengths(
			tx,
			oldLeadManagedResidentialEntries,
			leadManagedResidentialEntries(updatedLead.ResidentialEntries),
		); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to update residential entry usage",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
			return
		}

		// Handle Associates
	var validAssociates []*models.AssociateCustomer
	var associates []models.AssociateCustomer
	if assocJSON := c.PostForm("associates"); assocJSON != "" {
		if err := json.Unmarshal([]byte(assocJSON), &associates); err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Failed to parse associates",
				Code:    "VALIDATION_ERROR",
				Details: err.Error(),
			})
			return
		}
		if err := tx.Exec("DELETE FROM lead_associates WHERE lead_id = ?", updatedLead.ID); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to update associates",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
			return
		}

		if len(associates) > 0 {

			for _, associate := range associates {
				if associate.Details.FirstName == "" && associate.Details.LastName == "" && associate.Details.CompanyName == "" {
					continue
				}

				associate.Details.Email = strings.TrimSpace(associate.Details.Email)
				associate.Details.Phone = strings.TrimSpace(associate.Details.Phone)

				var existingAssociate models.AssociateCustomer
				if associate.Details.Email != "" || associate.Details.Phone != "" {
					query := tx.Model(&models.AssociateCustomer{})
					conditions := []string{}
					args := []interface{}{}
					if associate.Details.Email != "" {
						conditions = append(conditions, "details->>'email' = ?")
						args = append(args, associate.Details.Email)
					}
					if associate.Details.Phone != "" {
						conditions = append(conditions, "details->>'phone' = ?")
						args = append(args, associate.Details.Phone)
					}
					if len(conditions) > 0 {
						if err := query.Where(strings.Join(conditions, " OR "), args...).First(&existingAssociate); err == nil {
							validAssociates = append(validAssociates, &existingAssociate)
							continue
						}
					}
				}

				var company models.Customer
				if associate.Details.CompanyName != "" {
					if err := tx.Where("name = ?", associate.Details.CompanyName).First(&company); err != nil {
						if errors.Is(err, gorm.ErrRecordNotFound) {
							company = models.Customer{
								ID:                          utils.GenerateEmployeeID(db),
								Name:                        associate.Details.CompanyName,
								Phone:                       associate.Details.CompanyPhone,
								Email:                       associate.Details.CompanyEmail,
								SecondaryEmail:              associate.Details.CompanySecondaryEmail,
								Address:                     associate.Details.CompanyAddress,
								BillingAddress:              associate.Details.CompanyBillingAddress,
								Details:                     associate.Details.CompanyDetails,
								ContactName:                 associate.Details.CompanyContactName,
								SalesPerson1:                associate.Details.CompanySalesPerson1,
								SalesPerson2:                associate.Details.CompanySalesPerson2,
								ProjectManager:              associate.Details.CompanyProjectManager,
								CustomerSince:               associate.Details.CompanyCustomerSince,
								SalesTax:                    associate.Details.CompanySalesTax,
								EINNumber:                   associate.Details.CompanyEINNumber,
								TaxExempt:                   associate.Details.CompanyTaxExempt,
								TaxExemptReason:             associate.Details.CompanyTaxExemptReason,
								ExemptCertificateNumber:     associate.Details.CompanyExemptCertificateNumber,
								ExemptExpiryDate:            associate.Details.CompanyExemptExpiryDate,
								SpecialDeliveryInstructions: associate.Details.CompanySpecialDeliveryInstructions,
								CollectionNotes:             associate.Details.CompanyCollectionNotes,
							}
							if err := tx.Create(&company); err != nil {
								tx.Rollback()
								c.JSON(http.StatusInternalServerError, ErrorResponse{
									Error:   "Failed to create company",
									Code:    "DATABASE_ERROR",
									Details: err.Error(),
								})
								return
							}
						} else {
							tx.Rollback()
							c.JSON(http.StatusInternalServerError, ErrorResponse{
								Error:   "Database error",
								Code:    "DATABASE_ERROR",
								Details: err.Error(),
							})
							return
						}
					}
				}

				newAssociate := models.AssociateCustomer{
					ID:   utils.GenerateEmployeeID(db),
					Role: models.AssociateRole(associate.Role),
					Details: models.AssociateDetails{
						FirstName:                          associate.Details.FirstName,
						LastName:                           associate.Details.LastName,
						Email:                              associate.Details.Email,
						Phone:                              associate.Details.Phone,
						CompanyName:                        associate.Details.CompanyName,
						CompanyPhone:                       associate.Details.CompanyPhone,
						CompanyEmail:                       associate.Details.CompanyEmail,
						CompanySecondaryEmail:              associate.Details.CompanySecondaryEmail,
						CompanyContactName:                 associate.Details.CompanyContactName,
						CompanyAddress:                     associate.Details.CompanyAddress,
						CompanyBillingAddress:              associate.Details.CompanyBillingAddress,
						CompanyDetails:                     associate.Details.CompanyDetails,
						CompanySalesPerson1:                associate.Details.CompanySalesPerson1,
						CompanySalesPerson2:                associate.Details.CompanySalesPerson2,
						CompanyProjectManager:              associate.Details.CompanyProjectManager,
						CompanyCustomerSince:               associate.Details.CompanyCustomerSince,
						CompanySalesTax:                    associate.Details.CompanySalesTax,
						CompanyEINNumber:                   associate.Details.CompanyEINNumber,
						CompanyTaxExempt:                   associate.Details.CompanyTaxExempt,
						CompanyTaxExemptReason:             associate.Details.CompanyTaxExemptReason,
						CompanyExemptCertificateNumber:     associate.Details.CompanyExemptCertificateNumber,
						CompanyExemptExpiryDate:            associate.Details.CompanyExemptExpiryDate,
						CompanySpecialDeliveryInstructions: associate.Details.CompanySpecialDeliveryInstructions,
						CompanyCollectionNotes:             associate.Details.CompanyCollectionNotes,
						CompanyID:                          company.ID,
						Company:                            company,
					},
				}
				if err := tx.Create(&newAssociate); err != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, ErrorResponse{
						Error:   "Failed to save associate",
						Code:    "DATABASE_ERROR",
						Details: err.Error(),
					})
					return
				}
				validAssociates = append(validAssociates, &newAssociate)
			}

			// Link Associates
			for _, associate := range validAssociates {
				if err := tx.Exec("INSERT INTO lead_associates (lead_id, associate_customer_id) VALUES (?, ?)", updatedLead.ID, associate.ID); err != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, ErrorResponse{
						Error:   "Failed to link associate",
						Code:    "DATABASE_ERROR",
						Details: err.Error(),
					})
					return
				}
			}
		}
		updatedLead.Associates = validAssociates
	} else {
		updatedLead.Associates = existingLead.Associates
	}

	// Handle File Uploads
	if form, err := c.MultipartForm(); err == nil {
		bucket := h.getBucketName(c)

		// New file uploads — upload to S3, save real FilePath
		for i, fileHeader := range form.File["Files"] {
			file, err := fileHeader.Open()
			if err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "Failed to open file",
					Code:    "FILE_ERROR",
					Details: err.Error(),
				})
				return
			}
			defer file.Close()

			categoryKey := fmt.Sprintf("Files[%d][Category]", i)
			projectTypeKey := fmt.Sprintf("Files[%d][ProjectType]", i)
			category := "Other"
			projecttype := ""
			if val, ok := form.Value[categoryKey]; ok && len(val) > 0 {
				category = val[0]
			}
			if val, ok := form.Value[projectTypeKey]; ok && len(val) > 0 {
				projecttype = val[0]
			}

			fileExt := filepath.Ext(fileHeader.Filename)
			uniqueID := uuid.New().String()
			s3Key := fmt.Sprintf("files/%s%s", uniqueID, fileExt)
			storedFileName := fmt.Sprintf("%s%s", uniqueID, fileExt)

			_, err = h.s3Client.PutObject(&s3.PutObjectInput{
				Bucket:             aws.String(bucket),
				Key:                aws.String(s3Key),
				Body:               file,
				ContentDisposition: aws.String("attachment"),
				ContentType:        aws.String(fileHeader.Header.Get("Content-Type")),
			})
			if err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "Failed to upload file to S3",
					Code:    "FILE_ERROR",
					Details: err.Error(),
				})
				return
			}

			// filePath := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucket, h.region, s3Key)

			newFile := models.File{
				ID:             uint(rand.Uint32()),
				LeadID:         &updatedLead.ID,
				Name:           fileHeader.Filename,
				Type:           fileHeader.Header.Get("Content-Type"),
				Size:           fmt.Sprintf("%d", fileHeader.Size),
				UploadedAt:     time.Now(),
				Category:       category,
				ProjectType:    projecttype,
				StoredFileName: storedFileName,
				FilePath:       s3Key,
			}
			if err := tx.Create(&newFile); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "Failed to save file to database",
					Code:    "DATABASE_ERROR",
					Details: err.Error(),
				})
				return
			}
		}

		// Preserve existing files — update their metadata only
		if existingFilesJSON, ok := form.Value["ExistingFiles"]; ok && len(existingFilesJSON) > 0 {
			var existingFileMeta []struct {
				ID          uint   `json:"ID"`
				Name        string `json:"Name"`
				Category    string `json:"Category"`
				ProjectType string `json:"ProjectType"`
			}
			if err := json.Unmarshal([]byte(existingFilesJSON[0]), &existingFileMeta); err == nil {
				for _, ef := range existingFileMeta {
					if ef.ID == 0 {
						continue
					}
					if updateErr := tx.Model(&models.File{}).
						Where("id = ? AND lead_id = ?", ef.ID, updatedLead.ID).
						Updates(map[string]interface{}{
							"name":         ef.Name,
							"category":     ef.Category,
							"project_type": ef.ProjectType,
						}); updateErr != nil {
						fmt.Printf("WARN: failed to update existing file %d: %v\n", ef.ID, updateErr)
					}
				}
			}
		}

		// Profile picture upload
		for _, fileHeader := range form.File["Profile"] {
			file, err := fileHeader.Open()
			if err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "Failed to open profile file",
					Code:    "FILE_ERROR",
					Details: err.Error(),
				})
				return
			}
			defer file.Close()

			fileExt := filepath.Ext(fileHeader.Filename)
			s3Key := fmt.Sprintf("files/%s%s", uuid.New().String(), fileExt)

			_, err = h.s3Client.PutObject(&s3.PutObjectInput{
				Bucket:      aws.String(bucket),
				Key:         aws.String(s3Key),
				Body:        file,
				ContentType: aws.String(fileHeader.Header.Get("Content-Type")),
			})
			if err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "Failed to upload profile picture",
					Code:    "FILE_ERROR",
					Details: err.Error(),
				})
				return
			}
			existingLead.ProfilePic = s3Key
		}
		if updatedLead.ProfilePic == "" {
			updatedLead.ProfilePic = existingLead.ProfilePic
		}
	}

	// Copy all updated values onto existingLead for saving
	existingLead.JobName = updatedLead.JobName
	existingLead.Status = updatedLead.Status
	existingLead.DetailedLeadStatus = updatedLead.DetailedLeadStatus
	existingLead.DeliveryType = updatedLead.DeliveryType
	existingLead.LeadType = updatedLead.LeadType
	existingLead.PaymentTerm = updatedLead.PaymentTerm
	existingLead.PricingType = updatedLead.PricingType
	existingLead.Sales1PersonID = updatedLead.Sales1PersonID
	existingLead.Sales1PersonName = updatedLead.Sales1PersonName
	existingLead.Sales2PersonID = updatedLead.Sales2PersonID
	existingLead.Sales2PersonName = updatedLead.Sales2PersonName
	existingLead.Details.Phone = updatedLead.Details.Phone
	existingLead.Details.Secondaryphone = updatedLead.Details.Secondaryphone
	existingLead.Details.Cellphone = updatedLead.Details.Cellphone
	existingLead.Details.ShipToEmail = updatedLead.Details.ShipToEmail
	existingLead.Details.SecondaryEmail = updatedLead.Details.SecondaryEmail
	existingLead.Details.BillToCustomer = updatedLead.Details.BillToCustomer
	existingLead.Details.FaxNumber = updatedLead.Details.FaxNumber
	existingLead.Details.MobileNumber = updatedLead.Details.MobileNumber
	existingLead.Details.NotesHistory = updatedLead.Details.NotesHistory
	existingLead.Details.ShippingAddressID = updatedLead.Details.ShippingAddressID
	existingLead.Details.ShippingAddress = updatedLead.Details.ShippingAddress
	existingLead.Details.BillingAddressID = updatedLead.Details.BillingAddressID
	existingLead.Details.BillingAddress = updatedLead.Details.BillingAddress
	existingLead.Details.IsPickUpDelivery = updatedLead.Details.IsPickUpDelivery
	existingLead.ResidentialEntries = updatedLead.ResidentialEntries
	existingLead.Details.CustomerPhone = updatedLead.Details.CustomerPhone
	existingLead.Details.CustomerEmail = updatedLead.Details.CustomerEmail
	existingLead.Associates = updatedLead.Associates

	userID, userName := middleware.GetRequestUser(c)

	creatorLabel := userName
	if creatorLabel == "" || creatorLabel == "System" {
		creatorLabel = userID
	}
	if (userID == "system" || userID == "") && existingLead.Sales1PersonID != "" {
		userID = existingLead.Sales1PersonID
		userName = existingLead.Sales1PersonName
		creatorLabel = userName
	}

	leadChanges := getLeadChanges(oldLead, existingLead)

	if oldShippingFormatted != newShippingFormatted {
		leadChanges["shipping_address"] = map[string]string{
			"old_value": oldShippingFormatted,
			"new_value": newShippingFormatted,
		}
	}
	if oldBillingFormatted != newBillingFormatted {
		leadChanges["billing_address"] = map[string]string{
			"old_value": oldBillingFormatted,
			"new_value": newBillingFormatted,
		}
	}

	event := activity.ActivityEvent{
		EntityType: "lead",
		EntityID:   existingLead.ID,
		Action:     "updated",
		UserID:     userID,
		UserName:   userName,
		Payload: map[string]any{
			"lead_name":      existingLead.JobName,
			"updated_by":     creatorLabel,
			"changed_fields": leadChanges,
		},
	}
	if err := activity.InsertActivityEvent(c.Request.Context(), tx.GetUnderlying(), event); err != nil {
		log.Printf("activity event publish failed: %v", err)
	}

	// Save updated lead
	if err := tx.Save(&existingLead); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to update lead",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	// Commit transaction
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
		"message": "Lead updated successfully",
		"id":      existingLead.ID,
	})
}

func getLeadChanges(existingLead, updatedLead models.Lead) map[string]map[string]string {
	changes := make(map[string]map[string]string)

	if existingLead.JobName != updatedLead.JobName {
		changes["job_name"] = map[string]string{"old_value": existingLead.JobName, "new_value": updatedLead.JobName}
	}
	if existingLead.Status != updatedLead.Status {
		changes["status"] = map[string]string{"old_value": string(existingLead.Status), "new_value": string(updatedLead.Status)}
	}
	if existingLead.Sales1PersonName != updatedLead.Sales1PersonName {
		changes["sales1_person"] = map[string]string{"old_value": existingLead.Sales1PersonName, "new_value": updatedLead.Sales1PersonName}
	}
	if existingLead.Sales2PersonName != updatedLead.Sales2PersonName {
		changes["sales2_person"] = map[string]string{"old_value": existingLead.Sales2PersonName, "new_value": updatedLead.Sales2PersonName}
	}
	if existingLead.DeliveryType != updatedLead.DeliveryType {
		changes["delivery_type"] = map[string]string{"old_value": string(existingLead.DeliveryType), "new_value": string(updatedLead.DeliveryType)}
	}
	if existingLead.LeadType != updatedLead.LeadType {
		changes["lead_type"] = map[string]string{"old_value": string(existingLead.LeadType), "new_value": string(updatedLead.LeadType)}
	}
	if existingLead.PaymentTerm != updatedLead.PaymentTerm {
		changes["payment_term"] = map[string]string{"old_value": string(existingLead.PaymentTerm), "new_value": string(updatedLead.PaymentTerm)}
	}
	if existingLead.PricingType != updatedLead.PricingType {
		changes["pricing_type"] = map[string]string{"old_value": string(existingLead.PricingType), "new_value": string(updatedLead.PricingType)}
	}
	if existingLead.ProfilePic != updatedLead.ProfilePic {
		changes["profile_pic"] = map[string]string{"old_value": existingLead.ProfilePic, "new_value": updatedLead.ProfilePic}
	}
	if existingLead.Details.Phone != updatedLead.Details.Phone {
		changes["phone"] = map[string]string{"old_value": existingLead.Details.Phone, "new_value": updatedLead.Details.Phone}
	}
	if existingLead.Details.ShipToEmail != updatedLead.Details.ShipToEmail {
		changes["ship_to_email"] = map[string]string{"old_value": existingLead.Details.ShipToEmail, "new_value": updatedLead.Details.ShipToEmail}
	}
	if existingLead.Details.SecondaryEmail != updatedLead.Details.SecondaryEmail {
		changes["secondary_email"] = map[string]string{"old_value": existingLead.Details.SecondaryEmail, "new_value": updatedLead.Details.SecondaryEmail}
	}
	if existingLead.Details.BillToCustomer != updatedLead.Details.BillToCustomer {
		changes["bill_to_customer"] = map[string]string{"old_value": existingLead.Details.BillToCustomer, "new_value": updatedLead.Details.BillToCustomer}
	}
	if existingLead.Details.FaxNumber != updatedLead.Details.FaxNumber {
		changes["fax_number"] = map[string]string{"old_value": existingLead.Details.FaxNumber, "new_value": updatedLead.Details.FaxNumber}
	}
	if existingLead.Details.MobileNumber != updatedLead.Details.MobileNumber {
		changes["mobile_number"] = map[string]string{"old_value": existingLead.Details.MobileNumber, "new_value": updatedLead.Details.MobileNumber}
	}

	// CountyTax
	oldTaxName := ""
	if existingLead.Details.CountyTax != nil {
		oldTaxName = existingLead.Details.CountyTax.Name
	}
	newTaxName := ""
	if updatedLead.Details.CountyTax != nil {
		newTaxName = updatedLead.Details.CountyTax.Name
	}
	if oldTaxName != newTaxName {
		changes["county_tax"] = map[string]string{"old_value": oldTaxName, "new_value": newTaxName}
	}

	// Discount
	// oldDiscountName := ""
	// if existingLead.Details.Discount != nil {
	// 	oldDiscountName = existingLead.Details.Discount.Name
	// }
	// newDiscountName := ""
	// if updatedLead.Details.Discount != nil {
	// 	newDiscountName = updatedLead.Details.Discount.Name
	// }
	// if oldDiscountName != newDiscountName {
	// 	changes["discount"] = map[string]string{"old_value": oldDiscountName, "new_value": newDiscountName}
	// }

	// IsPickUpDelivery
	if existingLead.Details.IsPickUpDelivery != updatedLead.Details.IsPickUpDelivery {
		changes["is_pickup_delivery"] = map[string]string{
			"old_value": fmt.Sprintf("%t", existingLead.Details.IsPickUpDelivery),
			"new_value": fmt.Sprintf("%t", updatedLead.Details.IsPickUpDelivery),
		}
	}

	// Associates
	oldAssocData := make(map[string]*models.AssociateCustomer)
	for _, a := range existingLead.Associates {
		oldAssocData[a.ID] = a
	}
	newAssocData := make(map[string]*models.AssociateCustomer)
	for _, a := range updatedLead.Associates {
		newAssocData[a.ID] = a
	}

	processedNew := make(map[string]bool)
	processedOld := make(map[string]bool)

	var totalOld []string
	var totalNew []string

	for nid, na := range newAssocData {
		name := strings.TrimSpace(fmt.Sprintf("%s %s", na.Details.FirstName, na.Details.LastName))

		if oa, ok := oldAssocData[nid]; ok {
			oldVal, newVal := buildFieldChanges(oa, na)
			if oldVal != "" || newVal != "" {
				totalOld = append(totalOld, oldVal)
				totalNew = append(totalNew, newVal)
			}
			processedNew[nid] = true
			processedOld[nid] = true
			continue
		}

		if name != "" {
			for oid, oa := range oldAssocData {
				if processedOld[oid] {
					continue
				}
				oldName := strings.TrimSpace(fmt.Sprintf("%s %s", oa.Details.FirstName, oa.Details.LastName))
				if strings.EqualFold(oldName, name) {
					oldVal, newVal := buildFieldChanges(oa, na)
					if oldVal != "" || newVal != "" {
						totalOld = append(totalOld, oldVal)
						totalNew = append(totalNew, newVal)
					}
					processedNew[nid] = true
					processedOld[oid] = true
					break
				}
			}
		}
	}

	for nid, na := range newAssocData {
		if !processedNew[nid] {
			totalNew = append(totalNew, buildAssociateDetail(na))
		}
	}

	for oid, oa := range oldAssocData {
		if !processedOld[oid] {
			totalOld = append(totalOld, buildAssociateDetail(oa))
		}
	}

	if len(totalOld) > 0 || len(totalNew) > 0 {
		changes["associates_update"] = map[string]string{
			"old_value": strings.Join(totalOld, "; "),
			"new_value": strings.Join(totalNew, "; "),
		}
	}

	// ResidentialEntries
	oldMap := make(map[uint]string)
	for _, e := range existingLead.ResidentialEntries {
		oldMap[e.ID] = getResidentialEntryLabel(e)
	}
	newMap := make(map[uint]string)
	for _, e := range updatedLead.ResidentialEntries {
		if e.ID != 0 {
			newMap[e.ID] = getResidentialEntryLabel(e)
		}
	}

	var added []string
	for _, e := range updatedLead.ResidentialEntries {
		if e.ID == 0 {
			added = append(added, getResidentialEntryLabel(e))
		} else if _, exists := oldMap[e.ID]; !exists {
			added = append(added, getResidentialEntryLabel(e))
		}
	}
	var removed []string
	for id, label := range oldMap {
		if _, exists := newMap[id]; !exists {
			removed = append(removed, label)
		}
	}
	if len(added) > 0 {
		changes["residential_projects_added"] = map[string]string{"old_value": "", "new_value": strings.Join(added, ", ")}
	}
	if len(removed) > 0 {
		changes["residential_projects_removed"] = map[string]string{"old_value": strings.Join(removed, ", "), "new_value": ""}
	}

	return changes
}

func buildFieldChanges(oa, na *models.AssociateCustomer) (string, string) {
	if oa == nil || na == nil {
		return "", ""
	}
	oldParts := []string{}
	newParts := []string{}

	compareAndAppend := func(label string, oldVal, newVal string) {
		oldVal = strings.TrimSpace(oldVal)
		newVal = strings.TrimSpace(newVal)
		if oldVal != newVal {
			if oldVal != "" {
				oldParts = append(oldParts, label+":"+oldVal)
			}
			if newVal != "" {
				newParts = append(newParts, label+":"+newVal)
			}
		}
	}

	compareAndAppend("first", oa.Details.FirstName, na.Details.FirstName)
	compareAndAppend("last", oa.Details.LastName, na.Details.LastName)
	compareAndAppend("email", oa.Details.Email, na.Details.Email)
	compareAndAppend("phone", oa.Details.Phone, na.Details.Phone)
	compareAndAppend("company", oa.Details.CompanyName, na.Details.CompanyName)
	compareAndAppend("role", string(oa.Role), string(na.Role))

	return strings.Join(oldParts, ", "), strings.Join(newParts, ", ")
}

func getResidentialEntryLabel(entry models.ResidentialEntry) string {
	namePart := ""
	if entry.ManagerResidentialEntry.ID != 0 {
		namePart = entry.ManagerResidentialEntry.ResidentialEntryName
	}

	if namePart != "" && entry.Details != "" {
		return fmt.Sprintf("%s-%s", namePart, entry.Details)
	}
	if namePart != "" {
		return namePart
	}
	if entry.Details != "" {
		return entry.Details
	}
	return "Project"
}

func buildAssociateDetail(a *models.AssociateCustomer) string {
	if a == nil {
		return ""
	}
	var parts []string
	if v := strings.TrimSpace(a.Details.FirstName); v != "" {
		parts = append(parts, "first:"+v)
	}
	if v := strings.TrimSpace(a.Details.LastName); v != "" {
		parts = append(parts, "last:"+v)
	}
	if v := strings.TrimSpace(a.Details.CompanyName); v != "" {
		parts = append(parts, "company:"+v)
	}
	if v := strings.TrimSpace(a.Details.Email); v != "" {
		parts = append(parts, "email:"+v)
	}
	if v := strings.TrimSpace(a.Details.Phone); v != "" {
		parts = append(parts, "phone:"+v)
	}
	if v := strings.TrimSpace(string(a.Role)); v != "" {
		parts = append(parts, "role:"+v)
	}
	return strings.Join(parts, ", ")
}

func formatAddress(a *models.Address) string {
	if a == nil {
		return ""
	}
	parts := []string{}
	if a.Street != "" {
		parts = append(parts, a.Street)
	}
	if a.UnitSuite != "" {
		parts = append(parts, a.UnitSuite)
	}
	if a.City != "" {
		parts = append(parts, a.City)
	}
	if a.State != "" {
		parts = append(parts, a.State)
	}
	if a.Zip != "" {
		parts = append(parts, a.Zip)
	}
	if a.County != "" {
		parts = append(parts, a.County)
	}
	if len(parts) == 0 {
		return fmt.Sprintf("Address #%d", a.ID)
	}
	return strings.Join(parts, ", ")
}

func (h *LeadHandler) UpdateLeadSchedule(c *gin.Context) {
	db := middleware.GetDB(c)
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid lead ID",
			Code:  "BAD_REQUEST",
		})
		return
	}

	var payload struct {
		ScheduleType  string `json:"scheduleType"`
		ScheduledDate string `json:"scheduledDate"`
	}
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Code:    "VALIDATION_ERROR",
			Details: err.Error(),
		})
		return
	}

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Unexpected error occurred",
				Code:    "INTERNAL_SERVER_ERROR",
				Details: fmt.Sprintf("%v", r),
			})
		}
	}()

	scheduleType := models.ScheduleType(payload.ScheduleType)
	if scheduleType != models.ScheduleToday && scheduleType != models.ScheduleTomorrow && scheduleType != models.ScheduleCustom {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid schedule type",
			Code:    "VALIDATION_ERROR",
			Details: "scheduleType must be Today, Tomorrow or Custom",
		})
		return
	}

	var scheduleDate *time.Time
	if payload.ScheduledDate != "" {
		t, err := time.Parse("2006-01-02", payload.ScheduledDate)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Invalid scheduled date format",
				Code:    "VALIDATION_ERROR",
				Details: err.Error(),
			})
			return
		}
		scheduleDate = &t
	} else if scheduleType == models.ScheduleToday {
		t := time.Now().UTC()
		t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
		scheduleDate = &t
	} else if scheduleType == models.ScheduleTomorrow {
		t := time.Now().UTC().AddDate(0, 0, 1)
		t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
		scheduleDate = &t
	} else {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Custom date required",
			Code:    "VALIDATION_ERROR",
			Details: "scheduledDate is required for Custom schedule type",
		})
		return
	}

	var existingLead models.Lead
	if err := tx.First(&existingLead, "id = ?", id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Lead not found",
				Code:    "NOT_FOUND",
				Details: "No lead found with the specified ID",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Database error",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	var leadSchedule models.LeadSchedule
	if err := tx.Where("lead_id = ?", id).First(&leadSchedule); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			leadSchedule = models.LeadSchedule{
				ID:            uuid.NewString(),
				LeadID:        id,
				ScheduleType:  scheduleType,
				ScheduledDate: scheduleDate,
			}
			if err := tx.Create(&leadSchedule); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "Failed to save lead schedule",
					Code:    "DATABASE_ERROR",
					Details: err.Error(),
				})
				return
			}
		} else {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Database error",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
			return
		}
	} else {
		leadSchedule.ScheduleType = scheduleType
		leadSchedule.ScheduledDate = scheduleDate
		if err := tx.Save(&leadSchedule); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to update lead schedule",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
			return
		}
	}

	userID, userName := middleware.GetRequestUser(c)
	event := activity.ActivityEvent{
		EntityType: "lead",
		EntityID:   id,
		Action:     "schedule_updated",
		UserID:     userID,
		UserName:   userName,
		Payload: map[string]any{
			"new_date":   scheduleDate,
			"type":       scheduleType,
			"updated_by": userName,
		},
	}
	if err := activity.InsertActivityEvent(c.Request.Context(), tx.GetUnderlying(), event); err != nil {
		log.Printf("activity event publish failed: %v", err)
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
		"message": "Lead schedule saved successfully",
		"data":    leadSchedule,
	})
}
func (h *LeadHandler) GetAllLeads(c *gin.Context) {
	db := middleware.GetDB(c)
	companyID := strings.TrimSpace(c.Query("company_id"))
	if companyID == "" {
		companyID = strings.TrimSpace(c.Query("companyId"))
	}

	var totalCount int64
	if err := db.Model(&models.Lead{}).Count(&totalCount).Error(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to count leads",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}
	var leads []models.Lead
	query := db.Model(&models.Lead{}).
		Preload("Associates").
		Preload("ResidentialEntries.ManagerResidentialEntry").
		Preload("ResidentialEntries.ManagerResidentialType").
		Preload("Estimates.MaterialRows").
		Preload("Estimates.ServiceRows").
		Preload("Files").
		Preload("Schedule").
		Order("created_at desc")

	if companyID != "" {
		query = query.
			Joins("JOIN lead_associates ON lead_associates.lead_id = leads.id").
			Joins("JOIN associate_customers ON associate_customers.id = lead_associates.associate_customer_id").
			Where("associate_customers.details->>'company_id' = ?", companyID).
			Distinct()
	}

	if err := query.Find(&leads); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to retrieve leads",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}
	for i := range leads {
		if leads[i].Details.ShippingAddressID != nil {
			var shippingAddress models.Address
			if err := db.First(&shippingAddress, *leads[i].Details.ShippingAddressID); err != nil {
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "Failed to retrieve agents",
					Code:    "DATABASE_ERROR",
					Details: err.Error(),
				})
			} else {
				leads[i].Details.ShippingAddress = &shippingAddress
			}
		}

		if leads[i].Details.BillingAddressID != nil {
			var billingAddress models.Address
			if err := db.First(&billingAddress, *leads[i].Details.BillingAddressID); err != nil {

			} else {
				leads[i].Details.BillingAddress = &billingAddress
			}
		}
		if leads[i].Sales1PersonID != "" {

			var Sales1Person models.Agent
			if err := db.First(&Sales1Person, &leads[i].Sales1PersonID); err != nil {

			} else {
				leads[i].Sales1Person = &Sales1Person

			}
		}
		if leads[i].Sales2PersonID != "" {
			var Sales2Person models.Agent
			if err := db.First(&Sales2Person, &leads[i].Sales2PersonID); err != nil {
			} else {
				leads[i].Sales2Person = &Sales2Person
			}
		}
		if leads[i].Details.PickupDeliveryAddressID != nil {
			var pickupAddress models.PickupLocations
			if err := db.First(&pickupAddress, *leads[i].Details.PickupDeliveryAddressID); err != nil {
			} else {
				leads[i].Details.PickupDeliveryAddress = &pickupAddress
			}
		}

	}
	h.batchLoadRelatedData(&leads, c)
	for i := range leads {
		h.enrichFilesWithPresignedURLs(c, leads[i].Files)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Leads retrieved successfully",
		"data":    leads,
	})
}
func (h *LeadHandler) batchLoadRelatedData(leads *[]models.Lead, c *gin.Context) {
	db := middleware.GetDB(c)

	if len(*leads) == 0 {
		return
	}

	// Collect all IDs for batch queries
	var shippingAddressIDs, billingAddressIDs, pickupAddressIDs []interface{}
	var sales1PersonIDs, sales2PersonIDs []interface{}

	leadIndexMap := make(map[string]int) // Map ID to index for faster lookup

	for i, lead := range *leads {
		leadIndexMap[lead.ID] = i

		if lead.Details.ShippingAddressID != nil {
			shippingAddressIDs = append(shippingAddressIDs, *lead.Details.ShippingAddressID)
		}
		if lead.Details.BillingAddressID != nil {
			billingAddressIDs = append(billingAddressIDs, *lead.Details.BillingAddressID)
		}
		if lead.Details.PickupDeliveryAddressID != nil {
			pickupAddressIDs = append(pickupAddressIDs, *lead.Details.PickupDeliveryAddressID)
		}
		if lead.Sales1PersonID != "" {
			sales1PersonIDs = append(sales1PersonIDs, lead.Sales1PersonID)
		}
		if lead.Sales2PersonID != "" {
			sales2PersonIDs = append(sales2PersonIDs, lead.Sales2PersonID)
		}
	}

	// Use sync.WaitGroup for concurrent batch queries
	var wg sync.WaitGroup

	// Batch load addresses
	wg.Add(3)

	var shippingAddresses, billingAddresses []models.Address
	var pickupAddresses []models.PickupLocations

	go func() {
		defer wg.Done()
		if len(shippingAddressIDs) > 0 {
			db.Where("id IN ?", shippingAddressIDs).Find(&shippingAddresses)
		}
	}()

	go func() {
		defer wg.Done()
		if len(billingAddressIDs) > 0 {
			db.Where("id IN ?", billingAddressIDs).Find(&billingAddresses)
		}
	}()

	go func() {
		defer wg.Done()
		if len(pickupAddressIDs) > 0 {
			db.Where("id IN ?", pickupAddressIDs).Find(&pickupAddresses)
		}
	}()

	// Batch load agents
	wg.Add(2)
	var sales1Persons, sales2Persons []models.Agent

	go func() {
		defer wg.Done()
		if len(sales1PersonIDs) > 0 {
			db.Where("id IN ?", sales1PersonIDs).Find(&sales1Persons)
		}
	}()

	go func() {
		defer wg.Done()
		if len(sales2PersonIDs) > 0 {
			db.Where("id IN ?", sales2PersonIDs).Find(&sales2Persons)
		}
	}()

	wg.Wait()

	// Create maps for O(1) lookup
	// shippingMap := make(map[string]models.Address)
	// billingMap := make(map[string]models.Address)
	// pickupMap := make(map[string]models.PickupLocations)
	sales1Map := make(map[string]models.Agent)
	sales2Map := make(map[string]models.Agent)

	// for _, addr := range shippingAddresses {
	// 	shippingMap[addr.ID] = addr
	// }
	// for _, addr := range billingAddresses {
	// 	billingMap[addr.ID] = addr
	// }
	// for _, addr := range pickupAddresses {
	// 	pickupMap[addr.ID] = addr
	// }
	for _, agent := range sales1Persons {
		sales1Map[agent.ID] = agent
	}
	for _, agent := range sales2Persons {
		sales2Map[agent.ID] = agent
	}

	// Assign loaded data back to leads
	for i := range *leads {
		lead := &(*leads)[i]

		// if lead.Details.ShippingAddressID != nil {
		// 	if addr, exists := shippingMap[*lead.Details.ShippingAddressID]; exists {
		// 		lead.Details.ShippingAddress = &addr
		// 	}
		// }

		// if lead.Details.BillingAddressID != nil {
		// 	if addr, exists := billingMap[*lead.Details.BillingAddressID]; exists {
		// 		lead.Details.BillingAddress = &addr
		// 	}
		// }

		// if lead.Details.PickupDeliveryAddressID != nil {
		// 	if addr, exists := pickupMap[*lead.Details.PickupDeliveryAddressID]; exists {
		// 		lead.Details.PickupDeliveryAddress = &addr
		// 	}
		// }

		if lead.Sales1PersonID != "" {
			if agent, exists := sales1Map[lead.Sales1PersonID]; exists {
				lead.Sales1Person = &agent
			}
		}

		if lead.Sales2PersonID != "" {
			if agent, exists := sales2Map[lead.Sales2PersonID]; exists {
				lead.Sales2Person = &agent
			}
		}
	}
}

// DeleteLead deletes a lead by ID
func (h *LeadHandler) DeleteLead(c *gin.Context) {
	db := middleware.GetDB(c)

	// Extract and validate ID from URL parameter
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid lead ID",
			Code:  "BAD_REQUEST",
		})
		return
	}

	// Begin transaction
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Unexpected error occurred",
				Code:    "INTERNAL_SERVER_ERROR",
				Details: fmt.Sprintf("%v", r),
			})
		}
	}()

	// Check if lead exists
	var existingLead models.Lead
	if err := tx.Preload("Files").First(&existingLead, "id = ?", id); err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Lead not found",
			Code:    "NOT_FOUND",
			Details: "No lead found with the specified ID",
		})
		return
	}

	// Delete associated files from filesystem
	for _, file := range existingLead.Files {
		if file.FilePath != "" {
			if err := os.Remove(file.FilePath); err != nil {
			}
		}
	}
	if err := tx.Exec("DELETE FROM lead_associates WHERE lead_id = ?", id); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to delete lead associations",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	// Delete files
	if err := tx.Where("lead_id = ?", id).Delete(&models.File{}); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to delete lead files",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	// Delete residential entries
	if err := tx.Where("lead_id = ?", id).Delete(&models.ResidentialEntry{}); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to delete residential entries",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	// Delete estimates and their related data
	var estimates []models.Estimate
	if err := tx.Where("lead_id = ?", id).Find(&estimates); err == nil {
		for _, estimate := range estimates {
			// Delete material rows
			if err := tx.Where("estimate_id = ?", estimate.ID).Delete(&models.MaterialRow{}); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "Failed to delete estimate material rows",
					Code:    "DATABASE_ERROR",
					Details: err.Error(),
				})
				return
			}
			// Delete service rows
			if err := tx.Where("estimate_id = ?", estimate.ID).Delete(&models.ServiceRow{}); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "Failed to delete estimate service rows",
					Code:    "DATABASE_ERROR",
					Details: err.Error(),
				})
				return
			}
		}
		// Delete estimates
		if err := tx.Where("lead_id = ?", id).Delete(&models.Estimate{}); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to delete estimates",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
			return
		}
	}

	// Delete addresses if they exist and are not shared
	if existingLead.Details.ShippingAddressID != nil {
		// Check if this address is used by other leads
		var count int64
		tx.Model(&models.Lead{}).Where("details->>'shipping_address_id' = ? AND id != ?",
			fmt.Sprintf("%d", *existingLead.Details.ShippingAddressID), id).Count(&count)
		if count == 0 {
			if err := tx.Delete(&models.Address{}, *existingLead.Details.ShippingAddressID); err != nil {
				// Log warning but don't fail
			}
		}
	}

	if existingLead.Details.BillingAddressID != nil &&
		existingLead.Details.BillingAddressID != existingLead.Details.ShippingAddressID {
		// Check if this address is used by other leads
		var count int64
		tx.Model(&models.Lead{}).Where("details->>'billing_address_id' = ? AND id != ?",
			fmt.Sprintf("%d", *existingLead.Details.BillingAddressID), id).Count(&count)
		if count == 0 {
			if err := tx.Delete(&models.Address{}, *existingLead.Details.BillingAddressID); err != nil {
			}
		}
	}

	// Finally, delete the lead itself
	if err := tx.Delete(&existingLead); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to delete lead",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	// Commit transaction
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
		"message": "Lead deleted successfully",
		"id":      id,
	})
}

func (h *LeadHandler) ToggleFavorite(c *gin.Context) {
	db := middleware.GetDB(c)

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid lead ID",
			Code:  "BAD_REQUEST",
		})
		return
	}

	// Begin transaction
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Unexpected error occurred",
				Code:    "INTERNAL_SERVER_ERROR",
				Details: fmt.Sprintf("%v", r),
			})
		}
	}()
	var existingLead models.Lead
	if err := tx.First(&existingLead, "id = ?", id); err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Lead not found",
				Code:    "NOT_FOUND",
				Details: "No lead found with the specified ID",
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

	var updateData struct {
		Favorite string `json:"favorite" gorm:"text"`
	}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Code:    "VALIDATION_ERROR",
			Details: err.Error(),
		})
		return
	}

	// Capture old favorite status before updating
	oldFavoriteStatus := existingLead.Favorite

	// Update the favorite field
	existingLead.Favorite = updateData.Favorite

	// Save the updated lead
	if err := tx.Save(&existingLead); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to update lead",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	// Get user info for activity event
	userID, userName := getRequestUser(c)
	creatorLabel := userName
	if creatorLabel == "" {
		creatorLabel = userID
	}

	// Attribution fallback for consistency (similar to CreateLead/UpdateLead)
	if (userID == "system" || userID == "") && existingLead.Sales1PersonID != "" {
		var s1 models.Agent
		if err := db.First(&s1, "id = ?", existingLead.Sales1PersonID); err == nil {
			if userID == "system" {
				userID = s1.ID
			}
			if userName == "System" || userName == "" {
				userName = s1.Name
			}
		}
	}

	activityEvent := activity.ActivityEvent{
		EntityType: "lead",
		EntityID:   existingLead.ID,
		Action:     "favorite_toggled",
		UserID:     userID,
		UserName:   userName,
		Payload: map[string]any{
			"lead_name":    existingLead.JobName,
			"old_favorite": oldFavoriteStatus,
			"new_favorite": existingLead.Favorite,
			"updated_by":   creatorLabel,
		},
	}
	if err := activity.InsertActivityEvent(c.Request.Context(), tx.GetUnderlying(), activityEvent); err != nil { // This line was already correct from previous fix
		log.Printf("activity event publish failed: %v", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to commit transaction",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"message": "Lead favorite status updated successfully",
		"lead":    existingLead,
	})
}

func (h *LeadHandler) UpdateStatus(c *gin.Context) {
	db := middleware.GetDB(c)

	id := c.Param("id")
	status := c.Query("status")
	rejectionReason := c.Query("rejection_reason")
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid lead ID",
			Code:  "BAD_REQUEST",
		})
		return
	}

	// Begin transaction
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Unexpected error occurred",
				Code:    "INTERNAL_SERVER_ERROR",
				Details: fmt.Sprintf("%v", r),
			})
		}
	}()

	// Check if lead exists
	var existingLead models.Lead
	if err := tx.First(&existingLead, "id = ?", id); err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Lead not found",
				Code:    "NOT_FOUND",
				Details: "No lead found with the specified ID",
			})
			return
		}
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Database error",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	// Capture old status before it's updated
	oldStatus := existingLead.Status

	// Save the updated lead
	switch status {
	case "NEW":
		existingLead.Status = models.LeadStatus(status)
		existingLead.RejectionReason = ""
	case "IN_PROGRESS":
		existingLead.Status = models.LeadStatus(status)
		existingLead.RejectionReason = ""
	case "COMPLETED":
		existingLead.Status = models.LeadStatus(status)
		existingLead.RejectionReason = ""

		// Upsert the job synchronously
		// immediately on the very next request
		if err := tx.Save(&existingLead); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to update lead status",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
			return
		}
		jobService := services.NewJobService()
		if err := jobService.UpsertJobFromLead(db, existingLead); err != nil {
			// Log but don't fail the status update — job sync is best-effort
			fmt.Printf("WARN: UpsertJobFromLead failed for lead %s: %v\n", existingLead.ID, err)
		}
	case "REJECTED":
		existingLead.Status = models.LeadStatus(status)
		if rejectionReason == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Rejection reason is required for REJECTED status"})
			return
		}
		existingLead.RejectionReason = rejectionReason
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status"})
		return
	}
	if err := tx.Save(&existingLead); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to update lead",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	userID, userName := middleware.GetRequestUser(c)
	creatorLabel := userName
	if creatorLabel == "" {
		creatorLabel = userID
	}

	// Attribution fallback for consistency (similar to CreateLead/UpdateLead)
	if (userID == "system" || userID == "") && existingLead.Sales1PersonID != "" {
		userID = existingLead.Sales1PersonID
		userName = existingLead.Sales1PersonName
		creatorLabel = userName
	}

	// Publish activity event for status update (Fix: ensures dragging to Rejected/Sold shows in dashboard)
	activityEvent := activity.ActivityEvent{
		EntityType: "lead",
		EntityID:   existingLead.ID,
		Action:     "status_updated",
		UserID:     userID,
		UserName:   userName,
		Payload: map[string]any{
			"lead_name":  existingLead.JobName,
			"old_status": oldStatus,
			"new_status": status,
			"updated_by": creatorLabel,
		},
	}
	if err := activity.InsertActivityEvent(c.Request.Context(), tx.GetUnderlying(), activityEvent); err != nil { // This line was already correct from previous fix
		log.Printf("activity event publish failed: %v", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to commit transaction",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"message": "Lead favorite status updated successfully",
		"lead":    existingLead,
	})
}

// func (h *LeadHandler) UpdateStatus(c *gin.Context) {
// 	db := middleware.GetDB(c)

// 	id := c.Param("id")
// 	status := c.Query("status")
// 	rejectionReason := c.Query("rejection_reason")
// 	if id == "" {
// 		c.JSON(http.StatusBadRequest, ErrorResponse{
// 			Error: "Invalid lead ID",
// 			Code:  "BAD_REQUEST",
// 		})
// 		return
// 	}

// 	// Begin transaction
// 	tx := db.Begin()
// 	defer func() {
// 		if r := recover(); r != nil {
// 			tx.Rollback()
// 			c.JSON(http.StatusInternalServerError, ErrorResponse{
// 				Error:   "Unexpected error occurred",
// 				Code:    "INTERNAL_SERVER_ERROR",
// 				Details: fmt.Sprintf("%v", r),
// 			})
// 		}
// 	}()

// 	// Check if lead exists
// 	var existingLead models.Lead
// 	if err := tx.First(&existingLead, "id = ?", id); err != nil {
// 		tx.Rollback()
// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			c.JSON(http.StatusNotFound, ErrorResponse{
// 				Error:   "Lead not found",
// 				Code:    "NOT_FOUND",
// 				Details: "No lead found with the specified ID",
// 			})
// 		} else {
// 			c.JSON(http.StatusInternalServerError, ErrorResponse{
// 				Error:   "Database error",
// 				Code:    "DATABASE_ERROR",
// 				Details: err.Error(),
// 			})
// 		}
// 		return
// 	}
// 	// Save the updated lead
// 	switch status {
// 	case "NEW":
// 		existingLead.Status = models.LeadStatus(status)
// 		existingLead.RejectionReason = ""
// 	case "IN_PROGRESS":
// 		existingLead.Status = models.LeadStatus(status)
// 		existingLead.RejectionReason = ""
// 	case "COMPLETED":
// 		existingLead.Status = models.LeadStatus(status)
// 		existingLead.RejectionReason = ""

// 		// Upsert the job synchronously
// 		// immediately on the very next request
// 		if err := tx.Save(&existingLead); err != nil {
// 			tx.Rollback()
// 			c.JSON(http.StatusInternalServerError, ErrorResponse{
// 				Error:   "Failed to update lead status",
// 				Code:    "DATABASE_ERROR",
// 				Details: err.Error(),
// 			})
// 			return
// 		}
// 		jobService := services.NewJobService()
// 		if err := jobService.UpsertJobFromLead(db, existingLead); err != nil {
// 			// Log but don't fail the status update — job sync is best-effort
// 			fmt.Printf("WARN: UpsertJobFromLead failed for lead %s: %v\n", existingLead.ID, err)
// 		}
// 	case "REJECTED":
// 		existingLead.Status = models.LeadStatus(status)
// 		if rejectionReason == "" {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "Rejection reason is required for REJECTED status"})
// 			return
// 		}
// 		existingLead.RejectionReason = rejectionReason
// 	default:
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status"})
// 		return
// 	}
// 	if err := tx.Save(&existingLead); err != nil {
// 		tx.Rollback()
// 		c.JSON(http.StatusInternalServerError, ErrorResponse{
// 			Error:   "Failed to update lead",
// 			Code:    "DATABASE_ERROR",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	// Commit the transaction
// 	if err := tx.Commit(); err != nil {
// 		c.JSON(http.StatusInternalServerError, ErrorResponse{
// 			Error:   "Failed to commit transaction",
// 			Code:    "DATABASE_ERROR",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

//		// Return success response
//		c.JSON(http.StatusOK, gin.H{
//			"message": "Lead favorite status updated successfully",
//			"lead":    existingLead,
//		})
//	}
func (h *LeadHandler) ServeProfilePic(c *gin.Context) {
	db := middleware.GetDB(c)

	ID := c.Param("id")

	var lead models.Lead
	if err := db.Where("id = ?", ID).First(&lead); err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"error": "Lead not found",
			"id":    ID,
		})
		return
	}

	if lead.ProfilePic == "" {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"error": "No profile picture found for this lead",
		})
		return
	}

	// Support both legacy full URLs and new bare S3 keys (Option A migration)
	s3Key := lead.ProfilePic
	if strings.HasPrefix(s3Key, "https://") || strings.HasPrefix(s3Key, "http://") {
		s3Key = h.extractS3Key(s3Key)
	}

	if s3Key == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid profile picture path"})
		return
	}

	req, _ := h.s3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket:                     aws.String(h.getBucketName(c)),
		Key:                        aws.String(s3Key),
		ResponseContentDisposition: attachmentDispositionForKey(s3Key),
	})

	presignedURL, err := req.Presign(1 * time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate profile picture URL"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"download_url": presignedURL,
		"expires_in":   "1 hour",
	})
}
