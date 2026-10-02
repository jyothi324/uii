package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"inventory/db"
	"inventory/middleware"
	"inventory/models"
	"inventory/utils"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type FileHandler struct {
	s3Client s3iface.S3API
	region   string
}

func NewFileHandler(s3Client s3iface.S3API, region string) *FileHandler {
	return &FileHandler{
		s3Client: s3Client,
		region:   region,
	}
}

func (h *FileHandler) generatePresignedURL(bucket, key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("empty key")
	}
	if strings.HasPrefix(key, "https://") || strings.HasPrefix(key, "http://") {
		parts := strings.SplitN(key, "/", 4)
		if len(parts) == 4 {
			key = parts[3]
		}
	}
	req, _ := h.s3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return req.Presign(1 * time.Hour)
}

func (h *FileHandler) GetBucket(c *gin.Context) string {
	subdomain := middleware.GetRawSubdomain(c)
	if subdomain == "" {
		return "tulip-app-files"
	}
	return fmt.Sprintf("tulip-%s-files", subdomain)
}

// ─── Create ───────────────────────────────────────────────────────────────────

func (h *FileHandler) HandleCreateMaterial() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)
		form, err := c.MultipartForm()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse multipart form"})
			return
		}

		name := strings.TrimSpace(c.PostForm("name"))
		category := strings.TrimSpace(c.PostForm("category"))
		unit := strings.TrimSpace(c.PostForm("unit"))
		serialnum := fmt.Sprintf("%d", utils.GenerateEmployeeID(database))
		// barcode := fmt.Sprintf("%d", utils.GenerateEmployeeID(database))
		// presentLocation := strings.TrimSpace(c.PostForm("present_location"))
		// slabInput := strings.TrimSpace(c.PostForm("slabno"))
		// binInput := strings.TrimSpace(c.PostForm("bin"))
		formType := strings.TrimSpace(c.PostForm("type"))
		if formType != "" {
			var materialType models.MaterialType
			if err := database.Where("name = ?", formType).First(&materialType); err == nil {
				unit = materialType.Unit
			}
		}
		subcategory := strings.TrimSpace(c.PostForm("subcategories"))
		finish := strings.TrimSpace(c.PostForm("finishes"))
		origin := strings.TrimSpace(c.PostForm("origin"))
		baseColor := strings.TrimSpace(c.PostForm("base_colors"))
		level := strings.TrimSpace(c.PostForm("level"))
		lotNo := strings.TrimSpace(c.PostForm("lotno"))
		dimensions := strings.TrimSpace(c.PostForm("dimensions"))
		squareFeetStr := c.PostForm("square_feet")

		var squareFeet float64

		if squareFeetStr != "" {
			squareFeet, err = strconv.ParseFloat(squareFeetStr, 64)
			if err != nil {
				c.JSON(400, gin.H{"error": "Invalid square_feet value"})
				return
			}
		}

		// var slabno *string
		// if slabInput != "" {
		// 	slabno = &slabInput
		// } else {
		// 	slabno = nil
		// }

		// var bin *string
		// if binInput != "" {
		// 	bin = &binInput
		// } else {
		// 	bin = nil
		// }

		if name == "" || category == "" || serialnum == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name, category, serialnum  are required"})
			return
		}

		files := form.File["image"]

		var pictureURLs pq.StringArray
		for _, fileHeader := range files {
			file, err := fileHeader.Open()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open image"})
				return
			}
			defer file.Close()

			fileExt := filepath.Ext(fileHeader.Filename)
			uniqueID := uuid.New().String()
			s3Key := fmt.Sprintf("materials/%s%s", uniqueID, fileExt)

			bucket := h.GetBucket(c)

			_, err = h.s3Client.PutObject(&s3.PutObjectInput{
				Bucket:      aws.String(bucket),
				Key:         aws.String(s3Key),
				Body:        file,
				ContentType: aws.String(fileHeader.Header.Get("Content-Type")),
			})
			if err != nil {
				log.Println("s3 upload error:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
				return
			}
			pictureURLs = append(pictureURLs, s3Key)
		}
		// Always store as UTC so the value is an unambiguous instant,
		// independent of the app server's OS/timezone configuration.
		now := time.Now().UTC()
		material := models.Materials{
			ID:        utils.GenerateEmployeeID(database),
			Serialnum: serialnum,
			// Barcode:         barcode,
			// Slabno:          slabno,
			Name:     name,
			Category: category,
			// PresentLocation: presentLocation,
			// Bin:             bin,
			Unit:        unit,
			Pictures:    pictureURLs,
			Type:        formType,
			SubCategory: subcategory,
			Finish:      finish,
			Origin:      origin,
			BaseColor:   baseColor,
			Level:       level,
			LotNo:       lotNo,
			Dimensions:  dimensions,
			SquareFeet:  squareFeet,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		if err := database.Create(&material); err != nil {
			log.Println("insert material error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create material"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "Material created successfully", "id": material.ID})
	}
}

// parseDateFilterBoundary turns a start_date/end_date query parameter into
// a concrete UTC instant for the created_at BETWEEN filter.
//
// Preferred input is a full RFC3339 timestamp (e.g.
// "2026-09-23T18:30:00.000Z") — the UTC instant marking the start/end of
// the *caller's local* calendar day. Since CreatedAt is now always stored
// as UTC (see time.Now().UTC() at every write site), comparing against a
// UTC instant here is correct no matter what time zone the caller, the app
// server, or the database session happen to be in — there's no implicit
// "assume this string means UTC" step left anywhere.
//
// A bare "YYYY-MM-DD" is still accepted for backward compatibility with
// any other caller of this endpoint, but is NOT time-zone-correct: it's
// treated as a UTC calendar day, so a caller outside UTC sending a plain
// date will still see the boundary-day bug. Callers should send RFC3339.
func parseDateFilterBoundary(raw string, endOfDay bool) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t.UTC(), nil
	}
	d, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return time.Time{}, err
	}
	if endOfDay {
		d = d.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	}
	return d.UTC(), nil
}

// ─── Fetch All ────────────────────────────────────────────────────────────────

func (h *FileHandler) HandleFetchMaterials() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)

		var materials []models.Materials

		startDate := c.Query("start_date")
		endDate := c.Query("end_date")

		query := database

		if startDate != "" && endDate != "" {
			start, startErr := parseDateFilterBoundary(startDate, false)
			end, endErr := parseDateFilterBoundary(endDate, true)
			if startErr != nil || endErr != nil {
				log.Printf("invalid date filter values start=%q end=%q startErr=%v endErr=%v", startDate, endDate, startErr, endErr)
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date or end_date"})
				return
			}
			query = query.Where("created_at BETWEEN ? AND ?", start, end)
		}

		if err := query.Preload("Chunks", func(db *gorm.DB) *gorm.DB {
			return db.Order("id ASC")
		}).Find(&materials); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch materials",
			})
			return
		}

		total := len(materials)

		log.Printf("Fetched %d materials", total)

		// Backfill missing Unit values on the fly (never persisted) for materials
		// created before Unit was derived from their Type — avoids a DB migration.
		missingUnitTypes := map[string]bool{}
		for i := range materials {
			if strings.TrimSpace(materials[i].Unit) == "" && materials[i].Type != "" {
				missingUnitTypes[materials[i].Type] = true
			}
		}
		if len(missingUnitTypes) > 0 {
			typeNames := make([]string, 0, len(missingUnitTypes))
			for t := range missingUnitTypes {
				typeNames = append(typeNames, t)
			}
			var materialTypes []models.MaterialType
			if err := database.Where("name IN ?", typeNames).Find(&materialTypes); err == nil {
				unitByType := make(map[string]string, len(materialTypes))
				for _, mt := range materialTypes {
					unitByType[mt.Name] = mt.Unit
				}
				for i := range materials {
					if strings.TrimSpace(materials[i].Unit) == "" {
						if unit, ok := unitByType[materials[i].Type]; ok && unit != "" {
							materials[i].Unit = unit
						}
					}
				}
			}
		}

		for i := range materials {
			bucket := h.GetBucket(c)
			enriched := make([]string, 0, len(materials[i].Pictures))
			for _, key := range materials[i].Pictures {
				if url, err := h.generatePresignedURL(bucket, key); err == nil {
					enriched = append(enriched, url)
				} else {
					enriched = append(enriched, key)
				}
			}
			materials[i].Pictures = enriched
		}
		for i := range materials {
			for j := range materials[i].Chunks {
				bucket := h.GetBucket(c)
				enrichedChunk := make([]string, 0, len(materials[i].Chunks[j].Pictures))
				for _, key := range materials[i].Chunks[j].Pictures {
					if url, err := h.generatePresignedURL(bucket, key); err == nil {
						enrichedChunk = append(enrichedChunk, url)
					} else {
						enrichedChunk = append(enrichedChunk, key)
					}
				}
				materials[i].Chunks[j].Pictures = enrichedChunk
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"data":        materials,
			"total":       total,
			"page":        1,
			"page_size":   total,
			"total_pages": 1,
		})
	}
}

// ─── Get By ID ────────────────────────────────────────────────────────────────

func (h *FileHandler) HandleMaterialGetByID() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing id parameter"})
			return
		}

		var material models.Materials
		if err := database.Preload("Chunks").First(&material, id); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Println("material not found:", id)
				c.JSON(http.StatusNotFound, gin.H{"error": "Material not found"})
				return
			}
			log.Println("fetch material error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch material"})
			return
		}

		// Backfill a missing Unit on the fly (never persisted) — see HandleFetchMaterials.
		if strings.TrimSpace(material.Unit) == "" && material.Type != "" {
			var materialType models.MaterialType
			if err := database.Where("name = ?", material.Type).First(&materialType); err == nil && materialType.Unit != "" {
				material.Unit = materialType.Unit
			}
		}

		bucket := h.GetBucket(c)
		enriched := make([]string, 0, len(material.Pictures))
		for _, key := range material.Pictures {
			if url, err := h.generatePresignedURL(bucket, key); err == nil {
				enriched = append(enriched, url)
			} else {
				enriched = append(enriched, key)
			}
		}
		material.Pictures = enriched
		for i := range material.Chunks {
			enrichedChunk := make([]string, 0, len(material.Chunks[i].Pictures))
			for _, key := range material.Chunks[i].Pictures {
				if url, err := h.generatePresignedURL(bucket, key); err == nil {
					enrichedChunk = append(enrichedChunk, url)
				} else {
					enrichedChunk = append(enrichedChunk, key)
				}
			}
			material.Chunks[i].Pictures = enrichedChunk
		}
		c.JSON(http.StatusOK, material)
	}
}

// ─── maxChunkSuffix ───────────────────────────────────────────────────────────
func maxChunkSuffix(database db.Database, materialID uint) int {
	var chunks []models.Materialchunks
	database.Where("material_id = ?", materialID).Find(&chunks)

	max := 0
	for _, ch := range chunks {
		base := deriveBaseSerial(ch.Serialnum)
		idx := strings.LastIndex(base, "-")
		if idx == -1 {
			continue
		}
		n, err := strconv.Atoi(base[idx+1:])
		if err != nil {
			continue
		}
		if n > max {
			max = n
		}
	}
	return max
}

// parseSupplierDetails safely parses the supplier_details JSON string from form data.
// Returns an empty JSON object ({}) if the input is empty or invalid.
func parseSupplierDetails(raw string) datatypes.JSON {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return datatypes.JSON([]byte("{}"))
	}
	// Validate it's actually JSON
	var check interface{}
	if err := json.Unmarshal([]byte(raw), &check); err != nil {
		log.Printf("invalid supplier_details JSON, using empty object: %v", err)
		return datatypes.JSON([]byte("{}"))
	}
	return datatypes.JSON([]byte(raw))
}

// editNoteEntry represents a single timestamped note entry for a chunk edit.
type editNoteEntry struct {
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
}

// appendEditNote appends a new note entry to the existing edit_notes JSON array
// stored on a chunk, returning the updated JSON array.
func appendEditNote(existing datatypes.JSON, note string) (datatypes.JSON, error) {
	var entries []editNoteEntry
	if len(existing) > 0 {
		if err := json.Unmarshal(existing, &entries); err != nil {
			entries = []editNoteEntry{}
		}
	}
	entries = append(entries, editNoteEntry{Note: note, CreatedAt: time.Now().UTC()})
	return json.Marshal(entries)
}

// ─── Update ───────────────────────────────────────────────────────────────────

func (h *FileHandler) HandleMaterialUpdate() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing id parameter"})
			return
		}

		var m models.Materials
		if err := database.Preload("Chunks").First(&m, id); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Material not found"})
				return
			}
			log.Printf("fetch error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// ── Image uploads ────────────────────────────────────────────────────
		form, _ := c.MultipartForm()
		if form != nil && len(form.File["image"]) > 0 {
			bucket := h.GetBucket(c)
			for _, fileHeader := range form.File["image"] {
				file, err := fileHeader.Open()
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open image"})
					return
				}
				s3Key := fmt.Sprintf("materials/%s%s", uuid.New().String(), filepath.Ext(fileHeader.Filename))
				_, err = h.s3Client.PutObject(&s3.PutObjectInput{
					Bucket:      aws.String(bucket),
					Key:         aws.String(s3Key),
					Body:        file,
					ContentType: aws.String(fileHeader.Header.Get("Content-Type")),
				})
				file.Close()
				if err != nil {
					log.Println("s3 upload error:", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
					return
				}
				if m.Pictures == nil {
					m.Pictures = pq.StringArray{}
				}
				m.Pictures = append(m.Pictures, s3Key)
			}
		}

		// ── Text field updates ────────────────────────────────────────────────
		if name := strings.TrimSpace(c.PostForm("name")); name != "" {
			m.Name = name
		}
		if cat := strings.TrimSpace(c.PostForm("category")); cat != "" {
			m.Category = cat
		}
		if unit := strings.TrimSpace(c.PostForm("unit")); unit != "" {
			m.Unit = unit
		}
		// if _, exists := c.GetPostForm("slabno"); exists {
		// 	val := strings.TrimSpace(c.PostForm("slabno"))
		// 	if val == "" {
		// 		m.Slabno = nil
		// 	} else {
		// 		m.Slabno = &val
		// 	}
		// }
		// if val := strings.TrimSpace(c.PostForm("present_location")); val != "" {
		// 	m.PresentLocation = val
		// }
		// if _, exists := c.GetPostForm("bin"); exists {
		// 	val := strings.TrimSpace(c.PostForm("bin"))
		// 	if val == "" {
		// 		m.Bin = nil
		// 	} else {
		// 		m.Bin = &val
		// 	}
		// }
		if val := c.PostForm("available"); val != "" {
			if n, err := strconv.Atoi(val); err == nil {
				m.Available = n
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid available value"})
				return
			}
		}
		if val := c.PostForm("threshold"); val != "" {
			if n, err := strconv.Atoi(val); err == nil {
				m.Threshold = n
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid threshold value"})
				return
			}
		}
		if val := c.PostForm("retail_cost"); val != "" {
			if f, err := strconv.ParseFloat(val, 64); err == nil {
				m.RetailCost = f
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid retail_cost value"})
				return
			}
		}
		if val := c.PostForm("procost"); val != "" {
			if f, err := strconv.ParseFloat(val, 64); err == nil {
				m.ProCost = f
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid procost value"})
				return
			}
		}
		if val := strings.TrimSpace(c.PostForm("type")); val != "" {
			m.Type = val
			var materialType models.MaterialType
			if err := database.Where("name = ?", val).First(&materialType); err == nil {
				m.Unit = materialType.Unit
			}
		}
		if val := strings.TrimSpace(c.PostForm("subcategories")); val != "" {
			m.SubCategory = val
		}
		if val := strings.TrimSpace(c.PostForm("origin")); val != "" {
			m.Origin = val
		}
		if val := strings.TrimSpace(c.PostForm("finishes")); val != "" {
			m.Finish = val
		}
		if val := strings.TrimSpace(c.PostForm("base_colors")); val != "" {
			m.BaseColor = val
		}
		if val := strings.TrimSpace(c.PostForm("level")); val != "" {
			m.Level = val
		}
		if val := strings.TrimSpace(c.PostForm("lotno")); val != "" {
			m.LotNo = val
		}
		if val := strings.TrimSpace(c.PostForm("dimensions")); val != "" {
			m.Dimensions = val
		}
		if val := strings.TrimSpace(c.PostForm("square_feet")); val != "" {
			if f, err := strconv.ParseFloat(val, 64); err == nil {
				m.SquareFeet = f
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid square_feet value"})
				return
			}
		}

		if err := database.Save(&m); err != nil {
			log.Printf("save material error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update material"})
			return
		}

		// ── Parse supplier_details from form (snapshot, persists even if supplier deleted) ──
		supplierDetailsJSON := parseSupplierDetails(c.PostForm("supplier_details"))

		// ── Sync material chunks to new Available count ───────────────────────
		if availVal := c.PostForm("available"); availVal != "" {
			newCount := m.Available

			if receivedOrder := c.PostForm("procurement"); receivedOrder == "true" {
				lot1 := strings.TrimSpace(c.PostForm("lot"))
				inventoryType := strings.TrimSpace(c.PostForm("inventory_type"))
				invoiceNo := strings.TrimSpace(c.PostForm("invoice_no"))

				purchaseReceivedIDStr := strings.TrimSpace(c.PostForm("purchase_received_id"))
				var purchaseReceivedID uint
				if purchaseReceivedIDStr != "" {
					if parsed, err := strconv.ParseUint(purchaseReceivedIDStr, 10, 64); err == nil {
						purchaseReceivedID = uint(parsed)
					}
				}
				presentLocation := strings.TrimSpace(c.PostForm("present_location"))

				var bin *string
				if val := strings.TrimSpace(c.PostForm("bin")); val != "" {
					bin = &val
				}

				// retailCost, _ := strconv.ParseFloat(c.PostForm("retail_cost"), 64)
				// proCost, _ := strconv.ParseFloat(c.PostForm("procost"), 64)

				availableChunkCount := 0
				for _, ch := range m.Chunks {
					if ch.Status == "available" {
						availableChunkCount++
					}
				}

				chunksToCreate := newCount - availableChunkCount
				if chunksToCreate > 0 {
					serialnum := fmt.Sprintf("%d", utils.GenerateEmployeeID(database))
					startIdx := maxChunkSuffix(database, m.ID) + 1
					now := time.Now().UTC()
					var newChunks []models.Materialchunks
					for i := 0; i < chunksToCreate; i++ {
						newChunks = append(newChunks, models.Materialchunks{
							MaterialID:         m.ID,
							PurchaseReceivedID: purchaseReceivedID,
							Serialnum:          fmt.Sprintf("%s-%d", serialnum, startIdx+i),
							// Barcode:            "",
							Slabno:          nil,
							PresentLocation: presentLocation,
							Dimensions:      "",
							Lot:             lot1,
							BaseColor:       m.BaseColor,
							Bin:             bin,
							RetailCost:      m.RetailCost,
							Pictures:        m.Pictures,
							ProCost:         m.ProCost,
							CreatedAt:       now,
							UpdatedAt:       now,
							Status:          "available",
							SupplierDetails: supplierDetailsJSON,
							InventoryType:   inventoryType,
							InvoiceNo:       invoiceNo,
						})
					}
					if err := database.Create(&newChunks); err != nil {
						log.Printf("create chunks error: %v", err)
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create new chunks"})
						return
					}
				}

			} else {
				// Non-procurement manual update
				currentCount := len(m.Chunks)
				if newCount > currentCount {
					startIdx := maxChunkSuffix(database, m.ID) + 1
					now := time.Now().UTC()
					var newChunks []models.Materialchunks
					for i := 0; i < (newCount - currentCount); i++ {
						newChunks = append(newChunks, models.Materialchunks{
							MaterialID: m.ID,
							Serialnum:  fmt.Sprintf("%s-%d", m.Serialnum, startIdx+i),
							// Barcode:         "",
							Slabno:          nil,
							PresentLocation: "",
							Dimensions:      "",
							Lot:             "",
							RetailCost:      m.RetailCost,
							Pictures:        m.Pictures,
							ProCost:         m.ProCost,
							BaseColor:       m.BaseColor,
							Bin:             nil,
							CreatedAt:       now,
							UpdatedAt:       now,
							Status:          "available",
							SupplierDetails: datatypes.JSON([]byte("{}")),
						})
					}
					if err := database.Create(&newChunks); err != nil {
						log.Printf("create chunks error: %v", err)
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create new chunks"})
						return
					}
				}
			}
		}

		bucket2 := h.GetBucket(c)
		enrichedPics := make([]string, 0, len(m.Pictures))
		for _, key := range m.Pictures {
			if url, err := h.generatePresignedURL(bucket2, key); err == nil {
				enrichedPics = append(enrichedPics, url)
			} else {
				enrichedPics = append(enrichedPics, key)
			}
		}
		m.Pictures = enrichedPics
		c.JSON(http.StatusOK, m)
	}
}

func (h *FileHandler) HandleMaterialChunkUpdate() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing chunk id"})
			return
		}

		var chunk models.Materialchunks
		if err := database.First(&chunk, id); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Chunk not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// CRITICAL: Store the original status and job_id BEFORE any changes
		originalStatus := chunk.Status
		originalJobID := chunk.JobID

		newDimensions := strings.TrimSpace(c.PostForm("dimensions"))
		noRevision := c.PostForm("no_revision") == "true"

		// Check if this is a cut operation on an assigned chunk
		isCutOperation := (originalStatus == "assigned" || originalStatus == "reserved") &&
			newDimensions != "" &&
			newDimensions != chunk.Dimensions &&
			!noRevision

		// Update other fields first
		if val := strings.TrimSpace(c.PostForm("present_location")); val != "" {
			chunk.PresentLocation = val
		}
		if val := c.PostForm("retail_cost"); val != "" {
			if f, err := strconv.ParseFloat(val, 64); err == nil {
				chunk.RetailCost = f
			}
		}
		if val := c.PostForm("procost"); val != "" {
			if f, err := strconv.ParseFloat(val, 64); err == nil {
				chunk.ProCost = f
			}
		}
		if val := c.PostForm("slabno"); val != "" {
			chunk.Slabno = &val
		}
		if val := strings.TrimSpace(c.PostForm("lot")); val != "" {
			chunk.Lot = val
		}
		if val, exists := c.GetPostForm("base_colors"); exists {
			chunk.BaseColor = strings.TrimSpace(val)
		}
		if val, exists := c.GetPostForm("bin"); exists {
			trimmed := strings.TrimSpace(val)
			if trimmed == "" {
				chunk.Bin = nil
			} else {
				chunk.Bin = &trimmed
			}
		}
		if val := strings.TrimSpace(c.PostForm("supplier_details")); val != "" {
			chunk.SupplierDetails = parseSupplierDetails(val)
		}
		if val, exists := c.GetPostForm("level"); exists {
			chunk.Level = strings.TrimSpace(val)
		}
		if val, exists := c.GetPostForm("inventory_type"); exists {
			chunk.InventoryType = strings.TrimSpace(val)
		}
		if val, exists := c.GetPostForm("invoice_no"); exists {
			chunk.InvoiceNo = strings.TrimSpace(val)
		}
		if val := strings.TrimSpace(c.PostForm("received_date")); val != "" {
			if parsedDate, err := time.Parse("2006-01-02", val); err == nil {
				chunk.CreatedAt = parsedDate
			}
		}
		var notes *string
		if val := strings.TrimSpace(c.PostForm("notes")); val != "" {
			notes = &val
		}

		if editNoteVal := strings.TrimSpace(c.PostForm("edit_note")); editNoteVal != "" {
			updatedEditNotes, err := appendEditNote(chunk.EditNotes, editNoteVal)
			if err != nil {
				log.Printf("failed to append edit note for chunk %s: %v", id, err)
			} else {
				chunk.EditNotes = updatedEditNotes
			}
		}

		// Handle image uploads
		form, _ := c.MultipartForm()
		if form != nil && len(form.File["image"]) > 0 {
			bucket := h.GetBucket(c)
			chunk.Pictures = pq.StringArray{}
			for _, fileHeader := range form.File["image"] {
				file, err := fileHeader.Open()
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open image"})
					return
				}
				s3Key := fmt.Sprintf("chunks/%s%s", uuid.New().String(), filepath.Ext(fileHeader.Filename))
				_, err = h.s3Client.PutObject(&s3.PutObjectInput{
					Bucket:      aws.String(bucket),
					Key:         aws.String(s3Key),
					Body:        file,
					ContentType: aws.String(fileHeader.Header.Get("Content-Type")),
				})
				file.Close()
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
					return
				}
				chunk.Pictures = append(chunk.Pictures, s3Key)
			}
		}

		// Handle cut operation for assigned chunks
		if isCutOperation {
			// 2. Create remnant chunk with status "available"
			baseSerial := deriveBaseSerial(chunk.Serialnum)

			// Find next revision number
			var maxSerial string
			database.Model(&models.Materialchunks{}).
				Where("material_id = ? AND serialnum LIKE ?", chunk.MaterialID, baseSerial+"-r%").
				Order("serialnum DESC").
				Limit(1).
				Pluck("serialnum", &maxSerial)

			nextRev := 1
			if maxSerial != "" {
				if idx := strings.LastIndex(maxSerial, "-r"); idx != -1 {
					if n, err := strconv.Atoi(maxSerial[idx+2:]); err == nil {
						nextRev = n + 1
					}
				}
			}

			now := time.Now().UTC()
			newSerial := fmt.Sprintf("%s-r%d", baseSerial, nextRev)

			// 1. Create the NEW remnant chunk with the dimensions provided by the user.
			// This new chunk becomes "available".
			remnantChunk := models.Materialchunks{
				MaterialID:      chunk.MaterialID,
				Serialnum:       newSerial, // The remnant gets the new revision serial
				Slabno:          chunk.Slabno,
				PresentLocation: chunk.PresentLocation,
				Dimensions:      newDimensions, // The part that was cut off
				Lot:             chunk.Lot,
				RetailCost:      chunk.RetailCost,
				Pictures:        chunk.Pictures,
				ProCost:         chunk.ProCost,
				BaseColor:       chunk.BaseColor,
				Bin:             chunk.Bin,
				CreatedAt:       now,
				UpdatedAt:       now,
				Status:          "available",
				JobID:           "",
				SupplierDetails: chunk.SupplierDetails,
				Notes:           notes,
			}
			database.Create(&remnantChunk)

			// Recalculate available count
			// This operation doesn't change the number of chunks assigned to jobs,
			// but it does add one 'available' chunk. We'll update the count.
			database.Exec("UPDATE materials SET available = available + 1, updated_at = ? WHERE id = ?", time.Now(), chunk.MaterialID)

			// Prepare response with enriched pictures
			bucket := h.GetBucket(c)
			enrichPictures := func(pics pq.StringArray) []string {
				out := make([]string, 0, len(pics))
				for _, key := range pics {
					if url, err := h.generatePresignedURL(bucket, key); err == nil {
						out = append(out, url)
					} else {
						out = append(out, key)
					}
				}
				return out
			}

			c.JSON(http.StatusOK, gin.H{
				"message":        "Remnant created successfully from assigned chunk.",
				"original_chunk": chunk,
				"remnant_chunk": gin.H{
					"id":         remnantChunk.ID,
					"serialnum":  remnantChunk.Serialnum,
					"dimensions": remnantChunk.Dimensions,
					"status":     remnantChunk.Status,
					"pictures":   enrichPictures(remnantChunk.Pictures),
				},
			})
			return // ⚠️ CRITICAL: Prevent fall-through to the next block
		}

		// ── Non-cut operation (regular edit or remnant creation for non-assigned) ──
		if newDimensions != "" && newDimensions != chunk.Dimensions && !noRevision {
			// Create a new revision with the new dimensions
			baseSerial := deriveBaseSerial(chunk.Serialnum)
			var revisionCount int64
			database.Model(&models.Materialchunks{}).
				Where("material_id = ? AND serialnum LIKE ?", chunk.MaterialID, baseSerial+"%-r%").
				Count(&revisionCount)
			nextRev := int(revisionCount) + 1
			newSerial := fmt.Sprintf("%s-r%d", baseSerial, nextRev)
			now := time.Now().UTC()

			// Create the new revision (remnant) with status "available"
			newChunk := models.Materialchunks{
				MaterialID:      chunk.MaterialID,
				Serialnum:       newSerial,
				Slabno:          chunk.Slabno,
				PresentLocation: chunk.PresentLocation,
				Dimensions:      newDimensions,
				Lot:             chunk.Lot,
				RetailCost:      chunk.RetailCost,
				Pictures:        chunk.Pictures,
				ProCost:         chunk.ProCost,
				BaseColor:       chunk.BaseColor,
				Bin:             chunk.Bin,
				CreatedAt:       now,
				UpdatedAt:       now,
				Status:          "available",
				SupplierDetails: chunk.SupplierDetails,
				JobID:           "",
				Notes:           notes,
			}
			if err := database.Create(&newChunk); err != nil {
				log.Printf("failed to create revision chunk: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create revision chunk"})
				return
			}

			// Also update the dimensions on the original chunk (if needed)
			// But note: for remnant creation, the original chunk keeps its original dimensions
			// Only the new revision gets the new dimensions

			// ✅ FIX: Keep the original chunk's status UNCHANGED
			// We must explicitly set the status back to originalStatus and save
			chunk.Status = originalStatus // ← THIS IS THE FIX
			chunk.JobID = originalJobID   // ← Also preserve job_id

			// Save the original chunk with its original status
			if err := database.Save(&chunk); err != nil {
				log.Printf("failed to update original chunk: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update original chunk"})
				return
			}

			// Recalculate material available count
			var allChunks []models.Materialchunks
			database.Where("material_id = ?", chunk.MaterialID).Find(&allChunks)
			availableCount := 0
			for _, ch := range allChunks {
				if ch.Status == "available" {
					availableCount++
				}
			}
			var material models.Materials
			if err := database.First(&material, chunk.MaterialID); err == nil {
				material.Available = availableCount
				database.Save(&material)
			}

			bucket := h.GetBucket(c)
			enrichPictures := func(pics pq.StringArray) []string {
				out := make([]string, 0, len(pics))
				for _, key := range pics {
					if url, err := h.generatePresignedURL(bucket, key); err == nil {
						out = append(out, url)
					} else {
						out = append(out, key)
					}
				}
				return out
			}
			chunk.Pictures = enrichPictures(chunk.Pictures)
			newChunk.Pictures = enrichPictures(newChunk.Pictures)

			c.JSON(http.StatusOK, gin.H{
				"original": chunk,
				"revision": newChunk,
				"message":  fmt.Sprintf("Remnant created as %s with dimensions %s", newSerial, newDimensions),
			})
			return
		}

		// Simple dimension update (no revision) - preserve status
		if newDimensions != "" {
			chunk.Dimensions = newDimensions
		}

		// ✅ For simple updates, also preserve the original status
		chunk.Status = originalStatus
		chunk.JobID = originalJobID

		if err := database.Save(&chunk); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update chunk"})
			return
		}

		bucket := h.GetBucket(c)
		enrichedPics := make([]string, 0, len(chunk.Pictures))
		for _, key := range chunk.Pictures {
			if url, err := h.generatePresignedURL(bucket, key); err == nil {
				enrichedPics = append(enrichedPics, url)
			} else {
				enrichedPics = append(enrichedPics, key)
			}
		}
		chunk.Pictures = enrichedPics
		c.JSON(http.StatusOK, chunk)
	}
}

func deriveBaseSerial(serial string) string {
	idx := strings.LastIndex(serial, "-r")
	if idx == -1 {
		return serial
	}
	suffix := serial[idx+2:]
	allDigits := len(suffix) > 0
	for _, ch := range suffix {
		if ch < '0' || ch > '9' {
			allDigits = false
			break
		}
	}
	if allDigits {
		return serial[:idx]
	}
	return serial
}
func subtractDimensions(oldDim, newDim string) (string, error) {
	if strings.TrimSpace(oldDim) == "" {
		return newDim, nil
	}
	oldParts := strings.Split(oldDim, "x")
	newParts := strings.Split(newDim, "x")
	if len(oldParts) != 2 || len(newParts) != 2 {
		return "", fmt.Errorf("invalid dimension format")
	}
	oldW, err1 := strconv.Atoi(strings.TrimSpace(oldParts[0]))
	oldH, err2 := strconv.Atoi(strings.TrimSpace(oldParts[1]))
	newW, err3 := strconv.Atoi(strings.TrimSpace(newParts[0]))
	newH, err4 := strconv.Atoi(strings.TrimSpace(newParts[1]))
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		return "", fmt.Errorf("invalid dimension numbers")
	}
	resultW := oldW - newW
	resultH := oldH - newH
	if resultW < 0 {
		resultW = 0
	}
	if resultH < 0 {
		resultH = 0
	}
	return fmt.Sprintf("%dx%d", resultW, resultH), nil
}

// ─── Delete ───────────────────────────────────────────────────────────────────

func (h *FileHandler) DeleteMaterialByID() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing id parameter"})
			return
		}

		var material models.Materials
		if err := database.Preload("Chunks").First(&material, id); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Println("material not found:", id)
				c.JSON(http.StatusNotFound, gin.H{"error": "Material not found"})
				return
			}
			log.Println("fetch material error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch material"})
			return
		}

		if len(material.Chunks) > 0 {
			if err := database.Where("material_id = ?", material.ID).Delete(&models.Materialchunks{}); err != nil {
				log.Println("delete chunks error:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete material chunks"})
				return
			}
		}

		if err := database.Delete(&material); err != nil {
			log.Println("delete material error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete material"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Material deleted successfully"})
	}
}

// ─── Delete Single Image ──────────────────────────────────────────────────────

func (h *FileHandler) HandleDeleteMaterialImage() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing id parameter"})
			return
		}

		var body struct {
			ImageURL string `json:"image_url"`
		}
		if err := c.ShouldBindJSON(&body); err != nil || body.ImageURL == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "image_url is required in request body"})
			return
		}

		var m models.Materials
		if err := database.First(&m, id); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Material not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		normalizeKey := func(raw string) string {
			if strings.HasPrefix(raw, "https://") || strings.HasPrefix(raw, "http://") {
				parts := strings.SplitN(raw, "/", 4)
				if len(parts) == 4 {
					return parts[3]
				}
			}
			return raw
		}
		deleteKey := normalizeKey(body.ImageURL)

		newPictures := pq.StringArray{}
		found := false
		for _, pic := range m.Pictures {
			if normalizeKey(pic) == deleteKey {
				found = true
				continue
			}
			newPictures = append(newPictures, pic)
		}

		if !found {
			c.JSON(http.StatusNotFound, gin.H{"error": "Image URL not found in this material"})
			return
		}

		m.Pictures = newPictures
		if err := database.Save(&m); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update material"})
			return
		}

		bucket := h.GetBucket(c)
		enrichedPics := make([]string, 0, len(m.Pictures))
		for _, key := range m.Pictures {
			if url, err := h.generatePresignedURL(bucket, key); err == nil {
				enrichedPics = append(enrichedPics, url)
			} else {
				enrichedPics = append(enrichedPics, key)
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message":  "Image deleted successfully",
			"pictures": enrichedPics,
		})
	}
}

// ─── Replace Single Image ─────────────────────────────────────────────────────

func (h *FileHandler) HandleReplaceMaterialImage() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing id parameter"})
			return
		}

		replaceURL := strings.TrimSpace(c.PostForm("replace_url"))
		if replaceURL == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "replace_url is required"})
			return
		}

		var m models.Materials
		if err := database.First(&m, id); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Material not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		normalizeKey := func(raw string) string {
			if strings.HasPrefix(raw, "https://") || strings.HasPrefix(raw, "http://") {
				if idx := strings.Index(raw, "?"); idx != -1 {
					raw = raw[:idx]
				}
				parts := strings.SplitN(raw, "/", 4)
				if len(parts) == 4 {
					return parts[3]
				}
			}
			return raw
		}

		replaceKey := normalizeKey(replaceURL)
		found := false
		for _, pic := range m.Pictures {
			if normalizeKey(pic) == replaceKey {
				found = true
				break
			}
		}
		if !found {
			c.JSON(http.StatusNotFound, gin.H{"error": "replace_url not found in this material"})
			return
		}

		fileHeader, err := c.FormFile("image")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "image file is required"})
			return
		}
		file, err := fileHeader.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open image"})
			return
		}
		defer file.Close()

		bucket := h.GetBucket(c)
		newS3Key := fmt.Sprintf("materials/%s%s", uuid.New().String(), filepath.Ext(fileHeader.Filename))
		_, err = h.s3Client.PutObject(&s3.PutObjectInput{
			Bucket:      aws.String(bucket),
			Key:         aws.String(newS3Key),
			Body:        file,
			ContentType: aws.String(fileHeader.Header.Get("Content-Type")),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
			return
		}

		newPictures := pq.StringArray{}
		for _, pic := range m.Pictures {
			if normalizeKey(pic) == replaceKey {
				newPictures = append(newPictures, newS3Key)
			} else {
				newPictures = append(newPictures, pic)
			}
		}
		m.Pictures = newPictures

		if err := database.Save(&m); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update image"})
			return
		}

		presignedNew, presignErr := h.generatePresignedURL(bucket, newS3Key)
		if presignErr != nil {
			presignedNew = newS3Key
		}
		enrichedPics := make([]string, 0, len(m.Pictures))
		for _, key := range m.Pictures {
			if url, err := h.generatePresignedURL(bucket, key); err == nil {
				enrichedPics = append(enrichedPics, url)
			} else {
				enrichedPics = append(enrichedPics, key)
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"message":  "Image replaced successfully",
			"new_url":  presignedNew,
			"pictures": enrichedPics,
		})
	}
}

// ─── Chunk Status Update (single) ─────────────────────────────────────────────

func (h *FileHandler) HandleChunkStatusUpdate() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing chunk id"})
			return
		}

		var body struct {
			Status string `json:"status"`
		}
		if err := c.ShouldBindJSON(&body); err != nil || body.Status == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "status is required"})
			return
		}
		if body.Status != "available" && body.Status != "assigned" && body.Status != "reserved" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "status must be 'available', 'assigned', or 'reserved'"})
			return
		}

		var chunk models.Materialchunks
		if err := database.First(&chunk, id); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Chunk not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		clearJobID := ""
		jobIDToSet := chunk.JobID
		if body.Status == "available" {
			jobIDToSet = clearJobID
		}

		if err := database.Exec(
			"UPDATE materialchunks SET status = ?, job_id = ?, updated_at = ? WHERE id = ?",
			body.Status, jobIDToSet, time.Now(), chunk.ID,
		); err != nil {
			log.Printf("failed to update chunk status %s: %v", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update chunk status"})
			return
		}

		// Recalculate available count for this material
		var allChunks []models.Materialchunks
		if err := database.Where("material_id = ?", chunk.MaterialID).Find(&allChunks); err != nil {
			log.Printf("failed to fetch chunks for material %d: %v", chunk.MaterialID, err)
		} else {
			availableCount := 0
			for _, ch := range allChunks {
				if ch.ID == chunk.ID {
					if body.Status == "available" {
						availableCount++
					}
					continue
				}
				if ch.Status == "available" {
					availableCount++
				}
			}
			if err := database.Exec(
				"UPDATE materials SET available = ?, updated_at = ? WHERE id = ?",
				availableCount, time.Now(), chunk.MaterialID,
			); err != nil {
				log.Printf("failed to update material available count: %v", err)
			}
		}

		var updatedChunk models.Materialchunks
		if err := database.First(&updatedChunk, id); err == nil {
			c.JSON(http.StatusOK, updatedChunk)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "chunk status updated"})
	}
}

// ─── Bulk Chunk Status Update ─────────────────────────────────────────────────

func (h *FileHandler) HandleBulkChunkStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)

		var body struct {
			ChunkIDs   []string `json:"chunk_ids"`
			Status     string   `json:"status"`
			JobID      string   `json:"job_id"`
			ReleaseIDs []string `json:"release_ids"`
			Force      bool     `json:"force"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		if body.Status != "available" && body.Status != "assigned" && body.Status != "reserved" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "status must be available, assigned, or reserved"})
			return
		}

		affectedMaterialIDs := map[uint]bool{}

		// ── Release block ─────────────────────────────────────────────────────
		if len(body.ReleaseIDs) > 0 {
			var toRelease []models.Materialchunks
			if err := database.Where("id IN ?", body.ReleaseIDs).Find(&toRelease); err != nil {
				log.Printf("release fetch error: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch chunks to release"})
				return
			}

			for i := range toRelease {
				// Only release assigned chunks
				if toRelease[i].Status != "assigned" && toRelease[i].Status != "reserved" {
					continue
				}
				// Only release if belongs to requesting job or force release (empty job_id)
				if body.JobID != "" && toRelease[i].JobID != body.JobID {
					continue
				}
				// Use Exec to force write empty string — GORM Save skips zero values
				if err := database.Exec(
					"UPDATE materialchunks SET status = ?, job_id = ?, updated_at = ? WHERE id = ?",
					"available", "", time.Now(), toRelease[i].ID,
				); err != nil {
					log.Printf("failed to release chunk %d: %v", toRelease[i].ID, err)
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": fmt.Sprintf("failed to release chunk %d", toRelease[i].ID),
					})
					return
				}
				affectedMaterialIDs[toRelease[i].MaterialID] = true
			}
		}

		updatedCount := 0
		skippedIDs := []string{}

		if len(body.ChunkIDs) > 0 {
			var chunks []models.Materialchunks
			if err := database.Where("id IN ?", body.ChunkIDs).Find(&chunks); err != nil {
				log.Printf("chunks fetch error: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch chunks"})
				return
			}

			chunkByID := map[string]*models.Materialchunks{}
			for i := range chunks {
				chunkByID[fmt.Sprintf("%d", chunks[i].ID)] = &chunks[i]
			}

			for _, id := range body.ChunkIDs {
				ch, ok := chunkByID[id]
				if !ok {
					log.Printf("chunk %s not found in DB", id)
					skippedIDs = append(skippedIDs, id)
					continue
				}

				// If this chunk is free or already owned by same job — assign directly
				if ch.Status == "available" || ch.Status == "reserved" || ch.JobID == "" || ch.JobID == body.JobID {
					if err := database.Exec(
						"UPDATE materialchunks SET status = ?, job_id = ?, updated_at = ? WHERE id = ?",
						body.Status, body.JobID, time.Now(), ch.ID,
					); err != nil {
						log.Printf("failed to update chunk %s: %v", id, err)
						c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to update chunk %s", id)})
						return
					}
					affectedMaterialIDs[ch.MaterialID] = true
					updatedCount++
					continue
				}

				// ── Chunk is taken by another job — find another available chunk from same material ──
				log.Printf("chunk %s already assigned to job %q, looking for alternative in material %d", id, ch.JobID, ch.MaterialID)

				var alternative models.Materialchunks
				err := database.
					Where("material_id = ? AND status = ? AND id NOT IN ?", ch.MaterialID, "available", body.ChunkIDs).
					First(&alternative)

				if err != nil {
					// No alternative found — truly skip
					log.Printf("no available alternative for chunk %s in material %d", id, ch.MaterialID)
					skippedIDs = append(skippedIDs, id)
					continue
				}

				// Assign the alternative instead
				if err := database.Exec(
					"UPDATE materialchunks SET status = ?, job_id = ?, updated_at = ? WHERE id = ?",
					body.Status, body.JobID, time.Now(), alternative.ID,
				); err != nil {
					log.Printf("failed to assign alternative chunk %d: %v", alternative.ID, err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to assign alternative chunk for %s", id)})
					return
				}

				log.Printf("assigned alternative chunk %d instead of requested chunk %s", alternative.ID, id)
				affectedMaterialIDs[alternative.MaterialID] = true
				updatedCount++
			}
		}

		// ── Recalculate available count per affected material ─────────────────
		for matID := range affectedMaterialIDs {
			var allChunks []models.Materialchunks
			if err := database.Where("material_id = ?", matID).Find(&allChunks); err != nil {
				log.Printf("failed to fetch chunks for material %d: %v", matID, err)
				continue
			}
			availableCount := 0
			for _, ch := range allChunks {
				if ch.Status == "available" {
					availableCount++
				}
			}
			var material models.Materials
			if err := database.First(&material, matID); err == nil {
				if err := database.Exec(
					"UPDATE materials SET available = ?, updated_at = ? WHERE id = ?",
					availableCount, time.Now(), matID,
				); err != nil {
					log.Printf("failed to update material available count %d: %v", matID, err)
				}
			}
		}

		log.Printf("Bulk status: updated=%d skipped=%v jobID=%q", updatedCount, skippedIDs, body.JobID)

		c.JSON(http.StatusOK, gin.H{
			"message":       "chunk statuses updated",
			"updated_count": updatedCount,
			"skipped_ids":   skippedIDs,
		})
	}
}

func (h *FileHandler) HandleDeleteMaterialChunk() gin.HandlerFunc {
	return func(c *gin.Context) {
		database := middleware.GetDB(c)
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing chunk id"})
			return
		}

		var chunk models.Materialchunks
		if err := database.First(&chunk, id); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Chunk not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		if err := database.Delete(&chunk); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete chunk"})
			return
		}

		// Recalculate available count for this material
		var allChunks []models.Materialchunks
		if err := database.Where("material_id = ?", chunk.MaterialID).Find(&allChunks); err != nil {
			log.Printf("failed to fetch chunks for material %d: %v", chunk.MaterialID, err)
		} else {
			availableCount := 0
			for _, ch := range allChunks {
				if ch.Status == "available" {
					availableCount++
				}
			}
			var material models.Materials
			if err := database.First(&material, chunk.MaterialID); err == nil {
				material.Available = availableCount
				database.Save(&material)
			}
		}

		c.JSON(http.StatusOK, gin.H{"message": "Chunk deleted successfully"})
	}
}
