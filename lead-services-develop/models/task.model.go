package models

import (
	"time"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Task struct {
	ID          uint           `gorm:"primaryKey;"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	Title       string         `gorm:"type:text;not null;index"`
	Description string         `gorm:"type:text"`
	DueDate     string         `gorm:"type:text"` //not using currently
	EndDate     string         `gorm:"type:text"` 
	EndTime     string         `gorm:"type:text"`
	CompletedAt *time.Time     `gorm:"type:timestamp"`
	Priority    string         `gorm:"type:text;not null"`
	Status      string         `gorm:"type:text;not null;index"`
	TaskType    string         `gorm:"type:text;not null"`
	AssigneeID  string         `gorm:"type:text"`
	Photos      datatypes.JSON `gorm:"type:json"`
}
