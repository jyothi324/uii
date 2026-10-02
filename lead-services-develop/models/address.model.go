package models


// not required as per the recent changes in lead structure
type Address struct {
	ID        uint    `gorm:"primaryKey;autoIncrement"`
	State     string  `gorm:"type:text;not null"`
	County    string  `gorm:"type:text;not null"`
	Street    string  `gorm:"type:varchar(255);not null"`
	UnitSuite string  `gorm:"type:varchar(255);not null"`
	City      string  `gorm:"type:varchar(255);not null"`
	Zip       string  `gorm:"default:false"`
	Formatted string  `json:"formatted"`
	Lat       float64 `json:"lat"`
	Lng       float64 `json:"lng"`
}
