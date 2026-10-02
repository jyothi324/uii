package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type JSONB json.RawMessage

func (j JSONB) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("[]"), nil
	}
	return j, nil
}

func (j *JSONB) UnmarshalJSON(data []byte) error {
	if j == nil {
		return errors.New("JSONB: UnmarshalJSON on nil pointer")
	}
	*j = append((*j)[0:0], data...)
	return nil
}

func (j JSONB) Value() (driver.Value, error) {
	if len(j) == 0 {
		return "[]", nil
	}
	return string(j), nil
}

func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	s, ok := value.([]byte)
	if !ok {
		return errors.New("invalid scan source for JSONB")
	}
	*j = append((*j)[0:0], s...)
	return nil
}

type LinkedItem struct {
	ID                   string `json:"id"`
	Type                 string `json:"type"`
	Reference            string `json:"reference"`
	ResidentialEntryID   string `json:"residential_entry_id,omitempty"`
	ResidentialEntryName string `json:"residential_entry_name,omitempty"`
	StageID              string `json:"stage_id,omitempty"`
	StageName            string `json:"stage_name,omitempty"`
	TaskID               string `json:"task_id,omitempty"`
	TaskName             string `json:"task_name,omitempty"`
	NoteType string `json:"note_type,omitempty"`
	// EstimateID is the estimate that backs a job (used for job customer/internal notes).
	EstimateID string `json:"estimate_id,omitempty"`

	// edits/deletes can find the exact copy of the note. Not sent by the frontend.
	NoteCreatedAt string `json:"note_created_at,omitempty"` 
	TargetType    string `json:"target_type,omitempty"`     
	TargetID      string `json:"target_id,omitempty"`       
	StageNoteID   uint   `json:"stage_note_id,omitempty"`   
}

type Scratchpad struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	Content        string     `gorm:"not null" json:"content"`
	Priority       string     `gorm:"default:'medium';index" json:"priority"`
	LinkedItems    JSONB      `gorm:"type:jsonb" json:"linked_items"`
	Tags           JSONB      `gorm:"type:jsonb" json:"tags"`
	IsFavorite     bool       `gorm:"default:false" json:"is_favorite"`
	AssignedAgents JSONB      `gorm:"type:jsonb;default:'[]'" json:"assigned_agents"`
	ReminderAt     *time.Time `gorm:"index" json:"reminder_at"`
}
