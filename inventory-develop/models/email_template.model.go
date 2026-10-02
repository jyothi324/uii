package models

import "time"

// EmailTemplate stores per-tenant email templates editable from settings.
type EmailTemplate struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"type:text;not null;uniqueIndex" json:"name"`
	Subject   string    `gorm:"type:text" json:"subject"`
	Content   string    `gorm:"type:text" json:"content"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
