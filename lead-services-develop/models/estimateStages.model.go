package models

import "time"

// EstimateProjectTask is scoped to the Estimate "Project Management" tab only.
// It intentionally does NOT reuse the Job tracker's task table.
type EstimateProjectTask struct {
	ID                 string      `gorm:"primaryKey"`
	StageID            string      `gorm:"index;not null"`
	ResidentialEntryID uint        `gorm:"index;not null"`
	Name               string      `gorm:"type:text"`
	StartDate          string      `gorm:"type:text"`
	StartTime          string      `gorm:"type:text"`
	EndTime            string      `gorm:"type:text"`
	Duration           int         `gorm:"type:int"`
	AssignedTo         string      `gorm:"type:text"`
	NotesHistory       []NoteEntry        `gorm:"type:jsonb;serializer:json"`
	Projects           []StageTaskProject `gorm:"type:jsonb;serializer:json"`
	Completed          bool        `gorm:"default:false"`
	Cancelled          bool        `gorm:"default:false"`
	IsRunning          bool        `gorm:"default:false"`
	Elapsed            int         `gorm:"default:0"` // seconds
	LastResumedAt      *time.Time
	Photos             []string  `gorm:"type:jsonb;serializer:json"`
	CreatedAt          time.Time `gorm:"autoCreateTime"`
	UpdatedAt          time.Time `gorm:"autoUpdateTime"`
}

// EstimateStageUser tracks who is assigned to the Template stage on an
// Estimate. Separate table from the Job tracker's stage-user assignments.
type EstimateStageUser struct {
	ID                 string    `gorm:"primaryKey"`
	StageID            string    `gorm:"index;not null"`
	ResidentialEntryID uint      `gorm:"index;not null"`
	UserID             string    `gorm:"type:text"`
	Name               string    `gorm:"type:text"`
	TaskName           string    `gorm:"type:text"`
	CreatedAt          time.Time `gorm:"autoCreateTime"`
	UpdatedAt          time.Time `gorm:"autoUpdateTime"`
}

// EstimateStageNote stores notes added from the Estimate Project Management
// tab. Separate from both Estimate.Details.NotesHistory (estimate-level
// notes) and the Job tracker's stage notes.
type EstimateStageNote struct {
	ID                 uint      `gorm:"primaryKey;autoIncrement"`
	StageID            string    `gorm:"index;not null"`
	ResidentialEntryID uint      `gorm:"index;not null"`
	Text               string    `gorm:"type:text"`
	CreatedAt          time.Time `gorm:"autoCreateTime"`
}