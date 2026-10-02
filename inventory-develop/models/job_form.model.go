package models

import (
	"time"

	"gorm.io/datatypes"
)

// type JobFormTemplate struct {
// 	ID           uint           `gorm:"primaryKey" json:"id"`
// 	Stage        string         `gorm:"type:text;not null;uniqueIndex" json:"stage"`
// 	Name         string         `gorm:"type:text;not null" json:"name"`
// 	Description  string         `gorm:"type:text" json:"description"`
// 	TemplateData datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'" json:"template_data"`
// 	IsActive     bool           `gorm:"not null;default:true" json:"is_active"`
// 	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
// 	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
// }

type JobTaskForm struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	TemplateID       *uint          `gorm:"index" json:"template_id"`
	JobID            string         `gorm:"type:text;not null;uniqueIndex:idx_job_task_stage" json:"job_id"`
	TaskID           string         `gorm:"type:text;not null;uniqueIndex:idx_job_task_stage" json:"task_id"`
	TaskName         string         `gorm:"type:text" json:"task_name"`
	Stage            string         `gorm:"type:text;not null;uniqueIndex:idx_job_task_stage;index" json:"stage"`
	SalesPerson      string         `gorm:"type:text" json:"salesperson"`
	FormData         datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'" json:"form_data"`
	TemplateSnapshot datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'" json:"template_snapshot"`
	TaskSnapshot     datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'" json:"task_snapshot"`
	CreatedAt        time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}
