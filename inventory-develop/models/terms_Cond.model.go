package models

// type Terms struct {
// 	ID      uint   `gorm:"primaryKey" json:"id"`
// 	Content string `gorm:"type:text;not null" json:"content"`
// }

type Policy struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Title    string `gorm:"type:text;not null" json:"title"`
	Content  string `gorm:"type:text" json:"content"`
	Position int    `json:"position"`
	Selected bool   `gorm:"default:false" json:"selected"`
	DocType string `gorm:"type:varchar(20);not null;default:''" json:"doc_type"`
}
