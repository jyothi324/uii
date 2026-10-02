package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler struct to hold database connection

// NewHandler creates a new handler instance

// Common response structures
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type PaginationResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// Helper functions
func (h *Handler) parseUUID(c *gin.Context, param string) (uuid.UUID, bool) {
	idStr := c.Param(param)
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid UUID format",
			Code:    "INVALID_UUID",
			Details: "The provided ID is not a valid UUID",
		})
		return uuid.Nil, false
	}
	return id, true
}

func (h *Handler) getPaginationParams(c *gin.Context) (int, int) {
	page := 1
	pageSize := 20 // Default page size

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	return page, pageSize
}

// func (h *Handler) handleDBError(c *gin.Context, err error, operation string) {
// 	if h.DB.Error() {
// 		c.JSON(http.StatusNotFound, ErrorResponse{
// 			Error:   "Resource not found",
// 			Code:    "NOT_FOUND",
// 			Details: "The requested resource does not exist",
// 		})
// 		return
// 	}

// 	// Log the actual error for debugging (you should use a proper logger)
// 	// log.Printf("Database error during %s: %v", operation, err)

// 	c.JSON(http.StatusInternalServerError, ErrorResponse{
// 		Error:   "Internal server error",
// 		Code:    "DATABASE_ERROR",
// 		Details: "An error occurred while processing your request",
// 	})
// }

// // =============================================================================
// // LEAD HANDLERS
// // =============================================================================

// func (h *Handler) CreateLead(c *gin.Context) {
// 	var lead models.Lead

// 	if err := c.ShouldBindJSON(&lead); err != nil {
// 		c.JSON(http.StatusBadRequest, ErrorResponse{
// 			Error:   "Invalid request body",
// 			Code:    "VALIDATION_ERROR",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	// Validate required fields
// 	if lead.DeliveryType == "" || lead.LeadType == "" || lead.DetailedLeadStatus == "" ||
// 		lead.LeadStatus == "" || lead.PaymentTerm == "" || lead.PricingType == "" {
// 		c.JSON(http.StatusBadRequest, ErrorResponse{
// 			Error:   "Missing required fields",
// 			Code:    "VALIDATION_ERROR",
// 			Details: "DeliveryType, LeadType, DetailedLeadStatus, LeadStatus, PaymentTerm, and PricingType are required",
// 		})
// 		return
// 	}

// 	if err := h.DB.Create(&lead).Error; err != nil {
// 		h.handleDBError(c, err, "create lead")
// 		return
// 	}

// 	c.JSON(http.StatusCreated, SuccessResponse{
// 		Message: "Lead created successfully",
// 		Data:    lead,
// 	})
// }

// func (h *Handler) GetLead(c *gin.Context) {
// 	id, ok := h.parseUUID(c, "id")
// 	if !ok {
// 		return
// 	}

// 	var lead models.Lead
// 	query := h.DB.Preload("Address").
// 		Preload("AssociateCustomer").
// 		Preload("Files").
// 		Preload("ResidentialEntries").
// 		Preload("Estimates").
// 		Preload("Agents")

// 	if err := query.First(&lead, id).Error; err != nil {
// 		h.handleDBError(c, err, "get lead")
// 		return
// 	}

// 	c.JSON(http.StatusOK, lead)
// }

// func (h *Handler) UpdateLead(c *gin.Context) {
// 	id, ok := h.parseUUID(c, "id")
// 	if !ok {
// 		return
// 	}

// 	var existingLead models.Lead
// 	if err := h.DB.First(&existingLead, id).Error; err != nil {
// 		h.handleDBError(c, err, "find lead for update")
// 		return
// 	}

// 	var updateData models.Lead
// 	if err := c.ShouldBindJSON(&updateData); err != nil {
// 		c.JSON(http.StatusBadRequest, ErrorResponse{
// 			Error:   "Invalid request body",
// 			Code:    "VALIDATION_ERROR",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	// Preserve the ID and timestamps
// 	updateData.ID = existingLead.ID

// 	if err := h.DB.Save(&updateData).Error; err != nil {
// 		h.handleDBError(c, err, "update lead")
// 		return
// 	}

// 	c.JSON(http.StatusOK, SuccessResponse{
// 		Message: "Lead updated successfully",
// 		Data:    updateData,
// 	})
// }

// func (h *Handler) DeleteLead(c *gin.Context) {
// 	id, ok := h.parseUUID(c, "id")
// 	if !ok {
// 		return
// 	}

// 	// Check if lead exists
// 	var lead models.Lead
// 	if err := h.DB.First(&lead, id).Error; err != nil {
// 		h.handleDBError(c, err, "find lead for deletion")
// 		return
// 	}

// 	if err := h.DB.Delete(&lead).Error; err != nil {
// 		h.handleDBError(c, err, "delete lead")
// 		return
// 	}

// 	c.JSON(http.StatusOK, SuccessResponse{
// 		Message: "Lead deleted successfully",
// 	})
// }

// func (h *Handler) ListLeads(c *gin.Context) {
// 	page, pageSize := h.getPaginationParams(c)
// 	offset := (page - 1) * pageSize

// 	var leads []models.Lead
// 	var total int64

// 	// Count total records
// 	if err := h.DB.Model(&models.Lead{}).Count(&total).Error; err != nil {
// 		h.handleDBError(c, err, "count leads")
// 		return
// 	}

// 	// Apply filters
// 	query := h.DB.Preload("Address").
// 		Preload("AssociateCustomer").
// 		Preload("Files").
// 		Preload("ResidentialEntries").
// 		Preload("Estimates").
// 		Preload("Agents")

// 	// Add search functionality
// 	if search := c.Query("search"); search != "" {
// 		searchTerm := "%" + strings.ToLower(search) + "%"
// 		query = query.Where("LOWER(lead_type) LIKE ? OR LOWER(lead_status) LIKE ? OR LOWER(detailed_lead_status) LIKE ?",
// 			searchTerm, searchTerm, searchTerm)
// 	}

// 	// Add status filter
// 	if status := c.Query("status"); status != "" {
// 		query = query.Where("lead_status = ?", status)
// 	}

// 	// Add lead type filter
// 	if leadType := c.Query("lead_type"); leadType != "" {
// 		query = query.Where("lead_type = ?", leadType)
// 	}

// 	if err := query.Limit(pageSize).Offset(offset).Find(&leads).Error; err != nil {
// 		h.handleDBError(c, err, "list leads")
// 		return
// 	}

// 	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

// 	c.JSON(http.StatusOK, PaginationResponse{
// 		Data:       leads,
// 		Total:      total,
// 		Page:       page,
// 		PageSize:   pageSize,
// 		TotalPages: totalPages,
// 	})
// }

// // =============================================================================
// // ESTIMATE HANDLERS
// // =============================================================================

// func (h *Handler) CreateEstimate(c *gin.Context) {
// 	var estimate models.Estimate

// 	if err := c.ShouldBindJSON(&estimate); err != nil {
// 		c.JSON(http.StatusBadRequest, ErrorResponse{
// 			Error:   "Invalid request body",
// 			Code:    "VALIDATION_ERROR",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	// Validate required fields
// 	if estimate.Name == "" || estimate.Status == "" {
// 		c.JSON(http.StatusBadRequest, ErrorResponse{
// 			Error:   "Missing required fields",
// 			Code:    "VALIDATION_ERROR",
// 			Details: "Name and Status are required",
// 		})
// 		return
// 	}

// 	if err := h.DB.Create(&estimate).Error; err != nil {
// 		h.handleDBError(c, err, "create estimate")
// 		return
// 	}

// 	c.JSON(http.StatusCreated, SuccessResponse{
// 		Message: "Estimate created successfully",
// 		Data:    estimate,
// 	})
// }

// func (h *Handler) GetEstimate(c *gin.Context) {
// 	id, ok := h.parseUUID(c, "id")
// 	if !ok {
// 		return
// 	}

// 	var estimate models.Estimate
// 	query := h.DB.Preload("Lead").
// 		Preload("MaterialRows").
// 		Preload("ServiceRows").
// 		Preload("ResidentialEntries")

// 	if err := query.First(&estimate, id).Error; err != nil {
// 		h.handleDBError(c, err, "get estimate")
// 		return
// 	}

// 	c.JSON(http.StatusOK, estimate)
// }

// func (h *Handler) UpdateEstimate(c *gin.Context) {
// 	id, ok := h.parseUUID(c, "id")
// 	if !ok {
// 		return
// 	}

// 	var existingEstimate models.Estimate
// 	if err := h.DB.First(&existingEstimate, id).Error; err != nil {
// 		h.handleDBError(c, err, "find estimate for update")
// 		return
// 	}

// 	var updateData models.Estimate
// 	if err := c.ShouldBindJSON(&updateData); err != nil {
// 		c.JSON(http.StatusBadRequest, ErrorResponse{
// 			Error:   "Invalid request body",
// 			Code:    "VALIDATION_ERROR",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	// Preserve the ID
// 	updateData.ID = existingEstimate.ID

// 	if err := h.DB.Save(&updateData).Error; err != nil {
// 		h.handleDBError(c, err, "update estimate")
// 		return
// 	}

// 	c.JSON(http.StatusOK, SuccessResponse{
// 		Message: "Estimate updated successfully",
// 		Data:    updateData,
// 	})
// }

// func (h *Handler) DeleteEstimate(c *gin.Context) {
// 	id, ok := h.parseUUID(c, "id")
// 	if !ok {
// 		return
// 	}

// 	// Check if estimate exists
// 	var estimate models.Estimate
// 	if err := h.DB.First(&estimate, id).Error; err != nil {
// 		h.handleDBError(c, err, "find estimate for deletion")
// 		return
// 	}

// 	if err := h.DB.Delete(&estimate).Error; err != nil {
// 		h.handleDBError(c, err, "delete estimate")
// 		return
// 	}

// 	c.JSON(http.StatusOK, SuccessResponse{
// 		Message: "Estimate deleted successfully",
// 	})
// }

// func (h *Handler) ListEstimates(c *gin.Context) {
// 	page, pageSize := h.getPaginationParams(c)
// 	offset := (page - 1) * pageSize

// 	var estimates []models.Estimate
// 	var total int64

// 	// Count total records
// 	if err := h.DB.Model(&models.Estimate{}).Count(&total).Error; err != nil {
// 		h.handleDBError(c, err, "count estimates")
// 		return
// 	}

// 	query := h.DB.Preload("Lead").
// 		Preload("MaterialRows").
// 		Preload("ServiceRows").
// 		Preload("ResidentialEntries")

// 	// Add search functionality
// 	if search := c.Query("search"); search != "" {
// 		searchTerm := "%" + strings.ToLower(search) + "%"
// 		query = query.Where("LOWER(name) LIKE ? OR LOWER(status) LIKE ?", searchTerm, searchTerm)
// 	}

// 	// Add status filter
// 	if status := c.Query("status"); status != "" {
// 		query = query.Where("status = ?", status)
// 	}

// 	// Add lead filter
// 	if leadID := c.Query("lead_id"); leadID != "" {
// 		if id, err := uuid.Parse(leadID); err == nil {
// 			query = query.Where("lead_id = ?", id)
// 		}
// 	}

// 	if err := query.Limit(pageSize).Offset(offset).Find(&estimates).Error; err != nil {
// 		h.handleDBError(c, err, "list estimates")
// 		return
// 	}

// 	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

// 	c.JSON(http.StatusOK, PaginationResponse{
// 		Data:       estimates,
// 		Total:      total,
// 		Page:       page,
// 		PageSize:   pageSize,
// 		TotalPages: totalPages,
// 	})
// }

// // Response represents a standard API response
// type Response struct {
// 	Data    interface{} `json:"data,omitempty"`
// 	Message string      `json:"message,omitempty"`
// 	Error   string      `json:"error,omitempty"`
// }

// // ListOptions represents query parameters for listing
// type ListOptions struct {
// 	Page     int    `form:"page,default=1"`
// 	Limit    int    `form:"limit,default=10"`
// 	Sort     string `form:"sort,default=created_at"`
// 	Order    string `form:"order,default=desc"`
// 	Search   string `form:"search"`
// 	Preloads []string
// }

// // Handler represents the main handler struct

// // bindJSON binds JSON payload and handles errors
// func (h *Handler) bindJSON(c *gin.Context, obj interface{}) bool {
// 	if err := c.ShouldBindJSON(obj); err != nil {
// 		c.JSON(http.StatusBadRequest, Response{Error: err.Error()})
// 		return false
// 	}
// 	return true
// }

// // applyPreloads applies preloads to query
// func (h *Handler) applyPreloads(db *gorm.DB, preloads []string) *gorm.DB {
// 	for _, preload := range preloads {
// 		db = db.Preload(preload)
// 	}
// 	return db
// }

// // applyListOptions applies pagination, sorting, and filtering
// func (h *Handler) applyListOptions(db *gorm.DB, opts ListOptions) *gorm.DB {
// 	// Apply sorting
// 	if opts.Sort != "" {
// 		orderClause := opts.Sort
// 		if opts.Order == "desc" {
// 			orderClause += " DESC"
// 		}
// 		db = db.Order(orderClause)
// 	}

// 	// Apply pagination
// 	if opts.Limit > 0 {
// 		offset := (opts.Page - 1) * opts.Limit
// 		db = db.Offset(offset).Limit(opts.Limit)
// 	}

// 	return db
// }

// // Generic CRUD operations

// // Create creates a new record
// func (h *Handler) Create(model interface{}) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		if !h.bindJSON(c, model) {
// 			return
// 		}

// 		if err := h.DB.Create(model).Error; err != nil {
// 			c.JSON(http.StatusInternalServerError, Response{Error: err.Error()})
// 			return
// 		}

// 		c.JSON(http.StatusCreated, Response{Data: model})
// 	}
// }

// // Get retrieves a single record by ID
// func (h *Handler) Get(model interface{}, preloads []string) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		id, ok := h.parseUUID(c, "id")
// 		if !ok {
// 			return
// 		}

// 		db := h.DB
// 		db = h.applyPreloads(db, preloads)

// 		if err := db.First(model, id).Error; err != nil {
// 			if err == gorm.ErrRecordNotFound {
// 				c.JSON(http.StatusNotFound, Response{Error: "Record not found"})
// 			} else {
// 				c.JSON(http.StatusInternalServerError, Response{Error: err.Error()})
// 			}
// 			return
// 		}

// 		c.JSON(http.StatusOK, Response{Data: model})
// 	}
// }

// // Update updates an existing record
// func (h *Handler) Update(model interface{}) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		id, ok := h.parseUUID(c, "id")
// 		if !ok {
// 			return
// 		}

// 		if err := h.DB.First(model, id).Error; err != nil {
// 			if err == gorm.ErrRecordNotFound {
// 				c.JSON(http.StatusNotFound, Response{Error: "Record not found"})
// 			} else {
// 				c.JSON(http.StatusInternalServerError, Response{Error: err.Error()})
// 			}
// 			return
// 		}

// 		if !h.bindJSON(c, model) {
// 			return
// 		}

// 		if err := h.DB.Save(model).Error; err != nil {
// 			c.JSON(http.StatusInternalServerError, Response{Error: err.Error()})
// 			return
// 		}

// 		c.JSON(http.StatusOK, Response{Data: model})
// 	}
// }

// // Delete removes a record by ID
// func (h *Handler) Delete(model interface{}, entityName string) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		id, ok := h.parseUUID(c, "id")
// 		if !ok {
// 			return
// 		}

// 		if err := h.DB.Delete(model, id).Error; err != nil {
// 			c.JSON(http.StatusInternalServerError, Response{Error: err.Error()})
// 			return
// 		}

// 		c.JSON(http.StatusOK, Response{Message: entityName + " deleted successfully"})
// 	}
// }

// // List retrieves multiple records with pagination and filtering
// func (h *Handler) List(model interface{}, preloads []string) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		var opts ListOptions
// 		if err := c.ShouldBindQuery(&opts); err != nil {
// 			c.JSON(http.StatusBadRequest, Response{Error: err.Error()})
// 			return
// 		}

// 		db := h.DB
// 		db = h.applyPreloads(db, preloads)
// 		db = h.applyListOptions(db, opts)

// 		if err := db.Find(model).Error; err != nil {
// 			c.JSON(http.StatusInternalServerError, Response{Error: err.Error()})
// 			return
// 		}

// 		c.JSON(http.StatusOK, Response{Data: model})
// 	}
// }

// // Specific Handlers

// // FollowUpRules
// func (h *Handler) ListFollowUpRules(c *gin.Context) {
// 	var followUpRules []models.FollowUpRule
// 	h.List(&followUpRules, nil)(c)
// }

// // Tasks
// func (h *Handler) CreateTask(c *gin.Context) {
// 	var task models.Task
// 	h.Create(&task)(c)
// }

// func (h *Handler) GetTask(c *gin.Context) {
// 	var task models.Task
// 	h.Get(&task, nil)(c)
// }

// func (h *Handler) UpdateTask(c *gin.Context) {
// 	var task models.Task
// 	h.Update(&task)(c)
// }

// func (h *Handler) DeleteTask(c *gin.Context) {
// 	h.Delete(&models.Task{}, "Task")(c)
// }

// func (h *Handler) ListTasks(c *gin.Context) {
// 	var tasks []models.Task
// 	h.List(&tasks, nil)(c)
// }

// // Addresses
// func (h *Handler) CreateAddress(c *gin.Context) {
// 	var address models.Address
// 	h.Create(&address)(c)
// }

// func (h *Handler) GetAddress(c *gin.Context) {
// 	var address models.Address
// 	h.Get(&address, nil)(c)
// }

// func (h *Handler) UpdateAddress(c *gin.Context) {
// 	var address models.Address
// 	h.Update(&address)(c)
// }

// func (h *Handler) DeleteAddress(c *gin.Context) {
// 	h.Delete(&models.Address{}, "Address")(c)
// }

// func (h *Handler) ListAddresses(c *gin.Context) {
// 	var addresses []models.Address
// 	h.List(&addresses, nil)(c)
// }

// // AssociateCustomers
// func (h *Handler) CreateAssociateCustomer(c *gin.Context) {
// 	var associateCustomer models.AssociateCustomer
// 	h.Create(&associateCustomer)(c)
// }

// func (h *Handler) GetAssociateCustomer(c *gin.Context) {
// 	var associateCustomer models.AssociateCustomer
// 	h.Get(&associateCustomer, nil)(c)
// }

// func (h *Handler) UpdateAssociateCustomer(c *gin.Context) {
// 	var associateCustomer models.AssociateCustomer
// 	h.Update(&associateCustomer)(c)
// }

// func (h *Handler) DeleteAssociateCustomer(c *gin.Context) {
// 	h.Delete(&models.AssociateCustomer{}, "AssociateCustomer")(c)
// }

// func (h *Handler) ListAssociateCustomers(c *gin.Context) {
// 	var associateCustomers []models.AssociateCustomer
// 	h.List(&associateCustomers, nil)(c)
// }

// // Files
// func (h *Handler) CreateFile(c *gin.Context) {
// 	var file models.File
// 	h.Create(&file)(c)
// }

// func (h *Handler) GetFile(c *gin.Context) {
// 	var file models.File
// 	h.Get(&file, []string{"Lead"})(c)
// }

// func (h *Handler) UpdateFile(c *gin.Context) {
// 	var file models.File
// 	h.Update(&file)(c)
// }

// func (h *Handler) DeleteFile(c *gin.Context) {
// 	h.Delete(&models.File{}, "File")(c)
// }

// func (h *Handler) ListFiles(c *gin.Context) {
// 	var files []models.File
// 	h.List(&files, []string{"Lead"})(c)
// }

// // ResidentialEntries
// func (h *Handler) CreateResidentialEntry(c *gin.Context) {
// 	var residentialEntry models.ResidentialEntry
// 	h.Create(&residentialEntry)(c)
// }

// func (h *Handler) GetResidentialEntry(c *gin.Context) {
// 	var residentialEntry models.ResidentialEntry
// 	h.Get(&residentialEntry, []string{"Lead", "Estimate"})(c)
// }

// func (h *Handler) UpdateResidentialEntry(c *gin.Context) {
// 	var residentialEntry models.ResidentialEntry
// 	h.Update(&residentialEntry)(c)
// }

// func (h *Handler) DeleteResidentialEntry(c *gin.Context) {
// 	h.Delete(&models.ResidentialEntry{}, "ResidentialEntry")(c)
// }

// func (h *Handler) ListResidentialEntries(c *gin.Context) {
// 	var residentialEntries []models.ResidentialEntry
// 	h.List(&residentialEntries, []string{"Lead", "Estimate"})(c)
// }

// // ManagerResidentialEntries
// func (h *Handler) CreateManagerResidentialEntry(c *gin.Context) {
// 	var managerResidentialEntry models.ManagerResidentialEntry
// 	h.Create(&managerResidentialEntry)(c)
// }

// func (h *Handler) GetManagerResidentialEntry(c *gin.Context) {
// 	var managerResidentialEntry models.ManagerResidentialEntry
// 	h.Get(&managerResidentialEntry, nil)(c)
// }

// func (h *Handler) UpdateManagerResidentialEntry(c *gin.Context) {
// 	var managerResidentialEntry models.ManagerResidentialEntry
// 	h.Update(&managerResidentialEntry)(c)
// }

// func (h *Handler) DeleteManagerResidentialEntry(c *gin.Context) {
// 	h.Delete(&models.ManagerResidentialEntry{}, "ManagerResidentialEntry")(c)
// }

// func (h *Handler) ListManagerResidentialEntries(c *gin.Context) {
// 	var managerResidentialEntries []models.ManagerResidentialEntry
// 	h.List(&managerResidentialEntries, nil)(c)
// }

// // MaterialRows
// func (h *Handler) CreateMaterialRow(c *gin.Context) {
// 	var materialRow models.MaterialRow
// 	h.Create(&materialRow)(c)
// }

// func (h *Handler) GetMaterialRow(c *gin.Context) {
// 	var materialRow models.MaterialRow
// 	h.Get(&materialRow, []string{"Estimate"})(c)
// }

// func (h *Handler) UpdateMaterialRow(c *gin.Context) {
// 	var materialRow models.MaterialRow
// 	h.Update(&materialRow)(c)
// }

// func (h *Handler) DeleteMaterialRow(c *gin.Context) {
// 	h.Delete(&models.MaterialRow{}, "MaterialRow")(c)
// }

// func (h *Handler) ListMaterialRows(c *gin.Context) {
// 	var materialRows []models.MaterialRow
// 	h.List(&materialRows, []string{"Estimate"})(c)
// }

// // ServiceRows
// func (h *Handler) CreateServiceRow(c *gin.Context) {
// 	var serviceRow models.ServiceRow
// 	h.Create(&serviceRow)(c)
// }

// func (h *Handler) GetServiceRow(c *gin.Context) {
// 	var serviceRow models.ServiceRow
// 	h.Get(&serviceRow, []string{"Estimate"})(c)
// }

// func (h *Handler) UpdateServiceRow(c *gin.Context) {
// 	var serviceRow models.ServiceRow
// 	h.Update(&serviceRow)(c)
// }

// func (h *Handler) DeleteServiceRow(c *gin.Context) {
// 	h.Delete(&models.ServiceRow{}, "ServiceRow")(c)
// }

// func (h *Handler) ListServiceRows(c *gin.Context) {
// 	var serviceRows []models.ServiceRow
// 	h.List(&serviceRows, []string{"Estimate"})(c)
// }

// // AgentLeads (Special case for composite key)
// func (h *Handler) CreateAgentLead(c *gin.Context) {
// 	var agentLead models.AgentLead
// 	h.Create(&agentLead)(c)
// }

// func (h *Handler) GetAgentLead(c *gin.Context) {
// 	agentID, ok := h.parseUUID(c, "agentID")
// 	if !ok {
// 		return
// 	}
// 	leadID, ok := h.parseUUID(c, "leadID")
// 	if !ok {
// 		return
// 	}

// 	var agentLead models.AgentLead
// 	if err := h.DB.Preload("Agent").Preload("Lead").
// 		First(&agentLead, "agent_id = ? AND lead_id = ?", agentID, leadID).Error; err != nil {
// 		if err == gorm.ErrRecordNotFound {
// 			c.JSON(http.StatusNotFound, Response{Error: "AgentLead not found"})
// 		} else {
// 			c.JSON(http.StatusInternalServerError, Response{Error: err.Error()})
// 		}
// 		return
// 	}

// 	c.JSON(http.StatusOK, Response{Data: agentLead})
// }

// func (h *Handler) DeleteAgentLead(c *gin.Context) {
// 	agentID, ok := h.parseUUID(c, "agentID")
// 	if !ok {
// 		return
// 	}
// 	leadID, ok := h.parseUUID(c, "leadID")
// 	if !ok {
// 		return
// 	}

// 	if err := h.DB.Delete(&models.AgentLead{}, "agent_id = ? AND lead_id = ?", agentID, leadID).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, Response{Error: err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, Response{Message: "AgentLead deleted successfully"})
// }

// func (h *Handler) ListAgentLeads(c *gin.Context) {
// 	var agentLeads []models.AgentLead
// 	h.List(&agentLeads, []string{"Agent", "Lead"})(c)
// }

// // Batch operations
// func (h *Handler) BatchDelete(model interface{}, entityName string) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		var ids []uuid.UUID
// 		if !h.bindJSON(c, &ids) {
// 			return
// 		}

// 		if len(ids) == 0 {
// 			c.JSON(http.StatusBadRequest, Response{Error: "No IDs provided"})
// 			return
// 		}

// 		result := h.DB.Delete(model, ids)
// 		if result.Error != nil {
// 			c.JSON(http.StatusInternalServerError, Response{Error: result.Error.Error()})
// 			return
// 		}

// 		c.JSON(http.StatusOK, Response{
// 			Message: entityName + "s deleted successfully",
// 			Data:    map[string]int64{"deleted_count": result.RowsAffected},
// 		})
// 	}
// }

// // Health check
// func (h *Handler) HealthCheck(c *gin.Context) {
// 	c.JSON(http.StatusOK, Response{
// 		Message: "Service is healthy",
// 		Data: map[string]interface{}{
// 			"status":    "ok",
// 			"timestamp": time.Now().Unix(),
// 		},
// 	})
// }

// // SetupRoutes sets up the API routes
