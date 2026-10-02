package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type PurchaseStatus string

type PurchaseItem struct {
	MaterialID   uint    `json:"material_id"`
	Dimensions   string  `json:"dimensions"`
	Quantity     string  `json:"quantity"`
	CurrentPrice string  `json:"current_price"`
	SquareFeet   float64 `json:"square_feet"`
	RetailCost   string  `json:"retail_cost"`
	ProCost      string  `json:"procost"`
	Bin          string  `json:"bin"`
	BaseColor    string  `json:"base_color"`
}

type PurchaseItems []PurchaseItem

// Value implements the driver.Valuer interface for database storage
func (p PurchaseItems) Value() (driver.Value, error) {
	if len(p) == 0 {
		return []byte("[]"), nil
	}
	return json.Marshal(p)
}

// Scan implements the sql.Scanner interface for reading from the database
func (p *PurchaseItems) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, p)
	case string:
		return json.Unmarshal([]byte(v), p)
	case nil:
		*p = make(PurchaseItems, 0)
		return nil
	default:
		return fmt.Errorf("failed to scan PurchaseItems: unsupported type %T", value)
	}
}

type Purchase struct {
	ID                uint               `gorm:"primaryKey"`
	Name              string             `gorm:"not null" json:"name"`
	MaterialID        uint               `json:"material_id"`
	Materials         Materials          `gorm:"foreignKey:material_id;references:ID"`
	Quantity          string             `gorm:"type:text" json:"quantity"`
	Status            string             `gorm:"type:text;default:'Purchase'" json:"status"`
	RejectionReason   string             `gorm:"type:text" json:"rejection_reason,omitempty"`
	CurrentPrice      string             `gorm:"type:text" json:"current_price"`
	SupplierID        uint               `json:"supplier_id"`
	Supplier          Supplier           `gorm:"foreignKey:SupplierID" json:"supplier"`
	Dimensions        string             `gorm:"type:text;" json:"dimensions"`
	SquareFeet        float64            `gorm:"type:numeric" json:"square_feet"`
	PurchaseItems     PurchaseItems      `gorm:"type:jsonb" json:"purchase_items"`
	CreatedAt         time.Time          `gorm:"autoCreateTime"`
	UpdatedAt         time.Time          `gorm:"autoUpdateTime"`
	BillId            string             `json:"bill_id"`
	InventoryType     string             `gorm:"type:text" json:"inventory_type"`
	Note              string             `gorm:"type:text" json:"note"`
	CountyTaxID       *uint              `gorm:"index" json:"county_tax_id"`
	CountyName        string             `gorm:"type:text" json:"county_name"`
	TaxPercentage     float64            `gorm:"type:numeric;default:0" json:"tax_percentage"`
	DeliveryCharges   float64            `gorm:"type:numeric;default:0" json:"delivery_charges"`
	AdditionalCharges []AdditionalCharge `gorm:"type:jsonb;serializer:json" json:"additional_charges"` 
	ReturnReason      string             `gorm:"type:text" json:"return_reason,omitempty"`
}

type AdditionalCharge struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
}
