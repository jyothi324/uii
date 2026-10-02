package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
	"github.com/tuliporg/agent-management-api/db"
	"github.com/tuliporg/agent-management-api/middleware"
	"github.com/tuliporg/agent-management-api/models"
)

// parseLinkedItems safely decodes the linked_items JSON of a scratchpad.
func parseLinkedItems(raw models.JSONB) []models.LinkedItem {
	var items []models.LinkedItem
	if len(raw) == 0 {
		return nil
	}
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil
	}
	return items
}

// CreateScratchpad handles POST /api/scratchpads
func (h *Handler) CreateScratchpad(c *gin.Context) {
	db := middleware.GetDB(c)
	var scratchpad models.Scratchpad
	if err := c.ShouldBindJSON(&scratchpad); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := db.Create(&scratchpad); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create scratchpad"})
		return
	}

	// Sync note to linked entities and remember where each copy went
	items := parseLinkedItems(scratchpad.LinkedItems)
	if len(items) > 0 {
		for i := range items {
			items[i] = h.syncNoteToParent(db, items[i], scratchpad.Content)
		}
		if enriched, err := json.Marshal(items); err == nil {
			scratchpad.LinkedItems = models.JSONB(enriched)
			if err := db.Save(&scratchpad); err != nil {
				fmt.Printf("Error saving scratchpad link metadata: %v\n", err)
			}
		}
	}

	c.JSON(http.StatusCreated, scratchpad)
}

// GetAllScratchpads handles GET /api/scratchpads
func (h *Handler) GetAllScratchpads(c *gin.Context) {
	db := middleware.GetDB(c)
	var scratchpads []models.Scratchpad

	query := db.Model(&models.Scratchpad{})

	// Optional filters
	if priority := c.Query("priority"); priority != "" {
		query = query.Where("priority = ?", priority)
	}

	if err := query.Order("created_at DESC").Find(&scratchpads); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch scratchpads"})
		return
	}

	c.JSON(http.StatusOK, scratchpads)
}

// GetScratchpad handles GET /api/scratchpads/:id
func (h *Handler) GetScratchpad(c *gin.Context) {
	db := middleware.GetDB(c)
	id := c.Param("id")
	var scratchpad models.Scratchpad
	if err := db.First(&scratchpad, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Scratchpad not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch scratchpad"})
		return
	}
	c.JSON(http.StatusOK, scratchpad)
}

// UpdateScratchpad handles PUT /api/scratchpads/:id
func (h *Handler) UpdateScratchpad(c *gin.Context) {
	db := middleware.GetDB(c)
	id := c.Param("id")
	var scratchpad models.Scratchpad

	if err := db.First(&scratchpad, id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Scratchpad not found"})
		return
	}

	// Snapshot before binding: JSONB.UnmarshalJSON reuses the backing array, so
	// the old linked items must be copied first.
	oldContent := scratchpad.Content
	oldLinked := append(models.JSONB(nil), scratchpad.LinkedItems...)

	if err := c.ShouldBindJSON(&scratchpad); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Links are managed by the backend only; an edit must never lose them.
	scratchpad.LinkedItems = oldLinked

	if err := db.Save(&scratchpad); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update scratchpad"})
		return
	}

	// Push the edited text to every place the note was copied to
	if scratchpad.Content != oldContent {
		for _, item := range parseLinkedItems(scratchpad.LinkedItems) {
			h.updateLinkedNote(db, item, oldContent, scratchpad.Content, false)
		}
	}

	c.JSON(http.StatusOK, scratchpad)
}

// DeleteScratchpad handles DELETE /api/scratchpads/:id
func (h *Handler) DeleteScratchpad(c *gin.Context) {
	db := middleware.GetDB(c)
	id := c.Param("id")

	var scratchpad models.Scratchpad
	if err := db.First(&scratchpad, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Scratchpad not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch scratchpad"})
		return
	}

	if err := db.Delete(&scratchpad); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete scratchpad"})
		return
	}

	// Remove the copies of this note from every linked entity
	for _, item := range parseLinkedItems(scratchpad.LinkedItems) {
		h.updateLinkedNote(db, item, scratchpad.Content, "", true)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Scratchpad deleted", "id": id})
}

// SendAsEmail handles POST /api/scratchpads/:id/send-email
func (h *Handler) SendAsEmail(c *gin.Context) {
	db := middleware.GetDB(c)
	id := c.Param("id")
	var scratchpad models.Scratchpad
	if err := db.First(&scratchpad, id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Scratchpad not found"})
		return
	}

	var req struct {
		Recipients []string `json:"recipients"`
		Subject    string   `json:"subject"`
		Message    string   `json:"message"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.Recipients) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No recipients provided for email delivery."})
		return
	}

	if err := h.sendEmail(req.Recipients, req.Subject, req.Message, scratchpad); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email sent successfully"})
}

func (h *Handler) sendEmail(to []string, subject, customMsg string, s models.Scratchpad) error {
	from := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")

	if from == "" || pass == "" {
		return fmt.Errorf("SMTP credentials not configured")
	}

	body := fmt.Sprintf(
		"Subject: %s\n"+
			"MIME-version: 1.0;\n"+
			"Content-Type: text/html; charset=\"UTF-8\";\n\n"+
			"<html><body>"+
			"<h2>Scratchpad Update</h2>"+
			"<p><strong>Priority:</strong> %s</p>"+
			"<hr/>"+
			"<p>%s</p>"+
			"<br/>"+
			"<div style='background: #f4f4f4; padding: 10px; border-radius: 5px;'>"+
			"<h4>Content:</h4>"+
			"<p>%s</p>"+
			"</div>"+
			"<p><small>Created at: %s</small></p>"+
			"</body></html>",
		subject, s.Priority, customMsg, s.Content, s.CreatedAt.Format("Jan 02, 2006 15:04"),
	)

	auth := smtp.PlainAuth("", from, pass, host)
	err := smtp.SendMail(host+":"+port, auth, from, to, []byte(body))
	return err
}

// syncNoteToParent writes the scratchpad content into the linked entity's notes
// and returns the item enriched with link metadata (where the copy was stored),
// which is what later edits/deletes use to find it again.
func (h *Handler) syncNoteToParent(db db.Database, item models.LinkedItem, content string) models.LinkedItem {
	createdAt := time.Now().Format(time.RFC3339)
	newEntry := models.NoteEntry{Text: content, CreatedAt: createdAt}
	item.NoteCreatedAt = createdAt

	switch item.Type {
	case "lead":
		var lead models.Lead
		if err := db.First(&lead, "id = ?", item.ID); err == nil {
			lead.Details.NotesHistory = append(lead.Details.NotesHistory, newEntry)
			db.Save(&lead)
			item.TargetType, item.TargetID = "lead", item.ID
		}
	case "job":
		// If task name is provided, find task by stage_id + task_name
		if item.TaskName != "" && item.StageID != "" {
			var task models.StageTask
			if err := db.First(&task, "stage_id = ? AND name = ?", item.StageID, item.TaskName); err == nil {
				task.NotesHistory = append(task.NotesHistory, newEntry)
				db.Save(&task)
				item.TargetType, item.TargetID = "task", task.ID
				return item
			}
		}

		// If stage is selected, sync to stage only
		if item.StageID != "" {
			var stage models.Stage
			if err := db.First(&stage, "id = ?", item.StageID); err == nil {
				stage.AddNote(content)
				db.Save(&stage)
				item.TargetType, item.TargetID = "stage", stage.ID
				return item
			}
		}

		// No task or stage but a note type was chosen — store on the job's estimate
		if item.NoteType != "" {
			if estID := h.resolveJobEstimateID(db, item); estID != "" {
				if h.appendEstimateNote(db, estID, item.NoteType, newEntry) {
					item.TargetType, item.TargetID = "estimate", estID
					return item
				}
			}
		}

		// No task or stage — fallback to job's associated lead
		var job models.Job
		if err := db.First(&job, "id = ?", item.ID); err == nil && job.LeadID != "" {
			sub := h.syncNoteToParent(db, models.LinkedItem{ID: job.LeadID, Type: "lead"}, content)
			item.TargetType, item.TargetID, item.NoteCreatedAt = sub.TargetType, sub.TargetID, sub.NoteCreatedAt
		}
	case "estimate":
		// Stage selected — save into the estimate stage notes (Project Management)
		if item.StageID != "" {
			var stage models.Stage
			if err := db.First(&stage, "id = ?", item.StageID); err == nil {
				entryID := stage.ResidentialEntryID
				if entryID == 0 {
					if parsed, perr := strconv.ParseUint(item.ResidentialEntryID, 10, 64); perr == nil {
						entryID = uint(parsed)
					}
				}
				stageNote := models.EstimateStageNote{
					StageID:            stage.ID,
					ResidentialEntryID: entryID,
					Text:               content,
				}
				if err := db.Create(&stageNote); err != nil {
					fmt.Printf("Error saving estimate stage note: %v\n", err)
				} else {
					item.TargetType, item.TargetID, item.StageNoteID = "estimate_stage", stage.ID, stageNote.ID
				}
				return item
			}
		}

		// No stage — customer / internal notes (or legacy NotesHistory if no type given)
		if h.appendEstimateNote(db, item.ID, item.NoteType, newEntry) {
			item.TargetType, item.TargetID = "estimate", item.ID
		}
	}

	return item
}

// changeNoteEntries edits (remove=false) or removes (remove=true) one note in a
// NoteEntry list. It matches on CreatedAt + text first, then falls back to the
// latest entry with the same text. Returns the new list and whether it changed.
func changeNoteEntries(notes []models.NoteEntry, createdAt, oldText, newText string, remove bool) ([]models.NoteEntry, bool) {
	idx := -1
	if createdAt != "" {
		for i := range notes {
			if notes[i].CreatedAt == createdAt && notes[i].Text == oldText {
				idx = i
				break
			}
		}
	}
	if idx == -1 {
		for i := len(notes) - 1; i >= 0; i-- {
			if notes[i].Text == oldText {
				idx = i
				break
			}
		}
	}
	if idx == -1 {
		return notes, false
	}

	out := make([]models.NoteEntry, 0, len(notes))
	out = append(out, notes[:idx]...)
	if !remove {
		e := notes[idx]
		e.Text = newText
		out = append(out, e)
	}
	out = append(out, notes[idx+1:]...)
	return out, true
}

// updateLinkedNote pushes a scratchpad edit/delete to the copy of the note that
// syncNoteToParent stored. Items without link metadata (notes created before
// this feature) are skipped.
func (h *Handler) updateLinkedNote(db db.Database, item models.LinkedItem, oldText, newText string, remove bool) {
	switch item.TargetType {
	case "lead":
		var lead models.Lead
		if err := db.First(&lead, "id = ?", item.TargetID); err != nil {
			return
		}
		if notes, changed := changeNoteEntries(lead.Details.NotesHistory, item.NoteCreatedAt, oldText, newText, remove); changed {
			lead.Details.NotesHistory = notes
			db.Save(&lead)
		}

	case "task":
		var task models.StageTask
		if err := db.First(&task, "id = ?", item.TargetID); err != nil {
			return
		}
		if notes, changed := changeNoteEntries(task.NotesHistory, item.NoteCreatedAt, oldText, newText, remove); changed {
			task.NotesHistory = notes
			db.Save(&task)
		}

	case "stage":
		var stage models.Stage
		if err := db.First(&stage, "id = ?", item.TargetID); err != nil {
			return
		}
		notes := stage.GetNotesArray()
		idx := -1
		for i := len(notes) - 1; i >= 0; i-- {
			if notes[i] == oldText {
				idx = i
				break
			}
		}
		if idx == -1 {
			return
		}
		if remove {
			notes = append(notes[:idx], notes[idx+1:]...)
		} else {
			notes[idx] = newText
		}
		if b, err := json.Marshal(notes); err == nil {
			stage.Notes = datatypes.JSON(b)
			db.Save(&stage)
		}

	case "estimate":
		var est models.Estimate
		if err := db.First(&est, "id = ?", item.TargetID); err != nil {
			return
		}
		var list *[]models.NoteEntry
		switch strings.ToLower(strings.TrimSpace(item.NoteType)) {
		case "customer":
			list = &est.Details.CustomerNotesHistory
		case "internal":
			list = &est.Details.InternalNotesHistory
		default:
			list = &est.Details.NotesHistory
		}
		if notes, changed := changeNoteEntries(*list, item.NoteCreatedAt, oldText, newText, remove); changed {
			*list = notes
			db.Save(&est)
		}

	case "estimate_stage":
		if item.StageNoteID == 0 {
			return
		}
		if remove {
			db.Delete(&models.EstimateStageNote{}, "id = ?", item.StageNoteID)
		} else {
			db.Model(&models.EstimateStageNote{}).Where("id = ?", item.StageNoteID).Update("text", newText)
		}
	}
}

// removeScratchpadLinks is the reverse direction: when a note is deleted from a
// job stage / estimate stage, remove the matching link from its scratchpad, and
// delete the scratchpad itself once it has no links left.
// stageNoteID must be set for estimate_stage notes; for others the text is matched.
func removeScratchpadLinks(store db.Database, targetType, targetID, text string, stageNoteID uint) {
	probe, err := json.Marshal([]map[string]string{{"target_type": targetType, "target_id": targetID}})
	if err != nil {
		return
	}

	var pads []models.Scratchpad
	if err := store.Model(&models.Scratchpad{}).
		Where("linked_items::jsonb @> ?", string(probe)).
		Order("created_at DESC").
		Find(&pads); err != nil {
		fmt.Printf("removeScratchpadLinks: lookup failed: %v\n", err)
		return
	}

	for _, pad := range pads {
		items := parseLinkedItems(pad.LinkedItems)
		matched := -1
		for i, it := range items {
			if it.TargetType != targetType || it.TargetID != targetID {
				continue
			}
			if stageNoteID != 0 {
				if it.StageNoteID != stageNoteID {
					continue
				}
			} else if pad.Content != text {
				continue
			}
			matched = i
			break
		}
		if matched == -1 {
			continue
		}

		items = append(items[:matched], items[matched+1:]...)
		if len(items) == 0 {
			store.Delete(&pad)
		} else if b, err := json.Marshal(items); err == nil {
			pad.LinkedItems = models.JSONB(b)
			store.Save(&pad)
		}
		return // one scratchpad per deleted note
	}
}

// resolveJobEstimateID finds the estimate that backs a job. The scratchpad sends
// the job number as item.ID, which is NOT an estimate ID, so we resolve it via
// the job's lead and pick the latest estimate for that lead (the -vN version the
// job page works on). Falls back to any estimate ID the frontend supplied.
func (h *Handler) resolveJobEstimateID(db db.Database, item models.LinkedItem) string {
	leadID := ""

	var job models.Job
	if err := db.First(&job, "id = ?", item.ID); err == nil {
		leadID = strings.TrimSpace(fmt.Sprintf("%v", job.LeadID))
	}

	if leadID != "" {
		var est models.Estimate
		if err := db.Where("lead_id = ?", leadID).Order("created_at DESC").First(&est); err == nil {
			return est.ID
		}
	}

	// Fallbacks: estimate ID supplied by the frontend, then the item ID itself
	for _, candidate := range []string{strings.TrimSpace(item.EstimateID), strings.TrimSpace(item.ID)} {
		if candidate == "" {
			continue
		}
		var est models.Estimate
		if err := db.First(&est, "id = ?", candidate); err == nil {
			return est.ID
		}
	}

	return ""
}

// appendEstimateNote appends a note to an estimate's customer / internal notes
// (or the legacy NotesHistory when noteType is empty). Returns false if the
// estimate could not be found.
func (h *Handler) appendEstimateNote(db db.Database, estimateID, noteType string, entry models.NoteEntry) bool {
	var est models.Estimate
	if err := db.First(&est, "id = ?", estimateID); err != nil {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(noteType)) {
	case "customer":
		est.Details.CustomerNotesHistory = append(est.Details.CustomerNotesHistory, entry)
	case "internal":
		est.Details.InternalNotesHistory = append(est.Details.InternalNotesHistory, entry)
	default:
		est.Details.NotesHistory = append(est.Details.NotesHistory, entry)
	}
	db.Save(&est)
	return true
}

// GetScratchpadsByAgent handles GET /api/scratchpads/by-agent/:agent_id
func (h *Handler) GetScratchpadsByAgent(c *gin.Context) {
	db := middleware.GetDB(c)
	agentID := c.Param("agent_id")
	if agentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agent_id is required"})
		return
	}

	var scratchpads []models.Scratchpad
	// Match any scratchpad whose assigned_agents JSON array contains the given agent_id
	if err := db.Model(&models.Scratchpad{}).
		Where("assigned_agents::jsonb @> ?", fmt.Sprintf(`["%s"]`, agentID)).
		Order("created_at DESC").
		Find(&scratchpads); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch scratchpads"})
		return
	}

	c.JSON(http.StatusOK, scratchpads)
}