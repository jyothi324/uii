package models

import "gorm.io/datatypes"

type Profile struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	Name            string         `gorm:"type:text;not null" json:"Name"`
	Phone           string         `gorm:"type:text;not null" json:"Phone"`
	Secondaryphone  string         `gorm:"type:text" json:"SecondaryPhone"`
	Cellphone       string         `gorm:"type:text" json:"CellPhone"`
	Email           string         `gorm:"type:text;not null" json:"Email"`
	SecondaryEmail  string         `gorm:"type:text" json:"SecondaryEmail"`
	ShippingAddress datatypes.JSON `gorm:"type:json" json:"ShippingAddress"`
	BillingAddress  datatypes.JSON `gorm:"type:json" json:"BillingAddress"`
	SameAsShipping  bool           `gorm:"default:false" json:"SameAsShipping"`
	QBCustomerID    string         `gorm:"type:text;uniqueIndex:idx_profiles_qb_customer_id_unique,where:qb_customer_id IS NOT NULL AND qb_customer_id <> ''" json:"qb_customer_id,omitempty"`
	Details         datatypes.JSON `json:"Details"`
}
