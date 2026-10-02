package models

type CountyTax struct {
	ID            uint   `gorm:"primaryKey;"`
	Name          string `gorm:"type:text;not null;index"`
	TaxPercentage float32
}
