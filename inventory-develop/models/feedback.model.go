package models

import "time"

type Feedback struct {
	ID        uint      `json:"id"         gorm:"primaryKey;autoIncrement"`
	FirstName string    `json:"firstName"  gorm:"not null"`
	LastName  string    `json:"lastName"   gorm:"not null"`
	Email     string    `json:"email"      gorm:"not null"`
	Phone     *string   `json:"phone"`
	Message   string    `json:"message"    gorm:"type:text;not null"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}