package models

import (
	"gorm.io/datatypes"
	"time"
)

type Job struct {
	ID        string         `gorm:"primaryKey" json:"id"`
	JobID     string         `gorm:"uniqueIndex" json:"job_id"`
	LeadID    string         `gorm:"index" json:"lead_id"`
	EstimateID *string       `json:"estimate_id"`
	
	Completed bool `gorm:"default:false;index" json:"completed"`

	Stages    datatypes.JSON `gorm:"type:jsonb" json:"stages"`

	Details   datatypes.JSON `gorm:"type:jsonb" json:"details"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
