package handlers

import (
	"errors"
	"inventory/db"
	"inventory/middleware"
	"inventory/models"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ════════════════════════════════════════════════════════════════════════
// Shared helpers
// ════════════════════════════════════════════════════════════════════════

func joinTeam(team []string) string {
	cleaned := make([]string, 0, len(team))
	for _, m := range team {
		m = strings.TrimSpace(m)
		if m != "" {
			cleaned = append(cleaned, m)
		}
	}
	return strings.Join(cleaned, ", ")
}

type auditFilterBody struct {
	Search             string   `json:"search"`
	Category           []string `json:"category"`
	Type               string   `json:"type"`
	Group              string   `json:"group"`
	Bin                string   `json:"bin"`
	BinLike            string   `json:"bin_like"`
	SlabFilter         string   `json:"slab_filter"` // all, full, remnants, consignment, full_without_consignment
	ExcludeScanned     bool     `json:"exclude_scanned"`
	ExcludeConsignment bool     `json:"exclude_consignment"`
	ActiveOnly         bool     `json:"active"`
	InactiveOnly       bool     `json:"inactive"`
	ChunkID            *uint    `json:"chunk_id"`    // restricts to a single materialchunks row (Done/Not Found)
	MarkStatus         string   `json:"mark_status"` // "reconciled" (default) or "missing"
}

func buildCandidateQuery(database db.Database, sessionType string, f auditFilterBody) db.Database {
	d := database.Model(&models.Materialchunks{}).
		Joins("JOIN materials ON materials.id = materialchunks.material_id")

	if sessionType == "slab" {
		d = d.Where("materials.type ILIKE ?", "%slab%")
	} else {
		d = d.Where("(materials.type IS NULL OR materials.type NOT ILIKE ?)", "%slab%")
	}
	if f.Search != "" {
		like := "%" + f.Search + "%"
		d = d.Where("(materials.name ILIKE ? OR materialchunks.serialnum ILIKE ?)", like, like)
	}
	if len(f.Category) > 0 {
		d = d.Where("materials.category IN ?", f.Category)
	}
	if f.Type != "" {
		d = d.Where("materials.type = ?", f.Type)
	}
	if f.Group != "" {
		d = d.Where("materials.sub_category = ?", f.Group)
	}
	if f.Bin != "" {
		d = d.Where("materialchunks.bin = ?", f.Bin)
	}
	if f.ChunkID != nil {
		d = d.Where("materialchunks.id = ?", *f.ChunkID)
	}
	if f.BinLike != "" {
		d = d.Where("materialchunks.bin ILIKE ?", "%"+f.BinLike+"%")
	}
	switch f.SlabFilter {
	case "full":
		d = d.Where("materialchunks.serialnum NOT LIKE ?", "%-r%")
	case "remnants":
		d = d.Where("materialchunks.serialnum LIKE ?", "%-r%")
	case "consignment":
		d = d.Where("materialchunks.inventory_type = ?", "Consignment Inventory")
	case "full_without_consignment":
		d = d.Where("materialchunks.serialnum NOT LIKE ?", "%-r%").
			Where("(materialchunks.inventory_type IS NULL OR materialchunks.inventory_type != ?)", "Consignment Inventory")
	}
	if f.ExcludeScanned {
		d = d.Where("materialchunks.status != ?", "assigned")
	}
	if f.ExcludeConsignment {
		d = d.Where("(materialchunks.inventory_type IS NULL OR materialchunks.inventory_type != ?)", "Consignment Inventory")
	}
	if f.ActiveOnly {
		d = d.Where("materialchunks.status != ?", "returned")
	}
	if f.InactiveOnly {
		d = d.Where("materialchunks.status = ?", "returned")
	}
	return d
}

type candidateRow struct {
	models.Materialchunks
	MaterialName        string `json:"material_name"`
	MaterialType        string `json:"material_type"`
	MaterialCategory     string `json:"material_category"`
	MaterialSubCategory  string `json:"material_sub_category"`
}

const candidateSelect = "materialchunks.*, materials.name as material_name, materials.type as material_type, materials.category as material_category, materials.sub_category as material_sub_category"

func computeAuditCounts(lines []models.AuditLine) (physical, errorCount int) {
	for _, l := range lines {
		if l.ScanStatus == "reconciled" {
			physical++
		}
		if l.ScanStatus == "missing" || l.ScanStatus == "unidentified" {
			errorCount++
		}
	}
	return
}

func parseLockedMaterialIDs(raw string) map[uint]bool {
	set := make(map[uint]bool)
	if raw == "" {
		return set
	}
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if id, err := strconv.ParseUint(part, 10, 64); err == nil {
			set[uint(id)] = true
		}
	}
	return set
}

func joinMaterialIDs(ids map[uint]bool) string {
	parts := make([]string, 0, len(ids))
	for id := range ids {
		parts = append(parts, strconv.FormatUint(uint64(id), 10))
	}
	return strings.Join(parts, ",")
}

// lockNewMaterialsIntoSystemCount is called on every candidate search. For any
// material appearing in the results that hasn't been locked in before, it
// counts that material's TOTAL matching items once and adds it to
// session.SystemCount — stable for the life of the session, grows only when a
// new material is first touched.
func lockNewMaterialsIntoSystemCount(database db.Database, session *models.AuditSession, sessionType string, body auditFilterBody) {
	var materialIDs []uint
	buildCandidateQuery(database, sessionType, body).
		Distinct().
		Pluck("materialchunks.material_id", &materialIDs)
	if len(materialIDs) == 0 {
		return
	}

	locked := parseLockedMaterialIDs(session.LockedMaterialIDs)
	newlyLocked := false
	for _, matID := range materialIDs {
		if locked[matID] {
			continue
		}
		var total int64
		buildCandidateQuery(database, sessionType, auditFilterBody{SlabFilter: "all"}).
			Where("materialchunks.material_id = ?", matID).
			Count(&total)
		session.SystemCount += int(total)
		locked[matID] = true
		newlyLocked = true
	}
	if newlyLocked {
		session.LockedMaterialIDs = joinMaterialIDs(locked)
		database.Save(session)
	}
}

type sessionWithCounts struct {
	models.AuditSession
	System   int `json:"system"`
	Physical int `json:"physical"`
	Error    int `json:"error"`
}

func attachCounts(database db.Database, sessions []models.AuditSession) []sessionWithCounts {
	out := make([]sessionWithCounts, 0, len(sessions))
	for _, s := range sessions {
		var lines []models.AuditLine
		database.Where("audit_session_id = ?", s.ID).Find(&lines)
		phys, errs := computeAuditCounts(lines)
		out = append(out, sessionWithCounts{AuditSession: s, System: s.SystemCount, Physical: phys, Error: errs})
	}
	return out
}

// ════════════════════════════════════════════════════════════════════════
// 1) /audit/sessions — GET (list), POST (create)
// ════════════════════════════════════════════════════════════════════════

// sessionIDParam resolves the session id from the URL path (used by
// PUT/DELETE /audit/sessions/:id) or, when no path param is present, from
// the "id" query string (used by GET /audit/sessions?id=<id>).
func sessionIDParam(c *gin.Context) string {
	if id := c.Param("id"); id != "" {
		return id
	}
	return c.Query("id")
}

func HandleAuditSessionsCollection() gin.HandlerFunc {
	return func(c *gin.Context) {
		lines := c.Query("lines") == "true"
		id := c.Query("id")
		switch c.Request.Method {
		case http.MethodGet:
			switch {
			case lines:
				searchAuditCandidates(c)
			case id != "" && id != "-1":
				getAuditSession(c)
			default:
				listAuditSessions(c)
			}
		case http.MethodPost:
			switch {
			case lines:
				addAuditLines(c)
			case id != "" && id != "-1":
				updateAuditSession(c)
			default:
				createAuditSession(c)
			}
		default:
			c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
		}
	}
}

func listAuditSessions(c *gin.Context) {
	database := middleware.GetDB(c)
	status := c.DefaultQuery("status", "active")
	if status != "active" && status != "completed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status must be 'active' or 'completed'"})
		return
	}
	order := "start_date DESC"
	if status == "completed" {
		order = "completed_at DESC"
	}
	var sessions []models.AuditSession
	if err := database.Where("status = ?", status).Order(order).Find(&sessions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sessions"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": attachCounts(database, sessions)})
}

type createAuditSessionBody struct {
	Title          string   `json:"title"`
	StartDate      string   `json:"start_date"` // ISO-8601 instant
	Location       string   `json:"location"`
	SessionType    string   `json:"session_type"` // "slab" or "non_slab"
	InternalNotes  string   `json:"internal_notes"`
	ProductionTeam []string `json:"production_team"`
}

func createAuditSession(c *gin.Context) {
	database := middleware.GetDB(c)
	var body createAuditSessionBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if body.Title == "" || body.StartDate == "" || body.Location == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title, start_date and location are required"})
		return
	}
	if body.SessionType != "slab" && body.SessionType != "non_slab" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_type must be 'slab' or 'non_slab'"})
		return
	}
	startDate, err := time.Parse(time.RFC3339Nano, body.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format, expected ISO-8601"})
		return
	}
	session := models.AuditSession{
		Title:          body.Title,
		StartDate:      startDate,
		Location:       body.Location,
		SessionType:    body.SessionType,
		InternalNotes:  body.InternalNotes,
		ProductionTeam: joinTeam(body.ProductionTeam),
		Status:         "active",
	}
	if err := database.Create(&session); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create audit session"})
		return
	}
	c.JSON(http.StatusCreated, session)
}

// ════════════════════════════════════════════════════════════════════════
// 2) /audit/sessions/:id — GET (detail), PUT (edit / complete), DELETE
// ════════════════════════════════════════════════════════════════════════

func HandleAuditSessionItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		switch c.Request.Method {
		case http.MethodDelete:
			deleteAuditSession(c)
		default:
			c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
		}
	}
}

func getAuditSession(c *gin.Context) {
	database := middleware.GetDB(c)
	id := sessionIDParam(c)
	var session models.AuditSession
	if err := database.Preload("Lines", func(db *gorm.DB) *gorm.DB {
		return db.Order("id ASC")
	}).First(&session, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Audit session not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	phys, errs := computeAuditCounts(session.Lines)
	c.JSON(http.StatusOK, gin.H{"session": session, "system": session.SystemCount, "physical": phys, "error": errs})
}

type updateAuditSessionBody struct {
	Action         string   `json:"action"` // "" = edit fields, "complete" = mark completed
	Title          string   `json:"title"`
	StartDate      string   `json:"start_date"` // ISO-8601 instant
	Location       string   `json:"location"`
	SessionType    string   `json:"session_type"`
	InternalNotes  string   `json:"internal_notes"`
	ProductionTeam []string `json:"production_team"`
	CompletedBy    string   `json:"completed_by"`
}

func updateAuditSession(c *gin.Context) {
	database := middleware.GetDB(c)
	id := sessionIDParam(c)
	var session models.AuditSession
	if err := database.First(&session, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Audit session not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	var body updateAuditSessionBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if body.Action == "complete" {
		if session.Status != "active" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Session is not active"})
			return
		}
		now := time.Now()
		session.Status = "completed"
		session.CompletedAt = &now
		session.CompletedBy = body.CompletedBy
		if err := database.Save(&session); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete audit session"})
			return
		}
		c.JSON(http.StatusOK, session)
		return
	}

	// Regular field edit
	if session.Status != "active" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only active sessions can be edited"})
		return
	}
	if body.Title == "" || body.StartDate == "" || body.Location == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title, start_date and location are required"})
		return
	}
	if body.SessionType != "slab" && body.SessionType != "non_slab" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_type must be 'slab' or 'non_slab'"})
		return
	}
	startDate, err := time.Parse(time.RFC3339Nano, body.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format, expected ISO-8601"})
		return
	}
	if body.SessionType != session.SessionType && session.LockedMaterialIDs != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot change session type after inventory has already been searched in this session"})
		return
	}

	session.Title = body.Title
	session.StartDate = startDate
	session.Location = body.Location
	session.SessionType = body.SessionType
	session.InternalNotes = body.InternalNotes
	session.ProductionTeam = joinTeam(body.ProductionTeam)

	if err := database.Save(&session); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update audit session"})
		return
	}
	c.JSON(http.StatusOK, session)
}

func deleteAuditSession(c *gin.Context) {
	database := middleware.GetDB(c)
	id := c.Param("id")
	var session models.AuditSession
	if err := database.First(&session, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Audit session not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if err := database.Where("audit_session_id = ?", session.ID).Delete(&models.AuditLine{}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete audit lines"})
		return
	}
	if err := database.Delete(&session); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete audit session"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Audit session deleted"})
}

// ════════════════════════════════════════════════════════════════════════
// 3) /audit/sessions/:id/lines — GET (search candidates), POST (add: bulk /
//    single Done-NotFound / spare item, dispatched by body shape)
// ════════════════════════════════════════════════════════════════════════

func HandleAuditSessionLines() gin.HandlerFunc {
	return func(c *gin.Context) {
		switch c.Request.Method {
		case http.MethodGet:
			searchAuditCandidates(c)
		case http.MethodPost:
			addAuditLines(c)
		default:
			c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
		}
	}
}

func loadSession(database db.Database, id string) (models.AuditSession, error) {
	var session models.AuditSession
	err := database.First(&session, id)
	return session, err
}

func searchAuditCandidates(c *gin.Context) {
	database := middleware.GetDB(c)
	session, err := loadSession(database, sessionIDParam(c))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Audit session not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	body := auditFilterBody{
		Search:             c.Query("search"),
		Type:               c.Query("type"),
		Group:              c.Query("group"),
		Bin:                c.Query("bin"),
		BinLike:            c.Query("bin_like"),
		SlabFilter:         c.DefaultQuery("slab_filter", "all"),
		ExcludeScanned:     c.Query("exclude_scanned") == "true",
		ExcludeConsignment: c.Query("exclude_consignment") == "true",
		ActiveOnly:         c.Query("active") == "true",
		InactiveOnly:       c.Query("inactive") == "true",
	}
	if cats := c.QueryArray("category"); len(cats) > 0 {
		body.Category = cats
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	pageSize := 50

	var existingSerials []string
	database.Model(&models.AuditLine{}).Where("audit_session_id = ?", session.ID).Pluck("serialnum", &existingSerials)

	var total int64
	countQuery := buildCandidateQuery(database, session.SessionType, body)
	if len(existingSerials) > 0 {
		countQuery = countQuery.Where("materialchunks.serialnum NOT IN ?", existingSerials)
	}
	countQuery.Count(&total)

	var rows []candidateRow
	dataQuery := buildCandidateQuery(database, session.SessionType, body)
	if len(existingSerials) > 0 {
		dataQuery = dataQuery.Where("materialchunks.serialnum NOT IN ?", existingSerials)
	}
	dataQuery.
		Select(candidateSelect).
		Order("materialchunks.id ASC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&rows)

	if body.Search != "" {
		lockNewMaterialsIntoSystemCount(database, &session, session.SessionType, body)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        rows,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": int((total + int64(pageSize) - 1) / int64(pageSize)),
	})
}

type addLinesBody struct {
	auditFilterBody        // used for bulk/single add (filters or chunk_id)
	Serial           string `json:"serial"` // presence signals a spare/extra item add instead
	Note             string `json:"note"`
}

func addAuditLines(c *gin.Context) {
	database := middleware.GetDB(c)
	session, err := loadSession(database, sessionIDParam(c))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Audit session not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if session.Status != "active" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session is not active"})
		return
	}

	var body addLinesBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if strings.TrimSpace(body.Serial) != "" {
		addSpareLine(c, database, session, strings.TrimSpace(body.Serial), body.Note)
		return
	}
	addFilteredLines(c, database, session, body.auditFilterBody)
}

func addFilteredLines(c *gin.Context, database db.Database, session models.AuditSession, filter auditFilterBody) {
	status := "reconciled"
	if filter.MarkStatus == "missing" {
		status = "missing"
	}

	q := buildCandidateQuery(database, session.SessionType, filter)

	var existingSerials []string
	database.Model(&models.AuditLine{}).Where("audit_session_id = ?", session.ID).Pluck("serialnum", &existingSerials)
	if len(existingSerials) > 0 {
		q = q.Where("materialchunks.serialnum NOT IN ?", existingSerials)
	}

	var rows []candidateRow
	if err := q.Select(candidateSelect).Scan(&rows); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch matching inventory"})
		return
	}
	if len(rows) == 0 {
		c.JSON(http.StatusOK, gin.H{"added": 0, "message": "No new matching inventory found"})
		return
	}

	now := time.Now()
	lines := make([]models.AuditLine, 0, len(rows))
	for _, r := range rows {
		slabno := ""
		if r.Slabno != nil {
			slabno = *r.Slabno
		}
		bin := ""
		if r.Bin != nil {
			bin = *r.Bin
		}
		cid := r.ID
		lines = append(lines, models.AuditLine{
			AuditSessionID: session.ID,
			MaterialID:     r.MaterialID,
			ChunkID:        &cid,
			Serialnum:      r.Serialnum,
			Lot:            r.Lot,
			SlabNumber:     slabno,
			Bin:            bin,
			ProductName:    r.MaterialName,
			Type:           r.MaterialType,
			Category:       r.MaterialCategory,
			SubCategory:    r.MaterialSubCategory,
			Dimensions:     r.Dimensions,
			ExpectedStatus: r.Status,
			ScanStatus:     status,
			ScannedAt:      &now,
			CreatedAt:      now,
			UpdatedAt:      now,
		})
	}
	if err := database.Create(&lines); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add inventory lines"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"added": len(lines)})
}

func addSpareLine(c *gin.Context, database db.Database, session models.AuditSession, serial, note string) {
	var existing models.AuditLine
	err := database.Where("audit_session_id = ? AND serialnum = ?", session.ID, serial).First(&existing)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "This serial has already been added to the session"})
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	now := time.Now()
	newLine := models.AuditLine{
		AuditSessionID: session.ID,
		Serialnum:      serial,
		ScanStatus:     "unidentified", // recognized only if we can enrich it below
		IsExtra:        true,
		Notes:          strings.TrimSpace(note),
		ScannedAt:      &now,
	}

	var chunk models.Materialchunks
	if err := database.Where("serialnum = ?", serial).First(&chunk); err == nil {
		var mat models.Materials
		database.First(&mat, chunk.MaterialID)
		newLine.MaterialID = chunk.MaterialID
		cid := chunk.ID
		newLine.ChunkID = &cid
		newLine.Lot = chunk.Lot
		if chunk.Slabno != nil {
			newLine.SlabNumber = *chunk.Slabno
		}
		if chunk.Bin != nil {
			newLine.Bin = *chunk.Bin
		}
		newLine.ProductName = mat.Name
		newLine.Type = mat.Type
		newLine.Category = mat.Category
		newLine.SubCategory = mat.SubCategory
		newLine.Dimensions = chunk.Dimensions
		newLine.ExpectedStatus = chunk.Status
		newLine.ScanStatus = "reconciled" // known to the system, so it's a recognized spare item
	}

	if err := database.Create(&newLine); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add spare item"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"line": newLine})
}

// ════════════════════════════════════════════════════════════════════════
// 4) /audit/meta — GET filter options (default), or material-name
//    autocomplete when ?search= is provided (needs ?session_type=)
// ════════════════════════════════════════════════════════════════════════

func HandleAuditMeta() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)
		search := strings.TrimSpace(c.Query("search"))

		if search != "" {
			sessionType := c.Query("session_type")
			if sessionType != "slab" && sessionType != "non_slab" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "session_type must be 'slab' or 'non_slab'"})
				return
			}
			d := database.Model(&models.Materials{}).
				Joins("JOIN materialchunks ON materialchunks.material_id = materials.id")
			if sessionType == "slab" {
				d = d.Where("materials.type ILIKE ?", "%slab%")
			} else {
				d = d.Where("(materials.type IS NULL OR materials.type NOT ILIKE ?)", "%slab%")
			}
			d = d.Where("materials.name ILIKE ?", "%"+search+"%")

			var names []string
			d.Distinct().Order("materials.name ASC").Limit(20).Pluck("materials.name", &names)
			c.JSON(http.StatusOK, gin.H{"materials": names})
			return
		}

		var categories, types, groups, locations []string
		database.Model(&models.Materials{}).Where("category != ''").Distinct().Pluck("category", &categories)
		database.Model(&models.Materials{}).Where("type != ''").Distinct().Pluck("type", &types)
		database.Model(&models.Materials{}).Where("sub_category != ''").Distinct().Pluck("sub_category", &groups)
		database.Model(&models.Materialchunks{}).Where("present_location != ''").Distinct().Pluck("present_location", &locations)

		c.JSON(http.StatusOK, gin.H{
			"categories": categories,
			"types":      types,
			"groups":     groups,
			"locations":  locations,
		})
	}
}