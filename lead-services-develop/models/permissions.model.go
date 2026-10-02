package models

type Permissions struct {
	ID         int          `json:"id" gorm:"primaryKey"`
	Action     string       `json:"action"`
	Source     string       `json:"source"`
	UserGroups []UserGroups `json:"user_groups" gorm:"many2many:permission_user_groups;"`
	User       []Agent      `json:"users" gorm:"many2many:permission_agents;"`
	// Roles      pq.StringArray `gorm:"type:text[]" json:"roles"`
}

type BulkPermissionUpdate struct {
	ID     int    `json:"id"`
	Source string `json:"source"`
	Action string `json:"action"`

	User []struct {
		ID string `json:"id"`
	} `json:"User"`

	UserGroups []struct {
		ID int `json:"id"`
	} `json:"UserGroups"`
}

type BulkPermissionRequest struct {
	Updates []BulkPermissionUpdate `json:"updates"`
	ShowPricesUpdate *ShowPricesUpdate      `json:"show_prices_update,omitempty"`

}

type ShowPricesUpdate struct {
	AgentIDs   []string `json:"agent_ids"`
	ShowPrices bool     `json:"show_prices"`
}
type UpdatePermissionRequest struct {
	Permissions
	ShowPricesUpdate *ShowPricesUpdate `json:"show_prices_update,omitempty"`
}