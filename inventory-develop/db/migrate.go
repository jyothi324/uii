package db

import (
	"fmt"
	"inventory/models"
)

func Migrate(db Database) {
	err := db.AutoMigrate(
		&models.Purchase{},
		&models.PurchaseReceived{},
		&models.Materials{},
		&models.Supplier{},
		&models.Materialchunks{},
		&models.VisualizeCustomer{},
		&models.Feedback{},
		&models.MaterialType{},
		&models.MaterialCategory{},
		&models.MaterialSubCategory{},
		&models.MaterialOrigin{},
		&models.MaterialFinish{},
		&models.MaterialBaseColor{},
		&models.MaterialLevel{},
		&models.StageColor{},
		// &models.Terms{},
		&models.Policy{},
		&models.TenantBrand{},
		&models.JobTaskForm{},
		&models.Verbiage{},
		&models.EmailTemplate{},
		&models.JobFormQuestion{},
		&models.JobFormSectionTitles{},
		&models.MessageTemplate{},
		&models.AuditSession{},
		&models.AuditLine{},
	)
	if err != nil {
		fmt.Println("Error migrating database:", err)
	} else {
		fmt.Println("Database migrated successfully")
	}

}
