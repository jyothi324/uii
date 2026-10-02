package models

import "time"

type ImagePair struct {
	Uploaded string `json:"uploaded"`
	Result   string `json:"result"`
}

type VisualizeCustomer struct {
	ID            uint        `gorm:"primaryKey;autoIncrement" json:"id"`
	FirstName     string      `gorm:"type:varchar(100);not null" json:"first_name"`
	LastName      string      `gorm:"type:varchar(100);not null" json:"last_name"`
	Email         string      `gorm:"type:varchar(255);unique;not null;index" json:"email"`
	Phone         string      `gorm:"type:varchar(50)" json:"phone"`
	CompanyName   string      `gorm:"type:varchar(255)" json:"company_name"`
	CompanyEIN    string      `gorm:"type:varchar(100)" json:"company_ein"`
	EmailVerified bool        `gorm:"default:false" json:"email_verified"`
	UsageCount    int         `gorm:"default:5" json:"usage_count"`
	Images        []ImagePair `gorm:"serializer:json" json:"images"`
	CreatedAt     time.Time   `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time   `json:"updated_at" gorm:"autoUpdateTime"`
}
