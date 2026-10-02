package models

import "time"

// JobFormQuestion is a master template question, managed from Settings.
// Deleting one removes it from the template for all future job forms.
// Existing job forms are unaffected because they hold their own snapshot
// in JobTaskForm.TemplateSnapshot / FormData.
type JobFormQuestion struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Stage     string    `gorm:"type:text;not null;index" json:"stage"`
	Category  string    `gorm:"type:text;not null;index" json:"category"` // e.g. "fabricator_checklist"
	Text      string    `gorm:"type:text;not null" json:"text"`
	SortOrder int       `gorm:"not null;default:0" json:"sort_order"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// JobFormSectionTitles holds the editable section titles shown in the form
// (Materials / Services / Fabricator Checklist), one row per stage.
type JobFormSectionTitles struct {
	ID                       uint      `gorm:"primaryKey" json:"id"`
	Stage                    string    `gorm:"type:text;not null;uniqueIndex" json:"stage"`
	MaterialsTitle           string    `gorm:"type:text" json:"materials_title"`
	ServicesTitle            string    `gorm:"type:text" json:"services_title"`
	FabricatorChecklistTitle string    `gorm:"type:text" json:"fabricator_checklist_title"`
	CreatedAt                time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt                time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}