package models

type ManagerResidentialType struct {
	ID                   uint   `gorm:"primaryKey"`
	ResidentialEntryType string `json:"type" gorm:"type:text;"`
	Length               int64  `json:"length" gorm:"default:0;not null"`
}
