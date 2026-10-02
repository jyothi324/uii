package models

import (
	"time"

	"gorm.io/gorm"
)

type PickupLocations struct {
	ID        uint           `gorm:"primaryKey;autoIncrement"`
	Name      string         `gorm:"type:text;not null;"`
	Street    string         `gorm:"type:text;not null"`
	UnitSuite string         `gorm:"type:text"`
	City      string         `gorm:"type:text;not null"`
	State     string         `gorm:"type:text;not null"`
	Zip       string         `gorm:"type:text;not null"`
	Country   string         `gorm:"type:text;not null"`
	CreatedAt time.Time      `json:"-" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"-" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	Lat       float64
	Lng       float64
}
