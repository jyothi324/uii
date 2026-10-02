package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// ─── Types nested inside Details ──────────────────────────────────────────────

type Address struct {
	Street    string   `json:"street"`
	UnitSuite string   `json:"unitSuite"`
	City      string   `json:"city"`
	State     string   `json:"state"`
	Zip       string   `json:"zip"`
	Country   string   `json:"country"`
	Lat       *float64 `json:"lat,omitempty"`
	Lng       *float64 `json:"lng,omitempty"`
}

type Note struct {
	Text      string `json:"text"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// ─── SupplierDetails – stored as a single "details" JSONB column ──────────────

type SupplierDetails struct {
	Address              Address `json:"address"`
	PaymentTerms         string  `json:"paymentTerms,omitempty"`
	PricingType          string  `json:"pricingType,omitempty"`
	IsReferred           bool    `json:"isReferred,omitempty"`
	ReferrerName         string  `json:"referrerName,omitempty"`
	ReferrerPhone        string  `json:"referrerPhone,omitempty"`
	ReferrerEmail        string  `json:"referrerEmail,omitempty"`
	ReferrerCompanyName  string  `json:"referrerCompanyName,omitempty"`
	ReferrerCompanyPhone string  `json:"referrerCompanyPhone,omitempty"`
	ReferrerCompanyEmail string  `json:"referrerCompanyEmail,omitempty"`
	Notes                []Note  `json:"notes,omitempty"`
}

func (d SupplierDetails) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (d *SupplierDetails) Scan(value interface{}) error {
	if value == nil {
		*d = SupplierDetails{}
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan SupplierDetails: expected []byte, got %%T", value)
	}
	return json.Unmarshal(b, d)
}

// ─── Supplier ─────────────────────────────────────────────────────────────────

type Supplier struct {
	ID       uint    `gorm:"primaryKey"            json:"id"`
	Name     string  `gorm:"type:text;not null"    json:"name"`
	Phone    *string `gorm:"type:text;uniqueIndex" json:"phone"`
	Email    *string `gorm:"type:text;uniqueIndex" json:"email"`
	// Location string  `gorm:"type:text"             json:"location"`
	Category string  `gorm:"type:text;not null"    json:"category"`
	Website  *string `gorm:"type:text"             json:"website"`

	// All new fields live inside this single JSONB column
	Details SupplierDetails `gorm:"type:jsonb;column:details;default:\'{}\'" json:"details"`
		PrimaryContact       *string `json:"primaryContact,omitempty"`
	SecondaryContact     *string `json:"secondaryContact,omitempty"`
	PrimaryPhone         *string `json:"primaryPhone,omitempty"`
	SecondaryPhone       *string `json:"secondaryPhone,omitempty"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}