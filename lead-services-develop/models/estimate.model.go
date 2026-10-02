package models

import (
	"time"

	"gorm.io/gorm"
)

type Estimate struct {
	ID                 string             `gorm:"primaryKey;"`
	CreatedAt          time.Time          `gorm:"autoCreateTime"`
	UpdatedAt          time.Time          `gorm:"autoUpdateTime"`
	DeletedAt          gorm.DeletedAt     `gorm:"index"`
	LeadID             *string            `gorm:"default:null"`
	Lead               Lead               `gorm:"foreignKey:LeadID;constraint:OnDelete:SET NULL;omitempty"`
	Name               string             `gorm:"type:text;not null;"`
	Sales1PersonID     string
	Sales1PersonName   string
	Sales1Person       *Agent `gorm:"foreignKey:Sales1PersonID"`
	Sales2PersonID     string
	Sales2PersonName   string
	Sales2Person       *Agent `gorm:"foreignKey:Sales2PersonID"`
	MaterialTotal      float64            `gorm:"type:numeric;not null"`
	ServiceTotal       float64            `gorm:"type:numeric;not null"`
	Details            EstimateDetails    `gorm:"type:jsonb;serializer:json"`
	Status             string             `gorm:"type:text;not null;"`
	MaterialRows       []MaterialRow      `gorm:"foreignKey:EstimateID;constraint:OnDelete:SET NULL"`
	ServiceRows        []ServiceRow       `gorm:"foreignKey:EstimateID;constraint:OnDelete:SET NULL"`
	ResidentialEntries []ResidentialEntry `gorm:"foreignKey:EstimateID;constraint:OnDelete:SET NULL"`
	EstimatedSqFt  float64     `json:"EstimatedSqFt"`  
    CorrectedSqFt  float64     `json:"CorrectedSqFt"`
	Schedule           *EstimateSchedule  `gorm:"foreignKey:EstimateID;constraint:OnDelete:CASCADE"`
}

type EstimateDetails struct {
	LaborCost             float64     `gorm:"type:numeric;not null"`
	// DiscountID            *uint       `gorm:"default:null"`                                  // not required as per recent changes
	// Discount              *Discount   `gorm:"foreignKey:DiscountID;constraint:OnDelete:SET NULL"` // not required as per recent changes
	TaxID                 *uint       `gorm:"default:null"`
	Tax                   *CountyTax  `gorm:"foreignKey:TaxID;constraint:OnDelete:SET NULL"`
	Total                 float64     `gorm:"type:numeric;not null"`
	TaxAmount             float64     `gorm:"type:numeric;not null"`
	DiscountAmount        float64     `gorm:"type:numeric;not null"`
	NotesHistory          []NoteEntry `json:"notesHistory"`
	InternalNotesHistory  []NoteEntry `json:"InternalNotesHistory"`
	CustomerNotesHistory  []NoteEntry `json:"CustomerNotesHistory"`
	FollowUpCompletedDate string      `json:"followUpCompletedDate,omitempty"`
	Payments              []Payment   `gorm:"foreignKey:EstimateID;constraint:OnDelete:SET NULL"`
	// Installation Details card (Job page only) 
	InstallationSalesPersonID   string `json:",omitempty"`
	InstallationSalesPersonName string `json:",omitempty"`
	InstallationDateTime        string `json:",omitempty"`
}

type Payment struct {
	InvoiceId	*string    `gorm:"default:null"`
	Total		float64    `gorm:"type:numeric;not null"`
	PaymentRecorded []PaymentRecorded `gorm:"foreignKey:PaymentID;constraint:OnDelete:SET NULL"`
}

type PaymentRecorded struct {
	Date 	  time.Time `gorm:"autoCreateTime"`
	PaymentType string    `gorm:"type:text;not null"`
	Amount	  float64   `gorm:"type:numeric;not null"`
	RemainingAmount float64 `gorm:"type:numeric;not null"`
}
