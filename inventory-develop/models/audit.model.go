package models

import "time"

type AuditSession struct {
	ID            uint        `gorm:"primaryKey"`
	Title         string      `gorm:"type:text;not null" json:"title"`
	StartDate time.Time `gorm:"type:timestamptz;not null" json:"start_date"`
	Location      string      `gorm:"type:text;not null" json:"location"`
	SessionType   string      `gorm:"type:text;not null;default:'slab'" json:"session_type"` // "slab" or "non_slab"
	InternalNotes string      `gorm:"type:text" json:"internal_notes"`
	Status         string     `gorm:"type:text;not null;default:'active'" json:"status"` // active, completed
	ProductionTeam string     `gorm:"type:text" json:"production_team"` // comma-separated list of production team member names
	SystemCount    int        `gorm:"not null;default:0" json:"system_count"`      // locked-in total, grows only when a new material is first searched
	LockedMaterialIDs string  `gorm:"type:text" json:"-"`                          // comma-separated material IDs already counted into SystemCount
	CompletedBy    string     `gorm:"type:text" json:"completed_by"`
	CompletedAt   *time.Time  `json:"completed_at"`
	CreatedAt     time.Time   `gorm:"autoCreateTime" json:"CreatedAt"`
	UpdatedAt     time.Time   `gorm:"autoUpdateTime" json:"UpdatedAt"`
	Lines         []AuditLine `gorm:"foreignKey:AuditSessionID" json:"lines"`
}

// AuditLine represents one physically serialized inventory unit (a chunk)
// snapshotted into a session, or an "extra" unidentified scan discovered
// during the count that didn't match anything expected.
type AuditLine struct {
	ID             uint       `gorm:"primaryKey"`
	AuditSessionID uint       `gorm:"index;not null" json:"audit_session_id"`
	MaterialID     uint       `gorm:"index" json:"material_id"`
	ChunkID        *uint      `gorm:"index" json:"chunk_id"`
	Serialnum      string     `gorm:"type:text;index" json:"serialnum"`
	Lot            string     `gorm:"type:text" json:"lot"`
	SlabNumber     string     `gorm:"type:text" json:"slab_number"`
	Bin            string     `gorm:"type:text" json:"bin"`
	ProductName    string     `gorm:"type:text" json:"product_name"`
	Type           string     `gorm:"type:text" json:"type"`
	Category       string     `gorm:"type:text" json:"category"`
	SubCategory    string     `gorm:"type:text" json:"sub_category"`
	Dimensions     string     `gorm:"type:text" json:"dimensions"`
	ExpectedStatus string     `gorm:"type:text" json:"expected_status"`
	ScanStatus     string     `gorm:"type:text;not null;default:'unreconciled'" json:"scan_status"` // unreconciled, reconciled, unidentified, missing
	IsExtra        bool       `gorm:"default:false" json:"is_extra"`
	Notes          string     `gorm:"type:text" json:"notes"`
	ScannedAt      *time.Time `json:"scanned_at"`
	CreatedAt      time.Time  `gorm:"autoCreateTime" json:"CreatedAt"`
	UpdatedAt      time.Time  `gorm:"autoUpdateTime" json:"UpdatedAt"`
}