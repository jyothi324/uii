package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tuliporg/agent-management-api/db"
	"github.com/tuliporg/agent-management-api/internal/activity"
	"github.com/tuliporg/agent-management-api/middleware"
	"github.com/tuliporg/agent-management-api/models"
	"github.com/tuliporg/agent-management-api/utils"
	"gorm.io/gorm"
)

// sortEstimateRowsBySortOrder orders MaterialRows/ServiceRows in-place using the
// persisted Details.SortOrder value, so drag-reordered sequences survive a save + refetch.
func sortEstimateRowsBySortOrder(estimate *models.Estimate) {
	sort.SliceStable(estimate.MaterialRows, func(i, j int) bool {
		return estimate.MaterialRows[i].Details.SortOrder < estimate.MaterialRows[j].Details.SortOrder
	})
	sort.SliceStable(estimate.ServiceRows, func(i, j int) bool {
		return estimate.ServiceRows[i].Details.SortOrder < estimate.ServiceRows[j].Details.SortOrder
	})
}

func setEstimateSalesPersons(store db.Database, estimate *models.Estimate) error {
	estimate.Sales1Person = nil
	estimate.Sales1PersonName = ""
	if strings.TrimSpace(estimate.Sales1PersonID) != "" {
		var sales1 models.Agent
		if err := store.First(&sales1, "id = ?", estimate.Sales1PersonID); err != nil {
			return fmt.Errorf("invalid Sales1PersonID: %w", err)
		}
		estimate.Sales1Person = &sales1
		estimate.Sales1PersonName = sales1.Name
	}

	estimate.Sales2Person = nil
	estimate.Sales2PersonName = ""
	if strings.TrimSpace(estimate.Sales2PersonID) != "" {
		var sales2 models.Agent
		if err := store.First(&sales2, "id = ?", estimate.Sales2PersonID); err != nil {
			return fmt.Errorf("invalid Sales2PersonID: %w", err)
		}
		estimate.Sales2Person = &sales2
		estimate.Sales2PersonName = sales2.Name
	}

	return nil
}

func validateEstimateSalesPersons(estimate *models.Estimate) error {
	if strings.TrimSpace(estimate.Sales1PersonID) == "" {
		return fmt.Errorf("Sales1PersonID is required")
	}

	return nil
}

// resolveInstallationSalesPersonName looks up the Agent's display name from
// InstallationSalesPersonID, the same way Sales1/Sales2 are resolved — but
// only when an ID is actually present. The installer field is now free text
// entered directly on the Job page (not picked from the agent list), so it
// never comes with an ID; in that — now the common — case this must leave
// whatever name the frontend sent untouched. It used to unconditionally
// blank the name whenever the ID was empty, which silently discarded every
// typed installer name on save (it would still show for the rest of that
// session, then vanish on the next load once the API returned this
// now-blank field back). An invalid or stale ID is still treated as
// non-fatal — the existing name is just left as-is — since this field is
// optional and shouldn't block saving the rest of the estimate.
func resolveInstallationSalesPersonName(store db.Database, details *models.EstimateDetails) {
	if details == nil {
		return
	}
	id := strings.TrimSpace(details.InstallationSalesPersonID)
	if id == "" {
		return
	}
	var agent models.Agent
	if err := store.First(&agent, "id = ?", id); err != nil {
		log.Printf("installation sales person lookup failed for ID %q: %v", id, err)
		return
	}
	details.InstallationSalesPersonName = agent.Name
}

func normalizeResidentialEntryLookups(store db.Database, entry *models.ResidentialEntry) error {
	if entry.ManagerResidentialEntryID != 0 {
		var managerEntry models.ManagerResidentialEntry
		if err := store.Where("id = ?", entry.ManagerResidentialEntryID).First(&managerEntry); err != nil {
			return fmt.Errorf("invalid ManagerResidentialEntryID: %w", err)
		}
		entry.ManagerResidentialEntryID = managerEntry.ID
	} else if entry.ManagerResidentialEntry.ID != 0 {
		var managerEntry models.ManagerResidentialEntry
		if err := store.Where("id = ?", entry.ManagerResidentialEntry.ID).First(&managerEntry); err != nil {
			return fmt.Errorf("invalid ManagerResidentialEntry.ID: %w", err)
		}
		entry.ManagerResidentialEntryID = managerEntry.ID
	} else if name := strings.TrimSpace(entry.ManagerResidentialEntry.ResidentialEntryName); name != "" {
		var managerEntry models.ManagerResidentialEntry
		if err := store.Where("LOWER(residential_entry_name) = LOWER(?)", name).First(&managerEntry); err != nil {
			return fmt.Errorf("invalid ManagerResidentialEntry name %q: %w", name, err)
		}
		entry.ManagerResidentialEntryID = managerEntry.ID
	}

	if entry.ManagerResidentialTypeID != 0 {
		var managerType models.ManagerResidentialType
		if err := store.Where("id = ?", entry.ManagerResidentialTypeID).First(&managerType); err != nil {
			return fmt.Errorf("invalid ManagerResidentialTypeID: %w", err)
		}
		entry.ManagerResidentialTypeID = managerType.ID
	} else if entry.ManagerResidentialType.ID != 0 {
		var managerType models.ManagerResidentialType
		if err := store.Where("id = ?", entry.ManagerResidentialType.ID).First(&managerType); err != nil {
			return fmt.Errorf("invalid ManagerResidentialType.ID: %w", err)
		}
		entry.ManagerResidentialTypeID = managerType.ID
	} else if entryType := strings.TrimSpace(entry.ManagerResidentialType.ResidentialEntryType); entryType != "" {
		var managerType models.ManagerResidentialType
		if err := store.Where("LOWER(residential_entry_type) = LOWER(?)", entryType).First(&managerType); err != nil {
			return fmt.Errorf("invalid ManagerResidentialType type %q: %w", entryType, err)
		}
		entry.ManagerResidentialTypeID = managerType.ID
	}

	clearResidentialEntryAssociations(entry)

	return nil
}

func clearResidentialEntryAssociations(entry *models.ResidentialEntry) {
	entry.ManagerResidentialEntry = models.ManagerResidentialEntry{}
	entry.ManagerResidentialType = models.ManagerResidentialType{}
	entry.Lead = models.Lead{}
	entry.Estimate = models.Estimate{}
}

func clearMaterialRowAssociations(row *models.MaterialRow) {
	row.Estimate = models.Estimate{}
	row.ResidentialEntry = models.ResidentialEntry{}
}

func clearServiceRowAssociations(row *models.ServiceRow) {
	row.Estimate = models.Estimate{}
	row.ResidentialEntry = models.ResidentialEntry{}
}

func residentialEntryLookupErrorResponse(err error) ErrorResponse {
	details := err.Error()
	message := "Invalid residential entry lookup"
	switch {
	case strings.Contains(details, "ManagerResidentialEntryID"),
		strings.Contains(details, "ManagerResidentialEntry.ID"),
		strings.Contains(details, "ManagerResidentialEntry name"):
		message = "Invalid ManagerResidentialEntryID"
	case strings.Contains(details, "ManagerResidentialTypeID"),
		strings.Contains(details, "ManagerResidentialType.ID"),
		strings.Contains(details, "ManagerResidentialType type"):
		message = "Invalid ManagerResidentialTypeID"
	}

	return ErrorResponse{
		Error:   message,
		Code:    "VALIDATION_ERROR",
		Details: details,
	}
}

func incrementManagerResidentialLengths(store db.Database, entries []models.ResidentialEntry) error {
	entryCounts, typeCounts := countManagerResidentialLookups(entries)

	for id, delta := range entryCounts {
		if err := applyManagerResidentialLengthDelta(
			store,
			&models.ManagerResidentialEntry{},
			id,
			delta,
			"entry",
		); err != nil {
			return fmt.Errorf("failed to increment manager residential entry length: %w", err)
		}
	}

	for id, delta := range typeCounts {
		if err := applyManagerResidentialLengthDelta(
			store,
			&models.ManagerResidentialType{},
			id,
			delta,
			"type",
		); err != nil {
			return fmt.Errorf("failed to increment manager residential type length: %w", err)
		}
	}

	return nil
}

func countManagerResidentialLookups(entries []models.ResidentialEntry) (map[uint]int, map[uint]int) {
	entryCounts := make(map[uint]int)
	typeCounts := make(map[uint]int)

	for _, entry := range entries {
		if entry.ManagerResidentialEntryID != 0 {
			entryCounts[entry.ManagerResidentialEntryID]++
		}
		if entry.ManagerResidentialTypeID != 0 {
			typeCounts[entry.ManagerResidentialTypeID]++
		}
	}

	return entryCounts, typeCounts
}

func applyManagerResidentialLengthDelta(store db.Database, model interface{}, id uint, delta int, label string) error {
	if id == 0 || delta == 0 {
		return nil
	}

	value := gorm.Expr("length + ?", delta)
	if delta < 0 {
		decrement := -delta
		value = gorm.Expr("CASE WHEN length >= ? THEN length - ? ELSE 0 END", decrement, decrement)
	}

	if err := store.Model(model).
		Where("id = ?", id).
		Update("length", value); err != nil {
		return fmt.Errorf("failed to update manager residential %s length: %w", label, err)
	}

	return nil
}

func syncManagerResidentialLengths(store db.Database, before, after []models.ResidentialEntry) error {
	beforeEntryCounts, beforeTypeCounts := countManagerResidentialLookups(before)
	afterEntryCounts, afterTypeCounts := countManagerResidentialLookups(after)

	for id, afterCount := range afterEntryCounts {
		delta := afterCount - beforeEntryCounts[id]
		if err := applyManagerResidentialLengthDelta(store, &models.ManagerResidentialEntry{}, id, delta, "entry"); err != nil {
			return err
		}
		delete(beforeEntryCounts, id)
	}
	for id, beforeCount := range beforeEntryCounts {
		if err := applyManagerResidentialLengthDelta(store, &models.ManagerResidentialEntry{}, id, -beforeCount, "entry"); err != nil {
			return err
		}
	}

	for id, afterCount := range afterTypeCounts {
		delta := afterCount - beforeTypeCounts[id]
		if err := applyManagerResidentialLengthDelta(store, &models.ManagerResidentialType{}, id, delta, "type"); err != nil {
			return err
		}
		delete(beforeTypeCounts, id)
	}
	for id, beforeCount := range beforeTypeCounts {
		if err := applyManagerResidentialLengthDelta(store, &models.ManagerResidentialType{}, id, -beforeCount, "type"); err != nil {
			return err
		}
	}

	return nil
}

func inheritEstimateSalesPersons(estimate *models.Estimate, lead *models.Lead, existing *models.Estimate) {
	if strings.TrimSpace(estimate.Sales1PersonID) == "" {
		switch {
		case lead != nil && strings.TrimSpace(lead.Sales1PersonID) != "":
			estimate.Sales1PersonID = lead.Sales1PersonID
		case existing != nil && strings.TrimSpace(existing.Sales1PersonID) != "":
			estimate.Sales1PersonID = existing.Sales1PersonID
		}
	}

	if strings.TrimSpace(estimate.Sales2PersonID) == "" {
		switch {
		case lead != nil && strings.TrimSpace(lead.Sales2PersonID) != "":
			estimate.Sales2PersonID = lead.Sales2PersonID
		case existing != nil && strings.TrimSpace(existing.Sales2PersonID) != "":
			estimate.Sales2PersonID = existing.Sales2PersonID
		}
	}
}

func (h *Handler) DuplicateEstimateJob(c *gin.Context) {
	db := middleware.GetDB(c)

	// ✅ Safety check for nil DB
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error — database not initialized",
		})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid ID",
			Code:    "VALIDATION_ERROR",
			Details: "ID parameter is required",
		})
		return
	}

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Unexpected error during duplication",
			})
		}
	}()

	// ✅ Find the original estimate with its lead
	var originalEstimate models.Estimate
	if err := tx.Preload("Lead").
		Preload("MaterialRows").
		Preload("ServiceRows").
		Preload("ResidentialEntries").
		Preload("ResidentialEntries.ManagerResidentialEntry").
		Preload("ResidentialEntries.ManagerResidentialType").
		Where("id = ?", id).First(&originalEstimate); err != nil {

		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Estimate not found",
				Code:    "NOT_FOUND",
				Details: "No estimate found with the provided ID",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to fetch estimate",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	// ✅ Remove all residential entries for this lead
	if err := tx.Model(&models.ResidentialEntry{}).
		Where("lead_id = ?", originalEstimate.Lead.ID).
		Update("lead_id", nil); err != nil {

		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to delink old residential entries",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}
	// ✅ Find existing versioned duplicates
	var existingVersions []models.Estimate
	versionPattern := id + "-v%"
	if err := tx.Where("id LIKE ?", versionPattern).Find(&existingVersions); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to check existing versions",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	// ✅ Determine next version
	nextVersion := 1
	if len(existingVersions) > 0 {
		maxVersion := 0
		for _, estimate := range existingVersions {
			if parts := strings.Split(fmt.Sprintf("%v", estimate.ID), "-v"); len(parts) == 2 {
				if v, err := strconv.Atoi(parts[1]); err == nil && v > maxVersion {
					maxVersion = v
				}
			}
		}
		nextVersion = maxVersion + 1
	}

	// ✅ Prepare duplicate
	duplicateEstimate := originalEstimate
	duplicateEstimate.ID = fmt.Sprintf("%s-v%d", id, nextVersion)
	duplicateEstimate.CreatedAt = time.Now()
	duplicateEstimate.UpdatedAt = time.Now()
	duplicateEstimate.MaterialRows = nil
	duplicateEstimate.ServiceRows = nil
	duplicateEstimate.ResidentialEntries = nil

	// ✅ Create new estimate
	if err := tx.Create(&duplicateEstimate); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to duplicate estimate",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	// ✅ Map old to new residential entries
	residentialMapping := make(map[uint]uint)
	createdResidentialEntries := make([]models.ResidentialEntry, 0, len(originalEstimate.ResidentialEntries))

	// ✅ Duplicate ResidentialEntries and their stages
	for _, resEntry := range originalEstimate.ResidentialEntries {
		oldID := resEntry.ID
		newEntry := resEntry
		newEntry.ID = 0
		newEntry.EstimateID = &duplicateEstimate.ID
		// ✅ Assign to the same lead (pointer to string)
		newEntry.LeadID = &originalEstimate.Lead.ID
		clearResidentialEntryAssociations(&newEntry)

		if err := tx.Create(&newEntry); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to create residential entry",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
			return
		}
		residentialMapping[oldID] = newEntry.ID
		createdResidentialEntries = append(createdResidentialEntries, newEntry)
		var stages []models.Stage
		if err := tx.Where("residential_entry_id = ?", oldID).Find(&stages); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to fetch stages",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
			return
		}

		// ✅ Update stages to point  new ResidentialEntry
		for _, st := range stages {
			if err := tx.Model(&models.Stage{}).
				Where("id = ?", st.ID).
				Update("residential_entry_id", newEntry.ID); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "Failed to update stage residential_entry_id",
					Code:    "DATABASE_ERROR",
					Details: err.Error(),
				})
				return
			}
		}
	}
	// if err := incrementManagerResidentialLengths(tx, createdResidentialEntries); err != nil {
	// 	tx.Rollback()
	// 	c.JSON(http.StatusInternalServerError, ErrorResponse{
	// 		Error:   "Failed to update residential entry usage",
	// 		Code:    "DATABASE_ERROR",
	// 		Details: err.Error(),
	// 	})
	// 	return
	// }

	// ✅ Duplicate MaterialRows
	for _, m := range originalEstimate.MaterialRows {
		newRow := m
		newRow.ID = 0
		newRow.EstimateID = duplicateEstimate.ID
		if newID, ok := residentialMapping[m.ResidentialEntryID]; ok {
			newRow.ResidentialEntryID = newID
		}
		if err := tx.Create(&newRow); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to create material row",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
			return
		}
	}

	// ✅ Duplicate ServiceRows
	for _, s := range originalEstimate.ServiceRows {
		newRow := s
		newRow.ID = 0
		newRow.EstimateID = duplicateEstimate.ID
		if newID, ok := residentialMapping[s.ResidentialEntryID]; ok {
			newRow.ResidentialEntryID = newID
		}
		if err := tx.Create(&newRow); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to create service row",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
			return
		}
	}

	// ✅ Reload for response
	if err := tx.Preload("Lead").
		Preload("MaterialRows").
		Preload("ServiceRows").
		Preload("ResidentialEntries").
		Preload("ResidentialEntries.ManagerResidentialEntry").
		Preload("ResidentialEntries.ManagerResidentialType").
		Where("id = ?", duplicateEstimate.ID).First(&duplicateEstimate); err != nil {
	}

	userID, userName := middleware.GetRequestUser(c)
	creatorLabel := userName
	if creatorLabel == "" {
		creatorLabel = userID
	}

	event := activity.ActivityEvent{
		EntityType: "estimate",
		EntityID:   duplicateEstimate.ID,
		Action:     "duplicated",
		UserID:     userID,
		UserName:   userName,
		Payload: map[string]any{
			"original_id": id,
			"lead_name":   duplicateEstimate.Lead.JobName,
			"total":       duplicateEstimate.Details.Total,
			"created_by":  creatorLabel,
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

	c.JSON(http.StatusCreated, gin.H{
		"message":      "Estimate duplicated successfully",
		"original_id":  id,
		"duplicate_id": duplicateEstimate.ID,
		"version":      nextVersion,
		"duplicate":    duplicateEstimate,
	})
}

func (h *Handler) CreateEstimate(c *gin.Context) {
	db := middleware.GetDB(c)

	var input models.Estimate
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid input",
			Code:    "VALIDATION_ERROR",
			Details: err.Error(),
		})
		return
	}
	resolveInstallationSalesPersonName(db, &input.Details)

	input.ID = utils.GenerateEmployeeID(db)
	// Validate County Tax
	var CountyTax models.CountyTax
	id := input.Details.TaxID
	if *id != 0 {
		if err := db.First(&CountyTax, "id = ?", id); err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, ErrorResponse{
					Error:   "County tax entry not found",
					Code:    "NOT_FOUND",
					Details: err.Error(),
				})
			} else {
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "Failed to find county tax entry",
					Code:    "DATABASE_ERROR",
					Details: err.Error(),
				})
			}
			return
		}
	}

	// Validate Discount
	// var Discount models.Discount
	// discountid := input.Details.DiscountID
	// if *discountid != 0 {
	// 	if err := db.First(&Discount, "id = ?", discountid); err != nil {
	// 		if err == gorm.ErrRecordNotFound {
	// 			c.JSON(http.StatusNotFound, ErrorResponse{
	// 				Error:   "Discount entry not found",
	// 				Code:    "NOT_FOUND",
	// 				Details: err.Error(),
	// 			})
	// 		} else {
	// 			c.JSON(http.StatusInternalServerError, ErrorResponse{
	// 				Error:   "Failed to find discount entry",
	// 				Code:    "DATABASE_ERROR",
	// 				Details: err.Error(),
	// 			})
	// 		}
	// 		return
	// 	}
	// }

	// input.Details.Discount = &Discount
	input.Details.Tax = &CountyTax

	// Validate Lead
	var lead models.Lead
	if input.LeadID != nil {
		if err := db.First(&lead, "id = ?", input.LeadID); err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusBadRequest, ErrorResponse{
					Error:   "Invalid Lead ID",
					Code:    "VALIDATION_ERROR",
					Details: "Lead not found",
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
		input.Lead = lead
	}
	inheritEstimateSalesPersons(&input, &lead, nil)
	if err := validateEstimateSalesPersons(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid salesperson data",
			Code:    "VALIDATION_ERROR",
			Details: err.Error(),
		})
		return
	}
	if err := setEstimateSalesPersons(db, &input); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Invalid salesperson ID",
				Code:    "VALIDATION_ERROR",
				Details: err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to validate salesperson",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}
	lead.DetailedLeadStatus = "ESTIMATE_SENT"
	if err := db.Save(&lead); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update lead status", "detail": err})
		return
	}
	// ONLY validate manager entries - NO database operations

	for i := range input.ResidentialEntries {
		input.ResidentialEntries[i].LeadID = input.LeadID
		// var jobid = utils.GenerateEmployeeID(db)

		// Check if residential entry already exists in database
		var existingEntry models.ResidentialEntry
		if err := db.Where("id = ?", input.ResidentialEntries[i].ID).First(&existingEntry); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Entry doesn't exist in DB, create stages
				// defaultStages := []models.Stage{
				// 	{
				// 		ID:     uuid.NewString(),
				// 		LeadID: lead.ID,
				// 		Name:   "Template",
				// 		JobID:  jobid,

				// 		State:              "In Progress", // First stage starts as In Progress
				// 		CreatedAt:          time.Now(),
				// 		Date:               func() *time.Time { t := time.Now(); return &t }(),
				// 		Description:        "Prepare and finalize the template for the project",
				// 		ResidentialEntryID: input.ResidentialEntries[i].ID,
				// 	},
				// 	{
				// 		ID:                 uuid.NewString(),
				// 		LeadID:             lead.ID,
				// 		JobID:              jobid,
				// 		Name:               "CAD",
				// 		State:              "Not Started",
				// 		CreatedAt:          time.Now(),
				// 		Description:        "Create CAD designs for the project",
				// 		ResidentialEntryID: input.ResidentialEntries[i].ID,
				// 	},
				// 	{
				// 		ID:                 uuid.NewString(),
				// 		LeadID:             lead.ID,
				// 		JobID:              jobid,
				// 		Name:               "Cutting",
				// 		State:              "Not Started",
				// 		CreatedAt:          time.Now(),
				// 		Description:        "Cut materials according to specifications",
				// 		ResidentialEntryID: input.ResidentialEntries[i].ID,
				// 	},
				// 	{
				// 		ID:                 uuid.NewString(),
				// 		LeadID:             lead.ID,
				// 		JobID:              jobid,
				// 		Name:               "Fabrication",
				// 		State:              "Not Started",
				// 		CreatedAt:          time.Now(),
				// 		Description:        "Fabricate components for the project",
				// 		ResidentialEntryID: input.ResidentialEntries[i].ID,
				// 	},
				// 	{
				// 		ID:                 uuid.NewString(),
				// 		LeadID:             lead.ID,
				// 		JobID:              jobid,
				// 		Name:               "Installation",
				// 		State:              "Not Started",
				// 		CreatedAt:          time.Now(),
				// 		Description:        "Install the project components on-site",
				// 		ResidentialEntryID: input.ResidentialEntries[i].ID,
				// 	},
				// 	{
				// 		ID:                 uuid.NewString(),
				// 		LeadID:             lead.ID,
				// 		JobID:              jobid,
				// 		Name:               "Service Call",
				// 		State:              "Not Started",
				// 		CreatedAt:          time.Now(),
				// 		Description:        "Handle any service calls or maintenance",
				// 		ResidentialEntryID: input.ResidentialEntries[i].ID,
				// 	},
				// }

				// // Create default stages ONLY for new entries
				// fmt.Println("Creating default stages for lead:", lead.ID)
				// for _, stage := range defaultStages {
				// 	if err := db.Create(&stage); err != nil {
				// 		db.Rollback()
				// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create default stages: " + err.Error()})
				// 		return
				// 	}
				// 	fmt.Println("Stage created:", stage.Name)
				// }
			} else {
				// Some other database error
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
				return
			}
		}
		// If entry exists in DB, stages are NOT created (skip stage creation)

		if err := normalizeResidentialEntryLookups(db, &input.ResidentialEntries[i]); err != nil {
			c.JSON(http.StatusBadRequest, residentialEntryLookupErrorResponse(err))
			return
		}
	}

	// Clear EstimateID for rows
	for i := range input.MaterialRows {
		input.MaterialRows[i].EstimateID = ""
	}
	for i := range input.ServiceRows {
		input.ServiceRows[i].EstimateID = ""
	}

	// Start transaction
	tx := db.Begin()

	residentialEntriesInput := input.ResidentialEntries
	materialRowsInput := input.MaterialRows
	serviceRowsInput := input.ServiceRows
	input.ResidentialEntries = nil
	input.MaterialRows = nil
	input.ServiceRows = nil

	// Create estimate
	if err := tx.Create(&input); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to create estimate",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	input.ResidentialEntries = residentialEntriesInput
	input.MaterialRows = materialRowsInput
	input.ServiceRows = serviceRowsInput
	// After inserting all residential
	// A lead has one job: reuse its job ID so a project added to a lead that is
	// already sold still belongs to that job instead of spawning a new one.
	leadJobID := ""
	if input.LeadID != nil {
		leadJobID = *input.LeadID
	}
	jobid := resolveLeadJobID(tx, leadJobID)
	if jobid == "" {
		jobid = utils.GenerateEmployeeID(db)
	}

	for i := range input.ResidentialEntries {
		var residential models.ResidentialEntry
		input.ResidentialEntries[i].EstimateID = &input.ID
		input.ResidentialEntries[i].LeadID = input.LeadID
		if err := tx.Where("id = ?", input.ResidentialEntries[i].ID).First(&residential); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Create new entry
				if err := tx.Create(&input.ResidentialEntries[i]); err != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, ErrorResponse{
						Error:   "Failed to create residential entry",
						Code:    "DATABASE_ERROR",
						Details: err.Error(),
					})
					return
				}
				// ONLY create stages when residential entry is newly created
				defaultStages := []models.Stage{
					{
						ID:     uuid.NewString(),
						LeadID: lead.ID,
						Name:   "Template",
						JobID:  jobid,

						State:              "In Progress", // First stage starts as In Progress
						CreatedAt:          time.Now(),
						Date:               func() *time.Time { t := time.Now(); return &t }(),
						Description:        "Prepare and finalize the template for the project",
						ResidentialEntryID: input.ResidentialEntries[i].ID,
					},
					{
						ID:     uuid.NewString(),
						LeadID: lead.ID,
						JobID:  jobid,

						Name:               "CAD",
						State:              "Not Started",
						CreatedAt:          time.Now(),
						Description:        "Create CAD designs for the project",
						ResidentialEntryID: input.ResidentialEntries[i].ID,
					},
					{
						ID:     uuid.NewString(),
						LeadID: lead.ID,
						JobID:  jobid,

						Name:               "Cutting",
						State:              "Not Started",
						CreatedAt:          time.Now(),
						Description:        "Cut materials according to specifications",
						ResidentialEntryID: input.ResidentialEntries[i].ID,
					},
					{
						ID:     uuid.NewString(),
						LeadID: lead.ID,
						JobID:  jobid,

						Name:               "Fabrication",
						State:              "Not Started",
						CreatedAt:          time.Now(),
						Description:        "Fabricate components for the project",
						ResidentialEntryID: input.ResidentialEntries[i].ID,
					},
					{
						ID:     uuid.NewString(),
						LeadID: lead.ID,
						JobID:  jobid,

						Name:               "Installation",
						State:              "Not Started",
						CreatedAt:          time.Now(),
						Description:        "Install the project components on-site",
						ResidentialEntryID: input.ResidentialEntries[i].ID,
					},
					{
						ID:     uuid.NewString(),
						LeadID: lead.ID,
						JobID:  jobid,

						Name:               "Service Call",
						State:              "Not Started",
						CreatedAt:          time.Now(),
						Description:        "Handle any service calls or maintenance",
						ResidentialEntryID: input.ResidentialEntries[i].ID,
					},
				}

				// Create default stages
				for _, stage := range defaultStages {
					if err := tx.Create(&stage); err != nil {
						tx.Rollback()
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create default stages: " + err.Error()})
						return
					}
				}
			} else {
				// Other database error
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "Database error checking residential entry",
					Code:    "DATABASE_ERROR",
					Details: err.Error(),
				})
				return
			}
		} else {
			// Existing entry — sync EstimateID, LeadID, and ServicesTaxExempt
			residential.EstimateID = &input.ID
			residential.LeadID = input.LeadID
			residential.ServicesTaxExempt = input.ResidentialEntries[i].ServicesTaxExempt
			residential.Details = input.ResidentialEntries[i].Details
			residential.ManagerResidentialEntryID = input.ResidentialEntries[i].ManagerResidentialEntryID
			residential.ManagerResidentialTypeID = input.ResidentialEntries[i].ManagerResidentialTypeID
			if err := tx.Save(&residential); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "Failed to update residential entry",
					Code:    "DATABASE_ERROR",
					Details: err.Error(),
				})
				return
			}
		}
	}
	// if err := incrementManagerResidentialLengths(tx, input.ResidentialEntries); err != nil {
	// 	tx.Rollback()
	// 	c.JSON(http.StatusInternalServerError, ErrorResponse{
	// 		Error:   "Failed to update residential entry usage",
	// 		Code:    "DATABASE_ERROR",
	// 		Details: err.Error(),
	// 	})
	// 	return
	// }
	// ✅ Reset sequence inside the transaction, after inserts
	if err := tx.Exec(`SELECT setval('residential_entries_id_seq', (SELECT COALESCE(MAX(id), 1) FROM residential_entries))`); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to reset sequence",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	for i := range input.MaterialRows {
		input.MaterialRows[i].EstimateID = input.ID
		clearMaterialRowAssociations(&input.MaterialRows[i])
		if input.MaterialRows[i].ResidentialEntryID != 0 {
			found := false
			for _, entry := range input.ResidentialEntries {
				if entry.ID == input.MaterialRows[i].ResidentialEntryID {
					found = true
					break
				}
			}
			if !found && len(input.ResidentialEntries) > 0 {
				input.MaterialRows[i].ResidentialEntryID = input.ResidentialEntries[0].ID
			}
		} else if len(input.ResidentialEntries) > 0 {
			input.MaterialRows[i].ResidentialEntryID = input.ResidentialEntries[0].ID
		}

		if err := tx.Save(&input.MaterialRows[i]); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to save material row",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
			return
		}
	}

	for i := range input.ServiceRows {
		row := &input.ServiceRows[i]
		row.EstimateID = input.ID
		clearServiceRowAssociations(row)

		if row.ResidentialEntryID != 0 {
			found := false
			for _, entry := range input.ResidentialEntries {
				if entry.ID == row.ResidentialEntryID {
					found = true
					break
				}
			}
			if !found && len(input.ResidentialEntries) > 0 {
				row.ResidentialEntryID = input.ResidentialEntries[0].ID
			}
		} else if len(input.ResidentialEntries) > 0 {
			row.ResidentialEntryID = input.ResidentialEntries[0].ID
		}

		if row.Details.IsGroup {
			// Create group row without nested services
			groupDetails := row.Details
			// groupDetails.Services = nil

			groupRow := models.ServiceRow{
				Details:            groupDetails,
				EstimateID:         row.EstimateID,
				ResidentialEntryID: row.ResidentialEntryID,
			}

			if err := tx.Create(&groupRow); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "Failed to save service group",
					Code:    "DATABASE_ERROR",
					Details: err.Error(),
				})
				return
			}
		} else {
			if err := tx.Save(row); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "Failed to save service row",
					Code:    "DATABASE_ERROR",
					Details: err.Error(),
				})
				return
			}
		}
	}
	input.Lead.DetailedLeadStatus = "ESTIMATE_SENT"

	userID, userName := middleware.GetRequestUser(c)
	creatorLabel := userName
	if creatorLabel == "" {
		creatorLabel = userID
	}

	event := activity.ActivityEvent{
		EntityType: "estimate",
		EntityID:   input.ID,
		Action:     "created",
		UserID:     userID,
		UserName:   userName,
		Payload: map[string]any{
			"lead_name":  lead.JobName,
			"total":      input.Details.Total,
			"created_by": creatorLabel,
		},
	}
	if err := activity.InsertActivityEvent(c.Request.Context(), tx.GetUnderlying(), event); err != nil {
		log.Printf("activity event publish failed: %v", err)
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
	c.JSON(http.StatusCreated, gin.H{
		"message": "estimate created successfully",
		"data":    input,
	})
}

// GetEstimate handles GET /estimates/:id
func (h *Handler) GetEstimate(c *gin.Context) {
	db := middleware.GetDB(c)

	id := c.Param("id")
	var estimate models.Estimate

	if err := db.Preload("Lead").
		Preload("Sales1Person").
		Preload("Sales2Person").
		Preload("MaterialRows").
		Preload("ServiceRows").
		Preload("ResidentialEntries").
		Preload("ResidentialEntries.ManagerResidentialEntry").
		Preload("ResidentialEntries.ManagerResidentialType").
		Preload("Schedule").
		First(&estimate, "id = ?", id); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Estimate not found",
				Code:    "NOT_FOUND",
				Details: "No estimate found with the given ID",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to fetch estimate",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	sortEstimateRowsBySortOrder(&estimate)

	if estimate.LeadID != nil {
		var lead models.Lead
		if err := db.Model(&models.Lead{}).
			Where("id = ?", estimate.LeadID).
			Preload("Files").
			Preload("Associates").
			First(&lead); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusBadRequest, ErrorResponse{
					Error:   "Invalid Lead ID",
					Code:    "VALIDATION_ERROR",
					Details: "Lead not found",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Database error",
				Code:    "DATABASE_ERROR",
				Details: "An error occurred while querying the database",
			})
			return
		}

		// Hydrate shipping address
		if lead.Details.ShippingAddressID != nil && *lead.Details.ShippingAddressID != 0 {
			var shippingAddr models.Address
			if err := db.First(&shippingAddr, "id = ?", *lead.Details.ShippingAddressID); err == nil {
				lead.Details.ShippingAddress = &shippingAddr
			}
		}

		// Hydrate billing address
		if lead.Details.BillingAddressID != nil && *lead.Details.BillingAddressID != 0 {
			var billingAddr models.Address
			if err := db.First(&billingAddr, "id = ?", *lead.Details.BillingAddressID); err == nil {
				lead.Details.BillingAddress = &billingAddr
			}
		}

		estimate.Lead = lead
	}
	var finalServiceRows []models.ServiceRow

	newNestedGroupIDs := make(map[uint]bool)

	for _, row := range estimate.ServiceRows {
		if row.Details.IsGroup && row.Details.Services != nil && len(row.Details.Services) > 0 {
			newNestedGroupIDs[row.ID] = true
			finalServiceRows = append(finalServiceRows, row)
		}
	}

	for _, row := range estimate.ServiceRows {
		if !row.Details.IsGroup && row.Details.ServiceGroupName == "" {
			finalServiceRows = append(finalServiceRows, row)
		}
	}

	estimate.ServiceRows = finalServiceRows

	c.JSON(http.StatusOK, gin.H{
		"message": "estimate retrieved successfully",
		"data":    estimate,
	})
}

// GetEstimates handles GET /estimates
func (h *Handler) GetEstimates(c *gin.Context) {
	db := middleware.GetDB(c)

	// Parse query parameters for pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page <= 0 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	var estimates []models.Estimate

	// Get estimates with pagination and preloaded relations
	if err := db.Preload("Lead").
		Preload("Sales1Person").
		Preload("Sales2Person").
		Preload("MaterialRows").
		Preload("ServiceRows").
		Preload("ResidentialEntries").
		Preload("Schedule").
		Offset(offset).
		Limit(limit).
		Find(&estimates); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to fetch estimates",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	// Get total count for pagination metadata
	var total int64
	if err := db.Model(&models.Estimate{}).Count(&total).Error(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to count estimates",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "estimates retrieved successfully",
		"data": gin.H{
			"estimates":  estimates,
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}
func (h *Handler) UpdateEstimate(c *gin.Context) {
	db := middleware.GetDB(c)

	id := c.Param("id")

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Fetch existing estimate with associations
	var existing models.Estimate
	if err := tx.Preload("Lead").
		Preload("MaterialRows").
		Preload("ServiceRows").
		Preload("ResidentialEntries").
		First(&existing, "id = ?", id); err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Estimate not found"})
		return
	}
	var input models.Estimate
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid input",
			Code:    "VALIDATION_ERROR",
			Details: err.Error(),
		})
		return
	}
	resolveInstallationSalesPersonName(tx, &input.Details)
	if err := validateEstimateSalesPersons(&input); err != nil {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid salesperson data",
			Code:    "VALIDATION_ERROR",
			Details: err.Error(),
		})
		return
	}
	// Updates must honor the submitted salesperson values, including clears.
	if err := setEstimateSalesPersons(tx, &input); err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Invalid salesperson ID",
				Code:    "VALIDATION_ERROR",
				Details: err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to validate salesperson",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	// ── Snapshot old values BEFORE any mutation (used for diff later) ──
	oldSales1Name := existing.Sales1PersonName
	oldSales2Name := existing.Sales2PersonName
	oldStatus := existing.Status
	oldMaterialTotal := existing.MaterialTotal
	oldServiceTotal := existing.ServiceTotal
	oldEstimatedSqFt := existing.EstimatedSqFt
	oldCorrectedSqFt := existing.CorrectedSqFt
	oldTotal := existing.Details.Total
	oldTaxAmount := existing.Details.TaxAmount
	oldDiscountAmount := existing.Details.DiscountAmount
	oldLaborCost := existing.Details.LaborCost
	// Snapshot tax/discount IDs for county-tax and discount changes
	oldTaxID := uint(0)
	if existing.Details.TaxID != nil {
		oldTaxID = *existing.Details.TaxID
	}
	// oldDiscountID := uint(0)
	// if existing.Details.DiscountID != nil {
	// 	oldDiscountID = *existing.Details.DiscountID
	// }
	// Snapshot material/service rows for diff
	oldMaterialRows := make([]models.MaterialRow, len(existing.MaterialRows))
	copy(oldMaterialRows, existing.MaterialRows)
	oldServiceRows := make([]models.ServiceRow, len(existing.ServiceRows))
	copy(oldServiceRows, existing.ServiceRows)
	oldResidentialEntries := make([]models.ResidentialEntry, len(existing.ResidentialEntries))
	copy(oldResidentialEntries, existing.ResidentialEntries)

	// Update main estimate fields
	existing.Name = input.Name
	existing.MaterialTotal = input.MaterialTotal
	existing.ServiceTotal = input.ServiceTotal
	existing.Status = input.Status
	existing.LeadID = input.LeadID
	existing.Sales1PersonID = input.Sales1PersonID
	existing.Sales1PersonName = input.Sales1PersonName
	existing.Sales1Person = input.Sales1Person
	existing.Sales2PersonID = input.Sales2PersonID
	existing.Sales2PersonName = input.Sales2PersonName
	existing.Sales2Person = input.Sales2Person
	existing.EstimatedSqFt = input.EstimatedSqFt
	existing.CorrectedSqFt = input.CorrectedSqFt

	existing.Details = input.Details

	// Save the main estimate
	if err := tx.Save(&existing); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update estimate"})
		return
	}

	// === MaterialRows ===
	// Collect input MaterialRow IDs
	inputMaterialRowIDs := make(map[uint]bool)
	for _, row := range input.MaterialRows {
		if row.ID != 0 {
			inputMaterialRowIDs[row.ID] = true
		}
	}

	// Delete MaterialRows not in input
	for _, existingRow := range existing.MaterialRows {
		if !inputMaterialRowIDs[existingRow.ID] {
			if err := tx.Delete(&models.MaterialRow{}, "id = ?", existingRow.ID); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to delete material row",
					"details": err.Error(),
				})
				return
			}
		}
	}

	// Update or create MaterialRows
	for _, row := range input.MaterialRows {
		row.EstimateID = existing.ID
		clearMaterialRowAssociations(&row)
		// Auto-assign ResidentialEntryID if needed
		if row.ResidentialEntryID == 0 && len(input.ResidentialEntries) > 0 {
			row.ResidentialEntryID = input.ResidentialEntries[0].ID
		}

		if row.ID != 0 {
			if err := tx.Save(&row); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to save material row",
					"details": err.Error(),
				})
				return
			}
		} else {
			if err := tx.Create(&row); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create material row"})
				return
			}
		}
	}

	// === ServiceRows ===
	// Collect input ServiceRow IDs
	inputServiceRowIDs := make(map[uint]bool)
	for _, row := range input.ServiceRows {
		if row.ID != 0 {
			inputServiceRowIDs[row.ID] = true
		}
	}

	// Delete ServiceRows not in input
	for _, existingRow := range existing.ServiceRows {
		if !inputServiceRowIDs[existingRow.ID] {
			if err := tx.Delete(&models.ServiceRow{}, "id = ?", existingRow.ID); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to delete service row",
					"details": err.Error(),
				})
				return
			}
		}
	}

	// Update or create ServiceRows
	for _, row := range input.ServiceRows {
		row.EstimateID = existing.ID
		clearServiceRowAssociations(&row)
		if row.ResidentialEntryID == 0 && len(input.ResidentialEntries) > 0 {
			row.ResidentialEntryID = input.ResidentialEntries[0].ID
		}

		if row.ID != 0 {
			if err := tx.Save(&row); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to save service row",
					"details": err.Error(),
				})
				return
			}
		} else {
			if err := tx.Create(&row); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create service row"})
				return
			}
		}
	}

	// === ResidentialEntries ===
	// var jobid string
	// var latestStage models.Stage

	// err := tx.Where("lead_id = ?", existing.Lead.ID).
	// 	Order("updated_at DESC"). // fetch the latest updated stage
	// 	First(&latestStage)
	// if err != nil {
	// 	if errors.Is(err, gorm.ErrRecordNotFound) {
	// 		// No stages exist, generate a new JobID
	// 		jobid = utils.GenerateEmployeeID(db)
	// 	} else {
	// 		tx.Rollback()
	// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing stages: " + err.Error()})
	// 		return
	// 	}
	// } else {
	// 	// Use the JobID from the latest stage
	// 	jobid = latestStage.JobID
	// }
	//var jobid = uuid.NewString()
	inputResidentialEntryIDs := make(map[uint]bool)
	for _, entry := range input.ResidentialEntries {
		if entry.ID != 0 {
			inputResidentialEntryIDs[entry.ID] = true
		}
	}

	for _, existingEntry := range oldResidentialEntries {
		if inputResidentialEntryIDs[existingEntry.ID] {
			continue
		}

		if err := tx.Where("residential_entry_id = ?", existingEntry.ID).Delete(&models.Stage{}); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to delete stages for removed residential entry",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
			return
		}

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

	for i := range input.ResidentialEntries {
		var residential models.ResidentialEntry
		input.ResidentialEntries[i].EstimateID = &existing.ID
		input.ResidentialEntries[i].LeadID = input.LeadID
		if err := normalizeResidentialEntryLookups(tx, &input.ResidentialEntries[i]); err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, residentialEntryLookupErrorResponse(err))
			return
		}
		err := tx.Where("id = ?", input.ResidentialEntries[i].ID).First(&residential)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Create new entry
				if err := tx.Create(&input.ResidentialEntries[i]); err != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, ErrorResponse{
						Error:   "Failed to create residential entry",
						Code:    "DATABASE_ERROR",
						Details: err.Error(),
					})
					return
				}
				// defaultStages := []models.Stage{
				// 	{
				// 		ID:                 uuid.NewString(),
				// 		LeadID:             existing.Lead.ID,
				// 		JobID:              jobid,
				// 		Name:               "Template",
				// 		State:              "In Progress", // First stage starts as In Progress
				// 		CreatedAt:          time.Now(),
				// 		Date:               func() *time.Time { t := time.Now(); return &t }(),
				// 		Description:        "Prepare and finalize the template for the project",
				// 		ResidentialEntryID: input.ResidentialEntries[i].ID,
				// 	},
				// 	{
				// 		ID:     uuid.NewString(),
				// 		LeadID: existing.Lead.ID,
				// 		JobID:  jobid,

				// 		Name:               "CAD",
				// 		State:              "Not Started",
				// 		CreatedAt:          time.Now(),
				// 		Description:        "Create CAD designs for the project",
				// 		ResidentialEntryID: input.ResidentialEntries[i].ID,
				// 	},
				// 	{
				// 		ID:     uuid.NewString(),
				// 		LeadID: existing.Lead.ID,
				// 		JobID:  jobid,

				// 		Name:               "Cutting",
				// 		State:              "Not Started",
				// 		CreatedAt:          time.Now(),
				// 		Description:        "Cut materials according to specifications",
				// 		ResidentialEntryID: input.ResidentialEntries[i].ID,
				// 	},
				// 	{
				// 		ID:     uuid.NewString(),
				// 		LeadID: existing.Lead.ID,
				// 		JobID:  jobid,

				// 		Name:               "Fabrication",
				// 		State:              "Not Started",
				// 		CreatedAt:          time.Now(),
				// 		Description:        "Fabricate components for the project",
				// 		ResidentialEntryID: input.ResidentialEntries[i].ID,
				// 	},
				// 	{
				// 		ID:     uuid.NewString(),
				// 		LeadID: existing.Lead.ID,
				// 		JobID:  jobid,

				// 		Name:               "Installation",
				// 		State:              "Not Started",
				// 		CreatedAt:          time.Now(),
				// 		Description:        "Install the project components on-site",
				// 		ResidentialEntryID: input.ResidentialEntries[i].ID,
				// 	},
				// 	{
				// 		ID:     uuid.NewString(),
				// 		LeadID: existing.Lead.ID,
				// 		JobID:  jobid,

				// 		Name:               "Service Call",
				// 		State:              "Not Started",
				// 		CreatedAt:          time.Now(),
				// 		Description:        "Handle any service calls or maintenance",
				// 		ResidentialEntryID: input.ResidentialEntries[i].ID,
				// 	},
				// }

				// // Create default stages
				// fmt.Println("Creating default stages for lead:", existing.Lead.ID)
				// for _, stage := range defaultStages {
				// 	if err := tx.Create(&stage); err != nil {
				// 		tx.Rollback()
				// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create default stages: " + err.Error()})
				// 		return
				// 	}
				// 	fmt.Println("Stage created:", stage.Name)
				// }

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
			// Entry exists — update fields from input
			residential.EstimateID = &existing.ID
			residential.LeadID = input.LeadID
			residential.ServicesTaxExempt = input.ResidentialEntries[i].ServicesTaxExempt
			residential.Details = input.ResidentialEntries[i].Details
			residential.ManagerResidentialEntryID = input.ResidentialEntries[i].ManagerResidentialEntryID
			residential.ManagerResidentialTypeID = input.ResidentialEntries[i].ManagerResidentialTypeID

			if err := tx.Save(&residential); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "Failed to update residential entry",
					Code:    "DATABASE_ERROR",
					Details: err.Error(),
				})
				return
			}
		}

	}

	if err := syncManagerResidentialLengths(tx, oldResidentialEntries, input.ResidentialEntries); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to update residential entry usage",
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

	// ── Build diff payload (uses pre-mutation snapshots) ─────────
	diffPayload := map[string]any{
		"lead_name":  existing.Lead.JobName,
		"updated_by": creatorLabel,
	}

	// Scalar fields — compare snapshots vs input
	if oldTotal != input.Details.Total {
		diffPayload["total_old"] = oldTotal
		diffPayload["total_new"] = input.Details.Total
	} else {
		diffPayload["total"] = oldTotal
	}
	if oldSales1Name != input.Sales1PersonName {
		diffPayload["sales1_old"] = oldSales1Name
		diffPayload["sales1_new"] = input.Sales1PersonName
	}
	if oldSales2Name != input.Sales2PersonName {
		diffPayload["sales2_old"] = oldSales2Name
		diffPayload["sales2_new"] = input.Sales2PersonName
	}
	if oldEstimatedSqFt != input.EstimatedSqFt {
		diffPayload["estimated_sqft_old"] = oldEstimatedSqFt
		diffPayload["estimated_sqft_new"] = input.EstimatedSqFt
	}
	if oldCorrectedSqFt != input.CorrectedSqFt {
		diffPayload["corrected_sqft_old"] = oldCorrectedSqFt
		diffPayload["corrected_sqft_new"] = input.CorrectedSqFt
	}
	if oldTaxAmount != input.Details.TaxAmount {
		diffPayload["tax_amount_old"] = oldTaxAmount
		diffPayload["tax_amount_new"] = input.Details.TaxAmount
	}
	if oldDiscountAmount != input.Details.DiscountAmount {
		diffPayload["discount_amount_old"] = oldDiscountAmount
		diffPayload["discount_amount_new"] = input.Details.DiscountAmount
	}
	if oldLaborCost != input.Details.LaborCost {
		diffPayload["labor_cost_old"] = oldLaborCost
		diffPayload["labor_cost_new"] = input.Details.LaborCost
	}
	if oldMaterialTotal != input.MaterialTotal {
		diffPayload["material_total_old"] = oldMaterialTotal
		diffPayload["material_total_new"] = input.MaterialTotal
	}
	if oldServiceTotal != input.ServiceTotal {
		diffPayload["service_total_old"] = oldServiceTotal
		diffPayload["service_total_new"] = input.ServiceTotal
	}
	if oldStatus != input.Status {
		diffPayload["status_old"] = oldStatus
		diffPayload["status_new"] = input.Status
	}
	// County tax change
	newTaxID := uint(0)
	if input.Details.TaxID != nil {
		newTaxID = *input.Details.TaxID
	}
	if oldTaxID != newTaxID {
		diffPayload["county_tax_id_old"] = oldTaxID
		diffPayload["county_tax_id_new"] = newTaxID
	}
	// Discount change
	// newDiscountID := uint(0)
	// if input.Details.DiscountID != nil {
	// 	newDiscountID = *input.Details.DiscountID
	// }
	// if oldDiscountID != newDiscountID {
	// 	diffPayload["discount_id_old"] = oldDiscountID
	// 	diffPayload["discount_id_new"] = newDiscountID
	// }

	// Material rows diff — use snapshot
	existingMatMap := make(map[uint]models.MaterialRow)
	for _, r := range oldMaterialRows {
		existingMatMap[r.ID] = r
	}
	var addedMaterials, removedMaterials, changedMaterials []map[string]any
	inputMatIDs := make(map[uint]bool)
	for _, r := range input.MaterialRows {
		inputMatIDs[r.ID] = true
		if r.ID == 0 {
			addedMaterials = append(addedMaterials, map[string]any{
				"material": r.Details.Material,
				"qty":      r.Details.Quantity,
				"price":    r.Details.UnitPrice,
			})
		} else if old, ok := existingMatMap[r.ID]; ok {
			changes := map[string]any{"material": r.Details.Material}
			changed := false
			if old.Details.Quantity != r.Details.Quantity {
				changes["qty_old"] = old.Details.Quantity
				changes["qty_new"] = r.Details.Quantity
				changed = true
			}
			if old.Details.UnitPrice != r.Details.UnitPrice {
				changes["price_old"] = old.Details.UnitPrice
				changes["price_new"] = r.Details.UnitPrice
				changed = true
			}
			if old.Details.AdditionalDiscount != r.Details.AdditionalDiscount {
				changes["discount_old"] = old.Details.AdditionalDiscount
				changes["discount_new"] = r.Details.AdditionalDiscount
				changed = true
			}
			if changed {
				changedMaterials = append(changedMaterials, changes)
			}
		}
	}
	for id, old := range existingMatMap {
		if !inputMatIDs[id] {
			removedMaterials = append(removedMaterials, map[string]any{
				"material": old.Details.Material,
				"qty":      old.Details.Quantity,
				"price":    old.Details.UnitPrice,
			})
		}
	}
	if len(addedMaterials) > 0 {
		diffPayload["materials_added"] = addedMaterials
	}
	if len(removedMaterials) > 0 {
		diffPayload["materials_removed"] = removedMaterials
	}
	if len(changedMaterials) > 0 {
		diffPayload["materials_changed"] = changedMaterials
	}

	// Service rows diff — use snapshot
	existingSvcMap := make(map[uint]models.ServiceRow)
	for _, r := range oldServiceRows {
		existingSvcMap[r.ID] = r
	}
	var addedServices, removedServices, changedServices []map[string]any
	inputSvcIDs := make(map[uint]bool)
	for _, r := range input.ServiceRows {
		inputSvcIDs[r.ID] = true
		svcName := r.Details.Name
		if svcName == "" {
			svcName = r.Details.PackageName
		}
		if svcName == "" {
			svcName = r.Details.ServiceGroupName
		}
		if r.ID == 0 {
			addedServices = append(addedServices, map[string]any{
				"service": svcName,
				"qty":     r.Details.Quantity,
				"price":   r.Details.UnitPrice,
			})
		} else if old, ok := existingSvcMap[r.ID]; ok {
			changes := map[string]any{"service": svcName}
			changed := false
			if old.Details.Quantity != r.Details.Quantity {
				changes["qty_old"] = old.Details.Quantity
				changes["qty_new"] = r.Details.Quantity
				changed = true
			}
			if old.Details.UnitPrice != r.Details.UnitPrice {
				changes["price_old"] = old.Details.UnitPrice
				changes["price_new"] = r.Details.UnitPrice
				changed = true
			}
			if old.Details.AdditionalDiscount != r.Details.AdditionalDiscount {
				changes["discount_old"] = old.Details.AdditionalDiscount
				changes["discount_new"] = r.Details.AdditionalDiscount
				changed = true
			}
			if changed {
				changedServices = append(changedServices, changes)
			}
		}
	}
	for id, old := range existingSvcMap {
		if !inputSvcIDs[id] {
			svcName := old.Details.Name
			if svcName == "" {
				svcName = old.Details.PackageName
			}
			removedServices = append(removedServices, map[string]any{
				"service": svcName,
				"qty":     old.Details.Quantity,
				"price":   old.Details.UnitPrice,
			})
		}
	}
	if len(addedServices) > 0 {
		diffPayload["services_added"] = addedServices
	}
	if len(removedServices) > 0 {
		diffPayload["services_removed"] = removedServices
	}
	if len(changedServices) > 0 {
		diffPayload["services_changed"] = changedServices
	}

	event := activity.ActivityEvent{
		EntityType: "estimate",
		EntityID:   existing.ID,
		Action:     "updated",
		UserID:     userID,
		UserName:   userName,
		Payload:    diffPayload,
	}
	if err := activity.InsertActivityEvent(c.Request.Context(), tx.GetUnderlying(), event); err != nil {
		log.Printf("activity event publish failed: %v", err)
	}

	// Final commit
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	// Return updated estimate
	var result models.Estimate
	if err := db.Preload("Lead").
		Preload("Sales1Person").
		Preload("Sales2Person").
		Preload("MaterialRows").
		Preload("ServiceRows").
		Preload("ResidentialEntries").
		First(&result, "id = ?", id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated estimate"})
		return
	}

	sortEstimateRowsBySortOrder(&result)

	c.JSON(http.StatusOK, gin.H{
		"message": "Estimate updated successfully",
		"data":    result,
	})
}
func (h *Handler) UpdateJobEstimate(c *gin.Context) {
	db := middleware.GetDB(c)

	id := c.Param("id")

	// Start a transaction
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Fetch existing estimate with associations
	var existing models.Estimate
	if err := tx.Preload("Lead").
		Preload("MaterialRows").
		Preload("ServiceRows").
		Preload("ResidentialEntries").
		First(&existing, "id = ?", id); err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Estimate not found"})
		return
	}
	var input models.Estimate
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid input",
			Code:    "VALIDATION_ERROR",
			Details: err.Error(),
		})
		return
	}
	resolveInstallationSalesPersonName(tx, &input.Details)
	if err := validateEstimateSalesPersons(&input); err != nil {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid salesperson data",
			Code:    "VALIDATION_ERROR",
			Details: err.Error(),
		})
		return
	}
	// Updates must honor the submitted salesperson values, including clears.
	if err := setEstimateSalesPersons(tx, &input); err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Invalid salesperson ID",
				Code:    "VALIDATION_ERROR",
				Details: err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to validate salesperson",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	// ── Snapshot old values BEFORE any mutation (used for diff later) ──
	oldSales1Name := existing.Sales1PersonName
	oldSales2Name := existing.Sales2PersonName
	oldStatus := existing.Status
	oldMaterialTotal := existing.MaterialTotal
	oldServiceTotal := existing.ServiceTotal
	oldEstimatedSqFt := existing.EstimatedSqFt
	oldCorrectedSqFt := existing.CorrectedSqFt
	oldTotal := existing.Details.Total
	oldTaxAmount := existing.Details.TaxAmount
	oldDiscountAmount := existing.Details.DiscountAmount
	oldLaborCost := existing.Details.LaborCost
	oldTaxID := uint(0)
	if existing.Details.TaxID != nil {
		oldTaxID = *existing.Details.TaxID
	}
	// oldDiscountID := uint(0)
	// if existing.Details.DiscountID != nil {
	// 	oldDiscountID = *existing.Details.DiscountID
	// }
	oldMaterialRows := make([]models.MaterialRow, len(existing.MaterialRows))
	copy(oldMaterialRows, existing.MaterialRows)
	oldServiceRows := make([]models.ServiceRow, len(existing.ServiceRows))
	copy(oldServiceRows, existing.ServiceRows)
	oldResidentialEntries := make([]models.ResidentialEntry, len(existing.ResidentialEntries))
	copy(oldResidentialEntries, existing.ResidentialEntries)

	// Update main estimate fields
	existing.Name = input.Name
	existing.MaterialTotal = input.MaterialTotal
	existing.ServiceTotal = input.ServiceTotal
	existing.Status = input.Status
	existing.LeadID = input.LeadID
	existing.Sales1PersonID = input.Sales1PersonID
	existing.Sales1PersonName = input.Sales1PersonName
	existing.Sales1Person = input.Sales1Person
	existing.Sales2PersonID = input.Sales2PersonID
	existing.Sales2PersonName = input.Sales2PersonName
	existing.Sales2Person = input.Sales2Person
	existing.EstimatedSqFt = input.EstimatedSqFt
	existing.CorrectedSqFt = input.CorrectedSqFt

	existing.Details = input.Details

	// Save the main estimate
	if err := tx.Save(&existing); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update estimate"})
		return
	}

	// === MaterialRows ===
	// Collect input MaterialRow IDs
	inputMaterialRowIDs := make(map[uint]bool)
	for _, row := range input.MaterialRows {
		if row.ID != 0 {
			inputMaterialRowIDs[row.ID] = true
		}
	}

	// Delete MaterialRows not in input
	for _, existingRow := range existing.MaterialRows {
		if !inputMaterialRowIDs[existingRow.ID] {
			if err := tx.Delete(&models.MaterialRow{}, "id = ?", existingRow.ID); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to delete material row",
					"details": err.Error(),
				})
				return
			}
		}
	}

	// Update or create MaterialRows
	for _, row := range input.MaterialRows {
		row.EstimateID = existing.ID
		clearMaterialRowAssociations(&row)
		// Auto-assign ResidentialEntryID if needed
		if row.ResidentialEntryID == 0 && len(input.ResidentialEntries) > 0 {
			row.ResidentialEntryID = input.ResidentialEntries[0].ID
		}

		if row.ID != 0 {
			if err := tx.Save(&row); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to save material row",
					"details": err.Error(),
				})
				return
			}
		} else {
			if err := tx.Create(&row); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create material row"})
				return
			}
		}
	}

	// === ServiceRows ===
	// Collect input ServiceRow IDs
	inputServiceRowIDs := make(map[uint]bool)
	for _, row := range input.ServiceRows {
		if row.ID != 0 {
			inputServiceRowIDs[row.ID] = true
		}
	}

	// Delete ServiceRows not in input
	for _, existingRow := range existing.ServiceRows {
		if !inputServiceRowIDs[existingRow.ID] {
			if err := tx.Delete(&models.ServiceRow{}, "id = ?", existingRow.ID); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to delete service row",
					"details": err.Error(),
				})
				return
			}
		}
	}

	// Update or create ServiceRows
	for _, row := range input.ServiceRows {
		row.EstimateID = existing.ID
		clearServiceRowAssociations(&row)
		if row.ResidentialEntryID == 0 && len(input.ResidentialEntries) > 0 {
			row.ResidentialEntryID = input.ResidentialEntries[0].ID
		}

		if row.ID != 0 {
			if err := tx.Save(&row); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to save service row",
					"details": err.Error(),
				})
				return
			}
		} else {
			if err := tx.Create(&row); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create service row"})
				return
			}
		}
	}

	// === ResidentialEntries ===
	// Every residential entry of a lead's job shares one job ID, so a project
	// added to the job after it was created still belongs to that job. Resolve
	// the lead's existing job ID (job row first, then any of its stages) and
	// only mint a new one for a lead that has no job yet.
	jobid := resolveLeadJobID(tx, existing.Lead.ID)
	if jobid == "" {
		jobid = utils.GenerateEmployeeID(db)
	}

	inputResidentialEntryIDs := make(map[uint]bool)
	for _, entry := range input.ResidentialEntries {
		if entry.ID != 0 {
			inputResidentialEntryIDs[entry.ID] = true
		}
	}

	for _, existingEntry := range oldResidentialEntries {
		if inputResidentialEntryIDs[existingEntry.ID] {
			continue
		}

		if err := tx.Where("residential_entry_id = ?", existingEntry.ID).Delete(&models.Stage{}); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to delete stages for removed residential entry",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
			return
		}

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

	for i := range input.ResidentialEntries {
		var residential models.ResidentialEntry
		input.ResidentialEntries[i].EstimateID = &existing.ID
		input.ResidentialEntries[i].LeadID = input.LeadID
		if err := normalizeResidentialEntryLookups(tx, &input.ResidentialEntries[i]); err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, residentialEntryLookupErrorResponse(err))
			return
		}
		err := tx.Where("id = ?", input.ResidentialEntries[i].ID).First(&residential)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Create new entry
				if err := tx.Create(&input.ResidentialEntries[i]); err != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, ErrorResponse{
						Error:   "Failed to create residential entry",
						Code:    "DATABASE_ERROR",
						Details: err.Error(),
					})
					return
				}
				defaultStages := []models.Stage{
					{
						ID:                 uuid.NewString(),
						LeadID:             existing.Lead.ID,
						JobID:              jobid,
						Name:               "Template",
						State:              "In Progress", // First stage starts as In Progress
						CreatedAt:          time.Now(),
						Date:               func() *time.Time { t := time.Now(); return &t }(),
						Description:        "Prepare and finalize the template for the project",
						ResidentialEntryID: input.ResidentialEntries[i].ID,
					},
					{
						ID:     uuid.NewString(),
						LeadID: existing.Lead.ID,
						JobID:  jobid,

						Name:               "CAD",
						State:              "Not Started",
						CreatedAt:          time.Now(),
						Description:        "Create CAD designs for the project",
						ResidentialEntryID: input.ResidentialEntries[i].ID,
					},
					{
						ID:     uuid.NewString(),
						LeadID: existing.Lead.ID,
						JobID:  jobid,

						Name:               "Cutting",
						State:              "Not Started",
						CreatedAt:          time.Now(),
						Description:        "Cut materials according to specifications",
						ResidentialEntryID: input.ResidentialEntries[i].ID,
					},
					{
						ID:     uuid.NewString(),
						LeadID: existing.Lead.ID,
						JobID:  jobid,

						Name:               "Fabrication",
						State:              "Not Started",
						CreatedAt:          time.Now(),
						Description:        "Fabricate components for the project",
						ResidentialEntryID: input.ResidentialEntries[i].ID,
					},
					{
						ID:     uuid.NewString(),
						LeadID: existing.Lead.ID,
						JobID:  jobid,

						Name:               "Installation",
						State:              "Not Started",
						CreatedAt:          time.Now(),
						Description:        "Install the project components on-site",
						ResidentialEntryID: input.ResidentialEntries[i].ID,
					},
					{
						ID:     uuid.NewString(),
						LeadID: existing.Lead.ID,
						JobID:  jobid,

						Name:               "Service Call",
						State:              "Not Started",
						CreatedAt:          time.Now(),
						Description:        "Handle any service calls or maintenance",
						ResidentialEntryID: input.ResidentialEntries[i].ID,
					},
				}

				for _, stage := range defaultStages {
					if err := tx.Create(&stage); err != nil {
						tx.Rollback()
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create default stages: " + err.Error()})
						return
					}
					fmt.Println("Stage created:", stage.Name)
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
			// Entry exists — update EstimateID and ServicesTaxExempt
			residential.EstimateID = &existing.ID
			residential.LeadID = input.LeadID
			residential.ServicesTaxExempt = input.ResidentialEntries[i].ServicesTaxExempt
			residential.Details = input.ResidentialEntries[i].Details
			residential.ManagerResidentialEntryID = input.ResidentialEntries[i].ManagerResidentialEntryID
			residential.ManagerResidentialTypeID = input.ResidentialEntries[i].ManagerResidentialTypeID
			if err := tx.Save(&residential); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "Failed to update residential entry",
					Code:    "DATABASE_ERROR",
					Details: err.Error(),
				})
				return
			}
		}

	}

	if err := syncManagerResidentialLengths(tx, oldResidentialEntries, input.ResidentialEntries); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to update residential entry usage",
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

	// ── Build diff payload (uses pre-mutation snapshots) ─────────
	diffPayload := map[string]any{
		"lead_name":  existing.Lead.JobName,
		"updated_by": creatorLabel,
	}

	// Scalar fields — compare snapshots vs input
	if oldTotal != input.Details.Total {
		diffPayload["total_old"] = oldTotal
		diffPayload["total_new"] = input.Details.Total
	} else {
		diffPayload["total"] = oldTotal
	}
	if oldSales1Name != input.Sales1PersonName {
		diffPayload["sales1_old"] = oldSales1Name
		diffPayload["sales1_new"] = input.Sales1PersonName
	}
	if oldSales2Name != input.Sales2PersonName {
		diffPayload["sales2_old"] = oldSales2Name
		diffPayload["sales2_new"] = input.Sales2PersonName
	}
	if oldEstimatedSqFt != input.EstimatedSqFt {
		diffPayload["estimated_sqft_old"] = oldEstimatedSqFt
		diffPayload["estimated_sqft_new"] = input.EstimatedSqFt
	}
	if oldCorrectedSqFt != input.CorrectedSqFt {
		diffPayload["corrected_sqft_old"] = oldCorrectedSqFt
		diffPayload["corrected_sqft_new"] = input.CorrectedSqFt
	}
	if oldTaxAmount != input.Details.TaxAmount {
		diffPayload["tax_amount_old"] = oldTaxAmount
		diffPayload["tax_amount_new"] = input.Details.TaxAmount
	}
	if oldDiscountAmount != input.Details.DiscountAmount {
		diffPayload["discount_amount_old"] = oldDiscountAmount
		diffPayload["discount_amount_new"] = input.Details.DiscountAmount
	}
	if oldLaborCost != input.Details.LaborCost {
		diffPayload["labor_cost_old"] = oldLaborCost
		diffPayload["labor_cost_new"] = input.Details.LaborCost
	}
	if oldMaterialTotal != input.MaterialTotal {
		diffPayload["material_total_old"] = oldMaterialTotal
		diffPayload["material_total_new"] = input.MaterialTotal
	}
	if oldServiceTotal != input.ServiceTotal {
		diffPayload["service_total_old"] = oldServiceTotal
		diffPayload["service_total_new"] = input.ServiceTotal
	}
	if oldStatus != input.Status {
		diffPayload["status_old"] = oldStatus
		diffPayload["status_new"] = input.Status
	}
	// County tax change
	newTaxID := uint(0)
	if input.Details.TaxID != nil {
		newTaxID = *input.Details.TaxID
	}
	if oldTaxID != newTaxID {
		diffPayload["county_tax_id_old"] = oldTaxID
		diffPayload["county_tax_id_new"] = newTaxID
	}
	// Discount change
	// newDiscountID := uint(0)
	// if input.Details.DiscountID != nil {
	// 	newDiscountID = *input.Details.DiscountID
	// }
	// if oldDiscountID != newDiscountID {
	// 	diffPayload["discount_id_old"] = oldDiscountID
	// 	diffPayload["discount_id_new"] = newDiscountID
	// }

	// Material rows diff — use snapshot
	existingMatMap := make(map[uint]models.MaterialRow)
	for _, r := range oldMaterialRows {
		existingMatMap[r.ID] = r
	}
	var addedMaterials, removedMaterials, changedMaterials []map[string]any
	inputMatIDs := make(map[uint]bool)
	for _, r := range input.MaterialRows {
		inputMatIDs[r.ID] = true
		if r.ID == 0 {
			addedMaterials = append(addedMaterials, map[string]any{
				"material": r.Details.Material,
				"qty":      r.Details.Quantity,
				"price":    r.Details.UnitPrice,
			})
		} else if old, ok := existingMatMap[r.ID]; ok {
			changes := map[string]any{"material": r.Details.Material}
			changed := false
			if old.Details.Quantity != r.Details.Quantity {
				changes["qty_old"] = old.Details.Quantity
				changes["qty_new"] = r.Details.Quantity
				changed = true
			}
			if old.Details.UnitPrice != r.Details.UnitPrice {
				changes["price_old"] = old.Details.UnitPrice
				changes["price_new"] = r.Details.UnitPrice
				changed = true
			}
			if old.Details.AdditionalDiscount != r.Details.AdditionalDiscount {
				changes["discount_old"] = old.Details.AdditionalDiscount
				changes["discount_new"] = r.Details.AdditionalDiscount
				changed = true
			}
			if changed {
				changedMaterials = append(changedMaterials, changes)
			}
		}
	}
	for id, old := range existingMatMap {
		if !inputMatIDs[id] {
			removedMaterials = append(removedMaterials, map[string]any{
				"material": old.Details.Material,
				"qty":      old.Details.Quantity,
				"price":    old.Details.UnitPrice,
			})
		}
	}
	if len(addedMaterials) > 0 {
		diffPayload["materials_added"] = addedMaterials
	}
	if len(removedMaterials) > 0 {
		diffPayload["materials_removed"] = removedMaterials
	}
	if len(changedMaterials) > 0 {
		diffPayload["materials_changed"] = changedMaterials
	}

	// Service rows diff — use snapshot
	existingSvcMap := make(map[uint]models.ServiceRow)
	for _, r := range oldServiceRows {
		existingSvcMap[r.ID] = r
	}
	var addedServices, removedServices, changedServices []map[string]any
	inputSvcIDs := make(map[uint]bool)
	for _, r := range input.ServiceRows {
		inputSvcIDs[r.ID] = true
		svcName := r.Details.Name
		if svcName == "" {
			svcName = r.Details.PackageName
		}
		if svcName == "" {
			svcName = r.Details.ServiceGroupName
		}
		if r.ID == 0 {
			addedServices = append(addedServices, map[string]any{
				"service": svcName,
				"qty":     r.Details.Quantity,
				"price":   r.Details.UnitPrice,
			})
		} else if old, ok := existingSvcMap[r.ID]; ok {
			changes := map[string]any{"service": svcName}
			changed := false
			if old.Details.Quantity != r.Details.Quantity {
				changes["qty_old"] = old.Details.Quantity
				changes["qty_new"] = r.Details.Quantity
				changed = true
			}
			if old.Details.UnitPrice != r.Details.UnitPrice {
				changes["price_old"] = old.Details.UnitPrice
				changes["price_new"] = r.Details.UnitPrice
				changed = true
			}
			if old.Details.AdditionalDiscount != r.Details.AdditionalDiscount {
				changes["discount_old"] = old.Details.AdditionalDiscount
				changes["discount_new"] = r.Details.AdditionalDiscount
				changed = true
			}
			if changed {
				changedServices = append(changedServices, changes)
			}
		}
	}
	for id, old := range existingSvcMap {
		if !inputSvcIDs[id] {
			svcName := old.Details.Name
			if svcName == "" {
				svcName = old.Details.PackageName
			}
			removedServices = append(removedServices, map[string]any{
				"service": svcName,
				"qty":     old.Details.Quantity,
				"price":   old.Details.UnitPrice,
			})
		}
	}
	if len(addedServices) > 0 {
		diffPayload["services_added"] = addedServices
	}
	if len(removedServices) > 0 {
		diffPayload["services_removed"] = removedServices
	}
	if len(changedServices) > 0 {
		diffPayload["services_changed"] = changedServices
	}

	event := activity.ActivityEvent{
		EntityType: "job",
		EntityID:   jobid,
		Action:     "estimate updated",
		UserID:     userID,
		UserName:   userName,
		Payload:    diffPayload,
	}
	if err := activity.InsertActivityEvent(c.Request.Context(), tx.GetUnderlying(), event); err != nil {
		log.Printf("activity event publish failed: %v", err)
	}

	// Final commit
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	// Return updated estimate
	var result models.Estimate
	if err := db.Preload("Lead").
		Preload("MaterialRows").
		Preload("ServiceRows").
		Preload("ResidentialEntries").
		First(&result, "id = ?", id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated estimate"})
		return
	}

	sortEstimateRowsBySortOrder(&result)

	c.JSON(http.StatusOK, gin.H{
		"message": "Estimate updated successfully",
		"data":    result,
	})
}
func (h *Handler) GetAllEstimates(c *gin.Context) {
	db := middleware.GetDB(c)

	var totalCount int64
	if err := db.Model(&models.Estimate{}).Count(&totalCount).Error(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to count leads",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}
	var estimates []models.Estimate
	query := db.Model(&models.Estimate{}).
		Preload("ResidentialEntries").
		Preload("ResidentialEntries.ManagerResidentialEntry").
		Preload("ResidentialEntries.ManagerResidentialType").
		Preload("MaterialRows").
		Preload("ServiceRows").
		Preload("Lead").
		Preload("Schedule").
		Preload("Sales1Person").
		Preload("Sales2Person").
		Order("created_at desc")

	if err := query.Find(&estimates); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to retrieve leads",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "estimates retrieved successfully",
		"data":    estimates,
	})
}

// GetEstimatesByLead handles GET /leads/:leadId/estimates
func (h *Handler) GetEstimatesByLead(c *gin.Context) {
	db := middleware.GetDB(c)

	leadID := c.Param("leadId")

	var estimates []models.Estimate

	if err := db.Preload("Lead").
		Preload("Sales1Person").
		Preload("Sales2Person").
		Preload("MaterialRows").
		Preload("ServiceRows").
		Preload("ResidentialEntries").
		Preload("Schedule").
		Where("lead_id = ?", leadID).
		Find(&estimates); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to fetch estimates for lead",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "estimates retrieved successfully",
		"data":    estimates,
	})
}
func (h *Handler) DuplicateEstimate(c *gin.Context) {
	db := middleware.GetDB(c)

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid ID",
			Code:    "VALIDATION_ERROR",
			Details: "ID parameter is required",
		})
		return
	}

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Unexpected error during duplication",
			})
		}
	}()

	// Find the original estimate
	var originalEstimate models.Estimate
	if err := tx.Preload("Lead").
		Preload("Sales1Person").
		Preload("Sales2Person").
		Preload("MaterialRows").
		Preload("ServiceRows").
		Preload("ResidentialEntries").
		Preload("ResidentialEntries.ManagerResidentialEntry").
		Preload("ResidentialEntries.ManagerResidentialType").Where("id = ?", id).First(&originalEstimate); err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Estimate not found",
				Code:    "NOT_FOUND",
				Details: "No estimate found with the provided ID",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to fetch estimate",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	// Find the next version number by checking existing versions
	var existingVersions []models.Estimate
	versionPattern := id + "-v%"

	if err := tx.Where("id LIKE ?", versionPattern).Find(&existingVersions); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to check existing versions",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}
	// Determine the next version number
	nextVersion := 1
	if len(existingVersions) > 0 {
		// Find the highest version number
		maxVersion := 0
		for _, estimate := range existingVersions {
			// Extract version number from ID (assuming format: "123-v2")
			idStr := fmt.Sprintf("%v", estimate.ID)
			if strings.Contains(idStr, "-v") {
				parts := strings.Split(idStr, "-v")
				if len(parts) == 2 {
					if version, err := strconv.Atoi(parts[1]); err == nil && version > maxVersion {
						maxVersion = version
					}
				}
			}
		}
		nextVersion = maxVersion + 1
	}

	// Create a duplicate with versioned ID
	duplicateEstimate := originalEstimate
	duplicateEstimate.ID = fmt.Sprintf("%s-v%d", id, nextVersion)

	// Update timestamps
	duplicateEstimate.CreatedAt = time.Now()
	duplicateEstimate.UpdatedAt = time.Now()

	// Clear the related slices - we'll create them separately
	duplicateEstimate.MaterialRows = []models.MaterialRow{}
	duplicateEstimate.ServiceRows = []models.ServiceRow{}
	duplicateEstimate.ResidentialEntries = []models.ResidentialEntry{}

	// Create mapping of old ResidentialEntry ID to new ResidentialEntry ID
	residentialMapping := make(map[uint]uint)

	// First, create new ResidentialEntries
	if len(originalEstimate.ResidentialEntries) > 0 {
		for _, residentialEntry := range originalEstimate.ResidentialEntries {
			// oldResidentialID := residentialEntry.ID
			newResidentialEntry := residentialEntry
			newResidentialEntry.ID = 0 // New ID will be generated
			newResidentialEntry.EstimateID = &duplicateEstimate.ID
			clearResidentialEntryAssociations(&newResidentialEntry)

			duplicateEstimate.ResidentialEntries = append(duplicateEstimate.ResidentialEntries, newResidentialEntry)
		}
	}

	// Save the duplicate estimate first to get new ResidentialEntry IDs
	if err := tx.Create(&duplicateEstimate); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to duplicate estimate",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	// Build the mapping after save (when new IDs are generated)
	for i, originalEntry := range originalEstimate.ResidentialEntries {
		oldID := originalEntry.ID
		newID := duplicateEstimate.ResidentialEntries[i].ID
		residentialMapping[oldID] = newID
	}
	// if err := incrementManagerResidentialLengths(tx, duplicateEstimate.ResidentialEntries); err != nil {
	// 	tx.Rollback()
	// 	c.JSON(http.StatusInternalServerError, ErrorResponse{
	// 		Error:   "Failed to update residential entry usage",
	// 		Code:    "DATABASE_ERROR",
	// 		Details: err.Error(),
	// 	})
	// 	return
	// }

	// Create MaterialRows with correct ResidentialEntry mapping
	for _, materialRow := range originalEstimate.MaterialRows {
		newMaterialRow := materialRow
		newMaterialRow.ID = 0
		newMaterialRow.EstimateID = duplicateEstimate.ID
		// Map to the correct new residential entry ID
		newMaterialRow.ResidentialEntryID = residentialMapping[materialRow.ResidentialEntryID]

		if err := tx.Create(&newMaterialRow); err != nil {
			// Log error but continue
			fmt.Printf("Failed to create material row: %v\n", err)
		}
	}

	// Create ServiceRows with correct ResidentialEntry mapping
	for _, serviceRow := range originalEstimate.ServiceRows {
		newServiceRow := serviceRow
		newServiceRow.ID = 0
		newServiceRow.EstimateID = duplicateEstimate.ID
		// Map to the correct new residential entry ID
		newServiceRow.ResidentialEntryID = residentialMapping[serviceRow.ResidentialEntryID]

		if err := tx.Create(&newServiceRow); err != nil {
			// Log error but continue
			fmt.Printf("Failed to create service row: %v\n", err)
		}
	}

	userID, userName := middleware.GetRequestUser(c)
	creatorLabel := userName
	if creatorLabel == "" {
		creatorLabel = userID
	}

	event := activity.ActivityEvent{
		EntityType: "estimate",
		EntityID:   duplicateEstimate.ID,
		Action:     "duplicated",
		UserID:     userID,
		UserName:   userName,
		Payload: map[string]any{
			"original_id": id,
			"lead_name":   duplicateEstimate.Lead.JobName,
			"total":       duplicateEstimate.Details.Total,
			"created_by":  creatorLabel,
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

	c.JSON(http.StatusCreated, gin.H{
		"message":      "Estimate duplicated successfully",
		"original_id":  id,
		"duplicate_id": duplicateEstimate.ID,
		"version":      nextVersion,
		"duplicate":    duplicateEstimate,
	})
}

// func (h *Handler) PromoteVersionToBase(c *gin.Context) {
// 	id := c.Param("id")
// 	if id == "" {
// 		c.JSON(http.StatusBadRequest, ErrorResponse{
// 			Error:   "Invalid ID",
// 			Code:    "VALIDATION_ERROR",
// 			Details: "ID parameter is required",
// 		})
// 		return
// 	}

// 	// Extract base ID and check if it's a version
// 	var baseID string
// 	var isVersion bool
// 	if strings.Contains(id, "-v") {
// 		parts := strings.Split(id, "-v")
// 		baseID = parts[0]
// 		isVersion = true
// 	} else {
// 		baseID = id
// 		isVersion = false
// 	}

// 	// Start transaction
// 	tx := db.Begin()
// 	defer func() {
// 		if r := recover(); r != nil {
// 			tx.Rollback()
// 		}
// 	}()

// 	if isVersion {
// 		// CASE 1: Promote version to base by making it the base

// 		// First verify the version exists
// 		var versionEstimate models.Estimate
// 		if err := tx.Where("id = ?", id).First(&versionEstimate); err != nil {
// 			tx.Rollback()
// 			if errors.Is(err, gorm.ErrRecordNotFound) {
// 				c.JSON(http.StatusNotFound, ErrorResponse{
// 					Error:   "Version estimate not found",
// 					Code:    "NOT_FOUND",
// 					Details: "The specified version estimate does not exist",
// 				})
// 				return
// 			}
// 			c.JSON(http.StatusInternalServerError, ErrorResponse{
// 				Error:   "Failed to fetch version estimate",
// 				Code:    "DATABASE_ERROR",
// 				Details: err.Error(),
// 			})
// 			return
// 		}

// 		// Delete the current base estimate and its associated records
// 		// Delete base MaterialRows
// 		if err := tx.Where("estimate_id = ?", baseID).Delete(&models.MaterialRow{}); err != nil {
// 			tx.Rollback()
// 			c.JSON(http.StatusInternalServerError, ErrorResponse{
// 				Error:   "Failed to delete base material rows",
// 				Code:    "DATABASE_ERROR",
// 				Details: err.Error(),
// 			})
// 			return
// 		}

// 		// Delete base ServiceRows
// 		if err := tx.Where("estimate_id = ?", baseID).Delete(&models.ServiceRow{}); err != nil {
// 			tx.Rollback()
// 			c.JSON(http.StatusInternalServerError, ErrorResponse{
// 				Error:   "Failed to delete base service rows",
// 				Code:    "DATABASE_ERROR",
// 				Details: err.Error(),
// 			})
// 			return
// 		}

// 		// Delete base ResidentialEntries
// 		if err := tx.Where("estimate_id = ?", baseID).Delete(&models.ResidentialEntry{}); err != nil {
// 			tx.Rollback()
// 			c.JSON(http.StatusInternalServerError, ErrorResponse{
// 				Error:   "Failed to delete base residential entries",
// 				Code:    "DATABASE_ERROR",
// 				Details: err.Error(),
// 			})
// 			return
// 		}

// 		// Delete base estimate
// 		if err := tx.Unscoped().Where("id = ?", baseID).Delete(&models.Estimate{}); err != nil {
// 			tx.Rollback()
// 			c.JSON(http.StatusInternalServerError, ErrorResponse{
// 				Error:   "Failed to delete base estimate",
// 				Code:    "DATABASE_ERROR",
// 				Details: err.Error(),
// 			})
// 			return
// 		}

// 		// Update version estimate ID to base ID
// 		if err := tx.Model(&models.Estimate{}).Where("id = ?", id).Update("id", baseID); err != nil {
// 			tx.Rollback()
// 			c.JSON(http.StatusInternalServerError, ErrorResponse{
// 				Error:   "Failed to update version estimate ID to base",
// 				Code:    "DATABASE_ERROR",
// 				Details: err.Error(),
// 			})
// 			return
// 		}

// 		// Update MaterialRows to point to base ID
// 		if err := tx.Model(&models.MaterialRow{}).Where("estimate_id = ?", id).Update("estimate_id", baseID); err != nil {
// 			tx.Rollback()
// 			c.JSON(http.StatusInternalServerError, ErrorResponse{
// 				Error:   "Failed to update material rows estimate_id",
// 				Code:    "DATABASE_ERROR",
// 				Details: err.Error(),
// 			})
// 			return
// 		}

// 		// Update ServiceRows to point to base ID
// 		if err := tx.Model(&models.ServiceRow{}).Where("estimate_id = ?", id).Update("estimate_id", baseID); err != nil {
// 			tx.Rollback()
// 			c.JSON(http.StatusInternalServerError, ErrorResponse{
// 				Error:   "Failed to update service rows estimate_id",
// 				Code:    "DATABASE_ERROR",
// 				Details: err.Error(),
// 			})
// 			return
// 		}

// 		// Update ResidentialEntries to point to base ID
// 		if err := tx.Model(&models.ResidentialEntry{}).Where("estimate_id = ?", id).Update("estimate_id", baseID); err != nil {
// 			tx.Rollback()
// 			c.JSON(http.StatusInternalServerError, ErrorResponse{
// 				Error:   "Failed to update residential entries estimate_id",
// 				Code:    "DATABASE_ERROR",
// 				Details: err.Error(),
// 			})
// 			return
// 		}
// 	}

// 	// Delete all remaining versions (for both cases)
// 	versionPattern := baseID + "-v%"

// 	// Delete associated records from all remaining versions
// 	if err := tx.Where("estimate_id LIKE ? AND estimate_id != ?", versionPattern, baseID).Delete(&models.MaterialRow{}); err != nil {
// 		tx.Rollback()
// 		c.JSON(http.StatusInternalServerError, ErrorResponse{
// 			Error:   "Failed to delete remaining version material rows",
// 			Code:    "DATABASE_ERROR",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	if err := tx.Where("estimate_id LIKE ? AND estimate_id != ?", versionPattern, baseID).Delete(&models.ServiceRow{}); err != nil {
// 		tx.Rollback()
// 		c.JSON(http.StatusInternalServerError, ErrorResponse{
// 			Error:   "Failed to delete remaining version service rows",
// 			Code:    "DATABASE_ERROR",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	if err := tx.Where("estimate_id LIKE ? AND estimate_id != ?", versionPattern, baseID).Delete(&models.ResidentialEntry{}); err != nil {
// 		tx.Rollback()
// 		c.JSON(http.StatusInternalServerError, ErrorResponse{
// 			Error:   "Failed to delete remaining version residential entries",
// 			Code:    "DATABASE_ERROR",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	// Delete all remaining versioned estimates (hard delete)
// 	if err := tx.Unscoped().
// 		Where("id LIKE ? AND id != ?", versionPattern, baseID).
// 		Delete(&models.Estimate{}); err != nil {
// 		tx.Rollback()
// 		c.JSON(http.StatusInternalServerError, ErrorResponse{
// 			Error:   "Failed to delete remaining versioned estimates",
// 			Code:    "DATABASE_ERROR",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	// Commit transaction
// 	if err := tx.Commit(); err != nil {
// 		c.JSON(http.StatusInternalServerError, ErrorResponse{
// 			Error:   "Transaction failed",
// 			Code:    "TRANSACTION_ERROR",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

//		// Return appropriate success response
//		if isVersion {
//			c.JSON(http.StatusOK, gin.H{
//				"message":     "Version promoted to base and all other versions deleted",
//				"promoted_id": baseID,
//			})
//		} else {
//			c.JSON(http.StatusOK, gin.H{
//				"message": "All versioned estimates deleted, base unchanged",
//				"base_id": baseID,
//			})
//		}
//	}
func (h *Handler) PromoteVersionToBase(c *gin.Context) {
	db := middleware.GetDB(c)

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid ID",
			Code:    "VALIDATION_ERROR",
			Details: "ID parameter is required",
		})
		return
	}
	var baseID string
	var isVersion bool
	if strings.Contains(id, "-v") {
		parts := strings.Split(id, "-v")
		baseID = parts[0]
		isVersion = true
	} else {
		baseID = id
		isVersion = false
	}

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if isVersion {
		var versionEstimate models.Estimate
		if err := tx.Where("id = ?", id).First(&versionEstimate); err != nil {
			tx.Rollback()
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, ErrorResponse{
					Error:   "Version estimate not found",
					Code:    "NOT_FOUND",
					Details: "The specified version estimate does not exist",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to fetch version estimate",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
			return
		}
		if err := tx.Where("estimate_id = ?", baseID).Delete(&models.MaterialRow{}); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to delete base material rows",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
			return
		}
		if err := tx.Where("estimate_id = ?", baseID).Delete(&models.ServiceRow{}); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to delete base service rows",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
			return
		}
		if err := tx.Where("estimate_id = ?", baseID).Delete(&models.ResidentialEntry{}); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to delete base residential entries",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
			return
		}
		if err := tx.Unscoped().Where("id = ?", baseID).Delete(&models.Estimate{}); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to delete base estimate",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
			return
		}
		if err := tx.Model(&models.Estimate{}).Where("id = ?", id).Update("id", baseID); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to update version estimate ID to base",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
			return
		}
		if err := tx.Model(&models.MaterialRow{}).Where("estimate_id = ?", id).Update("estimate_id", baseID); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to update material rows estimate_id",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
			return
		}
		if err := tx.Model(&models.ServiceRow{}).Where("estimate_id = ?", id).Update("estimate_id", baseID); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to update service rows estimate_id",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
			return
		}
		var residentialEntries []models.ResidentialEntry
		if err := tx.Where("estimate_id = ?", id).Find(&residentialEntries); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to fetch residential entries",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
			return
		}
		// var jobid = utils.GenerateEmployeeID(db)
		for _, residentialEntry := range residentialEntries {
			if err := tx.Model(&models.ResidentialEntry{}).Where("id = ?", residentialEntry.ID).Update("estimate_id", baseID); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "Failed to update residential entry estimate_id",
					Code:    "DATABASE_ERROR",
					Details: err.Error(),
				})
				return
			}

			// defaultStages := []models.Stage{
			// 	{
			// 		ID:                 uuid.NewString(),
			// 		LeadID:             residentialEntry.Lead.ID,
			// 		JobID:              jobid,
			// 		Name:               "Template",
			// 		State:              "In Progress",
			// 		CreatedAt:          time.Now(),
			// 		Date:               func() *time.Time { t := time.Now(); return &t }(),
			// 		Description:        "Prepare and finalize the template for the project",
			// 		ResidentialEntryID: residentialEntry.ID,
			// 	},
			// 	{
			// 		ID:                 uuid.NewString(),
			// 		LeadID:             residentialEntry.Lead.ID,
			// 		JobID:              jobid,
			// 		Name:               "CAD",
			// 		State:              "Not Started",
			// 		CreatedAt:          time.Now(),
			// 		Description:        "Create CAD designs for the project",
			// 		ResidentialEntryID: residentialEntry.ID,
			// 	},
			// 	{
			// 		ID:                 uuid.NewString(),
			// 		LeadID:             residentialEntry.Lead.ID,
			// 		JobID:              jobid,
			// 		Name:               "Cutting",
			// 		State:              "Not Started",
			// 		CreatedAt:          time.Now(),
			// 		Description:        "Cut materials according to specifications",
			// 		ResidentialEntryID: residentialEntry.ID,
			// 	},
			// 	{
			// 		ID:                 uuid.NewString(),
			// 		LeadID:             residentialEntry.Lead.ID,
			// 		JobID:              jobid,
			// 		Name:               "Fabrication",
			// 		State:              "Not Started",
			// 		CreatedAt:          time.Now(),
			// 		Description:        "Fabricate components for the project",
			// 		ResidentialEntryID: residentialEntry.ID,
			// 	},
			// 	{
			// 		ID:                 uuid.NewString(),
			// 		LeadID:             residentialEntry.Lead.ID,
			// 		JobID:              jobid,
			// 		Name:               "Installation",
			// 		State:              "Not Started",
			// 		CreatedAt:          time.Now(),
			// 		Description:        "Install the project components on-site",
			// 		ResidentialEntryID: residentialEntry.ID,
			// 	},
			// 	{
			// 		ID:                 uuid.NewString(),
			// 		LeadID:             residentialEntry.Lead.ID,
			// 		JobID:              jobid,
			// 		Name:               "Service Call",
			// 		State:              "Not Started",
			// 		CreatedAt:          time.Now(),
			// 		Description:        "Handle any service calls or maintenance",
			// 		ResidentialEntryID: residentialEntry.ID,
			// 	},
			// }
			// for _, stage := range defaultStages {
			// 	if err := tx.Create(&stage); err != nil {
			// 		tx.Rollback()
			// 		c.JSON(http.StatusInternalServerError, ErrorResponse{
			// 			Error:   "Failed to create default stages",
			// 			Code:    "DATABASE_ERROR",
			// 			Details: err.Error(),
			// 		})
			// 		return
			// 	}
			// }
		}
	}

	userID, userName := middleware.GetRequestUser(c)
	creatorLabel := userName
	if creatorLabel == "" {
		creatorLabel = userID
	}

	event := activity.ActivityEvent{
		EntityType: "estimate",
		EntityID:   baseID,
		Action:     "version_promoted",
		UserID:     userID,
		UserName:   userName,
		Payload: map[string]any{
			"original_version_id": id,
			"promoted_by":         creatorLabel,
		},
	}
	if err := activity.InsertActivityEvent(c.Request.Context(), tx.GetUnderlying(), event); err != nil {
		log.Printf("activity event publish failed: %v", err)
	}

	versionPattern := baseID + "-v%"
	if err := tx.Where("estimate_id LIKE ? AND estimate_id != ?", versionPattern, baseID).Delete(&models.MaterialRow{}); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to delete remaining version material rows",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	if err := tx.Where("estimate_id LIKE ? AND estimate_id != ?", versionPattern, baseID).Delete(&models.ServiceRow{}); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to delete remaining version service rows",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	if err := tx.Where("estimate_id LIKE ? AND estimate_id != ?", versionPattern, baseID).Delete(&models.ResidentialEntry{}); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to delete remaining version residential entries",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	// Delete all remaining versioned estimates (hard delete)
	if err := tx.Unscoped().
		Where("id LIKE ? AND id != ?", versionPattern, baseID).
		Delete(&models.Estimate{}); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to delete remaining versioned estimates",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Transaction failed",
			Code:    "TRANSACTION_ERROR",
			Details: err.Error(),
		})
		return
	}

	// Return appropriate success response
	if isVersion {
		c.JSON(http.StatusOK, gin.H{
			"message":     "Version promoted to base and all other versions deleted",
			"promoted_id": baseID,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message": "All versioned estimates deleted, base unchanged",
			"base_id": baseID,
		})
	}
}

// GetEstimatesByStatus handles GET /estimates/status/:status
func (h *Handler) GetEstimatesByStatus(c *gin.Context) {
	db := middleware.GetDB(c)

	status := c.Param("status")

	var estimates []models.Estimate

	if err := db.Preload("Lead").
		Preload("Sales1Person").
		Preload("Sales2Person").
		Preload("MaterialRows").
		Preload("ServiceRows").
		Preload("ResidentialEntries").
		Where("status = ?", status).
		Find(&estimates); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to fetch estimates by status",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "estimates retrieved successfully",
		"data":    estimates,
	})
}

func (h *Handler) GetEstimateVersions(c *gin.Context) {
	db := middleware.GetDB(c)

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid ID",
			Code:    "VALIDATION_ERROR",
			Details: "ID parameter is required",
		})
		return
	}

	// Find all versions for this estimate
	var versions []models.Estimate
	versionPattern := id + "-v%"

	if err := db.Where("id LIKE ?", versionPattern).Select("id").Find(&versions); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to fetch versions",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	// Extract version IDs
	var versionIds []string
	for _, version := range versions {
		versionIds = append(versionIds, version.ID)
	}

	c.JSON(http.StatusOK, gin.H{
		"original_id": id,
		"versions":    versionIds,
		"count":       len(versionIds),
	})
}

// resolveLeadJobID returns the job ID a lead's stages must all carry: the
// lead's job row, else any stage the lead already owns. A lead has one job, so
// a project added to the job later still belongs to that same job. Returns ""
// when the lead has no job yet, in which case the caller mints one.
func resolveLeadJobID(tx db.Database, leadID string) string {
	if leadID == "" || tx == nil {
		return ""
	}
	var job models.Job
	if err := tx.Where("lead_id = ? AND job_id != ''", leadID).First(&job); err == nil {
		return job.JobID
	}
	var stage models.Stage
	if err := tx.Where("lead_id = ? AND job_id != ''", leadID).First(&stage); err == nil {
		return stage.JobID
	}
	return ""
}

func (h *Handler) CreateStagesForEstimate(c *gin.Context) {
	db := middleware.GetDB(c)
	estimateID := c.Param("id")
	if estimateID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid estimate ID",
			Code:    "VALIDATION_ERROR",
			Details: "Estimate ID is required",
		})
		return
	}

	// Fetch the estimate with residential entries
	var estimate models.Estimate
	if err := db.Preload("ResidentialEntries").
		Preload("Lead").
		Where("id = ?", estimateID).
		First(&estimate); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Estimate not found",
				Code:    "NOT_FOUND",
				Details: "No estimate found with the provided ID",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to fetch estimate",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	// Validate that estimate has residential entries
	if len(estimate.ResidentialEntries) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "No residential entries found",
			Code:    "VALIDATION_ERROR",
			Details: "Cannot create stages without residential entries",
		})
		return
	}

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Define default stages
	stageTemplates := []struct {
		Name        string
		State       string
		Description string
	}{
		{
			Name:        "Template",
			State:       "In Progress",
			Description: "Prepare and finalize the template for the project",
		},
		{
			Name:        "CAD",
			State:       "Not Started",
			Description: "Create CAD designs for the project",
		},
		{
			Name:        "Cutting",
			State:       "Not Started",
			Description: "Cut materials according to specifications",
		},
		{
			Name:        "Fabrication",
			State:       "Not Started",
			Description: "Fabricate components for the project",
		},
		{
			Name:        "Installation",
			State:       "Not Started",
			Description: "Install the project components on-site",
		},
		{
			Name:        "Service Call",
			State:       "Not Started",
			Description: "Handle any service calls or maintenance",
		},
	}

	createdStagesCount := 0

	var allExistingStagesForEstimate []models.Stage
	entryIDs := make([]uint, 0, len(estimate.ResidentialEntries))
	for _, entry := range estimate.ResidentialEntries {
		entryIDs = append(entryIDs, entry.ID)
	}
	if len(entryIDs) > 0 {
		if err := tx.Where("residential_entry_id IN ?", entryIDs).Find(&allExistingStagesForEstimate); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to check existing stages",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
			return
		}
	}

	canonicalJobID := resolveLeadJobID(tx, estimate.Lead.ID)
	if canonicalJobID == "" {
		for _, s := range allExistingStagesForEstimate {
			if s.JobID != "" {
				canonicalJobID = s.JobID
				break
			}
		}
	}
	if canonicalJobID == "" {
		canonicalJobID = utils.GenerateEmployeeID(db)
	}

	for _, s := range allExistingStagesForEstimate {
		if s.JobID != canonicalJobID {
			if err := tx.Model(&models.Stage{}).Where("id = ?", s.ID).Update("job_id", canonicalJobID); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "Failed to unify job ID across residential entries",
					Code:    "DATABASE_ERROR",
					Details: err.Error(),
				})
				return
			}
		}
	}

	for _, entry := range estimate.ResidentialEntries {
		existingStageByName := make(map[string]models.Stage)
		for _, s := range allExistingStagesForEstimate {
			if s.ResidentialEntryID == entry.ID {
				existingStageByName[s.Name] = s
			}
		}

		var templateStage *models.Stage
		for _, stageTemplate := range stageTemplates {
			if existing, ok := existingStageByName[stageTemplate.Name]; ok {
				fmt.Printf("Stage %s already exists for residential entry ID: %d, skipping create...\n", stageTemplate.Name, entry.ID)
				if stageTemplate.Name == "Template" {
					s := existing
					s.JobID = canonicalJobID
					templateStage = &s
				}
				continue
			}

			stage := models.Stage{
				ID:                 uuid.NewString(),
				LeadID:             *estimate.LeadID,
				JobID:              canonicalJobID,
				Name:               stageTemplate.Name,
				State:              stageTemplate.State,
				Description:        stageTemplate.Description,
				Date:               func() *time.Time { t := time.Now(); return &t }(),
				ResidentialEntryID: entry.ID,
				CreatedAt:          time.Now(),
			}

			if err := tx.Create(&stage); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "Failed to create stage",
					Code:    "DATABASE_ERROR",
					Details: err.Error(),
				})
				return
			}

			if stageTemplate.Name == "Template" {
				templateStage = &stage
			}

			fmt.Printf("Created stage: %s for residential entry ID: %d\n", stage.Name, entry.ID)
			createdStagesCount++
		}

		// Carry over Estimate Project Management's Template-stage tasks and
		// user assignments onto the Job tracker's own tables. See
		// migrateEstimateProjectDataToJobStage for details — this is a
		// one-time, idempotent copy.
		if templateStage != nil {
			if err := migrateEstimateProjectDataToJobStage(tx, entry.ID, *templateStage); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "Failed to migrate estimate scheduling data to job",
					Code:    "DATABASE_ERROR",
					Details: err.Error(),
				})
				return
			}
		}
	}

	userID, userName := middleware.GetRequestUser(c)
	creatorLabel := userName
	if creatorLabel == "" {
		creatorLabel = userID
	}

	event := activity.ActivityEvent{
		EntityType: "estimate",
		EntityID:   estimateID,
		Action:     "stages_created",
		UserID:     userID,
		UserName:   userName,
		Payload: map[string]any{
			"job_id":       canonicalJobID,
			"stages_count": createdStagesCount,
			"created_by":   creatorLabel,
		},
	}
	if err := activity.InsertActivityEvent(c.Request.Context(), tx.GetUnderlying(), event); err != nil {
		log.Printf("activity event publish failed: %v", err)
	}

	// Log the job creation event
	jobCreatedEvent := activity.ActivityEvent{
		EntityType: "job",
		EntityID:   canonicalJobID,
		Action:     "created",
		UserID:     userID,
		UserName:   userName,
		Payload: map[string]any{
			"job_name":    estimate.Lead.JobName,
			"estimate_id": estimateID,
			"created_by":  creatorLabel,
		},
	}
	if err := activity.InsertActivityEvent(c.Request.Context(), tx.GetUnderlying(), jobCreatedEvent); err != nil {
		log.Printf("job created activity event failed: %v", err)
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

	c.JSON(http.StatusCreated, gin.H{
		"message":              "Stages created successfully",
		"estimate_id":          estimateID,
		"job_id":               canonicalJobID,
		"stages_created_count": createdStagesCount,
		"residential_entries":  len(estimate.ResidentialEntries),
	})
}

func (h *Handler) GetEstimateActivity(c *gin.Context) {
	db := middleware.GetDB(c)
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid ID",
			Code:    "VALIDATION_ERROR",
			Details: "ID parameter is required",
		})
		return
	}

	// Derive base ID — strip version suffix if present
	baseID := id
	if strings.Contains(id, "-v") {
		parts := strings.Split(id, "-v")
		baseID = parts[0]
	}

	// Match base ID and all versioned IDs (e.g. ABC, ABC-v1, ABC-v2 ...)
	versionPattern := baseID + "-v%"

	type ActivityEventRow struct {
		ID         uint      `json:"id"`
		EntityType string    `json:"entity_type"`
		EntityID   string    `json:"entity_id"`
		Action     string    `json:"action"`
		UserID     string    `json:"user_id"`
		UserName   string    `json:"user_name"`
		Payload    []byte    `json:"-"`
		CreatedAt  time.Time `json:"created_at"`
	}

	type ActivityEventResponse struct {
		ID         uint           `json:"id"`
		EntityType string         `json:"entity_type"`
		EntityID   string         `json:"entity_id"`
		Action     string         `json:"action"`
		UserID     string         `json:"user_id"`
		UserName   string         `json:"user_name"`
		Payload    map[string]any `json:"payload"`
		CreatedAt  time.Time      `json:"created_at"`
	}

	var rows []ActivityEventRow
	if err := db.GetUnderlying().
		Table("activity_events").
		Select("id, entity_type, entity_id, action, user_id, user_name, payload, created_at").
		Where("entity_type = ? AND (entity_id = ? OR entity_id LIKE ?)", "estimate", baseID, versionPattern).
		Order("created_at DESC").
		Scan(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to fetch activity events",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	events := make([]ActivityEventResponse, 0, len(rows))
	for _, row := range rows {
		var payload map[string]any
		if len(row.Payload) > 0 {
			if err := json.Unmarshal(row.Payload, &payload); err != nil {
				payload = nil
			}
		}
		events = append(events, ActivityEventResponse{
			ID:         row.ID,
			EntityType: row.EntityType,
			EntityID:   row.EntityID,
			Action:     row.Action,
			UserID:     row.UserID,
			UserName:   row.UserName,
			Payload:    payload,
			CreatedAt:  row.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Activity events retrieved successfully",
		"data":    events,
	})
}

func (h *Handler) AddEstimateNote(c *gin.Context) {
	db := middleware.GetDB(c)
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid estimate ID",
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

	var estimate models.Estimate
	if err := tx.Preload("Schedule").Preload("Lead").First(&estimate, "id = ?", id); err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Estimate not found",
				Code:    "NOT_FOUND",
				Details: "No estimate found with the specified ID",
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
	estimate.Details.NotesHistory = append(estimate.Details.NotesHistory, note)
	if payload.CompleteFollowUp {
		completedDate := strings.TrimSpace(payload.CompletedDate)
		if completedDate == "" && estimate.Schedule != nil && estimate.Schedule.ScheduledDate != nil {
			completedDate = estimate.Schedule.ScheduledDate.Format("2006-01-02")
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
			estimate.Details.FollowUpCompletedDate = completedDate
		}
	}

	userID, userName := middleware.GetRequestUser(c)
	event := activity.ActivityEvent{
		EntityType: "estimate",
		EntityID:   estimate.ID,
		Action:     "note_added",
		UserID:     userID,
		UserName:   userName,
		Payload: map[string]any{
			"estimate_name":   estimate.Name,
			"lead_name":       estimate.Lead.JobName,
			"note_text":       note.Text,
			"note_created_at": note.CreatedAt,
			"updated_by":      userName,
		},
	}
	if err := activity.InsertActivityEvent(c.Request.Context(), tx.GetUnderlying(), event); err != nil {
		log.Printf("activity event publish failed: %v", err)
	}

	if err := tx.Save(&estimate); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to save estimate note",
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
		"message":               "Estimate note added successfully",
		"note":                  note,
		"notesHistory":          estimate.Details.NotesHistory,
		"followUpCompletedDate": estimate.Details.FollowUpCompletedDate,
	})
}

func (h *Handler) UpdateEstimateSchedule(c *gin.Context) {
	db := middleware.GetDB(c)
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid estimate ID",
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

	scheduleType := models.ScheduleType(payload.ScheduleType)
	if scheduleType != models.ScheduleToday && scheduleType != models.ScheduleTomorrow && scheduleType != models.ScheduleCustom {
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
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Custom date required",
			Code:    "VALIDATION_ERROR",
			Details: "scheduledDate is required for Custom schedule type",
		})
		return
	}

	var existingEstimate models.Estimate
	if err := db.First(&existingEstimate, "id = ?", id); err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Estimate not found",
			Code:    "NOT_FOUND",
			Details: "No estimate found with the specified ID",
		})
		return
	}

	var estimateSchedule models.EstimateSchedule
	if err := db.Where("estimate_id = ?", id).First(&estimateSchedule); err != nil {
		estimateSchedule = models.EstimateSchedule{ID: uuid.NewString(), EstimateID: id, ScheduleType: scheduleType, ScheduledDate: scheduleDate}
		db.Create(&estimateSchedule)
	} else {
		estimateSchedule.ScheduleType = scheduleType
		estimateSchedule.ScheduledDate = scheduleDate
		db.Save(&estimateSchedule)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Estimate schedule saved successfully",
		"data":    estimateSchedule,
	})
}
