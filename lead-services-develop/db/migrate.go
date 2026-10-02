package db

import (
	"fmt"

	"github.com/tuliporg/agent-management-api/internal/activity"
	"github.com/tuliporg/agent-management-api/models"
)

func Migrate(db Database) {
	err := db.AutoMigrate(
		&models.Agent{},
		&models.Address{},
		&models.PickupLocations{},
		&models.AssociateCustomer{},
		&models.Lead{},
		&models.File{},
		&models.ResidentialEntry{},
		&models.Estimate{},
		&models.CountyTax{},
		&models.Service{},
		&models.ServiceCategory{},
		&models.ServiceGroup{},
		&models.Task{},
		&models.ManagerResidentialEntry{},
		&models.MaterialRow{},
		&models.ServiceRow{},
		&models.ManagerResidentialType{},
		&models.Customer{},
		&models.Stage{},
		&models.StageTask{},
		&models.StageUser{},
		&models.LeadSchedule{},
		&models.EstimateSchedule{},
		&models.Profile{},
		&models.Permissions{},
		&models.UserGroups{},
		&models.Job{},
		&models.Scratchpad{},
		&models.Otp{},
		&models.EstimateProjectTask{},
		&models.EstimateStageUser{},
		&models.EstimateStageNote{},
		&activity.ActivityEvent{},
		&models.DeliveryTypeOption{}, 
		&models.LeadSourceOption{}, 
		&models.AssociateTypeOption{},
	)
	if err != nil {
		panic("failed to auto-migrate models: " + err.Error())
	} else {
		fmt.Errorf("Auto Migration Succesfull")
	}
}
