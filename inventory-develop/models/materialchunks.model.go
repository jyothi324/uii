package models

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/datatypes"
)

type Materialchunks struct {
	ID                 uint   `gorm:"primaryKey"`
	MaterialID         uint   `gorm:"index" json:"material_id"`
	PurchaseReceivedID uint   `gorm:"index;default:0" json:"purchase_received_id"`
	Serialnum          string `gorm:"type:text;not null;" json:"serialnum"`
	// Barcode         string         `gorm:"type:text;not null;" json:"barcode"`
	Slabno          *string        `gorm:"type:text;" json:"slabno"`
	Dimensions      string         `gorm:"type:text;" json:"dimensions"`
	Lot             string         `gorm:"type:text;" json:"lot"`
	PresentLocation string         `gorm:"type:text;not null" json:"present_location"`
	RetailCost      float64        `gorm:"type:numeric;not null" json:"retail_cost"`
	ProCost         float64        `gorm:"type:numeric;not null" json:"procost"`
	Pictures        pq.StringArray `gorm:"type:text[];not null;default:'{}'" json:"pictures"`
	CreatedAt       time.Time      `gorm:"autoCreateTime"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime"`
	Status          string         `json:"status"`
	BaseColor       string         `gorm:"type:text;default:''" json:"base_color"`
	Level           string         `gorm:"type:text;default:''" json:"level"`
	JobID           string         `gorm:"type:varchar(255);not null;default:''" json:"job_id"`
	Bin             *string        `gorm:"type:text" json:"bin"`
	SupplierDetails datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"supplier_details"`
	Notes           *string        `gorm:"type:text" json:"notes"`
	InventoryType   string         `gorm:"type:text" json:"inventory_type"`
	InvoiceNo       string         `gorm:"type:text;default:''" json:"invoice_no"`
	EditNotes       datatypes.JSON `gorm:"type:jsonb;default:'[]'" json:"edit_notes"`
}
