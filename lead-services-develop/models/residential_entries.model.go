package models

type ResidentialEntry struct {
	ID                        uint                    `gorm:"primaryKey;autoIncrement"`
	ManagerResidentialEntryID uint                    // BIGINT
	ManagerResidentialEntry   ManagerResidentialEntry `gorm:"foreignKey:ManagerResidentialEntryID;constraint:OnDelete:SET NULL"`
	ManagerResidentialTypeID  uint
	ManagerResidentialType    ManagerResidentialType `gorm:"foreignKey:ManagerResidentialTypeID;constraint:OnDelete:SET NULL"`
	Details                   string                 `gorm:"type:text"`
	LeadID                    *string
	Lead                      Lead `json:"-" gorm:"foreignKey:LeadID;constraint:OnDelete:SET NULL"`
	EstimateID                *string
	Estimate                  Estimate `json:"-" gorm:"foreignKey:EstimateID;constraint:OnDelete:SET NULL"`
	ServicesTaxExempt         bool     `gorm:"default:true" json:"ServicesTaxExempt"`
}
