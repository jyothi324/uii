package models

import "time"

type EstimateSchedule struct {
	ID            string       `gorm:"type:varchar(36);primaryKey" json:"id"`
	EstimateID    string       `gorm:"type:text;not null;index;unique" json:"estimateId"`
	Estimate      *Estimate    `gorm:"foreignKey:EstimateID;constraint:OnDelete:CASCADE" json:"-"`
	ScheduleType  ScheduleType `gorm:"type:varchar(20);not null" json:"ScheduleType"`
	ScheduledDate *time.Time   `gorm:"type:date" json:"ScheduledDate"`
	CreatedAt     time.Time    `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt     time.Time    `gorm:"autoUpdateTime" json:"updatedAt"`
}
