package models

type StageColor struct {
	ID    uint   `gorm:"primaryKey" json:"id"`
	Stage string `gorm:"type:text;not null" json:"stage_name"`
	Color string `gorm:"type:text;not null" json:"color_hex"`
}
