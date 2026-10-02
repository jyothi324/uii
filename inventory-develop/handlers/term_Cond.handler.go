package handlers

import (
	"errors"
	"inventory/middleware"
	"inventory/models"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// parsePolicyDocType reads the optional ?type= query param. Valid values are
// "estimate" and "job". An empty value keeps the legacy (unscoped) behaviour.
func parsePolicyDocType(c *gin.Context) (string, bool) {
	t := strings.ToLower(strings.TrimSpace(c.Query("type")))
	switch t {
	case "", "estimate", "job":
		return t, true
	}
	return "", false
}

func HandleSaveTerms() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := middleware.GetDB(c)
		if db == nil {
			log.Println("DB is nil")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection unavailable"})
			return
		}

		docType, ok := parsePolicyDocType(c)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid type, expected 'estimate' or 'job'"})
			return
		}
		var req struct {
			Policies []struct {
				ID       string `json:"id"`
				Title    string `json:"title"`
				Content  string `json:"content"`
				Selected bool   `json:"selected"`
			} `json:"policies"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			log.Println("JSON bind error:", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		if len(req.Policies) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "at least one policy is required"})
			return
		}

		// Replace-all approach: wipe existing policies, insert fresh. When a
		// type is given, only that type's policies are replaced.
		var delErr error
		if docType == "" {
			delErr = db.Model(&models.Policy{}).Delete(&models.Policy{}, "1 = 1")
		} else {
			delErr = db.Model(&models.Policy{}).Delete(&models.Policy{}, "doc_type = ?", docType)
		}
		if delErr != nil {
			log.Println("Delete error:", delErr)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save policies"})
			return
		}

		// Only one policy may be marked as selected at a time. If the
		// caller sent more than one (or none), fall back to the first
		// one sent so a policy is always selected for printing.
		selectedIndex := -1
		for i, p := range req.Policies {
			if p.Selected {
				selectedIndex = i
				break
			}
		}
		if selectedIndex == -1 {
			selectedIndex = 0
		}

		newPolicies := make([]models.Policy, 0, len(req.Policies))
		for i, p := range req.Policies {
			newPolicies = append(newPolicies, models.Policy{
				Title:    p.Title,
				Content:  p.Content,
				Position: i,
				Selected: i == selectedIndex,
				DocType:  docType,
			})
		}

		if err := db.Create(&newPolicies); err != nil {
			log.Println("Create error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save policies"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "saved successfully"})
	}
}

func HandleGetTerms() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := middleware.GetDB(c)
		if db == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection unavailable"})
			return
		}

		docType, ok := parsePolicyDocType(c)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid type, expected 'estimate' or 'job'"})
			return
		}

		var policies []models.Policy
		var findErr error
		if docType == "" {
			findErr = db.Order("position asc").Find(&policies)
		} else {
			findErr = db.Where("doc_type = ?", docType).Order("position asc").Find(&policies)
			if findErr == nil && len(policies) == 0 {
				// Nothing saved for this type yet: fall back to the legacy
				// shared policies so existing tenants keep their current terms.
				findErr = db.Where("doc_type = '' OR doc_type IS NULL").Order("position asc").Find(&policies)
			}
		}
		if findErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
			return
		}

		if len(policies) > 0 {
			// Backward compatibility: rows saved before the "selected"
			// column existed will all have Selected == false. In that
			// case, report the first one as selected so the frontend
			// always has exactly one policy to print with.
			hasSelected := false
			for _, p := range policies {
				if p.Selected {
					hasSelected = true
					break
				}
			}

			resp := make([]gin.H, 0, len(policies))
			for i, p := range policies {
				selected := p.Selected
				if !hasSelected && i == 0 {
					selected = true
				}
				resp = append(resp, gin.H{
					"id":       p.ID,
					"title":    p.Title,
					"content":  p.Content,
					"selected": selected,
				})
			}
			c.JSON(http.StatusOK, gin.H{"policies": resp})
			return
		}

		// Fallback: no policies yet — check legacy single-Terms table.
		// var terms models.Terms
		// dbErr := db.Model(&models.Terms{}).First(&terms)
		// if dbErr == nil && terms.Content != "" {
		// 	c.JSON(http.StatusOK, gin.H{"content": terms.Content})
		// 	return
		// }

		c.JSON(http.StatusOK, gin.H{"policies": []gin.H{}})
	}
}

func HandleSaveVerbiage() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := middleware.GetDB(c)
		if db == nil {
			log.Println("DB is nil")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection unavailable"})
			return
		}

		var req struct {
			ChangeOrder  string `json:"change_order"`
			Cancellation string `json:"cancellation"`
			Disputes     string `json:"disputes"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			log.Println("JSON bind error:", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		var verbiage models.Verbiage
		dbErr := db.Model(&models.Verbiage{}).First(&verbiage)

		if dbErr != nil {
			if errors.Is(dbErr, gorm.ErrRecordNotFound) {
				verbiage = models.Verbiage{
					ChangeOrder:  req.ChangeOrder,
					Cancellation: req.Cancellation,
					Disputes:     req.Disputes,
				}
				if err := db.Create(&verbiage); err != nil {
					log.Println("Create error:", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save verbiage"})
					return
				}
			} else {
				log.Println("Real DB error:", dbErr)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
				return
			}
		} else {
			verbiage.ChangeOrder = req.ChangeOrder
			verbiage.Cancellation = req.Cancellation
			verbiage.Disputes = req.Disputes
			if err := db.Save(&verbiage); err != nil {
				log.Println("Save error:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update verbiage"})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{"message": "saved successfully"})
	}
}

func HandleGetVerbiage() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := middleware.GetDB(c)
		if db == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection unavailable"})
			return
		}

		var verbiage models.Verbiage
		dbErr := db.Model(&models.Verbiage{}).First(&verbiage)
		if dbErr != nil {
			if errors.Is(dbErr, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusOK, gin.H{
					"change_order": "",
					"cancellation": "",
					"disputes":     "",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"change_order": verbiage.ChangeOrder,
			"cancellation": verbiage.Cancellation,
			"disputes":     verbiage.Disputes,
		})
	}
}
