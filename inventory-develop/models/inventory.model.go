package models



type ManagerResidentialEntry struct {
	ID                   uint   `gorm:"primaryKey"`
	ResidentialEntryName string `json:"name" gorm:"type:text"`
}

type ManagerResidentialType struct {
	ID                   uint   `gorm:"primaryKey"`
	ResidentialEntryType string `json:"type" gorm:"type:text;"`
}


