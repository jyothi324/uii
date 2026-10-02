package models

import (
	"time"
)

// Agent represents an agent entity in the system.
type Role string

const (
	Agents  Role = "Agent"
	Manager Role = "Manager"
)

type AgentDetails struct {
	Branch      string `json:"branch" gorm:"type:varchar(255);not null"`
	Phone       string `json:"phone" gorm:"type:varchar(255);not null"`
	TOTPSecret  string `json:"totpsecret" gorm:"type:varchar(255);not null"`
	TOTPEnabled bool   `json:"totpenabled" gorm:"default:false"`
}

type Agent struct {
	ID             string       `gorm:"type:varchar(27);primaryKey"`
	Name           string       `json:"name" gorm:"type:varchar(100);not null"`
	Email          string       `json:"email" gorm:"type:varchar(255);unique;not null;index"`
	Password       string       `json:"password" gorm:"type:varchar(255);not null"` //need to be removed
	PasswordHashed string       `json:"password_hashed" gorm:"type:varchar(255);not null"`
	Details        AgentDetails `json:"details" gorm:"type:jsonb;serializer:json"`
	Role           Role         `json:"role" gorm:"type:varchar(20);"`
	AssigneeType   string       `json:"assignee_type" gorm:"type:varchar(20);"`
	Superuser      bool         `json:"superuser" gorm:"default:false"`
	ShowPrices     bool         `json:"show_prices" gorm:"default:false"`
	CreatedAt      time.Time    `json:"-" gorm:"autoCreateTime"`
	UpdatedAt      time.Time    `json:"-" gorm:"autoUpdateTime"`
	// DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}

type CreateAgentRequest struct {
    Agent
    TOTPCode   string `json:"totp_code"`
    TOTPSecret string `json:"totp_secret"`
}
