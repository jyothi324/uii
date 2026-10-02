package models

type DeliveryTypeOption struct {
	ID      uint             `gorm:"primaryKey" json:"ID"`
	Name    string           `json:"name" gorm:"type:text;not null"`
	Address CustomerAddress `json:"address" gorm:"type:jsonb;serializer:json"`
	IsPickup *bool `json:"isPickup" gorm:"column:is_pickup"`
}


type LeadSourceOption struct {
	ID   uint   `gorm:"primaryKey" json:"ID"`
	Name string `json:"name" gorm:"type:text;not null"`
}

type AssociateTypeOption struct {
	ID   uint   `gorm:"primaryKey" json:"ID"`
	Name string `json:"name" gorm:"type:text;not null"`
}