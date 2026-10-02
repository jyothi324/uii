package models

type Verbiage struct {
	ID           uint   `gorm:"primaryKey"`
	ChangeOrder  string `json:"change_order"`
	Cancellation string `json:"cancellation"`
	Disputes     string `json:"disputes"`
}