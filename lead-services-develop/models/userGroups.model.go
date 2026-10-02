package models

type UserGroups struct {
	ID        int     `json:"id" gorm:"primaryKey"`
	GroupName string  `json:"group_name"`
	Users     []Agent `json:"users" gorm:"many2many:user_group_agents;"`
}
