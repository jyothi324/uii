package models

import (
	"time"

	"gorm.io/gorm"
)

type EnumType[T ~string] string

func (e EnumType[T]) GormDataType() string {
	return string(e)
}

type DeliveryTypes struct {
	CUSTOMER_PICKUP   DeliveryType
	CONTRACTOR_PICKUP DeliveryType
	SHIPPING          DeliveryType
}
type DetailedLeadStatus string

// not using all these statuses currently (need to be clarified)
const (
	ESTIMATE_NOT_SENT    DetailedLeadStatus = "ESTIMATE_NOT_SENT"
	ESTIMATE_SENT        DetailedLeadStatus = "ESTIMATE_SENT"
	ESTIMATE_APPROVED    DetailedLeadStatus = "ESTIMATE_APPROVED"
	QUOTATION_REQUESTED  DetailedLeadStatus = "QUOTATION_REQUESTED"
	QUOTATION_SENT       DetailedLeadStatus = "QUOTATION_SENT"
	QUOTATION_REJECTED   DetailedLeadStatus = "QUOTATION_REJECTED"
	QUOTATION_APPROVED   DetailedLeadStatus = "QUOTATION_APPROVED"
	WAITING_FOR_APPROVAL DetailedLeadStatus = "WAITING_FOR_APPROVAL"
	CLOSED               DetailedLeadStatus = "CLOSED"
)

type ReferrerDetails struct {
	Name       string `json:"name" gorm:"type:varchar(255);not null"`
	Email      string `json:"email" gorm:"type:varchar(255);not null"`
	Phone      string `json:"phone" gorm:"type:varchar(255);not null"`
	IsReferred bool   `json:"IsReferred" gorm:"default:false"`
}
type LeadType string

// not using these variables currently (need to be clarified)
const (
	WALK_IN LeadType = "walk_in"
	NORMAL  LeadType = "Normal"
	Website LeadType = "Website"
)

type DeliveryType string

const (
	CUSTOMER_PICKUP   DeliveryType = "Customer_Pickup"
	CONTRACTOR_PICKUP DeliveryType = "Contractor_Pickup"
	SHIPPING          DeliveryType = "Shipping"
)

type LeadStatus string

const (
	NEW         LeadStatus = "NEW"
	IN_PROGRESS LeadStatus = "IN_PROGRESS"
	COMPLETED   LeadStatus = "COMPLETED"
	REJECTED    LeadStatus = "REJECTED"
)

type PricingType string

const (
	PRO    PricingType = "PRO"
	RETAIL PricingType = "RETAIL"
)

type PaymentTerm string

const (
	CASH   PaymentTerm = "CASH"
	CREDIT PaymentTerm = "CREDIT"
	CHEQUE PaymentTerm = "CHEQUE"
)

type NoteEntry struct {
	Text      string `json:"text"`
	CreatedAt string `json:"createdAt"`
}

type LeadDetails struct {
	Phone                   string           `gorm:"type:text"`
	Secondaryphone          string           `gorm:"type:text" json:"SecondaryPhone"`
	Cellphone               string           `gorm:"type:text" json:"CellPhone"`
	// DiscountID              *uint            `gorm:"default:null;autoIncrement"`
	// Discount                *Discount        `gorm:"foreignKey:DiscountID;constraint:OnDelete:SET NULL"`
	TaxID                   *uint            `gorm:"default:null;autoIncrement"`
	CountyTax               *CountyTax       `gorm:"foreignKey:TaxID;constraint:OnDelete:SET NULL"`
	FaxNumber               string           `gorm:"type:text"`
	MobileNumber            string           `gorm:"type:text"`
	ShippingAddressID       *uint            `gorm:"default:null;autoIncrement"`                                //not using
	ShippingAddress         *Address         `gorm:"foreignKey:ShippingAddressID;constraint:OnDelete:SET NULL"` //not using
	BillingAddressID        *uint            `gorm:"default:null;autoIncrement"`                                //not using
	BillingAddress          *Address         `gorm:"foreignKey:BillingAddressID;constraint:OnDelete:SET NULL"`  //not using
	IsPickUpDelivery        bool             `gorm:"default:false"`
	BillToCustomer          string           `gorm:"type:text"`
	ShipToMobile            string           `gorm:"type:text"`
	ShipToEmail             string           `gorm:"type:text"`
	PickupDeliveryAddressID *uint            `gorm:"default:null"`
	PickupDeliveryAddress   *PickupLocations `gorm:"foreignKey:PickupDeliveryAddressID;constraint:OnDelete:SET NULL"`
	SecondaryEmail          string           `gorm:"type:text"`
	CustomerPhone           string           `gorm:"type:text" json:"CustomerPhone"`
	CustomerEmail           string           `gorm:"type:text" json:"CustomerEmail"`
	// Notes                   string           `gorm:"type:text"`
	IsCompanyContact      bool        `gorm:"default:false" json:"IsCompanyContact"`
	NotesHistory          []NoteEntry `json:"NotesHistory"`
	FollowUpCompletedDate string      `json:"FollowUpCompletedDate,omitempty"`
}

// Lead represents the leads table
type Lead struct {
	ID                 string             `gorm:"type:varchar(50);primaryKey;"`
	CreatedAt          time.Time          `gorm:"autoCreateTime"`
	UpdatedAt          time.Time          `gorm:"autoUpdateTime"`
	DeletedAt          gorm.DeletedAt     `gorm:"index"`
	LeadType           LeadType           `gorm:"type:varchar(20);not null"`
	JobName            string             `gorm:"type:varchar(100);not null"`
	DeliveryType       DeliveryType       `gorm:"type:varchar(50)"`
	Status             LeadStatus         `gorm:"not null;"`
	Details            LeadDetails        `gorm:"type:jsonb;serializer:json"`
	Referrer           ReferrerDetails    `gorm:"type:jsonb;serializer:json"`
	DetailedLeadStatus DetailedLeadStatus `gorm:"type:varchar(50)"`
	PaymentTerm        PaymentTerm        `gorm:"type:varchar(50)"`
	PricingType        PricingType        `gorm:"type:varchar(50)"`
	Favorite           string             `gorm:"type:text"`
	ProfilePic         string             `gorm:"type:text"`
	Sales1PersonID     string
	Sales1PersonName   string
	Sales1Person       *Agent `gorm:"foreignKey:Sales1PersonID"`
	Sales2PersonID     string
	Sales2PersonName   string
	Sales2Person       *Agent               `gorm:"foreignKey:Sales2PersonID"`
	Associates         []*AssociateCustomer `gorm:"many2many:lead_associates;references:ID"`
	Files              []File               `gorm:"foreignKey:LeadID;constraint:OnDelete:SET NULL"`
	ResidentialEntries []ResidentialEntry   `gorm:"foreignKey:LeadID;constraint:OnDelete:SET NULL"`
	Estimates          []Estimate           `gorm:"foreignKey:LeadID;constraint:OnDelete:SET NULL"`
	Schedule           *LeadSchedule        `gorm:"foreignKey:LeadID;constraint:OnDelete:CASCADE"`
	RejectionReason    string               `gorm:"type:varchar(255)"`
	QBCustomerID    		string        	 `gorm:"type:text;uniqueIndex:idx_profiles_qb_customer_id_unique,where:qb_customer_id IS NOT NULL AND qb_customer_id <> ''" json:"qb_customer_id,omitempty"`
}
