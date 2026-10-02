package models

type ManagerResidentialEntry struct {
	ID                   uint   `gorm:"primaryKey"`
	ResidentialEntryName string `json:"name" gorm:"type:text"`
	Length               int64  `json:"length" gorm:"default:0;not null"`
}
