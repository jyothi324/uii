package models

import "github.com/lib/pq"

type MaterialRowDetails struct {
	Description string ` gorm:"type:varchar(255);not null"`
	Material    string ` gorm:"type:varchar(255);not null"`
	Category    string ` gorm:"type:varchar(255);not null"`
	SubCategory string ` gorm:"type:varchar(255);not null"` //not using
	Quantity    float64

	AdditionalDiscount float64 //not using
	UnitPrice          float64
	Total              float64
	IsCustom           bool
	SortOrder          int
}

type MaterialRow struct {
	ID                 uint               `gorm:"primaryKey;"`
	Details            MaterialRowDetails `gorm:"type:jsonb;serializer:json"`
	EstimateID         string
	Estimate           Estimate `json:"-" gorm:"foreignKey:EstimateID;constraint:OnDelete:SET NULL"`
	ResidentialEntryID uint
	ResidentialEntry   ResidentialEntry ` gorm:"foreignKey:ResidentialEntryID;constraint:OnDelete:SET NULL"`
	ChunkIDs           pq.StringArray   `gorm:"type:text[]"`
}
