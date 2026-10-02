package models

import "time"

type Otp struct {
	Email     string    `gorm:"primaryKey" json:"email"`
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expires_at"`
}