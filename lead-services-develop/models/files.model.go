package models

import (
	"time"
)

type File struct {
	ID             uint      `gorm:"primaryKey;"`
	Name           string    `gorm:"type:text;not null"`
	Type           string    `gorm:"type:text;not null"`
	Size           string    `gorm:"type:text;not null"`
	UploadedAt     time.Time `gorm:"type:timestamp;not null"`
	ProjectType    string
	StoredFileName string `gorm:"type:text;not null"`
	Category       string `gorm:"type:text;"`
	LeadID         *string
	FilePath       string `gorm:"type:text;not null"`
	Lead           Lead   `json:"-" gorm:"foreignKey:LeadID;constraint:OnDelete:SET NULL"`
	PresignedURL string `json:"PresignedURL,omitempty" gorm:"-"`
}
