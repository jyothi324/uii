package models

import "time"

type PurchaseReceived struct {
	ID              uint          `gorm:"primaryKey"`
	PurchaseID uint          `gorm:"index" json:"purchase_id"`
	InvoiceID       string        `gorm:"type:text;not null" json:"invoice_id"`
	Purchase   Purchase  `gorm:"foreignKey:PurchaseID;references:ID" json:"purchase"`
	Price           float64       `gorm:"type:numeric;not null"`
	ReceivedQuantity float64   `json:"received_quantity"`
	Image            string    `gorm:"type:text;default:''" json:"image"`
	CreatedAt       time.Time     `gorm:"autoCreateTime"`
	UpdatedAt       time.Time     `gorm:"autoUpdateTime"`
}
