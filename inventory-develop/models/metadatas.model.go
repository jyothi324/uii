package models

import (
	"time"
)

type MaterialType struct {
	ID        		uint           `gorm:"primaryKey"`
	Name            string         `gorm:"type:text;not null;index" json:"name"`
	Unit      string    `gorm:"type:text" json:"unit"`
	CreatedAt       time.Time      `gorm:"autoCreateTime"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime"`

}

type MaterialCategory struct{
	ID        		uint           `gorm:"primaryKey"`
	Name			string         `gorm:"type:text;not null;index" json:"name"`
	CreatedAt       time.Time      `gorm:"autoCreateTime"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime"`
}

type MaterialSubCategory struct{
	ID        		uint           `gorm:"primaryKey"`
	Name			string         `gorm:"type:text;not null;index" json:"name"`
	CreatedAt       time.Time      `gorm:"autoCreateTime"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime"`
}

type MaterialOrigin struct{
	ID        		uint           `gorm:"primaryKey"`
	Name			string         `gorm:"type:text;not null;index" json:"name"`
	CreatedAt       time.Time      `gorm:"autoCreateTime"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime"`
}

type MaterialFinish struct{
	ID        		uint           `gorm:"primaryKey"`
	Name			string         `gorm:"type:text;not null;index" json:"name"`
	CreatedAt       time.Time      `gorm:"autoCreateTime"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime"`
}

type MaterialBaseColor struct{
	ID        		uint           `gorm:"primaryKey"`
	Name			string         `gorm:"type:text;not null;index" json:"name"`
	CreatedAt       time.Time      `gorm:"autoCreateTime"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime"`
}

type MaterialLevel struct {
	ID				uint           `gorm:"primaryKey"`
	Name			string         `gorm:"type:text;not null;index" json:"name"`
	CreatedAt       time.Time      `gorm:"autoCreateTime"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime"`
}

type MetadataModel interface {
	SetName(string)
	SetCreatedAt(time.Time)
	SetUpdatedAt(time.Time)
}

func (m *MaterialType) SetName(n string)         { m.Name = n }
func (m *MaterialType) SetCreatedAt(t time.Time) { m.CreatedAt = t }
func (m *MaterialType) SetUpdatedAt(t time.Time) { m.UpdatedAt = t }

func (m *MaterialCategory) SetName(n string)         { m.Name = n }
func (m *MaterialCategory) SetCreatedAt(t time.Time) { m.CreatedAt = t }
func (m *MaterialCategory) SetUpdatedAt(t time.Time) { m.UpdatedAt = t }

func (m *MaterialSubCategory) SetName(n string)         { m.Name = n }
func (m *MaterialSubCategory) SetCreatedAt(t time.Time) { m.CreatedAt = t }
func (m *MaterialSubCategory) SetUpdatedAt(t time.Time) { m.UpdatedAt = t }

func (m *MaterialOrigin) SetName(n string)         { m.Name = n }
func (m *MaterialOrigin) SetCreatedAt(t time.Time) { m.CreatedAt = t }
func (m *MaterialOrigin) SetUpdatedAt(t time.Time) { m.UpdatedAt = t }

func (m *MaterialFinish) SetName(n string)         { m.Name = n }
func (m *MaterialFinish) SetCreatedAt(t time.Time) { m.CreatedAt = t }
func (m *MaterialFinish) SetUpdatedAt(t time.Time) { m.UpdatedAt = t }

func (m *MaterialBaseColor) SetName(n string)         { m.Name = n }
func (m *MaterialBaseColor) SetCreatedAt(t time.Time) { m.CreatedAt = t }
func (m *MaterialBaseColor) SetUpdatedAt(t time.Time) { m.UpdatedAt = t }

func (m *MaterialLevel) SetName(n string)         { m.Name = n }
func (m *MaterialLevel) SetCreatedAt(t time.Time) { m.CreatedAt = t }
func (m *MaterialLevel) SetUpdatedAt(t time.Time) { m.UpdatedAt = t }