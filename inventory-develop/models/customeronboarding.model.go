package models

import (
	"time"
)

// CustomerOnboarding represents a new tenant registration and its infrastructure status.
type CustomerOnboarding struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	CompanyName      string    `gorm:"column:company_name;not null" json:"company_name"`
	AdminName        string    `gorm:"column:admin_name;not null" json:"admin_name"`
	Email            string    `gorm:"column:email;not null;index" json:"email"`
	Address          string    `gorm:"column:address;type:text" json:"address"`
	Phone            string    `gorm:"column:phone" json:"phone"`
	Subdomain        string    `gorm:"column:subdomain;unique;not null" json:"subdomain"`
	LogoURL          string    `gorm:"column:logo_url" json:"logo_url"`
	Region           string    `gorm:"column:region" json:"region"`
	Timezone         string    `gorm:"column:timezone" json:"timezone"`
	Users            int       `gorm:"column:users" json:"users"`
	DBInstanceID     string    `gorm:"column:db_instance_id" json:"db_instance_id"`
	ProvisioningDate time.Time `gorm:"column:provisioning_date" json:"provisioning_date"`
	DeploymentStatus string    `gorm:"column:deployment_status;default:pending" json:"deployment_status"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
