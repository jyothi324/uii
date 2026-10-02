package models

import (
	"time"

	"gorm.io/gorm"
)

type MessageTemplate struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Name      string         `gorm:"uniqueIndex;not null" json:"name"`
	Subject   string         `json:"subject"`
	Content   string         `json:"content"`
	Category  string         `json:"category"`
}
