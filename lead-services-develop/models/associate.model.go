package models

import "time"

type AssociateRole string

const (
	ARCHITECT  AssociateRole = "ARCHITECT"
	DESIGNER   AssociateRole = "DESIGNER"
	CONTRACTOR AssociateRole = "CONTRACTOR"
)

type AssociateCustomer struct {
	ID      string           `gorm:"primaryKey;"`
	Details AssociateDetails `json:"details" gorm:"type:jsonb;serializer:json"`
	Leads   []*Lead          `json:"-" gorm:"many2many:lead_associates;"`
	Name    string           `gorm:"type:text;not null;"`
	Role    AssociateRole    `gorm:"type:text;not null"`
}
type AssociateDetails struct {
	FirstName    string
	LastName     string
	Email        string
	Phone        string
	CompanyName  string
	CompanyPhone string
	CompanyEmail string

	CompanyContactName                 string          `json:"CompanyContactName"`
	CompanySecondaryEmail              string          `json:"CompanySecondaryEmail"`
	CompanyAddress                     CustomerAddress `json:"CompanyAddress" gorm:"type:jsonb;serializer:json"`
	CompanyBillingAddress              CustomerAddress `json:"CompanyBillingAddress" gorm:"type:jsonb;serializer:json"`
	CompanyDetails                     CustomerDetails `json:"CompanyDetails" gorm:"type:jsonb;serializer:json"`
	CompanySalesPerson1                string          `json:"CompanySalesPerson1"`
	CompanySalesPerson2                string          `json:"CompanySalesPerson2"`
	CompanyProjectManager              string          `json:"CompanyProjectManager"`
	CompanyCustomerSince               string          `json:"CompanyCustomerSince"`
	CompanySalesTax                    string          `json:"CompanySalesTax"`
	CompanyEINNumber                   string          `json:"CompanyEINNumber"`
	CompanyTaxExempt                   bool            `json:"CompanyTaxExempt"`
	CompanyTaxExemptReason             string          `json:"CompanyTaxExemptReason"`
	CompanyExemptCertificateNumber     string          `json:"CompanyExemptCertificateNumber"`
	CompanyExemptExpiryDate            string          `json:"CompanyExemptExpiryDate"`
	CompanySpecialDeliveryInstructions string          `json:"CompanySpecialDeliveryInstructions"`
	CompanyCollectionNotes             string          `json:"CompanyCollectionNotes"`

	CompanyID string   `json:"company_id"`
	Company   Customer `json:"company" gorm:"foreignKey:CompanyID;references:ID;constraint:OnDelete:SET NULL"`
}
type CustomerAddress struct {
	Street    string   `json:"Street"`
	UnitSuite string   `json:"UnitSuite"`
	City      string   `json:"City"`
	State     string   `json:"State"`
	Zip       string   `json:"Zip"`
	Country   string   `json:"Country"`
	Lat       *float64 `json:"lat,omitempty"`
	Lng       *float64 `json:"lng,omitempty"`
}

type CustomerDetails struct {
	PaymentTerms string `json:"PaymentTerms"`
	PricingType  string `json:"PricingType"`
}

type AdditionalPersonContact struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

type Customer struct {
	ID                          string                    `json:"id" gorm:"primaryKey"`
	Name                        string                    `json:"name"`
	Phone                       string                    `json:"phone"`
	Email                       string                    `json:"email"`
	SecondaryEmail              string                    `json:"secondaryEmail" gorm:"type:text"`
	AccountingEmail             string                    `json:"accountingEmail" gorm:"type:text"`
	SecondaryPhoneNumber        string                    `json:"secondaryPhoneNumber" gorm:"type:text"`
	CellNumber                  string                    `json:"cellNumber" gorm:"type:text"`
	Address                     CustomerAddress           `json:"address"      gorm:"type:jsonb;serializer:json"`
	BillingAddress              CustomerAddress           `json:"billingAddress" gorm:"type:jsonb;serializer:json"`
	ContactName                 string                    `json:"contactName" gorm:"type:text"`
	AdditionalPersonContacts    []AdditionalPersonContact `json:"additionalPersonContacts" gorm:"type:jsonb;serializer:json"`
	SalesPerson1                string                    `json:"SalesPerson1" gorm:"type:text"`
	SalesPerson2                string                    `json:"SalesPerson2" gorm:"type:text"`
	ProjectManager              string                    `json:"ProjectManager" gorm:"type:text"`
	CustomerSince               string                    `json:"CustomerSince" gorm:"type:text"`
	SalesTax                    string                    `json:"SalesTax" gorm:"type:text"`
	EINNumber                   string                    `json:"EINNumber" gorm:"type:text"`
	TaxExempt                   bool                      `json:"TaxExempt" gorm:"type:boolean"`
	TaxExemptReason             string                    `json:"TaxExemptReason" gorm:"type:text"`
	ExemptCertificateNumber     string                    `json:"ExemptCertificateNumber" gorm:"type:text"`
	ExemptExpiryDate            string                    `json:"ExemptExpiryDate" gorm:"type:text"`
	SpecialDeliveryInstructions string                    `json:"SpecialDeliveryInstructions" gorm:"type:text"`
	CollectionNotes             string                    `json:"CollectionNotes" gorm:"type:text"`
	Details                     CustomerDetails           `json:"details"  gorm:"type:jsonb;serializer:json"`
	CreatedAt                   time.Time                 `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt                   time.Time                 `json:"updatedAt" gorm:"autoUpdateTime"`
	QBCustomerID    		string        	 `gorm:"type:text;uniqueIndex:idx_profiles_qb_customer_id_unique,where:qb_customer_id IS NOT NULL AND qb_customer_id <> ''" json:"qb_customer_id,omitempty"`
}
