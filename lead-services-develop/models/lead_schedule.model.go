package models

import "time"

type ScheduleType string

const (
	ScheduleToday    ScheduleType = "Today"
	ScheduleTomorrow ScheduleType = "Tomorrow"
	ScheduleCustom   ScheduleType = "Custom"
)

type LeadSchedule struct {
	ID            string       `gorm:"type:varchar(36);primaryKey" json:"id"`
	LeadID        string       `gorm:"type:varchar(20);not null;index;unique" json:"leadId"`
	Lead          *Lead        `gorm:"foreignKey:LeadID;constraint:OnDelete:CASCADE" json:"-"`
	ScheduleType  ScheduleType `gorm:"type:varchar(20);not null" json:"ScheduleType"`
	ScheduledDate *time.Time   `gorm:"type:date" json:"ScheduledDate"`
	CreatedAt     time.Time    `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt     time.Time    `gorm:"autoUpdateTime" json:"updatedAt"`
}
