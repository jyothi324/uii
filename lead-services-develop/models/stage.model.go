package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Stage represents a job stage in the lead management process
type Stage struct {
	ID                 string `gorm:"type:varchar(36);primaryKey"`
	LeadID             string `gorm:"type:varchar(36);not null;index"`
	JobID              string
	Name               string         `gorm:"type:varchar(100);not null"`
	Description        string         `gorm:"type:text"`
	State              string         `gorm:"type:varchar(20);check:state IN ('Not Started','In Progress','Completed')"`
	Date               *time.Time     `gorm:"default:null"`
	Notes              datatypes.JSON `gorm:"type:jsonb"`
	Team               string         `gorm:"type:varchar(255)"`
	CreatedAt          time.Time      `gorm:"autoCreateTime"`
	UpdatedAt          time.Time      `gorm:"autoUpdateTime"`
	DeletedAt          gorm.DeletedAt `gorm:"index"`
	ResidentialEntryID uint
	Lead               *Lead           `gorm:"foreignKey:LeadID;constraint:OnDelete:CASCADE"`
	Tasks              []StageTask     `gorm:"foreignKey:StageID;constraint:OnDelete:CASCADE"`
	// Materials          []StageMaterial `gorm:"foreignKey:StageID;constraint:OnDelete:CASCADE"` //not using currently
	Users              []StageUser     `gorm:"foreignKey:StageID;constraint:OnDelete:CASCADE"` 
}

type NoteEntries []NoteEntry

func (n NoteEntries) Value() (driver.Value, error) {
	if n == nil {
		return "[]", nil
	}
	b, err := json.Marshal(n)
	return string(b), err
}

func (n *NoteEntries) Scan(value interface{}) error {
	if value == nil {
		*n = NoteEntries{}
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case string:
		bytes = []byte(v)
	case []byte:
		bytes = v
	default:
		return fmt.Errorf("cannot scan type %T into NoteEntries", value)
	}
	return json.Unmarshal(bytes, n)
}

type StageTask struct {
	ID            string    `gorm:"type:varchar(36);primaryKey"`
	StageID       string    `gorm:"type:varchar(36);not null;index"`
	Name          string    `gorm:"type:varchar(255);not null"`
	DueDate       time.Time `gorm:"not null"`
	StartDate     time.Time
	Duration      int              `gorm:"not null"` // Planned duration in minutes
	Completed     bool             `gorm:"default:false"`
	Cancelled     bool             `gorm:"default:false"`
	IsRunning     bool             `gorm:"default:false"`
	Elapsed       int              `gorm:"default:0"` // seconds elapsed
	StartTime     *time.Time       `gorm:"type:timestamp with time zone"`
	LastResumedAt *time.Time       `gorm:"type:timestamp with time zone"`
	EndTime       *time.Time       `gorm:"type:timestamp with time zone"`
	CreatedAt     time.Time        `gorm:"autoCreateTime"`
	UpdatedAt     time.Time        `gorm:"autoUpdateTime"`
	AssignedTo 	  string 		   `gorm:"type:text;index"`
	DeletedAt     gorm.DeletedAt   `gorm:"index"`
	NotesHistory  NoteEntries      `gorm:"type:json" json:"NotesHistory"`
	Photos        datatypes.JSON   `gorm:"type:json"`
	Projects      datatypes.JSON   `gorm:"type:jsonb;default:'[]'" json:"projects"`
	Address       datatypes.JSON   `gorm:"type:jsonb"`
	// SourceEstimateTaskID is set ONLY on rows copied from Estimate Project
	// Management when an estimate is converted to a job. 
	SourceEstimateTaskID string     `gorm:"type:varchar(36);index"`
	Stage         *Stage           `gorm:"foreignKey:StageID;constraint:OnDelete:CASCADE"`
	// Timers        []StageTaskTimer `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE"`
}

type StageTaskProject struct {
	ID                        uint   `json:"id"`
	ResidentialEntryID        uint   `json:"residentialEntryId"`
	ManagerResidentialEntryID uint   `json:"managerResidentialEntryId"`
	ManagerResidentialTypeID  uint   `json:"managerResidentialTypeId"`
	Name                      string `json:"name"`
	Type                      string `json:"type"`
	Details                   string `json:"details"`
	Title                     string `json:"title"`
}

// StageTaskTimer represents a timer for a task
// type StageTaskTimer struct {
// 	ID                 string         `gorm:"type:varchar(36);primaryKey"`
// 	TaskID             string         `gorm:"type:varchar(36);not null;index"`
// 	StartTime          *time.Time     `gorm:"default:null"`
// 	LastResumedAt      *time.Time     `gorm:"default:null"`
// 	IsRunning          bool           `gorm:"default:false"`
// 	ElapsedTime        int            `gorm:"default:0"` // seconds
// 	CreatedAt          time.Time      `gorm:"autoCreateTime"`
// 	UpdatedAt          time.Time      `gorm:"autoUpdateTime"`
// 	DeletedAt          gorm.DeletedAt `gorm:"index"`
// 	ResidentialEntryID uint
// 	Task               *StageTask `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE"`
// }

// StageMaterial represents materials associated with a stage
// type StageMaterial struct {
// 	ID        string         `gorm:"type:varchar(36);primaryKey"`
// 	StageID   string         `gorm:"type:varchar(36);not null;index"`
// 	Name      string         `gorm:"type:varchar(255);not null"`
// 	CreatedAt time.Time      `gorm:"autoCreateTime"`
// 	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
// 	DeletedAt gorm.DeletedAt `gorm:"index"`

// 	Stage *Stage `gorm:"foreignKey:StageID;constraint:OnDelete:CASCADE"`
// }

// StageUser represents users assigned to a stage
type StageUser struct {
	ID        string         `gorm:"type:varchar(36);primaryKey"`
	StageID   string         `gorm:"type:varchar(36);not null;index"`
	UserID    string         `gorm:"type:varchar(36);not null"`
	Name      string         `gorm:"type:varchar(255);not null"`
	TaskName string		 	 `gorm:"type:varchar(255);not null;default:''"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Stage *Stage `gorm:"foreignKey:StageID;constraint:OnDelete:CASCADE"`
}

// func (s *Stage) AddNote(note string) {
// 	var notes []string
// 	if len(s.Notes) > 0 {
// 		_ = json.Unmarshal(s.Notes, &notes)
// 	}
// 	notes = append(notes, note)
// 	notesJSON, _ := json.Marshal(notes)
// 	s.Notes = datatypes.JSON(notesJSON)
// }

// // GetNotesArray returns the notes as a Go string slice
// func (s *Stage) GetNotesArray() []string {
// 	var notes []string
// 	if len(s.Notes) > 0 {
// 		_ = json.Unmarshal(s.Notes, &notes)
// 	}
// 	return notes
// }

// AddNote appends a note and ensures s.Notes becomes valid JSON.
// Returns an error if JSON marshalling fails.
func (s *Stage) AddNote(note string) error {
	var notes []string
	// If there is existing JSON, try to unmarshal it; if invalid, reset to empty slice.
	if len(s.Notes) > 0 {
		if err := json.Unmarshal(s.Notes, &notes); err != nil {
			notes = []string{}
		}
	}
	notes = append(notes, note)

	notesJSON, err := json.Marshal(notes)
	if err != nil {
		return err
	}
	s.Notes = datatypes.JSON(notesJSON)
	return nil
}

// GetNotesArray returns notes as a slice, safely handling empty/invalid JSON.
func (s *Stage) GetNotesArray() []string {
	var notes []string
	if len(s.Notes) == 0 {
		return []string{}
	}
	if err := json.Unmarshal(s.Notes, &notes); err != nil {
		// invalid JSON -> return empty slice
		return []string{}
	}
	return notes
}

// DeleteNote deletes the note at index (no-op if index invalid).
// Returns an error if marshalling fails.
func (s *Stage) DeleteNote(index int) error {
	var notes []string
	if len(s.Notes) > 0 {
		if err := json.Unmarshal(s.Notes, &notes); err != nil {
			notes = []string{}
		}
	}
	if index < 0 || index >= len(notes) {
		// nothing to delete
		return nil
	}
	notes = append(notes[:index], notes[index+1:]...)
	notesJSON, err := json.Marshal(notes)
	if err != nil {
		return err
	}
	s.Notes = datatypes.JSON(notesJSON)
	return nil
}
