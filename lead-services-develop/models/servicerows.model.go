package models

type ServiceRow struct {
	ID                 uint              `gorm:"primaryKey;"`
	Details            ServiceRowDetails `gorm:"type:jsonb;serializer:json"`
	EstimateID         string
	Estimate           Estimate `json:"-" gorm:"foreignKey:EstimateID;constraint:OnDelete:SET NULL"`
	ResidentialEntryID uint
	ResidentialEntry   ResidentialEntry ` gorm:"foreignKey:ResidentialEntryID;constraint:OnDelete:SET NULL"`
}

type ServiceRowDetails struct {
	Name               string
	PackageName        string
	CategoryID         uint
	CategoryName       string
	ServiceGroupName   string
	AdditionalDiscount float64 //not using
	Description        string
	UnitPrice          float64
	Total              float64
	Quantity           uint
	IsCustom           bool
	IsGroup            bool                `json:"IsGroup,omitempty"`
	Services           []ServiceRowDetails `json:"services,omitempty"`
	SortOrder          int
}
