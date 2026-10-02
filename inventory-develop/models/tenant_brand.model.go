package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type TenantAddress struct {
	CompanyName string `json:"companyName"`
	Street    string `json:"street"`
	UnitSuite string `json:"unitSuite"`
	City      string `json:"city"`
	State     string `json:"state"`
	Zip       string `json:"zip"`
	Country   string `json:"country"`
	Website   string `json:"website"`
	Email     string `json:"email"`
	Contact   string `json:"contact"`
	Latitude  string `json:"latitude"`   // ← add
	Longitude string `json:"longitude"` 
}

func (a TenantAddress) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *TenantAddress) Scan(value interface{}) error {
	if value == nil {
		*a = TenantAddress{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan TenantAddress: expected []byte, got %T", value)
	}

	return json.Unmarshal(bytes, a)
}

// TenantBrand stores this tenant's own branding info, synced from the www onboarding record.
type TenantBrand struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	CompanyName string    `gorm:"column:company_name;not null" json:"company_name"`
	LogoURL     string    `gorm:"column:logo_url" json:"logo_url"`
	Subdomain   string    `gorm:"column:subdomain;not null" json:"subdomain"`
	Address     TenantAddress `gorm:"type:jsonb;column:address;default:'{}'" json:"address"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
