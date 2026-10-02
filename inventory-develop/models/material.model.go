package models

import (
	"time"

	"github.com/lib/pq"
)

type Materials struct {
	ID              uint             `gorm:"primaryKey"`
	Serialnum       string           `gorm:"type:text;not null;uniqueIndex" json:"serialnum"`
	// Barcode         string           `gorm:"type:text;not null;uniqueIndex" json:"barcode"`
	// Slabno          *string          `gorm:"type:text;uniqueIndex" json:"slabno"`
	Name            string           `gorm:"type:text;not null;index" json:"name"`
	Category        string           `gorm:"type:text;not null" json:"category"`
	// PresentLocation string           `gorm:"type:text;not null" json:"present_location"`
	// Bin             *string          `gorm:"type:text" json:"bin"`
	Threshold       int              `gorm:"type:integer;default:0" json:"threshold"`
	Available       int              `gorm:"type:integer;default:0" json:"available"`
	RetailCost      float64          `gorm:"type:numeric;not null" json:"retail_cost"`
	ProCost         float64          `gorm:"type:numeric;not null" json:"procost"`
	Pictures        pq.StringArray   `gorm:"type:text[];default:'{}'" json:"pictures"`
	Unit            string           `gorm:"type:text"`
	Type            string           `gorm:"type:text" json:"type"`
	SubCategory     string           `gorm:"type:text" json:"subcategories"`
	Origin          string           `gorm:"type:text" json:"origin"`
	Finish          string           `gorm:"type:text" json:"finishes"`
	BaseColor       string           `gorm:"type:text" json:"base_colors"`
	Level           string           `gorm:"type:text" json:"level"`
	LotNo		    string           `gorm:"type:text" json:"lot_no"`
	Dimensions      string           `gorm:"type:text" json:"dimensions"`
	SquareFeet      float64          `gorm:"type:numeric" json:"square_feet"`
	CreatedAt       time.Time        `gorm:"autoCreateTime"`
	UpdatedAt       time.Time        `gorm:"autoUpdateTime"`
	Chunks          []Materialchunks `gorm:"foreignKey:MaterialID" json:"chunks"`
}
